package middleware

import (
    "net/http"
    "github.com/gin-contrib/sessions"
    "github.com/gin-gonic/gin"
)

func Authentication(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        
        session_role := session.Get("role")

        if session_role != role {
            c.JSON(http.StatusUnauthorized, gin.H{
                "message": "unauthorized",
            })
            c.Abort()
            return
        }
        c.Next() 
    }
}