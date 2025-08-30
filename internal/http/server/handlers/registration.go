package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"

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

func (h *RegistrationHandler) Create(c *gin.Context) {
	// Parse form data
	err := c.Request.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// Get form values
	ukmID := c.PostForm("ukm_id")
	nrp := c.PostForm("nrp")
	code := c.PostForm("code")
	driveURL := c.PostForm("drive_url")

	// Debug: Print received values
	fmt.Printf("Received values - UkmID: %s, NRP: %s, Code: %s, DriveURL: %s\n", ukmID, nrp, code, driveURL)

	// Validate required fields
	if ukmID == "" || nrp == "" || code == "" || driveURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Handle file upload
	file, header, err := c.Request.FormFile("payment")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment file is required"})
		return
	}
	defer file.Close()

	// Validate file type
	allowedTypes := []string{".jpg", ".jpeg", ".png", ".pdf"}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	isValidType := false
	for _, allowedType := range allowedTypes {
		if ext == allowedType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPG, PNG, and PDF files are allowed"})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "uploads/payments"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s_%s%s", nrp, ukmID, code, ext)
	filePath := filepath.Join(uploadsDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create registration record
	registration := &models.DetailRegistration{
		NRP:      nrp,
		UkmID:    ukmID,
		Payment:  filename, // Store filename instead of integer
		// Code:     code,
		DriveURL: driveURL,
	}

	// Debug: Print the data being inserted
	fmt.Printf("Creating registration: NRP=%s, UkmID=%s, Payment=%s, Code=%s, DriveURL=%s\n",
		nrp, ukmID, filename, code, driveURL)

	if err := h.registrationRepo.Create(c.Request.Context(), registration); err != nil {
		// Clean up uploaded file if database insert fails
		os.Remove(filePath)
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
	})
}
