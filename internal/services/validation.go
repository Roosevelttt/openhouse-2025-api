package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"

	"gorm.io/gorm"
)

// Define constants for the possible outcomes
const (
	ValidationSuccess         = "SUCCESS"
	ValidationAlreadyDone     = "ALREADY_VALIDATED"
	ValidationFileRejected    = "FILE_REJECTED"
	ValidationFileNotReviewed = "FILE_NOT_REVIEWED"
	RejectionSuccess          = "SUCCESS"
	RejectionAlreadyDone      = "ALREADY_REJECTED"
)

// The service has been renamed to ValidationService to reflect its broader responsibilities.
// It now requires the main *gorm.DB for handling transactions.
type ValidationService struct {
	db         *gorm.DB
	statusRepo *repositories.ValidationRepository // Assuming a unified repository
	userRepo   *repositories.UserRepository       // For logging
	ukmRepo    *repositories.UkmRepository        // For logging
}

// The constructor now accepts the DB handle and the unified StatusRepository.
func NewValidationService(db *gorm.DB, statusRepo *repositories.ValidationRepository, userRepo *repositories.UserRepository, ukmRepo *repositories.UkmRepository) *ValidationService {
	return &ValidationService{
		db:         db,
		statusRepo: statusRepo,
		userRepo:   userRepo,
		ukmRepo:    ukmRepo,
	}
}

type AdminContext struct {
	NRP   string
	Name  string
	UkmID string
}

// ProcessValidation is the new, flexible method for all acceptance logic.
func (s *ValidationService) ProcessValidation(ctx context.Context, admin AdminContext, validationType, nrp, ukmID string) (string, error) {
	detailReg, err := s.statusRepo.FindRegistration(nrp, ukmID)
	if err != nil {
		return "", err // Handle not found or other db errors
	}

	switch validationType {
	//case "file":
	//	// sudah di reject
	//	if detailReg.FileValidated == 2 || detailReg.PaymentValidated == 2 {
	//		return ValidationFileRejected, nil // Already rejected
	//	}
	//	// Logic for validating a selection file
	//	if detailReg.FileValidated == 1 {
	//		return ValidationAlreadyDone, nil // Already accepted
	//	}
	//	// Update the record to '1' (Accepted)
	//	if err := s.statusRepo.UpdateStatus(s.db, nrp, ukmID, "file_validated", 1); err != nil {
	//		return "", err
	//	}
	//	go s.logValidationAction(admin, detailReg, "selection file")
	//	return ValidationSuccess, nil

	case "payment":
		// sudah di accept
		if detailReg.PaymentValidated == 1 {
			return ValidationAlreadyDone, nil
		}
		// sudah di reject
		if detailReg.PaymentValidated == 2 {
			return RejectionAlreadyDone, nil
		}
		// This is the original logic from ValidatePayment
		//if detailReg.FileValidated == 0 {
		//	return ValidationFileNotReviewed, nil
		//}

		// Update the record to '1' (Accepted)
		if err := s.statusRepo.UpdateStatus(s.db, nrp, ukmID, "payment_validated", 1); err != nil {
			return "", err
		}
		go s.logValidationAction(admin, detailReg, "payment file")
		return ValidationSuccess, nil

	default:
		return "", errors.New("invalid validation type")
	}
}

// ProcessRejection is the method from the old RejectionService, now part of ValidationService.
func (s *ValidationService) ProcessRejection(ctx context.Context, admin AdminContext, reqType, nrp, ukmID string) (string, error) {
	var fieldToUpdate string
	var logFileType string

	detailReg, err := s.statusRepo.FindRegistration(nrp, ukmID)
	if err != nil {
		return "", err // Handle not found or other db errors
	}

	switch reqType {
	case "payment":
		if detailReg.PaymentValidated == 1 {
			return ValidationAlreadyDone, nil
		}
		if detailReg.PaymentValidated == 2 {
			return RejectionAlreadyDone, nil
		}
		fieldToUpdate = "payment_validated"
		logFileType = "payment file"
	//case "file":
	//	if detailReg.FileValidated == 1 || detailReg.PaymentValidated == 1 {
	//		return ValidationAlreadyDone, nil
	//	}
	//	if detailReg.FileValidated == 2 {
	//		return RejectionAlreadyDone, nil
	//	}
	//	fieldToUpdate = "file_validated"
	//	logFileType = "selection file"
	default:
		return "", errors.New("invalid rejection type")
	}

	// Use a transaction to ensure both database updates succeed or fail together.
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Increment the UKM slot
		if err := s.statusRepo.IncrementUkmSlot(tx, ukmID); err != nil {
			return fmt.Errorf("failed to increment slot: %w", err)
		}

		// 2. Update the registration status to '2' (Rejected)
		if err := s.statusRepo.UpdateStatus(tx, nrp, ukmID, fieldToUpdate, 2); err != nil {
			return fmt.Errorf("failed to update registration status: %w", err)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	// Side effects like logging and sending emails happen after the database transaction is successful.
	go s.logAndNotifyRejection(admin, nrp, ukmID, logFileType, reqType)

	return RejectionSuccess, nil
}

// logValidationAction is now updated to accept the type of file being validated
func (s *ValidationService) logValidationAction(admin AdminContext, detailReg *models.DetailRegistration, fileType string) {
	adminUkm, _ := s.ukmRepo.FindByID(context.Background(), admin.UkmID)
	user, _ := s.userRepo.FindByNRP(context.Background(), detailReg.NRP)
	dataUkm, _ := s.ukmRepo.FindByID(context.Background(), detailReg.UkmID)

	logMsg := fmt.Sprintf("%s-%s-%s, has accepted %s of %s-%s-%s",
		admin.NRP, admin.Name, adminUkm.Name,
		fileType, // Use the dynamic file type
		user.NRP, user.Name, dataUkm.Name)
	log.Println(logMsg)
}

// logAndNotifyRejection is the helper for rejections.
func (s *ValidationService) logAndNotifyRejection(admin AdminContext, nrp, ukmID, logFileType, mailType string) {
	adminUkm, _ := s.ukmRepo.FindByID(context.Background(), admin.UkmID)
	user, _ := s.userRepo.FindByNRP(context.Background(), nrp)
	dataUkm, _ := s.ukmRepo.FindByID(context.Background(), ukmID)

	logMsg := fmt.Sprintf("%s-%s-%s, has rejected %s of %s-%s-%s",
		admin.NRP, admin.Name, adminUkm.Name,
		logFileType,
		user.NRP, user.Name, dataUkm.Name)
	log.Println(logMsg)

	// Email sending logic would be called here
	log.Printf("INFO: Sending rejection email for %s to %s for UKM %s", mailType, user.NRP, dataUkm.Name)
}
