package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
	"openhouse-2025-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	registrationRepo *repositories.RegistrationRepository
}

func NewRegistrationHandler(registrationRepo *repositories.RegistrationRepository) *RegistrationHandler {
	return &RegistrationHandler{
		registrationRepo: registrationRepo,
	}
}

func (h *RegistrationHandler) Test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Registration endpoint is working"})
}

func (h *RegistrationHandler) isRegistrationClosed() bool {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatalf("Could not load timezone: %v", err)
	}

	deadline := time.Date(2025, 9, 25, 0, 0, 0, 0, loc)
	now := time.Now().In(loc)

	return now.After(deadline)
}

func (h *RegistrationHandler) Create(c *gin.Context) {
	if h.isRegistrationClosed() {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Registration is closed.",
		})
		return
	}

	// Parse form data
	err := c.Request.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// Get form values
	ukmID := c.PostForm("ukm_id")
	nrp := c.PostForm("nrp")
	driveURL := c.PostForm("drive_url")

	// Debug: Print received values
	fmt.Printf("Received values - UkmID: %s, NRP: %s, DriveURL: %s\n", ukmID, nrp, driveURL)

	// Validate required fields
	if ukmID == "" || nrp == "" || driveURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Handle file upload
	file, header, err := c.Request.FormFile("payment")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment file is required", "details": err.Error()})
		return
	}
	defer file.Close()

	prefix := fmt.Sprintf("%s_%s_payment", nrp, ukmID)
	result, err := utils.SavePaymentFile(file, header, prefix)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to save payment file", "details": err.Error()})
		return
	}

	// Create registration record
	registration := &models.DetailRegistration{
		NRP:      nrp,
		UkmID:    ukmID,
		Payment:  &result.FileName, // Store filename as pointer to string
		DriveURL: driveURL,
	}

	// Debug: Print the data being inserted
	fmt.Printf("Creating registration: NRP=%s, UkmID=%s, Payment=%s, DriveURL=%s\n",
		nrp, ukmID, result.FileName, driveURL)

	if err := h.registrationRepo.Create(c.Request.Context(), registration); err != nil {
		// Log the actual error for debugging
		fmt.Printf("Database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create registration",
			"details": err.Error(), // Include the actual error for debugging
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration created successfully",
		"id":      registration.ID,
		"file":    result.RelativeURL,
	})
}
