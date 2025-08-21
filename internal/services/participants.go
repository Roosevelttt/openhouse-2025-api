package services

import (
	"context"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
)

type ParticipantsService struct {
	participants *repositories.ParticipantsRepository
}

func NewParticipantsService(repo *repositories.ParticipantsRepository) *ParticipantsService {
	return &ParticipantsService{participants: repo}
}

func (s *ParticipantsService) ListParticipants(ctx context.Context) ([]models.Participant, error) {
	return s.participants.List(ctx)
}
