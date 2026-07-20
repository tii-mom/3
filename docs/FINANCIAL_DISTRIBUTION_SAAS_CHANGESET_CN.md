# 3API 财务、分销与 SaaS 变更清单

## 交付分组

建议按以下顺序审阅和提交。当前工作区包含用户原有改动，未在未经确认的情况下自动暂存。

1. 财务底座
   - `176_credit_accounts_and_vouchers.sql`
   - `179_financial_runtime_controls.sql`
   - `internal/pkg/creditctx`
   - `internal/pkg/creditledger`
   - 用户余额、兑换和用量扣费接入点
2. 用户额度兑换码
   - `internal/service/voucher_service.go`
   - 用户及管理员 voucher handler、route、API 和页面
3. 算力公司部门组织与提现
   - `177_distribution_program.sql`
   - `180_distribution_reversals.sql`
   - `internal/service/distribution_service.go`
   - 邀请关系闭包、支付完成接入、用户及管理员页面
4. 白牌 SaaS MVP
   - `178_saas_control_plane.sql`
   - `internal/service/saas_service.go`
   - 批发 Key、实时余额计费、域名、套餐、部署任务和合作提现页面
5. 运行保障
   - `internal/service/financial_maintenance_service.go`
   - outbox、迁移契约测试、上线手册和本清单

## 运行开关

| 开关 | 默认值 | 作用 |
| --- | --- | --- |
| `credit_bucket_enforce_enabled` | `false` | `false` 时旧余额权威且额度桶影子写入，`true` 时额度桶接管兼容余额 |
| `balance_voucher_enabled` | `false` | 控制用户创建额度兑换码 |
| `distribution_programs.enabled` | `false` | 控制算力公司首充奖励、团队业绩和部门绩效 |
| `distribution_programs.stack_with_legacy` | `false` | 是否同时发放旧单层返利 |
| `saas_control_plane_enabled` | `false` | 控制租户创建和批发 Key 访问 |

## 新增管理 API

- `GET/PUT /api/v1/admin/distribution/financial-runtime`
- `POST /api/v1/admin/distribution/config/versions`
- `POST /api/v1/admin/distribution/recharge-events/:id/reverse`
- `/api/v1/admin/distribution/*`
- `/api/v1/admin/vouchers/*`
- `/api/v1/admin/saas/*`

资金开关、政策版本、租户创建、批发加款、收款详情和提现状态操作均要求管理员 TOTP。

## 上线阻断项

- 尚未在真实 PostgreSQL 执行 `175-180` 迁移和二次幂等验证。
- 尚未基于生产副本生成余额与关系回填对账报告。
- 尚未执行真实支付回调、并发提现、拒付冲正和批发余额耗尽端到端测试。
- 尚未完成备份恢复、压力测试和 2-3 个租户的 7 天灰度。
- 官方上游差异必须在独立集成分支处理，不能直接合并到当前脏工作区。
