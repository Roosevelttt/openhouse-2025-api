package services

import (
	"context"
	"errors"
	"fmt"

	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
)

type AuthService struct {
	cfg *config.Config
	users *repositories.UserRepository
}

func NewAuthService(cfg *config.Config, users *repositories.UserRepository) *AuthService {
	return &AuthService{cfg: cfg, users: users}
}

func (s *AuthService) GetGoogleAuthURL() string {
	// Placeholder for Google OAuth URL construction; swap with oauth2.Config.
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=profile email", s.cfg.GoogleClientID, s.cfg.GoogleRedirectURL)
}

func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (string, *models.User, error) {
	if code == "" { return "", nil, errors.New("missing code") }
	// TODO: Exchange code with Google, fetch profile; for now return mock user.
	u := &models.User{NRP: "sample", Name: "Sample User"}
	if err := s.users.UpsertByNRP(ctx, u); err != nil { return "", nil, err }
	// TODO: issue JWT signed with s.cfg.JWTSecret
	return "mock-jwt", u, nil
}

