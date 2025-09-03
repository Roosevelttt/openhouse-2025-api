package handlers

import (
	"log"
	"net/http"
	"time"

	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
	"openhouse-2025-api/internal/services"

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

	var reg models.DetailRegistration
	if err := c.ShouldBindJSON(&reg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nrp := c.GetString("user_nrp")
	if nrp == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not logged in"})
		return
	}

	reg.NRP = nrp

	err := h.service.RegisterWithReservation(c.Request.Context(), reservationID, &reg)
	if err != nil {
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
		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"message":        "Slot berhasil dipesan, silakan lanjutkan pembayaran dalam 5 menit",
			"reservation_id": result.ReservationID,
			"expires_at":     result.ExpiresAt,
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
