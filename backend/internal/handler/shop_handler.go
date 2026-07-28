package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ShopHandler struct {
	shopService *service.ShopService
}

func NewShopHandler(shopService *service.ShopService) *ShopHandler {
	return &ShopHandler{shopService: shopService}
}

func (h *ShopHandler) ListBanners(c *gin.Context) {
	items, err := h.shopService.ListPublicBanners(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *ShopHandler) ListProducts(c *gin.Context) {
	items, err := h.shopService.ListPublicProducts(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

type createShopOrderRequest struct {
	ProductID   int64  `json:"product_id" binding:"required"`
	PaymentType string `json:"payment_type" binding:"required"`
	ReturnURL   string `json:"return_url"`
	OpenID      string `json:"openid"`
	IsMobile    *bool  `json:"is_mobile,omitempty"`
}

func (h *ShopHandler) CreateOrder(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	var req createShopOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	mobile := isMobile(c)
	if req.IsMobile != nil {
		mobile = *req.IsMobile
	}
	result, err := h.shopService.CreateOrderAndPayment(
		c.Request.Context(),
		subject.UserID,
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

func (h *ShopHandler) MyOrders(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.shopService.ListMyOrders(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}
