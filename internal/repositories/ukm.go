package repositories

import (
	"context"
	"database/sql"

	"openhouse-2025-api/internal/models"
)

type UkmRepository struct { db *sql.DB }
func NewUkmRepository(db *sql.DB) *UkmRepository { return &UkmRepository{db: db} }

func (r *UkmRepository) List(ctx context.Context) ([]models.Ukm, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, slug, current_slot, max_slot, regist_fee FROM ukms ORDER BY name`)
	if err != nil { return nil, err }
	defer rows.Close()
	var res []models.Ukm
	for rows.Next() {
		var u models.Ukm
		if err := rows.Scan(&u.ID, &u.Name, &u.Slug, &u.CurrentSlot, &u.MaxSlot, &u.RegistFee); err != nil { return nil, err }
		res = append(res, u)
	}
	return res, rows.Err()
}

