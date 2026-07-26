import { describe, expect, it } from 'vitest'
import { hasAuthBootstrapStorageHint, shouldBootstrapAuthForRoute } from '../authBootstrap'

function storage(values: Record<string, string> = {}) {
  return {
    getItem(key: string) {
      return values[key] ?? null
    },
  } as Storage
}

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

describe('hasAuthBootstrapStorageHint', () => {
  it('does not restore auth for an anonymous public browser session', () => {
    expect(hasAuthBootstrapStorageHint(storage(), storage())).toBe(false)
  })

  it.each([
    ['local auth user', storage({ auth_user: '{"id":1}' }), storage()],
    ['local legacy refresh token', storage({ refresh_token: 'legacy' }), storage()],
    ['local token expiry', storage({ token_expires_at: '1000' }), storage()],
    ['session legacy refresh token', storage(), storage({ refresh_token: 'legacy' })],
    ['session token expiry', storage(), storage({ token_expires_at: '1000' })],
  ])('restores auth when %s exists', (_label, local, session) => {
    expect(hasAuthBootstrapStorageHint(local, session)).toBe(true)
  })
})
