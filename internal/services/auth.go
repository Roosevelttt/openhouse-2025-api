package services

import (
	// "context"
	// "errors"
	// "fmt"
	"encoding/json"
	"net/http"
	"os"

    "github.com/gin-gonic/gin"
    "github.com/markbates/goth"
    "github.com/markbates/goth/gothic"
    "github.com/markbates/goth/providers/google"

	"openhouse-2025-api/internal/config"
	// "openhouse-2025-api/internal/models"
	"openhouse-2025-api/internal/repositories"
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
    res, err  :=  json.Marshal(user)
    if  err  !=  nil {
        c.AbortWithError(http.StatusInternalServerError, err)
        return
    }

    jsonString  :=  string(res)
    c.JSON(http.StatusAccepted, jsonString)
}

// func NewAuthService(cfg *config.Config, users *repositories.UserRepository) *AuthService {
// 	return &AuthService{cfg: cfg, users: users}
// }

// func (s *AuthService) GetGoogleAuthURL() string {
// 	// Placeholder for Google OAuth URL construction; swap with oauth2.Config.
// 	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=profile email", s.cfg.GoogleClientID, s.cfg.GoogleRedirectURL)
// }

// func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (string, *models.User, error) {
// 	if code == "" { return "", nil, errors.New("missing code") }
// 	// TODO: Exchange code with Google, fetch profile; for now return mock user.
// 	u := &models.User{NRP: "sample", Name: "Sample User"}
// 	if err := s.users.UpsertByNRP(ctx, u); err != nil { return "", nil, err }
// 	// TODO: issue JWT signed with s.cfg.JWTSecret
// 	return "mock-jwt", u, nil
// }

