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
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.SessionManager(cfg))

	s := server.NewServer(cfg)

	// Serve static files for uploads
	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		api.GET("/debug/session", s.Session.DebugSession)
		api.GET("/auth/google/start", s.Auth.BeginGoogleAuth)
		api.GET("/auth/google/callback", s.Auth.OAuthCallback)
		api.POST("/auth/logout", s.Auth.Logout)

		api.GET("/ukms", s.Ukm.List)
		api.GET("/ukms/:id", s.Ukm.GetByID)
		api.GET("/ukms/slug/:slug", s.Ukm.GetBySlug)

		// User routes
		userRoutes := api.Group("/user")
		userRoutes.Use(middleware.AuthMiddleware("user", "admin"))
		{
			userRoutes.POST("/session/values", s.Session.GetValues)
			userRoutes.GET("/biodata", s.User.GetBiodata)
			userRoutes.POST("/biodata", s.User.UpdateBiodata)
		}

		// Registration routes
		registrationRoutes := api.Group("/registrations")
		registrationRoutes.Use(middleware.AuthMiddleware("user", "admin"))
		{
			registrationRoutes.POST("/reserve", s.Participants.ReserveSlot)
			registrationRoutes.POST("/access-payment/:ukm_id", s.Participants.AccessPaymentPage)
			registrationRoutes.GET("/check-reservation/:ukm_id", s.Participants.CheckUserReservation)
			registrationRoutes.POST("/with-reservation/:reservationId", s.Participants.RegisterWithReservation)
		}

		// Public registration routes (no auth needed)
		{
			api.GET("/registrations/test", s.Registration.Test)
			api.POST("/registrations", s.Registration.Create)
		}

		// Admin routes
		adminRoutes := api.Group("/admin")
		adminRoutes.Use(middleware.AuthMiddleware("admin"))
		{
			adminRoutes.GET("/participants", s.Participants.List)

			ukm := adminRoutes.Group("/ukm")
			{
				ukm.GET("/groupchat", s.Groupchat.Get)
				ukm.PUT("/groupchat", s.Groupchat.Update)
			}

			// CRUD UKM
			ukmManagement := adminRoutes.Group("/ukms")
			{
				ukmManagement.POST("", s.Ukm.Create)
				ukmManagement.PUT("/:id", s.Ukm.Update)
				ukmManagement.DELETE("/:id", s.Ukm.Delete)
			}

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

		api.GET("/divisions", s.Division.List)
		
		adminManagement := api.Group("/admins")
		adminManagement.Use(middleware.AuthMiddleware("admin"))
		{
			adminManagement.GET("", s.Admin.List)
			adminManagement.POST("", s.Admin.Create)
			adminManagement.PUT("/:id", s.Admin.Update)
			adminManagement.DELETE("/:id", s.Admin.Delete)
		}
	}

	return r
}
