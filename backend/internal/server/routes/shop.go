package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterShopRoutes(
	v1 *gin.RouterGroup,
	shopHandler *handler.ShopHandler,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
) {
	shop := v1.Group("/shop")
	shop.Use(gin.HandlerFunc(jwtAuth))
	shop.Use(middleware.BackendModeUserGuard(settingService))
	{
		shop.GET("/banners", shopHandler.ListBanners)
		shop.GET("/products", shopHandler.ListProducts)
		shop.POST("/orders", shopHandler.CreateOrder)
		shop.GET("/orders/my", shopHandler.MyOrders)
	}
}
