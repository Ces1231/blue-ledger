package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	direction := flag.String("direction", "up", "Migration direction: up or down")
	steps := flag.Int("steps", 0, "Number of steps (0 means all for up, 1 for down)")
	flag.Parse()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrator: %v\n", err)
		os.Exit(1)
	}
	defer m.Close()

	var migrateErr error

	switch *direction {
	case "up":
		if *steps > 0 {
			migrateErr = m.Steps(*steps)
		} else {
			migrateErr = m.Up()
		}
	case "down":
		n := *steps
		if n <= 0 {
			n = 1
		}
		migrateErr = m.Steps(-n)
	case "drop":
		migrateErr = m.Drop()
	default:
		fmt.Fprintf(os.Stderr, "unknown direction: %s (use up, down, or drop)\n", *direction)
		os.Exit(1)
	}

	if migrateErr != nil && migrateErr != migrate.ErrNoChange {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", migrateErr)
		os.Exit(1)
	}

	v, dirty, _ := m.Version()
	fmt.Printf("migration complete — version: %d, dirty: %v\n", v, dirty)
}
