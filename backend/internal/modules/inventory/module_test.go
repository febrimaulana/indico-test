package inventory

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestNewModuleWiresEverything(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	module := NewModule(nil, log, ModuleConfig{
		ExpiryDuration: time.Minute,
		SweepInterval:  time.Minute,
		SweepBatchSize: 10,
	})

	require.NotNil(t, module)
	require.NotNil(t, module.Handler)
	require.NotNil(t, module.Service)
	require.NotNil(t, module.ExpiryWorker)
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	module := NewModule(nil, log, ModuleConfig{ExpiryDuration: time.Minute, SweepInterval: time.Minute, SweepBatchSize: 10})

	engine := gin.New()
	module.RegisterRoutes(engine.Group("/api/v1"))

	routes := map[string]bool{}
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	require.True(t, routes["POST /api/v1/inventory/reserve"])
	require.True(t, routes["POST /api/v1/inventory/confirm"])
	require.True(t, routes["GET /api/v1/inventory/stock"])
}
