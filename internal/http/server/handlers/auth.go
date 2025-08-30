package handlers

import (
	"openhouse-2025-api/internal/services"

	"github.com/gin-gonic/gin"
	// "net/http"
	// "github.com/gin-contrib/sessions"
)

// var r = gin.Default()
type AuthHandler struct{ service *services.AuthService }

func NewAuthHandler(s *services.AuthService) *AuthHandler { return &AuthHandler{service: s} }

func (h *AuthHandler) BeginGoogleAuth(c *gin.Context) {

	h.service.BeginGoogleAuth(c)

}

func (h *AuthHandler) OAuthCallback(c *gin.Context) {

	h.service.OAuthCallback(c)

}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.service.Logout(c)
}
