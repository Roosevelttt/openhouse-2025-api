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
	r.Use(middleware.SessionManager(cfg))

	s := server.NewServer(cfg)

	api := r.Group("/api")
	{
		api.GET("/debug/session", s.Session.DebugSession)
		api.GET("/auth/google/start", s.Auth.BeginGoogleAuth)
		api.GET("/auth/google/callback", s.Auth.OAuthCallback)
		api.POST("/auth/logout", s.Auth.Logout)

		// User routes
		userRoutes := api.Group("/user")
		userRoutes.Use(middleware.AuthMiddleware("user", "admin"))
		{
			userRoutes.POST("/session/values", s.Session.GetValues)
			userRoutes.GET("/ukms", s.Ukm.List)
		}

		// Admin routes
		adminRoutes := api.Group("/admin")
		adminRoutes.Use(middleware.AuthMiddleware("admin"))
		{
			adminRoutes.GET("/participants", s.Participants.List)

			payment := adminRoutes.Group("/payment")
			{
				payment.POST("/validate", s.Validation.Validate)
				payment.POST("/reject", s.Validation.Reject)
			}

			export := adminRoutes.Group("/export")
			{
				export.GET("/participants", s.Export.ExportParticipants)
			}
		}
	}

	return r
}
