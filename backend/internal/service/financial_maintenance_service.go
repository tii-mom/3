package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const (
	financialMaintenanceInterval = time.Minute
	financialMaintenanceTimeout  = 45 * time.Second
	financialOutboxStream        = "financial:events"
)

// FinancialMaintenanceService performs retryable financial housekeeping. The
// PostgreSQL rows remain authoritative; Redis only wakes downstream consumers.
type FinancialMaintenanceService struct {
	db           *sql.DB
	vouchers     *VoucherService
	distribution *DistributionService
	saas         *SaaSService
	redis        *redis.Client

	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewFinancialMaintenanceService(db *sql.DB, vouchers *VoucherService, distribution *DistributionService, saas *SaaSService, redisClient *redis.Client) *FinancialMaintenanceService {
	return &FinancialMaintenanceService{db: db, vouchers: vouchers, distribution: distribution, saas: saas, redis: redisClient}
}

func (s *FinancialMaintenanceService) Start() {
	if s == nil || s.db == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.wg.Add(1)
	go s.run(ctx)
}

func (s *FinancialMaintenanceService) Stop() {
	if s == nil {
		return
	}
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
		s.wg.Wait()
	}
}

func (s *FinancialMaintenanceService) run(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(financialMaintenanceInterval)
	defer ticker.Stop()
	s.runOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

func (s *FinancialMaintenanceService) runOnce(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, financialMaintenanceTimeout)
	defer cancel()
	if s.vouchers != nil {
		if err := s.vouchers.ExpireDue(ctx, 200); err != nil {
			logger.LegacyPrintf("service.financial_maintenance", "expire vouchers failed: %v", err)
		}
	}
	if s.distribution != nil {
		if err := s.distribution.ThawDue(ctx, 200); err != nil {
			logger.LegacyPrintf("service.financial_maintenance", "thaw distribution commissions failed: %v", err)
		}
	}
	if s.saas != nil {
		if err := s.saas.ThawPartnersDue(ctx, 200); err != nil {
			logger.LegacyPrintf("service.financial_maintenance", "thaw partner commissions failed: %v", err)
		}
	}
	if err := s.dispatchOutbox(ctx, 100); err != nil {
		logger.LegacyPrintf("service.financial_maintenance", "dispatch financial outbox failed: %v", err)
	}
}

type financialOutboxItem struct {
	id             int64
	tenantID       int64
	aggregateType  string
	aggregateID    string
	eventType      string
	payload        string
	idempotencyKey string
	attempts       int
}

func (s *FinancialMaintenanceService) dispatchOutbox(ctx context.Context, limit int) error {
	if s == nil || s.db == nil || s.redis == nil {
		return nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.QueryContext(ctx, `
SELECT id, tenant_id, aggregate_type, aggregate_id, event_type, payload::text, idempotency_key, attempts
FROM financial_outbox_events
WHERE status = 'pending' AND available_at <= NOW()
ORDER BY id
FOR UPDATE SKIP LOCKED
LIMIT $1`, limit)
	if err != nil {
		return err
	}
	items := make([]financialOutboxItem, 0, limit)
	for rows.Next() {
		var item financialOutboxItem
		if err := rows.Scan(&item.id, &item.tenantID, &item.aggregateType, &item.aggregateID, &item.eventType, &item.payload, &item.idempotencyKey, &item.attempts); err != nil {
			_ = rows.Close()
			return err
		}
		items = append(items, item)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, item := range items {
		_, publishErr := s.redis.XAdd(ctx, &redis.XAddArgs{
			Stream: financialOutboxStream,
			MaxLen: 100000,
			Approx: true,
			Values: map[string]any{
				"outbox_id": item.id, "tenant_id": item.tenantID,
				"aggregate_type": item.aggregateType, "aggregate_id": item.aggregateID,
				"event_type": item.eventType, "payload": item.payload,
				"idempotency_key": item.idempotencyKey,
			},
		}).Result()
		if publishErr != nil {
			delay := financialOutboxRetryDelay(item.attempts + 1)
			_, updateErr := tx.ExecContext(ctx, `UPDATE financial_outbox_events SET attempts = attempts + 1, last_error = $2, available_at = NOW() + $3::interval WHERE id = $1`, item.id, truncateFinancialError(publishErr.Error()), fmt.Sprintf("%d seconds", int(delay.Seconds())))
			if updateErr != nil {
				return updateErr
			}
			break
		}
		if _, err := tx.ExecContext(ctx, `UPDATE financial_outbox_events SET status = 'processed', attempts = attempts + 1, processed_at = NOW(), last_error = NULL WHERE id = $1`, item.id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func financialOutboxRetryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 8 {
		attempt = 8
	}
	return time.Duration(1<<uint(attempt-1)) * time.Minute
}

func truncateFinancialError(value string) string {
	const max = 1000
	if len(value) <= max {
		return value
	}
	return value[:max]
}
