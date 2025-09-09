package router

import (
	"net/http"
	"log"
	// "os"

	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/http/server"
	"openhouse-2025-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
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

	public := r.Group("/api")
	{
		public.GET("/auth/google/start", s.Auth.BeginGoogleAuth)
		public.GET("/auth/google/callback", s.Auth.OAuthCallback)
		
		
		public.GET("/csrf-tok", func(c *gin.Context) {
			// Retrieve the token associated with the user's session.
			// This function is provided by the gorilla/csrf library.
			token := csrf.Token(c.Request)
			log.Printf("CSRF Token generated: %s", token)
			
			// Set the token in a custom response header. Your frontend will read this.
			c.Header("X-CSRF-Token", token)
			
			// Send a success response with a simple JSON body.
			c.JSON(http.StatusOK, gin.H{
				"message": "CSRF token provided",
				"csrfToken": token, // fallback
			})
		})
		
		public.GET("/ukms", s.Ukm.List)
		public.GET("/ukms/:id", s.Ukm.GetByID)
		public.GET("/ukms/slug/:slug", s.Ukm.GetBySlug)
		public.GET("/divisions", s.Division.List)
		public.GET("/admins", s.Admin.List)
		
		public.GET("/debug/session", s.Session.DebugSession)
	}

	protected := r.Group("/api")
	protected.Use(middleware.CSRF(cfg))
	{
		protected.POST("/auth/logout", s.Auth.Logout)

		userRoutes := protected.Group("/user")
		{
			userRoutes.POST("/session/values", s.Session.GetValues)
		}

		userRoutes = protected.Group("/user")
		userRoutes.Use(middleware.AuthMiddleware("user", "admin"))
		{
			userRoutes.GET("/biodata", s.User.GetBiodata)
			userRoutes.POST("/biodata", s.User.UpdateBiodata)
		}

		// Registration routes
		registrationRoutes := protected.Group("/registrations")
		registrationRoutes.Use(middleware.AuthMiddleware("user", "admin"))
		{
			registrationRoutes.POST("/reserve", s.Participants.ReserveSlot)
			registrationRoutes.POST("/access-payment/:ukm_id", s.Participants.AccessPaymentPage)
			registrationRoutes.GET("/check-reservation/:ukm_id", s.Participants.CheckUserReservation)
			registrationRoutes.POST("/with-reservation/:reservationId", s.Participants.RegisterWithReservation)
			registrationRoutes.GET("/check-registration/:ukm_id", s.Participants.CheckRegistration)
		}

		// Admin routes
		adminRoutes := protected.Group("/admin")
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

		adminManagement := protected.Group("/admins")
		adminManagement.Use(middleware.AuthMiddleware("admin"))
		{
			adminManagement.POST("", s.Admin.Create)
			adminManagement.PUT("/:id", s.Admin.Update)
			adminManagement.DELETE("/:id", s.Admin.Delete)
		}
	}

	return r
}
