// package googleAuth

// import (
//     "encoding/json"
//     "net/http"

//     "github.com/gin-gonic/gin"
//     "github.com/markbates/goth"
//     "github.com/markbates/goth/gothic"
//     "github.com/markbates/goth/providers/google"
// )

// var r = gin.Default()

// func init() {
//     goth.UseProviders(google.New("922329790485-55jrc60i4v62j3n99nqpqbqhesd93on2.apps.googleusercontent.com", "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", "http://localhost:8080/v1/api/auth/google/callback", "email", "profile"))
// }

// func  BeginGoogleAuth(c  *gin.Context) {
//     q  :=  c.Request.URL.Query()
//     q.Add("provider", "google")
//     c.Request.URL.RawQuery  =  q.Encode()
//     gothic.BeginAuthHandler(c.Writer, c.Request)
// }

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