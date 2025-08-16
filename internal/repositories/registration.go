package repositories

import (
	"context"
	"database/sql"
)

type RegistrationRepository struct { db *sql.DB }
func NewRegistrationRepository(db *sql.DB) *RegistrationRepository { return &RegistrationRepository{db: db} }

func (r *RegistrationRepository) CountParticipantsByUkm(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT ukm_id, COUNT(*) as cnt FROM detail_registrations WHERE payment_validated=1 GROUP BY ukm_id`)
	if err != nil { return nil, err }
	defer rows.Close()
	result := map[string]int{}
	for rows.Next() {
		var ukmID string
		var cnt int
		if err := rows.Scan(&ukmID, &cnt); err != nil { return nil, err }
		result[ukmID] = cnt
	}
	return result, rows.Err()
}

