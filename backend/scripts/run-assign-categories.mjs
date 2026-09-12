// 一次性执行脚本：把 assign-shop-categories.sql 里的 12 条 UPDATE 跑进生产库。
// 设计：① 只执行文件里非注释的 UPDATE 语句，绝不跑其它；② 整批包在事务里，
// 先跑完再 SELECT 校验（品类数 ≥ 2、更新行数 = 12），通过才 COMMIT，否则 ROLLBACK。
// 用法：DATABASE_URL='postgres://...' node run-assign-categories.mjs
// 注意：DATABASE_URL 从环境变量读，本脚本不会把它打印到任何输出。

import { readFileSync } from 'node:fs'
import { Client } from 'pg'

const url = process.env.DATABASE_URL
if (!url) {
  console.error('[!] 未设置 DATABASE_URL 环境变量，退出。')
  process.exit(1)
}

const sqlPath = new URL('./assign-shop-categories.sql', import.meta.url)
const raw = readFileSync(sqlPath, 'utf8')

// 按 ; 切分，去掉每段的 -- 注释行，过滤空段 → 仅剩真正的 UPDATE 语句
const statements = raw
  .split(';')
  .map((block) =>
    block
      .split('\n')
      .filter((line) => !line.trim().startsWith('--'))
      .join('\n')
      .trim()
  )
  .filter((s) => /^update\s/i.test(s))

console.log(`[i] 解析到 ${statements.length} 条 UPDATE 语句，准备在事务中执行。`)

// 生产库通常要求 SSL；若连接串已带 sslmode=disable 则不开，否则放行（rejectUnauthorized=false 以兼容自签/ Hyperdrive）
const needsSsl = !/sslmode=disable/i.test(url)
const client = new Client({
  connectionString: url,
  ssl: needsSsl ? { rejectUnauthorized: false } : false,
})

await client.connect()
try {
  await client.query('BEGIN')
  for (const stmt of statements) {
    const res = await client.query(stmt)
    const n = res.rowCount ?? 0
    const idMatch = stmt.match(/id\s*=\s*(\d+)/i)
    console.log(`  ✓ id=${idMatch ? idMatch[1] : '?'} 影响行数=${n}`)
    if (n !== 1) {
      throw new Error(`id=${idMatch?.[1]} 期望影响 1 行，实际 ${n}（可能 id 不存在或 tenant 不匹配），回滚。`)
    }
  }

  const verify = await client.query(
    "SELECT category, count(*)::int AS cnt FROM shop_products WHERE tenant_id = 1 GROUP BY category ORDER BY category"
  )
  console.log('\n[i] 校验（tenant_id=1 品类分布）：')
  let total = 0
  for (const r of verify.rows) {
    console.log(`  - ${r.category}: ${r.cnt}`)
    total += r.cnt
  }
  const distinct = verify.rows.length
  if (distinct < 2) throw new Error(`品类数=${distinct} < 2，4 列对比仍不生效，回滚。`)
  if (total !== 12) throw new Error(`总商品数=${total} ≠ 12，回滚。`)

  await client.query('COMMIT')
  console.log(`\n[✓] 已提交。${distinct} 个品类 / 共 ${total} 个商品，首页分类导航 + 商城品类 Tab 现在可渲染。`)
} catch (err) {
  await client.query('ROLLBACK')
  console.error(`\n[✗] 已回滚：${err.message}`)
  process.exit(1)
} finally {
  await client.end()
}
