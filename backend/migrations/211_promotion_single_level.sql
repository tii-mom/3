-- 把「算力公司」的多层分销收敛为「推广计划」的单层三档阶梯返点。
--
-- 背景：
--   原方案（v1 见 177、v2 见 188）是多层分销：按团队业绩逐层拨出
--   10/0/0/0/0、10/4/3/2/1、15/6/4/3/2、20/8/6/4/2，合计 10%~40%。
--   两个致命问题：
--     1) 只适配「预付款型」的 API 余额充值，照搬到现货商城卖一单亏一单；
--     2) 门槛沿用老 API 中转站「批发/预付款客户」的量级（¥100 万 / ¥1000 万），
--        对现在的商城 + 小额充值完全不适用 —— 几乎没人能升档，阶梯等于死的，
--        玩家看不到升级希望，激励等于没有。
--
-- 目标（单层：只结算「直接邀请人」，service 层 promotionMaxDepth = 1）：
--     T0  ¥0        → 5%    （人人起步，零门槛）
--     T1  ¥5,000    → 8%
--     T2  ¥50,000   → 10%   （封顶）
--   门槛口径 = 该用户「直接邀请的人」的累计已生效充值额
--   （distribution_members.team_volume_cny_minor，本迁移按 depth = 1 重算，见第 3 步）；
--   level2_bps ~ level5_bps 新版本一律写 0。
--
-- 版本号：
--   沿用 188 的写法取 current_config_version + 1（本次落到 v3）。
--   绝不能硬编码版本号：v2 已被 188 占用，直接写 2 会让
--   INSERT ... ON CONFLICT DO NOTHING 全部静默命中、UPDATE 的
--   WHERE current_config_version < 2 也不命中，整份迁移空转且不报错。
--
-- 数据覆盖范围（用户已确认「算力公司数据可以覆盖」）：
--   只动 distribution_* 系列表（配置版本、档位表、成员档位与直属业绩、手工指定档位）。
--   **不触碰** users / user_credit_accounts / payment_orders / shop_orders，
--   也**不改任何钱包余额与已发佣金流水**：distribution_cash_wallets 的
--   available / frozen / withdrawing / lifetime_earned、distribution_commissions
--   全部原样保留。⇒ 用户的账户余额与已经到手的返点完全不受影响。
--
-- 幂等：重复执行时，若已存在单层配置（T1 的 level2~level5 全为 0）则直接返回。

DO $$
DECLARE
    compute_program_id BIGINT;
    previous_version INTEGER;
    next_version INTEGER;
BEGIN
    SELECT id, current_config_version
    INTO compute_program_id, previous_version
    FROM distribution_programs
    WHERE tenant_id = 1 AND code = 'compute_company'
    FOR UPDATE;

    IF compute_program_id IS NULL THEN
        RAISE EXCEPTION 'compute_company distribution program is missing';
    END IF;

    -- 幂等闸门：历史版本（177 的 v1、188 的 v2）的 T1 都带着 4/3/2/1 的深层费率，
    -- 只有本迁移产生的单层版本才会让 T1 的 level2~level5 全为 0。
    IF EXISTS (
        SELECT 1
        FROM distribution_tier_configs t
        WHERE t.program_id = compute_program_id
          AND t.tier = 1
          AND t.level2_bps = 0
          AND t.level3_bps = 0
          AND t.level4_bps = 0
          AND t.level5_bps = 0
    ) THEN
        RETURN;
    END IF;

    next_version := previous_version + 1;

    -- 1) 新版本的单层三档阶梯。
    INSERT INTO distribution_tier_configs
        (program_id, config_version, tier, threshold_cny_minor, level1_bps, level2_bps, level3_bps, level4_bps, level5_bps)
    VALUES
        (compute_program_id, next_version, 0,       0,  500, 0, 0, 0, 0),
        (compute_program_id, next_version, 1,  500000,  800, 0, 0, 0, 0),
        (compute_program_id, next_version, 2, 5000000, 1000, 0, 0, 0, 0)
    ON CONFLICT (program_id, config_version, tier) DO NOTHING;

    INSERT INTO distribution_policy_versions
        (program_id, config_version, commission_freeze_hours, withdrawal_min_cny_minor,
         withdrawal_daily_limit, withdrawal_fee_bps, first_recharge_bonus_bps,
         first_recharge_bonus_cap_usd)
    SELECT p.id, next_version, p.commission_freeze_hours, p.withdrawal_min_cny_minor,
           p.withdrawal_daily_limit, p.withdrawal_fee_bps, p.first_recharge_bonus_bps,
           p.first_recharge_bonus_cap_usd
    FROM distribution_programs p
    WHERE p.id = compute_program_id
    ON CONFLICT (program_id, config_version) DO NOTHING;

    UPDATE distribution_programs
    SET current_config_version = next_version,
        name = '推广计划',
        updated_at = NOW()
    WHERE id = compute_program_id;

    -- 2) 档位手工指定重新定档：阶梯从 T0~T3 变成 T0~T2，旧的指定全部作废，
    --    统一回到「按重算后的直属业绩自动落档」，避免留下指向不存在档位的脏值。
    --    这里不带 program_id 过滤（库内只有 compute_company 一个程序），
    --    确保下面的 CHECK 收紧不会被任何残留的旧值挡住。
    UPDATE distribution_members
    SET tier_override = NULL,
        tier_override_by = NULL,
        tier_override_at = NULL,
        tier_override_reason = NULL,
        updated_at = NOW()
    WHERE tier_override IS NOT NULL;

    ALTER TABLE distribution_members DROP CONSTRAINT IF EXISTS distribution_members_tier_override_check;
    ALTER TABLE distribution_members
        ADD CONSTRAINT distribution_members_tier_override_check CHECK (tier_override IS NULL OR tier_override BETWEEN 0 AND 2);

    -- 3) 直属业绩按新口径重算：只累计 depth = 1（直接邀请）的已生效充值。
    --    存量里含 2~5 层的旧口径业绩会让老成员档位虚高，这里一并覆盖掉。
    --    注意必须在重算业绩之后再算档位：PostgreSQL 的 UPDATE 里所有 SET 表达式
    --    读到的都是旧行值，放同一条语句里会用旧业绩算档位。
    UPDATE distribution_members AS m
    SET team_volume_cny_minor = COALESCE((
            SELECT SUM(e.base_cny_minor)
            FROM distribution_relations r
            JOIN distribution_recharge_events e
              ON e.program_id = r.program_id
             AND e.user_id = r.descendant_user_id
             AND e.status = 'APPLIED'
            WHERE r.program_id = m.program_id
              AND r.ancestor_user_id = m.user_id
              AND r.depth = 1
        ), 0),
        updated_at = NOW()
    WHERE m.program_id = compute_program_id;

    UPDATE distribution_members AS m
    SET current_tier = COALESCE((
            SELECT t.tier
            FROM distribution_tier_configs t
            WHERE t.program_id = m.program_id
              AND t.config_version = next_version
              AND t.threshold_cny_minor <= m.team_volume_cny_minor
            ORDER BY t.tier DESC
            LIMIT 1
        ), 0),
        updated_at = NOW()
    WHERE m.program_id = compute_program_id;
END $$;
