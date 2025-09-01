package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileUploadResult struct {
	FileName    string
	FilePath    string
	RelativeURL string
}

var AllowedImageTypes = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

func ValidateImageFile(header *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))

	for _, allowedType := range AllowedImageTypes {
		if ext == allowedType {
			return nil
		}
	}

	return fmt.Errorf("invalid file type %s. Allowed types: %s", ext, strings.Join(AllowedImageTypes, ", "))
}

// SaveUploadedFile saves an uploaded file to the specified directory
func SaveUploadedFile(file multipart.File, header *multipart.FileHeader, uploadDir, prefix string) (*FileUploadResult, error) {
	// Create upload directory
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Filename
	ext := strings.ToLower(filepath.Ext(header.Filename))
	uniqueID := uuid.New().String()[:8]
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s%s", prefix, timestamp, uniqueID, ext)

	filePath := filepath.Join(uploadDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		// Clean up file if copy fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Extract the path relative to the uploads directory
	relativePath := strings.TrimPrefix(uploadDir, "uploads/")
	relativeURL := filepath.Join("/uploads", relativePath, filename)
	relativeURL = strings.ReplaceAll(relativeURL, "\\", "/")

	return &FileUploadResult{
		FileName:    filename,
		FilePath:    filePath,
		RelativeURL: relativeURL,
	}, nil
}

func SaveUkmLogo(file multipart.File, header *multipart.FileHeader) (*FileUploadResult, error) {
	if err := ValidateImageFile(header); err != nil {
		return nil, err
	}

	return SaveUploadedFile(file, header, "uploads/ukm/logos", "logo")
}

func SaveUkmPoster(file multipart.File, header *multipart.FileHeader) (*FileUploadResult, error) {
	if err := ValidateImageFile(header); err != nil {
		return nil, err
	}

	return SaveUploadedFile(file, header, "uploads/ukm/posters", "poster")
}

func SaveUkmImage(file multipart.File, header *multipart.FileHeader) (*FileUploadResult, error) {
	if err := ValidateImageFile(header); err != nil {
		return nil, err
	}

	return SaveUploadedFile(file, header, "uploads/ukm/images", "image")
}

func DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil 
	}

	return os.Remove(filePath)
}

func ValidateYouTubeURL(url string) error {
	if url == "" {
		return nil 
	}

	url = strings.ToLower(url)
	validPrefixes := []string{
		"https://www.youtube.com/",
		"https://youtube.com/",
		"https://youtu.be/",
		"https://m.youtube.com/",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(url, prefix) {
			return nil
		}
	}

	return fmt.Errorf("invalid YouTube URL. Must start with: %s", strings.Join(validPrefixes, ", "))
}
