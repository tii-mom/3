import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const dir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(dir, '../../../views/HomeView.vue'), 'utf8')

// 旧首页支持后台自定义 HTML 区块（DOMPurify + sandboxed iframe）。
// 首页改造为销售页后该能力已下线，因此这里断言的是当前实现的安全保证，
// 同时保留防回归：一旦重新引入 v-html / iframe，必须配套净化与沙箱隔离。
describe('HomeView custom content security', () => {
  it('does not render untrusted HTML via v-html', () => {
    expect(source).not.toContain('v-html')
    expect(source).not.toContain('dompurify')
  })

  it('does not embed external content in iframes', () => {
    expect(source).not.toContain('<iframe')
    expect(source).not.toContain('allow-same-origin')
  })
})
