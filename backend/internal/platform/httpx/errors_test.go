package httpx

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppErrorWithoutCause(t *testing.T) {
	err := New(400, CodeInvalidRequest, "pesan uji")
	require.Equal(t, "INVALID_REQUEST: pesan uji", err.Error())
	require.NoError(t, err.Unwrap())
}

func TestAppErrorWithCause(t *testing.T) {
	cause := errors.New("penyebab")
	err := ErrInternal.Wrap(cause)
	require.Contains(t, err.Error(), "penyebab")
	require.ErrorIs(t, err, cause)
}

func TestWithDetailsReturnsCopy(t *testing.T) {
	base := ErrInvalidRequest
	detailed := base.WithDetails(ErrorDetail{Field: "quantity", Rule: "gt", Message: "message"})
	require.Empty(t, base.Details)
	require.Len(t, detailed.Details, 1)
	require.Len(t, base.Details, 0)
}

func TestInternal(t *testing.T) {
	err := Internal(errors.New("rahasia"))
	require.Equal(t, 500, err.Status)
	require.Equal(t, CodeInternalError, err.Code)
	require.Equal(t, "terjadi kesalahan pada server", err.Message)
}
