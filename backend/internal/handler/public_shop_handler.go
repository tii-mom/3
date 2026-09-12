package handler

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// PublicShopHandler 承载官网首页的免登录直充链路。
// 与需要登录的 ShopHandler 分离，避免公开路由意外获得用户态权限。
type PublicShopHandler struct {
	shopService   *service.ShopService
	configService *service.PaymentConfigService
}

func NewPublicShopHandler(
	shopService *service.ShopService,
	configService *service.PaymentConfigService,
) *PublicShopHandler {
	return &PublicShopHandler{shopService: shopService, configService: configService}
}

func (h *PublicShopHandler) ListProducts(c *gin.Context) {
	items, err := h.shopService.ListPublicProducts(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

// ListCategories 官网首页的分组导航数据源（只返回后台启用中的品类）。
// 首页分组、标签、顺序全部以后台商城配置为准，前端不再硬编码品类枚举。
func (h *PublicShopHandler) ListCategories(c *gin.Context) {
	items, err := h.shopService.ListPublicCategories(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

// ListPaymentMethods 公开支付渠道与金额区间，供官网首页免登录下单使用。
// 与登录态 /payment/checkout-info 同源，但只返回渠道名称与限额，不暴露价格档位等其他配置。
func (h *PublicShopHandler) ListPaymentMethods(c *gin.Context) {
	methods := map[string]service.MethodLimits{}
	var globalMin, globalMax float64
	forceQRCode := false

	if h.configService != nil {
		limits, err := h.configService.GetAvailableMethodLimits(c.Request.Context())
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if limits.Methods != nil {
			methods = limits.Methods
		}
		globalMin, globalMax = limits.GlobalMin, limits.GlobalMax
		if cfg, cfgErr := h.configService.GetPaymentConfig(c.Request.Context()); cfgErr == nil {
			forceQRCode = cfg.AlipayForceQRCode
		}
	}

	response.Success(c, gin.H{
		"methods":             methods,
		"global_min":          globalMin,
		"global_max":          globalMax,
		"alipay_force_qrcode": forceQRCode,
	})
}

// ListBanners 官网首页轮播，与控制台商城共用同一份 banner 数据。
func (h *PublicShopHandler) ListBanners(c *gin.Context) {
	items, err := h.shopService.ListPublicBanners(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

type createPublicShopOrderRequest struct {
	ProductID   int64  `json:"product_id" binding:"required"`
	PaymentType string `json:"payment_type" binding:"required"`
	Contact     string `json:"contact" binding:"required"`
	ReturnURL   string `json:"return_url"`
	OpenID      string `json:"openid"`
	IsMobile    *bool  `json:"is_mobile,omitempty"`
}

func (h *PublicShopHandler) CreateOrder(c *gin.Context) {
	var req createPublicShopOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	mobile := isMobile(c)
	if req.IsMobile != nil {
		mobile = *req.IsMobile
	}
	result, err := h.shopService.CreateGuestOrderAndPayment(
		c.Request.Context(),
		req.Contact,
		req.ProductID,
		strings.TrimSpace(req.PaymentType),
		req.ReturnURL,
		c.ClientIP(),
		c.Request.Host,
		c.Request.Referer(),
		c.GetHeader("Accept-Language"),
		mobile,
		isWeChatBrowser(c),
		strings.TrimSpace(req.OpenID),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

type submitPublicDeliveryRequest struct {
	OrderNo        string `json:"order_no" binding:"required"`
	Contact        string `json:"contact" binding:"required"`
	Payload        string `json:"payload" binding:"required"`
	RentalDuration string `json:"rental_duration"`
}

func (h *PublicShopHandler) SubmitDelivery(c *gin.Context) {
	var req submitPublicDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.shopService.SubmitGuestDeliveryInfo(c.Request.Context(), req.OrderNo, req.Contact, req.Payload, req.RentalDuration); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"submitted": true})
}

func (h *PublicShopHandler) LookupOrder(c *gin.Context) {
	orderNo := strings.TrimSpace(c.Query("order_no"))
	contact := strings.TrimSpace(c.Query("contact"))
	if orderNo == "" || contact == "" {
		response.BadRequest(c, "order_no and contact are required")
		return
	}
	item, err := h.shopService.LookupGuestOrder(c.Request.Context(), orderNo, contact)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *AdminShopDeliveryHandler) Decrypt(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "invalid order id")
		return
	}
	plain, err := h.shopService.AdminDecryptDelivery(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"payload": plain})
}

type AdminShopDeliveryHandler struct {
	shopService *service.ShopService
}

func NewAdminShopDeliveryHandler(shopService *service.ShopService) *AdminShopDeliveryHandler {
	return &AdminShopDeliveryHandler{shopService: shopService}
}
