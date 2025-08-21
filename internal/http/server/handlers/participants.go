package handlers

import (
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
	participants, err := h.service.ListParticipants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, participants)
}
