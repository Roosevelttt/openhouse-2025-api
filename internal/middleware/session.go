package middleware

import (
	"log"
	"os"

	"openhouse-2025-api/internal/config"

    "github.com/markbates/goth/gothic"
    "github.com/gin-contrib/sessions"
    // "github.com/gin-contrib/sessions/cookie"
    "github.com/gin-contrib/sessions/redis"
    "github.com/gin-gonic/gin"
)

func SessionManager(cfg *config.Config) gin.HandlerFunc {
	// hashkey hrs ada
	hashKey := os.Getenv("SESSION_HASH_KEY")
	if hashKey == "" {
		log.Fatal("SESSION_HASH_KEY environment variable is not set.")
	}

    // redisAddr := os.Getenv("REDIS_ADDR")
    redisAddr := "localhost:8001"
    // log.Fatal("REDIS_ADDR :" + redisAddr)
    if redisAddr == "" {
        log.Fatal("REDIS_ADDR environment variable is not set.")
    }

    
    // store := cookie.NewStore([]byte(hashKey))
    store, err := redis.NewStore(10, "tcp", redisAddr, "waxomoly", "", []byte(hashKey))
    if err != nil {
        log.Fatalf("Could not connect to Redis: %v", err)
    }
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