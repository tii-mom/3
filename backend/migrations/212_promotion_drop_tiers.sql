-- 推广计划正式下线「档位」概念：返佣只按商品上架时设置的返佣比例结算。
--
-- 背景（承接 211）：
--   211 把多层分销收敛为「单层 + 三档阶梯（T0/T1/T2，5%/8%/10%）」。
--   但「档位」本身就是一个错误的抽象：不同商品利润天差地别，用一个全局口径
--   （团队累计充值额）去决定统一返点率，既不公平也不可控，还会把后台重新
--   拖回「多层分销 + 人工调档」的叙事。
--
--   现在的规则只有一条：**返佣比例跟着商品走**。
--   商品上架时设置 shop_products.commission_bps，下单时快照进
--   shop_orders.snapshot_commission_bps，结算写 shop_commission_records。
--   余额充值（API 额度）不再产生任何推广佣金。
--
-- 本迁移做什么：
--   ① 把已作废的档位状态清零：distribution_members.current_tier / tier_override；
--   ② 给这些列和 distribution_tier_configs 表打上 DEPRECATED 注释，
--      便于日后排查时一眼看出它们不再参与任何结算。
--
--   列与表一律**保留不删**：历史佣金流水（distribution_commissions）仍要能追溯，
--   直接 DROP 会让旧记录失去上下文，也会切断线上回滚路径。
--
-- 不触碰：users / user_credit_accounts / payment_orders / shop_orders，
--         以及任何钱包余额与已发佣金流水（金额一律原样保留）。
--
-- 幂等：两条 UPDATE 都带 WHERE 条件，重复执行不会产生额外写入。

DO $$
DECLARE
    compute_program_id BIGINT;
BEGIN
    SELECT id
    INTO compute_program_id
    FROM distribution_programs
    WHERE tenant_id = 1 AND code = 'compute_company';

    IF compute_program_id IS NULL THEN
        RETURN;
    END IF;

    -- 1) 手工指定档位作废（推广计划已无档位，后台也不再有调档入口）。
    UPDATE distribution_members
    SET tier_override = NULL,
        tier_override_by = NULL,
        tier_override_at = NULL,
        tier_override_reason = NULL,
        updated_at = NOW()
    WHERE tier_override IS NOT NULL;

    -- 2) 自动档位归零（仅清理非 0 的残留值，避免全表无谓写入）。
    UPDATE distribution_members
    SET current_tier = 0,
        updated_at = NOW()
    WHERE current_tier <> 0;
END $$;

-- 3) 列 / 表废弃标注：不改变数据，只让「它不再参与结算」这件事写进 schema。
COMMENT ON COLUMN distribution_members.current_tier IS
    'DEPRECATED（推广计划已无档位）：返佣按商品上架时的 commission_bps 结算，此列不再参与任何计算。';
COMMENT ON COLUMN distribution_members.tier_override IS
    'DEPRECATED（推广计划已无档位）：后台调档入口已下线，此列不再参与任何计算。';
COMMENT ON COLUMN distribution_members.tier_override_by IS
    'DEPRECATED（推广计划已无档位）：后台调档入口已下线，此列不再参与任何计算。';
COMMENT ON COLUMN distribution_members.tier_override_at IS
    'DEPRECATED（推广计划已无档位）：后台调档入口已下线，此列不再参与任何计算。';
COMMENT ON COLUMN distribution_members.tier_override_reason IS
    'DEPRECATED（推广计划已无档位）：后台调档入口已下线，此列不再参与任何计算。';
COMMENT ON COLUMN distribution_members.team_volume_cny_minor IS
    'DEPRECATED：团队业绩已改为「直属邀请成员在商城的实付额」，由 shop_orders 实时汇总，此列不再参与任何计算。';
COMMENT ON TABLE distribution_tier_configs IS
    'DEPRECATED（推广计划已无档位）：档位配置表保留仅为历史追溯，新配置版本不再写入，运行时也不再读取。';
