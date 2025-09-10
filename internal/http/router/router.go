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

	api := r.Group("/api")
	{

		api.GET("/csrf-tok", middleware.CSRF(cfg), func(c *gin.Context) {
			// Retrieve the token associated with the user's session.
			// This function is provided by the gorilla/csrf library.
			token := csrf.Token(c.Request)
			log.Printf("CSRF Token generated: %s", token)

			if token == "" {
				log.Printf("Request details: URL=%s, Method=%s", c.Request.URL.Path, c.Request.Method)
			}

			// Set the token in a custom response header. Your frontend will read this.
			c.Header("X-CSRF-Token", token)

			// Send a success response with a simple JSON body.
			c.JSON(http.StatusOK, gin.H{
				"message": "CSRF token provided",
				"csrfToken": token, // fallback
			})
		})

		api.GET("/debug/session", s.Session.DebugSession)
		api.GET("/auth/google/start", s.Auth.BeginGoogleAuth)
		api.GET("/auth/google/callback", s.Auth.OAuthCallback)
		api.POST("/auth/logout", middleware.CSRF(cfg), s.Auth.Logout)

		api.GET("/ukms", s.Ukm.List)
		api.GET("/ukms/:id", s.Ukm.GetByID)
		api.GET("/ukms/slug/:slug", s.Ukm.GetBySlug)

		userRoutes := api.Group("/user")
		userRoutes.Use(middleware.AuthMiddleware("user", "admin"))
		{
			userRoutes.POST("/session/values", s.Session.GetValues)
			userRoutes.GET("/biodata", s.User.GetBiodata)
			userRoutes.POST("/biodata", middleware.CSRF(cfg), s.User.UpdateBiodata)
		}

		registrationRoutes := api.Group("/registrations")
		registrationRoutes.Use(middleware.AuthMiddleware("user", "admin"))
		{
			registrationRoutes.POST("/reserve", middleware.CSRF(cfg), s.Participants.ReserveSlot)
			registrationRoutes.POST("/access-payment/:ukm_id", middleware.CSRF(cfg), s.Participants.AccessPaymentPage)
			registrationRoutes.GET("/check-reservation/:ukm_id", s.Participants.CheckUserReservation)
			registrationRoutes.POST("/with-reservation/:reservationId", middleware.CSRF(cfg), s.Participants.RegisterWithReservation)
			registrationRoutes.GET("/check-registration/:ukm_id", s.Participants.CheckRegistration)
		}

		api.GET("/registrations/test", s.Registration.Test)
		api.POST("/registrations", middleware.CSRF(cfg), s.Registration.Create)

		adminRoutes := api.Group("/admin")
		adminRoutes.Use(middleware.AuthMiddleware("admin"))
		{
			adminRoutes.GET("/participants", s.Participants.List)

			ukm := adminRoutes.Group("/ukm")
			{
				ukm.GET("/groupchat", s.Groupchat.Get)
				ukm.PUT("/groupchat", middleware.CSRF(cfg), s.Groupchat.Update)
			}

			// CRUD UKM
			ukmManagement := adminRoutes.Group("/ukms")
			{
				ukmManagement.POST("", middleware.CSRF(cfg), s.Ukm.Create)
				ukmManagement.PUT("/:id", middleware.CSRF(cfg), s.Ukm.Update)
				ukmManagement.DELETE("/:id", middleware.CSRF(cfg), s.Ukm.Delete)
			}

			payment := adminRoutes.Group("/payment")
			{
				payment.POST("/validate", middleware.CSRF(cfg), s.Validation.Validate)
				payment.POST("/reject", middleware.CSRF(cfg), s.Validation.Reject)
			}

			export := adminRoutes.Group("/export")
			{
				export.GET("/participants", s.Export.ExportParticipants)
				export.POST("/participants/daily-recap", s.Export.TriggerDailyRecap)
				export.GET("/participants/google-sheets/start", s.Export.GoogleSheetsOAuthStart)
				export.GET("/participants/google-sheets/callback", s.Export.GoogleSheetsOAuthCallback)
			}
		}

		api.GET("/divisions", s.Division.List)

		adminManagement := api.Group("/admins")
		adminManagement.Use(middleware.AuthMiddleware("admin"))
		{
			adminManagement.GET("", s.Admin.List)
			adminManagement.POST("", middleware.CSRF(cfg), s.Admin.Create)
			adminManagement.PUT("/:id", middleware.CSRF(cfg), s.Admin.Update)
			adminManagement.DELETE("/:id", middleware.CSRF(cfg), s.Admin.Delete)
		}
	}

	return r
}
