package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"openhouse-2025-api/internal/services"
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
	
	participants, err := h.service.ListParticipants(c.Request.Context())
	if err != nil {
		log.Printf("Error getting participants: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Found %d participants", len(participants))
	c.JSON(http.StatusOK, participants)
}
