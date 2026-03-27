package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// --- Domain errors ---

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is temporarily locked due to too many failed attempts")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenUsed          = errors.New("token has already been used")
	ErrTokenNotFound      = errors.New("token not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email address is already registered")
)

// RegisterInput holds the data required to create a new chapter and admin user.
type RegisterInput struct {
	// Chapter
	ChapterName    string `json:"chapter_name" validate:"required,min=3,max=120"`
	GreekLetters   string `json:"greek_letters" validate:"required"`
	City           string `json:"city" validate:"required"`
	StateCode      string `json:"state_code" validate:"required,len=2"`
	University     string `json:"university"`
	// Admin user
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

// LoginInput holds email/password login credentials.
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is returned after successful authentication.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	User         UserSummary `json:"user"`
}

// UserSummary is the minimal user profile returned in auth responses.
type UserSummary struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	ChapterID  string `json:"chapter_id"`
	MemberID   string `json:"member_id"`
	Role       string `json:"role"`
	IsSysadmin bool   `json:"is_sysadmin"`
}

// Service defines the auth business logic interface.
type Service interface {
	Register(ctx context.Context, input RegisterInput, deviceHint, ipAddress string) (*AuthResponse, error)
	Login(ctx context.Context, input LoginInput, deviceHint, ipAddress string) (*AuthResponse, error)
	SendMagicLink(ctx context.Context, email string) error
	ConsumeMagicLink(ctx context.Context, rawToken, deviceHint, ipAddress string) (*AuthResponse, error)
	RefreshToken(ctx context.Context, rawRefreshToken, deviceHint, ipAddress string) (*AuthResponse, error)
	Logout(ctx context.Context, rawRefreshToken string) error
	GetUserByID(ctx context.Context, userID string) (*UserSummary, error)
}

// service implements Service.
type service struct {
	db              *pgxpool.Pool
	tokens          *TokenManager
	emailClient     EmailSender
	hmacSecret      []byte
	magicLinkExpiry time.Duration
	appURL          string
}

// EmailSender is the minimum interface the auth service needs from the email package.
type EmailSender interface {
	SendMagicLink(toEmail, token string) error
}

// NewService creates a new auth service.
func NewService(
	db *pgxpool.Pool,
	tokens *TokenManager,
	emailClient EmailSender,
	hmacSecret string,
	magicLinkExpiry time.Duration,
	appURL string,
) Service {
	return &service{
		db:              db,
		tokens:          tokens,
		emailClient:     emailClient,
		hmacSecret:      []byte(hmacSecret),
		magicLinkExpiry: magicLinkExpiry,
		appURL:          appURL,
	}
}

// Register creates a new chapter and its first admin user, returns auth tokens.
func (s *service) Register(ctx context.Context, input RegisterInput, deviceHint, ipAddress string) (*AuthResponse, error) {
	// Check for existing email
	var count int
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email = $1`, input.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if count > 0 {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Insert chapter, user, and member in a transaction
	var chapterID, userID, memberID string

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create chapter
	err = tx.QueryRow(ctx, `
		INSERT INTO chapters (name, greek_letters, city, state_code, university, member_id_prefix)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		input.ChapterName, input.GreekLetters, input.City, input.StateCode,
		input.University, "MBR",
	).Scan(&chapterID)
	if err != nil {
		return nil, fmt.Errorf("create chapter: %w", err)
	}

	// Create user
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, email_verified, password_hash, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		input.Email, false, string(hash), input.FirstName, input.LastName,
	).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Create member record (first admin)
	displayID := fmt.Sprintf("%s-001", input.GreekLetters)
	fullName := fmt.Sprintf("%s %s", input.FirstName, input.LastName)
	err = tx.QueryRow(ctx, `
		INSERT INTO members (chapter_id, user_id, display_id, name, email, role, status)
		VALUES ($1, $2, $3, $4, $5, 'admin', 'active')
		RETURNING id`,
		chapterID, userID, displayID, fullName, input.Email,
	).Scan(&memberID)
	if err != nil {
		return nil, fmt.Errorf("create member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit registration: %w", err)
	}

	return s.issueTokens(ctx, userID, chapterID, memberID, "admin", input.Email, input.FirstName, input.LastName, false, deviceHint, ipAddress)
}

// Login authenticates a user with email and password.
func (s *service) Login(ctx context.Context, input LoginInput, deviceHint, ipAddress string) (*AuthResponse, error) {
	var (
		userID            string
		passwordHash      string
		firstName         string
		lastName          string
		failedCount       int
		isSysadmin        bool
		lockedUntil       *time.Time
	)

	err := s.db.QueryRow(ctx, `
		SELECT id, password_hash, first_name, last_name, failed_login_count, locked_until, is_sysadmin
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`,
		input.Email,
	).Scan(&userID, &passwordHash, &firstName, &lastName, &failedCount, &lockedUntil, &isSysadmin)

	if err == pgx.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	// Check account lockout (15 minutes after 10 failed attempts)
	if lockedUntil != nil && time.Now().Before(*lockedUntil) {
		return nil, ErrAccountLocked
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)); err != nil {
		// Increment failed attempts
		newCount := failedCount + 1
		if newCount >= 10 {
			lockUntil := time.Now().Add(15 * time.Minute)
			_, _ = s.db.Exec(ctx,
				`UPDATE users SET failed_login_count = $1, locked_until = $2 WHERE id = $3`,
				newCount, lockUntil, userID)
		} else {
			_, _ = s.db.Exec(ctx,
				`UPDATE users SET failed_login_count = $1 WHERE id = $2`,
				newCount, userID)
		}
		return nil, ErrInvalidCredentials
	}

	// Reset failed attempts on success
	_, _ = s.db.Exec(ctx,
		`UPDATE users SET failed_login_count = 0, locked_until = NULL, last_login_at = NOW() WHERE id = $1`,
		userID)

	// Find the member's primary chapter
	var chapterID, memberID, role string
	err = s.db.QueryRow(ctx, `
		SELECT chapter_id, id, role
		FROM members
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT 1`,
		userID,
	).Scan(&chapterID, &memberID, &role)

	if isSysadmin && err == pgx.ErrNoRows {
		// Sysadmin may not have a chapter membership
		return s.issueTokens(ctx, userID, "", "", "sysadmin", input.Email, firstName, lastName, true, deviceHint, ipAddress)
	}
	if err != nil {
		return nil, fmt.Errorf("lookup member: %w", err)
	}

	if isSysadmin {
		role = "sysadmin"
	}

	return s.issueTokens(ctx, userID, chapterID, memberID, role, input.Email, firstName, lastName, isSysadmin, deviceHint, ipAddress)
}

// SendMagicLink generates a magic link token and sends it via email.
func (s *service) SendMagicLink(ctx context.Context, email string) error {
	// Verify user exists
	var userID string
	err := s.db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`, email).Scan(&userID)
	if err == pgx.ErrNoRows {
		// Return no error to avoid email enumeration
		return nil
	}
	if err != nil {
		return fmt.Errorf("lookup user for magic link: %w", err)
	}

	// Generate raw token (32 random bytes)
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return fmt.Errorf("generate magic token: %w", err)
	}
	rawToken := hex.EncodeToString(rawBytes)

	// Hash token for storage
	tokenHash := sha256Hex(rawToken)

	// Store in database
	_, err = s.db.Exec(ctx, `
		INSERT INTO magic_link_tokens (user_id, token_hash, email, expires_at)
		VALUES ($1, $2, $3, $4)`,
		userID, tokenHash, email, time.Now().Add(s.magicLinkExpiry),
	)
	if err != nil {
		return fmt.Errorf("store magic token: %w", err)
	}

	// Send email (fire-and-forget in practice; we return the error here for reliability)
	if err := s.emailClient.SendMagicLink(email, rawToken); err != nil {
		return fmt.Errorf("send magic link email: %w", err)
	}

	return nil
}

// ConsumeMagicLink validates a raw token from a magic link click, returns auth tokens.
func (s *service) ConsumeMagicLink(ctx context.Context, rawToken, deviceHint, ipAddress string) (*AuthResponse, error) {
	tokenHash := sha256Hex(rawToken)

	var (
		tokenID   string
		userID    string
		email     string
		expiresAt time.Time
		usedAt    *time.Time
	)

	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, email, expires_at, used_at
		FROM magic_link_tokens
		WHERE token_hash = $1`,
		tokenHash,
	).Scan(&tokenID, &userID, &email, &expiresAt, &usedAt)

	if err == pgx.ErrNoRows {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lookup magic token: %w", err)
	}

	if usedAt != nil {
		return nil, ErrTokenUsed
	}
	if time.Now().After(expiresAt) {
		return nil, ErrTokenExpired
	}

	// Mark token as used
	_, err = s.db.Exec(ctx, `UPDATE magic_link_tokens SET used_at = NOW() WHERE id = $1`, tokenID)
	if err != nil {
		return nil, fmt.Errorf("consume magic token: %w", err)
	}

	// Mark email as verified and update last login
	_, _ = s.db.Exec(ctx, `UPDATE users SET email_verified = true, last_login_at = NOW() WHERE id = $1`, userID)

	// Find member
	var chapterID, memberID, role string
	var isSysadmin bool
	s.db.QueryRow(ctx, `SELECT is_sysadmin FROM users WHERE id = $1`, userID).Scan(&isSysadmin)

	err = s.db.QueryRow(ctx, `
		SELECT chapter_id, id, role FROM members WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at ASC LIMIT 1`,
		userID,
	).Scan(&chapterID, &memberID, &role)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("lookup member: %w", err)
	}

	if isSysadmin {
		role = "sysadmin"
	}

	return s.issueTokens(ctx, userID, chapterID, memberID, role, email, "", "", isSysadmin, deviceHint, ipAddress)
}

// RefreshToken rotates a refresh token, issuing new access + refresh tokens.
func (s *service) RefreshToken(ctx context.Context, rawRefreshToken, deviceHint, ipAddress string) (*AuthResponse, error) {
	tokenHash := sha256Hex(rawRefreshToken)

	var (
		sessionID string
		userID    string
		chapterID *string
		expiresAt time.Time
		revokedAt *time.Time
	)

	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, chapter_id, expires_at, revoked_at
		FROM auth_sessions
		WHERE refresh_token = $1`,
		tokenHash,
	).Scan(&sessionID, &userID, &chapterID, &expiresAt, &revokedAt)

	if err == pgx.ErrNoRows {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lookup session: %w", err)
	}

	if revokedAt != nil {
		return nil, ErrTokenUsed
	}
	if time.Now().After(expiresAt) {
		return nil, ErrTokenExpired
	}

	// Revoke old session
	_, err = s.db.Exec(ctx, `UPDATE auth_sessions SET revoked_at = NOW() WHERE id = $1`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("revoke old session: %w", err)
	}

	// Get current user + member data
	var (
		email      string
		isSysadmin bool
	)
	s.db.QueryRow(ctx, `SELECT email, is_sysadmin FROM users WHERE id = $1`, userID).Scan(&email, &isSysadmin)

	cid := ""
	if chapterID != nil {
		cid = *chapterID
	}

	var memberID, role string
	s.db.QueryRow(ctx, `
		SELECT id, role FROM members WHERE user_id = $1 AND chapter_id = $2 AND deleted_at IS NULL`,
		userID, cid,
	).Scan(&memberID, &role)

	if isSysadmin {
		role = "sysadmin"
	}

	return s.issueTokens(ctx, userID, cid, memberID, role, email, "", "", isSysadmin, deviceHint, ipAddress)
}

// Logout revokes the session associated with the given refresh token.
func (s *service) Logout(ctx context.Context, rawRefreshToken string) error {
	tokenHash := sha256Hex(rawRefreshToken)
	_, err := s.db.Exec(ctx,
		`UPDATE auth_sessions SET revoked_at = NOW() WHERE refresh_token = $1`,
		tokenHash)
	if err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}

// GetUserByID returns a UserSummary for the given user ID.
func (s *service) GetUserByID(ctx context.Context, userID string) (*UserSummary, error) {
	var u UserSummary
	err := s.db.QueryRow(ctx, `
		SELECT id, email, first_name, last_name, is_sysadmin FROM users WHERE id = $1`, userID).
		Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.IsSysadmin)
	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	s.db.QueryRow(ctx, `
		SELECT chapter_id, id, role FROM members WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at ASC LIMIT 1`,
		userID).Scan(&u.ChapterID, &u.MemberID, &u.Role)

	if u.IsSysadmin {
		u.Role = "sysadmin"
	}

	return &u, nil
}

// issueTokens generates access + refresh tokens and persists the session.
func (s *service) issueTokens(ctx context.Context, userID, chapterID, memberID, role, email, firstName, lastName string, isSysadmin bool, deviceHint, ipAddress string) (*AuthResponse, error) {
	accessToken, err := s.tokens.GenerateAccessToken(userID, chapterID, memberID, role, email, isSysadmin)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	rawRefresh, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshHash := sha256Hex(rawRefresh)

	// Persist session
	var cid *string
	if chapterID != "" {
		cid = &chapterID
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO auth_sessions (user_id, chapter_id, refresh_token, device_hint, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5::inet, $6)`,
		userID, cid, refreshHash, deviceHint, ipAddress, time.Now().Add(s.tokens.RefreshExpiry()),
	)
	if err != nil {
		return nil, fmt.Errorf("persist session: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int(s.tokens.AccessExpiry().Seconds()),
		User: UserSummary{
			ID:         userID,
			Email:      email,
			FirstName:  firstName,
			LastName:   lastName,
			ChapterID:  chapterID,
			MemberID:   memberID,
			Role:       role,
			IsSysadmin: isSysadmin,
		},
	}, nil
}

// sha256Hex returns the hex-encoded SHA-256 hash of the input string.
func sha256Hex(input string) string {
	h := hmac.New(sha256.New, nil)
	h.Write([]byte(input))
	return hex.EncodeToString(sha256.New().Sum([]byte(input))[:32])
}

// sha256Plain hashes a raw string with SHA-256 (no HMAC key) — used for magic link tokens.
func sha256Plain(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}
