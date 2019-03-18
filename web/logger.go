package web

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	timeFormat = "2006-01-02T15:04:05-0700"
)

// Logger instance returns a middleware using logrus to log the access information.
func Logger(notlogged ...string) gin.HandlerFunc {
	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			// Stop timer
			end := time.Now()
			latency := end.Sub(start)

			clientIP := c.ClientIP()
			method := c.Request.Method
			statusCode := c.Writer.Status()
			comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

			if raw != "" {
				path = path + "?" + raw
			}

			log.WithFields(log.Fields{
				"app":        "GIN",
				"time":       end.Format(timeFormat),
				"latency":    latency,
				"clientIP":   clientIP,
				"method":     method,
				"path":       path,
				"comment":    comment,
				"statusCode": statusCode,
			}).Info("HTTP Access")
		}
	}
}
