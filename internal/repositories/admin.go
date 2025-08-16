package repositories

import (
	"context"
	"database/sql"

	"openhouse-2025-api/internal/models"
)

type AdminRepository struct { db *sql.DB }
func NewAdminRepository(db *sql.DB) *AdminRepository { return &AdminRepository{db: db} }

func (r *AdminRepository) FindByNRP(ctx context.Context, nrp string) (*models.Admin, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, ukm_id, division_id, name, nrp, field, created_at, updated_at FROM admins WHERE nrp=?`, nrp)
	var a models.Admin
	if err := row.Scan(&a.ID, &a.UkmID, &a.DivisionID, &a.Name, &a.NRP, &a.Field, &a.CreatedAt, &a.UpdatedAt); err != nil { return nil, err }
	return &a, nil
}

func (r *AdminRepository) Create(ctx context.Context, a *models.Admin) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO admins (id, ukm_id, division_id, name, nrp, field, created_at, updated_at) VALUES (UUID(),?,?,?,?,?,NOW(),NOW())`, a.UkmID, a.DivisionID, a.Name, a.NRP, a.Field)
	return err
}

