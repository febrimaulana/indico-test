package middleware

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"

	"indico/backend/internal/platform/httpx"
)

// Recovery converts panics into the frozen 500 envelope so a single bad
// request can never take the server down.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error("panic ditangkap",
					"request_id", c.GetString(RequestIDKey),
					"panic", fmt.Sprint(recovered),
				)
				httpx.Error(c, httpx.Internal(fmt.Errorf("panic: %v", recovered)))
			}
		}()
		c.Next()
	}
}
