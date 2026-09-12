-- ============================================================================
-- 3api 商城：将 12 个「公开」商品分到真实品类（当前全部为 other）
-- ----------------------------------------------------------------------------
-- 目的：让官网首页 #plans 的 4 列对比 + 商城品类 Tab 能正常渲染分组。
-- 约束：仅更新 category 列，不动表结构、不加列、不跑迁移。
--       满足用户「不要进行修改或重置，当前有用户数据」前提：原子事务，
--       任一商品受影响行数 != 1 立即整体回滚，绝不产生半成品改动。
-- 范围：仅 tenant_id=1 且 status='published' 且 deleted_at IS NULL 的 12 个公开商品。
--       另有 6 个非公开行（draft/archived+deleted：id 3/6/7/9/10/16）保持不动。
-- 枚举：category 取值必须是 shop_products_category_check 的成员：
--       gpt_topup | gpt_account | x_premium | gemini | codex | other
-- ============================================================================

-- 执行方式（生产 VPS）：用 DO 块整体包成事务，逐条断言 ROW_COUNT = 1，
-- 任一条不符 → RAISE → 整个事务回滚，数据库保持原样。
DO $$
DECLARE
  r int;
BEGIN
  UPDATE shop_products SET category = 'gpt_account', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 1  AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=1 affected %', r; END IF;

  UPDATE shop_products SET category = 'gpt_topup', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 2  AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=2 affected %', r; END IF;

  UPDATE shop_products SET category = 'gpt_topup', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 4  AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=4 affected %', r; END IF;

  UPDATE shop_products SET category = 'gpt_topup', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 5  AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=5 affected %', r; END IF;

  UPDATE shop_products SET category = 'codex', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 8  AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=8 affected %', r; END IF;

  UPDATE shop_products SET category = 'gpt_topup', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 11 AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=11 affected %', r; END IF;

  UPDATE shop_products SET category = 'gpt_topup', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 12 AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=12 affected %', r; END IF;

  UPDATE shop_products SET category = 'x_premium', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 13 AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=13 affected %', r; END IF;

  UPDATE shop_products SET category = 'gemini', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 14 AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=14 affected %', r; END IF;

  UPDATE shop_products SET category = 'x_premium', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 15 AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=15 affected %', r; END IF;

  UPDATE shop_products SET category = 'gemini', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 17 AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=17 affected %', r; END IF;

  UPDATE shop_products SET category = 'other', updated_at = NOW()
    WHERE tenant_id = 1 AND id = 18 AND status = 'published' AND deleted_at IS NULL;
  GET DIAGNOSTICS r = ROW_COUNT; IF r <> 1 THEN RAISE EXCEPTION 'id=18 affected %', r; END IF;
END $$;

-- 校验：执行后应看到 6 个品类（gpt_topup / gpt_account / x_premium / gemini / codex / other）
-- SELECT category, count(*) FROM shop_products WHERE tenant_id = 1 GROUP BY category ORDER BY category;
