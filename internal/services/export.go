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
	timeZone         *time.Location
}

func NewExportService(participantsRepo *repositories.ParticipantsRepository) *ExportService {
	timeZone, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		timeZone = time.UTC
	}

	return &ExportService{
		participantsRepo: participantsRepo,
		timeZone:         timeZone,
	}
}

func (s *ExportService) GetTimezone() *time.Location {
	return s.timeZone
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

		// Check if p.CreatedAt is nil before calling .Format()
		var createdAtStr string
		if p.CreatedAt != nil {
			localTime := p.CreatedAt.In(s.timeZone)
			createdAtStr = localTime.Format("02-Jan-2006 15:04:05") // Format: DD-MMM-YYYY HH:MM:SS
		} else {
			createdAtStr = "N/A" // Or "" for an empty cell
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), p.NRP)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), p.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), p.LineID)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), p.Phone)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), p.UkmName)
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

func (s *ExportService) GenerateDailyParticipantsExcel(ctx context.Context, adminDivisionSlug string, adminUkmID string) (*bytes.Buffer, error) {
	// 1. Fetch all participant data
	participants, err := s.participantsRepo.List(ctx, adminDivisionSlug, adminUkmID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch participants: %w", err)
	}

	// 2. Group participants by date
	participantsByDate := make(map[string][]models.Participant)
	for _, p := range participants {
		var dateKey string
		if p.CreatedAt != nil {
			localTime := p.CreatedAt.In(s.timeZone)
			dateKey = localTime.Format("2006-01-02") // Format: YYYY-MM-DD
		} else {
			dateKey = "unknown"
		}
		participantsByDate[dateKey] = append(participantsByDate[dateKey], p)
	}

	// 3. Create a new Excel file in memory
	f := excelize.NewFile()

	// 4. Create a sheet for each date
	sheetIndex := 0
	for dateKey, participantsForDate := range participantsByDate {
		var sheetName string
		if dateKey == "unknown" {
			sheetName = "Unknown Date"
		} else {
			if parsedDate, err := time.Parse("2006-01-02", dateKey); err == nil {
				sheetName = parsedDate.In(s.timeZone).Format("02-Jan-2006") // Format: DD-MMM-YYYY
			} else {
				sheetName = dateKey
			}
		}

		// Create a new sheet
		index, err := f.NewSheet(sheetName)
		if err != nil {
			return nil, err
		}

		// Define and write the header row
		headers := []string{
			"NRP", "Name", "Line ID", "Phone", "UKM",
			"Payment Validated", "Registered At (Asia/Jakarta)",
		}
		for i, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Write the data rows for this date
		for i, p := range participantsForDate {
			rowNum := i + 2 // Start from row 2

			// Format timestamp in Asia/Jakarta timezone
			var createdAtStr string
			if p.CreatedAt != nil {
				localTime := p.CreatedAt.In(s.timeZone)
				createdAtStr = localTime.Format("02-Jan-2006 15:04:05") // Format: DD-MMM-YYYY HH:MM:SS
			} else {
				createdAtStr = "N/A"
			}

			f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), p.NRP)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), p.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), p.LineID)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), p.Phone)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), p.UkmName)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), mapPaymentStatus(p.PaymentValidated))
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), createdAtStr)
		}

		// Set this as the active sheet for the first date
		if sheetIndex == 0 {
			f.SetActiveSheet(index)
		}
		sheetIndex++
	}

	// Delete the default "Sheet1" if we created other sheets
	if len(participantsByDate) > 0 {
		f.DeleteSheet("Sheet1")
	}

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
