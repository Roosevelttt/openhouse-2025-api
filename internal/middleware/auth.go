package middleware

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Authentication(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		session_role := session.Get("role")

		log.Printf("Authentication middleware: expected role='%s', session role='%v'", role, session_role)

		if session_role != role {
			log.Printf("Authentication failed: role mismatch")
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "unauthorized",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}


func AuthenticationWithRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session_role := session.Get("role")

		log.Printf("AuthenticationWithRoles: allowed roles=%v, session role='%v'", allowedRoles, session_role)

		for _, allowedRole := range allowedRoles {
			if session_role == allowedRole {
				log.Printf("Access granted: session role '%v' matches allowed role '%s'", session_role, allowedRole)

				// Set user context based on role
				switch session_role {
                case "admin":
					adminNRP := session.Get("nrp")
					adminID := session.Get("admin_id")
					adminName := session.Get("admin_name")
					adminUkmID := session.Get("admin_ukm_id")
					adminDivisionID := session.Get("admin_division_id")

					c.Set("admin_nrp", adminNRP)
					c.Set("admin_id", adminID)
					c.Set("admin_name", adminName)
					c.Set("admin_ukm_id", adminUkmID)
					c.Set("admin_division_id", adminDivisionID)
					c.Set("user_role", "admin")
				case "user":
					userNRP := session.Get("nrp")
					c.Set("user_nrp", userNRP)
					c.Set("user_role", "user")
				}

				c.Next()
				return
			}
		}

		log.Printf("Access denied: session role '%v' not in allowed roles %v", session_role, allowedRoles)
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		c.Abort()
	}
}

func AdminAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		// Check if user has admin role
		session_role := session.Get("role")
		if session_role != "admin" {
			log.Printf("Admin authentication failed: role='%v', expected='admin'", session_role)
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "admin access required",
			})
			c.Abort()
			return
		}

		// Set admin context in gin context for handlers to use
		adminNRP := session.Get("nrp")
		adminID := session.Get("admin_id")
		adminName := session.Get("admin_name")
		adminUkmID := session.Get("admin_ukm_id")
		adminDivisionID := session.Get("admin_division_id")

		if adminNRP == nil || adminID == nil || adminName == nil {
			log.Printf("Admin authentication failed: missing session data")
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid admin session",
			})
			c.Abort()
			return
		}

		// Set admin context for handlers
		c.Set("admin_nrp", adminNRP)
		c.Set("admin_id", adminID)
		c.Set("admin_name", adminName)
		c.Set("admin_ukm_id", adminUkmID)
		c.Set("admin_division_id", adminDivisionID)

		log.Printf("Admin authenticated: NRP=%s, Name=%s", adminNRP, adminName)
		c.Next()
	}
}
