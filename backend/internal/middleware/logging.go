package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger returns a Gin middleware that emits structured request logs via zap.
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.String("user_id", c.GetString("user_id")),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				log.Error("request error", append(fields, zap.String("error", e))...)
			}
			return
		}

		switch {
		case status >= 500:
			log.Error("server error", fields...)
		case status >= 400:
			log.Warn("client error", fields...)
		default:
			log.Info("request", fields...)
		}
	}
}

// Recovery returns a Gin middleware that recovers from panics and logs them.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered", zap.Any("error", r))
				c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}
