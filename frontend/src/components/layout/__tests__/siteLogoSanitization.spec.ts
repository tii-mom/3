import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const sidebarSource = readFileSync(resolve(dir, '../AppSidebar.vue'), 'utf8')
const homeViewSource = readFileSync(resolve(dir, '../../../views/HomeView.vue'), 'utf8')
const keyUsageViewSource = readFileSync(resolve(dir, '../../../views/KeyUsageView.vue'), 'utf8')

describe('site_logo sanitization', () => {
  it('AppSidebar imports sanitizeUrl and applies it to siteLogo', () => {
    expect(sidebarSource).toContain("import { sanitizeUrl } from '@/utils/url'")
    expect(sidebarSource).toContain('sanitizeUrl(appStore.siteLogo')
  })

  // 销售版首页不再展示站点 Logo；保留防回归：若重新引入必须经 sanitizeUrl。
  it('HomeView: 若引入 site_logo 则必须经过 sanitizeUrl 处理', () => {
    if (homeViewSource.includes('site_logo') || homeViewSource.includes('siteLogo')) {
      expect(homeViewSource).toContain("import { sanitizeUrl } from '@/utils/url'")
      expect(homeViewSource).toMatch(/sanitizeUrl\([^)]*site_?[Ll]ogo/)
    } else {
      expect(homeViewSource).not.toContain('site_logo')
    }
  })

  it('KeyUsageView applies sanitizeUrl to siteLogo', () => {
    expect(keyUsageViewSource).toContain('sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo')
  })

  it('all logo renderers pass allowRelative and allowDataUrl options', () => {
    // 销售版首页已移除 Logo 渲染，仅校验仍然渲染 Logo 的两处
    for (const src of [sidebarSource, keyUsageViewSource]) {
      expect(src).toContain('allowRelative: true')
      expect(src).toContain('allowDataUrl: true')
    }
  })
})
