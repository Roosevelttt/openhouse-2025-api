package repositories

import (
	"context"
	"database/sql"

	"openhouse-2025-api/internal/models"
)

type ParticipantsRepository struct{ db *sql.DB }

func NewParticipantsRepository(db *sql.DB) *ParticipantsRepository {
	return &ParticipantsRepository{db: db}
}

func (r *ParticipantsRepository) List(ctx context.Context, adminDivisionSlug string, adminUkmID string) ([]models.Participant, error) {
	query := `
		SELECT u.id ,u.nrp, u.name, line_id, phone, 
		       uk.id AS ukm_id, uk.name AS ukm_name, 
		       dr.payment, dr.drive_url, dr.file_validated, dr.payment_validated, dr.created_at 
		FROM users u 
		JOIN detail_registrations dr ON u.nrp=dr.nrp
		JOIN ukms uk ON uk.id=dr.ukm_id
		`

	// Siapkan slice untuk menyimpan argumen kueri secara dinamis
	var args []interface{}

	if adminDivisionSlug != "it" && adminDivisionSlug != "bph" {
		query += `WHERE uk.id = ?`
		args = append(args, adminUkmID)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.Participant
	for rows.Next() {
		var p models.Participant
		if err := rows.Scan(
			&p.ID,
			&p.NRP,
			&p.Name,
			&p.LineID,
			&p.Phone,
			&p.UkmId,
			&p.UkmName,
			&p.Payment,
			&p.DriveURL,
			&p.FileValidated,
			&p.PaymentValidated,
			&p.CreatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, rows.Err()
}
