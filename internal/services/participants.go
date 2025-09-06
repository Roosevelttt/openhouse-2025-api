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
	mailService  *MailService
	userRepo     *repositories.UserRepository
	ukmRepo      *repositories.UkmRepository
}

func NewParticipantsService(
	repo *repositories.ParticipantsRepository,
	regRepo *repositories.RegistrationRepository,
	mailService *MailService,
	userRepo *repositories.UserRepository,
	ukmRepo *repositories.UkmRepository,
) *ParticipantsService {
	return &ParticipantsService{
		participants: repo,
		registration: regRepo,
		mailService:  mailService,
		userRepo:     userRepo,
		ukmRepo:      ukmRepo,
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
	// First, complete the registration
	err := s.registration.ConsumeReservation(ctx, reservationID, reg)
	if err != nil {
		return err
	}

	// Send email notifications if mail service is configured
	if s.mailService != nil {
		// Get user details
		user, err := s.userRepo.FindByNRP(ctx, reg.NRP)
		if err != nil {
			// Log error but don't fail the registration
			fmt.Printf("Warning: Could not get user details for email: %v\n", err)
			return nil
		}

		// Get UKM details
		ukm, err := s.ukmRepo.FindByID(ctx, reg.UkmID)
		if err != nil {
			// Log error but don't fail the registration
			fmt.Printf("Warning: Could not get UKM details for email: %v\n", err)
			return nil
		}

		// --- Email Sending Logic (using payment confirmation template) ---
		fmt.Printf("INFO: Sending payment confirmation email to %s for UKM %s\n", user.NRP, ukm.Name)

		// Prepare payment confirmation data
		emailData := PaymentConfirmationData{
			UserName: user.Name,
			UserNRP:  user.NRP,
			UkmName:  ukm.Name,
		}

		// Use SendPaymentConfirmationEmailTemplate for payment submissions
		go func() {
			err := s.mailService.SendPaymentConfirmationEmailTemplate(user.NRP, emailData)
			if err != nil {
				fmt.Printf("ERROR: Failed to send payment confirmation email to %s: %v\n", user.NRP, err)
			} else {
				fmt.Printf("INFO: Payment confirmation email sent successfully to %s\n", user.NRP)
			}
		}()
	}

	return nil
}

// CheckUserReservation checks if user has a valid reservation for a UKM
func (s *ParticipantsService) CheckUserReservation(ctx context.Context, nrp, ukmID string) (*models.SlotReservation, error) {
	return s.registration.GetUserReservation(ctx, nrp, ukmID)
}
