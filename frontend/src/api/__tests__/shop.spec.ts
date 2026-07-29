import { afterEach, describe, expect, it, vi } from 'vitest'

describe('resolveShopAssetUrl', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.resetModules()
  })

  it('routes uploaded shop assets to the API origin on the production site', async () => {
    vi.stubGlobal('window', {
      location: { hostname: '3api.shop' },
    })
    const { resolveShopAssetUrl } = await import('@/api/shop')

    expect(resolveShopAssetUrl('/api/v1/shop/assets/demo.png')).toBe(
      'https://api.3api.shop/api/v1/shop/assets/demo.png',
    )
    expect(resolveShopAssetUrl('/shop/assets/demo.png')).toBe(
      'https://api.3api.shop/api/v1/shop/assets/demo.png',
    )
  })

  it('keeps external and inline image URLs unchanged', async () => {
    const { resolveShopAssetUrl } = await import('@/api/shop')

    expect(resolveShopAssetUrl('https://cdn.example.com/item.png')).toBe('https://cdn.example.com/item.png')
    expect(resolveShopAssetUrl('data:image/png;base64,abc')).toBe('data:image/png;base64,abc')
    expect(resolveShopAssetUrl('')).toBe('')
  })
})
