package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/repositories"
)

type DivisionSeed struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func main() {
	cfg := config.Load()
	sqlDB, _, _ := repositories.NewDatabaseConnections(cfg)
	defer sqlDB.Close()

	// Load from JSON only
	if err := seedFromJSON(sqlDB, "db/seeds/divisions.json"); err != nil {
		log.Fatalf("seeding from JSON failed: %v", err)
	}
	log.Println("division seeding completed")
}

func seedFromJSON(db *sql.DB, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var list []DivisionSeed
	dec := json.NewDecoder(f)
	if err := dec.Decode(&list); err != nil {
		return err
	}
	for _, s := range list {
		if err := upsertDivisionBySlug(db, s); err != nil {
			return err
		}
	}
	return nil
}

func upsertDivisionBySlug(db *sql.DB, s DivisionSeed) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var id string
	err = tx.QueryRow(`SELECT id FROM divisions WHERE slug = ? LIMIT 1`, s.Slug).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		// Insert new row with generated UUID()
		_, err = tx.Exec(`INSERT INTO divisions
			(id, name, slug, created_at, updated_at)
			VALUES (UUID(), ?, ?, NOW(), NOW())`,
			s.Name, s.Slug,
		)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		// Update existing row by slug
		_, err = tx.Exec(`UPDATE divisions SET
			name = ?, updated_at = NOW()
			WHERE slug = ?`,
			s.Name, s.Slug,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
