import { describe, it, expect, afterEach, vi } from 'vitest'

describe('resolveApiClientBaseURL', () => {
  afterEach(() => {
    vi.unstubAllEnvs()
    vi.unstubAllGlobals()
    vi.resetModules()
  })

  it('appends /api/v1 when site setting is origin-only', async () => {
    vi.resetModules()
    const { resolveApiClientBaseURL } = await import('@/api/url')
    expect(resolveApiClientBaseURL('https://api.3api.shop')).toBe(
      'https://api.3api.shop/api/v1',
    )
    expect(resolveApiClientBaseURL('https://api.3api.shop/')).toBe(
      'https://api.3api.shop/api/v1',
    )
  })

  it('keeps an already-versioned absolute base', async () => {
    vi.resetModules()
    const { resolveApiClientBaseURL } = await import('@/api/url')
    expect(resolveApiClientBaseURL('https://api.3api.shop/api/v1')).toBe(
      'https://api.3api.shop/api/v1',
    )
  })

  it('appends /api/v1 under a custom path prefix', async () => {
    vi.resetModules()
    const { resolveApiClientBaseURL } = await import('@/api/url')
    expect(resolveApiClientBaseURL('https://example.com/sub2api')).toBe(
      'https://example.com/sub2api/api/v1',
    )
  })

  it('falls back to host-based API URL when empty', async () => {
    vi.resetModules()
    vi.stubGlobal('window', {
      location: { hostname: '3api.shop' },
    })
    const { resolveApiClientBaseURL } = await import('@/api/url')
    expect(resolveApiClientBaseURL('')).toBe('https://api.3api.shop/api/v1')
    expect(resolveApiClientBaseURL(null)).toBe('https://api.3api.shop/api/v1')
  })
})
