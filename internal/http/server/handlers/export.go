package handlers

import (
	"fmt"
	"net/http"
	"openhouse-2025-api/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	exportSvc *services.ExportService
}

func NewExportHandler(exportSvc *services.ExportService) *ExportHandler {
	return &ExportHandler{exportSvc: exportSvc}
}

func (h *ExportHandler) ExportParticipants(c *gin.Context) {
	buffer, err := h.exportSvc.GenerateParticipantsExcel(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
		return
	}

	// Set the correct headers for a file download
	fileName := fmt.Sprintf("participants_export_%s.xlsx", time.Now().Format("2006-01-02"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())
}
