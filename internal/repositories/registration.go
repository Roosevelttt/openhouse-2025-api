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

// CheckExistingRegistration checks if a user has already registered for a UKM
func (r *RegistrationRepository) CheckExistingRegistration(ctx context.Context, nrp, ukmID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM detail_registrations 
		WHERE nrp = ? AND ukm_id = ?
	`, nrp, ukmID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
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

	// Get UKM current slot and quota with SELECT FOR UPDATE to prevent race conditions
	var currentSlot, quota int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(current_slot, 0), max_slot FROM ukms WHERE id = ? FOR UPDATE
	`, ukmID).Scan(&currentSlot, &quota)
	if err != nil {
		return "", fmt.Errorf("UKM not found")
	}

	var currentReserved int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM slot_reservations WHERE ukm_id = ? AND expires_at > UTC_TIMESTAMP()
	`, ukmID).Scan(&currentReserved)
	if err != nil {
		return "", err
	}

	// Check if slots are available (implement race condition control)
	totalTaken := currentSlot + currentReserved
	if totalTaken >= quota {
		return "", fmt.Errorf("no slots available")
	}

	// For race condition control: limit concurrent reservations to remaining slots
	remainingSlots := quota - currentSlot
	fmt.Printf("DEBUG ReserveSlot: currentSlot=%d, currentReserved=%d, quota=%d, remainingSlots=%d\n",
		currentSlot, currentReserved, quota, remainingSlots)
	if currentReserved >= remainingSlots {
		fmt.Printf("DEBUG ReserveSlot: BLOCKING - too many concurrent reservations\n")
		return "", fmt.Errorf("no slots available - too many concurrent reservations")
	}

	// Check if user already has a reservation for this UKM
	var existingReservation string
	err = tx.QueryRowContext(ctx, `
		SELECT reservation_id FROM slot_reservations 
		WHERE nrp = ? AND ukm_id = ? AND expires_at > UTC_TIMESTAMP()
	`, nrp, ukmID).Scan(&existingReservation)
	if err == nil {
		// User already has a valid reservation
		return existingReservation, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	// Create new reservation with database-calculated expiry time
	reservationID := uuid.New().String()
	expiresAt := time.Now().UTC().Add(5 * time.Minute) // 5 minutes from now in UTC

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

	// Check if user has already registered for this UKM
	var existingRegistrations int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM detail_registrations 
		WHERE nrp = ? AND ukm_id = ?
	`, reg.NRP, reg.UkmID).Scan(&existingRegistrations)
	if err != nil {
		return err
	}

	if existingRegistrations > 0 {
		return fmt.Errorf("user has already registered for this UKM")
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
func (r *RegistrationRepository) ReserveSlotForPayment(ctx context.Context, nrp string, ukmID string) (*models.SlotReservationResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Check if user has already registered for this UKM
	var existingRegistrations int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM detail_registrations 
		WHERE nrp = ? AND ukm_id = ?
	`, nrp, ukmID).Scan(&existingRegistrations)
	if err != nil {
		return nil, err
	}

	if existingRegistrations > 0 {
		return nil, fmt.Errorf("user has already registered for this UKM")
	}

	// Get UKM current slot and quota with SELECT FOR UPDATE to prevent race conditions
	var currentSlot, quota int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(current_slot, 0), max_slot FROM ukms WHERE id = ? FOR UPDATE
	`, ukmID).Scan(&currentSlot, &quota)
	if err != nil {
		return nil, fmt.Errorf("UKM not found")
	}

	var currentReserved int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM slot_reservations WHERE ukm_id = ? AND expires_at > UTC_TIMESTAMP()
	`, ukmID).Scan(&currentReserved)
	if err != nil {
		return nil, err
	}

	// Check if slots are available (implement race condition control)
	totalTaken := currentSlot + currentReserved
	if totalTaken >= quota {
		return nil, nil // Return nil to indicate no slots available
	}

	// For race condition control: limit concurrent reservations to remaining slots
	remainingSlots := quota - currentSlot
	if currentReserved >= remainingSlots {
		return nil, nil // Limit concurrent payment access to remaining slots
	}

	// Check if user already has a reservation for this UKM
	var existingReservation string
	var existingExpiry time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT reservation_id, expires_at FROM slot_reservations 
		WHERE nrp = ? AND ukm_id = ? AND expires_at > UTC_TIMESTAMP()
	`, nrp, ukmID).Scan(&existingReservation, &existingExpiry)
	if err == nil {
		// User already has a valid reservation, return existing details
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return &models.SlotReservationResult{
			ReservationID: existingReservation,
			ExpiresAt:     existingExpiry,
			CurrentSlot:   currentSlot,
			MaxSlot:       quota,
		}, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new reservation with Go-calculated expiry time to avoid timezone issues
	reservationID := uuid.New().String()
	expiresAt := time.Now().UTC().Add(5 * time.Minute) // 5 minutes from now in UTC

	_, err = tx.ExecContext(ctx, `
		INSERT INTO slot_reservations (reservation_id, nrp, ukm_id, expires_at) 
		VALUES (?, ?, ?, ?)
	`, reservationID, nrp, ukmID, expiresAt)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &models.SlotReservationResult{
		ReservationID: reservationID,
		ExpiresAt:     expiresAt,
		CurrentSlot:   currentSlot,
		MaxSlot:       quota,
	}, nil
}
