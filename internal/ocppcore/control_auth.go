package ocppcore

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireControlToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := strings.TrimSpace(os.Getenv("OCPP_CONTROL_TOKEN"))
		if expected == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "OCPP_CONTROL_TOKEN is not configured",
			})
			c.Abort()
			return
		}

		got := strings.TrimSpace(c.GetHeader("X-Control-Token"))
		if got == "" {
			auth := strings.TrimSpace(c.GetHeader("Authorization"))
			if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
				got = strings.TrimSpace(auth[7:])
			}
		}

		if got == "" || got != expected {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid control token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}