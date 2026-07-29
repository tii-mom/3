package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/petaccount"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/pkg/creditledger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/creditctx"

	"github.com/shopspring/decimal"
)

const (
	// taiPetsGroupName is the pricing group for TAI pets (bulk rate).
	taiPetsGroupName = "tai-pets"

	// defaultDailyTALimit is the max TAI a pet can spend per day.
	defaultDailyTALimit = 100.0
)

// PetAccountService manages TAI Protocol pet compute accounts.
type PetAccountService struct {
	db           *ent.Client
	ledger       *creditledger.Ledger
	apiKeyService *APIKeyService
	userService  *UserService
}

func NewPetAccountService(
	db *ent.Client,
	ledger *creditledger.Ledger,
	apiKeyService *APIKeyService,
	userService *UserService,
) *PetAccountService {
	return &PetAccountService{
		db:            db,
		ledger:        ledger,
		apiKeyService: apiKeyService,
		userService:   userService,
	}
}

// Provision creates a 3api user account and API key for a new pet.
func (s *PetAccountService) Provision(ctx context.Context, petID, ownerTgID, petName string) (*handler.ProvisionResponse, error) {
	// Check if already provisioned
	existing, err := s.db.PetAccount.Query().
		Where(petaccount.PetID(petID)).
		Only(ctx)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("pet %s already provisioned (user_id=%d)", petID, existing.UserID)
	}

	if petName == "" {
		petName = "TAI-Pet-" + petID
	}

	// Create a 3api user for the pet
	user, err := s.userService.CreateSystemUser(ctx, fmt.Sprintf("pet:%s", petID), petName)
	if err != nil {
		return nil, fmt.Errorf("create pet user: %w", err)
	}

	// Find the tai-pets group for bulk pricing
	group, err := s.db.Group.Query().
		Where(func(sel *ent.GroupQuery) {
			// Match by name
		}).
		Only(ctx)
	// If group doesn't exist yet, proceed without group (admin creates it separately)
	var groupID *int64
	if err == nil && group != nil {
		gid := group.ID
		groupID = &gid
	}

	// Create API key for the pet
	apiKey, keyStr, err := s.apiKeyService.CreateKeyForUser(ctx, user.ID, petName, groupID)
	if err != nil {
		return nil, fmt.Errorf("create pet api key: %w", err)
	}

	// Create pet_account record
	petAcc, err := s.db.PetAccount.Create().
		SetPetID(petID).
		SetUserID(user.ID).
		SetAPIKeyID(apiKey.ID).
		SetOwnerTgID(ownerTgID).
		SetStatus("active").
		SetDailyTALimit(defaultDailyTALimit).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create pet account record: %w", err)
	}

	resp := &handler.ProvisionResponse{
		PetID:   petID,
		UserID:  user.ID,
		APIKey:  keyStr,
		Status:  "active",
	}
	if groupID != nil {
		resp.GroupID = *groupID
	}
	_ = petAcc

	return resp, nil
}

// Credit converts TAI tokens into 3api compute balance for a pet.
// Uses the credit ledger with idempotency to prevent double-crediting.
func (s *PetAccountService) Credit(ctx context.Context, petID string, taiAmount, creditAmount float64, idempotencyKey string) (*handler.CreditResponse, error) {
	petAcc, err := s.db.PetAccount.Query().
		Where(petaccount.PetID(petID)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("pet %s not found: %w", petID, err)
	}

	if petAcc.Status != "active" {
		return nil, fmt.Errorf("pet %s is %s, cannot credit", petID, petAcc.Status)
	}

	// Check daily limit
	s.resetDailyIfNeeded(ctx, petAcc)
	if petAcc.DailyTAIUsed+taiAmount > petAcc.DailyTALimit {
		return nil, fmt.Errorf("pet %s daily TAI limit exceeded (%.2f/%.2f)",
			petID, petAcc.DailyTAIUsed+taiAmount, petAcc.DailyTALimit)
	}

	// Apply credit via ledger (idempotent)
	creditCtx := creditctx.WithMetadata(ctx, creditctx.Metadata{
		EntryType:      "tai_compute_credit",
		SourceType:     "tai_pet",
		SourceID:       petID,
		IdempotencyKey: idempotencyKey,
		Transferable:   true,
		CountRecharge:  true,
	})

	err = s.ledger.Apply(creditCtx, s.db, petAcc.UserID, decimal.NewFromFloat(creditAmount), nil, false)
	if err != nil {
		return nil, fmt.Errorf("credit ledger apply: %w", err)
	}

	// Update pet account tracking
	_, err = s.db.PetAccount.UpdateOneID(petAcc.ID).
		AddTaiSpentTotal(taiAmount).
		AddComputeCreditsTotal(creditAmount).
		AddDailyTAIUsed(taiAmount).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update pet account: %w", err)
	}

	// Get new balance
	user, _ := s.db.User.Get(ctx, petAcc.UserID)
	newBalance := 0.0
	if user != nil {
		newBalance = user.Balance
	}

	return &handler.CreditResponse{
		PetID:         petID,
		Credited:      creditAmount,
		NewBalance:    newBalance,
		TAISpentTotal: petAcc.TaiSpentTotal + taiAmount,
	}, nil
}

// GetStatus returns a pet's current account status.
func (s *PetAccountService) GetStatus(ctx context.Context, petID string) (*handler.PetStatusResponse, error) {
	petAcc, err := s.db.PetAccount.Query().
		Where(petaccount.PetID(petID)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("pet %s not found", petID)
	}

	s.resetDailyIfNeeded(ctx, petAcc)

	user, _ := s.db.User.Get(ctx, petAcc.UserID)
	balance := 0.0
	if user != nil {
		balance = user.Balance
	}

	apiKeyStatus := "unknown"
	if petAcc.APIKeyID != nil {
		key, err := s.db.APIKey.Get(ctx, *petAcc.APIKeyID)
		if err == nil {
			apiKeyStatus = key.Status
		}
	}

	return &handler.PetStatusResponse{
		PetID:        petID,
		Status:       petAcc.Status,
		Balance:      balance,
		TAISpentTotal: petAcc.TaiSpentTotal,
		DailyTAIUsed: petAcc.DailyTAIUsed,
		DailyTALimit: petAcc.DailyTALimit,
		APIKeyStatus: apiKeyStatus,
	}, nil
}

// SetStatus changes a pet's status and syncs the API key.
func (s *PetAccountService) SetStatus(ctx context.Context, petID, status string) error {
	petAcc, err := s.db.PetAccount.Query().
		Where(petaccount.PetID(petID)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("pet %s not found", petID)
	}

	// Sync API key status
	if petAcc.APIKeyID != nil {
		keyStatus := "active"
		if status == "suspended" {
			keyStatus = "disabled"
		}
		_, _ = s.db.APIKey.UpdateOneID(*petAcc.APIKeyID).
			SetStatus(keyStatus).
			Save(ctx)
	}

	_, err = s.db.PetAccount.UpdateOneID(petAcc.ID).
		SetStatus(status).
		Save(ctx)
	return err
}

// BatchUsage returns usage for multiple pets.
func (s *PetAccountService) BatchUsage(ctx context.Context, petIDs []string) ([]handler.PetUsageItem, error) {
	pets, err := s.db.PetAccount.Query().
		Where(petaccount.PetIDIn(petIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]handler.PetUsageItem, 0, len(pets))
	for _, p := range pets {
		balance := 0.0
		user, _ := s.db.User.Get(ctx, p.UserID)
		if user != nil {
			balance = user.Balance
		}
		items = append(items, handler.PetUsageItem{
			PetID:        p.PetID,
			Balance:      balance,
			DailyTAIUsed: p.DailyTAIUsed,
			Status:       p.Status,
		})
	}
	return items, nil
}

// resetDailyIfNeeded resets daily usage counter if the day has passed.
func (s *PetAccountService) resetDailyIfNeeded(ctx context.Context, petAcc *ent.PetAccount) {
	now := time.Now()
	if petAcc.DailyResetAt == nil || now.After(*petAcc.DailyResetAt) {
		nextReset := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
		_, _ = s.db.PetAccount.UpdateOneID(petAcc.ID).
			SetDailyTAIUsed(0).
			SetDailyResetAt(nextReset).
			Save(ctx)
		petAcc.DailyTAIUsed = 0
	}
}
