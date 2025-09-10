package main

import (
	"log"
	"net/http"
	"time"

	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/router"
	"openhouse-2025-api/internal/repositories"
	"openhouse-2025-api/internal/services"
)

func main() {
	cfg := config.Load()

	go runDailyRecapScheduler(cfg)

	r := router.New(cfg)

	addr := ":" + cfg.HTTPPort
	log.Printf("openhouse-2025-api starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func runDailyRecapScheduler(cfg *config.Config) {
	time.Sleep(10 * time.Second)

	sqlDB, _, err := repositories.NewDatabaseConnections(cfg)
	if err != nil {
		log.Printf("Failed to connect to database for scheduler: %v", err)
		return
	}
	defer sqlDB.Close()

	participantsRepo := repositories.NewParticipantsRepository(sqlDB)
	googleSheetsSvc := services.NewGoogleSheetsService(cfg)
	if googleSheetsSvc == nil {
		log.Println("Google Sheets service not configured")
		return
	}

	// Load the saved token if it exists
	err = googleSheetsSvc.LoadToken(cfg.GoogleSheetsTokenFilePath)
	if err != nil {
		log.Printf("Failed to load Google Sheets token: %v", err)
	}

	dailyRecapSvc := services.NewDailyRecapService(cfg, googleSheetsSvc, participantsRepo)
	if dailyRecapSvc == nil {
		log.Println("Daily recap service not configured")
		return
	}

	//test
	log.Println("Running initial daily recap...")
	if err := dailyRecapSvc.RunDailyRecap(); err != nil {
		log.Printf("Initial daily recap failed: %v", err)
	} else {
		log.Println("Initial daily recap completed successfully")
	}

	// Schedule daily recap to run every day at the configured time (default: 11:59 PM WIB)
	for {
		now := time.Now()
		// Set timezone to Asia/Jakarta (WIB)
		wib, _ := time.LoadLocation("Asia/Jakarta")
		nowWIB := now.In(wib)

		nextRun := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day(), cfg.GoogleSheetsDailyRecapHour, cfg.GoogleSheetsDailyRecapMinute, 0, 0, wib)
		if nowWIB.After(nextRun) {
			// If it's already past 11:59 PM today, schedule for tomorrow
			nextRun = nextRun.Add(24 * time.Hour)
		}

		// Wait until next run time
		duration := nextRun.Sub(nowWIB)
		log.Printf("Next daily recap scheduled for: %s (in %s)", nextRun.Format("2006-01-02 15:04:05 MST"), duration.String())

		time.Sleep(duration)

		// Run daily recap
		log.Println("Running scheduled daily recap...")
		if err := dailyRecapSvc.RunDailyRecap(); err != nil {
			log.Printf("Scheduled daily recap failed: %v", err)
		} else {
			log.Println("Scheduled daily recap completed successfully")
		}

		time.Sleep(2 * time.Minute)
	}
}
