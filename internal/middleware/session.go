package middleware

import (
	"log"
	"os"

	"openhouse-2025-api/internal/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func SessionManager(cfg *config.Config) gin.HandlerFunc {
	// hashkey hrs ada
	hashKey := os.Getenv("SESSION_HASH_KEY")
	if hashKey == "" {
		log.Fatal("SESSION_HASH_KEY environment variable is not set.")
	}

	// Check if we're in development mode
	isDevelopment := cfg.IsDevelopment()

	store := cookie.NewStore([]byte(hashKey))
	store.Options(sessions.Options{
		MaxAge:   60 * 60 * 24 * 7, // 7 days in seconds
		Path:     "/",
		HttpOnly: true,
		SameSite: 0,
		// Secure: false for development to allow HTTP connections
		Secure: !isDevelopment,
	})
	gothic.Store = store

	return sessions.Sessions("mysession", store)
}