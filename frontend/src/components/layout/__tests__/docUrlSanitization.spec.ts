import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const headerSource = readFileSync(resolve(dir, '../AppHeader.vue'), 'utf8')
const homeViewSource = readFileSync(resolve(dir, '../../../views/HomeView.vue'), 'utf8')
const keyUsageViewSource = readFileSync(resolve(dir, '../../../views/KeyUsageView.vue'), 'utf8')

describe('doc_url sanitization', () => {
  it('AppHeader imports sanitizeUrl', () => {
    expect(headerSource).toContain("import { sanitizeUrl } from '@/utils/url'")
  })

  it('AppHeader applies sanitizeUrl to docUrl', () => {
    expect(headerSource).toContain('sanitizeUrl(appStore.docUrl)')
  })

  // 首页已由「API 业务首页」改造为「GPT 充值销售页」，不再渲染文档链接。
  // 这里保留防回归约束：若将来重新引入 doc_url，必须经过 sanitizeUrl。
  it('HomeView: 若引入 doc_url 则必须经过 sanitizeUrl 处理', () => {
    if (homeViewSource.includes('doc_url') || homeViewSource.includes('docUrl')) {
      expect(homeViewSource).toContain("import { sanitizeUrl } from '@/utils/url'")
      expect(homeViewSource).toMatch(/sanitizeUrl\([^)]*doc_?[Uu]rl/)
    } else {
      expect(homeViewSource).not.toContain('doc_url')
    }
  })

  it('KeyUsageView imports sanitizeUrl', () => {
    expect(keyUsageViewSource).toContain("import { sanitizeUrl } from '@/utils/url'")
  })

  it('KeyUsageView applies sanitizeUrl to docUrl', () => {
    expect(keyUsageViewSource).toContain('sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl')
  })
})
