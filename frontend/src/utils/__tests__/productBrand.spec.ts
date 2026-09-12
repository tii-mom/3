import { describe, expect, it } from 'vitest'
import type { PublicProduct } from '@/api/publicShop'
import { productBrandMeta, resolveBrandDisplay, hasProductImage } from '@/utils/productBrand'

/** 构造最小可测商品，避开无关字段。 */
function makeProduct(overrides: Partial<PublicProduct> = {}): PublicProduct {
  return {
    id: 1,
    name: '测试商品',
    description: '',
    image_url: '',
    product_type: 'account',
    price_cny_minor: 10000,
    original_price_cny_minor: 0,
    sold_count: 0,
    fulfillment_mode: 'one_time',
    delivery_form_hint: '',
    badge_text: '',
    spec_label: '',
    highlight: false,
    category: 'gpt_account',
    gallery: [],
    sort_order: 0,
    ...overrides
  } as PublicProduct
}

describe('productBrandMeta', () => {
  const gptCategories = ['gpt_topup', 'gpt_account', 'gpt_rental', 'codex', 'gpt_usage', 'gpt_api']

  it.each(gptCategories)('GPT 系品类 %s 一律命中 chatgpt 品牌并强制品牌图优先', (cat) => {
    const meta = productBrandMeta(makeProduct({ category: cat as PublicProduct['category'] }))
    expect(meta.key).toBe('chatgpt')
    // 关键守卫：preferBrandImage 必须为真，否则后台占位图会顶掉 OpenAI 结。
    expect(meta.preferBrandImage).toBe(true)
  })

  it('x_premium 命中 X 品牌', () => {
    expect(productBrandMeta(makeProduct({ category: 'x_premium' })).key).toBe('x')
  })

  it('gemini 命中 Gemini 品牌', () => {
    expect(productBrandMeta(makeProduct({ category: 'gemini' })).key).toBe('gemini')
  })

  it('未知品类兜底 generic', () => {
    expect(productBrandMeta(makeProduct({ category: 'something_else' as PublicProduct['category'] })).key).toBe(
      'generic'
    )
  })
})

describe('resolveBrandDisplay（图标优先级守卫）', () => {
  it('GPT 商品即使后台配了 image_url，也恒定用品牌图（历史坑：曾整批传成 Codex 云朵图）', () => {
    // 真实线上场景：7 个 GPT 商品 image_url 都是同一张 Codex 云朵占位图（387909 字节）。
    const product = makeProduct({
      category: 'gpt_account',
      image_url: 'https://cdn.example.com/codex-cloud.png'
    })
    expect(hasProductImage(product)).toBe(true) // 确认后台确实带了图（破局点）
    const d = resolveBrandDisplay(product)
    // 回归断言：绝不能因为 image_url 存在就显示后台图。
    expect(d.useProductImage).toBe(false)
    expect(d.useBrandImage).toBe(true)
    expect(d.useGlyph).toBe(false)
    expect(d.meta.key).toBe('chatgpt')
  })

  it('GPT 商品无 image_url 时也用品牌图', () => {
    const d = resolveBrandDisplay(makeProduct({ category: 'gpt_account', image_url: '' }))
    expect(d.useProductImage).toBe(false)
    expect(d.useBrandImage).toBe(true)
  })

  it('非强制优先品类（generic）带后台图时，用后台商品图', () => {
    const d = resolveBrandDisplay(
      makeProduct({ category: 'other' as PublicProduct['category'], image_url: 'https://cdn.example.com/p.png' })
    )
    expect(d.useProductImage).toBe(true)
    expect(d.useBrandImage).toBe(false)
  })

  it('非强制优先品类且无后台图时，回退内联字形', () => {
    const d = resolveBrandDisplay(makeProduct({ category: 'other' as PublicProduct['category'], image_url: '' }))
    expect(d.useProductImage).toBe(false)
    expect(d.useBrandImage).toBe(false)
    expect(d.useGlyph).toBe(true)
  })

  it('x_premium 无后台图时回退字形（X 无品牌图片资源）', () => {
    const d = resolveBrandDisplay(makeProduct({ category: 'x_premium', image_url: '' }))
    expect(d.useGlyph).toBe(true)
    expect(d.useBrandImage).toBe(false)
  })
})
