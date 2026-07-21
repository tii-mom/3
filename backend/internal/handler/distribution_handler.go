package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type DistributionHandler struct {
	service *service.DistributionService
}

func NewDistributionHandler(distributionService *service.DistributionService) *DistributionHandler {
	return &DistributionHandler{service: distributionService}
}

func (h *DistributionHandler) Dashboard(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	result, err := h.service.Dashboard(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *DistributionHandler) Tree(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	parentID, _ := strconv.ParseInt(c.Query("parent_user_id"), 10, 64)
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.Tree(c.Request.Context(), userID, parentID, c.Query("search"), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) Ledger(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.Ledger(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

type payoutAccountRequest struct {
	AlipayAccount string `json:"alipay_account" binding:"required"`
	RealName      string `json:"real_name" binding:"required"`
}

type conversionRequest struct {
	AmountCNYMinor int64  `json:"amount_cny_minor" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *DistributionHandler) GetPayoutAccount(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	account, err := h.service.GetPayoutAccount(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, account)
}

func (h *DistributionHandler) SavePayoutAccount(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var request payoutAccountRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	account, err := h.service.SavePayoutAccount(c.Request.Context(), userID, request.AlipayAccount, request.RealName)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, account)
}

type withdrawalRequest struct {
	AmountMinor int64 `json:"amount_cny_minor" binding:"required"`
}

func (h *DistributionHandler) ListWithdrawals(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.ListWithdrawals(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) CreateWithdrawal(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var request withdrawalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.service.CreateWithdrawal(c.Request.Context(), userID, request.AmountMinor)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}

func (h *DistributionHandler) ConvertToPlatformBalance(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var request conversionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.service.ConvertToPlatformBalance(c.Request.Context(), userID, request.AmountCNYMinor, request.IdempotencyKey)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}

func authenticatedUserID(c *gin.Context) (int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return 0, false
	}
	return subject.UserID, true
}
