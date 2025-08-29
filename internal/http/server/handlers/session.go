package handlers

import (
	"openhouse-2025-api/internal/services"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct{ service *services.SessionService }

func NewSessionHandler(s *services.SessionService) *SessionHandler {
	return &SessionHandler{service: s}
}

func (h *SessionHandler) GetValues(c *gin.Context) {

	h.service.GetSessionValues(c)

}

func (h *SessionHandler) DebugSession(c *gin.Context) {
	h.service.DebugSession(c)
}