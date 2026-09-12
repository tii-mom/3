-- 210: 新增「GPT 租号」品类，补齐首页页脚「租号」入口的落点
--
-- 背景：首页长期对外承诺三种交付方式「代充值 / 成品号 / 租号」（首页 .compare 三栏、
--       下单流程文案、页脚「服务」三入口都写了租号），后端 206 迁移也早已支持
--       fulfillment_mode='rental' + shop_orders.rental_duration；
--       但 208 的品类种子里只有 gpt_topup / gpt_account / gpt_usage / x_premium /
--       gemini / codex / other —— 没有租号品类，于是页脚「租号」只能指向 #plans 空锚，
--       后台也无法把租赁商品归到一个语义正确的品类下。
--
-- 安全前提（沿用既有约定）：
--   1. 只插入一行品类种子，**不修改、不删除任何现有商品与品类行**；
--   2. 语句幂等（ON CONFLICT (tenant_id, slug) DO NOTHING），重复执行/灾备重放无副作用；
--   3. 不新增约束、不改列类型。slug 满足 208 的 ^[a-z][a-z0-9_]{0,31}$ 校验。
--
-- sort_order=25 落在「GPT 成品号(20)」与「GPT 使用服务(30)」之间，
-- 与首页页脚「代充值 → 成品号 → 租号」的叙述顺序一致。
--
-- 注：本迁移只建「品类」。具体租赁商品（价格 / 时长 / 库存）仍由后台
--     /admin/shop/products 上架，fulfillment_mode 选 rental。
INSERT INTO shop_categories (tenant_id, slug, label, blurb, sort_order, enabled) VALUES
    (1, 'gpt_rental', 'GPT 租号', '按月租赁，短期试用', 25, TRUE)
ON CONFLICT (tenant_id, slug) DO NOTHING;
