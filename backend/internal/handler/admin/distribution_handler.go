package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type DistributionHandler struct {
	service *service.DistributionService
	totp    *service.TotpService
}

func NewDistributionHandler(distributionService *service.DistributionService, totpService *service.TotpService) *DistributionHandler {
	return &DistributionHandler{service: distributionService, totp: totpService}
}

func (h *DistributionHandler) ListWithdrawals(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.AdminListWithdrawals(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) ListCommissions(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.AdminListCommissions(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) ListRechargeEvents(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.AdminListRechargeEvents(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *DistributionHandler) ListRelations(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.AdminListRelations(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

type distributionReversalRequest struct {
	ReversalType string `json:"reversal_type" binding:"required"`
	Reason       string `json:"reason" binding:"required"`
	TOTPCode     string `json:"totp_code" binding:"required"`
}

func (h *DistributionHandler) ReverseRecharge(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || eventID <= 0 {
		response.BadRequest(c, "Invalid recharge event id")
		return
	}
	var request distributionReversalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	item, err := h.service.ReverseRecharge(c.Request.Context(), eventID, subject.UserID, request.ReversalType, request.Reason)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *DistributionHandler) GetConfig(c *gin.Context) {
	config, err := h.service.ProgramConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, config)
}

func (h *DistributionHandler) GetFinancialRuntimeConfig(c *gin.Context) {
	config, err := h.service.FinancialRuntimeConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, config)
}

type distributionConfigRequest struct {
	Enabled         bool   `json:"enabled"`
	StackWithLegacy bool   `json:"stack_with_legacy"`
	TOTPCode        string `json:"totp_code" binding:"required"`
}

type financialRuntimeConfigRequest struct {
	CreditBucketEnforceEnabled bool   `json:"credit_bucket_enforce_enabled"`
	TOTPCode                   string `json:"totp_code" binding:"required"`
}

func (h *DistributionHandler) UpdateFinancialRuntimeConfig(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var request financialRuntimeConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.service.UpdateFinancialRuntimeConfig(c.Request.Context(), request.CreditBucketEnforceEnabled); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"credit_bucket_enforce_enabled": request.CreditBucketEnforceEnabled})
}

type distributionPolicyRequest struct {
	service.DistributionPolicyInput
	TOTPCode string `json:"totp_code" binding:"required"`
}

func (h *DistributionHandler) CreatePolicyVersion(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var request distributionPolicyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	version, err := h.service.CreatePolicyVersion(c.Request.Context(), subject.UserID, request.DistributionPolicyInput)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, gin.H{"config_version": version})
}

func (h *DistributionHandler) UpdateConfig(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var request distributionConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.service.UpdateProgramConfig(c.Request.Context(), request.Enabled, request.StackWithLegacy); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"enabled": request.Enabled, "stack_with_legacy": request.StackWithLegacy})
}

type withdrawalTransitionRequest struct {
	Status           string `json:"status" binding:"required"`
	Reason           string `json:"reason"`
	PaymentReference string `json:"payment_reference"`
	ProofURL         string `json:"proof_url"`
	TOTPCode         string `json:"totp_code" binding:"required"`
}

func (h *DistributionHandler) TransitionWithdrawal(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	withdrawalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || withdrawalID <= 0 {
		response.BadRequest(c, "Invalid withdrawal id")
		return
	}
	var request withdrawalTransitionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	item, err := h.service.AdminTransitionWithdrawal(c.Request.Context(), withdrawalID, subject.UserID, request.Status, request.Reason, request.PaymentReference, request.ProofURL)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

type payoutDetailsRequest struct {
	TOTPCode string `json:"totp_code" binding:"required"`
}

func (h *DistributionHandler) PayoutDetails(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	withdrawalID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || withdrawalID <= 0 {
		response.BadRequest(c, "Invalid withdrawal id")
		return
	}
	var request payoutDetailsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	details, err := h.service.AdminPayoutDetails(c.Request.Context(), withdrawalID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, details)
}
