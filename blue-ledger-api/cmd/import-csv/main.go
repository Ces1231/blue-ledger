package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// CSV importer for Blue Ledger member roster
// Usage: go run ./cmd/import-csv --file=/path/to/roster.csv

type MemberRecord struct {
	PBS            string
	FirstName      string
	LastName       string
	PreferredName  string
	Email          string
	PhoneNumber    string
	TShirtSize     string
	PoloSize       string
	BlazerSize     string
	BirthdayMonth  string
	Sigmaversary   string
}

func main() {
	_ = godotenv.Load()

	csvFile := flag.String("file", "", "Path to CSV roster file")
	flag.Parse()

	if *csvFile == "" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/import-csv --file=/path/to/roster.csv")
		os.Exit(1)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Read CSV
	file, err := os.Open(*csvFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open CSV: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read CSV: %v\n", err)
		os.Exit(1)
	}

	if len(records) < 2 {
		fmt.Fprintln(os.Stderr, "CSV must have header + data rows")
		os.Exit(1)
	}

	// Parse records (skip header)
	members := []MemberRecord{}
	for _, row := range records[1:] {
		if len(row) < 5 {
			continue // Skip malformed rows
		}
		members = append(members, MemberRecord{
			PBS:            strings.TrimSpace(row[0]),
			FirstName:      strings.TrimSpace(row[1]),
			LastName:       strings.TrimSpace(row[2]),
			PreferredName:  strings.TrimSpace(row[3]),
			Email:          strings.TrimSpace(row[4]),
			PhoneNumber:    strings.TrimSpace(row[5]),
			TShirtSize:     strings.TrimSpace(row[6]),
			PoloSize:       strings.TrimSpace(row[7]),
			BlazerSize:     strings.TrimSpace(row[8]),
			BirthdayMonth:  strings.TrimSpace(row[9]),
			Sigmaversary:   strings.TrimSpace(row[10]),
		})
	}

	fmt.Printf("Loaded %d members from CSV\n", len(members))

	// Create chapter if not exists
	chapterID := uuid.New().String()
	err = pool.QueryRow(ctx, `
		INSERT INTO chapters (id, name, greek_letters, city, state_code, university, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT DO NOTHING
		RETURNING id
	`, chapterID, "Tau Sigma Sigma", "ΤΣΣ", "Atlanta", "GA", "Georgia Tech").Scan(&chapterID)
	if err != nil && err != pgx.ErrNoRows {
		fmt.Fprintf(os.Stderr, "Failed to create chapter: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Chapter ID: %s\n", chapterID)

	// Create admin user
	adminUserID := uuid.New().String()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("BlueLedger2026!"), bcrypt.DefaultCost)
	err = pool.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, is_sysadmin, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (email) DO NOTHING
		RETURNING id
	`, adminUserID, "admin@tausigmasigma.org", string(hashedPassword), "Admin", "User", true).Scan(&adminUserID)
	if err != nil && err != pgx.ErrNoRows {
		fmt.Fprintf(os.Stderr, "Failed to create admin user: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Admin user created: admin@tausigmasigma.org\n")

	// Import members
	importedCount := 0
	skippedCount := 0
	for _, m := range members {
		// Skip rows without names
		if m.FirstName == "" || m.LastName == "" {
			skippedCount++
			continue
		}

		// Generate email if missing
		email := m.Email
		if email == "" {
			email = fmt.Sprintf("%s.%s@member.tausigmasigma.org", strings.ToLower(m.FirstName), strings.ToLower(m.LastName))
		}

		// Create or get user
		var actualUserID string
		err := pool.QueryRow(ctx, `
			INSERT INTO users (id, email, first_name, last_name, created_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (email) DO UPDATE SET id = EXCLUDED.id
			RETURNING id
		`, uuid.New().String(), email, m.FirstName, m.LastName).Scan(&actualUserID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create user %s %s: %v\n", m.FirstName, m.LastName, err)
			continue
		}

		// Create member
		// Generate a sequential display_id
		displayID := fmt.Sprintf("ΤΣΣ-%03d", importedCount+1)
		displayName := m.PreferredName
		if displayName == "" {
			displayName = fmt.Sprintf("%s %s", m.FirstName, m.LastName)
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO members (id, chapter_id, user_id, display_id, name, email, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (chapter_id, user_id) DO NOTHING
		`, uuid.New().String(), chapterID, actualUserID, displayID, displayName, email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create member %s %s: %v\n", m.FirstName, m.LastName, err)
			continue
		}

		importedCount++
	}

	fmt.Printf("\n✅ Imported %d members\n", importedCount)
	fmt.Printf("⏭️  Skipped %d incomplete records\n", skippedCount)
	fmt.Printf("\n🔑 Credentials for testing:\n")
	fmt.Printf("   Email: admin@tausigmasigma.org\n")
	fmt.Printf("   Password: BlueLedger2026!\n")
}
