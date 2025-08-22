package server

import (
	"log"
	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/server/handlers"
	"openhouse-2025-api/internal/repositories"
	"openhouse-2025-api/internal/services"
)

type Server struct {
	Auth         *handlers.AuthHandler
	Ukm          *handlers.UkmHandler
	Participants *handlers.ParticipantsHandler
	Payment      *handlers.PaymentHandler
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
	ukmRepo := repositories.NewUkmRepository(gormDB) // Pake GORM
	regRepo := repositories.NewRegistrationRepository(sqlDB)
	participantsRepo := repositories.NewParticipantsRepository(sqlDB)

	// --- New repositories that use GORM get the *gorm.DB ---
	paymentRepo := repositories.NewPaymentRepository(gormDB)

	// --- Services ---
	authSvc := services.NewAuthService(cfg, userRepo)
	ukmSvc := services.NewUkmService(ukmRepo, regRepo)
	participantsSvc := services.NewParticipantsService(participantsRepo)
	paymentSvc := services.NewPaymentService(paymentRepo, userRepo, ukmRepo)

	// --- Handlers ---
	return &Server{
		Auth:         handlers.NewAuthHandler(authSvc),
		Ukm:          handlers.NewUkmHandler(ukmSvc),
		Participants: handlers.NewParticipantsHandler(participantsSvc),
		Payment:      handlers.NewPaymentHandler(paymentSvc),
	}
}
