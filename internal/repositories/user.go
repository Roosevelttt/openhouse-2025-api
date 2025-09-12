package repositories

import (
	"context"
	"database/sql"

	"openhouse-2025-api/internal/models"

	"github.com/google/uuid"
)

type UserRepository struct{ db *sql.DB }

func NewUserRepository(db *sql.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) UpsertByNRP(ctx context.Context, u *models.User) error {
	var existingID string
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE nrp=?`, u.NRP).Scan(&existingID)

	id := uuid.New().String()
	if err == nil && existingID != "" {
		id = existingID
	}

	_, err = r.db.ExecContext(ctx, `INSERT INTO users (id, nrp, name, line_id, phone) VALUES (?,?,?,?,?)
		ON DUPLICATE KEY UPDATE name=VALUES(name), line_id=VALUES(line_id), phone=VALUES(phone)`,
		id, u.NRP, u.Name, u.LineID, u.Phone)
	return err
}

func (r *UserRepository) FindByNRP(ctx context.Context, nrp string) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT nrp, name, line_id, phone FROM users WHERE nrp=?`, nrp)
	var u models.User
	if err := row.Scan(&u.NRP, &u.Name, &u.LineID, &u.Phone); err != nil {
		return nil, err
	}
	return &u, nil
}
