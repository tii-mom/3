package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type SaaSHandler struct {
	service      *service.SaaSService
	totp         *service.TotpService
	distribution *service.DistributionService
}

func NewSaaSHandler(saasService *service.SaaSService, totpService *service.TotpService, distributionService *service.DistributionService) *SaaSHandler {
	return &SaaSHandler{service: saasService, totp: totpService, distribution: distributionService}
}

func (h *SaaSHandler) GetConfig(c *gin.Context) {
	enabled, err := h.service.Enabled(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	applicationsEnabled, err := h.service.ApplicationEnabled(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"enabled": enabled, "applications_enabled": applicationsEnabled})
}

type saasConfigRequest struct {
	Enabled             *bool  `json:"enabled"`
	ApplicationsEnabled *bool  `json:"applications_enabled"`
	TOTPCode            string `json:"totp_code" binding:"required"`
}

func (h *SaaSHandler) UpdateConfig(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var request saasConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.service.UpdateFeatureFlags(c.Request.Context(), request.Enabled, request.ApplicationsEnabled); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	enabled, err := h.service.Enabled(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	applicationsEnabled, err := h.service.ApplicationEnabled(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"enabled": enabled, "applications_enabled": applicationsEnabled})
}

func (h *SaaSHandler) ListApplications(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.AdminListApplications(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

type transitionSaaSApplicationRequest struct {
	Status     string `json:"status" binding:"required"`
	ReviewNote string `json:"review_note"`
	Slug       string `json:"slug"`
	SiteName   string `json:"site_name"`
	SiteLogo   string `json:"site_logo"`
	TOTPCode   string `json:"totp_code" binding:"required"`
}

func (h *SaaSHandler) TransitionApplication(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	applicationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || applicationID <= 0 {
		response.BadRequest(c, "Invalid application id")
		return
	}
	var request transitionSaaSApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if strings.EqualFold(strings.TrimSpace(request.Status), "APPROVED") {
		result, err := h.service.ApproveApplication(c.Request.Context(), applicationID, subject.UserID, service.ApproveSaaSApplicationInput{
			Slug: request.Slug, SiteName: request.SiteName, SiteLogo: request.SiteLogo, ReviewNote: request.ReviewNote,
		})
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, result)
		return
	}
	item, err := h.service.ReviewApplication(c.Request.Context(), applicationID, subject.UserID, request.Status, request.ReviewNote)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

type createTenantRequest struct {
	Slug         string `json:"slug" binding:"required"`
	Name         string `json:"name" binding:"required"`
	SiteName     string `json:"site_name"`
	SiteLogo     string `json:"site_logo"`
	CoreUserID   int64  `json:"core_user_id" binding:"required"`
	ReferralCode string `json:"referral_code"`
	TOTPCode     string `json:"totp_code" binding:"required"`
}

func (h *SaaSHandler) CreateTenant(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var request createTenantRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result, err := h.service.CreateTenant(c.Request.Context(), service.CreateSaaSTenantInput{Slug: request.Slug, Name: request.Name, SiteName: request.SiteName, SiteLogo: request.SiteLogo, CoreUserID: request.CoreUserID, ReferralCode: request.ReferralCode})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, result)
}

func (h *SaaSHandler) ListTenants(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.ListTenants(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

type fundWholesaleRequest struct {
	Amount    string `json:"amount_usd" binding:"required"`
	Reference string `json:"reference" binding:"required"`
	TOTPCode  string `json:"totp_code" binding:"required"`
}

func (h *SaaSHandler) FundWholesale(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || tenantID <= 0 {
		response.BadRequest(c, "Invalid tenant id")
		return
	}
	var request fundWholesaleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	balance, err := h.service.FundWholesaleWallet(c.Request.Context(), tenantID, request.Amount, request.Reference)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"balance_usd": balance})
}

type addDomainRequest struct {
	Domain string `json:"domain" binding:"required"`
}

func (h *SaaSHandler) AddDomain(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || tenantID <= 0 {
		response.BadRequest(c, "Invalid tenant id")
		return
	}
	var request addDomainRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.service.AddDomain(c.Request.Context(), tenantID, request.Domain)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}

func (h *SaaSHandler) VerifyDomain(c *gin.Context) {
	domainID, err := strconv.ParseInt(c.Param("domain_id"), 10, 64)
	if err != nil || domainID <= 0 {
		response.BadRequest(c, "Invalid domain id")
		return
	}
	item, err := h.service.VerifyDomain(c.Request.Context(), domainID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

type createPlanRequest struct {
	Name          string `json:"name" binding:"required"`
	BillingPeriod string `json:"billing_period" binding:"required"`
	PriceMinor    int64  `json:"price_cny_minor"`
	Limits        string `json:"limits"`
}

func (h *SaaSHandler) CreatePlan(c *gin.Context) {
	var request createPlanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.service.CreatePlan(c.Request.Context(), request.Name, request.BillingPeriod, request.PriceMinor, request.Limits)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}

func (h *SaaSHandler) ListPlans(c *gin.Context) {
	items, err := h.service.ListPlans(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *SaaSHandler) ListSubscriptions(c *gin.Context) {
	tenantID, _ := strconv.ParseInt(c.Query("tenant_id"), 10, 64)
	items, err := h.service.ListSubscriptions(c.Request.Context(), tenantID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *SaaSHandler) ListDomains(c *gin.Context) {
	tenantID, _ := strconv.ParseInt(c.Query("tenant_id"), 10, 64)
	items, err := h.service.ListDomains(c.Request.Context(), tenantID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *SaaSHandler) ListResourceAllocations(c *gin.Context) {
	tenantID, _ := strconv.ParseInt(c.Query("tenant_id"), 10, 64)
	items, err := h.service.ListResourceAllocations(c.Request.Context(), tenantID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *SaaSHandler) ListProvisioningJobs(c *gin.Context) {
	tenantID, _ := strconv.ParseInt(c.Query("tenant_id"), 10, 64)
	items, err := h.service.ListProvisioningJobs(c.Request.Context(), tenantID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

type resourceAllocationRequest struct {
	GroupID          int64  `json:"group_id" binding:"required"`
	AllocationType   string `json:"allocation_type" binding:"required"`
	ConcurrencyLimit int    `json:"concurrency_limit"`
	MonthlyLimitUSD  string `json:"monthly_limit_usd"`
	TOTPCode         string `json:"totp_code" binding:"required"`
}

func (h *SaaSHandler) AssignResourcePool(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || tenantID <= 0 {
		response.BadRequest(c, "Invalid tenant id")
		return
	}
	var request resourceAllocationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.service.AssignResourcePool(c.Request.Context(), tenantID, request.GroupID, request.AllocationType, request.ConcurrencyLimit, request.MonthlyLimitUSD); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}

type paidSubscriptionRequest struct {
	TenantID  int64  `json:"tenant_id" binding:"required"`
	PlanID    int64  `json:"plan_id" binding:"required"`
	PaidMinor int64  `json:"paid_cny_minor" binding:"required"`
	Reference string `json:"reference" binding:"required"`
	TOTPCode  string `json:"totp_code" binding:"required"`
}

func (h *SaaSHandler) RecordPaidSubscription(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var request paidSubscriptionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	id, err := h.service.RecordPaidSubscription(c.Request.Context(), request.TenantID, request.PlanID, request.PaidMinor, request.Reference)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, gin.H{"subscription_id": id})
}

func (h *SaaSHandler) ListPartnerWithdrawals(c *gin.Context) {
	items, err := h.service.AdminListPartnerWithdrawals(c.Request.Context(), c.Query("status"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *SaaSHandler) TransitionPartnerWithdrawal(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
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
	item, err := h.service.AdminTransitionPartnerWithdrawal(c.Request.Context(), id, subject.UserID, request.Status, request.Reason, request.PaymentReference, request.ProofURL)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

func (h *SaaSHandler) PartnerPayoutDetails(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
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
	details, err := h.distribution.AdminPartnerPayoutDetails(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, details)
}
