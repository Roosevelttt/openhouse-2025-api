package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
)

type DailyRecapService struct {
	googleSheetsSvc  *GoogleSheetsService
	participantsRepo *repositories.ParticipantsRepository
	spreadsheetID    string
	timeZone         *time.Location
}

func NewDailyRecapService(
	cfg *config.Config,
	googleSheetsSvc *GoogleSheetsService,
	participantsRepo *repositories.ParticipantsRepository,
) *DailyRecapService {
	// Set timezone to Asia/Jakarta (WIB) or UTC as default
	timeZone, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("Failed to load timezone Asia/Jakarta, using UTC: %v", err)
		timeZone = time.UTC
	}

	return &DailyRecapService{
		googleSheetsSvc:  googleSheetsSvc,
		participantsRepo: participantsRepo,
		spreadsheetID:    cfg.GoogleSheetsSpreadsheetID,
		timeZone:         timeZone,
	}
}

func (d *DailyRecapService) RunDailyRecap() error {
	log.Println("Starting daily participant data recap to Google Sheets")

	if d.spreadsheetID == "" {
		return fmt.Errorf("google sheets spreadsheet ID not configured")
	}

	// Fetch all participants (no filtering by division)
	participants, err := d.participantsRepo.List(context.Background(), "it", "") // "it" division gets all participants
	if err != nil {
		return fmt.Errorf("failed to fetch participants: %v", err)
	}

	values := d.prepareDataForSheets(participants)

	// Get today's date for the sheet name
	now := time.Now().In(d.timeZone)
	sheetName := now.Format("02-Jan-2006") // Format: DD-MMM-YYYY

	// Create or clear the sheet for today
	err = d.prepareSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to prepare sheet: %v", err)
	}

	// Write data to the spreadsheet
	rangeName := fmt.Sprintf("%s!A1", sheetName)
	err = d.googleSheetsSvc.WriteDataToSheet(d.spreadsheetID, rangeName, values)
	if err != nil {
		return fmt.Errorf("failed to write to spreadsheet: %v", err)
	}

	return nil
}

func (d *DailyRecapService) prepareSheet(sheetName string) error {
	// Check if the sheet already exists
	sheetExists, err := d.googleSheetsSvc.SheetExists(d.spreadsheetID, sheetName)
	if err != nil {
		return fmt.Errorf("failed to check if sheet exists: %v", err)
	}

	if sheetExists {
		// Clear the existing sheet data
		rangeName := fmt.Sprintf("%s!A:Z", sheetName)
		err = d.googleSheetsSvc.ClearSheetData(d.spreadsheetID, rangeName)
		if err != nil {
			return fmt.Errorf("failed to clear sheet data: %v", err)
		}
	} else {
		// Create a new sheet
		err = d.googleSheetsSvc.CreateSheet(d.spreadsheetID, sheetName)
		if err != nil {
			return fmt.Errorf("failed to create sheet: %v", err)
		}
	}

	return nil
}

func (d *DailyRecapService) prepareDataForSheets(participants []models.Participant) [][]interface{} {
	headers := []interface{}{
		"NRP", "Nama", "UKM", "Line ID", "Phone", "Status Payment", "Timestamp",
	}

	var values [][]interface{}
	values = append(values, headers)

	for _, p := range participants {
		// Map payment status to readable format
		paymentStatus := mapDailyRecapPaymentStatus(p.PaymentValidated)

		// Format timestamp
		var timestamp string
		if p.CreatedAt != nil {
			// Convert to configured timezone and format nicely
			localTime := p.CreatedAt.In(d.timeZone)
			timestamp = localTime.Format("02-Jan-2006 15:04:05") // Format: DD-MMM-YYYY HH:MM:SS
		} else {
			timestamp = ""
		}

		row := []interface{}{
			p.NRP,
			p.Name,
			p.UkmName,
			p.LineID,
			p.Phone,
			paymentStatus,
			timestamp,
		}

		values = append(values, row)
	}

	return values
}

func mapDailyRecapPaymentStatus(status int) string {
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
