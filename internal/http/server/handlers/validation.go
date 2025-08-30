package handlers

import (
	"errors"
	"net/http"
	"openhouse-2025-api/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentHandler struct {
	paymentSvc *services.ValidationService
}

func NewPaymentHandler(paymentSvc *services.ValidationService) *PaymentHandler {
	return &PaymentHandler{paymentSvc: paymentSvc}
}

type validateRequest struct {
	NRP  string `json:"nrp" binding:"required"`
	UKM  string `json:"ukm" binding:"required"`
	Type string `json:"type" binding:"required,oneof=payment file"` // Must be 'payment' or 'file'
}

func (h *PaymentHandler) Validate(c *gin.Context) {
	var req validateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get admin context from middleware
	adminNRP, exists := c.Get("admin_nrp")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin context not found"})
		return
	}
	adminName, _ := c.Get("admin_name")
	adminUkmID, _ := c.Get("admin_ukm_id")

	adminCtx := services.AdminContext{
		NRP:   adminNRP.(string),
		Name:  adminName.(string),
		UkmID: adminUkmID.(string),
	}

	// Delegate the core logic to the service
	status, err := h.paymentSvc.ProcessValidation(c.Request.Context(), adminCtx, req.Type, req.NRP, req.UKM)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Registration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "An internal error occurred"})
		return
	}

	// Map the service status to a JSON response
	switch status {
	case services.ValidationSuccess:
		c.JSON(http.StatusOK, gin.H{"message": "true"})
	case services.ValidationAlreadyDone:
		c.JSON(http.StatusConflict, gin.H{"message": "false"})
	case services.RejectionAlreadyDone:
		c.JSON(http.StatusForbidden, gin.H{"message": "warning"})
	// Ga dipakai
	//case services.ValidationFileNotReviewed:
	//	c.JSON(http.StatusPreconditionFailed, gin.H{"message": "not_yet"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown validation status"})
	}
}

func (h *PaymentHandler) Reject(c *gin.Context) {
	var req validateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get admin context from middleware
	adminNRP, exists := c.Get("admin_nrp")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin context not found"})
		return
	}
	adminName, _ := c.Get("admin_name")
	adminUkmID, _ := c.Get("admin_ukm_id")

	adminCtx := services.AdminContext{
		NRP:   adminNRP.(string),
		Name:  adminName.(string),
		UkmID: adminUkmID.(string),
	}

	// Delegate the core logic to the service
	status, err := h.paymentSvc.ProcessRejection(c.Request.Context(), adminCtx, req.Type, req.NRP, req.UKM)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "Registration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "An internal error occurred"})
		return
	}

	// Map the service status to a JSON response
	switch status {
	case services.RejectionSuccess:
		c.JSON(http.StatusOK, gin.H{"message": "true"})
	case services.RejectionAlreadyDone:
		c.JSON(http.StatusConflict, gin.H{"message": "false"})
	case services.ValidationAlreadyDone:
		c.JSON(http.StatusConflict, gin.H{"message": "validated"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Unknown validation status"})
	}
}
