	package services

	import (
		// "encoding/json"
		"net/http"
		"os"

		"github.com/gin-gonic/gin"
		"github.com/markbates/goth"
		"github.com/markbates/goth/gothic"
		"github.com/markbates/goth/providers/google"

		"openhouse-2025-api/internal/config"
		"openhouse-2025-api/internal/repositories"

		// for session
		"github.com/gin-contrib/sessions"
		// gsessions "github.com/gorilla/sessions"
		"strings"
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
		nrp := strings.Split(email, "@")[0]
		
		session := sessions.Default(c)
		session.Set("nrp", nrp)

		session.Save()
		

		// sessionData := session.Get("nrp")
		// c.JSON(http.StatusAccepted, sessionData)

		// gw gtw mau arahin ke mana
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




