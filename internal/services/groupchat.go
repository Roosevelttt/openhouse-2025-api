package services

import (
	"context"
	"openhouse-2025-api/internal/repositories"
)

type GroupchatService struct {
	ukmRepo *repositories.UkmRepository
}

func NewGroupchatService(ukmRepo *repositories.UkmRepository) *GroupchatService {
	return &GroupchatService{ukmRepo: ukmRepo}
}

func (s *GroupchatService) GetLink(ctx context.Context, ukmID string) (*string, error) {
	return s.ukmRepo.GetGroupchatLink(ctx, ukmID)
}

func (s *GroupchatService) UpdateLink(ctx context.Context, ukmID string, newLink string) error {
	return s.ukmRepo.UpdateGroupchatLink(ctx, ukmID, newLink)
}
