-- Ensure every active user has a complete, zero-safe distribution profile.
-- This migration is additive and idempotent; it does not alter balances,
-- commissions, recharge events, or existing invitation ownership.
WITH programs AS (
    SELECT id, tenant_id
    FROM distribution_programs
    WHERE tenant_id = 1 AND code = 'compute_company'
)
INSERT INTO distribution_members (program_id, tenant_id, user_id)
SELECT p.id, p.tenant_id, u.id
FROM programs p
CROSS JOIN users u
WHERE u.deleted_at IS NULL
ON CONFLICT DO NOTHING;

WITH programs AS (
    SELECT id, tenant_id
    FROM distribution_programs
    WHERE tenant_id = 1 AND code = 'compute_company'
)
INSERT INTO distribution_relations (
    program_id, tenant_id, ancestor_user_id, descendant_user_id, depth
)
SELECT p.id, p.tenant_id, u.id, u.id, 0
FROM programs p
CROSS JOIN users u
WHERE u.deleted_at IS NULL
ON CONFLICT DO NOTHING;

WITH RECURSIVE chain AS (
    SELECT ua.inviter_id AS ancestor_user_id,
           ua.user_id AS descendant_user_id,
           1 AS depth
    FROM user_affiliates ua
    JOIN users child ON child.id = ua.user_id AND child.deleted_at IS NULL
    JOIN users parent ON parent.id = ua.inviter_id AND parent.deleted_at IS NULL
    WHERE ua.inviter_id IS NOT NULL AND ua.inviter_id <> ua.user_id

    UNION ALL

    SELECT parent.inviter_id,
           chain.descendant_user_id,
           chain.depth + 1
    FROM chain
    JOIN user_affiliates parent ON parent.user_id = chain.ancestor_user_id
    WHERE chain.depth < 5
      AND parent.inviter_id IS NOT NULL
      AND parent.inviter_id <> chain.descendant_user_id
), programs AS (
    SELECT id, tenant_id
    FROM distribution_programs
    WHERE tenant_id = 1 AND code = 'compute_company'
)
INSERT INTO distribution_relations (
    program_id, tenant_id, ancestor_user_id, descendant_user_id, depth
)
SELECT p.id, p.tenant_id,
       chain.ancestor_user_id, chain.descendant_user_id, chain.depth
FROM programs p
CROSS JOIN chain
WHERE chain.ancestor_user_id <> chain.descendant_user_id
ON CONFLICT DO NOTHING;

WITH programs AS (
    SELECT id, tenant_id
    FROM distribution_programs
    WHERE tenant_id = 1 AND code = 'compute_company'
)
INSERT INTO distribution_cash_wallets (program_id, tenant_id, user_id)
SELECT p.id, p.tenant_id, u.id
FROM programs p
CROSS JOIN users u
WHERE u.deleted_at IS NULL
ON CONFLICT DO NOTHING;
