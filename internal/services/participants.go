package services

import (
	"context"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
	"time"
)

type ParticipantsService struct {
	participants *repositories.ParticipantsRepository
	registration *repositories.RegistrationRepository
}

func NewParticipantsService(repo *repositories.ParticipantsRepository, regRepo *repositories.RegistrationRepository) *ParticipantsService {
	return &ParticipantsService{
		participants: repo,
		registration: regRepo,
	}
}

func (s *ParticipantsService) ListParticipants(ctx context.Context, adminDivisionSlug string, adminUkmID string) ([]models.Participant, error) {
	return s.participants.List(ctx, adminDivisionSlug, adminUkmID)
}

// ReserveSlot reserves a slot for the user
func (s *ParticipantsService) ReserveSlot(ctx context.Context, nrp, ukmID string) (*models.ReserveSlotResponse, error) {
	reservationID, err := s.registration.ReserveSlot(ctx, nrp, ukmID)
	if err != nil {
		return nil, err
	}

	// Calculate expiry time (10 minutes from now)
	expiresAt := time.Now().Add(10 * time.Minute)

	return &models.ReserveSlotResponse{
		ReservationID: reservationID,
		ExpiresAt:     expiresAt,
	}, nil
}

// RegisterWithReservation creates a registration using an existing reservation
func (s *ParticipantsService) RegisterWithReservation(ctx context.Context, reservationID string, reg *models.DetailRegistration) error {
	return s.registration.ConsumeReservation(ctx, reservationID, reg)
}
