/**
 * 首屏促销条的「最优折扣」推导。
 *
 * 为什么要单独抽出来：首页促销条上显示的优惠金额必须是**真实存在**的，
 * 即商品后台确实配了划线原价（original_price > price）。凭空写一个「立减 ¥20」
 * 而结算时对不上，属于虚假宣传。所以统一走这个纯函数，只认数据。
 */

export interface DealLike {
  price_cny_minor: number
  original_price_cny_minor: number
}

export interface BestDeal<T> {
  product: T
  /** 相对划线原价省下的金额（分），恒 > 0 */
  savingMinor: number
}

/**
 * 挑出折扣金额最大的商品；没有任何商品配了有效划线原价时返回 null。
 * 金额相等时保留先出现的那个，保证同一份数据每次渲染结果一致。
 */
export function pickBestDeal<T extends DealLike>(products: readonly T[]): BestDeal<T> | null {
  let best: BestDeal<T> | null = null
  for (const product of products) {
    const saving = (product.original_price_cny_minor || 0) - (product.price_cny_minor || 0)
    if (saving <= 0) continue
    if (!best || saving > best.savingMinor) best = { product, savingMinor: saving }
  }
  return best
}
