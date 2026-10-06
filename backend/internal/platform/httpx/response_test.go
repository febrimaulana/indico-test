package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return context, recorder
}

func decodeEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()
	var envelope ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	return envelope
}

func TestSuccess(t *testing.T) {
	context, recorder := newTestContext()
	Success(context, http.StatusCreated, map[string]string{"status": "success"})

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.JSONEq(t, `{"status":"success"}`, recorder.Body.String())
}

func TestErrorWithAppError(t *testing.T) {
	context, recorder := newTestContext()
	appErr := New(http.StatusConflict, CodeInsufficientStock, "stok tidak mencukupi").
		WithDetails(ErrorDetail{Field: "quantity", Rule: "gt", Message: "pesan"})
	Error(context, appErr)

	require.Equal(t, http.StatusConflict, recorder.Code)
	envelope := decodeEnvelope(t, recorder)
	require.Equal(t, "error", envelope.Status)
	require.Equal(t, CodeInsufficientStock, envelope.Error.Code)
	require.Equal(t, "stok tidak mencukupi", envelope.Error.Message)
	require.Len(t, envelope.Error.Details, 1)
}

func TestErrorWithWrappedAppError(t *testing.T) {
	context, recorder := newTestContext()
	expired := New(http.StatusGone, CodeReservationExpired, "reservasi sudah kedaluwarsa")
	Error(context, fmt.Errorf("konteks tambahan: %w", expired))

	require.Equal(t, http.StatusGone, recorder.Code)
	require.Equal(t, CodeReservationExpired, decodeEnvelope(t, recorder).Error.Code)
}

func TestErrorWithUnknownError(t *testing.T) {
	context, recorder := newTestContext()
	Error(context, errors.New("rahasia internal"))

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Equal(t, CodeInternalError, decodeEnvelope(t, recorder).Error.Code)
	require.NotContains(t, recorder.Body.String(), "rahasia internal")
}
