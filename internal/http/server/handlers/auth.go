package handlers

import (
    "github.com/gin-gonic/gin"
    "openhouse-2025-api/internal/services"

    // "net/http"
    // "github.com/gin-contrib/sessions"
)

// var r = gin.Default()
type AuthHandler struct { service *services.AuthService }
func NewAuthHandler(s *services.AuthService) *AuthHandler { return &AuthHandler{service: s} }


func (h *AuthHandler) BeginGoogleAuth(c *gin.Context) {

    h.service.BeginGoogleAuth(c) 
    
}

func (h *AuthHandler) OAuthCallback(c *gin.Context) {

    h.service.OAuthCallback(c) 
    
}

