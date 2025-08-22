package repositories

import (
	"context"
	"openhouse-2025-api/internal/models"

	"gorm.io/gorm"
)

// UkmRepository now holds a GORM database handle.
type UkmRepository struct {
	db *gorm.DB
}

// NewUkmRepository's constructor now accepts a *gorm.DB instance.
func NewUkmRepository(db *gorm.DB) *UkmRepository {
	return &UkmRepository{db: db}
}

// List retrieves all Ukm records from the database, ordered by name.
func (r *UkmRepository) List(ctx context.Context) ([]models.Ukm, error) {
	var ukms []models.Ukm

	// GORM's chainable methods replace the raw SQL query and the manual scanning loop.
	// .WithContext(ctx) passes the request context down to the database driver.
	// .Order("name") applies the sorting.
	// .Find(&ukms) executes the query and scans the results directly into the slice.
	result := r.db.WithContext(ctx).Order("name").Find(&ukms)

	if result.Error != nil {
		return nil, result.Error
	}

	return ukms, nil
}

// FindByID retrieves a single Ukm by its primary key.
func (r *UkmRepository) FindByID(ctx context.Context, ukmID string) (*models.Ukm, error) {
	var ukm models.Ukm

	// For primary key lookups, GORM's .First() method is the most idiomatic and efficient.
	// It's equivalent to `WHERE id = ? LIMIT 1`.
	// It will automatically return `gorm.ErrRecordNotFound` if no record matches the ID.
	result := r.db.WithContext(ctx).First(&ukm, ukmID)

	if result.Error != nil {
		return nil, result.Error
	}

	return &ukm, nil
}
