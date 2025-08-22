package services

import (
	"context"
	"fmt"
	"log"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
)

// Define constants for the possible outcomes
const (
	ValidationSuccess         = "SUCCESS"
	ValidationAlreadyDone     = "ALREADY_VALIDATED"
	ValidationFileRejected    = "FILE_REJECTED"
	ValidationFileNotReviewed = "FILE_NOT_REVIEWED"
)

type PaymentService struct {
	paymentRepo *repositories.PaymentRepository
	userRepo    *repositories.UserRepository // For logging
	ukmRepo     *repositories.UkmRepository  // For logging
}

func NewPaymentService(paymentRepo *repositories.PaymentRepository, userRepo *repositories.UserRepository, ukmRepo *repositories.UkmRepository) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
		ukmRepo:     ukmRepo,
	}
}

type AdminContext struct {
	NRP   string
	Name  string
	UkmID string
}

// ValidatePayment contains the core business logic
func (s *PaymentService) ValidatePayment(ctx context.Context, admin AdminContext, participantNRP, ukmID string) (string, error) {
	// 1. Fetch the registration record using the GORM repository
	detailReg, err := s.paymentRepo.FindRegistration(participantNRP, ukmID)
	if err != nil {
		return "", err // Let the handler deal with not found or other db errors
	}

	// 2. Apply the validation logic
	if detailReg.FileValidated == 1 { // File has been accepted
		if detailReg.PaymentValidated == 0 { // Payment not yet validated, proceed
			// Update the record
			if err := s.paymentRepo.UpdatePaymentStatus(participantNRP, ukmID, 1); err != nil {
				return "", err
			}

			// Log the action (as the service's responsibility)
			go s.logValidationAction(admin, detailReg) // Run in a goroutine so it doesn't block the response

			return ValidationSuccess, nil
		} else {
			// Payment was already validated
			return ValidationAlreadyDone, nil
		}
	} else if detailReg.FileValidated == 2 { // File was rejected
		return ValidationFileRejected, nil
	} else { // File selection has not been validated yet
		return ValidationFileNotReviewed, nil
	}
}

// logValidationAction is a helper for logging
func (s *PaymentService) logValidationAction(admin AdminContext, detailReg *models.DetailRegistration) {
	adminUkm, err := s.ukmRepo.FindByID(context.Background(), admin.UkmID)
	if err != nil {
		log.Printf("Logging Error: Could not fetch admin UKM name: %v", err)
		return
	}

	user, err := s.userRepo.FindByNRP(context.Background(), detailReg.NRP)
	if err != nil {
		log.Printf("Logging Error: Could not fetch user details: %v", err)
		return
	}

	dataUkm, err := s.ukmRepo.FindByID(context.Background(), detailReg.UkmID)
	if err != nil {
		log.Printf("Logging Error: Could not fetch target UKM name: %v", err)
		return
	}

	logMsg := fmt.Sprintf("%s-%s-%s, has accepted payment file of %s-%s-%s",
		admin.NRP, admin.Name, adminUkm.Name,
		user.NRP, user.Name, dataUkm.Name)

	log.Println(logMsg)
}
