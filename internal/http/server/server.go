package server

import (
	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/server/handlers"
	"openhouse-2025-api/internal/repositories"
	"openhouse-2025-api/internal/services"
)

type Server struct {
	Session *handlers.SessionHandler
	Auth *handlers.AuthHandler
	Ukm  *handlers.UkmHandler
}

func NewServer(cfg *config.Config) *Server {
	db := repositories.MustConnectMySQL(cfg)

	userRepo := repositories.NewUserRepository(db)
	ukmRepo := repositories.NewUkmRepository(db)
	regRepo := repositories.NewRegistrationRepository(db)

	sessionSvc := services.NewSessionService()
	authSvc := services.NewAuthService(cfg, userRepo)
	ukmSvc := services.NewUkmService(ukmRepo, regRepo)

	return &Server{
		Session: handlers.NewSessionHandler(sessionSvc),
		Auth: handlers.NewAuthHandler(authSvc),
		Ukm:  handlers.NewUkmHandler(ukmSvc),
	}
}

