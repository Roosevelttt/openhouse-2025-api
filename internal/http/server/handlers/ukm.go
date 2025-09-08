package handlers

import (
	"net/http"
	"strconv"


	"github.com/gin-gonic/gin"
	"openhouse-2025-api/internal/services"
)

type UkmHandler struct{ service *services.UkmService }

func NewUkmHandler(s *services.UkmService) *UkmHandler { return &UkmHandler{service: s} }

func (h *UkmHandler) List(c *gin.Context) {
	ukms, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ukms)
}

// Create creates a new UKM with file uploads
func (h *UkmHandler) Create(c *gin.Context) {

	// Parse basic fields
	req := &services.CreateUkmRequest{
		Name:        c.PostForm("name"),
		Slug:        c.PostForm("slug"),
		Description: c.PostForm("description"),
		Groupchat:   c.PostForm("groupchat"),
		VideoURL:    c.PostForm("video_url"),
	}

	// Parse optional integer fields
	if currentSlotStr := c.PostForm("current_slot"); currentSlotStr != "" {
		if val, err := strconv.Atoi(currentSlotStr); err == nil {
			req.CurrentSlot = &val
		}
	}
	if maxSlotStr := c.PostForm("max_slot"); maxSlotStr != "" {
		if val, err := strconv.Atoi(maxSlotStr); err == nil {
			req.MaxSlot = &val
		}
	}
	if registFeeStr := c.PostForm("regist_fee"); registFeeStr != "" {
		if val, err := strconv.Atoi(registFeeStr); err == nil {
			req.RegistFee = &val
		}
	}

	// Validate required fields
	if req.Name == "" || req.Slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and slug are required"})
		return
	}

	// Handle file uploads
	if logo, err := c.FormFile("logo"); err == nil {
		req.Logo = logo
	}
	if poster, err := c.FormFile("poster"); err == nil {
		req.Poster = poster
	}
	if form := c.Request.MultipartForm; form != nil && form.File["images"] != nil {
		req.Images = form.File["images"]
	}

	// Create UKM
	ukm, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ukm)
}

func (h *UkmHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	ukm, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "UKM not found"})
		return
	}

	c.JSON(http.StatusOK, ukm)
}

func (h *UkmHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
		return
	}

	ukm, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "UKM not found"})
		return
	}

	c.JSON(http.StatusOK, ukm)
}

func (h *UkmHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	// Parse basic fields
	req := &services.UpdateUkmRequest{
		ID:           id,
		Name:         c.PostForm("name"),
		Slug:         c.PostForm("slug"),
		Description:  c.PostForm("description"),
		Groupchat:    c.PostForm("groupchat"),
		VideoURL:     c.PostForm("video_url"),
		RemoveImages: c.PostForm("remove_images"),
	}

	// Parse optional integer fields
	if currentSlotStr := c.PostForm("current_slot"); currentSlotStr != "" {
		if val, err := strconv.Atoi(currentSlotStr); err == nil {
			req.CurrentSlot = &val
		}
	}
	if maxSlotStr := c.PostForm("max_slot"); maxSlotStr != "" {
		if val, err := strconv.Atoi(maxSlotStr); err == nil {
			req.MaxSlot = &val
		}
	}
	if registFeeStr := c.PostForm("regist_fee"); registFeeStr != "" {
		if val, err := strconv.Atoi(registFeeStr); err == nil {
			req.RegistFee = &val
		}
	}

	// Handle file uploads
	if logo, err := c.FormFile("logo"); err == nil {
		req.Logo = logo
	}
	if poster, err := c.FormFile("poster"); err == nil {
		req.Poster = poster
	}
	if form := c.Request.MultipartForm; form != nil && form.File["images"] != nil {
		req.Images = form.File["images"]
	}

	// Update UKM
	ukm, err := h.service.Update(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ukm)
}

func (h *UkmHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "UKM deleted successfully"})
}
