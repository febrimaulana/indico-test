package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func testLogger(buffer *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buffer, nil))
}

func TestRequestIDGeneratesWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/", func(c *gin.Context) {
		require.NotEmpty(t, c.GetString(RequestIDKey))
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, recorder.Code)

	generated := recorder.Header().Get(RequestIDHeader)
	require.Regexp(t, regexp.MustCompile(`^[0-9a-f]{32}$`), generated)
}

func TestRequestIDPreservesIncoming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(RequestIDHeader, "req-dari-klien")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	require.Equal(t, "req-dari-klien", recorder.Header().Get(RequestIDHeader))
}

func TestRequestIDRandFailureFallback(t *testing.T) {
	original := randRead
	t.Cleanup(func() { randRead = original })
	randRead = func([]byte) (int, error) { return 0, errors.New("boom") }

	id := generateRequestID()
	require.True(t, strings.HasPrefix(id, "req_"))
	require.Greater(t, len(id), len("req_"))
}

func TestLoggerWritesRequestLine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buffer bytes.Buffer
	engine := gin.New()
	engine.Use(RequestID(), Logger(testLogger(&buffer)))
	engine.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ping", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, buffer.String(), "request selesai")
	require.Contains(t, buffer.String(), "/ping")
}

func TestRecoveryReturnsEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buffer bytes.Buffer
	engine := gin.New()
	engine.Use(RequestID(), Recovery(testLogger(&buffer)))
	engine.GET("/panic", func(*gin.Context) { panic("boom") })

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/panic", nil))

	require.Equal(t, http.StatusInternalServerError, recorder.Code)

	var envelope struct {
		Status string `json:"status"`
		Error  struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.Equal(t, "error", envelope.Status)
	require.Equal(t, "INTERNAL_ERROR", envelope.Error.Code)
	require.Contains(t, buffer.String(), "panic ditangkap")
}
