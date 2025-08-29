package handlers

import (
	"net/http"
	"openhouse-2025-api/internal/services"

	"github.com/gin-gonic/gin"
)

type GroupchatHandler struct {
	service *services.GroupchatService
}

func NewGroupchatHandler(service *services.GroupchatService) *GroupchatHandler {
	return &GroupchatHandler{service: service}
}

// GET /api/admin/ukm/groupchat
func (h *GroupchatHandler) Get(c *gin.Context) {
	adminUkmID := c.GetString("admin_ukm_id")
	if adminUkmID == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin is not associated with any UKM"})
		return
	}

	link, err := h.service.GetLink(c.Request.Context(), adminUkmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve groupchat link"})
		return
	}

	// If link is nil, return an empty string for consistency on the frontend
	if link == nil {
		c.JSON(http.StatusOK, gin.H{"groupchat_link": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groupchat_link": *link})
}

type updateLinkRequest struct {
	Link string `json:"link" binding:"required,url"`
}

// PUT /api/admin/ukm/groupchat
func (h *GroupchatHandler) Update(c *gin.Context) {
	adminUkmID := c.GetString("admin_ukm_id")
	if adminUkmID == "" {
		c.JSON(http.StatusForbidden, gin.H{"message": "Admin is not associated with any UKM"})
		return
	}

	var req updateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: " + err.Error()})
		return
	}

	err := h.service.UpdateLink(c.Request.Context(), adminUkmID, req.Link)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update group chat link! UKM Not found or link has not changed."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Groupchat link updated successfully"})
}
