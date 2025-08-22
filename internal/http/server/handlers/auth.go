package handlers

import (
    "github.com/gin-gonic/gin"
    "openhouse-2025-api/internal/services"
)

// var r = gin.Default()
type AuthHandler struct { service *services.AuthService }
func NewAuthHandler(s *services.AuthService) *AuthHandler { return &AuthHandler{service: s} }


func (h *AuthHandler) BeginGoogleAuth(c *gin.Context) {

    h.service.BeginGoogleAuth(c) 
    
}

func (h *AuthHandler) OAuthCallback(c *gin.Context) {

    h.service.OAuthCallback(c) 
    
}

// func  OAuthCallback(c  *gin.Context) {

//     q  :=  c.Request.URL.Query()
//     q.Add("provider", "google")
//     c.Request.URL.RawQuery  =  q.Encode()
//     user, err  :=  gothic.CompleteUserAuth(c.Writer, c.Request)
//     if  err  !=  nil {
//         c.AbortWithError(http.StatusInternalServerError, err)
//         return
//     }
//     res, err  :=  json.Marshal(user)
//     if  err  !=  nil {
//         c.AbortWithError(http.StatusInternalServerError, err)
//         return
//     }

//     jsonString  :=  string(res)
//     c.JSON(http.StatusAccepted, jsonString)
// }