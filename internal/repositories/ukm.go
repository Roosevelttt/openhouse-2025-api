package repositories

import (
	"context"
	"database/sql"
	"errors"
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
	result := r.db.WithContext(ctx).Where("id = ?", ukmID).First(&ukm)

	if result.Error != nil {
		return nil, result.Error
	}

	return &ukm, nil
}

// GetGroupchatLink retrieves only the groupchat_link for a specific UKM ID.
func (r *UkmRepository) GetGroupchatLink(ctx context.Context, ukmID string) (*string, error) {
	var groupchatLink sql.NullString // Use sql.NullString to handle NULL values

	result := r.db.Model(&models.Ukm{}).WithContext(ctx).Select("groupchat").Where("id = ?", ukmID).Scan(&groupchatLink)
	if result.Error != nil {
		// If no record is found, it's not a fatal error for this function. Return nil.
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	if groupchatLink.Valid {
		return &groupchatLink.String, nil
	}

	return nil, nil // Return nil if the link is NULL in the database
}

// UpdateGroupchatLink updates the groupchat_link for a specific UKM ID.
func (r *UkmRepository) UpdateGroupchatLink(ctx context.Context, ukmID string, newLink string) error {
	result := r.db.Model(&models.Ukm{}).WithContext(ctx).Where("id = ?", ukmID).Update("groupchat", newLink)

	if result.Error != nil {
		return result.Error
	}

	// It's good practice to check if a row was actually affected.
	// If not, it means the ukmID did not exist.
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *UkmRepository) Create(ctx context.Context, ukm *models.Ukm) error {
	result := r.db.WithContext(ctx).Create(ukm)
	return result.Error
}

func (r *UkmRepository) Update(ctx context.Context, ukm *models.Ukm) error {
	result := r.db.WithContext(ctx).Save(ukm)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UkmRepository) Delete(ctx context.Context, ukmID string) error {
	result := r.db.WithContext(ctx).Delete(&models.Ukm{}, "id = ?", ukmID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UkmRepository) FindBySlug(ctx context.Context, slug string) (*models.Ukm, error) {
	var ukm models.Ukm
	result := r.db.WithContext(ctx).Where("slug = ?", slug).First(&ukm)
	if result.Error != nil {
		return nil, result.Error
	}
	return &ukm, nil
}
