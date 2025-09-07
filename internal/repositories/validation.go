package repositories

import (
	"openhouse-2025-api/internal/models" // Assuming your GORM models are here

	"gorm.io/gorm"
)

// ValidationRepository uses GORM for its operations
type ValidationRepository struct {
	db *gorm.DB
}

// NewValidationRepository is the constructor
func NewValidationRepository(db *gorm.DB) *ValidationRepository {
	return &ValidationRepository{db: db}
}

// An example method using GORM
func (r *ValidationRepository) FindRegistration(nrp string, ukmID string) (*models.DetailRegistration, error) {
	var detailReg models.DetailRegistration
	err := r.db.Where("nrp = ? AND ukm_id = ?", nrp, ukmID).First(&detailReg).Error
	if err != nil {
		return nil, err
	}
	return &detailReg, nil
}

//// UpdatePaymentValidatedStatus updates the 'payment_validated' field for a specific registration record.
//func (r *ValidationRepository) UpdatePaymentValidatedStatus(nrp string, ukmID string, status int) error {
//	// We use .Model() to specify which table we are updating without needing to fetch the object first.
//	// This is more efficient for simple updates.
//	result := r.db.Model(&models.DetailRegistration{}).
//		Where("nrp = ? AND ukm_id = ?", nrp, ukmID).
//		Update("payment_validated", status)
//
//	if result.Error != nil {
//		return result.Error
//	}
//
//	// Good practice: Check if any row was actually affected.
//	// If the WHERE clause didn't match any record, GORM won't return an error,
//	// but no rows will be updated. We can treat this as a "not found" case.
//	if result.RowsAffected == 0 {
//		return gorm.ErrRecordNotFound
//	}
//
//	return nil
//}
//
//// UpdateIsSelectedStatus updates the 'isSelected' field for a specific registration record.
//func (r *ValidationRepository) UpdateIsSelectedStatus(nrp string, ukmID string, status int) error {
//	// We use .Model() to specify which table we are updating without needing to fetch the object first.
//	// This is more efficient for simple updates.
//	result := r.db.Model(&models.DetailRegistration{}).
//		Where("nrp = ? AND ukm_id = ?", nrp, ukmID).
//		Update("isSelected", status)
//
//	if result.Error != nil {
//		return result.Error
//	}
//
//	// Good practice: Check if any row was actually affected.
//	// If the WHERE clause didn't match any record, GORM won't return an error,
//	// but no rows will be updated. We can treat this as a "not found" case.
//	if result.RowsAffected == 0 {
//		return gorm.ErrRecordNotFound
//	}
//
//	return nil
//}

// UpdateStatus is the new, generic method to update any status field for a registration.
// It accepts a database handle, which can be the main DB instance or a transaction (tx).
func (r *ValidationRepository) UpdateStatus(db *gorm.DB, nrp, ukmID string, fieldName string, status int) error {
	// .Model() specifies the table.
	// .Where() finds the specific record.
	// .Update() takes the column name as a string and the new value.
	result := db.Model(&models.DetailRegistration{}).
		Where("nrp = ? AND ukm_id = ?", nrp, ukmID).
		Update(fieldName, status)

	if result.Error != nil {
		return result.Error
	}

	// This is a crucial check to ensure a record was actually updated.
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DecrementUkmSlot atomically increments the current_slot for a UKM.
func (r *ValidationRepository) DecrementUkmSlot(tx *gorm.DB, ukmID string) error {
	// gorm.Expr allows us to write raw SQL expressions for updates.
	// This is the safe way to do an increment.
	result := tx.Model(&models.Ukm{}).Where("id = ?", ukmID).Update("current_slot", gorm.Expr("COALESCE(current_slot, 0) - 1"))
	return result.Error
}
