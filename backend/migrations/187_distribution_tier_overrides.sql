-- Allow administrators to assign a distribution tier before the automatic
-- team-volume threshold is reached. The automatic tier remains in
-- distribution_members.current_tier and can be restored by clearing the
-- override.
ALTER TABLE distribution_members
    ADD COLUMN IF NOT EXISTS tier_override SMALLINT,
    ADD COLUMN IF NOT EXISTS tier_override_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS tier_override_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS tier_override_reason TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'distribution_members_tier_override_check'
    ) THEN
        ALTER TABLE distribution_members
            ADD CONSTRAINT distribution_members_tier_override_check
            CHECK (tier_override IS NULL OR tier_override BETWEEN 1 AND 3);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_distribution_members_tier_override
    ON distribution_members(program_id, tier_override)
    WHERE tier_override IS NOT NULL;
