-- Migration: add request attempt counters for latency and failover analysis.
--
-- These columns are intentionally additive and nullable-free with constant
-- defaults. They do not store request bodies or prompt content.

ALTER TABLE usage_logs
  ADD COLUMN IF NOT EXISTS attempt_count INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS failover_count INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN usage_logs.attempt_count IS 'Actual upstream account attempts for this request; at least 1.';
COMMENT ON COLUMN usage_logs.failover_count IS 'Number of account switches during this request.';

ALTER TABLE ops_error_logs
  ADD COLUMN IF NOT EXISTS attempt_count INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS failover_count INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN ops_error_logs.attempt_count IS 'Actual upstream account attempts for the failed request; at least 1.';
COMMENT ON COLUMN ops_error_logs.failover_count IS 'Number of account switches observed before the failed request completed.';
