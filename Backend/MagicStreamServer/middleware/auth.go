// middleware/auth.go
package middleware

import (
	"net/http"

	authservice "github.com/afdhali/magic-stream/Backend/MagicStreamServer/controllers/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT access token and extracts user ID
// Uses dependency injection instead of global variable
func AuthMiddleware(ts *authservice.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing Authorization header",
			})
			return
		}

		// Extract token from "Bearer <token>" format
		var tokenStr string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenStr = authHeader[7:]
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization format. Use: Bearer <token>",
			})
			return
		}

		// Validate token and extract user ID
		userID, err := ts.ValidateAccessToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Store user ID in context for use in handlers
		c.Set("user_id", userID)
		c.Next()
	}
}

// AdminOnly middleware checks if user has admin role
// Must be used after AuthMiddleware
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by a previous middleware or handler)
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "User role not found",
			})
			return
		}

		if role != "ADMIN" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			return
		}

		c.Next()
	}
}

// GetUserID extracts user ID from gin context
// Helper function for handlers
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}