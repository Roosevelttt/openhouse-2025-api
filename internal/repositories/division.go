package repositories

import (
	"context"
	"database/sql"

	"openhouse-2025-api/internal/models"
)

type DivisionRepository struct { db *sql.DB }
func NewDivisionRepository(db *sql.DB) *DivisionRepository { return &DivisionRepository{db: db} }

func (r *DivisionRepository) List(ctx context.Context) ([]models.Division, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, slug, created_at, updated_at FROM divisions ORDER BY name`)
	if err != nil { return nil, err }
	defer rows.Close()
	var res []models.Division
	for rows.Next() {
		var d models.Division
		if err := rows.Scan(&d.ID, &d.Name, &d.Slug, &d.CreatedAt, &d.UpdatedAt); err != nil { return nil, err }
		res = append(res, d)
	}
	return res, rows.Err()
}

