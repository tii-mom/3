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

/**
 * 交付模式 → 卡片「权益面板」要展示的事实（到账时间 / 账号归属 / 需提交什么）。
 * 零后端改动：卡片直接按 fulfillment_mode 派生，不需要新增字段。
 */
export interface ShopModeFact {
  /** 卡片里的小标签，如「代充值」「成品号」 */
  label: string
  /** 到账 / 可用时间 */
  turnaround: string
  /** 账号归属 */
  belongs: string
  /** 支付后需要提交什么 */
  submit: string
}

const SHOP_MODE_FACTS_MAP: Record<string, ShopModeFact> = {
  manual: { label: '人工处理', turnaround: '管理员处理', belongs: '按订单说明', submit: '无需提交资料' },
  session_topup: { label: '代充值', turnaround: '1-3 分钟到账', belongs: '你自己的号', submit: '需登录凭证' },
  account_delivery: { label: '成品号', turnaround: '发货后可用', belongs: '新开账号', submit: '需接收邮箱' },
  rental: { label: '租号', turnaround: '租期内可用', belongs: '平台提供', submit: '需邮箱与租期' },
}

export function shopModeFact(mode?: string | null): ShopModeFact {
  return SHOP_MODE_FACTS_MAP[mode || ''] || SHOP_MODE_FACTS_MAP.manual
}

/**
 * 品类 → 「套餐包含」权益清单（3-5 条）。
 * 派生式（决策 A1）：零后端改动；每商品的差异靠 spec_label / delivery_form_hint 在面板里体现。
 * 若要每商品独立编辑权益，需新增 benefits_json 列（本需求明确不碰数据库，故不做）。
 */
export const SHOP_CATEGORY_BENEFITS: Record<ShopCategory, string[]> = {
  gpt_topup: [
    '官方渠道代充，非共享账号',
    '全程不需要账号密码',
    '支付后 1-3 分钟到账',
    '非人为中断 30 天质保',
  ],
  gpt_account: [
    '开好即用的独享账号',
    '绑定你自己的邮箱，可改密',
    '支持网页 / App / API 多端登录',
    '非人为封号 30 天质保',
  ],
  x_premium: [
    'X Premium 订阅正式开通',
    '蓝 V 标识实时生效',
    '支持 Grok 等会员权益',
    '赠送额度按说明发放',
  ],
  gemini: [
    'Gemini Advanced 订阅开通',
    '支持 Gemini 深度研究',
    '中文与多语言可用',
    '非人为中断质保',
  ],
  codex: [
    'Codex / API 额度补充',
    '按量计费，用完即止',
    '支持 CLI 与 API 调用',
    '额度实时到账',
  ],
  other: [
    '专属客服对接',
    '按订单说明交付',
    '支持开具凭证',
    '有问题随时售后',
  ],
}

export function shopCategoryBenefits(category?: string | null): string[] {
  return SHOP_CATEGORY_BENEFITS[normalizeShopCategory(category)] || SHOP_CATEGORY_BENEFITS.other
}
