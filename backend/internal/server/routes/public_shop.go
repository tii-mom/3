package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterPublicShopRoutes 官网首页免登录直充链路。
// 刻意不挂载 JWT 中间件：下单只需联系方式，支付后凭订单号 + 联系方式查单。
// 防刷依赖 service 层对 IP 与联系方式的限流。
func RegisterPublicShopRoutes(v1 *gin.RouterGroup, publicShopHandler *handler.PublicShopHandler) {
	public := v1.Group("/public/shop")
	{
		public.GET("/categories", publicShopHandler.ListCategories)
		public.GET("/products", publicShopHandler.ListProducts)
		public.GET("/banners", publicShopHandler.ListBanners)
		public.GET("/payment-methods", publicShopHandler.ListPaymentMethods)
		public.POST("/orders", publicShopHandler.CreateOrder)
		public.POST("/orders/delivery", publicShopHandler.SubmitDelivery)
		public.GET("/orders/lookup", publicShopHandler.LookupOrder)
	}
}

// RegisterAdminShopDeliveryRoutes 后台解密交付资料，与其他 admin 路由同样鉴权与审计。
func RegisterAdminShopDeliveryRoutes(
	v1 *gin.RouterGroup,
	deliveryHandler *handler.AdminShopDeliveryHandler,
	adminAuth middleware.AdminAuthMiddleware,
	auditLog middleware.AuditLogMiddleware,
	settingService *service.SettingService,
) {
	group := v1.Group("/admin/shop")
	group.Use(gin.HandlerFunc(adminAuth))
	group.Use(gin.HandlerFunc(auditLog))
	group.Use(middleware.AdminComplianceGuard(settingService))
	group.GET("/orders/:id/delivery", deliveryHandler.Decrypt)
}
