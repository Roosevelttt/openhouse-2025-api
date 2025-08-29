package server

import (
	"log"
	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/server/handlers"
	"openhouse-2025-api/internal/repositories"
	"openhouse-2025-api/internal/services"
)

type Server struct {
	Session      *handlers.SessionHandler
	Auth         *handlers.AuthHandler
	Ukm          *handlers.UkmHandler
	Participants *handlers.ParticipantsHandler
	Validation   *handlers.PaymentHandler
	Export       *handlers.ExportHandler
}

func NewServer(cfg *config.Config) *Server {
	// Call our new connection function here
	sqlDB, gormDB, err := repositories.NewDatabaseConnections(cfg)
	if err != nil {
		// Since this is the root, panicking on a failed DB connection is acceptable.
		log.Fatalf("Could not connect to the database: %v", err)
	}

	// --- Repositories that use RAW SQL get the *sql.DB ---
	userRepo := repositories.NewUserRepository(sqlDB)
	adminRepo := repositories.NewAdminRepository(sqlDB)
	ukmRepo := repositories.NewUkmRepository(gormDB) // Pake GORM
	regRepo := repositories.NewRegistrationRepository(sqlDB)
	participantsRepo := repositories.NewParticipantsRepository(sqlDB)

	// --- New repositories that use GORM get the *gorm.DB ---
	validationRepo := repositories.NewValidationRepository(gormDB)

	// --- Services ---
	sessionSvc := services.NewSessionService()
	authSvc := services.NewAuthService(cfg, userRepo, adminRepo)
	ukmSvc := services.NewUkmService(ukmRepo, regRepo)
	participantsSvc := services.NewParticipantsService(participantsRepo)
	mailSvc := services.NewMailService(cfg)
	validationSvc := services.NewValidationService(gormDB, validationRepo, userRepo, ukmRepo, mailSvc)
	exportSvc := services.NewExportService(participantsRepo)

	// --- Handlers ---
	return &Server{
		Session:      handlers.NewSessionHandler(sessionSvc),
		Auth:         handlers.NewAuthHandler(authSvc),
		Ukm:          handlers.NewUkmHandler(ukmSvc),
		Participants: handlers.NewParticipantsHandler(participantsSvc),
		Validation:   handlers.NewPaymentHandler(validationSvc),
		Export:       handlers.NewExportHandler(exportSvc),
	}
}
