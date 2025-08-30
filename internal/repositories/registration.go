package repositories

import (
	"context"
	"database/sql"
	"openhouse-2025-api/internal/models"
	"time"

	"github.com/google/uuid"
)

type RegistrationRepository struct{ db *sql.DB }

func NewRegistrationRepository(db *sql.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) Create(ctx context.Context, reg *models.DetailRegistration) error {
	reg.ID = uuid.New().String()
	now := time.Now()
	reg.CreatedAt = &now
	reg.UpdatedAt = &now
	reg.FileValidated = 0    // Default: not validated
	reg.PaymentValidated = 0 // Default: not validated

	query := `INSERT INTO detail_registrations (id, nrp, ukm_id, payment, code, drive_url, file_validated, payment_validated, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query, reg.ID, reg.NRP, reg.UkmID, reg.Payment, reg.DriveURL, reg.FileValidated, reg.PaymentValidated, reg.CreatedAt, reg.UpdatedAt)
	return err
}

func (r *RegistrationRepository) CountParticipantsByUkm(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT ukm_id, COUNT(*) as cnt FROM detail_registrations WHERE payment_validated=1 GROUP BY ukm_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int{}
	for rows.Next() {
		var ukmID string
		var cnt int
		if err := rows.Scan(&ukmID, &cnt); err != nil {
			return nil, err
		}
		result[ukmID] = cnt
	}
	return result, rows.Err()
}
