package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterAIToolRoutes(
	v1 *gin.RouterGroup,
	toolHandler *handler.AIToolHandler,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	tools := v1.Group("/tools")
	tools.Use(gin.HandlerFunc(jwtAuth))
	tools.Use(middleware.BackendModeUserGuard(settingService))
	{
		tools.GET("", toolHandler.List)
	}
}
