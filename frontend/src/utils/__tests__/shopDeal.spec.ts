import { describe, expect, it } from 'vitest'
import { pickBestDeal } from '../shopDeal'

const p = (name: string, price: number, original: number) => ({
  name,
  price_cny_minor: price,
  original_price_cny_minor: original
})

describe('pickBestDeal', () => {
  it('没有任何划线原价时返回 null（首屏退化为无数字承诺）', () => {
    expect(pickBestDeal([p('a', 15800, 0), p('b', 9900, 9900)])).toBeNull()
  })

  it('空列表返回 null', () => {
    expect(pickBestDeal([])).toBeNull()
  })

  it('原价低于售价（脏数据）时忽略该商品，不能算出负优惠', () => {
    expect(pickBestDeal([p('a', 15800, 12000)])).toBeNull()
  })

  it('取折扣金额最大的商品', () => {
    const result = pickBestDeal([p('小折', 15800, 17800), p('大折', 9900, 19900)])
    expect(result?.product.name).toBe('大折')
    expect(result?.savingMinor).toBe(10000)
  })

  it('金额相同时保留先出现的，保证渲染稳定', () => {
    const result = pickBestDeal([p('先', 10000, 12000), p('后', 30000, 32000)])
    expect(result?.product.name).toBe('先')
  })

  it('忽略 0 折扣的商品，但保留真正有折扣的', () => {
    const result = pickBestDeal([p('无折', 10000, 10000), p('有折', 10000, 15000)])
    expect(result?.product.name).toBe('有折')
    expect(result?.savingMinor).toBe(5000)
  })
})
