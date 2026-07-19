package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type VoucherHandler struct {
	service *service.VoucherService
}

func NewVoucherHandler(voucherService *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{service: voucherService}
}

type createVoucherRequest struct {
	Amount   string `json:"amount" binding:"required"`
	TOTPCode string `json:"totp_code"`
}

func (h *VoucherHandler) Create(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var request createVoucherRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	voucher, err := h.service.Create(c.Request.Context(), subject.UserID, service.CreateVoucherInput{Amount: request.Amount, TOTPCode: request.TOTPCode})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, voucher)
}

func (h *VoucherHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.List(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *VoucherHandler) Availability(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	availability, err := h.service.Availability(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, availability)
}

func (h *VoucherHandler) Cancel(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	voucherID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || voucherID <= 0 {
		response.BadRequest(c, "Invalid voucher id")
		return
	}
	voucher, err := h.service.Cancel(c.Request.Context(), subject.UserID, voucherID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, voucher)
}
