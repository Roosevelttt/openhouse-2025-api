package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"openhouse-2025-api/internal/config"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleSheetsService struct {
	config *oauth2.Config
	token  *oauth2.Token
}

func NewGoogleSheetsService(cfg *config.Config) *GoogleSheetsService {
	clientID := cfg.GoogleClientID
	clientSecret := cfg.GoogleClientSecret
	redirectURL := cfg.GoogleSheetsRedirectURL

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{sheets.SpreadsheetsScope},
		Endpoint:     google.Endpoint,
	}

	return &GoogleSheetsService{
		config: config,
	}
}

func (g *GoogleSheetsService) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *GoogleSheetsService) ExchangeCode(code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	g.token = token
	return token, nil
}

func (g *GoogleSheetsService) SaveToken(path string) error {
	if g.token == nil {
		return fmt.Errorf("no token to save")
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(g.token)
}

func (g *GoogleSheetsService) LoadToken(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to read token file: %v", err)
	}
	defer file.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(token)
	if err != nil {
		return fmt.Errorf("unable to decode token file: %v", err)
	}

	g.token = token
	return nil
}

func (g *GoogleSheetsService) GetClient() (*http.Client, error) {
	if g.token == nil {
		return nil, fmt.Errorf("no token available")
	}
	return g.config.Client(context.Background(), g.token), nil
}

func (g *GoogleSheetsService) CreateSpreadsheet(title string) (*sheets.Spreadsheet, error) {
	client, err := g.GetClient()
	if err != nil {
		return nil, err
	}

	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}

	resp, err := srv.Spreadsheets.Create(spreadsheet).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create spreadsheet: %v", err)
	}

	return resp, nil
}

func (g *GoogleSheetsService) CreateSheet(spreadsheetID string, sheetName string) error {
	client, err := g.GetClient()
	if err != nil {
		return err
	}

	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	requests := []*sheets.Request{
		{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: sheetName,
				},
			},
		},
	}

	batchUpdate := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdate).Do()
	if err != nil {
		return fmt.Errorf("unable to create sheet: %v", err)
	}

	return nil
}

func (g *GoogleSheetsService) ClearSheetData(spreadsheetID string, rangeName string) error {
	client, err := g.GetClient()
	if err != nil {
		return err
	}

	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	_, err = srv.Spreadsheets.Values.Clear(spreadsheetID, rangeName, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return fmt.Errorf("unable to clear sheet data: %v", err)
	}

	return nil
}

func (g *GoogleSheetsService) SheetExists(spreadsheetID string, sheetName string) (bool, error) {
	client, err := g.GetClient()
	if err != nil {
		return false, err
	}

	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return false, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	resp, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return false, fmt.Errorf("unable to get spreadsheet info: %v", err)
	}

	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == sheetName {
			return true, nil
		}
	}

	return false, nil
}

func (g *GoogleSheetsService) WriteDataToSheet(spreadsheetID string, rangeName string, values [][]interface{}) error {
	client, err := g.GetClient()
	if err != nil {
		return err
	}

	srv, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	rb := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         values,
	}

	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, rangeName, rb).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to update sheet: %v", err)
	}

	return nil
}
