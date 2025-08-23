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
        // KELLY LIHAT INI LAGI
        Path:     "/",
        HttpOnly: true,  
        SameSite: 0, 
		Secure:   true,
    })
    gothic.Store = store

    return sessions.Sessions("mysession", store)
}