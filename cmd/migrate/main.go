package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	defaultMigrationSource = "file://deployments/postgresql/migrations"
)

func main() {
	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	sourceURL := os.Getenv("MIGRATIONS_SOURCE")
	if sourceURL == "" {
		sourceURL = defaultMigrationSource
	}

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		log.Fatalf("failed to initialize migrate: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("warning: failed to close migration source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("warning: failed to close migration db: %v", dbErr)
		}
	}()

	switch action {
	case "up":
		runUp(m)
	case "down":
		runDown(m)
	default:
		log.Fatalf("unsupported action %q (use: up|down)", action)
	}
}

func runUp(m *migrate.Migrate) {
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no change")
			return
		}
		log.Fatalf("migration up failed: %v", err)
	}
	log.Println("migration up complete")
}

func runDown(m *migrate.Migrate) {
	if err := m.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no change")
			return
		}
		log.Fatalf("migration down failed: %v", err)
	}
	fmt.Println("migration down complete")
}
