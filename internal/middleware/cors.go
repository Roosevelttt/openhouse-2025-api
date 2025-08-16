package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"openhouse-2025-api/internal/config"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	allowed := strings.Split(cfg.CORSOrigins, ",")
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allow := "*"
		for _, o := range allowed {
			if o == "*" || o == origin { allow = origin; break }
		}
		c.Header("Access-Control-Allow-Origin", allow)
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" { c.AbortWithStatus(204); return }
		c.Next()
	}
}

