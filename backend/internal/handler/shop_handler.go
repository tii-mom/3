package handler

import (
	"strconv"
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
	ProductID int64 `json:"product_id" binding:"required"`
	// PaymentType 在「全额返点余额抵扣」时可以为空（没有外部支付环节），
	// 是否必填由 service 按应付金额判断，避免把无支付单的订单挡在绑定校验外。
	PaymentType string `json:"payment_type"`
	ReturnURL   string `json:"return_url"`
	OpenID      string `json:"openid"`
	IsMobile    *bool  `json:"is_mobile,omitempty"`
	// UseWallet 为 true 时先扣人民币返点余额，差额再走微信/支付宝；
	// 余额足够全额抵扣时不产生外部支付单（结果里 fully_paid_by_wallet = true）。
	UseWallet bool `json:"use_wallet"`
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
		req.UseWallet,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

// WalletBalance 返回当前用户可用于商城抵扣的人民币返点余额。
// 与「API 美金额度」是两个完全独立的资金账户，前端下单页用它展示可抵扣上限。
func (h *ShopHandler) WalletBalance(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	balance, err := h.shopService.WalletBalance(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, balance)
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

type submitShopDeliveryRequest struct {
	Payload        string `json:"payload" binding:"required"`
	RentalDuration string `json:"rental_duration"`
}

// SubmitDelivery 登录用户为自己的订单补交交付资料（Session / 收货邮箱 / 租期）。
// 与首页免登录链路共用同一套加密存储，后台发货页看到的内容一致。
func (h *ShopHandler) SubmitDelivery(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "invalid order id")
		return
	}
	var req submitShopDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.shopService.SubmitUserDeliveryInfo(c.Request.Context(), subject.UserID, id, req.Payload, req.RentalDuration); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"submitted": true})
}

type claimShopOrderRequest struct {
	OrderNo string `json:"order_no" binding:"required"`
	Contact string `json:"contact" binding:"required"`
}

// ClaimOrder 认领此前以免登录方式在首页下的订单，归入当前账号的订单列表。
func (h *ShopHandler) ClaimOrder(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	var req claimShopOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.shopService.ClaimGuestOrder(c.Request.Context(), subject.UserID, req.OrderNo, req.Contact); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"claimed": true})
}
