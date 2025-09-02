package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/repositories"
)

type UkmSeed struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Current   int     `json:"current"`
	Max       int     `json:"max"`
	Fee       int     `json:"fee"`
	Desc      string  `json:"desc"`
	LogoURL   string  `json:"logo_url"`
	PosterURL *string `json:"poster_url"`
	Groupchat string  `json:"groupchat"`
}

func main() {
	cfg := config.Load()
	sqlDB, _, _ := repositories.NewDatabaseConnections(cfg)
	defer sqlDB.Close()

	// Load from JSON only
	if err := seedFromJSON(sqlDB, "db/seeds/ukms.json"); err != nil {
		log.Fatalf("seeding from JSON failed: %v", err)
	}
	log.Println("ukm seeding completed")
}

func seedFromJSON(db *sql.DB, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var list []UkmSeed
	dec := json.NewDecoder(f)
	if err := dec.Decode(&list); err != nil {
		return err
	}
	for _, s := range list {
		if err := upsertUkmBySlug(db, s); err != nil {
			return err
		}
	}
	return nil
}

func upsertUkmBySlug(db *sql.DB, s UkmSeed) error {
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
	err = tx.QueryRow(`SELECT id FROM ukms WHERE slug = ? LIMIT 1`, s.Slug).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		// Insert new row with generated UUID()
		_, err = tx.Exec(`INSERT INTO ukms
			(id, name, slug, current_slot, max_slot, regist_fee, description, logo_url, poster_url, groupchat)
			VALUES (UUID(), ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			s.Name, s.Slug, s.Current, s.Max, s.Fee, s.Desc, s.LogoURL, s.PosterURL, s.Groupchat,
		)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		// Update existing row by slug
		_, err = tx.Exec(`UPDATE ukms SET
			name = ?, current_slot = ?, max_slot = ?, regist_fee = ?, description = ?, logo_url = ?, poster_url = ?, groupchat = ?
			WHERE slug = ?`,
			s.Name, s.Current, s.Max, s.Fee, s.Desc, s.LogoURL, s.PosterURL, s.Groupchat, s.Slug,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
