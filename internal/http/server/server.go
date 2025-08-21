package server

import (
	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/server/handlers"
	"openhouse-2025-api/internal/repositories"
	"openhouse-2025-api/internal/services"
)

type Server struct {
	Auth         *handlers.AuthHandler
	Ukm          *handlers.UkmHandler
	Participants *handlers.ParticipantsHandler
}

func NewServer(cfg *config.Config) *Server {
	db := repositories.MustConnectMySQL(cfg)

	userRepo := repositories.NewUserRepository(db)
	ukmRepo := repositories.NewUkmRepository(db)
	regRepo := repositories.NewRegistrationRepository(db)
	participantsRepo := repositories.NewParticipantsRepository(db)

	authSvc := services.NewAuthService(cfg, userRepo)
	ukmSvc := services.NewUkmService(ukmRepo, regRepo)
	participantsSvc := services.NewParticipantsService(participantsRepo)

	return &Server{
		Auth:         handlers.NewAuthHandler(authSvc),
		Ukm:          handlers.NewUkmHandler(ukmSvc),
		Participants: handlers.NewParticipantsHandler(participantsSvc),
	}
}
