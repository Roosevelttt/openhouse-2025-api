package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"openhouse-2025-api/internal/services"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(s *services.AuthService) *AuthHandler { return &AuthHandler{service: s} }

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := h.service.GetGoogleAuthURL()
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	token, user, err := h.service.HandleGoogleCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

