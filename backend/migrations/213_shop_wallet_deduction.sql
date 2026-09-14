-- 商城下单支持「人民币返点余额抵扣」（组合支付）。
--
-- 背景：
--   推广返点结算进 distribution_cash_wallets（人民币），与只能调 API 的
--   平台额度（美元，user_credit_accounts）是两种钱。此前返点只能提现或
--   兑换成平台额度，无法直接在商城消费。
--
-- 本迁移只加两列，不改任何既有数据语义：
--   wallet_applied_cny_minor  本单用返点余额抵扣的金额（人民币分）
--   payable_cny_minor         本单还需外部支付（微信/支付宝）的金额（人民币分）
--
-- 恒等式：payable_cny_minor = snapshot_price_cny_minor - wallet_applied_cny_minor
-- 由 CHECK 约束保证，避免出现「抵扣 + 实付 ≠ 商品价」的脏订单。
--
-- 返佣基数说明（重要）：
--   shop_commission_records.base_cny_minor 取的是**实付金额**，不是商品原价。
--   用余额抵扣掉的那部分钱本身就是平台已经付出过的返点，再计一次返佣会形成
--   「返点 → 抵扣下单 → 再返点」的无锚增发。所以全额抵扣的订单不产生返佣。
--
-- 历史数据：
--   存量订单 wallet_applied 默认为 0，payable 回填成 snapshot_price_cny_minor，
--   与迁移前的行为完全一致（旧订单等价于「全额外部支付」）。
--
-- 幂等：ADD COLUMN IF NOT EXISTS + 带条件的 UPDATE，重复执行无副作用。

ALTER TABLE shop_orders
    ADD COLUMN IF NOT EXISTS wallet_applied_cny_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS payable_cny_minor BIGINT NOT NULL DEFAULT 0;

-- 存量回填：只处理「从未抵扣过」的订单（两者都还是默认值），
-- 这样即便迁移被重复执行，也不会把已抵扣订单的 payable 改回原价。
UPDATE shop_orders
SET payable_cny_minor = snapshot_price_cny_minor
WHERE wallet_applied_cny_minor = 0
  AND payable_cny_minor = 0
  AND snapshot_price_cny_minor > 0;

ALTER TABLE shop_orders DROP CONSTRAINT IF EXISTS shop_orders_wallet_applied_range_check;
ALTER TABLE shop_orders
    ADD CONSTRAINT shop_orders_wallet_applied_range_check
    CHECK (wallet_applied_cny_minor >= 0 AND wallet_applied_cny_minor <= snapshot_price_cny_minor);

ALTER TABLE shop_orders DROP CONSTRAINT IF EXISTS shop_orders_payable_identity_check;
ALTER TABLE shop_orders
    ADD CONSTRAINT shop_orders_payable_identity_check
    CHECK (payable_cny_minor = snapshot_price_cny_minor - wallet_applied_cny_minor);

-- 抵扣记录落 distribution_wallet_ledger（action 为自由文本，无需改表）：
--   shop_wallet_deduction         下单抵扣
--   shop_wallet_deduction_refund  订单关闭（取消/超时/失败）退回抵扣
