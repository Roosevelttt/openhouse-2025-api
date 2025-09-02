	package services

	import (
		"strings"
		"net/http"
		"os"

		"github.com/gin-gonic/gin"
		"github.com/markbates/goth"
		"github.com/markbates/goth/gothic"
		"github.com/markbates/goth/providers/google"

		"openhouse-2025-api/internal/config"
		"openhouse-2025-api/internal/repositories"

		"github.com/gin-contrib/sessions"
	)

	type AuthService struct {
		cfg *config.Config
		users *repositories.UserRepository
	}

	func NewAuthService(cfg *config.Config, users *repositories.UserRepository) *AuthService {
		return &AuthService{cfg: cfg, users: users}
	}

	func init() {
		goth.UseProviders(google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_REDIRECT_URL"), "email", "profile"))
	}

	func (s *AuthService) BeginGoogleAuth(c  *gin.Context) {
		q  :=  c.Request.URL.Query()
		q.Add("provider", "google")
		c.Request.URL.RawQuery  =  q.Encode()
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}

	func (s *AuthService) OAuthCallback(c  *gin.Context) {

		q  :=  c.Request.URL.Query()
		q.Add("provider", "google")
		c.Request.URL.RawQuery  =  q.Encode()
		user, err  :=  gothic.CompleteUserAuth(c.Writer, c.Request)
		if  err  !=  nil {
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

		session.Save()
		

		// sessionData := session.Get("nrp")
		// c.JSON(http.StatusAccepted, sessionData)

		// gw gtw mau arahin ke mana hehe
		c.Redirect(http.StatusFound, os.Getenv("CORS_ORIGINS") + "")
		
	}

	func (s *AuthService) Logout(c  *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()

		c.JSON(http.StatusAccepted, gin.H{
			"message": "User signed out successfully",
		})
	}




