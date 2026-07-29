package admin

import (
	"strconv"

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

type shopProductRequest struct {
	Name                  string `json:"name"`
	Description           string `json:"description"`
	ImageURL              string `json:"image_url"`
	ProductType           string `json:"product_type"`
	PriceCNYMinor         int64  `json:"price_cny_minor"`
	OriginalPriceCNYMinor int64  `json:"original_price_cny_minor"`
	GrantUSDAmount        string `json:"grant_usd_amount"`
	StockQuantity         *int64 `json:"stock_quantity"`
	CommissionBPS         int    `json:"commission_bps"`
	Status                string `json:"status"`
	SortOrder             int    `json:"sort_order"`
}

type shopBannerRequest struct {
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	ImageURL   string `json:"image_url"`
	ButtonText string `json:"button_text"`
	ProductID  *int64 `json:"product_id"`
	Enabled    bool   `json:"enabled"`
	SortOrder  int    `json:"sort_order"`
}

type shopFulfillOrderRequest struct {
	FulfillmentNote string `json:"fulfillment_note"`
}

func (h *ShopHandler) ListProducts(c *gin.Context) {
	items, err := h.shopService.AdminListProducts(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *ShopHandler) CreateProduct(c *gin.Context) {
	var req shopProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.shopService.CreateProduct(c.Request.Context(), req.toServiceInput())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *ShopHandler) UpdateProduct(c *gin.Context) {
	id, ok := parseAdminShopIDParam(c)
	if !ok {
		return
	}
	var req shopProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.shopService.UpdateProduct(c.Request.Context(), id, req.toServiceInput())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *ShopHandler) DeleteProduct(c *gin.Context) {
	id, ok := parseAdminShopIDParam(c)
	if !ok {
		return
	}
	if err := h.shopService.DeleteProduct(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "product archived"})
}

func (h *ShopHandler) ListBanners(c *gin.Context) {
	items, err := h.shopService.AdminListBanners(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *ShopHandler) CreateBanner(c *gin.Context) {
	var req shopBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.shopService.CreateBanner(c.Request.Context(), req.toServiceInput())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *ShopHandler) UpdateBanner(c *gin.Context) {
	id, ok := parseAdminShopIDParam(c)
	if !ok {
		return
	}
	var req shopBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.shopService.UpdateBanner(c.Request.Context(), id, req.toServiceInput())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *ShopHandler) DeleteBanner(c *gin.Context) {
	id, ok := parseAdminShopIDParam(c)
	if !ok {
		return
	}
	if err := h.shopService.DeleteBanner(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "banner removed"})
}

func (h *ShopHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.shopService.AdminListOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *ShopHandler) FulfillOrder(c *gin.Context) {
	id, ok := parseAdminShopIDParam(c)
	if !ok {
		return
	}
	var req shopFulfillOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.shopService.AdminFulfillOrder(c.Request.Context(), id, req.FulfillmentNote); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "order fulfilled"})
}

func parseAdminShopIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid ID")
		return 0, false
	}
	return id, true
}

func (r shopProductRequest) toServiceInput() service.UpsertShopProductInput {
	return service.UpsertShopProductInput{
		Name:                  r.Name,
		Description:           r.Description,
		ImageURL:              r.ImageURL,
		ProductType:           r.ProductType,
		PriceCNYMinor:         r.PriceCNYMinor,
		OriginalPriceCNYMinor: r.OriginalPriceCNYMinor,
		GrantUSDAmount:        r.GrantUSDAmount,
		StockQuantity:         r.StockQuantity,
		CommissionBPS:         r.CommissionBPS,
		Status:                r.Status,
		SortOrder:             r.SortOrder,
	}
}

func (r shopBannerRequest) toServiceInput() service.UpsertShopBannerInput {
	return service.UpsertShopBannerInput{
		Title:      r.Title,
		Subtitle:   r.Subtitle,
		ImageURL:   r.ImageURL,
		ButtonText: r.ButtonText,
		ProductID:  r.ProductID,
		Enabled:    r.Enabled,
		SortOrder:  r.SortOrder,
	}
}
