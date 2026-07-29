package handler

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PetInternalHandler handles server-to-server requests from TAI Protocol backend.
type PetInternalHandler struct {
	petService *service.PetAccountService
}

func NewPetInternalHandler(petService *service.PetAccountService) *PetInternalHandler {
	return &PetInternalHandler{petService: petService}
}

// ─── Request/Response DTOs ─────────────────────────────────────────

type ProvisionRequest struct {
	PetID     string `json:"pet_id" binding:"required,max=64"`
	OwnerTgID string `json:"owner_tg_id" binding:"required,max=64"`
	PetName   string `json:"pet_name" binding:"max=100"`
}

type ProvisionResponse struct {
	PetID     string `json:"pet_id"`
	UserID    int64  `json:"user_id"`
	APIKey    string `json:"api_key"`
	GroupID   int64  `json:"group_id"`
	Status    string `json:"status"`
}

type CreditRequest struct {
	PetID       string  `json:"pet_id" binding:"required"`
	TAIAmount   float64 `json:"tai_amount" binding:"required,gt=0"`
	CreditAmount float64 `json:"credit_amount" binding:"required,gt=0"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

type CreditResponse struct {
	PetID          string  `json:"pet_id"`
	Credited       float64 `json:"credited"`
	NewBalance     float64 `json:"new_balance"`
	TAISpentTotal  float64 `json:"tai_spent_total"`
}

type PetStatusResponse struct {
	PetID             string  `json:"pet_id"`
	Status            string  `json:"status"`
	Balance           float64 `json:"balance"`
	TAISpentTotal     float64 `json:"tai_spent_total"`
	DailyTAIUsed      float64 `json:"daily_tai_used"`
	DailyTALimit      float64 `json:"daily_tai_limit"`
	APIKeyStatus      string  `json:"api_key_status"`
}

type BatchUsageRequest struct {
	PetIDs []string `json:"pet_ids" binding:"required,min=1,max=100"`
}

type PetUsageItem struct {
	PetID        string  `json:"pet_id"`
	Balance      float64 `json:"balance"`
	DailyTAIUsed float64 `json:"daily_tai_used"`
	Status       string  `json:"status"`
}

// ─── Handlers ──────────────────────────────────────────────────────

// Provision creates a new 3api user + API key for a TAI pet.
func (h *PetInternalHandler) Provision(c *gin.Context) {
	var req ProvisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.petService.Provision(c.Request.Context(), req.PetID, req.OwnerTgID, req.PetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Credit adds compute balance to a pet's account (TAI → 3api balance conversion).
func (h *PetInternalHandler) Credit(c *gin.Context) {
	var req CreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.petService.Credit(c.Request.Context(), req.PetID, req.TAIAmount, req.CreditAmount, req.IdempotencyKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetStatus returns a pet's account status, balance, and usage.
func (h *PetInternalHandler) GetStatus(c *gin.Context) {
	petID := c.Param("pet_id")
	if petID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pet_id required"})
		return
	}

	resp, err := h.petService.GetStatus(c.Request.Context(), petID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Suspend disables a pet's compute access.
func (h *PetInternalHandler) Suspend(c *gin.Context) {
	petID := c.Param("pet_id")
	if err := h.petService.SetStatus(c.Request.Context(), petID, "suspended"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pet_id": petID, "status": "suspended"})
}

// Reactivate re-enables a suspended pet.
func (h *PetInternalHandler) Reactivate(c *gin.Context) {
	petID := c.Param("pet_id")
	if err := h.petService.SetStatus(c.Request.Context(), petID, "active"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pet_id": petID, "status": "active"})
}

// BatchUsage returns usage info for multiple pets in one call.
func (h *PetInternalHandler) BatchUsage(c *gin.Context) {
	var req BatchUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, err := h.petService.BatchUsage(c.Request.Context(), req.PetIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pets": items})
}
