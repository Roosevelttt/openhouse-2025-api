package services

import (
	"context"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// UpdateBiodata updates user biodata information
func (s *UserService) UpdateBiodata(ctx context.Context, nrp, name, lineID, phone string) error {
	user := &models.User{
		NRP:    nrp,
		Name:   name,
		LineID: lineID,
		Phone:  phone,
	}

	return s.userRepo.UpsertByNRP(ctx, user)
}

// GetUserByNRP retrieves user information by NRP
func (s *UserService) GetUserByNRP(ctx context.Context, nrp string) (*models.User, error) {
	return s.userRepo.FindByNRP(ctx, nrp)
}
