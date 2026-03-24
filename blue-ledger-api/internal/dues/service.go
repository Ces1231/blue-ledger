package dues

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v78"
)

// DuesRecord represents a dues_records row.
type DuesRecord struct {
	ID                 string     `json:"id"`
	ChapterID          string     `json:"chapter_id"`
	MemberID           string     `json:"member_id"`
	Semester           string     `json:"semester"`
	AmountCents        int        `json:"amount_cents"`
	DueDate            string     `json:"due_date"`
	PaidAt             *time.Time `json:"paid_at,omitempty"`
	PaymentMethod      *string    `json:"payment_method,omitempty"`
	ZeffyFormID        *string    `json:"zeffy_form_id,omitempty"`
	ZeffyTransactionID *string    `json:"zeffy_transaction_id,omitempty"`
	Status             string     `json:"status"`
	XPAwarded          bool       `json:"xp_awarded"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ZeffyWebhookPayload is the expected payload from Zapier → POST /webhooks/zeffy
type ZeffyWebhookPayload struct {
	MemberEmail     string `json:"member_email"`
	AmountCents     int    `json:"amount_cents"`
	TransactionID   string `json:"transaction_id"`
	Semester        string `json:"semester"`
}

// DuesService defines the dues domain business logic interface.
type DuesService interface {
	List(ctx context.Context, chapterID, semester string, page, perPage int) ([]*DuesRecord, int, error)
	ListForMember(ctx context.Context, memberID string) ([]*DuesRecord, error)
	Create(ctx context.Context, chapterID, memberID, semester string, amountCents int, dueDate string) (*DuesRecord, error)
	MarkPaid(ctx context.Context, duesID, paymentMethod string) (*DuesRecord, error)
	HandleZeffyWebhook(ctx context.Context, chapterID string, payload ZeffyWebhookPayload) error
	HandleStripeSubscriptionWebhook(ctx context.Context, event *stripe.Event) error
}

type duesService struct {
	db            *pgxpool.Pool
	stripeKey     string
	webhookSecret string
}

// NewDuesService creates a new dues service.
func NewDuesService(db *pgxpool.Pool, stripeKey, webhookSecret string) DuesService {
	return &duesService{
		db:            db,
		stripeKey:     stripeKey,
		webhookSecret: webhookSecret,
	}
}

func (s *duesService) List(ctx context.Context, chapterID, semester string, page, perPage int) ([]*DuesRecord, int, error) {
	offset := (page - 1) * perPage

	var total int
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM dues_records WHERE chapter_id = $1 AND ($2 = '' OR semester = $2)`,
		chapterID, semester).Scan(&total)

	rows, err := s.db.Query(ctx, `
		SELECT id, chapter_id, member_id, semester, amount_cents, due_date,
		       paid_at, payment_method, zeffy_form_id, zeffy_transaction_id,
		       status, xp_awarded, created_at, updated_at
		FROM dues_records
		WHERE chapter_id = $1 AND ($2 = '' OR semester = $2)
		ORDER BY due_date ASC
		LIMIT $3 OFFSET $4`,
		chapterID, semester, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list dues: %w", err)
	}
	defer rows.Close()

	return scanDuesRecords(rows, total)
}

func (s *duesService) ListForMember(ctx context.Context, memberID string) ([]*DuesRecord, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, chapter_id, member_id, semester, amount_cents, due_date,
		       paid_at, payment_method, zeffy_form_id, zeffy_transaction_id,
		       status, xp_awarded, created_at, updated_at
		FROM dues_records WHERE member_id = $1 ORDER BY due_date DESC`, memberID)
	if err != nil {
		return nil, fmt.Errorf("list member dues: %w", err)
	}
	defer rows.Close()

	records, _, err := scanDuesRecords(rows, 0)
	return records, err
}

func (s *duesService) Create(ctx context.Context, chapterID, memberID, semester string, amountCents int, dueDate string) (*DuesRecord, error) {
	var id string
	err := s.db.QueryRow(ctx, `
		INSERT INTO dues_records (chapter_id, member_id, semester, amount_cents, due_date)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (chapter_id, member_id, semester)
		DO UPDATE SET amount_cents = $4, due_date = $5, updated_at = NOW()
		RETURNING id`,
		chapterID, memberID, semester, amountCents, dueDate,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create dues record: %w", err)
	}

	return s.getByID(ctx, id)
}

func (s *duesService) MarkPaid(ctx context.Context, duesID, paymentMethod string) (*DuesRecord, error) {
	now := time.Now()
	_, err := s.db.Exec(ctx, `
		UPDATE dues_records
		SET status = 'paid', paid_at = $2, payment_method = $3, updated_at = NOW()
		WHERE id = $1`,
		duesID, now, paymentMethod)
	if err != nil {
		return nil, fmt.Errorf("mark dues paid: %w", err)
	}

	rec, err := s.getByID(ctx, duesID)
	if err != nil {
		return nil, err
	}

	// Update member dues_status
	s.db.Exec(ctx,
		`UPDATE members SET dues_status = 'paid', updated_at = NOW() WHERE id = $1`,
		rec.MemberID)

	return rec, nil
}

// HandleZeffyWebhook processes incoming Zeffy payment confirmations sent via Zapier.
// Payload: { member_email, amount_cents, transaction_id, semester }
func (s *duesService) HandleZeffyWebhook(ctx context.Context, chapterID string, payload ZeffyWebhookPayload) error {
	// Look up member by email
	var memberID string
	err := s.db.QueryRow(ctx, `
		SELECT m.id FROM members m
		JOIN users u ON u.id = m.user_id
		WHERE u.email = $1 AND m.chapter_id = $2 AND m.deleted_at IS NULL`,
		payload.MemberEmail, chapterID,
	).Scan(&memberID)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("member not found for email: %s", payload.MemberEmail)
	}
	if err != nil {
		return fmt.Errorf("lookup member: %w", err)
	}

	// Find the dues record
	var duesID string
	var xpAwarded bool
	var xpAmount int
	err = s.db.QueryRow(ctx, `
		SELECT id, xp_awarded,
		       COALESCE((SELECT xp FROM point_economy WHERE chapter_id = $3 AND activity = 'Dues Paid' LIMIT 1), 100)
		FROM dues_records
		WHERE chapter_id = $3 AND member_id = $1 AND semester = $2 AND status != 'paid'`,
		memberID, payload.Semester, chapterID,
	).Scan(&duesID, &xpAwarded, &xpAmount)

	if err == pgx.ErrNoRows {
		// No open dues record — create one and mark it paid
		duesID = ""
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin zeffy tx: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	if duesID == "" {
		err = tx.QueryRow(ctx, `
			INSERT INTO dues_records (chapter_id, member_id, semester, amount_cents, due_date, status, paid_at, payment_method, zeffy_transaction_id)
			VALUES ($1, $2, $3, $4, CURRENT_DATE, 'paid', $5, 'zeffy', $6)
			RETURNING id`,
			chapterID, memberID, payload.Semester, payload.AmountCents, now, payload.TransactionID,
		).Scan(&duesID)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE dues_records
			SET status = 'paid', paid_at = $2, payment_method = 'zeffy',
			    zeffy_transaction_id = $3, updated_at = NOW()
			WHERE id = $1`,
			duesID, now, payload.TransactionID)
	}
	if err != nil {
		return fmt.Errorf("update dues record: %w", err)
	}

	// Award XP if not already done
	if !xpAwarded && xpAmount > 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO engagement_log (chapter_id, member_id, activity, xp_awarded, source, reference_id, reference_type)
			VALUES ($1, $2, 'Dues Paid', $3, 'dues', $4, 'dues_record')`,
			chapterID, memberID, xpAmount, duesID)
		if err != nil {
			return fmt.Errorf("insert xp log: %w", err)
		}

		_, err = tx.Exec(ctx, `UPDATE members SET xp_total = xp_total + $1, xp_semester = xp_semester + $1, dues_status = 'paid', updated_at = NOW() WHERE id = $2`,
			xpAmount, memberID)
		if err != nil {
			return fmt.Errorf("update member xp: %w", err)
		}

		_, err = tx.Exec(ctx, `UPDATE dues_records SET xp_awarded = true WHERE id = $1`, duesID)
		if err != nil {
			return fmt.Errorf("mark xp awarded: %w", err)
		}
	} else {
		// Just update dues_status without XP
		_, _ = tx.Exec(ctx, `UPDATE members SET dues_status = 'paid', updated_at = NOW() WHERE id = $1`, memberID)
	}

	return tx.Commit(ctx)
}

// HandleStripeSubscriptionWebhook processes Stripe billing events for chapter subscriptions.
func (s *duesService) HandleStripeSubscriptionWebhook(ctx context.Context, event *stripe.Event) error {
	// Parse the subscription object from the event data
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("unmarshal subscription: %w", err)
	}

	newStatus := string(sub.Status)

	switch event.Type {
	case "customer.subscription.updated":
		_, err := s.db.Exec(ctx,
			`UPDATE chapters SET subscription_status = $1, stripe_subscription_id = $2, updated_at = NOW()
			 WHERE stripe_customer_id = $3`,
			newStatus, sub.ID, sub.Customer.ID)
		return err

	case "customer.subscription.deleted":
		_, err := s.db.Exec(ctx,
			`UPDATE chapters SET subscription_status = 'canceled', updated_at = NOW()
			 WHERE stripe_customer_id = $1`,
			sub.Customer.ID)
		return err

	case "invoice.payment_succeeded":
		_, err := s.db.Exec(ctx,
			`UPDATE chapters SET subscription_status = 'active', updated_at = NOW()
			 WHERE stripe_customer_id = $1`,
			sub.Customer.ID)
		return err

	case "invoice.payment_failed":
		_, err := s.db.Exec(ctx,
			`UPDATE chapters SET subscription_status = 'past_due', updated_at = NOW()
			 WHERE stripe_customer_id = $1`,
			sub.Customer.ID)
		return err
	}

	return nil
}

func (s *duesService) getByID(ctx context.Context, id string) (*DuesRecord, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, chapter_id, member_id, semester, amount_cents, due_date,
		       paid_at, payment_method, zeffy_form_id, zeffy_transaction_id,
		       status, xp_awarded, created_at, updated_at
		FROM dues_records WHERE id = $1`, id)

	var r DuesRecord
	err := row.Scan(&r.ID, &r.ChapterID, &r.MemberID, &r.Semester, &r.AmountCents,
		&r.DueDate, &r.PaidAt, &r.PaymentMethod, &r.ZeffyFormID, &r.ZeffyTransactionID,
		&r.Status, &r.XPAwarded, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get dues record: %w", err)
	}
	return &r, nil
}

func scanDuesRecords(rows pgx.Rows, total int) ([]*DuesRecord, int, error) {
	var records []*DuesRecord
	for rows.Next() {
		var r DuesRecord
		err := rows.Scan(&r.ID, &r.ChapterID, &r.MemberID, &r.Semester, &r.AmountCents,
			&r.DueDate, &r.PaidAt, &r.PaymentMethod, &r.ZeffyFormID, &r.ZeffyTransactionID,
			&r.Status, &r.XPAwarded, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan dues record: %w", err)
		}
		records = append(records, &r)
	}
	return records, total, nil
}
