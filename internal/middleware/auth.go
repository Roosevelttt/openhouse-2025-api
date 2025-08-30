package middleware

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionRole, ok := session.Get("role").(string)
		if !ok || sessionRole == "" {
			log.Println("Authentication failed: role not found in session")
			unauthorized(c)
			return
		}

		isAllowed := false
		for _, role := range allowedRoles {
			if sessionRole == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			log.Printf("Access denied: role '%s' is not in the allowed list %v", sessionRole, allowedRoles)
			unauthorized(c)
			return
		}

		// Set user context based on the verified role
		if err := setContextFromSession(c, session, sessionRole); err != nil {
			log.Printf("Authentication failed: could not set context. Error: %v", err)
			unauthorized(c)
			return
		}

		log.Printf("Access granted for role: '%s'", sessionRole)
		c.Next()
	}
}

func setContextFromSession(c *gin.Context, session sessions.Session, role string) error {
	c.Set("user_role", role)

	switch role {
	case "admin":
		adminNRP := session.Get("nrp")
		adminID := session.Get("admin_id")
		adminName := session.Get("admin_name")
		
		c.Set("admin_nrp", adminNRP)
		c.Set("admin_id", adminID)
		c.Set("admin_name", adminName)
		
		c.Set("admin_ukm_id", session.Get("admin_ukm_id"))
		c.Set("admin_division_id", session.Get("admin_division_id"))

	case "user":
		userNRP, ok := session.Get("nrp").(string)
		if !ok {
			return log.Output(1, "invalid or missing user nrp in session")
		}
		c.Set("user_nrp", userNRP)
	}
	return nil
}

func unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
	c.Abort()
}