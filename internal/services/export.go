package services

import (
	"bytes"
	"context"
	"fmt"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
	"time"

	"github.com/xuri/excelize/v2"
)

type ExportService struct {
	participantsRepo *repositories.ParticipantsRepository
}

func NewExportService(participantsRepo *repositories.ParticipantsRepository) *ExportService {
	return &ExportService{participantsRepo: participantsRepo}
}

func (s *ExportService) ListParticipantsForExport(ctx context.Context, adminDivisionSlug string, adminUkmID string) ([]models.Participant, error) {
	return s.participantsRepo.List(ctx, adminDivisionSlug, adminUkmID)
}

// GenerateParticipantsExcel creates an Excel file in memory and returns it as a byte buffer.
func (s *ExportService) GenerateParticipantsExcel(ctx context.Context, adminDivisionSlug string, adminUkmID string) (*bytes.Buffer, error) {
	// 1. Fetch all participant data
	participants, err := s.participantsRepo.List(ctx, adminDivisionSlug, adminUkmID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch participants: %w", err)
	}

	// 2. Create a new Excel file in memory
	f := excelize.NewFile()
	sheetName := "Participants"
	// Create a new sheet.
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	// 3. Define and write the header row
	headers := []string{
		"NRP", "Name", "Line ID", "Phone", "UKM",
		"Payment Validated", "Registered At",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// 4. Write the data rows
	for i, p := range participants {
		rowNum := i + 2 // Start from row 2

		// Gapakai file payment
		// Check if p.Payment is nil before dereferencing it
		//var paymentStr string
		//if p.Payment != nil {
		//	paymentStr = *p.Payment
		//} else {
		//	paymentStr = "N/A" // Or "" for an empty cell
		//}

		// Check if p.CreatedAt is nil before calling .Format()
		var createdAtStr string
		if p.CreatedAt != nil {
			createdAtStr = p.CreatedAt.Format(time.RFC1123)
		} else {
			createdAtStr = "N/A" // Or "" for an empty cell
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), p.NRP)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), p.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), p.LineID)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), p.Phone)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), p.UkmName)
		//f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), paymentStr) // Assuming Payment is a *string
		//f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), mapFileStatus(p.FileValidated))
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), mapPaymentStatus(p.PaymentValidated))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), createdAtStr)
	}

	// Set active sheet of the workbook.
	f.SetActiveSheet(index)
	// Delete the default "Sheet1".
	f.DeleteSheet("Sheet1")

	// 5. Write the file to a buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel to buffer: %w", err)
	}

	return buffer, nil
}

func mapPaymentStatus(status int) string {
	switch status {
	case 0:
		return "Pending"
	case 1:
		return "Validated"
	case 2:
		return "Rejected"
	default:
		return "Unknown"
	}
}
