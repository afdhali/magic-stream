package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecureHeaders adds security headers to all responses
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking attacks
		c.Header("X-Frame-Options", "DENY")
		
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection in older browsers
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Force HTTPS connections
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		
		// Content Security Policy - adjust based on your needs
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; img-src 'self' data: https:; font-src 'self' data: https://cdn.jsdelivr.net")
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions Policy (formerly Feature Policy)
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), accelerometer=(), gyroscope=()")
		
		// Remove server identification
		c.Header("X-Powered-By", "")
		c.Header("Server", "")
		
		c.Next()
	}
}

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// You can customize allowed origins based on your requirements
		allowedOrigins := []string{
			"http://localhost:5000",
			"http://localhost:5173",
			"localhost",
		}
		
		// Check if origin is allowed
		originAllowed := false
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				originAllowed = true
				break
			}
		}
		
		if originAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-CSRF-Token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// RateLimiter implements a simple rate limiting mechanism
// You might want to use a more robust solution like github.com/ulule/limiter
func RateLimiter(requestsPerMinute int) gin.HandlerFunc {
	// This is a simplified example. For production, use a proper rate limiting library
	return func(c *gin.Context) {
		// Implementation would go here
		// For now, just pass through
		c.Next()
	}
}

// RequestID adds a unique request ID to each request for tracking
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate a new request ID if not present
			// You can use UUID library for this
			requestID = generateRequestID()
		}
		
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Helper function to generate request ID
func generateRequestID() string {
	// Simple implementation - in production use UUID
	return "req_" + randomString(16)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}