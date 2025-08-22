package router

import (
	"net/http"

	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/server"
	"openhouse-2025-api/internal/middleware"

	"github.com/gin-gonic/gin"
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

		api.GET("/ukms", s.Ukm.List)
		api.GET("/participants", s.Participants.List)
		api.POST("/payment/validate", s.Payment.Validate)
	}

	return r
}
