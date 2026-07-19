package service

import (
	"context"
	"database/sql"
	"encoding/json"
)

func insertFinancialOutboxEvent(ctx context.Context, tx *sql.Tx, aggregateType, aggregateID, eventType, idempotencyKey string, payload map[string]any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO financial_outbox_events (tenant_id, aggregate_type, aggregate_id, event_type, payload, idempotency_key) VALUES (1, $1, $2, $3, $4::jsonb, $5) ON CONFLICT (tenant_id, idempotency_key) DO NOTHING`, aggregateType, aggregateID, eventType, string(data), idempotencyKey)
	return err
}
