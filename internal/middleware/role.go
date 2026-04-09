package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/om-1980/ChargeMitra_backend/pkg/response"
)

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}

	return func(c *gin.Context) {
		roleValue, exists := c.Get("user_role")
		if !exists {
			response.Forbidden(c, "role not found in token")
			c.Abort()
			return
		}

		role, ok := roleValue.(string)
		if !ok || !allowed[role] {
			response.Forbidden(c, "access denied")
			c.Abort()
			return
		}

		c.Next()
	}
}