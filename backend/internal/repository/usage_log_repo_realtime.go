package repository

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

func (r *usageLogRepository) GetRealtimeMetrics(ctx context.Context, start, end time.Time) (*usagestats.RealtimeMetrics, error) {
	const query = `
		WITH success_stats AS (
			SELECT
				COUNT(*) AS request_count,
				COALESCE(SUM(duration_ms) FILTER (WHERE duration_ms IS NOT NULL), 0) AS duration_sum,
				COUNT(duration_ms) AS duration_count
			FROM usage_logs
			WHERE created_at >= $1 AND created_at < $2 AND actual_cost > 0
		),
		recent_errors AS (
			SELECT COALESCE(
				duration_ms::bigint,
				response_latency_ms,
				CASE
					WHEN auth_latency_ms IS NOT NULL OR routing_latency_ms IS NOT NULL OR upstream_latency_ms IS NOT NULL
					THEN COALESCE(auth_latency_ms, 0) + COALESCE(routing_latency_ms, 0) + COALESCE(upstream_latency_ms, 0)
				END
			) AS duration_ms
			FROM ops_error_logs
			WHERE created_at >= $1 AND created_at < $2 AND COALESCE(status_code, 0) >= 400
		),
		error_stats AS (
			SELECT
				COUNT(*) AS request_count,
				COALESCE(SUM(duration_ms) FILTER (WHERE duration_ms IS NOT NULL), 0) AS duration_sum,
				COUNT(duration_ms) AS duration_count
			FROM recent_errors
		)
		SELECT
			s.request_count + e.request_count AS requests_per_minute,
			CASE
				WHEN s.duration_count + e.duration_count = 0 THEN 0
				ELSE (s.duration_sum + e.duration_sum)::double precision / (s.duration_count + e.duration_count)
			END AS average_response_time,
			CASE
				WHEN s.request_count + e.request_count = 0 THEN 0
				ELSE e.request_count::double precision * 100 / (s.request_count + e.request_count)
			END AS error_rate
		FROM success_stats s CROSS JOIN error_stats e`

	metrics := &usagestats.RealtimeMetrics{}
	if err := scanSingleRow(
		ctx,
		r.sql,
		query,
		[]any{start, end},
		&metrics.RequestsPerMinute,
		&metrics.AverageResponseTime,
		&metrics.ErrorRate,
	); err != nil {
		return nil, err
	}
	return metrics, nil
}
