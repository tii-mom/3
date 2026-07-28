-- 199: 3API AI 工具箱
-- 仅新增工具推荐表和默认工具数据，不修改或删除现有用户、订单、余额数据。

CREATE TABLE IF NOT EXISTS ai_tools (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL DEFAULT 1 REFERENCES saas_tenants(id) ON DELETE RESTRICT,
    name VARCHAR(160) NOT NULL,
    category VARCHAR(80) NOT NULL DEFAULT 'AI 编程',
    description TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_ai_tools_public
    ON ai_tools(tenant_id, status, sort_order, id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_ai_tools_admin
    ON ai_tools(tenant_id, sort_order, id)
    WHERE deleted_at IS NULL;

INSERT INTO ai_tools (tenant_id, name, category, description, url, status, sort_order)
SELECT 1, '2 年 36 元网络环境服务', '网络环境', '适合需要稳定网络环境的用户，购买前请自行确认服务内容和可用性。', 'https://kitty.work/', 'published', 10
WHERE NOT EXISTS (SELECT 1 FROM ai_tools WHERE tenant_id = 1 AND url = 'https://kitty.work/' AND deleted_at IS NULL);

INSERT INTO ai_tools (tenant_id, name, category, description, url, status, sort_order)
SELECT 1, '静态住宅 IP', '网络环境', '提供静态住宅 IP 服务，适合对访问环境稳定性有要求的场景。', 'https://1024proxy.com/', 'published', 20
WHERE NOT EXISTS (SELECT 1 FROM ai_tools WHERE tenant_id = 1 AND url = 'https://1024proxy.com/' AND deleted_at IS NULL);

INSERT INTO ai_tools (tenant_id, name, category, description, url, status, sort_order)
SELECT 1, 'CCS API 管理器', 'API / CCS 工具', '用于管理 CCS API 配置，适合需要快速切换和维护 API 配置的用户。', 'https://ccswitch.lovable.app/', 'published', 30
WHERE NOT EXISTS (SELECT 1 FROM ai_tools WHERE tenant_id = 1 AND url = 'https://ccswitch.lovable.app/' AND deleted_at IS NULL);

INSERT INTO ai_tools (tenant_id, name, category, description, url, status, sort_order)
SELECT 1, 'Cockpit-tools 多账号管理器', 'Codex / 多模型工具', '开源多账号管理工具，适合需要集中管理多个账号和配置的进阶用户。', 'https://github.com/jlcodes99/cockpit-tools', 'published', 40
WHERE NOT EXISTS (SELECT 1 FROM ai_tools WHERE tenant_id = 1 AND url = 'https://github.com/jlcodes99/cockpit-tools' AND deleted_at IS NULL);

INSERT INTO ai_tools (tenant_id, name, category, description, url, status, sort_order)
SELECT 1, 'Opencodex 多模型接入工具', 'Codex / 多模型工具', '可将 Claude、Gemini、Grok、DeepSeek、Ollama 等模型接入 Codex CLI、App、SDK 和 Claude Code。', 'https://github.com/lidge-jun/opencodex', 'published', 50
WHERE NOT EXISTS (SELECT 1 FROM ai_tools WHERE tenant_id = 1 AND url = 'https://github.com/lidge-jun/opencodex' AND deleted_at IS NULL);

INSERT INTO ai_tools (tenant_id, name, category, description, url, status, sort_order)
SELECT 1, '网络环境检测工具', '网络环境', '用于快速检测当前网络环境和 IP 信息，便于排查连接问题。', 'https://pingip.cn/', 'published', 60
WHERE NOT EXISTS (SELECT 1 FROM ai_tools WHERE tenant_id = 1 AND url = 'https://pingip.cn/' AND deleted_at IS NULL);
