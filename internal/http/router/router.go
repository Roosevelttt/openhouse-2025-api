package router

import (
	"net/http"
	// "log"
	// "os"

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
	r.Use(middleware.SessionManager())
	

	s := server.NewServer(cfg)

	api := r.Group("/api")
	{

		api.GET("/auth/google/start", s.Auth.BeginGoogleAuth)
		api.GET("auth/google/callback", s.Auth.OAuthCallback)
		

		api.GET("/ukms", s.Ukm.List)
		api.GET("/participants", s.Participants.List)
		api.POST("/payment/validate", s.Payment.Validate)

		
		api.Use(middleware.Authentication("user")) 
        {
            api.POST("/session/values", s.Session.GetValues)
        }
	}

	return r
}
