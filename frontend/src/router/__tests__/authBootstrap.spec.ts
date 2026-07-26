import { describe, expect, it } from 'vitest'
import { shouldBootstrapAuthForRoute } from '../authBootstrap'

describe('shouldBootstrapAuthForRoute', () => {
  it.each([
    ['/'],
    ['/home'],
    ['/api-relay'],
    ['/openai-api'],
    ['/codex-api'],
    ['/full-api'],
    ['/token-guide'],
    ['/chatgpt-plus-vs-api'],
    ['/compute-company'],
  ])('does not restore auth on public marketing route %s', (path) => {
    expect(shouldBootstrapAuthForRoute(path, { requiresAuth: false }, false)).toBe(false)
  })

  it.each([
    ['/dashboard', {}],
    ['/admin/finance', { requiresAdmin: true }],
    ['/distribution', { requiresAuth: true }],
    ['/login', { requiresAuth: false }],
    ['/register', { requiresAuth: false }],
    ['/setup', { requiresAuth: false }],
  ])('restores auth when identity can change route handling for %s', (path, meta) => {
    expect(shouldBootstrapAuthForRoute(path, meta, false)).toBe(true)
  })

  it('restores auth for any route while backend mode is enabled', () => {
    expect(shouldBootstrapAuthForRoute('/', { requiresAuth: false }, true)).toBe(true)
  })
})
