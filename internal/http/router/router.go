package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/server"
	"openhouse-2025-api/internal/middleware"
	// "openhouse-2025-api/internal/http/authentication"
)

func New(cfg *config.Config) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg))

	s := server.NewServer(cfg)

	api := r.Group("/api")
	{
		// api.GET("/auth/google", s.Auth.GoogleLogin)
		// api.GET("/auth/google/callback", s.Auth.GoogleCallback)

		api.GET("/auth/google/start", s.Auth.BeginGoogleAuth)
		api.GET("auth/google/callback", s.Auth.OAuthCallback)

		api.GET("/ukms", s.Ukm.List)
	}

	return r
}

