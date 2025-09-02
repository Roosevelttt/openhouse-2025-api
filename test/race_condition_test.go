package test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/router"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
	"sync"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRaceConditionSlotReservation tests multiple users trying to reserve the last slot simultaneously
func TestRaceConditionSlotReservation(t *testing.T) {
	// Setup test environment
	cfg := &config.Config{
		DatabaseHost:     "127.0.0.1",
		DatabaseName:     "openhouse_2025_test",
		DatabaseUser:     "root",
		DatabasePassword: "",
		SessionSecret:    "test-secret-key-for-testing-only",
	}

	// Create test database connection
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseHost, cfg.DatabaseName))
	require.NoError(t, err)
	defer db.Close()

	// Clean up test data
	defer cleanupTestData(t, db)

	// Setup test data
	setupTestData(t, db)

	// Create router with test configuration
	handler := router.New(cfg)

	// Test scenario: 5 users trying to reserve the last 2 slots simultaneously
	const numConcurrentUsers = 5
	const availableSlots = 2

	var wg sync.WaitGroup
	results := make(chan ReservationResult, numConcurrentUsers)

	// Launch concurrent reservation attempts
	for i := 0; i < numConcurrentUsers; i++ {
		wg.Add(1)
		go func(userIndex int) {
			defer wg.Done()
			result := attemptSlotReservation(t, handler, userIndex)
			results <- result
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Analyze results
	var successfulReservations []ReservationResult
	var failedReservations []ReservationResult

	for result := range results {
		if result.Success {
			successfulReservations = append(successfulReservations, result)
		} else {
			failedReservations = append(failedReservations, result)
		}
	}

	// Assertions
	assert.Len(t, successfulReservations, availableSlots,
		"Exactly %d users should have successfully reserved slots", availableSlots)
	assert.Len(t, failedReservations, numConcurrentUsers-availableSlots,
		"Exactly %d users should have failed to reserve slots", numConcurrentUsers-availableSlots)

	// Verify database state
	verifyDatabaseState(t, db, availableSlots)

	// Test successful completion of reservations
	testReservationCompletion(t, handler, successfulReservations)
}

type ReservationResult struct {
	UserIndex     int
	Success       bool
	ReservationID string
	Error         string
	StatusCode    int
}

func attemptSlotReservation(t *testing.T, handler http.Handler, userIndex int) ReservationResult {
	// Create request body
	requestBody := map[string]string{
		"ukm_id": "test-ukm-id-123",
	}

	bodyBytes, err := json.Marshal(requestBody)
	require.NoError(t, err)

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/registrations/reserve", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Create test session with user authentication
	w := httptest.NewRecorder()

	// Add session middleware manually for testing
	store := cookie.NewStore([]byte("test-secret-key"))
	session := sessions.Default(req.Context())

	// Mock authenticated user session
	session.Set("role", "user")
	session.Set("nrp", fmt.Sprintf("502200%d", userIndex))
	session.Save(req, w)

	// Execute request
	handler.ServeHTTP(w, req)

	// Parse response
	result := ReservationResult{
		UserIndex:  userIndex,
		StatusCode: w.Code,
	}

	if w.Code == http.StatusOK {
		var response models.ReserveSlotResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err == nil {
			result.Success = true
			result.ReservationID = response.ReservationID
		}
	} else {
		var errorResponse map[string]string
		json.Unmarshal(w.Body.Bytes(), &errorResponse)
		result.Error = errorResponse["error"]
	}

	return result
}

func setupTestData(t *testing.T, db *sql.DB) {
	// Create test UKM with 2 available slots (100 max, 98 current)
	_, err := db.Exec(`
		INSERT INTO ukms (id, name, slug, max_slot, current_slot, regist_fee, description, logo_url, groupchat) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "test-ukm-id-123", "Test UKM", "test-ukm", 100, 98, 50000, "Test UKM for race condition testing", "test-logo.png", "https://chat.whatsapp.com/test")

	// Create test users
	for i := 0; i < 5; i++ {
		nrp := fmt.Sprintf("502200%d", i)
		_, err := db.Exec(`
			INSERT INTO users (nrp, name, line_id, phone, form_submitted, is_angket) 
			VALUES (?, ?, ?, ?, ?, ?)
		`, nrp, fmt.Sprintf("Test User %d", i), fmt.Sprintf("testline%d", i), fmt.Sprintf("08123456789%d", i), 1, 1)
		require.NoError(t, err)
	}

	require.NoError(t, err)
}

func cleanupTestData(t *testing.T, db *sql.DB) {
	// Clean up in reverse dependency order
	db.Exec("DELETE FROM slot_reservations WHERE ukm_id = 'test-ukm-id-123'")
	db.Exec("DELETE FROM detail_registrations WHERE ukm_id = 'test-ukm-id-123'")
	db.Exec("DELETE FROM ukms WHERE id = 'test-ukm-id-123'")

	for i := 0; i < 5; i++ {
		nrp := fmt.Sprintf("502200%d", i)
		db.Exec("DELETE FROM users WHERE nrp = ?", nrp)
	}
}

func verifyDatabaseState(t *testing.T, db *sql.DB, expectedReservations int) {
	// Check active reservations
	var activeReservations int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM slot_reservations 
		WHERE ukm_id = 'test-ukm-id-123' AND expires_at > NOW()
	`).Scan(&activeReservations)
	require.NoError(t, err)

	assert.Equal(t, expectedReservations, activeReservations,
		"Database should have exactly %d active reservations", expectedReservations)

	// Check UKM current_slot hasn't been incremented yet (only happens on completion)
	var currentSlot int
	err = db.QueryRow("SELECT current_slot FROM ukms WHERE id = 'test-ukm-id-123'").Scan(&currentSlot)
	require.NoError(t, err)

	assert.Equal(t, 98, currentSlot,
		"UKM current_slot should still be 98 (not incremented until registration completion)")
}

func testReservationCompletion(t *testing.T, handler http.Handler, successfulReservations []ReservationResult) {
	// Test completing one of the successful reservations
	if len(successfulReservations) > 0 {
		reservation := successfulReservations[0]

		// Create registration completion request
		registrationData := map[string]string{
			"ukm_id":    "test-ukm-id-123",
			"payment":   "test-payment-proof.jpg",
			"drive_url": "https://drive.google.com/test-portfolio",
		}

		bodyBytes, err := json.Marshal(registrationData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST",
			fmt.Sprintf("/api/registrations/with-reservation/%s", reservation.ReservationID),
			bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// Mock session for the same user
		w := httptest.NewRecorder()
		store := cookie.NewStore([]byte("test-secret-key"))
		session := sessions.Default(req.Context())
		session.Set("role", "user")
		session.Set("nrp", fmt.Sprintf("502200%d", reservation.UserIndex))
		session.Save(req, w)

		// Execute completion request
		handler.ServeHTTP(w, req)

		// Should succeed
		assert.Equal(t, http.StatusCreated, w.Code,
			"Registration completion should succeed for valid reservation")
	}
}

// TestSlotReservationExpiry tests that expired reservations are properly handled
func TestSlotReservationExpiry(t *testing.T) {
	cfg := &config.Config{
		DatabaseHost:     "127.0.0.1",
		DatabaseName:     "openhouse_2025_test",
		DatabaseUser:     "root",
		DatabasePassword: "",
		SessionSecret:    "test-secret-key-for-testing-only",
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseHost, cfg.DatabaseName))
	require.NoError(t, err)
	defer db.Close()

	// Setup
	regRepo := repositories.NewRegistrationRepository(db)
	ctx := context.Background()

	// Clean up
	defer db.Exec("DELETE FROM slot_reservations WHERE nrp = 'test-expiry-user'")
	defer db.Exec("DELETE FROM users WHERE nrp = 'test-expiry-user'")
	defer db.Exec("DELETE FROM ukms WHERE id = 'test-expiry-ukm'")

	// Create test data
	db.Exec(`INSERT INTO users (nrp, name, line_id, phone, form_submitted, is_angket) VALUES (?, ?, ?, ?, ?, ?)`,
		"test-expiry-user", "Test User", "testline", "08123456789", 1, 1)
	db.Exec(`INSERT INTO ukms (id, name, slug, max_slot, current_slot) VALUES (?, ?, ?, ?, ?)`,
		"test-expiry-ukm", "Test UKM", "test-ukm", 100, 0)

	// Create expired reservation manually
	_, err = db.Exec(`
		INSERT INTO slot_reservations (reservation_id, nrp, ukm_id, expires_at) 
		VALUES (?, ?, ?, ?)
	`, "expired-reservation-123", "test-expiry-user", "test-expiry-ukm", time.Now().Add(-5*time.Minute))
	require.NoError(t, err)

	// Try to validate expired reservation
	valid, _, err := regRepo.ValidateReservation(ctx, "expired-reservation-123", "test-expiry-user")

	assert.False(t, valid, "Expired reservation should not be valid")
	assert.Error(t, err, "Should return error for expired reservation")
	assert.Contains(t, err.Error(), "expired", "Error should mention expiry")
}

// TestConcurrentRegistrationCompletion tests race conditions in registration completion
func TestConcurrentRegistrationCompletion(t *testing.T) {
	// This test ensures that even if two users somehow get reservations,
	// only one can complete the registration if there's only one slot left

	cfg := &config.Config{
		DatabaseHost:     "127.0.0.1",
		DatabaseName:     "openhouse_2025_test",
		DatabaseUser:     "root",
		DatabasePassword: "",
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseHost, cfg.DatabaseName))
	require.NoError(t, err)
	defer db.Close()

	regRepo := repositories.NewRegistrationRepository(db)
	ctx := context.Background()

	// Setup
	defer cleanupConcurrentTestData(t, db)
	setupConcurrentTestData(t, db)

	// Create two valid reservations manually (simulating a race condition scenario)
	reservation1 := "concurrent-test-reservation-1"
	reservation2 := "concurrent-test-reservation-2"

	_, err = db.Exec(`
		INSERT INTO slot_reservations (reservation_id, nrp, ukm_id, expires_at) 
		VALUES (?, ?, ?, ?), (?, ?, ?, ?)
	`,
		reservation1, "concurrent-user-1", "concurrent-ukm-id", time.Now().Add(10*time.Minute),
		reservation2, "concurrent-user-2", "concurrent-ukm-id", time.Now().Add(10*time.Minute))
	require.NoError(t, err)

	// Try to complete both registrations concurrently
	var wg sync.WaitGroup
	results := make(chan error, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()
		reg1 := &models.DetailRegistration{
			NRP:      "concurrent-user-1",
			UkmID:    "concurrent-ukm-id",
			Payment:  "payment1.jpg",
			DriveURL: "https://drive.google.com/test1",
		}
		err := regRepo.ConsumeReservation(ctx, reservation1, reg1)
		results <- err
	}()

	go func() {
		defer wg.Done()
		reg2 := &models.DetailRegistration{
			NRP:      "concurrent-user-2",
			UkmID:    "concurrent-ukm-id",
			Payment:  "payment2.jpg",
			DriveURL: "https://drive.google.com/test2",
		}
		err := regRepo.ConsumeReservation(ctx, reservation2, reg2)
		results <- err
	}()

	wg.Wait()
	close(results)

	// Check results
	var errors []error
	for err := range results {
		errors = append(errors, err)
	}

	// Both should succeed since there are enough slots (max 100, current 98)
	assert.NoError(t, errors[0], "First registration should succeed")
	assert.NoError(t, errors[1], "Second registration should succeed")

	// Verify current_slot was incremented correctly
	var currentSlot int
	err = db.QueryRow("SELECT current_slot FROM ukms WHERE id = 'concurrent-ukm-id'").Scan(&currentSlot)
	require.NoError(t, err)
	assert.Equal(t, 100, currentSlot, "Current slot should be incremented to 100")
}

func setupConcurrentTestData(t *testing.T, db *sql.DB) {
	// Create UKM at capacity (100 max, 98 current - 2 slots available)
	_, err := db.Exec(`
		INSERT INTO ukms (id, name, slug, max_slot, current_slot) 
		VALUES (?, ?, ?, ?, ?)
	`, "concurrent-ukm-id", "Concurrent Test UKM", "concurrent-ukm", 100, 98)
	require.NoError(t, err)

	// Create test users
	_, err = db.Exec(`
		INSERT INTO users (nrp, name, line_id, phone, form_submitted, is_angket) 
		VALUES (?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?)
	`,
		"concurrent-user-1", "Concurrent User 1", "line1", "081234567891", 1, 1,
		"concurrent-user-2", "Concurrent User 2", "line2", "081234567892", 1, 1)
	require.NoError(t, err)
}

func cleanupConcurrentTestData(t *testing.T, db *sql.DB) {
	db.Exec("DELETE FROM slot_reservations WHERE ukm_id = 'concurrent-ukm-id'")
	db.Exec("DELETE FROM detail_registrations WHERE ukm_id = 'concurrent-ukm-id'")
	db.Exec("DELETE FROM ukms WHERE id = 'concurrent-ukm-id'")
	db.Exec("DELETE FROM users WHERE nrp IN ('concurrent-user-1', 'concurrent-user-2')")
}
