package handlers

import (
	"net/http"
	"openhouse-2025-api/internal/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type UpdateBiodataRequest struct {
	LineID string `json:"line_id" binding:"required"`
	Phone  string `json:"phone" binding:"required"`
}

// UpdateBiodata updates user biodata (line_id and phone)
func (h *UserHandler) UpdateBiodata(c *gin.Context) {
	session := sessions.Default(c)

	// Check if user is authenticated
	nrp := session.Get("nrp")
	if nrp == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	name := session.Get("name")
	if name == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Name not found in session"})
		return
	}

	var req UpdateBiodataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate line_id
	if len(req.LineID) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Line ID must be at least 3 characters",
		})
		return
	}

	// Validate phone (basic validation)
	if len(req.Phone) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Phone number must be at least 10 characters",
		})
		return
	}

	err := h.userService.UpdateBiodata(c.Request.Context(), nrp.(string), name.(string), req.LineID, req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update biodata",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Biodata updated successfully",
		"data": gin.H{
			"nrp":     nrp,
			"name":    name,
			"line_id": req.LineID,
			"phone":   req.Phone,
		},
	})
}

// GetBiodata retrieves user biodata
func (h *UserHandler) GetBiodata(c *gin.Context) {
	session := sessions.Default(c)

	// Check if user is authenticated
	nrp := session.Get("nrp")
	if nrp == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	name := session.Get("name")
	if name == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Name not found in session"})
		return
	}

	user, err := h.userService.GetUserByNRP(c.Request.Context(), nrp.(string))
	if err != nil {
		// If user doesn't exist in database yet, return default data with session info
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusOK, gin.H{
				"message": "User data retrieved successfully",
				"data": gin.H{
					"nrp":     nrp,
					"name":    name,
					"line_id": "",
					"phone":   "",
				},
			})
			return
		}

		// For other database errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user data",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User data retrieved successfully",
		"data":    user,
	})
}
