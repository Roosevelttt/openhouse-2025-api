package repositories

import (
	"context"
	"database/sql"
	"openhouse-2025-api/internal/models"
)

type AdminRepository struct{ db *sql.DB }

func NewAdminRepository(db *sql.DB) *AdminRepository { return &AdminRepository{db: db} }

func (r *AdminRepository) FindByNRP(ctx context.Context, nrp string) (*models.Admin, error) {
	row := r.db.QueryRowContext(ctx, `
			SELECT a.id, a.ukm_id, u.name AS ukm_name, a.division_id, d.slug AS division_slug, a.name, a.nrp, a.field, a.created_at, a.updated_at 
			FROM admins a
			LEFT JOIN divisions d ON d.id=a.division_id
			LEFT JOIN ukms u ON u.id=a.ukm_id
			WHERE nrp=?`, nrp)
	var a models.Admin
	if err := row.Scan(&a.ID, &a.UkmID, &a.UkmName, &a.DivisionID, &a.DivisionSlug, &a.Name, &a.NRP, &a.Field, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepository) Create(ctx context.Context, a *models.Admin) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO admins (id, ukm_id, division_id, name, nrp, field, created_at, updated_at) VALUES (UUID(),?,?,?,?,?,NOW(),NOW())`, a.UkmID, a.DivisionID, a.Name, a.NRP, a.Field)
	return err
}

func (r *AdminRepository) List(ctx context.Context) ([]*models.Admin, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.ukm_id, u.name AS ukm_name, a.division_id, d.name AS division_name, a.name, a.nrp, a.field, a.created_at, a.updated_at 
		FROM admins a
		LEFT JOIN divisions d ON d.id=a.division_id
		LEFT JOIN ukms u ON u.id=a.ukm_id
		ORDER BY a.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []*models.Admin
	for rows.Next() {
		var a models.Admin
		if err := rows.Scan(&a.ID, &a.UkmID, &a.UkmName, &a.DivisionID, &a.DivisionName, &a.Name, &a.NRP, &a.Field, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		admins = append(admins, &a)
	}
	return admins, nil
}

func (r *AdminRepository) GetByID(ctx context.Context, id string) (*models.Admin, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT a.id, a.ukm_id, u.name AS ukm_name, a.division_id, d.name AS division_name, a.name, a.nrp, a.field, a.created_at, a.updated_at 
		FROM admins a
		LEFT JOIN divisions d ON d.id=a.division_id
		LEFT JOIN ukms u ON u.id=a.ukm_id
		WHERE a.id=?`, id)
	var a models.Admin
	if err := row.Scan(&a.ID, &a.UkmID, &a.UkmName, &a.DivisionID, &a.DivisionName, &a.Name, &a.NRP, &a.Field, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepository) Update(ctx context.Context, id string, a *models.Admin) error {
	_, err := r.db.ExecContext(ctx, `UPDATE admins SET ukm_id=?, division_id=?, name=?, nrp=?, field=?, updated_at=NOW() WHERE id=?`, a.UkmID, a.DivisionID, a.Name, a.NRP, a.Field, id)
	return err
}

func (r *AdminRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admins WHERE id=?`, id)
	return err
}
