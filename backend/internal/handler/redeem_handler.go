package handler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RedeemHandler handles redeem code-related requests
type RedeemHandler struct {
	redeemService  *service.RedeemService
	voucherService *service.VoucherService
}

// NewRedeemHandler creates a new RedeemHandler
func NewRedeemHandler(redeemService *service.RedeemService, voucherService *service.VoucherService) *RedeemHandler {
	return &RedeemHandler{
		redeemService:  redeemService,
		voucherService: voucherService,
	}
}

// RedeemRequest represents the redeem code request payload
type RedeemRequest struct {
	Code string `json:"code" binding:"required"`
}

// RedeemResponse represents the redeem response
type RedeemResponse struct {
	Message        string   `json:"message"`
	Type           string   `json:"type"`
	Value          float64  `json:"value"`
	NewBalance     *float64 `json:"new_balance,omitempty"`
	NewConcurrency *int     `json:"new_concurrency,omitempty"`
}

// Redeem handles redeeming a code
// POST /api/v1/redeem
func (h *RedeemHandler) Redeem(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(req.Code)), "VCH-") {
		result, err := h.voucherService.Redeem(c.Request.Context(), subject.UserID, req.Code)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		mapped, mapErr := mapVoucherRedeemResponse(result)
		if mapErr != nil {
			response.ErrorFrom(c, mapErr)
			return
		}
		response.Success(c, mapped)
		return
	}

	result, err := h.redeemService.Redeem(c.Request.Context(), subject.UserID, req.Code)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, RedeemResponse{Message: "success", Type: result.Type, Value: result.Value})
}

func mapVoucherRedeemResponse(result *service.VoucherRedeemResult) (RedeemResponse, error) {
	if result == nil {
		return RedeemResponse{}, fmt.Errorf("voucher redemption returned no result")
	}
	value, err := strconv.ParseFloat(result.FaceValue, 64)
	if err != nil {
		return RedeemResponse{}, fmt.Errorf("parse voucher face value: %w", err)
	}
	newBalance, err := strconv.ParseFloat(result.NewBalance, 64)
	if err != nil {
		return RedeemResponse{}, fmt.Errorf("parse voucher balance: %w", err)
	}
	return RedeemResponse{Message: "success", Type: service.RedeemTypeBalance, Value: value, NewBalance: &newBalance}, nil
}

// GetHistory returns the user's redemption history
// GET /api/v1/redeem/history
func (h *RedeemHandler) GetHistory(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Default limit is 25
	limit := 25

	codes, err := h.redeemService.GetUserHistory(c.Request.Context(), subject.UserID, limit)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	allCodes := append([]service.RedeemCode(nil), codes...)
	if h.voucherService != nil {
		vouchers, voucherErr := h.voucherService.ListRedeemedHistory(c.Request.Context(), subject.UserID, int64(limit))
		if voucherErr != nil {
			response.ErrorFrom(c, voucherErr)
			return
		}
		for i := range vouchers {
			value, parseErr := strconv.ParseFloat(vouchers[i].FaceValue, 64)
			if parseErr != nil {
				response.ErrorFrom(c, fmt.Errorf("parse voucher history value: %w", parseErr))
				return
			}
			allCodes = append(allCodes, service.RedeemCode{
				ID:        -vouchers[i].ID,
				Code:      "VCH-****-" + vouchers[i].CodeLast4,
				Type:      service.RedeemTypeBalance,
				Value:     value,
				Status:    service.StatusUsed,
				UsedBy:    vouchers[i].RedeemerUserID,
				UsedAt:    vouchers[i].RedeemedAt,
				CreatedAt: vouchers[i].CreatedAt,
			})
		}
	}
	sort.SliceStable(allCodes, func(i, j int) bool {
		return redeemHistoryTime(allCodes[i]).After(redeemHistoryTime(allCodes[j]))
	})
	if len(allCodes) > limit {
		allCodes = allCodes[:limit]
	}

	out := make([]dto.RedeemCode, 0, len(allCodes))
	for i := range allCodes {
		out = append(out, *dto.RedeemCodeFromService(&allCodes[i]))
	}
	response.Success(c, out)
}

func redeemHistoryTime(code service.RedeemCode) time.Time {
	if code.UsedAt != nil {
		return *code.UsedAt
	}
	return code.CreatedAt
}
