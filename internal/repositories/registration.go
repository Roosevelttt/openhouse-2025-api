package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"openhouse-2025-api/internal/models"
	"time"

	"github.com/google/uuid"
)

type RegistrationRepository struct{ db *sql.DB }

func NewRegistrationRepository(db *sql.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) Create(ctx context.Context, reg *models.DetailRegistration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	reg.ID = uuid.New().String()
	now := time.Now()
	reg.CreatedAt = &now
	reg.UpdatedAt = &now
	reg.FileValidated = 0    // Default: not validated
	reg.PaymentValidated = 0 // Default: not validated

	query := `INSERT INTO detail_registrations (id, nrp, ukm_id, payment, drive_url, file_validated, payment_validated, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query, reg.ID, reg.NRP, reg.UkmID, reg.Payment, reg.DriveURL, reg.FileValidated, reg.PaymentValidated, reg.CreatedAt, reg.UpdatedAt)
	if err != nil {
		return err
	}

	// Increment UKM current_slot
	_, err = tx.ExecContext(ctx, `
		UPDATE ukms SET current_slot = COALESCE(current_slot, 0) + 1 WHERE id = ?
	`, reg.UkmID)
	if err != nil {
		return err
	}

	return tx.Commit()
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

// ReserveSlot reserves a slot for the user for 10 minutes
func (r *RegistrationRepository) ReserveSlot(ctx context.Context, nrp, ukmID string) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// Check current registrations + reservations vs quota
	var currentRegistered int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM detail_registrations WHERE ukm_id = ? AND payment_validated = 1
	`, ukmID).Scan(&currentRegistered)
	if err != nil {
		return "", err
	}

	var currentReserved int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM slot_reservations WHERE ukm_id = ? AND expires_at > NOW()
	`, ukmID).Scan(&currentReserved)
	if err != nil {
		return "", err
	}

	// Get UKM quota
	var quota int
	err = tx.QueryRowContext(ctx, `
		SELECT max_slot FROM ukms WHERE id = ?
	`, ukmID).Scan(&quota)
	if err != nil {
		return "", fmt.Errorf("UKM not found")
	}

	// Check if slots are available
	if currentRegistered+currentReserved >= quota {
		return "", fmt.Errorf("no slots available")
	}

	// Check if user already has a reservation for this UKM
	var existingReservation string
	err = tx.QueryRowContext(ctx, `
		SELECT reservation_id FROM slot_reservations 
		WHERE nrp = ? AND ukm_id = ? AND expires_at > NOW()
	`, nrp, ukmID).Scan(&existingReservation)
	if err == nil {
		// User already has a valid reservation
		return existingReservation, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	// Create new reservation
	reservationID := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Minute)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO slot_reservations (reservation_id, nrp, ukm_id, expires_at) 
		VALUES (?, ?, ?, ?)
	`, reservationID, nrp, ukmID, expiresAt)
	if err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return reservationID, nil
}

// ValidateReservation checks if a reservation is still valid
func (r *RegistrationRepository) ValidateReservation(ctx context.Context, reservationID, nrp string) (bool, string, error) {
	var ukmID string
	var expiresAt time.Time

	err := r.db.QueryRowContext(ctx, `
		SELECT ukm_id, expires_at FROM slot_reservations 
		WHERE reservation_id = ? AND nrp = ?
	`, reservationID, nrp).Scan(&ukmID, &expiresAt)

	if err == sql.ErrNoRows {
		return false, "", fmt.Errorf("reservation not found")
	}
	if err != nil {
		return false, "", err
	}

	if time.Now().After(expiresAt) {
		return false, "", fmt.Errorf("reservation has expired")
	}

	return true, ukmID, nil
}

// ConsumeReservation converts a reservation into an actual registration
func (r *RegistrationRepository) ConsumeReservation(ctx context.Context, reservationID string, reg *models.DetailRegistration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Validate reservation still exists and is valid
	var ukmID string
	var expiresAt time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT ukm_id, expires_at FROM slot_reservations 
		WHERE reservation_id = ? AND nrp = ?
	`, reservationID, reg.NRP).Scan(&ukmID, &expiresAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("reservation not found")
	}
	if err != nil {
		return err
	}

	if time.Now().After(expiresAt) {
		return fmt.Errorf("reservation has expired")
	}

	if ukmID != reg.UkmID {
		return fmt.Errorf("reservation UKM mismatch")
	}

	// Create registration
	reg.ID = uuid.New().String()
	now := time.Now()
	reg.CreatedAt = &now
	reg.UpdatedAt = &now
	reg.FileValidated = 0
	reg.PaymentValidated = 0

	_, err = tx.ExecContext(ctx, `
		INSERT INTO detail_registrations (id, nrp, ukm_id, payment, drive_url, file_validated, payment_validated, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, reg.ID, reg.NRP, reg.UkmID, reg.Payment, reg.DriveURL, reg.FileValidated, reg.PaymentValidated, reg.CreatedAt, reg.UpdatedAt)
	if err != nil {
		return err
	}

	// Delete the reservation
	_, err = tx.ExecContext(ctx, `
		DELETE FROM slot_reservations WHERE reservation_id = ?
	`, reservationID)
	if err != nil {
		return err
	}

	// Increment UKM current_slot
	_, err = tx.ExecContext(ctx, `
		UPDATE ukms SET current_slot = COALESCE(current_slot, 0) + 1 WHERE id = ?
	`, reg.UkmID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ReserveSlotForPayment reserves a slot for payment page access with full details
func (r *RegistrationRepository) ReserveSlotForPayment(ctx context.Context, nrp string, ukmID int) (*models.SlotReservationResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check current registrations + reservations vs quota
	var currentRegistered int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM detail_registrations WHERE ukm_id = ? AND payment_validated = 1
	`, ukmID).Scan(&currentRegistered)
	if err != nil {
		return nil, err
	}

	var currentReserved int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM slot_reservations WHERE ukm_id = ? AND expires_at > NOW()
	`, ukmID).Scan(&currentReserved)
	if err != nil {
		return nil, err
	}

	// Get UKM quota
	var quota int
	err = tx.QueryRowContext(ctx, `
		SELECT max_slot FROM ukms WHERE id = ?
	`, ukmID).Scan(&quota)
	if err != nil {
		return nil, fmt.Errorf("UKM not found")
	}

	// Check if slots are available (implement race condition control)
	totalTaken := currentRegistered + currentReserved
	if totalTaken >= quota {
		return nil, nil // Return nil to indicate no slots available
	}

	// For race condition control: when near capacity, limit concurrent payment page access
	remainingSlots := quota - currentRegistered
	if remainingSlots <= 2 && currentReserved >= 2 {
		return nil, nil // Limit concurrent payment access when nearly full
	}

	// Check if user already has a reservation for this UKM
	var existingReservation string
	var existingExpiry time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT reservation_id, expires_at FROM slot_reservations 
		WHERE nrp = ? AND ukm_id = ? AND expires_at > NOW()
	`, nrp, ukmID).Scan(&existingReservation, &existingExpiry)
	if err == nil {
		// User already has a valid reservation, return existing details
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return &models.SlotReservationResult{
			ReservationID: existingReservation,
			ExpiresAt:     existingExpiry,
			CurrentSlot:   currentRegistered,
			MaxSlot:       quota,
		}, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new reservation with database-calculated expiry time
	reservationID := uuid.New().String()
	var expiresAt time.Time

	err = tx.QueryRowContext(ctx, `
		INSERT INTO slot_reservations (reservation_id, nrp, ukm_id, expires_at) 
		VALUES (?, ?, ?, DATE_ADD(NOW(), INTERVAL 5 MINUTE))
		RETURNING expires_at
	`, reservationID, nrp, ukmID).Scan(&expiresAt)
	if err != nil {
		// Fallback for databases that don't support RETURNING
		_, err = tx.ExecContext(ctx, `
			INSERT INTO slot_reservations (reservation_id, nrp, ukm_id, expires_at) 
			VALUES (?, ?, ?, DATE_ADD(NOW(), INTERVAL 5 MINUTE))
		`, reservationID, nrp, ukmID)
		if err != nil {
			return nil, err
		}

		// Get the expires_at value
		err = tx.QueryRowContext(ctx, `
			SELECT expires_at FROM slot_reservations WHERE reservation_id = ?
		`, reservationID).Scan(&expiresAt)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &models.SlotReservationResult{
		ReservationID: reservationID,
		ExpiresAt:     expiresAt,
		CurrentSlot:   currentRegistered,
		MaxSlot:       quota,
	}, nil
}
