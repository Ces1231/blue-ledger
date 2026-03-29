package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// Seed script — imports prototype data into the database for the founding chapter.
//
// Data seeded (from app/index.html prototype constants):
//   - Chapter:        Tau Sigma Sigma (ΤΣΣ), Atlanta GA
//   - Members:        7 active members + CES1231 (sysadmin)
//   - Badges:         10 definitions + member_badge awards
//   - Point Economy:  10 activity XP values
//   - Events:         4 events + RSVPs
//   - Announcements:  4 chapter posts
//   - Service Log:    6 service entries
//   - Props:          4 peer-recognition entries
//   - Engagement Log: 10 XP audit entries
//   - Store Items:    8 XP redemption items
//   - Fundraising:    3 campaigns
//   - Mentorships:    4 pairs
//   - Minutes:        2 meeting minutes sets
//   - Scholarships:   4 applications
//   - Quests:         4 multi-step quest definitions
//
// Usage:
//   go run ./cmd/seed
//   DATABASE_URL=postgres://... go run ./cmd/seed
//
// Idempotent: re-running on an already-seeded chapter is safe (skipped).

const seedPassword = "BlueLedger2026!"

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ping: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Blue Ledger — Prototype Data Seed")
	fmt.Println("===================================")

	if err := run(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\n✅ Seed complete!")
}

// ptr helpers
func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func run(ctx context.Context, pool *pgxpool.Pool) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	passwordHash := string(hash)

	// ── 1. Chapter ────────────────────────────────────────────────────────
	fmt.Print("  chapter... ")
	var chapterID string
	err = pool.QueryRow(ctx,
		`SELECT id FROM chapters WHERE greek_letters = 'ΤΣΣ' LIMIT 1`,
	).Scan(&chapterID)
	if err == pgx.ErrNoRows {
		err = pool.QueryRow(ctx, `
			INSERT INTO chapters
				(name, greek_letters, city, state_code, member_id_prefix,
				 subscription_status, plan_tier, semester_start)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
			"Tau Sigma Sigma Chapter", "ΤΣΣ", "Atlanta", "GA", "ΤΣΣ",
			"active", "chapter_pro", "2025-09-01",
		).Scan(&chapterID)
	}
	if err != nil {
		return fmt.Errorf("chapter: %w", err)
	}
	fmt.Printf("ok (id=%s)\n", chapterID)

	// ── Guard: already seeded? ────────────────────────────────────────────
	var seededCount int
	_ = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM members WHERE chapter_id=$1 AND display_id LIKE 'ΤΣΣ-%'`,
		chapterID,
	).Scan(&seededCount)
	if seededCount >= 7 {
		fmt.Println("  ↩  Chapter already has prototype members — skipping member-dependent seed.")
		fmt.Println("     Delete chapter row and members to re-seed from scratch.")
		return nil
	}

	// ── 2. Users ──────────────────────────────────────────────────────────
	fmt.Print("  users... ")
	type userSeed struct {
		email     string
		firstName string
		lastName  string
		isSys     bool
	}
	userSeeds := []userSeed{
		{"m.williams@chapter.org", "Marcus", "Williams", false},
		{"d.carter@chapter.org", "DeShawn", "Carter", false},
		{"e.brooks@chapter.org", "Elijah", "Brooks", false},
		{"j.hayes@chapter.org", "Jordan", "Hayes", false},
		{"i.ford@chapter.org", "Isaiah", "Ford", false},
		{"t.simmons@chapter.org", "Terrence", "Simmons", false},
		{"b.mitchell@chapter.org", "Brandon", "Mitchell", false},
		{"ces1231@blueledger.sys", "CES", "1231", true},
	}
	userIDs := make(map[string]string) // email → uuid
	for _, u := range userSeeds {
		var uid string
		err = pool.QueryRow(ctx, `
			INSERT INTO users (email, first_name, last_name, password_hash, email_verified, is_sysadmin)
			VALUES ($1,$2,$3,$4,TRUE,$5)
			ON CONFLICT (email) DO UPDATE SET first_name = EXCLUDED.first_name
			RETURNING id`,
			u.email, u.firstName, u.lastName, passwordHash, u.isSys,
		).Scan(&uid)
		if err != nil {
			return fmt.Errorf("user %s: %w", u.email, err)
		}
		userIDs[u.email] = uid
	}
	fmt.Printf("ok (%d users)\n", len(userIDs))

	// ── 3. Members ────────────────────────────────────────────────────────
	fmt.Print("  members... ")
	type memberSeed struct {
		displayID string
		email     string
		name      string
		initials  string
		role      string
		inducted  int
		employer  string
		title     string
		city      string
		linkedin  string
		xp        int
		level     string // level_key: neo/scholar/leader/sage/legend/icon
		dues      string // dues_status: paid/unpaid/late/waived/outstanding
		status    string // status: active/inactive/alumni/suspended/pledging
		avatarBg  string
		avatarFg  string
	}
	memberSeeds := []memberSeed{
		{"ΤΣΣ-001", "m.williams@chapter.org", "Marcus J. Williams", "MJ", "admin", 2019, "Deloitte", "Senior Analyst", "Atlanta, GA", "linkedin.com/in/mjwilliams", 1840, "legend", "paid", "active", "#C9A84C", "#001A4D"},
		{"ΤΣΣ-002", "d.carter@chapter.org", "DeShawn A. Carter", "DA", "chair", 2020, "JP Morgan", "Associate", "New York, NY", "linkedin.com/in/dacarter", 1250, "sage", "paid", "active", "#888888", "#ffffff"},
		{"ΤΣΣ-003", "e.brooks@chapter.org", "Elijah T. Brooks", "ET", "pia", 2021, "Howard University", "Graduate Student", "Washington, DC", "", 780, "leader", "paid", "active", "#CD7F32", "#ffffff"},
		{"ΤΣΣ-004", "j.hayes@chapter.org", "Jordan M. Hayes", "JM", "member", 2022, "City of Atlanta", "Project Coordinator", "Atlanta, GA", "", 340, "scholar", "paid", "active", "#185FA5", "#ffffff"},
		{"ΤΣΣ-005", "i.ford@chapter.org", "Isaiah R. Ford", "IR", "member", 2023, "Google", "Software Engineer", "San Francisco, CA", "linkedin.com/in/irford", 210, "neo", "late", "active", "#185FA5", "#ffffff"},
		{"ΤΣΣ-006", "t.simmons@chapter.org", "Terrence K. Simmons", "TS", "member", 2020, "Delta Air Lines", "Operations Analyst", "Atlanta, GA", "", 920, "leader", "paid", "active", "#888888", "#ffffff"},
		{"ΤΣΣ-007", "b.mitchell@chapter.org", "Brandon L. Mitchell", "BM", "member", 2021, "CDC", "Public Health Analyst", "Atlanta, GA", "", 650, "leader", "outstanding", "active", "#CD7F32", "#ffffff"},
		{"CES-001", "ces1231@blueledger.sys", "CES1231", "CE", "sysadmin", 2024, "Blue Ledger Systems", "System Administrator", "Atlanta, GA", "", 9999, "icon", "paid", "active", "#0D1117", "#58A6FF"},
	}
	memberIDs := make(map[string]string) // displayID → uuid
	for _, ms := range memberSeeds {
		var linkedinArg *string
		if ms.linkedin != "" {
			linkedinArg = strPtr(ms.linkedin)
		}
		var mid string
		err = pool.QueryRow(ctx, `
			INSERT INTO members
				(chapter_id, user_id, display_id, name, initials, email,
				 role, inducted_year, status, employer, title, city, linkedin,
				 xp_total, xp_semester, level_key, dues_status, avatar_bg, avatar_fg)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
			ON CONFLICT (chapter_id, display_id) DO UPDATE SET xp_total = EXCLUDED.xp_total
			RETURNING id`,
			chapterID, userIDs[ms.email], ms.displayID, ms.name, ms.initials, ms.email,
			ms.role, ms.inducted, ms.status, ms.employer, ms.title, ms.city, linkedinArg,
			ms.xp, ms.xp/4, ms.level, ms.dues, ms.avatarBg, ms.avatarFg,
		).Scan(&mid)
		if err != nil {
			return fmt.Errorf("member %s: %w", ms.displayID, err)
		}
		memberIDs[ms.displayID] = mid
	}
	fmt.Printf("ok (%d members)\n", len(memberIDs))
	m := func(did string) string { return memberIDs[did] }

	// ── 4. Badges ─────────────────────────────────────────────────────────
	fmt.Print("  badges... ")
	type badgeSeed struct {
		name     string
		icon     string
		category string
		req      string
		xp       int
		rarity   string
	}
	badgeSeeds := []badgeSeed{
		{"Architect", "🧩", "Onboarding", "Complete 100% of your profile", 100, "common"},
		{"Scholar", "📚", "Knowledge", "Score 100% on any quiz", 150, "uncommon"},
		{"Faithful", "🛡️", "Financials", "Pay dues on time 3 semesters in a row", 300, "rare"},
		{"Punctual", "⏱️", "Attendance", "Attend 10 meetings on time", 200, "uncommon"},
		{"25-Hour Builder", "🏗️", "Service", "Log 25 hours of community service", 250, "common"},
		{"The Centurion", "💯", "Service", "Log 100 hours of community service", 1000, "legendary"},
		{"Road Warrior", "🌍", "Attendance", "Attend a Regional or National Conference", 300, "rare"},
		{"Sigma of the Year", "👑", "All", "Win annual Chapter vote", 500, "legendary"},
		{"The Plug", "🔌", "Social", "Refer a brother who gets initiated", 200, "rare"},
		{"Chapter Legend", "🏛️", "All", "Reach 2,500+ XP lifetime", 500, "legendary"},
	}
	badgeIDs := make(map[string]string) // name → uuid
	for _, b := range badgeSeeds {
		var bid string
		err = pool.QueryRow(ctx, `
			INSERT INTO badges (chapter_id, name, icon, category, description, requirement, xp_reward, rarity)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (chapter_id, name) DO UPDATE SET icon = EXCLUDED.icon
			RETURNING id`,
			chapterID, b.name, b.icon, b.category, b.req, b.req, b.xp, b.rarity,
		).Scan(&bid)
		if err != nil {
			return fmt.Errorf("badge %s: %w", b.name, err)
		}
		badgeIDs[b.name] = bid
	}
	fmt.Printf("ok (%d badges)\n", len(badgeIDs))

	// ── 5. Member Badges ──────────────────────────────────────────────────
	fmt.Print("  member badges... ")
	type memberBadge struct{ did, badge string }
	memberBadges := []memberBadge{
		{"ΤΣΣ-001", "Architect"}, {"ΤΣΣ-001", "Scholar"}, {"ΤΣΣ-001", "Faithful"},
		{"ΤΣΣ-001", "Punctual"}, {"ΤΣΣ-001", "25-Hour Builder"},
		{"ΤΣΣ-002", "Architect"}, {"ΤΣΣ-002", "Road Warrior"}, {"ΤΣΣ-002", "Punctual"},
		{"ΤΣΣ-003", "Architect"}, {"ΤΣΣ-003", "Scholar"},
		{"ΤΣΣ-004", "Architect"},
		{"ΤΣΣ-006", "Architect"}, {"ΤΣΣ-006", "Faithful"},
		{"ΤΣΣ-007", "Architect"},
		{"CES-001", "Architect"},
	}
	for _, mb := range memberBadges {
		_, err = pool.Exec(ctx, `
			INSERT INTO member_badges (chapter_id, member_id, badge_id)
			VALUES ($1,$2,$3)
			ON CONFLICT (member_id, badge_id) DO NOTHING`,
			chapterID, m(mb.did), badgeIDs[mb.badge],
		)
		if err != nil {
			return fmt.Errorf("member badge %s/%s: %w", mb.did, mb.badge, err)
		}
	}
	fmt.Printf("ok (%d awarded)\n", len(memberBadges))

	// ── 6. Point Economy ──────────────────────────────────────────────────
	fmt.Print("  point economy... ")
	type peSeed struct {
		activity string
		xp       int
		category string
	}
	peSeeds := []peSeed{
		{"Chapter Meeting (on time)", 50, "Attendance"},
		{"Chapter Meeting (late)", 25, "Attendance"},
		{"Committee Meeting", 20, "Attendance"},
		{"Regional/National Conference", 150, "Attendance"},
		{"Community Service (per hour)", 30, "Service"},
		{"Dues Paid (On Time)", 50, "Financials"},
		{"Committee Chair Role", 100, "Leadership"},
		{"E-Board Officer Role", 200, "Leadership"},
		{"Fraternity History Quiz (Pass)", 75, "Knowledge"},
		{"Constitution Quiz (Pass)", 75, "Knowledge"},
	}
	for _, pe := range peSeeds {
		_, err = pool.Exec(ctx, `
			INSERT INTO point_economy (chapter_id, activity, xp, category)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (chapter_id, activity) DO UPDATE SET xp = EXCLUDED.xp`,
			chapterID, pe.activity, pe.xp, pe.category,
		)
		if err != nil {
			return fmt.Errorf("point_economy %s: %w", pe.activity, err)
		}
	}
	fmt.Printf("ok (%d activities)\n", len(peSeeds))

	// ── 7. Events ─────────────────────────────────────────────────────────
	fmt.Print("  events... ")
	type eventSeed struct {
		name      string
		date      string
		etime     string
		location  string
		etype     string
		xp        int
		deadline  string
		createdBy string
	}
	eventSeeds := []eventSeed{
		{"Fall Chapter Kickoff Meeting", "2025-09-05", "7:00 PM", "Chapter Hall", "Meeting", 50, "2025-09-04", "ΤΣΣ-001"},
		{"Habitat for Humanity Build", "2025-09-07", "8:00 AM", "2301 MLK Dr", "Service", 120, "2025-09-05", "ΤΣΣ-001"},
		{"Youth Mentorship Workshop", "2025-09-14", "10:00 AM", "Boys & Girls Club", "Service", 90, "2025-09-12", "ΤΣΣ-002"},
		{"Fall Chapter Meeting #2", "2025-09-19", "7:00 PM", "Chapter Hall", "Meeting", 50, "2025-09-18", "ΤΣΣ-002"},
	}
	eventIDs := make(map[string]string) // name → uuid
	for _, ev := range eventSeeds {
		var eid string
		_ = pool.QueryRow(ctx,
			`SELECT id FROM events WHERE chapter_id=$1 AND name=$2 LIMIT 1`,
			chapterID, ev.name,
		).Scan(&eid)
		if eid == "" {
			err = pool.QueryRow(ctx, `
				INSERT INTO events
					(chapter_id, name, event_date, event_time, location,
					 event_type, xp_attend, rsvp_deadline, created_by)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
				chapterID, ev.name, ev.date, ev.etime, ev.location,
				ev.etype, ev.xp, ev.deadline, m(ev.createdBy),
			).Scan(&eid)
			if err != nil {
				return fmt.Errorf("event %s: %w", ev.name, err)
			}
		}
		eventIDs[ev.name] = eid
	}
	fmt.Printf("ok (%d events)\n", len(eventIDs))

	// ── 8. RSVPs ──────────────────────────────────────────────────────────
	fmt.Print("  rsvps... ")
	type rsvpSeed struct {
		event    string
		did      string
		status   string
		attended bool
	}
	rsvpSeeds := []rsvpSeed{
		{"Fall Chapter Kickoff Meeting", "ΤΣΣ-001", "yes", true},
		{"Fall Chapter Kickoff Meeting", "ΤΣΣ-002", "yes", true},
		{"Fall Chapter Kickoff Meeting", "ΤΣΣ-003", "yes", true},
		{"Fall Chapter Kickoff Meeting", "ΤΣΣ-004", "yes", true},
		{"Fall Chapter Kickoff Meeting", "ΤΣΣ-005", "maybe", false},
		{"Fall Chapter Kickoff Meeting", "ΤΣΣ-006", "yes", true},
		{"Habitat for Humanity Build", "ΤΣΣ-001", "yes", true},
		{"Habitat for Humanity Build", "ΤΣΣ-003", "yes", true},
		{"Habitat for Humanity Build", "ΤΣΣ-006", "yes", true},
		{"Youth Mentorship Workshop", "ΤΣΣ-001", "yes", true},
		{"Youth Mentorship Workshop", "ΤΣΣ-002", "yes", true},
		{"Youth Mentorship Workshop", "ΤΣΣ-004", "yes", true},
	}
	for _, r := range rsvpSeeds {
		eid := eventIDs[r.event]
		if eid == "" {
			continue
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO rsvps (chapter_id, event_id, member_id, status, attended)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (event_id, member_id) DO NOTHING`,
			chapterID, eid, m(r.did), r.status, r.attended,
		)
		if err != nil {
			return fmt.Errorf("rsvp %s/%s: %w", r.event, r.did, err)
		}
	}
	fmt.Printf("ok (%d rsvps)\n", len(rsvpSeeds))

	// ── 9. Announcements ──────────────────────────────────────────────────
	fmt.Print("  announcements... ")
	type annSeed struct {
		title    string
		body     string
		category string
		pinned   bool
		author   string
	}
	annSeeds := []annSeed{
		{
			"The Blue Ledger is LIVE 🚀",
			"Welcome to the chapter engagement platform. Complete your profile now to earn the Architect badge and 100 XP. The leaderboard is live — check your rank!",
			"General", true, "ΤΣΣ-001",
		},
		{
			"Fall 2025 Dues Deadline — Sept 10",
			"Chapter dues of $85 are due by September 10th. Pay on time to earn +50 XP. Contact the Treasurer with any questions.",
			"Dues", true, "ΤΣΣ-001",
		},
		{
			"Habitat for Humanity Build — Sept 7",
			"We're building with Habitat this Sunday. Meet at the chapter hall at 7:45 AM. 4 hours of service = 120 XP. RSVP in the app!",
			"Event", false, "ΤΣΣ-002",
		},
		{
			"Congrats Marcus — Gold Legend! 🏆",
			"Brother Marcus Williams crossed 1,500 XP and reached Gold Legend status. First in the chapter to unlock the blazer. Keep grinding, brothers.",
			"General", false, "ΤΣΣ-001",
		},
	}
	for _, a := range annSeeds {
		var exists bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM announcements WHERE chapter_id=$1 AND title=$2)`,
			chapterID, a.title,
		).Scan(&exists)
		if !exists {
			_, err = pool.Exec(ctx, `
				INSERT INTO announcements (chapter_id, title, body, category, is_pinned, author_id)
				VALUES ($1,$2,$3,$4,$5,$6)`,
				chapterID, a.title, a.body, a.category, a.pinned, m(a.author),
			)
			if err != nil {
				return fmt.Errorf("announcement %q: %w", a.title, err)
			}
		}
	}
	fmt.Printf("ok (%d announcements)\n", len(annSeeds))

	// ── 10. Service Log ───────────────────────────────────────────────────
	fmt.Print("  service log... ")
	type svcSeed struct {
		did        string
		eventName  string
		org        string
		date       string
		hours      float64
		xp         int
		verified   bool
		verifiedBy string
	}
	svcSeeds := []svcSeed{
		{"ΤΣΣ-001", "Habitat for Humanity Build", "Habitat for Humanity", "2025-09-07", 4, 120, true, "ΤΣΣ-002"},
		{"ΤΣΣ-003", "Habitat for Humanity Build", "Habitat for Humanity", "2025-09-07", 4, 120, true, "ΤΣΣ-002"},
		{"ΤΣΣ-001", "Youth Mentorship Workshop", "Boys & Girls Club", "2025-09-14", 3, 90, true, "ΤΣΣ-002"},
		{"ΤΣΣ-002", "Youth Mentorship Workshop", "Boys & Girls Club", "2025-09-14", 3, 90, true, "ΤΣΣ-001"},
		{"ΤΣΣ-004", "Food Bank Volunteer", "Atlanta Community Food Bank", "2025-09-21", 2, 60, true, "ΤΣΣ-001"},
		{"ΤΣΣ-005", "STEM Tutoring", "City Schools", "2025-09-28", 2, 0, false, ""},
	}
	for _, s := range svcSeeds {
		var vbID *string
		if s.verifiedBy != "" {
			vbID = strPtr(m(s.verifiedBy))
		}
		var verifiedAt *time.Time
		if s.verified {
			t := time.Now()
			verifiedAt = &t
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO service_log
				(chapter_id, member_id, event_name, organization, service_date,
				 hours, xp_awarded, verified, verified_by, verified_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			chapterID, m(s.did), s.eventName, s.org, s.date,
			s.hours, s.xp, s.verified, vbID, verifiedAt,
		)
		if err != nil {
			return fmt.Errorf("service log %s/%s: %w", s.did, s.eventName, err)
		}
	}
	fmt.Printf("ok (%d entries)\n", len(svcSeeds))

	// ── 11. Props ─────────────────────────────────────────────────────────
	fmt.Print("  props... ")
	type propSeed struct {
		from, to string
		category string
		message  string
	}
	propSeeds := []propSeed{
		{"ΤΣΣ-002", "ΤΣΣ-001", "Leadership", "Led the Habitat build with total professionalism. Real chapter leader."},
		{"ΤΣΣ-001", "ΤΣΣ-003", "Academic", "Elijah's quiz score was perfect — sets the standard."},
		{"ΤΣΣ-004", "ΤΣΣ-002", "Brotherhood", "Always looks out for the younger brothers. True Sigma."},
		{"ΤΣΣ-003", "ΤΣΣ-001", "Service", "Showed up every single service event this semester."},
	}
	for _, p := range propSeeds {
		_, err = pool.Exec(ctx, `
			INSERT INTO props (chapter_id, from_id, to_id, category, message, xp_awarded)
			VALUES ($1,$2,$3,$4,$5,10)`,
			chapterID, m(p.from), m(p.to), p.category, p.message,
		)
		if err != nil {
			return fmt.Errorf("props %s→%s: %w", p.from, p.to, err)
		}
	}
	fmt.Printf("ok (%d props)\n", len(propSeeds))

	// ── 12. Engagement Log ────────────────────────────────────────────────
	fmt.Print("  engagement log... ")
	type engSeed struct {
		did      string
		activity string
		xp       int
		source   string
		date     string
	}
	engSeeds := []engSeed{
		{"ΤΣΣ-001", "Chapter Meeting (on time)", 50, "checkin", "2025-09-05"},
		{"ΤΣΣ-002", "Chapter Meeting (on time)", 50, "checkin", "2025-09-05"},
		{"ΤΣΣ-003", "Chapter Meeting (on time)", 50, "checkin", "2025-09-05"},
		{"ΤΣΣ-004", "Chapter Meeting (on time)", 50, "checkin", "2025-09-05"},
		{"ΤΣΣ-001", "Community Service (per hour)", 120, "service", "2025-09-07"},
		{"ΤΣΣ-001", "Dues Paid (On Time)", 50, "dues", "2025-09-10"},
		{"ΤΣΣ-005", "Fraternity History Quiz (Pass)", 75, "quiz", "2025-09-12"},
		{"ΤΣΣ-003", "Community Service (per hour)", 90, "service", "2025-09-14"},
		{"ΤΣΣ-002", "Committee Chair Role", 100, "admin", "2025-09-15"},
		{"ΤΣΣ-004", "Chapter Meeting (on time)", 50, "checkin", "2025-09-19"},
	}
	for _, e := range engSeeds {
		ts, _ := time.Parse("2006-01-02", e.date)
		_, err = pool.Exec(ctx, `
			INSERT INTO engagement_log
				(chapter_id, member_id, activity, xp_awarded, source, semester, created_at)
			VALUES ($1,$2,$3,$4,$5,'Fall 2025',$6)`,
			chapterID, m(e.did), e.activity, e.xp, e.source, ts,
		)
		if err != nil {
			return fmt.Errorf("engagement log %s/%s: %w", e.did, e.activity, err)
		}
	}
	fmt.Printf("ok (%d entries)\n", len(engSeeds))

	// ── 13. Store Items ───────────────────────────────────────────────────
	fmt.Print("  store items... ")
	type storeSeed struct {
		name     string
		desc     string
		cost     int
		icon     string
		category string
		stock    int
	}
	storeSeeds := []storeSeed{
		{"Chapter Polo Shirt", "Royal blue Tau Sigma Sigma embroidered polo. Available sizes S–XXL.", 500, "👕", "Apparel", 12},
		{"Gold Chapter Pin", "Official chapter lapel pin. Gold plated.", 300, "📌", "Accessories", 20},
		{"Chapter Blazer", "Navy blazer with embroidered chapter crest. Premium quality.", 1500, "🧥", "Apparel", 5},
		{"Meeting Excusal (1x)", "One excused absence from a chapter meeting. Valid one semester.", 200, "📋", "Privileges", 999},
		{"Early Dues Window (+3 days)", "Get a 3-day extension on your dues deadline. One-time use.", 150, "⏰", "Privileges", 999},
		{"Chapter Cap", "Official navy Tau Sigma Sigma snapback cap with gold embroidery.", 400, "🧢", "Apparel", 8},
		{"Custom Profile Frame", "Unlock a limited-edition gold profile frame for your avatar.", 250, "🖼️", "Digital", 999},
		{"Founder's Edition Badge", "Exclusive digital badge for early adopters. Never available again.", 1000, "🏅", "Digital", 10},
	}
	storeAdded := 0
	for _, s := range storeSeeds {
		var exists bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM store_items WHERE chapter_id=$1 AND name=$2)`,
			chapterID, s.name,
		).Scan(&exists)
		if !exists {
			_, err = pool.Exec(ctx, `
				INSERT INTO store_items (chapter_id, name, description, xp_cost, icon, category, stock)
				VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				chapterID, s.name, s.desc, s.cost, s.icon, s.category, s.stock,
			)
			if err != nil {
				return fmt.Errorf("store item %s: %w", s.name, err)
			}
			storeAdded++
		}
	}
	fmt.Printf("ok (%d added)\n", storeAdded)

	// ── 14. Fundraising Campaigns ─────────────────────────────────────────
	fmt.Print("  fundraising... ")
	type fundSeed struct {
		title        string
		desc         string
		goalCents    int
		currentCents int
		category     string
		deadline     string
		createdBy    string
	}
	fundSeeds := []fundSeed{
		{"Annual Scholarship Fund", "Annual scholarship for a graduating high school senior in our community.", 500000, 320000, "Scholarship", "2025-12-31", "ΤΣΣ-001"},
		{"Community Health Fair Expenses", "Supplies and equipment for our annual community health screening event.", 80000, 65000, "Community", "2025-10-05", "ΤΣΣ-002"},
		{"Chapter Hall Renovation", "Modernizing the chapter hall — new furniture, AV equipment, HVAC.", 1000000, 240000, "Chapter", "2026-06-01", "ΤΣΣ-001"},
	}
	for _, f := range fundSeeds {
		var exists bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM fundraising_campaigns WHERE chapter_id=$1 AND title=$2)`,
			chapterID, f.title,
		).Scan(&exists)
		if !exists {
			_, err = pool.Exec(ctx, `
				INSERT INTO fundraising_campaigns
					(chapter_id, title, description, goal_cents, current_cents, category, deadline, created_by)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
				chapterID, f.title, f.desc, f.goalCents, f.currentCents, f.category, f.deadline, m(f.createdBy),
			)
			if err != nil {
				return fmt.Errorf("fundraising %q: %w", f.title, err)
			}
		}
	}
	fmt.Printf("ok (%d campaigns)\n", len(fundSeeds))

	// ── 15. Mentorships ───────────────────────────────────────────────────
	fmt.Print("  mentorships... ")
	type mentorSeed struct {
		mentorDID string
		menteeDID string
		focus     []string
	}
	mentorSeeds := []mentorSeed{
		{"ΤΣΣ-001", "ΤΣΣ-004", []string{"Career Development", "Leadership"}},
		{"ΤΣΣ-001", "ΤΣΣ-005", []string{"Career Development", "Technology"}},
		{"ΤΣΣ-002", "ΤΣΣ-007", []string{"Finance", "Professional Development"}},
		{"ΤΣΣ-003", "ΤΣΣ-004", []string{"Academic Excellence", "Graduate School"}},
	}
	for _, ms := range mentorSeeds {
		var exists bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM mentorships WHERE mentor_id=$1 AND mentee_id=$2)`,
			m(ms.mentorDID), m(ms.menteeDID),
		).Scan(&exists)
		if !exists {
			_, err = pool.Exec(ctx, `
				INSERT INTO mentorships (chapter_id, mentor_id, mentee_id, focus_areas, is_active, started_at)
				VALUES ($1,$2,$3,$4,TRUE,NOW())`,
				chapterID, m(ms.mentorDID), m(ms.menteeDID), ms.focus,
			)
			if err != nil {
				return fmt.Errorf("mentorship %s→%s: %w", ms.mentorDID, ms.menteeDID, err)
			}
		}
	}
	fmt.Printf("ok (%d pairs)\n", len(mentorSeeds))

	// ── 16. Chapter Minutes ───────────────────────────────────────────────
	fmt.Print("  minutes... ")
	type minSeed struct {
		date     string
		title    string
		body     string
		recorder string
		quorum   bool
		status   string
	}
	minSeeds := []minSeed{
		{
			"2025-09-05",
			"Fall Chapter Kickoff Meeting — Sept 5, 2025",
			"## Call to Order\nChapter called to order at 7:08 PM by President Williams.\n\n## Roll Call\n22 brothers present. Quorum achieved.\n\n## Old Business\n- Dues collection reminder for Fall 2025 semester due September 10.\n\n## New Business\n- Blue Ledger platform launch announced. All members to complete profiles by September 12.\n- Habitat for Humanity build scheduled for September 7.\n- Youth Mentorship Workshop scheduled for September 14.\n\n## Announcements\n- Regional Conference registration deadline is October 15.\n\n## Adjournment\nMeeting adjourned at 8:42 PM.",
			"ΤΣΣ-003", true, "final",
		},
		{
			"2025-09-19",
			"Fall Chapter Meeting #2 — Sept 19, 2025",
			"## Call to Order\nChapter called to order at 7:05 PM by President Williams.\n\n## Roll Call\n18 brothers present. Quorum achieved.\n\n## Committee Reports\n- Programming (Carter): Fall social planned for October 18.\n- Community Service (Simmons): Habitat build recap — 10 volunteers, 40 hours logged.\n- Finance (Mitchell): Chapter balance at $4,230. Dues 85% collected.\n\n## New Business\n- Sigma of the Year nominations open November 1.\n- Study groups forming for CPA, Bar, and GRE prep.\n\n## Adjournment\nMeeting adjourned at 9:15 PM.",
			"ΤΣΣ-003", true, "draft",
		},
	}
	for _, mn := range minSeeds {
		var exists bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM chapter_minutes WHERE chapter_id=$1 AND meeting_date=$2)`,
			chapterID, mn.date,
		).Scan(&exists)
		if !exists {
			_, err = pool.Exec(ctx, `
				INSERT INTO chapter_minutes (chapter_id, meeting_date, title, body, recorder_id, quorum, status)
				VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				chapterID, mn.date, mn.title, mn.body, m(mn.recorder), mn.quorum, mn.status,
			)
			if err != nil {
				return fmt.Errorf("minutes %s: %w", mn.date, err)
			}
		}
	}
	fmt.Printf("ok (%d sets)\n", len(minSeeds))

	// ── 17. Scholarships ──────────────────────────────────────────────────
	fmt.Print("  scholarships... ")
	type scholSeed struct {
		name        string
		school      string
		gpa         string
		major       string
		year        int
		city        string
		semester    string
		amountCents int
		status      string
	}
	scholSeeds := []scholSeed{
		{"Darius J. Thompson", "Georgia Tech", "3.8", "Computer Science", 2026, "Atlanta, GA", "Spring 2026", 100000, "under_review"},
		{"Aaliyah C. Robinson", "Clark Atlanta University", "3.6", "Business Administration", 2026, "Atlanta, GA", "Spring 2026", 100000, "finalist"},
		{"Marcus K. Jefferson", "Morehouse College", "3.9", "Pre-Medicine", 2026, "Atlanta, GA", "Spring 2026", 150000, "submitted"},
		{"Jasmine R. Williams", "Spelman College", "3.7", "Psychology", 2026, "Atlanta, GA", "Spring 2026", 100000, "awarded"},
	}
	for _, s := range scholSeeds {
		var exists bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM scholarships WHERE chapter_id=$1 AND applicant_name=$2 AND semester=$3)`,
			chapterID, s.name, s.semester,
		).Scan(&exists)
		if !exists {
			_, err = pool.Exec(ctx, `
				INSERT INTO scholarships
					(chapter_id, applicant_name, school, gpa, major,
					 academic_year, city, semester, amount_cents, status, submitted_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,CURRENT_DATE)`,
				chapterID, s.name, s.school, s.gpa, s.major,
				s.year, s.city, s.semester, s.amountCents, s.status,
			)
			if err != nil {
				return fmt.Errorf("scholarship %s: %w", s.name, err)
			}
		}
	}
	fmt.Printf("ok (%d applications)\n", len(scholSeeds))

	// ── 18. Quests ────────────────────────────────────────────────────────
	fmt.Print("  quests... ")
	type questSeed struct {
		title string
		desc  string
		xp    int
		steps string
	}
	questSeeds := []questSeed{
		{
			"Welcome to the Chapter",
			"Complete your onboarding steps to earn your first badge and XP.",
			200,
			`[{"title":"Complete your profile","description":"Fill out all profile fields including employer, city, and LinkedIn","target_count":1},{"title":"Attend your first meeting","description":"Check in to any chapter meeting","target_count":1},{"title":"Give your first props","description":"Recognize a brother by sending props","target_count":1}]`,
		},
		{
			"Service Champion",
			"Log 10 hours of verified community service in a single semester.",
			300,
			`[{"title":"Log first service hours","description":"Submit your first service log entry","target_count":1},{"title":"Reach 5 service hours","description":"Log a cumulative 5 hours of verified service","target_count":5},{"title":"Reach 10 service hours","description":"Log a cumulative 10 hours of verified service","target_count":10}]`,
		},
		{
			"Knowledge Seeker",
			"Pass 3 chapter quizzes to prove your fraternity knowledge.",
			250,
			`[{"title":"Pass the History Quiz","description":"Score 100% on the Fraternity History quiz","target_count":1},{"title":"Pass the Constitution Quiz","description":"Score 100% on the Constitution quiz","target_count":1},{"title":"Pass a third quiz","description":"Score 100% on any remaining quiz","target_count":1}]`,
		},
		{
			"The Faithful",
			"Pay dues on time for 3 consecutive semesters.",
			400,
			`[{"title":"Pay dues on time (1st)","description":"Pay your chapter dues before the deadline","target_count":1},{"title":"Pay dues on time (2nd)","description":"Pay your chapter dues before the deadline again","target_count":1},{"title":"Pay dues on time (3rd)","description":"Complete the hat trick — dues paid on time 3 semesters in a row","target_count":1}]`,
		},
	}
	for _, q := range questSeeds {
		var exists bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM quests WHERE chapter_id=$1 AND title=$2)`,
			chapterID, q.title,
		).Scan(&exists)
		if !exists {
			_, err = pool.Exec(ctx, `
				INSERT INTO quests (chapter_id, title, description, xp_reward, steps, is_active)
				VALUES ($1,$2,$3,$4,$5::jsonb,TRUE)`,
				chapterID, q.title, q.desc, q.xp, q.steps,
			)
			if err != nil {
				return fmt.Errorf("quest %q: %w", q.title, err)
			}
		}
	}
	fmt.Printf("ok (%d quests)\n", len(questSeeds))

	return nil
}
