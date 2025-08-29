package handlers

import (
	"log"
	"net/http"

	"openhouse-2025-api/internal/services"

	"github.com/gin-gonic/gin"
)

type ParticipantsHandler struct {
	service *services.ParticipantsService
}

func NewParticipantsHandler(s *services.ParticipantsService) *ParticipantsHandler {
	return &ParticipantsHandler{service: s}
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
