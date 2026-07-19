import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const dir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(dir, '../../../views/HomeView.vue'), 'utf8')

describe('HomeView custom content security', () => {
  it('sanitizes custom HTML before using v-html', () => {
    expect(source).toContain("import DOMPurify from 'dompurify'")
    expect(source).toContain('v-html="sanitizedHomeContent"')
    expect(source).toContain('FORBID_TAGS')
    expect(source).not.toContain('v-html="homeContent"')
  })

  it('isolates external content and requires HTTPS in production', () => {
    expect(source).toContain('sandbox="allow-scripts allow-forms allow-popups"')
    expect(source).toContain('referrerpolicy="no-referrer"')
    expect(source).toContain("content.startsWith('https://')")
    expect(source).not.toContain('allow-same-origin')
  })
})
