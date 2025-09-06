package handlers

import (
	"net/http"

	"openhouse-2025-api/internal/repositories"

	"github.com/gin-gonic/gin"
)

type DivisionHandler struct {
	repo *repositories.DivisionRepository
}

func NewDivisionHandler(r *repositories.DivisionRepository) *DivisionHandler {
	return &DivisionHandler{repo: r}
}

func (h *DivisionHandler) List(c *gin.Context) {
	divisions, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, divisions)
}
