-- 206: 3API 公开销售订单交付扩展（GPT 充值 / 成品号 / 租号）
-- 仅新增列、索引与系统游客账号，不修改或删除任何现有数据。

-- 1. 商品：交付模式与展示配置
ALTER TABLE shop_products ADD COLUMN IF NOT EXISTS fulfillment_mode VARCHAR(24) NOT NULL DEFAULT 'manual';
ALTER TABLE shop_products ADD COLUMN IF NOT EXISTS delivery_form_hint TEXT NOT NULL DEFAULT '';
ALTER TABLE shop_products ADD COLUMN IF NOT EXISTS badge_text VARCHAR(24) NOT NULL DEFAULT '';
ALTER TABLE shop_products ADD COLUMN IF NOT EXISTS spec_label VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE shop_products ADD COLUMN IF NOT EXISTS highlight BOOLEAN NOT NULL DEFAULT FALSE;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'shop_products_fulfillment_mode_check'
    ) THEN
        ALTER TABLE shop_products
            ADD CONSTRAINT shop_products_fulfillment_mode_check
            CHECK (fulfillment_mode IN ('manual', 'session_topup', 'account_delivery', 'rental'));
    END IF;
END $$;

-- 2. 订单：游客归属、订单号与交付资料
--    delivery_payload 存 AES-256-GCM 密文（Session 凭证或收货信息），
--    delivery_hint 存脱敏片段（仅前若干位），供后台列表安全展示。
ALTER TABLE shop_orders ADD COLUMN IF NOT EXISTS order_no VARCHAR(32);
ALTER TABLE shop_orders ADD COLUMN IF NOT EXISTS guest_token VARCHAR(64);
ALTER TABLE shop_orders ADD COLUMN IF NOT EXISTS guest_contact VARCHAR(120) NOT NULL DEFAULT '';
ALTER TABLE shop_orders ADD COLUMN IF NOT EXISTS snapshot_fulfillment_mode VARCHAR(24) NOT NULL DEFAULT 'manual';
ALTER TABLE shop_orders ADD COLUMN IF NOT EXISTS delivery_payload TEXT NOT NULL DEFAULT '';
ALTER TABLE shop_orders ADD COLUMN IF NOT EXISTS delivery_hint VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE shop_orders ADD COLUMN IF NOT EXISTS delivery_submitted_at TIMESTAMPTZ;
ALTER TABLE shop_orders ADD COLUMN IF NOT EXISTS rental_duration VARCHAR(32) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS uq_shop_orders_order_no
    ON shop_orders(order_no) WHERE order_no IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_shop_orders_guest_token
    ON shop_orders(guest_token) WHERE guest_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_shop_orders_guest_contact
    ON shop_orders(tenant_id, guest_contact, created_at DESC) WHERE guest_contact <> '';

-- 3. 系统游客账号：承载免登录订单
--    password_hash 使用非法哈希值，任何登录尝试都会失败，该账号不可登录。
INSERT INTO users (email, password_hash, username, role, status, notes, created_at, updated_at)
SELECT 'guest@internal.3api.invalid',
       'nologin:system-guest-account-do-not-delete',
       '游客订单',
       'user',
       'active',
       '系统游客账号，用于承载免登录下单产生的订单，请勿删除或改密',
       NOW(),
       NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE email = 'guest@internal.3api.invalid' AND deleted_at IS NULL
);
