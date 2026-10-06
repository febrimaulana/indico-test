package inventory

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts every inventory endpoint on the given group. This is
// the single place where this module's routes are declared.
func (m *Module) RegisterRoutes(group *gin.RouterGroup) {
	inventory := group.Group("/inventory")
	inventory.POST("/reserve", m.Handler.Reserve)
	inventory.POST("/confirm", m.Handler.Confirm)
	inventory.GET("/stock", m.Handler.GetStock)
}
