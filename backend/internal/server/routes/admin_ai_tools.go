package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/gin-gonic/gin"
)

func registerAdminAIToolRoutes(adminGroup *gin.RouterGroup, h *admin.AIToolHandler) {
	tools := adminGroup.Group("/tools")
	{
		tools.GET("", h.List)
		tools.POST("", h.Create)
		tools.PUT("/:id", h.Update)
		tools.DELETE("/:id", h.Delete)
	}
}
