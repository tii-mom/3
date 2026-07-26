# 3API

3API 是一个 AI API 中转与运营管理平台，用于统一接入多模型服务，并提供用户、额度、支付、兑换码、订阅套餐、图片生成桥接、财务运营和 T0-T3 算力公司收益体系。

官网地址：https://3api.shop

## 项目包含什么

- 多模型 API 网关：支持 OpenAI 兼容客户端、Codex、Claude、Gemini、Grok、图片、Embedding 和流式场景。
- 管理后台：用户、分组、上游账号、API Key、支付套餐、财务运营、风控和审计记录。
- 用户控制台：Key 使用、购买、额度兑换/生成、算力公司收益、提现申请和平台余额。
- SEO 公共页面：构建期 HTML 快照、sitemap、robots、canonical，以及登录/后台等私有页面的 `noindex` 控制。
- 发布安全门禁：数据库迁移、财务对账、备份预检、隔离恢复预演和生产发布记录。

## 品牌规则

本仓库维护的所有对外产品名称、页面标题、描述、导航文案和公开文档，统一使用 **3API**。

项目基于上游源码演进而来，因此 Go module、数据库默认名、缓存 key、测试夹具或历史迁移记录中可能保留少量内部兼容名称。不要进行全局替换；只替换会展示给用户、运营人员或公开文档读者的内容。

发布前请运行品牌检查：

```bash
cd frontend
pnpm run brand:check
```

## 开发检查

前端：

```bash
cd frontend
pnpm install
pnpm run typecheck
pnpm run lint:check
pnpm run test:run
pnpm run build
```

后端：

```bash
cd backend
go test ./...
```

仓库格式检查：

```bash
git diff --check
```

## 发布安全

当前线上平台已有真实用户数据，生产发布必须保守处理：

1. 审查最终 diff，确认没有提交密钥、私有配置、证书、数据库备份或 `.env` 文件。
2. 执行前端检查、后端测试、品牌检查和格式检查。
3. 构建不可变发布产物。
4. 创建最新生产数据库备份。
5. 使用备份进行隔离恢复预演，并验证迁移结果。
6. 只部署通过预检的镜像或产物。
7. 执行健康检查、公开配置、首页、SEO、后台、支付、兑换码和财务运营冒烟。
8. 保留备份路径、SHA-256、部署版本、镜像摘要和回滚命令。

正常发布流程中禁止删除或重置生产数据库。

## 文档

- 支付配置：[docs/PAYMENT_CN.md](docs/PAYMENT_CN.md)
- Payment configuration: [docs/PAYMENT.md](docs/PAYMENT.md)
- 图片生成 API：[docs/IMAGE_GENERATION_API.md](docs/IMAGE_GENERATION_API.md)
- 管理端支付集成：[docs/ADMIN_PAYMENT_INTEGRATION_API.md](docs/ADMIN_PAYMENT_INTEGRATION_API.md)
