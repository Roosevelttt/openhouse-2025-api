package handlers

import (
    "github.com/gin-gonic/gin"
    "openhouse-2025-api/internal/services"

    
)

type SessionHandler struct { service *services.SessionService }
func NewSessionHandler(s *services.SessionService) *SessionHandler { return &SessionHandler{service: s} }

func (h *SessionHandler) GetValues(c *gin.Context) {

    h.service.GetSessionValues(c) 
    
}