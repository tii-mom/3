-- 201: 扩展 3API AI 工具箱默认介绍
-- 只更新仍保持系统默认文案的工具；如果管理员已手动改过，不覆盖。

UPDATE ai_tools
SET description = '用于日常访问 AI 工具时保持网络稳定，适合新手快速准备环境，购买前请确认套餐、地区和时长。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://kitty.work/'
  AND description IN (
    '适合需要稳定网络环境的用户，购买前请自行确认服务内容和可用性。',
    '稳定网络环境服务。'
  )
  AND deleted_at IS NULL;

UPDATE ai_tools
SET description = '提供静态住宅 IP，适合账号登录、环境固定和访问排查，使用前请确认地区、时长和平台规则。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://1024proxy.com/'
  AND description IN (
    '提供静态住宅 IP 服务，适合对访问环境稳定性有要求的场景。',
    '静态住宅 IP 服务。'
  )
  AND deleted_at IS NULL;

UPDATE ai_tools
SET description = '帮你集中保存和切换 CCS API 配置，适合多项目、多密钥用户，打开后按页面提示填写即可。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://ccswitch.lovable.app/'
  AND description IN (
    '用于管理 CCS API 配置，适合需要快速切换和维护 API 配置的用户。',
    '管理 CCS API 配置。'
  )
  AND deleted_at IS NULL;

UPDATE ai_tools
SET description = '开源多账号管理工具，适合同时维护多个账号和配置的进阶用户，使用前建议先阅读项目说明。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://github.com/jlcodes99/cockpit-tools'
  AND description IN (
    '开源多账号管理工具，适合需要集中管理多个账号和配置的进阶用户。',
    '多账号管理工具。'
  )
  AND deleted_at IS NULL;

UPDATE ai_tools
SET description = '把 Claude、Gemini、Grok、DeepSeek 等模型接入 Codex，用一个工作流完成多模型切换和调用。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://github.com/lidge-jun/opencodex'
  AND description IN (
    '可将 Claude、Gemini、Grok、DeepSeek、Ollama 等模型接入 Codex CLI、App、SDK 和 Claude Code。',
    '多模型接入工具。'
  )
  AND deleted_at IS NULL;

UPDATE ai_tools
SET description = '一键查看当前 IP、地区和网络连通情况，适合排查机场、代理或接口访问异常，打开即可检测。',
    updated_at = NOW()
WHERE tenant_id = 1
  AND url = 'https://pingip.cn/'
  AND description IN (
    '用于快速检测当前网络环境和 IP 信息，便于排查连接问题。',
    '查看网络和 IP 信息。'
  )
  AND deleted_at IS NULL;
