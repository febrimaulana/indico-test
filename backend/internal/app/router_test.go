package app

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"indico/backend/internal/modules/inventory"
	"indico/backend/internal/platform/config"
	"indico/backend/internal/platform/middleware"
)

func testRouter(t *testing.T) http.Handler {
	t.Helper()
	log := testLog()
	module := inventory.NewModule(nil, log, inventory.ModuleConfig{
		ExpiryDuration: time.Minute,
		SweepInterval:  time.Minute,
		SweepBatchSize: 10,
	})
	return BuildRouter(config.Config{CORSAllowedOrigins: []string{"*"}}, log, module)
}

func perform(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestBuildRouterHealthz(t *testing.T) {
	recorder := perform(testRouter(t), http.MethodGet, "/api/v1/healthz", "")

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
	require.NotEmpty(t, recorder.Header().Get(middleware.RequestIDHeader))
}

func TestBuildRouterUnknownRoute(t *testing.T) {
	recorder := perform(testRouter(t), http.MethodGet, "/tidak-ada", "")

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Contains(t, recorder.Body.String(), "ROUTE_NOT_FOUND")
}

func TestBuildRouterMethodNotAllowed(t *testing.T) {
	recorder := perform(testRouter(t), http.MethodPost, "/api/v1/healthz", "")

	require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	require.Contains(t, recorder.Body.String(), "METHOD_NOT_ALLOWED")
}

func TestBuildRouterInventoryRouteRegistered(t *testing.T) {
	recorder := perform(testRouter(t), http.MethodPost, "/api/v1/inventory/reserve", `{}`)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "INVALID_REQUEST")
}

func TestBuildRouterSwaggerServed(t *testing.T) {
	recorder := perform(testRouter(t), http.MethodGet, "/swagger/doc.json", "")

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "INDICO")
}

func TestBuildRouterValidatorRegistrationError(t *testing.T) {
	original := registerValidator
	t.Cleanup(func() { registerValidator = original })
	registerValidator = func() error { return errors.New("registrasi gagal") }

	recorder := perform(testRouter(t), http.MethodGet, "/api/v1/healthz", "")
	require.Equal(t, http.StatusOK, recorder.Code)
}
