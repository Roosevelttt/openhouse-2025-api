package services

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SessionService struct {
}

func NewSessionService() *SessionService {
	return &SessionService{}
}

type KeyRequest struct {
	Keys []string `json:"keys"`
}

func (s *SessionService) GetSessionValues(c *gin.Context) {
	session := sessions.Default(c)

	// check if user is authenticated
	nrp := session.Get("nrp")
	if nrp == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req KeyRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid request body: %s", err.Error()),
		})
		return
	}

	values := make(map[string]interface{})
	for _, key := range req.Keys {
		value := session.Get(key)

		if value != nil {
			values[key] = value
		} else {
			values[key] = "UNKNOWN KEY or EMPTY STRING"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"session_values": values,
	})
}

// DebugSession inspects all session contents for troubleshooting
func (s *SessionService) DebugSession(c *gin.Context) {
	session := sessions.Default(c)

	// Get all common session keys
	sessionData := map[string]interface{}{
		"role":              session.Get("role"),
		"nrp":               session.Get("nrp"),
		"admin_id":          session.Get("admin_id"),
		"admin_name":        session.Get("admin_name"),
		"admin_ukm_id":      session.Get("admin_ukm_id"),
		"admin_division_id": session.Get("admin_division_id"),
	}

	c.JSON(http.StatusOK, gin.H{
		"debug_session": sessionData,
		"session_id":    session.ID(),
		"has_session":   session.Get("nrp") != nil,
		"is_authenticated": sessionData["role"] != nil,
		"is_admin":      sessionData["role"] == "admin",
		"is_user":       sessionData["role"] == "user",
		"message":       "Session debug information",
	})
}