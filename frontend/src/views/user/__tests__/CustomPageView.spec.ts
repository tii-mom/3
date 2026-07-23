import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const dir = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(dir, '../CustomPageView.vue'), 'utf8')

describe('CustomPageView embed security', () => {
  it('isolates external embeds and suppresses referrer forwarding', () => {
    expect(source).toContain('sandbox="allow-scripts allow-forms allow-popups"')
    expect(source).toContain('referrerpolicy="no-referrer"')
  })
})
