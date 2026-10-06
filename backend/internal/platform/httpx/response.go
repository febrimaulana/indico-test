package httpx

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// Success writes a success payload with the given HTTP status.
func Success(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}

// Error maps any error to the frozen error envelope. Unknown errors become a
// 500 and the internal cause is never leaked to the client.
func Error(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.AbortWithStatusJSON(appErr.Status, ErrorResponse{
			Status: "error",
			Error: ErrorInfo{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
		return
	}
	c.AbortWithStatusJSON(ErrInternal.Status, ErrorResponse{
		Status: "error",
		Error: ErrorInfo{
			Code:    ErrInternal.Code,
			Message: ErrInternal.Message,
		},
	})
}
