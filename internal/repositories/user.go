package repositories

import (
	"context"
	"database/sql"

	"openhouse-2025-api/internal/models"
)

type UserRepository struct { db *sql.DB }

func NewUserRepository(db *sql.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) UpsertByNRP(ctx context.Context, u *models.User) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (nrp, name, line_id, phone) VALUES (?,?,?,?)
		ON DUPLICATE KEY UPDATE name=VALUES(name), line_id=VALUES(line_id), phone=VALUES(phone)`, u.NRP, u.Name, u.LineID, u.Phone)
	return err
}

func (r *UserRepository) FindByNRP(ctx context.Context, nrp string) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT nrp, name, line_id, phone FROM users WHERE nrp=?`, nrp)
	var u models.User
	if err := row.Scan(&u.NRP, &u.Name, &u.LineID, &u.Phone); err != nil { return nil, err }
	return &u, nil
}

