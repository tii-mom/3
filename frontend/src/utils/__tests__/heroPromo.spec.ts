import { describe, expect, it } from 'vitest'
import { planHeroPromo } from '../heroPromo'

const p = (name: string, price: number, original: number, highlight = false) => ({
  name,
  price_cny_minor: price,
  original_price_cny_minor: original,
  highlight
})

describe('planHeroPromo', () => {
  it('没有任何推荐位时 product 为 null（首屏不出现任意的商品名）', () => {
    const plan = planHeroPromo([p('Gemini 12个月（无质保）', 3800, 120000)])
    expect(plan.product).toBeNull()
    expect(plan.savingMinor).toBe(0)
  })

  it('优先取推荐位商品，而不是折扣最大的那个', () => {
    const plan = planHeroPromo([
      p('折扣很大但没推荐', 1000, 100000),
      p('推荐位商品', 15800, 17800, true)
    ])
    expect(plan.product?.name).toBe('推荐位商品')
    expect(plan.savingMinor).toBe(2000)
  })

  it('多个推荐位时取列表里最靠前的（列表已按展示顺序排好）', () => {
    const plan = planHeroPromo([
      p('第一个推荐', 10000, 12000, true),
      p('第二个推荐', 20000, 50000, true)
    ])
    expect(plan.product?.name).toBe('第一个推荐')
  })

  it('推荐位商品没有划线原价时 savingMinor 为 0，不能算出负优惠', () => {
    const plan = planHeroPromo([p('无折扣推荐', 10000, 0, true), p('原价更低', 10000, 8000, true)])
    expect(plan.product?.name).toBe('无折扣推荐')
    expect(plan.savingMinor).toBe(0)
  })

  it('minPriceMinor 取全站最低价，供通用文案使用', () => {
    const plan = planHeroPromo([p('a', 15800, 0), p('b', 995, 0), p('c', 4500, 0)])
    expect(plan.minPriceMinor).toBe(995)
  })

  it('空列表不报错', () => {
    const plan = planHeroPromo([])
    expect(plan.product).toBeNull()
    expect(plan.minPriceMinor).toBeNull()
  })
})
