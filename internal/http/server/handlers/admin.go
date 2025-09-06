package handlers

import (
	"net/http"

	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	repo *repositories.AdminRepository
}

func NewAdminHandler(r *repositories.AdminRepository) *AdminHandler {
	return &AdminHandler{repo: r}
}

func (h *AdminHandler) List(c *gin.Context) {
	admins, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, admins)
}

type CreateAdminRequest struct {
	Name       string `json:"name" binding:"required"`
	NRP        string `json:"nrp" binding:"required"`
	Field      string `json:"field" binding:"required,email"`
	DivisionID string `json:"division_id"`
	UkmID      string `json:"ukm_id"`
}

func (h *AdminHandler) Create(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// field value based on UKM and Division
	field := "Panitia"
	if req.UkmID != "" && req.DivisionID == "" {
		field = "Ketua UKM"
	}

	// Create admin model
	admin := &models.Admin{
		Name:  req.Name,
		NRP:   req.NRP,
		Field: field,
	}

	// Set Division ID if provided
	if req.DivisionID != "" {
		admin.DivisionID = &req.DivisionID
	}

	// Set UKM ID if provided
	if req.UkmID != "" {
		admin.UkmID = &req.UkmID
	}

	if err := h.repo.Create(c.Request.Context(), admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, admin)
}

type UpdateAdminRequest struct {
	Name       string `json:"name" binding:"required"`
	NRP        string `json:"nrp" binding:"required"`
	Field      string `json:"field" binding:"required,email"`
	DivisionID string `json:"division_id"`
	UkmID      string `json:"ukm_id"`
}

func (h *AdminHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	var req UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	field := "Panitia"
	if req.UkmID != "" && req.DivisionID == "" {
		field = "Ketua UKM"
	}

	admin := &models.Admin{
		Name:  req.Name,
		NRP:   req.NRP,
		Field: field,
	}

	if req.DivisionID != "" {
		admin.DivisionID = &req.DivisionID
	}

	if req.UkmID != "" {
		admin.UkmID = &req.UkmID
	}

	if err := h.repo.Update(c.Request.Context(), id, admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, admin)
}

func (h *AdminHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin deleted successfully"})
}
