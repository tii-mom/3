-- 209: 免登录（访客）下单限流改为跨进程持久化
--
-- 背景：原实现是进程内 map[string][]time.Time 的滑动窗口（见
--       internal/service/shop_guest_service.go 的 guestOrderLimiter）。
--       多副本部署时，每个副本各算各的窗口，重启还会把计数清零——
--       攻击者只要把请求打散到不同副本、或等一次重启，
--       「10 分钟 / IP 20 次、/ 联系方式 5 次」的上限就会被成倍放大。
--
-- 方案：固定窗口计数器落到数据库。bucket_key 作为主键提供行级锁，
--       并发请求由 PostgreSQL 串行化同一 key 的读改写，天然跨副本一致；
--       这也符合本项目「无 Redis 依赖也能正确工作」的部署假设。
--
-- 安全前提（沿用既有约定）：
--   1. 只新建表，不修改、不删除任何现有表与数据；
--   2. 所有语句幂等，重复执行无副作用，灾备重放安全。

CREATE TABLE IF NOT EXISTS shop_guest_order_windows (
    -- 形如 ip:1.2.3.4 / contact:a@b.com，长度按最长邮箱留足余量
    bucket_key   VARCHAR(320) NOT NULL PRIMARY KEY,
    window_start TIMESTAMPTZ  NOT NULL,
    hits         INTEGER      NOT NULL DEFAULT 0,
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 过期清理按 window_start 扫描
CREATE INDEX IF NOT EXISTS idx_shop_guest_order_windows_start
    ON shop_guest_order_windows(window_start);
