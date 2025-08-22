package repositories

import (
	"openhouse-2025-api/internal/models" // Assuming your GORM models are here

	"gorm.io/gorm"
)

// PaymentRepository uses GORM for its operations
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository is the constructor
func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// An example method using GORM
func (r *PaymentRepository) FindRegistration(nrp string, ukmID string) (*models.DetailRegistration, error) {
	var detailReg models.DetailRegistration
	err := r.db.Where("nrp = ? AND ukm_id = ?", nrp, ukmID).First(&detailReg).Error
	if err != nil {
		return nil, err
	}
	return &detailReg, nil
}

func (r *PaymentRepository) UpdateRegistration(detailReg *models.DetailRegistration) error {
	return nil
}

// UpdatePaymentStatus updates the 'payment_validated' field for a specific registration record.
func (r *PaymentRepository) UpdatePaymentStatus(nrp string, ukmID string, status int) error {
	// We use .Model() to specify which table we are updating without needing to fetch the object first.
	// This is more efficient for simple updates.
	result := r.db.Model(&models.DetailRegistration{}).
		Where("nrp = ? AND ukm_id = ?", nrp, ukmID).
		Update("payment_validated", status)

	if result.Error != nil {
		return result.Error
	}

	// Good practice: Check if any row was actually affected.
	// If the WHERE clause didn't match any record, GORM won't return an error,
	// but no rows will be updated. We can treat this as a "not found" case.
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// ... other GORM-based methods for validation update, etc.
