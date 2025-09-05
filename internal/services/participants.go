package services

import (
	"context"
	"fmt"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
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
	result, err := s.registration.ReserveSlotForPayment(ctx, nrp, ukmID)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("no slots available")
	}

	return &models.ReserveSlotResponse{
		ReservationID: result.ReservationID,
		ExpiresAt:     result.ExpiresAt, // Use actual database time
	}, nil
}

// RegisterWithReservation creates a registration using an existing reservation
func (s *ParticipantsService) RegisterWithReservation(ctx context.Context, reservationID string, reg *models.DetailRegistration) error {
	return s.registration.ConsumeReservation(ctx, reservationID, reg)
}

// CheckUserReservation checks if user has a valid reservation for a UKM
func (s *ParticipantsService) CheckUserReservation(ctx context.Context, nrp, ukmID string) (*models.SlotReservation, error) {
	return s.registration.GetUserReservation(ctx, nrp, ukmID)
}
