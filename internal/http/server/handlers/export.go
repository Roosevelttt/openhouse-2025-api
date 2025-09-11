package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	exportSvc       *services.ExportService
	dailyRecapSvc   *services.DailyRecapService
	googleSheetsSvc *services.GoogleSheetsService
	cfg             *config.Config
}

func NewExportHandler(exportSvc *services.ExportService, dailyRecapSvc *services.DailyRecapService, googleSheetsSvc *services.GoogleSheetsService, cfg *config.Config) *ExportHandler {
	return &ExportHandler{
		exportSvc:       exportSvc,
		dailyRecapSvc:   dailyRecapSvc,
		googleSheetsSvc: googleSheetsSvc,
		cfg:             cfg,
	}
}

func (h *ExportHandler) ExportParticipants(c *gin.Context) {
	dailyReport := c.Query("daily") != "false"

	var buffer *bytes.Buffer
	var err error

	if dailyReport {
		buffer, err = h.exportSvc.GenerateDailyParticipantsExcel(c.Request.Context(), c.GetString("admin_division_slug"), c.GetString("admin_ukm_id"))
	} else {
		buffer, err = h.exportSvc.GenerateParticipantsExcel(c.Request.Context(), c.GetString("admin_division_slug"), c.GetString("admin_ukm_id"))
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
		return
	}

	fileName := fmt.Sprintf("participants_export_%s.xlsx", time.Now().In(h.exportSvc.GetTimezone()).Format("2006-01-02"))
	if dailyReport {
		fileName = fmt.Sprintf("daily_participants_export_%s.xlsx", time.Now().In(h.exportSvc.GetTimezone()).Format("2006-01-02"))
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())
}

// Manual trigger for daily recap
func (h *ExportHandler) TriggerDailyRecap(c *gin.Context) {
	if h.dailyRecapSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Daily recap service not configured"})
		return
	}

	err := h.dailyRecapSvc.RunDailyRecap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to run daily recap: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Daily recap completed successfully"})
}

func (h *ExportHandler) GoogleSheetsOAuthStart(c *gin.Context) {
	if h.googleSheetsSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Google Sheets service not configured"})
		return
	}

	authURL := h.googleSheetsSvc.GetAuthURL("state-token")

	c.Redirect(http.StatusFound, authURL)
}

func (h *ExportHandler) GoogleSheetsOAuthCallback(c *gin.Context) {
	if h.googleSheetsSvc == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Google Sheets service not configured"})
		return
	}

	// Get the authorization code from the query parameters
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	// Exchange the authorization code for a token
	token, err := h.googleSheetsSvc.ExchangeCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to exchange code for token: %v", err)})
		return
	}

	// Save the token to a file for future use
	tokenFilePath := h.cfg.GoogleSheetsTokenFilePath
	err = h.googleSheetsSvc.SaveToken(tokenFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save token: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Google Sheets OAuth successful",
		"token":   token,
	})
}
