package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Embedded CSV data - hardcoded member roster
const rosterData = `PBS#,First Name,Last Name,Preferred Name,Email,Phone,T-Shirt,Polo,Blazer,Birthday,Sigmaversary
1001,George,Smith,George,george.smith@email.com,404-555-0101,L,L,40R,January 15,2020
1002,Terrell,Marshall,Terry,terrell.marshall@email.com,404-555-0102,M,M,38R,February 20,2021
1003,Adrian,Fuller,Adrian,adrian.fuller@email.com,404-555-0103,XL,XL,42R,March 10,2019
1004,Dr. Corico,McCray,Corico,,404-555-0104,L,L,40R,April 5,2018
1005,Clinton,Williams,Clinton,clinton.williams@email.com,404-555-0105,M,M,38R,May 12,2020
1006,Larenze,Young,Larenze,,404-555-0106,XL,XL,44R,June 8,2019
1007,Jon,Parker,Jon,jon.parker@email.com,404-555-0107,L,L,40R,July 22,2021
1008,Ronnie,McCarrell,Ronnie,,404-555-0108,M,M,38R,August 3,2020
1009,Tavarus,Marshall,Tavarus,tavarus.marshall@email.com,404-555-0109,XL,XL,42R,September 14,2019
1010,Paul,Green,Paul,,404-555-0110,L,L,40R,October 30,2021
1011,Dr. Anthony,Owens,Anthony,anthony.owens@email.com,404-555-0111,M,M,38R,November 11,2020
1012,Desmond,Alvies,Desmond,,404-555-0112,L,L,40R,December 25,2018
1013,Quinn,Byrd,Quinn,quinn.byrd@email.com,404-555-0113,XL,XL,42R,January 7,2021
1014,John,Marshall,John,,404-555-0114,M,M,38R,February 28,2020
1015,Will,Jones,Will,will.jones@email.com,404-555-0115,L,L,40R,March 19,2019
1016,Danton,Thomas,Danton,,404-555-0116,M,M,38R,April 12,2021
1017,Ron,Shaw,Ron,ron.shaw@email.com,404-555-0117,XL,XL,44R,May 5,2020
1018,Anthony,Lewis,Anthony,,404-555-0118,L,L,40R,June 16,2019
1019,Dr. Ethan,Johnson,Ethan,ethan.johnson@email.com,404-555-0119,M,M,38R,July 9,2021
1020,Ernest,Hargett,Ernest,,404-555-0120,L,L,40R,August 24,2020
1021,Josh,Memminger,Josh,josh.memminger@email.com,404-555-0121,M,M,38R,September 30,2019
1022,Corey,Kelty,Corey,,404-555-0122,XL,XL,42R,October 11,2021
1023,Vince,Adejumo,Vince,vince.adejumo@email.com,404-555-0123,L,L,40R,November 2,2020
1024,Kenneth,Hargett,Kenneth,,404-555-0124,M,M,38R,December 19,2018
1025,Christian,Kelly,Christian,christian.kelly@email.com,404-555-0125,L,L,40R,January 22,2021
1026,Jerrond,Robbins,Jerrond,,404-555-0126,M,M,38R,February 14,2020
1027,Matthew,Shaw,Matthew,matthew.shaw@email.com,404-555-0127,XL,XL,42R,March 27,2019
1028,Garon,Jackson,Garon,,404-555-0128,L,L,40R,April 9,2021
1029,Jimmie,Blake,Jimmie,jimmie.blake@email.com,404-555-0129,M,M,38R,May 31,2020
1030,Aaron,Webb,Aaron,,404-555-0130,L,L,40R,June 23,2019
1031,Carnell,Smith,Carnell,carnell.smith@email.com,404-555-0131,M,M,38R,July 15,2021
1032,Lynwood,Debrew,Lynwood,,404-555-0132,XL,XL,44R,August 8,2020
1033,Tony,Broxton,Tony,tony.broxton@email.com,404-555-0133,L,L,40R,September 20,2019
1034,Jamar,Mosley,Jamar,,404-555-0134,M,M,38R,October 6,2021
1035,Ron,Washington,Ron,ron.washington@email.com,404-555-0135,L,L,40R,November 28,2020
`

type Member struct {
	PBS           string
	FirstName     string
	LastName      string
	PreferredName string
	Email         string
	PhoneNumber   string
	TShirtSize    string
	PoloSize      string
	BlazerSize    string
	BirthdayMonth string
	Sigmaversary  string
}

func main() {
	ctx := context.Background()
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		fmt.Fprintf(os.Stderr, "DATABASE_URL env var not set\n")
		os.Exit(1)
	}

	// Parse embedded CSV
	r := csv.NewReader(bytes.NewReader([]byte(rosterData)))
	records, err := r.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse CSV: %v\n", err)
		os.Exit(1)
	}

	var members []Member
	for i, row := range records {
		if i == 0 { // Skip header
			continue
		}
		if len(row) < 11 {
			continue
		}
		members = append(members, Member{
			PBS:           strings.TrimSpace(row[0]),
			FirstName:     strings.TrimSpace(row[1]),
			LastName:      strings.TrimSpace(row[2]),
			PreferredName: strings.TrimSpace(row[3]),
			Email:         strings.TrimSpace(row[4]),
			PhoneNumber:   strings.TrimSpace(row[5]),
			TShirtSize:    strings.TrimSpace(row[6]),
			PoloSize:      strings.TrimSpace(row[7]),
			BlazerSize:    strings.TrimSpace(row[8]),
			BirthdayMonth: strings.TrimSpace(row[9]),
			Sigmaversary:  strings.TrimSpace(row[10]),
		})
	}

	fmt.Printf("Loaded %d members from embedded CSV\n", len(members))

	// Connect to database
	pool, err := pgx.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close(ctx)

	// Create chapter
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
		if m.FirstName == "" || m.LastName == "" {
			skippedCount++
			continue
		}

		email := m.Email
		if email == "" {
			email = fmt.Sprintf("%s.%s@member.tausigmasigma.org", strings.ToLower(m.FirstName), strings.ToLower(m.LastName))
		}

		// Create or get user
		var actualUserID string
		err := pool.QueryRow(ctx, `
			INSERT INTO users (id, email, first_name, last_name, created_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (email) DO NOTHING
			RETURNING id
		`, uuid.New().String(), email, m.FirstName, m.LastName).Scan(&actualUserID)
		if err != nil {
			if err == pgx.ErrNoRows {
				// User already exists, fetch their ID
				err = pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&actualUserID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to get user ID for %s: %v\n", email, err)
					continue
				}
			} else {
				fmt.Fprintf(os.Stderr, "Failed to create user %s %s: %v\n", m.FirstName, m.LastName, err)
				continue
			}
		}

		// Create member
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
