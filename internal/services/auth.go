package services

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"

	"github.com/gin-contrib/sessions"
)

type AuthService struct {
	cfg    *config.Config
	users  *repositories.UserRepository
	admins *repositories.AdminRepository
}

func NewAuthService(cfg *config.Config, users *repositories.UserRepository, admins *repositories.AdminRepository) *AuthService {
	return &AuthService{cfg: cfg, users: users, admins: admins}
}

func init() {
	goth.UseProviders(google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_REDIRECT_URL"), "email", "profile"))
}

func (s *AuthService) BeginGoogleAuth(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", "google")
	c.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (s *AuthService) OAuthCallback(c *gin.Context) {

	q := c.Request.URL.Query()
	q.Add("provider", "google")
	c.Request.URL.RawQuery = q.Encode()
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	// res, err  :=  json.Marshal(user)
	// if  err  !=  nil {
	//     c.AbortWithError(http.StatusInternalServerError, err)
	//     return
	// }

	// jsonString  :=  string(res)

		// set session
		email := user.Email
		parts := strings.Split(email, "@")
		nrp := parts[0]
		institution := parts[1]

		if institution != "john.petra.ac.id" {
			c.Redirect(http.StatusUnauthorized, os.Getenv("CORS_ORIGINS") + "/login")
		}
		
		session := sessions.Default(c)
		session.Set("role", "user")
		session.Set("nrp", nrp)
		session.Set("name", user.Name)
		session.Set("email", user.Email)
	}

	session.Save()

	// sessionData := session.Get("nrp")
	// c.JSON(http.StatusAccepted, sessionData)

	// Handle redirect after login
	redirectURL := os.Getenv("CORS_ORIGINS")

	// If user is admin, check if there was a stored redirect intention
	if admin != nil {
		// For now, redirect admin users to admin dashboard
		redirectURL = os.Getenv("CORS_ORIGINS") + "/admin"
	}

	c.Redirect(http.StatusFound, redirectURL)

}

func (s *AuthService) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "User signed out successfully",
	})
}
