package services

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"

	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
	"openhouse-2025-api/internal/utils"

	"github.com/google/uuid"
)

type UkmService struct {
	ukms *repositories.UkmRepository
	regs *repositories.RegistrationRepository
}

func NewUkmService(ukms *repositories.UkmRepository, regs *repositories.RegistrationRepository) *UkmService {
	return &UkmService{ukms: ukms, regs: regs}
}

func (s *UkmService) List(ctx context.Context) ([]models.Ukm, error) {
	return s.ukms.List(ctx)
}

type CreateUkmRequest struct {
	Name        string                  `form:"name" binding:"required"`
	Slug        string                  `form:"slug" binding:"required"`
	CurrentSlot *int                    `form:"current_slot"`
	MaxSlot     *int                    `form:"max_slot"`
	RegistFee   *int                    `form:"regist_fee"`
	Description string                  `form:"description"`
	Groupchat   string                  `form:"groupchat"`
	VideoURL    string                  `form:"video_url"`
	Logo        *multipart.FileHeader   `form:"logo"`
	Poster      *multipart.FileHeader   `form:"poster"`
	Images      []*multipart.FileHeader `form:"images"`
}

type UpdateUkmRequest struct {
	ID           string                  `form:"id" binding:"required"`
	Name         string                  `form:"name"`
	Slug         string                  `form:"slug"`
	CurrentSlot  *int                    `form:"current_slot"`
	MaxSlot      *int                    `form:"max_slot"`
	RegistFee    *int                    `form:"regist_fee"`
	Description  string                  `form:"description"`
	Groupchat    string                  `form:"groupchat"`
	VideoURL     string                  `form:"video_url"`
	Logo         *multipart.FileHeader   `form:"logo"`
	Poster       *multipart.FileHeader   `form:"poster"`
	Images       []*multipart.FileHeader `form:"images"`
	RemoveImages string                  `form:"remove_images"`
}

func (s *UkmService) Create(ctx context.Context, req *CreateUkmRequest) (*models.Ukm, error) {
	// Create UKM instance
	ukm := &models.Ukm{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Slug:        req.Slug,
		CurrentSlot: req.CurrentSlot,
		MaxSlot:     req.MaxSlot,
		RegistFee:   req.RegistFee,
		Description: req.Description,
		Groupchat:   req.Groupchat,
	}

	if req.VideoURL != "" {
		ukm.VideoURL = &req.VideoURL
	}

	// Handle logo upload
	if req.Logo != nil {
		file, err := req.Logo.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open logo file: %w", err)
		}
		defer file.Close()

		result, err := utils.SaveUkmLogo(file, req.Logo)
		if err != nil {
			return nil, fmt.Errorf("failed to save logo: %w", err)
		}
		ukm.LogoURL = result.RelativeURL
	}

	// Handle poster upload
	if req.Poster != nil {
		file, err := req.Poster.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open poster file: %w", err)
		}
		defer file.Close()

		result, err := utils.SaveUkmPoster(file, req.Poster)
		if err != nil {
			return nil, fmt.Errorf("failed to save poster: %w", err)
		}
		ukm.PosterURL = &result.RelativeURL
	}

	// Handle multiple images upload
	if len(req.Images) > 0 {
		imageURLs := make([]string, 0, len(req.Images))
		for _, imageHeader := range req.Images {
			file, err := imageHeader.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open image file: %w", err)
			}
			defer file.Close()

			result, err := utils.SaveUkmImage(file, imageHeader)
			if err != nil {
				return nil, fmt.Errorf("failed to save image: %w", err)
			}
			imageURLs = append(imageURLs, result.RelativeURL)
		}

		if len(imageURLs) > 0 {
			imageURLsJSON, err := json.Marshal(imageURLs)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal image URLs: %w", err)
			}
			imageURLsStr := string(imageURLsJSON)
			ukm.ImageURLs = &imageURLsStr
		}
	}

	// Save
	if err := s.ukms.Create(ctx, ukm); err != nil {
		return nil, fmt.Errorf("failed to create UKM: %w", err)
	}

	return ukm, nil
}

func (s *UkmService) GetByID(ctx context.Context, id string) (*models.Ukm, error) {
	return s.ukms.FindByID(ctx, id)
}

func (s *UkmService) GetBySlug(ctx context.Context, slug string) (*models.Ukm, error) {
	return s.ukms.FindBySlug(ctx, slug)
}

func (s *UkmService) Update(ctx context.Context, req *UpdateUkmRequest) (*models.Ukm, error) {
	existingUkm, err := s.ukms.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("UKM not found: %w", err)
	}

	if req.Name != "" {
		existingUkm.Name = req.Name
	}
	if req.Slug != "" {
		existingUkm.Slug = req.Slug
	}
	if req.CurrentSlot != nil {
		existingUkm.CurrentSlot = req.CurrentSlot
	}
	if req.MaxSlot != nil {
		existingUkm.MaxSlot = req.MaxSlot
	}
	if req.RegistFee != nil {
		existingUkm.RegistFee = req.RegistFee
	}
	if req.Description != "" {
		existingUkm.Description = req.Description
	}
	if req.Groupchat != "" {
		existingUkm.Groupchat = req.Groupchat
	}
	if req.VideoURL != "" {
		existingUkm.VideoURL = &req.VideoURL
	}

	if req.Logo != nil {
		if existingUkm.LogoURL != "" {
			utils.DeleteFile("uploads" + strings.TrimPrefix(existingUkm.LogoURL, "/uploads"))
		}

		file, err := req.Logo.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open logo file: %w", err)
		}
		defer file.Close()

		result, err := utils.SaveUkmLogo(file, req.Logo)
		if err != nil {
			return nil, fmt.Errorf("failed to save logo: %w", err)
		}
		existingUkm.LogoURL = result.RelativeURL
	}

	if req.Poster != nil {
		if existingUkm.PosterURL != nil && *existingUkm.PosterURL != "" {
			utils.DeleteFile("uploads" + strings.TrimPrefix(*existingUkm.PosterURL, "/uploads"))
		}

		file, err := req.Poster.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open poster file: %w", err)
		}
		defer file.Close()

		result, err := utils.SaveUkmPoster(file, req.Poster)
		if err != nil {
			return nil, fmt.Errorf("failed to save poster: %w", err)
		}
		existingUkm.PosterURL = &result.RelativeURL
	}

	var currentImages []string
	if existingUkm.ImageURLs != nil && *existingUkm.ImageURLs != "" {
		if err := json.Unmarshal([]byte(*existingUkm.ImageURLs), &currentImages); err != nil {
			currentImages = []string{} // Reset if unmarshal fails
		}
	}

	if req.RemoveImages != "" {
		removeList := strings.Split(req.RemoveImages, ",")
		for _, removeURL := range removeList {
			removeURL = strings.TrimSpace(removeURL)
			utils.DeleteFile("uploads" + strings.TrimPrefix(removeURL, "/uploads"))
			for i, img := range currentImages {
				if img == removeURL {
					currentImages = append(currentImages[:i], currentImages[i+1:]...)
					break
				}
			}
		}
	}

	if len(req.Images) > 0 {
		for _, imageHeader := range req.Images {
			file, err := imageHeader.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open image file: %w", err)
			}
			defer file.Close()

			result, err := utils.SaveUkmImage(file, imageHeader)
			if err != nil {
				return nil, fmt.Errorf("failed to save image: %w", err)
			}
			currentImages = append(currentImages, result.RelativeURL)
		}
	}

	if len(currentImages) > 0 {
		imageURLsJSON, err := json.Marshal(currentImages)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal image URLs: %w", err)
		}
		imageURLsStr := string(imageURLsJSON)
		existingUkm.ImageURLs = &imageURLsStr
	} else {
		existingUkm.ImageURLs = nil
	}

	// Save
	if err := s.ukms.Update(ctx, existingUkm); err != nil {
		return nil, fmt.Errorf("failed to update UKM: %w", err)
	}

	return existingUkm, nil
}

func (s *UkmService) Delete(ctx context.Context, id string) error {
	existingUkm, err := s.ukms.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("UKM not found: %w", err)
	}

	if existingUkm.LogoURL != "" {
		utils.DeleteFile("uploads" + strings.TrimPrefix(existingUkm.LogoURL, "/uploads"))
	}
	if existingUkm.PosterURL != nil && *existingUkm.PosterURL != "" {
		utils.DeleteFile("uploads" + strings.TrimPrefix(*existingUkm.PosterURL, "/uploads"))
	}
	if existingUkm.ImageURLs != nil && *existingUkm.ImageURLs != "" {
		var imageURLs []string
		if err := json.Unmarshal([]byte(*existingUkm.ImageURLs), &imageURLs); err == nil {
			for _, imageURL := range imageURLs {
				utils.DeleteFile("uploads" + strings.TrimPrefix(imageURL, "/uploads"))
			}
		}
	}

	return s.ukms.Delete(ctx, id)
}
