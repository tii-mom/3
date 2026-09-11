/**
 * 商城品类（迁移 207）。
 *
 * 品类回答「卖什么」，与 fulfillment_mode（怎么交付）正交：
 * 同一个品类下可以有代充、成品号、租号等不同交付方式。
 *
 * 后台表单、官网首页分组、结构化数据共用这一份枚举，
 * 避免三处各写一遍导致不一致。
 */

export type ShopCategory =
  | 'gpt_topup'
  | 'gpt_account'
  | 'x_premium'
  | 'gemini'
  | 'codex'
  | 'other'

export interface ShopCategoryMeta {
  value: ShopCategory
  label: string
  /** 首页分组下的一句说明，控制在 20 字以内 */
  blurb: string
}

export const SHOP_CATEGORIES: ShopCategoryMeta[] = [
  { value: 'gpt_topup', label: 'GPT 代充值', blurb: '给已有账号续费升级' },
  { value: 'gpt_account', label: 'GPT 成品号', blurb: '开好即用的独享账号' },
  { value: 'x_premium', label: 'X 蓝 V', blurb: 'X Premium 订阅开通' },
  { value: 'gemini', label: 'Gemini 会员', blurb: 'Gemini Advanced 订阅' },
  { value: 'codex', label: 'Codex 额度', blurb: 'Codex / API 额度补充' },
  { value: 'other', label: '其他服务', blurb: '其余增值服务' },
]

const CATEGORY_MAP = new Map(SHOP_CATEGORIES.map((item) => [item.value, item]))

export function isShopCategory(value: unknown): value is ShopCategory {
  return typeof value === 'string' && CATEGORY_MAP.has(value as ShopCategory)
}

/** 后端返回脏值时兜底到 other，保证首页分组永远能渲染。 */
export function normalizeShopCategory(value?: string | null): ShopCategory {
  return isShopCategory(value) ? value : 'other'
}

export function shopCategoryLabel(value?: string | null): string {
  return CATEGORY_MAP.get(normalizeShopCategory(value))?.label ?? '其他服务'
}

export function shopCategoryBlurb(value?: string | null): string {
  return CATEGORY_MAP.get(normalizeShopCategory(value))?.blurb ?? ''
}
