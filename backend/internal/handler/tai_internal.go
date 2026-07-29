package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TAIInternalHandler handles minimal server-to-server requests from TAI Protocol.
// Phase 0: platform pool account model — no per-pet provisioning needed.
type TAIInternalHandler struct {
	// platformUserID is the 3api user ID for the TAI platform pool account.
	platformUserID int64
	// db access would go here for balance queries
}

func NewTAIInternalHandler(platformUserID int64) *TAIInternalHandler {
	return &TAIInternalHandler{platformUserID: platformUserID}
}

// ─── DTOs ──────────────────────────────────────────────────────────

type UsageReportRequest struct {
	Period       string  `json:"period"`        // e.g. "2026-07-29"
	TotalCalls   int64   `json:"total_calls"`
	TotalTokens  int64   `json:"total_tokens"`
	TotalTAISpent float64 `json:"total_tai_spent"`
	ActivePets   int     `json:"active_pets"`
}

type AccountStatusResponse struct {
	PlatformUserID int64   `json:"platform_user_id"`
	Balance        float64 `json:"balance"`
	GroupName      string  `json:"group_name"`
	Status         string  `json:"status"`
}

// ─── Handlers ──────────────────────────────────────────────────────

// UsageReport receives aggregated usage data from TAI backend for reconciliation.
func (h *TAIInternalHandler) UsageReport(c *gin.Context) {
	var req UsageReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Store reconciliation record for finance audit
	// For Phase 0 this is informational only — billing is via the platform key's
	// normal usage deduction in the gateway.

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"period":  req.Period,
		"message": "usage report received",
	})
}

// AccountStatus returns the platform pool account's current state.
func (h *TAIInternalHandler) AccountStatus(c *gin.Context) {
	// TODO: Query user balance from DB
	c.JSON(http.StatusOK, AccountStatusResponse{
		PlatformUserID: h.platformUserID,
		Balance:        0, // TODO: actual query
		GroupName:      "tai-pets",
		Status:         "active",
	})
}

// BalanceCheck returns whether the platform account has sufficient balance.
// TAI backend calls this periodically to decide whether to pause pet operations.
func (h *TAIInternalHandler) BalanceCheck(c *gin.Context) {
	// TODO: Query actual balance
	balance := 0.0
	threshold := 10.0 // minimum USD balance before alerting

	c.JSON(http.StatusOK, gin.H{
		"balance":        balance,
		"sufficient":     balance > threshold,
		"threshold":      threshold,
		"recommend_topup": balance < threshold*2,
	})
}
