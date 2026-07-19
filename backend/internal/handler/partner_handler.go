package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type PartnerHandler struct{ service *service.SaaSService }

func NewPartnerHandler(saasService *service.SaaSService) *PartnerHandler {
	return &PartnerHandler{service: saasService}
}

func (h *PartnerHandler) Dashboard(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	item, err := h.service.PartnerDashboard(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
func (h *PartnerHandler) ListWithdrawals(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	items, err := h.service.ListPartnerWithdrawals(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}
func (h *PartnerHandler) CreateWithdrawal(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var request withdrawalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.service.CreatePartnerWithdrawal(c.Request.Context(), userID, request.AmountMinor)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}

func (h *PartnerHandler) TenantControl(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	item, err := h.service.TenantControl(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

type tenantControlRequest struct {
	SiteName         string `json:"site_name" binding:"required"`
	SiteLogo         string `json:"site_logo"`
	RetailMultiplier string `json:"retail_multiplier" binding:"required"`
	PaymentProvider  string `json:"payment_provider"`
	PaymentConfig    string `json:"payment_config"`
	InstanceConfig   string `json:"instance_config"`
}

func (h *PartnerHandler) UpdateTenantControl(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var request tenantControlRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.service.UpdateTenantControl(c.Request.Context(), userID, request.SiteName, request.SiteLogo, request.RetailMultiplier, request.PaymentProvider, request.PaymentConfig, request.InstanceConfig)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}

type tenantDomainRequest struct {
	Domain string `json:"domain" binding:"required"`
}

func (h *PartnerHandler) AddTenantDomain(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	control, err := h.service.TenantControl(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	var request tenantDomainRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.service.AddDomain(c.Request.Context(), control.Tenant.ID, request.Domain)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, item)
}
