/**
 * 首屏促销条的选品规则。
 *
 * 为什么要有规则、而不是直接写死一句折扣文案：
 * 1. 后台目前没有「新客立减 / 优惠券」这类规则。促销条上凭空写一个优惠金额，
 *    用户下单时对不上，属于虚假宣传 —— 所以任何金额都必须来自真实商品数据。
 * 2. 促销条是首屏最响的位置，一旦自动挑中一个名字不好看的商品
 *    （例如「XX（无质保）」），反而伤转化。因此**只推后台勾了「推荐位」的商品**，
 *    把选品权交回给店主；没有任何推荐位时退化成一句不含商品名的服务承诺。
 */

export interface PromoProductLike {
  name: string
  price_cny_minor: number
  original_price_cny_minor: number
  highlight: boolean
}

export interface HeroPromoPlan<T extends PromoProductLike> {
  /** 要主推的商品；null 表示退化为通用文案 */
  product: T | null
  /** 该商品相对划线原价省下的金额（分）；0 表示没有折扣，不显示省多少 */
  savingMinor: number
  /** 全站最低价（分），用于通用文案里的「最低 ¥X 起」；无商品时为 null */
  minPriceMinor: number | null
}

/**
 * @param products 已按展示顺序（推荐位优先 → sort_order → id）排好的商品列表
 */
export function planHeroPromo<T extends PromoProductLike>(products: readonly T[]): HeroPromoPlan<T> {
  const product = products.find((item) => item.highlight) ?? null

  const savingMinor = product
    ? Math.max(0, (product.original_price_cny_minor || 0) - (product.price_cny_minor || 0))
    : 0

  const minPriceMinor = products.reduce<number | null>(
    (min, item) => (min === null || item.price_cny_minor < min ? item.price_cny_minor : min),
    null
  )

  return { product, savingMinor, minPriceMinor }
}
