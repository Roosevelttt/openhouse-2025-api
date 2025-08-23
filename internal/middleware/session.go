package middleware

import (
    "os"
    "log"

    "github.com/markbates/goth/gothic"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-gonic/gin"
)

func SessionManager() gin.HandlerFunc {
    // hashkey hrs ada
    hashKey := os.Getenv("SESSION_HASH_KEY")
    if hashKey == "" {
        log.Fatal("SESSION_HASH_KEY environment variable is not set.")
    }

    
    store := cookie.NewStore([]byte(hashKey))
    store.Options(sessions.Options{
        MaxAge: 60 * 60 * 24 * 7, // 7 days in seconds
    })
    gothic.Store = store

    return sessions.Sessions("mysession", store)
}

// func main() {
//     r := gin.Default()
//     store := cookie.NewStore([]byte("secret"))
//     r.Use(sessions.Sessions("mysession", store))
//     // ...
// }

// import (

//     "net/http"
//     "github.com/gin-contrib/sessions"
// 	"github.com/gin-gonic/gin"
// )

// func Authentication() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         session := sessions.Default(c)
//         sessionID := session.Get("id")

//         if sessionID == nil {
//             c.JSON(http.StatusNotFound, gin.H{
//                 "message": "unauthorized",
//             })
//             c.Abort()
//         }
//     }
// }