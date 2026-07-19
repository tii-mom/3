# 3API 财务、分销与 SaaS 上线运行手册

## 1. 发布原则

- 所有新功能保持关闭，先完成迁移、回填、对账和影子写入观察。
- PostgreSQL 是资金、佣金、兑换码和 outbox 的唯一事实源；Redis 仅承载缓存、限流和事件唤醒。
- 不对生产库执行向下迁移。出现问题时先关闭功能开关并保留账本，再通过冲正记录修复。
- 首批只允许 2 至 3 个白牌租户，使用独立实例、独立数据库或 Schema、独立 Redis 命名空间。

## 2. 发布前门禁

1. 固定代码提交和不可变镜像摘要，记录 Go、Node、PostgreSQL、Redis 版本。
2. 对生产数据库做物理备份和逻辑备份，并在隔离环境完成恢复演练。
3. 在生产副本依次执行 `175`、`176`、`177`、`178`、`179`、`180` 迁移；重新运行迁移程序，确认第二次无新增变更。
4. 执行后端 `go test ./...`，前端 `pnpm typecheck && pnpm lint:check && pnpm test:run && pnpm build`。
5. 确认管理员 TOTP、支付宝字段加密密钥、Redis 持久化和告警均可用。
6. 确认官方上游差异已在独立集成分支审阅；禁止在脏工作区直接合并上游。

## 3. 迁移后对账

优先使用内置门禁执行迁移双跑、校验和和下列对账。默认只允许本机数据库；对隔离的远程预发布副本执行时，必须显式设置 `FINANCIAL_GATE_ALLOW_NON_LOCAL=true`。

```bash
cd backend
DATABASE_URL='postgres://user:password@127.0.0.1:5432/threeapi_gate?sslmode=disable' \
  go run ./cmd/financialgate
```

全新空白本地库可以追加 `-run-scenarios`，该模式会写入隔离夹具并验证重复充值、首充赠额、兑换码并发兑换和 SaaS 批发重复计费。已有用户时命令会拒绝运行。

```bash
DATABASE_URL='postgres://user:password@127.0.0.1:5432/threeapi_gate?sslmode=disable' \
  go run ./cmd/financialgate -run-scenarios

# 在已经完成夹具验收的隔离库执行分销压力门禁
DATABASE_URL='postgres://user:password@127.0.0.1:5432/threeapi_gate?sslmode=disable' \
  go run ./cmd/financialgate -stress-orders 1000 -stress-concurrency 32 -timeout 10m
```

以下查询必须返回零行或零值；非零记录进入人工复核，不能自动删除或覆盖。

```sql
-- 兼容余额与额度桶守恒
SELECT u.id, u.balance,
       a.transferable_credit + a.non_transferable_credit - a.debt AS bucket_balance
FROM users u
JOIN user_credit_accounts a ON a.user_id = u.id
WHERE u.balance <> a.transferable_credit + a.non_transferable_credit - a.debt;

-- 回填审计异常
SELECT * FROM financial_balance_migration_audit
WHERE reconciliation_status <> 'RECONCILED';

-- 影子写入期间产生的兼容余额与额度桶差异
SELECT * FROM financial_reconciliation_issues
WHERE status = 'OPEN';

-- 兑换码创建、退款、兑换状态异常
SELECT v.id, v.status, COUNT(l.id) AS ledger_entries
FROM balance_vouchers v
LEFT JOIN balance_voucher_ledger l ON l.voucher_id = v.id
GROUP BY v.id, v.status
HAVING COUNT(l.id) = 0;

-- 分销钱包不可出现负桶
SELECT * FROM distribution_cash_wallets
WHERE available_cny_minor < 0 OR frozen_cny_minor < 0
   OR withdrawing_cny_minor < 0 OR debt_cny_minor < 0;

-- 已冲正充值必须有且仅有一条不可变冲正事件
SELECT e.id, COUNT(r.id)
FROM distribution_recharge_events e
LEFT JOIN distribution_reversal_events r ON r.recharge_event_id = e.id
WHERE e.status = 'REVERSED'
GROUP BY e.id
HAVING COUNT(r.id) <> 1;

-- 同一订单、受益人、真实层级只能有一条佣金
SELECT program_id, source_order_id, beneficiary_user_id, depth, COUNT(*)
FROM distribution_commissions
GROUP BY program_id, source_order_id, beneficiary_user_id, depth
HAVING COUNT(*) > 1;

-- SaaS 批发钱包不可透支
SELECT * FROM saas_wholesale_wallets WHERE balance_usd < 0;

-- outbox 积压与失败重试
SELECT status, COUNT(*), MAX(attempts) FROM financial_outbox_events GROUP BY status;
```

## 4. 灰度顺序

1. 部署代码但保持 `credit_bucket_enforce_enabled=false`、`balance_voucher_enabled=false`、`distribution_enabled=false`、`saas_control_plane_enabled=false`。
2. 在额度桶影子模式观察至少 24 小时，每小时对比 `users.balance` 与额度桶，并确认 `financial_reconciliation_issues` 没有开放问题。
3. 管理后台通过 TOTP 启用 `credit_bucket_enforce_enabled`，验证非可转让优先、可转让其次、债务最后，观察至少 24 小时。
4. 对内部测试用户启用兑换码，完整走创建、风控锁定、解锁、兑换、撤销和到期退款。
5. 启用算力公司但保持 `stack_with_legacy=false`，用五层测试关系验证跨档订单和 7 天冻结。
6. 启用 SaaS 控制面，只创建试点租户；域名 TXT 校验成功后再配置 TLS 和流量。
7. 逐租户验证批发余额耗尽返回 `TENANT_WHOLESALE_BALANCE_INSUFFICIENT`，总开关关闭返回 `SAAS_CONTROL_PLANE_DISABLED`。

## 5. 功能启用检查

### 兑换码

- 创建仅扣可转让额度，金额为面值加 8% 手续费。
- 完整码只在创建响应出现，数据库只有 SHA-256 哈希和末四位。
- 兑换所得进入不可转让额度，不增加 `total_recharged`，不触发任何返利。
- 撤销和到期原事务退回面值与手续费；并发操作只有一个状态转换成功。

### 五级分销与提现

- 邀请关系无环且永久锁定，层级固定、不压缩。
- 团队业绩仅包含余额充值的人民币实付本金，支付附加费不计入。
- `1000/10000/100000 CNY` 边界订单按订单完成后的团队业绩立即升档。
- 每次调整冻结期、提现规则、首充奖励或档位比例时创建新配置版本；历史订单继续引用原版本。
- 提现只来自现金佣金钱包，状态仅允许 `SUBMITTED -> APPROVED -> PAID` 或 `REJECTED`。
- 管理员查看完整支付宝信息和改变提现状态必须通过 TOTP。
- 支付拒付只通过充值事件的“拒付冲正”执行；必须填写原因并通过 TOTP，重复提交返回同一冲正记录。
- 冲正后撤回充值本金、首充赠额、五层业绩、新五级佣金及同订单的旧邀请返利；已提现佣金形成债务，后续佣金解冻时优先偿还。

### 白牌 SaaS

- 租户支付配置和数据库密码为密文，上游凭证只存在 Core。
- 租户终端支付直接进入租户账户；平台只扣租户预充值批发钱包。
- 批发钱包不可提现，合作佣金进入独立现金钱包。
- 每个试点租户设置 CPU、内存、并发、月度额度、日志标签和告警。

## 6. 回滚与事故处理

1. 立即关闭对应功能开关；额度异常时先关闭 `credit_bucket_enforce_enabled` 回到影子模式，SaaS 开关关闭会使批发 Key 认证缓存失效并阻断新请求。
2. 保留数据库和 outbox，不删除失败事件，不手工修改不可变账本。
3. 暂停支付入口后生成余额桶、佣金钱包、兑换码和批发钱包快照。
4. 以原业务事务为单位写冲正记录；禁止直接改历史流水。
5. 恢复旧应用版本前确认旧版本不会绕过额度桶继续写 `users.balance`。
6. 事故关闭后重新执行全部对账 SQL、备份恢复抽检和端到端冒烟测试。

## 7. 试点验收

- 连续 7 天无账本不守恒、重复佣金、重复兑换、重复提现或批发透支。
- Redis 中断期间 outbox 保持 pending，恢复后可重投；消费者按幂等键去重。
- 单租户实例、数据库或支付配置故障不影响 Core 和其他租户。
- 记录试点租户的毛利、上游成本、支付成本、佣金、赠额和退款，确认商业模型后再扩容。

## 8. 2026-07-18 隔离预发布验收记录

- PostgreSQL `16.9` 隔离实例全量应用 `220` 个迁移；第二次执行仍为 `220`，`175-180` 嵌入文件校验和一致。
- 真实 PostgreSQL 验收发现并修复两处仅在驱动执行时出现的 SQL 类型推断问题：分销档位 `SMALLINT` 更新和充值 outbox JSON 参数。
- 12 路重复充值回调只产生 1 个充值事件、5 笔佣金，一级至五级比例为 `10/4/3/2/1%`，总佣金 `200.00 CNY`。
- `10,000 USD` 首充本金产生 `1,000 USD` 不可转让赠额；12 路兑换码并发兑换仅 1 次成功，撤销完整退回 `100 USD + 8 USD` 手续费。
- 12 路相同 SaaS 批发请求仅扣款 1 次；`1,000 USD - 1.25 USD = 998.75 USD`。
- 分销压力门禁：1,000 个独立订单、32 worker、1,000 个事件、5,000 笔佣金，耗时 `4.327s`，约 `231.1 orders/s`；门禁十项对账均为零。
- 冷物理备份归档大小 `12 MB`，SHA-256 `5537bba9ba5e22f12ec79de349ed914056dad0d53b681d33a867545856399a8d`；恢复到独立端口后迁移、校验和、资金和关系对账全部通过。
- 本机嵌入式 PostgreSQL 不含 `pg_dump/psql`。生产根 `Dockerfile` 已内置同版本工具，但仍须在最终 Docker 预发布环境完成一次逻辑 `pg_dump -> psql --single-transaction` 恢复演练，才可关闭备份恢复门禁。

## 9. 2026-07-19 Docker 预发布验收记录

- Docker Desktop `4.82.0`、Engine `29.6.1`、Compose `5.3.0`、ARM64 环境完成最终根 `Dockerfile` 构建。生产镜像为 `3api-financial-gate:20260719`，摘要 `sha256:1fc2b300c324e6c3d3e4ab093f4f482320fb44329ca7569763da6a2a3f92747f`，大小 `37,175,211` 字节。
- Docker 构建发现前端在固定 `1 GiB` Node 堆限制下 OOM；构建堆上限调整为 `2 GiB` 后成功。该内存只用于 Vite/Rollup 构建，不增加最终运行容器的内存占用。
- 最终镜像包含 `pg_dump/psql 18.4`。使用镜像内工具生成 `474,711` 字节逻辑备份，SHA-256 `f14883723207f0d9275847e4e24c66a242c77f0af0e5c54bd789ffe6b93de592`，通过 `psql --single-transaction` 恢复到全新数据库；恢复库 `220` 个迁移及十项财务对账全部通过。至此逻辑恢复门禁关闭。
- 应用、PostgreSQL `18.4`、Redis `8` 容器健康；应用主进程以 UID/GID `1000` 运行。`/health` 同时报告 PostgreSQL 和 Redis 为 `ok`。
- Testcontainers PostgreSQL/Redis 集成测试通过。Redis 停机时 `/health` 返回 `503`，outbox 保持 `pending` 并记录重试；Redis 恢复后事件变为 `processed` 且成功写入 `financial:events` Stream。
- Docker PostgreSQL 18 空库再次完成 12 路充值/兑换/批发幂等场景；500 个分销订单、32 worker 产生 500 个事件和 2,500 笔佣金，对账为零。批发余额不足请求被原子拒绝，钱包和幂等占位均未改变。
- 第二白牌实例在 `512 MiB / 1 CPU / 256 PID` 限制下健康运行，使用独立数据库和 Redis DB；停止该实例不影响主实例，恢复后自身健康检查通过。
- 最终回归通过：`go test ./...`、`go test -tags=unit ./internal/service`、`pnpm typecheck`、`pnpm lint:check`、`pnpm test:run`、`pnpm build`。
- 官方 `Wei-Shaw/sub2api` 主分支当前为 `0.1.161`（`d4b9797ff72024960a035cf22fdd8f213e149169`），本项目基线为 `0.1.152`，且内部仓库不是 GitHub fork、无可直接比较的共同提交；文件级差异约 `892` 项。必须在本功能分支固定提交后创建独立上游集成分支逐项合并和复测，禁止直接覆盖当前已验收代码。

## 10. 2026-07-19 官方 0.1.161 集成验收记录

- 以官方 `v0.1.152` 为三方合并基准，将官方 `v0.1.161` 功能提交 `19149ca196eeae4a4482e5299dc6fa4ba0b06c8c` 和主分支版本同步提交 `d4b9797ff72024960a035cf22fdd8f213e149169` 集成到独立分支；财务、批发、鉴权和部署边界保留 3API 定制实现。
- 升级后 PostgreSQL `18.4` 空库迁移总数为 `234`，第二次执行仍为 `234`；十项财务对账均为零。500 个订单、32 worker 产生 500 个充值事件和 2,500 笔佣金，耗时 `4.170s`，约 `119.9 orders/s`。
- 逻辑备份通过 `pg_restore --single-transaction` 恢复到独立数据库，恢复库保持 `234` 个迁移，十项财务对账均为零。
- 最终回归通过：`go test ./...`、`go test -tags=unit ./internal/service -count=1`、Testcontainers 仓储集成测试、`pnpm typecheck`、`pnpm lint:check`、182 个前端测试文件共 1,237 项测试以及 `pnpm build`。
- 生产镜像 `3api-financial-gate:upstream-0161` 摘要为 `sha256:452809d6f839b2ab8cf25b9e5d02d2acdfab9f0aed21d7dbee50831addf2154a`，大小 `38,136,988` 字节，内嵌版本 `0.1.161`。应用主进程以 `sub2api` 用户运行，`/health` 同时报 PostgreSQL 和 Redis 为 `ok`。
- 根 Dockerfile 已移除不必要的远程 Dockerfile frontend 语法声明，避免受限网络环境在业务构建开始前阻塞；Docker Desktop 自带 BuildKit 可直接解析现有 cache mount 指令。

## 11. 生产备份与隔离恢复预检

- 已合并基线提交为 `6c63dbb299326aa9ab076c621b2e5fd7c5439f32`；首个正式候选镜像为 `ghcr.io/tii-mom/3@sha256:9896d780f463915ee265dd51cee748299be8b92a3af581cf73dcb7c128b53c2b`。
- 后续候选镜像必须包含 `/app/financialgate`，并以新构建产生的 digest 为准。禁止使用 tag 或 `latest` 触发生产预检和部署。
- 手动运行 GitHub Actions 工作流 `Production Backup and Isolated Restore Preflight`，输入候选 digest 和确认短语 `BACKUP_AND_RESTORE_ONLY`。
- 预检对生产 PostgreSQL 只执行版本、用户数、迁移数、数据库大小查询和 `pg_dump`。它不会执行迁移、场景夹具、回填、删除或更新。
- 备份恢复到随机命名的临时 PostgreSQL 18 容器；候选镜像的迁移双跑、功能默认值和十项财务门禁只在该恢复副本执行。
- 备份文件、SHA-256 和财务门禁 JSON 保存在服务器 `/opt/sub2api-deploy/backups/preflight/`。临时容器和网络自动删除，备份及报告不自动删除。
- 只有预检成功且人工核对报告后，才允许手动运行生产部署工作流。部署后所有新功能仍保持关闭。
- 手动部署必须输入同一个已预检 digest 和确认短语 `DEPLOY_PREFLIGHTED_DIGEST`；工作流不会重新构建镜像，避免实际部署物与预检物不一致。
- 2026-07-19 本地端到端脚本演练通过：对已有 1 个用户、220 个迁移的 PostgreSQL 18.1 验收库只读备份，恢复副本升级到 234 个迁移并通过十项财务门禁；演练同时验证了环境文件权限拒绝和 archive 恢复的 `--no-owner --no-acl` 兼容处理。
