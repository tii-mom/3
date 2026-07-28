-- 198: 3API 商城基础表
-- 仅新增表/索引，不修改或删除现有用户、订单、余额数据。

CREATE TABLE IF NOT EXISTS shop_products (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    product_type VARCHAR(40) NOT NULL DEFAULT 'virtual' CHECK (product_type IN ('virtual', 'platform_usd_balance')),
    price_cny_minor BIGINT NOT NULL CHECK (price_cny_minor > 0),
    original_price_cny_minor BIGINT NOT NULL DEFAULT 0 CHECK (original_price_cny_minor >= 0),
    grant_usd_amount DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (grant_usd_amount >= 0),
    stock_quantity BIGINT,
    sold_count BIGINT NOT NULL DEFAULT 0 CHECK (sold_count >= 0),
    commission_bps INTEGER NOT NULL DEFAULT 0 CHECK (commission_bps >= 0 AND commission_bps <= 10000),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_shop_products_public
    ON shop_products(tenant_id, status, sort_order, id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS shop_banners (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    title VARCHAR(160) NOT NULL,
    subtitle TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    button_text VARCHAR(40) NOT NULL DEFAULT '立即查看',
    product_id BIGINT REFERENCES shop_products(id) ON DELETE SET NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_shop_banners_public
    ON shop_banners(tenant_id, enabled, sort_order, id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS shop_orders (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    product_id BIGINT NOT NULL REFERENCES shop_products(id) ON DELETE RESTRICT,
    payment_order_id BIGINT UNIQUE REFERENCES payment_orders(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'fulfilled', 'cancelled', 'refunded', 'failed')),
    fulfillment_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (fulfillment_status IN ('pending', 'fulfilled', 'failed')),
    commission_status VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (commission_status IN ('none', 'frozen', 'available', 'reversed')),
    snapshot_name VARCHAR(160) NOT NULL,
    snapshot_description TEXT NOT NULL DEFAULT '',
    snapshot_image_url TEXT NOT NULL DEFAULT '',
    snapshot_product_type VARCHAR(40) NOT NULL,
    snapshot_price_cny_minor BIGINT NOT NULL,
    snapshot_grant_usd_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    snapshot_commission_bps INTEGER NOT NULL DEFAULT 0,
    paid_at TIMESTAMPTZ,
    fulfilled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shop_orders_user
    ON shop_orders(tenant_id, user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shop_orders_admin
    ON shop_orders(tenant_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS shop_commission_records (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    shop_order_id BIGINT NOT NULL UNIQUE REFERENCES shop_orders(id) ON DELETE RESTRICT,
    product_id BIGINT NOT NULL REFERENCES shop_products(id) ON DELETE RESTRICT,
    buyer_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    beneficiary_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    base_cny_minor BIGINT NOT NULL CHECK (base_cny_minor > 0),
    commission_bps INTEGER NOT NULL CHECK (commission_bps > 0),
    amount_cny_minor BIGINT NOT NULL CHECK (amount_cny_minor > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'FROZEN' CHECK (status IN ('FROZEN', 'AVAILABLE', 'REVERSED')),
    frozen_until TIMESTAMPTZ NOT NULL,
    idempotency_key VARCHAR(180) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reversed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_shop_commission_beneficiary
    ON shop_commission_records(tenant_id, beneficiary_user_id, created_at DESC);
