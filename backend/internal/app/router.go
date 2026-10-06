package app

import (
	"log/slog"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "indico/backend/docs" // generated swagger spec
	"indico/backend/internal/modules/inventory"
	"indico/backend/internal/platform/config"
	"indico/backend/internal/platform/httpx"
	"indico/backend/internal/platform/middleware"
	"indico/backend/internal/platform/validator"
)

// registerValidator is a seam so tests can force the registration failure.
var registerValidator = validator.Register

// BuildRouter assembles the gin engine: global middleware, platform routes
// (health + swagger), module routes, and JSON fallbacks for unknown routes.
func BuildRouter(cfg config.Config, log *slog.Logger, inventoryModule *inventory.Module) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	if err := registerValidator(); err != nil {
		log.Error("gagal mendaftarkan validator", "error", err)
	}

	engine := gin.New()
	engine.HandleMethodNotAllowed = true

	engine.Use(
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Recovery(log),
	)
	engine.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{"Content-Type", "Accept", middleware.RequestIDHeader},
	}))

	engine.GET("/api/v1/healthz", func(c *gin.Context) {
		httpx.Success(c, http.StatusOK, gin.H{"status": "ok"})
	})
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := engine.Group("/api/v1")
	inventoryModule.RegisterRoutes(v1)

	engine.NoRoute(func(c *gin.Context) { httpx.Error(c, httpx.ErrRouteNotFound) })
	engine.NoMethod(func(c *gin.Context) { httpx.Error(c, httpx.ErrMethodNotAllowed) })
	return engine
}
