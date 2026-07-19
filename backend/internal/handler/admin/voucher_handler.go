package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type VoucherHandler struct {
	service *service.VoucherService
	totp    *service.TotpService
}

func NewVoucherHandler(voucherService *service.VoucherService, totpService *service.TotpService) *VoucherHandler {
	return &VoucherHandler{service: voucherService, totp: totpService}
}

func (h *VoucherHandler) GetConfig(c *gin.Context) {
	config, err := h.service.AdminConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, config)
}

type voucherConfigRequest struct {
	Enabled  bool   `json:"enabled"`
	TOTPCode string `json:"totp_code" binding:"required"`
}

func (h *VoucherHandler) UpdateConfig(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Admin not authenticated")
		return
	}
	var request voucherConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.totp.VerifyCode(c.Request.Context(), subject.UserID, request.TOTPCode); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := h.service.UpdateEnabled(c.Request.Context(), request.Enabled); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"enabled": request.Enabled})
}

func (h *VoucherHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.AdminList(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

type voucherRiskRequest struct {
	Locked bool   `json:"locked"`
	Reason string `json:"reason"`
}

func (h *VoucherHandler) SetRiskLock(c *gin.Context) {
	voucherID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || voucherID <= 0 {
		response.BadRequest(c, "Invalid voucher id")
		return
	}
	var request voucherRiskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	item, err := h.service.SetRiskLock(c.Request.Context(), voucherID, request.Locked, request.Reason)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
