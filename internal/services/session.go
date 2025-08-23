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