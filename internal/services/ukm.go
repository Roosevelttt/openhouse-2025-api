package services

import (
	"context"

	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
)

type UkmService struct {
	ukms *repositories.UkmRepository
	regs *repositories.RegistrationRepository
}

func NewUkmService(ukms *repositories.UkmRepository, regs *repositories.RegistrationRepository) *UkmService {
	return &UkmService{ukms: ukms, regs: regs}
}

func (s *UkmService) List(ctx context.Context) ([]models.Ukm, error) {
	return s.ukms.List(ctx)
}

