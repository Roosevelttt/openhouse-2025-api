package middleware

import (
	"log"
	"net/http"
	"openhouse-2025-api/internal/config" // Import your config package

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	adapter "github.com/gwatts/gin-adapter"
)

// CSRF creates a Gin middleware that provides CSRF protection.
// It is dynamically configured based on your application's environment.
func CSRF(cfg *config.Config) gin.HandlerFunc {
	// 1. Determine cookie settings based on the environment (e.g., "debug" or "release").
	isSecure := (cfg.GinMode == "release")
	sameSiteMode := csrf.SameSiteLaxMode
	if isSecure {
		// SameSiteNoneMode is required for cross-site requests in production (HTTPS).
		sameSiteMode = csrf.SameSiteNoneMode
	}

	// 2. Configure the csrf.Protect middleware with dynamic options.
	csrfMiddleware := csrf.Protect(
		// Use the secret key from your application's configuration.
		[]byte(cfg.CSRFAuthKey),

		// Set cookie flags based on the environment to work locally (HTTP) and in production (HTTPS).
		csrf.Secure(isSecure),
		csrf.SameSite(sameSiteMode),

		// Set the cookie path to "/" to ensure it's sent for all requests on your site.
		csrf.Path("/"),

		// Provide a user-friendly JSON error response on failure.
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message": "Forbidden - CSRF token is invalid or missing"}`))
			tokenFromCookie := csrf.Token(r)
			log.Printf("Expected Token yyy: %s", csrf.Token(r))
            log.Printf("Expected Token zzz: %s", tokenFromCookie)
			log.Printf("Received Token (from header): %s", r.Header.Get("X-CSRF-Token"))
		})),
	)

	// 3. Use the adapter to convert the standard http.Handler into a gin.HandlerFunc.
	return adapter.Wrap(csrfMiddleware)
}