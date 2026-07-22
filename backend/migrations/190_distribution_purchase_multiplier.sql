-- Use the payment balance purchase multiplier for new compute-company
-- conversions while preserving the legacy exchange snapshot on old rows.
ALTER TABLE distribution_usd_conversions
    ADD COLUMN IF NOT EXISTS cny_to_usd_rate DECIMAL(20,10),
    ADD COLUMN IF NOT EXISTS rate_source VARCHAR(64);

UPDATE distribution_usd_conversions
SET cny_to_usd_rate = ROUND(1 / usd_to_cny_rate, 10),
    rate_source = 'legacy_usd_to_cny_rate'
WHERE cny_to_usd_rate IS NULL;

UPDATE distribution_usd_conversions
SET rate_source = 'legacy_usd_to_cny_rate'
WHERE rate_source IS NULL;

ALTER TABLE distribution_usd_conversions
    ALTER COLUMN rate_source SET DEFAULT 'legacy_usd_to_cny_rate';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'distribution_usd_conversions_cny_to_usd_rate_check'
          AND conrelid = 'distribution_usd_conversions'::regclass
    ) THEN
        ALTER TABLE distribution_usd_conversions
            ADD CONSTRAINT distribution_usd_conversions_cny_to_usd_rate_check
            CHECK (cny_to_usd_rate IS NULL OR cny_to_usd_rate > 0);
    END IF;
END $$;
