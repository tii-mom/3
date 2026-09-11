-- 207: 商品分类与图廊（官网首页按品类分组展示 + 商品多图）
-- 仅新增列、约束与索引，不修改或删除任何现有数据。

-- 1. 品类：与 fulfillment_mode（怎么交付）正交，这里表达「卖什么」
ALTER TABLE shop_products ADD COLUMN IF NOT EXISTS category VARCHAR(32) NOT NULL DEFAULT 'other';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'shop_products_category_check'
    ) THEN
        ALTER TABLE shop_products
            ADD CONSTRAINT shop_products_category_check
            CHECK (category IN ('gpt_topup', 'gpt_account', 'x_premium', 'gemini', 'codex', 'other'));
    END IF;
END $$;

-- 2. 图廊：附加展示图 URL 的 JSON 数组（主图仍用 image_url）。
--    空数组表示无附加图，前端退化为只展示主图。
ALTER TABLE shop_products ADD COLUMN IF NOT EXISTS gallery_json TEXT NOT NULL DEFAULT '[]';

CREATE INDEX IF NOT EXISTS idx_shop_products_category
    ON shop_products(category, sort_order, id);
