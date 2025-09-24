package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a gin.HandlerFunc that logs HTTP requests.
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CORS middleware to handle Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Auth middleware for authentication (placeholder implementation)
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement authentication logic
		// For now, this is a placeholder that allows all requests
		c.Next()
	}
}

// RateLimit middleware for rate limiting (placeholder implementation)
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement rate limiting logic
		// For now, this is a placeholder that allows all requests
		c.Next()
	}
}