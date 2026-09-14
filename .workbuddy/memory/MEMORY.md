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

## 返利体系（至关重要，两轨且共用一个钱包）

- **`compute_company` 现为「推广计划」，只剩一条返佣来源：商城订单**（2026-09-13/14 定稿）。旧的「API 余额充值返点」路径已停发（余额充值不返佣），`177` 的 5 层 / `188` 的 T0~T3 档位全部废弃（代码删除 + 迁移 212 打 DEPRECATED，历史数据保留）。
- **商城返利**：`shop_service.go` `issueCommissionTx` 只发直接邀请人，金额 = **外部实付金额 × `commission_bps` ÷ 10000**，线上 12 SKU 实测 1%–10%。全额余额抵扣订单不返佣。
- 返点都写 `distribution_cash_wallets`（人民币），共用冻结时长与支付宝提现口（最低 ¥100）；钱包视图按 `distribution_wallet_ledger.source_type` 区分来源。
- **结构性红线**：现货类商城（GPT 代充/成品号）毛利薄，**不能照搬算力公司 20–40% 的拨出**，否则卖一单亏一单；能承受高拨出的只有预付款型的 API 充值。
- 重构方向已定（2026-09-11，待用户拍板）：**方案 A 解耦** = 商城只做「一级推荐 + 拼车团内生裂变」，算力公司参数不动但限定 API 充值路径、对纯商城用户不展示。
- **2026-09-13 拍板：算力公司 → 「推广计划」，单层**（`promotionMaxDepth = 1`，只发直接邀请人；删掉全部下钻），关系表只写 `depth=1`，历史 2~5 层保留追溯但不再结算。
- **2026-09-13 进一步拍板：彻底删掉 T0/T1/T2 档位（迁移 212 + 代码全删）**。返点比例**唯一来源 = 商品上架时的 `shop_products.commission_bps`**（逐商品设，利润高的商品比例单独定），**没有档位、没有升级、没有业绩门槛**。后台「档位策略表 / 手工指定档位」已下线，改为「推广员列表（邀请人数 / 团队业绩 / 钱包余额）」。
- **2026-09-14 拍板：两种钱严格分离**。① `distribution_cash_wallets`（**人民币返点**）：可提现到支付宝、**可在商城下单时抵扣**、可兑换成 API 平台额度。② `user_credit_accounts`（**API 平台额度，美元**）：只能调 API，不可提现、不可反向兑换、不可抵扣商城订单。兑换入口保留，但 UI 严格分区（`DistributionView.vue` 提现 tab 顶部有「两种余额，用途不同」对比条）。
- **返佣基数 = 商城订单的外部实付金额**（`shop_orders.payable_cny_minor`），**不是商品原价**；**余额充值不返佣**。用返点抵扣掉的那部分再计一次返点 = 无锚增发，故**全额抵扣订单不产生任何返佣**。
- **口径一致性（2026-09-14）**：对外展示的「业绩 / 消费额」也必须取 `payable_cny_minor`。`distribution_service.go` 的 Dashboard 邀请业绩 / Analytics 趋势 spend / Tree 成员业绩 / AdminListMembers 业绩，四处都已是 `SUM(o.payable_cny_minor)`。**新增任何按订单金额汇总的指标都要用 payable，不要用 `snapshot_price_cny_minor`。**
- **佣金明细读侧必须合并两张表（2026-09-14）**：`distribution_commissions`（历史充值返佣，**已无写入方**）与 `shop_commission_records`（当前唯一在写入的佣金）。只读前者 ⇒ 用户端明细与后台佣金列表双双为空。`Ledger()` / `AdminListCommissions()` 已改成 `WITH merged AS (... UNION ALL ...)`，并给 `DistributionCommission` 加了 `source` 字段。
  - ⚠️ **两张表自增 id 会重号**，前端列表 `:key` 必须用 `${item.source ?? 'distribution'}-${item.id}`（`DistributionView.vue` / `FinanceOperationsView.vue` 已改）。
  - 商城侧映射：`shop_order_id`→source_order_id、`buyer_user_id`→source_user_id、`depth` 恒 1（单层）、`tier` 恒 0（无档位）。

## 迁移编写红线（211 踩过）

- **新增「发布新配置版本」式迁移，版本号必须 `current_config_version + 1`，禁止硬编码。** 217/211 这类文件把新档位写进 `config_version = 2`，而版本 2 早被 `188_compute_company_t0.sql` 占用 ⇒ `ON CONFLICT DO NOTHING` 全静默 + `WHERE current_config_version < 2` 不命中 ⇒ **整份迁移空转、无任何报错**。正确写法见 188：`DO $$ ... next_version := previous_version + 1 ... $$`。
- 已有版本号占用表：177 → v1（T1~T3），188 → v2（T0~T3）。**211 → v3**；212（删档位）**非破坏性、不新增版本**（只把 `tier_override` 置 NULL / `current_tier` 置 0 + 打 DEPRECATED 注释，不 DROP 列/表）。档位既已废弃，后续迁移**不需要再递增版本号**，但若要再写版本式迁移，下一版应为 **v4**。
- 编写「不存在则加」的迁移时不能只靠 `ON CONFLICT DO NOTHING`：它会**静默**吞掉版本撞车。
- 分销档位校验已从 `validateDistributionPolicy` 中**删除**（不再有 `promotionTierCount` / 4 档断言）。

## 商城抵扣不变量（213 引入，极易踩）

- 恒等式：**`shop_orders.payable_cny_minor = snapshot_price_cny_minor - wallet_applied_cny_minor`**，由 CHECK 约束强制。
- ⚠️ **任何 `INSERT INTO shop_orders` 都必须显式写 `payable_cny_minor`**，否则默认 0 直接违反 CHECK。已经踩过一次：`shop_guest_service.go` 免登录下单漏写 → 游客买任何商品都失败。
- 抵扣走 `distribution_wallet_ledger`，action `shop_wallet_deduction` / `shop_wallet_deduction_refund`，幂等键 `shop:order:{id}:wallet-deduction` / `:wallet-refund`；钱包行 `FOR UPDATE` 锁，按 `min(可用, 应付)` 抵扣。
- 订单取消/超时/失败必须退回抵扣；全额抵扣订单（无外部支付单）交付失败也必须退回，否则「余额已扣、货没发」。
- `completeLinkedPaymentOrder` **只写不提交**，由调用方统一 Commit——内部再 Commit 会得到 `sql.ErrTxDone`（踩过）。
- 接口：`POST /shop/orders` 的 `use_wallet`、`GET /shop/wallet-balance`。`payment_type` 已放开为可选（全额抵扣无支付环节），是否必填在 service 里按应付金额判断。
- `distribution_tier_configs` 的 `tier` CHECK 已被 188 改成 `BETWEEN 0 AND 3`；`threshold_cny_minor >= 0`。
- 迁移回归测试里要断言「某段 SQL 不存在」时，用 `normalizedMigrationCode`（先剥 `--` 注释）而不是 `normalizedMigration`，否则注释里的示例文本会让 `NotContains` 自伤。

## 财务门 financialgate（生产门，不可删）

- `backend/cmd/financialgate` 被 `deploy/production-preflight.sh` 调用（**不带 `-run-scenarios`**），失败会**阻断生产发布** ⇒ 只能对齐，不能删。
- `requiredFinancialMigrations` 必须随财务模型同步，现含：175–180、188–191、194–197、**205、211、212、213**。漏加 = 新模型未落库也能过门。
- 对账检查（`reconciliationChecks`，全部必须为 0）：credit_bucket_balance_mismatch、migration_audit_unreconciled、open_reconciliation_issues、voucher_without_ledger、negative_distribution_wallet、**shop_order_deduction_identity**、**unrefunded_shop_wallet_deduction**、**partially_deducted_commission_base_mismatch**、reversed_recharge_without_single_reversal、duplicate_distribution_commission、invalid_distribution_relation、negative_wholesale_wallet、negative_partner_wallet。
- `-run-scenarios` 是**破坏性**的（要求空库），只在一次性库上跑。三个历史坑（已修）：闸门把 `nologin:` 系统账号算作非空；系统账号无信用桶导致开计划失败；邀请链需覆盖全部用户才有佣金。
- 本机验证配方（真库）：
  ```sh
  docker run -d --name fgate-pg -e POSTGRES_USER=gate -e POSTGRES_PASSWORD=gate \
    -e POSTGRES_DB=sub2api_gate -p 55432:5432 postgres:18-alpine
  cd backend && go build -o /tmp/financialgate ./cmd/financialgate
  /tmp/financialgate -database-url "postgres://gate:gate@127.0.0.1:55432/sub2api_gate?sslmode=disable" -run-scenarios
  docker rm -f fgate-pg
  ```
  （`validateTarget` 只允许 loopback，127.0.0.1 天然通过；跑完记得删容器。）

## 本机工具坑（会反复踩）

- **`grep` 不支持 `\|` 交替**：`grep "a\|b"` 恒不匹配。一律用 `grep -E "a|b"`。
- ⚠️ **同一轮里对同一个文件连发多个 Edit，会静默丢掉其中一部分**（工具仍报 success）。改完一个文件务必用 Grep / `typecheck` 复核落点；宁可一个文件一轮只改一处。
- **zsh 下 `--include=*.vue` 会报 `no matches found`**：必须给通配符加引号，或改用 Grep 工具。
- `pnpm` 不在 PATH 时用 `corepack pnpm`。
- HTTP 200 ≠ 文件存在：`3api.shop` 缺失的静态文件会被 SPA 回落成 **200 + text/html**，复核静态资源必须校验 `content_type`。

## 上游关系（Wei-Shaw/sub2api）——长期事实

- 本项目是 `github.com/Wei-Shaw/sub2api` 的**业务重度定制分叉**（3api.shop：推广计划 + 商城 + 返点钱包 + 中转站）。
- 远端：`origin = https://github.com/tii-mom/3.git`；`upstream = https://github.com/Wei-Shaw/sub2api.git`。
- **分叉基线 = 上游 v0.1.152（`fc9b48910`，2026-07-13，PR #4125）**。此后本地自研 150 提交。
- 上游最新 = **v0.2.4**（`bdb42e22f`，2026-09-09），落差 **2316 提交 / 36 个 tag / 跨 0.1.x→0.2.x 大版本**。
- **不可直接 merge 升级**：`git merge-tree` 试合并 = **498 文件冲突**；文件级重叠 1018/1347（76%）；
  上游改了 `ent/schema/{user,group,usage_log,subscription_plan,proxy}.go` 等自研依赖的核心表。
- 升级只走**选择性 cherry-pick**（安全/支付/稳定性补丁），不做全量 merge；任何同步前必须先提交工作区。
