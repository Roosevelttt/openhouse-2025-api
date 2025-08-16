package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"openhouse-2025-api/internal/services"
)

type UkmHandler struct { service *services.UkmService }
func NewUkmHandler(s *services.UkmService) *UkmHandler { return &UkmHandler{service: s} }

func (h *UkmHandler) List(c *gin.Context) {
	ukms, err := h.service.List(c.Request.Context())
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(http.StatusOK, ukms)
}

