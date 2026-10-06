// Package middleware contains cross-cutting Gin middleware: request ids,
// structured logging and panic recovery.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestIDHeader is echoed on every response and used in every log line.
const RequestIDHeader = "X-Request-ID"

// RequestIDKey is the gin context key holding the request id.
const RequestIDKey = "request_id"

// randRead is a seam so tests can force the crypto/rand failure path.
var randRead = rand.Read

// RequestID ensures every request has a stable identifier.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = generateRequestID()
		}
		c.Set(RequestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

func generateRequestID() string {
	buf := make([]byte, 16)
	if _, err := randRead(buf); err != nil {
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
