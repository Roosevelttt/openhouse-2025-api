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
	User         *handlers.UserHandler
	Ukm          *handlers.UkmHandler
	Participants *handlers.ParticipantsHandler
	Validation   *handlers.PaymentHandler
	Export       *handlers.ExportHandler
	Groupchat    *handlers.GroupchatHandler
	Registration *handlers.RegistrationHandler
	Admin        *handlers.AdminHandler
	Division     *handlers.DivisionHandler
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
	divisionRepo := repositories.NewDivisionRepository(sqlDB)
	ukmRepo := repositories.NewUkmRepository(gormDB) // Pake GORM
	regRepo := repositories.NewRegistrationRepository(sqlDB)
	participantsRepo := repositories.NewParticipantsRepository(sqlDB)

	// --- New repositories that use GORM get the *gorm.DB ---
	validationRepo := repositories.NewValidationRepository(gormDB)

	// --- Services ---
	sessionSvc := services.NewSessionService()
	authSvc := services.NewAuthService(cfg, userRepo, adminRepo)
	userSvc := services.NewUserService(userRepo)
	ukmSvc := services.NewUkmService(ukmRepo, regRepo)
	participantsSvc := services.NewParticipantsService(participantsRepo, regRepo)
	mailSvc := services.NewMailService(cfg)
	validationSvc := services.NewValidationService(gormDB, validationRepo, userRepo, ukmRepo, mailSvc)
	exportSvc := services.NewExportService(participantsRepo)
	groupchatSvc := services.NewGroupchatService(ukmRepo)

	// --- Handlers ---
	return &Server{
		Session:      handlers.NewSessionHandler(sessionSvc),
		Auth:         handlers.NewAuthHandler(authSvc),
		User:         handlers.NewUserHandler(userSvc),
		Ukm:          handlers.NewUkmHandler(ukmSvc),
		Participants: handlers.NewParticipantsHandler(participantsSvc, regRepo),
		Validation:   handlers.NewPaymentHandler(validationSvc),
		Export:       handlers.NewExportHandler(exportSvc),
		Groupchat:    handlers.NewGroupchatHandler(groupchatSvc),
		Registration: handlers.NewRegistrationHandler(regRepo),
		Admin:        handlers.NewAdminHandler(adminRepo),
		Division:     handlers.NewDivisionHandler(divisionRepo),
	}
}
