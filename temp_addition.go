// ReserveSlotForPayment reserves a slot specifically for payment page access
func (r *RegistrationRepository) ReserveSlotForPayment(ctx context.Context, nrp string, ukmID int) (*models.SlotReservationResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get UKM quota and current slot info
	var maxSlot, currentSlot int
	err = tx.QueryRowContext(ctx, `
		SELECT max_slot, COALESCE(current_slot, 0) FROM ukms WHERE id = ?
	`, ukmID).Scan(&maxSlot, &currentSlot)
	if err != nil {
		return nil, fmt.Errorf("UKM not found")
	}

	// Check current reservations
	var currentReserved int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM slot_reservations WHERE ukm_id = ? AND expires_at > NOW()
	`, ukmID).Scan(&currentReserved)
	if err != nil {
		return nil, err
	}

	// Check if slots are available (current_slot + active reservations >= max_slot)
	if currentSlot+currentReserved >= maxSlot {
		return nil, nil // No slots available
	}

	// Check if user already has a reservation for this UKM
	var existingReservation string
	var existingExpiry time.Time
	err = tx.QueryRowContext(ctx, `
		SELECT reservation_id, expires_at FROM slot_reservations 
		WHERE nrp = ? AND ukm_id = ? AND expires_at > NOW()
	`, nrp, ukmID).Scan(&existingReservation, &existingExpiry)
	if err == nil {
		// User already has a valid reservation, return it
		return &models.SlotReservationResult{
			ReservationID: existingReservation,
			ExpiresAt:     existingExpiry,
			CurrentSlot:   currentSlot + currentReserved,
			MaxSlot:       maxSlot,
		}, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new reservation
	reservationID := uuid.New().String()
	expiresAt := time.Now().Add(5 * time.Minute) // 5 minutes for payment

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
		CurrentSlot:   currentSlot + currentReserved + 1, // +1 for this new reservation
		MaxSlot:       maxSlot,
	}, nil
}
</content>
</invoke>
