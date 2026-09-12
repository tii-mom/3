-- 208: 商城品类改为后台可管理（数据表驱动，替代硬编码枚举 + CHECK 约束）
--
-- 背景：迁移 207 把品类写成代码枚举 + shop_products_category_check 约束，
--       导致后台无法自行新增品类。这里把品类提升为独立数据表，后台可增删改。
--
-- 安全前提（沿用既有约定）：
--   1. 只新建表 + 插入种子数据，**不修改、不删除任何现有商品行**；
--   2. 所有语句幂等，重复执行无副作用；
--   3. 仅解除 207 留下的 CHECK 约束（枚举不再由数据库硬编码），不动列类型、不动数据。

-- ---------------------------------------------------------------------------
-- 1. 品类表
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shop_categories (
    id         BIGSERIAL    PRIMARY KEY,
    tenant_id  BIGINT       NOT NULL DEFAULT 1,
    slug       VARCHAR(32)  NOT NULL,
    label      VARCHAR(64)  NOT NULL,
    blurb      VARCHAR(255) NOT NULL DEFAULT '',
    sort_order INTEGER      NOT NULL DEFAULT 0,
    enabled    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    -- slug 会进入前端 DOM id（cat-<slug>）与 URL，限制为安全字符集
    CONSTRAINT shop_categories_slug_check CHECK (slug ~ '^[a-z][a-z0-9_]{0,31}$')
);

-- 同一租户下 slug 唯一
CREATE UNIQUE INDEX IF NOT EXISTS uq_shop_categories_tenant_slug
    ON shop_categories(tenant_id, slug);

-- 首页按「启用 + 排序」取列表
CREATE INDEX IF NOT EXISTS idx_shop_categories_tenant_enabled_sort
    ON shop_categories(tenant_id, enabled, sort_order, id);

-- ---------------------------------------------------------------------------
-- 2. 种子：把 207 的 6 个硬编码品类落成数据行，顺序与前台展示一致
--    标签按官网首页的可读性重写（GPT 官方充值 / GPT 成品号 / GPT 使用服务 / X 会员 …）
-- ---------------------------------------------------------------------------
INSERT INTO shop_categories (tenant_id, slug, label, blurb, sort_order, enabled) VALUES
    (1, 'gpt_topup',   'GPT 官方充值', '给已有账号续费升级',      10, TRUE),
    (1, 'gpt_account', 'GPT 成品号',   '开好即用的独享账号',      20, TRUE),
    (1, 'gpt_usage',   'GPT 使用服务', 'Codex / API 额度等增值服务', 30, TRUE),
    (1, 'x_premium',   'X 会员',       'X Premium 订阅开通',      40, TRUE),
    (1, 'gemini',      'Gemini',       'Gemini Advanced 订阅',    50, TRUE),
    (1, 'codex',       'Codex 额度',   'Codex / API 额度补充',    60, TRUE),
    (1, 'other',       '其他服务',     '其余增值服务',            90, TRUE)
ON CONFLICT (tenant_id, slug) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3. 解除 207 的 CHECK 约束
--    品类集合从此由 shop_categories 表定义；保留列本身的 NOT NULL DEFAULT 'other'，
--    因此历史数据与老代码写入仍然合法。
-- ---------------------------------------------------------------------------
ALTER TABLE shop_products DROP CONSTRAINT IF EXISTS shop_products_category_check;
