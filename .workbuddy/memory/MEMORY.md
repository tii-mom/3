# 3api 项目长期约定（跨会话）

> 详细操作手册见 skill：`3api-deploy-release`（发布/CI 安全门）、`local-ui-preview-screenshot`（预览与线上复核）。

## 部署拓扑（一句话版）

- 前端 `3api.shop`：push 到 main 且改到 `frontend/**`、`docs/legal/**` 或 `pages-deploy.yml` → `pages-deploy.yml` 自动发。
- 后端镜像：任意 push 触发 `deploy.yml` 的 build-and-push，**只产镜像不上线**。
- 后端生产 `api.3api.shop`：必须手动 `production-preflight.yml` → `deploy.yml`（同一 digest）；`production` 环境无人工审批、5 分钟冷却后自动放行。

## 前端架构约定

- **`document.title` 只允许一个写入者。** `App.vue` 里 `resolveRouteDocumentTitle(...)`（走 `route.meta.titleKey` 本地化）必须先写、且之后不能再被覆盖；`utils/seo.ts` 的 `updateRouteSeo` **只对首页与 SEO 落地页**写 title，其余路由不得用 `route.meta.title` 覆盖。新增路由要给本地化标题就补 `titleKey`。
- **法务文档走「内置 Markdown 登记表」**：`frontend/docs/legal/*.md`（仓库根 `docs/legal/`）被 `LegalDocumentView.vue` 用 `?raw` 导入，经 `BUNDLED_DOCUMENTS`（Record<id,{titleKey,typeKey,zh,en}>）命中即渲染，**不依赖后端 `login_agreement_documents`**（线上那几张表是空的）。新增一篇 = 加 2 个 .md + 登记表一行 + i18n 两键。
- **法务/条款入口只放页脚**（`HomeView.vue` footer），购买界面（`OrderSheet.vue`）不放条款、不要求勾选。登录/注册页的协议勾选由后端 `login_agreement_enabled` 开关门控（线上 false）。
- 站点图标由 `public/logo.svg` 生成（Playwright + 系统 Chrome 截图 → PNG，ICO 为手写容器嵌 PNG），清单见 `frontend/public/`。

## 本机工具坑（会反复踩）

- **`grep` 不支持 `\|` 交替**：`grep "a\|b"` 恒不匹配。一律用 `grep -E "a|b"`。
- **zsh 下 `--include=*.vue` 会报 `no matches found`**：必须给通配符加引号，或改用 Grep 工具。
- `pnpm` 不在 PATH 时用 `corepack pnpm`。
- HTTP 200 ≠ 文件存在：`3api.shop` 缺失的静态文件会被 SPA 回落成 **200 + text/html**，复核静态资源必须校验 `content_type`。
