import { describe, expect, it } from 'vitest'
import { DEFAULT_SITE_SUBTITLE, normalizeSiteName, normalizeSiteSubtitle, PRODUCT_NAME } from '@/constants/brand'

describe('public brand defaults', () => {
  it('normalizes the historical site name without changing custom names', () => {
    expect(normalizeSiteName('Sub2API')).toBe(PRODUCT_NAME)
    expect(normalizeSiteName('')).toBe(PRODUCT_NAME)
    expect(normalizeSiteName('My API')).toBe('My API')
  })

  it('normalizes the historical subtitle', () => {
    expect(normalizeSiteSubtitle('Subscription to API Conversion Platform')).toBe(DEFAULT_SITE_SUBTITLE)
    expect(normalizeSiteSubtitle('')).toBe('')
    expect(normalizeSiteSubtitle('统一模型接入')).toBe('统一模型接入')
  })
})
