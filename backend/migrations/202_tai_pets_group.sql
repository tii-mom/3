-- 202_tai_pets_group.sql
-- TAI Protocol Phase 0: platform pool account for AI pet compute
-- Creates a bulk-pricing group and documents the setup for the platform API key.

-- Create the tai-pets pricing group (40% bulk discount)
INSERT INTO groups (
    name,
    rate_multiplier,
    status,
    platform,
    created_at,
    updated_at
)
SELECT
    'tai-pets',
    0.6000,          -- 40% discount vs retail
    'active',
    'openai',         -- default platform; pets use OpenAI-compatible endpoint
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM groups WHERE name = 'tai-pets'
);

-- NOTE: After running this migration, manually:
-- 1. Create a platform user: username = "tai-platform"
-- 2. Create an API key for that user, assigned to group "tai-pets"
-- 3. Set the key in TAI Protocol's .env as THREEAPI_PLATFORM_KEY
-- 4. Top up the platform user's balance (funded by TAI treasury revenue)
--
-- All pet API calls are proxied through TAI backend using this single key.
-- Per-pet tracking is handled in TAI Protocol's own database.

COMMENT ON TABLE groups IS 'Added tai-pets group for TAI Protocol bulk compute pricing (Phase 0)';
