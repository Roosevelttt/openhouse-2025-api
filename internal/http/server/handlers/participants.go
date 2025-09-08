package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
	"openhouse-2025-api/internal/services"
	"openhouse-2025-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type ParticipantsHandler struct {
	service          *services.ParticipantsService
	registrationRepo *repositories.RegistrationRepository
}

func NewParticipantsHandler(s *services.ParticipantsService, regRepo *repositories.RegistrationRepository) *ParticipantsHandler {
	return &ParticipantsHandler{
		service:          s,
		registrationRepo: regRepo,
	}
}

func (h *ParticipantsHandler) List(c *gin.Context) {
	log.Printf("=== ParticipantsHandler.List called ===")
	log.Printf("Admin NRP from context: %v", c.GetString("admin_nrp"))
	log.Printf("Admin Name from context: %v", c.GetString("admin_name"))
	log.Printf("Admin UKM ID from context: %v", c.GetString("admin_ukm_id"))
	log.Printf("Admin Division ID from context: %v", c.GetString("admin_division_id"))
	log.Printf("Admin Division Slug from context: %v", c.GetString("admin_division_slug"))

	participants, err := h.service.ListParticipants(c.Request.Context(), c.GetString("admin_division_slug"), c.GetString("admin_ukm_id"))
	if err != nil {
		log.Printf("Error getting participants: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Found %d participants", len(participants))
	c.JSON(http.StatusOK, participants)
}

// ReserveSlot reserves a slot for registration
func (h *ParticipantsHandler) ReserveSlot(c *gin.Context) {
	var req models.ReserveSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nrp := c.GetString("user_nrp")
	if nrp == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
		return
	}

	response, err := h.service.ReserveSlot(c.Request.Context(), nrp, req.UkmID)
	if err != nil {
		if err.Error() == "no slots available" {
			c.JSON(http.StatusConflict, gin.H{"error": "No slots available"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RegisterWithReservation creates registration using a reservation
func (h *ParticipantsHandler) RegisterWithReservation(c *gin.Context) {
	reservationID := c.Param("reservationId")
	if reservationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation ID is required"})
		return
	}

	// Parse multipart form data
	err := c.Request.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// Get form values
	ukmID := c.PostForm("ukm_id")
	driveURL := c.PostForm("drive_url")

	nrp := c.GetString("user_nrp")
	if nrp == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
		return
	}

	// Validate required fields
	if ukmID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UKM ID is required"})
		return
	}

	var filename *string = nil

	// Handle file upload (optional for free UKMs)
	file, header, err := c.Request.FormFile("payment")
	if err == nil {
		// File was uploaded, validate and save it
		defer file.Close()

		// Validate file type
		allowedTypes := []string{".jpg", ".jpeg", ".png", ".pdf"}
		ext := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(header.Filename, " ")))
		ext = ext[strings.LastIndex(ext, "."):]

		isValidType := false
		for _, allowedType := range allowedTypes {
			if ext == allowedType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPG, PNG, and PDF files are allowed"})
			return
		}

		// Get UKM name for the filename prefix
		// Assuming we can get UKM name from the service or we'll use UKM ID
		ukmName := ukmID // Fallback to UKM ID if name not available

		// Use SaveUploadedFile with proper naming format: nrp_ukmname_payment
		prefix := fmt.Sprintf("%s_%s_payment", nrp, ukmName)
		result, err := utils.SaveUploadedFile(file, header, "uploads/payments", prefix)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
			return
		}

		filename = &result.FileName
	}

	// Create registration record
	reg := &models.DetailRegistration{
		NRP:      nrp,
		UkmID:    ukmID,
		Payment:  filename, // Will be nil for free UKMs or if no file uploaded
		DriveURL: driveURL,
	}

	err = h.service.RegisterWithReservation(c.Request.Context(), reservationID, reg)
	if err != nil {
		// Clean up uploaded file if database insert fails
		if filename != nil {
			// Note: We'd need the full file path to clean up, but for now we'll rely on the service to handle cleanup
		}

		if err.Error() == "reservation not found" || err.Error() == "reservation has expired" {
			c.JSON(http.StatusGone, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Registration successful",
		"registration":   reg,
		"reservation_id": reservationID,
	})
}

// AccessPaymentPage handles when user tries to access payment page
func (h *ParticipantsHandler) AccessPaymentPage(c *gin.Context) {
	nrp := c.GetString("user_nrp")
	if nrp == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	ukmID := c.Param("ukm_id")
	if ukmID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "UKM ID is required",
		})
		return
	}

	// Try to reserve slot for payment page access
	result, err := h.registrationRepo.ReserveSlotForPayment(c.Request.Context(), nrp, ukmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal memesan slot: " + err.Error(),
		})
		return
	}

	if result != nil {
		// Debug logging
		fmt.Printf("DEBUG: Reservation created with expires_at: %v (UTC: %v)\n", result.ExpiresAt, result.ExpiresAt.UTC())
		fmt.Printf("DEBUG: Current time: %v (UTC: %v)\n", time.Now(), time.Now().UTC())
		fmt.Printf("DEBUG: Time until expiry: %v minutes\n", time.Until(result.ExpiresAt).Minutes())

		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"message":        "Slot berhasil dipesan, silakan lanjutkan pembayaran dalam 5 menit",
			"reservation_id": result.ReservationID,
			"expires_at":     result.ExpiresAt.Format(time.RFC3339), // Explicit RFC3339 formatting
			"current_slot":   result.CurrentSlot,
			"max_slot":       result.MaxSlot,
			"time_remaining": int(time.Until(result.ExpiresAt).Minutes()),
		})
	} else {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "UKM sudah penuh, silakan pilih UKM lain",
		})
	}
}

// CheckUserReservation checks if user has a valid reservation for a UKM
func (h *ParticipantsHandler) CheckUserReservation(c *gin.Context) {
	nrp := c.GetString("user_nrp")
	if nrp == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ukmID := c.Param("ukm_id")
	if ukmID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UKM ID is required"})
		return
	}

	reservation, err := h.service.CheckUserReservation(c.Request.Context(), nrp, ukmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if reservation == nil {
		c.JSON(http.StatusOK, gin.H{
			"has_reservation": false,
		})
		return
	}

	isExpired := time.Now().After(reservation.ExpiresAt)

	c.JSON(http.StatusOK, gin.H{
		"has_reservation": true,
		"reservation_id":  reservation.ReservationID,
		"expires_at":      reservation.ExpiresAt.Format(time.RFC3339),
		"is_expired":      isExpired,
	})
}
