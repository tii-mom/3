-- 200: 简化 3API AI 工具箱默认文案
-- 只更新迁移 199 写入的默认工具文案；如果管理员已经手动改过，不覆盖。

UPDATE ai_tools
SET description = '稳定网络环境服务。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://kitty.work/'
  AND description = '适合需要稳定网络环境的用户，购买前请自行确认服务内容和可用性。'
  AND deleted_at IS NULL;

UPDATE ai_tools
SET description = '静态住宅 IP 服务。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://1024proxy.com/'
  AND description = '提供静态住宅 IP 服务，适合对访问环境稳定性有要求的场景。'
  AND deleted_at IS NULL;

UPDATE ai_tools
SET category = 'API 工具',
    description = '管理 CCS API 配置。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://ccswitch.lovable.app/'
  AND category = 'API / CCS 工具'
  AND description = '用于管理 CCS API 配置，适合需要快速切换和维护 API 配置的用户。'
  AND deleted_at IS NULL;

UPDATE ai_tools
SET category = '多模型工具',
    description = '多账号管理工具。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://github.com/jlcodes99/cockpit-tools'
  AND category = 'Codex / 多模型工具'
  AND description = '开源多账号管理工具，适合需要集中管理多个账号和配置的进阶用户。'
  AND deleted_at IS NULL;

UPDATE ai_tools
SET category = '多模型工具',
    description = '多模型接入工具。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://github.com/lidge-jun/opencodex'
  AND category = 'Codex / 多模型工具'
  AND description = '可将 Claude、Gemini、Grok、DeepSeek、Ollama 等模型接入 Codex CLI、App、SDK 和 Claude Code。'
  AND deleted_at IS NULL;

UPDATE ai_tools
SET description = '查看网络和 IP 信息。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://pingip.cn/'
  AND description = '用于快速检测当前网络环境和 IP 信息，便于排查连接问题。'
  AND deleted_at IS NULL;
