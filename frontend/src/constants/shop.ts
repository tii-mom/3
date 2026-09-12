/**
 * 商城品类（迁移 207 引入，迁移 208 起改为后台可配置）。
 *
 * 品类回答「卖什么」，与 fulfillment_mode（怎么交付）正交：
 * 同一个品类下可以有代充、成品号、租号等不同交付方式。
 *
 * **权威数据源是后台**：`GET /public/shop/categories` 返回启用的品类（slug/label/blurb/排序）。
 * 本文件只保留「内置兜底」——接口失败或商品品类已被删除时，仍能渲染出可读的标签，
 * 而不是让整块分组消失。
 *
 * 因此 slug 是开放的字符串（`ShopCategory = string`），后台新建 gpt_usage 之类的新品类
 * 不需要改前端代码。
 */

/** 品类标识。后台可自由新增，故为开放字符串。 */
export type ShopCategory = string

export interface ShopCategoryMeta {
  value: string
  label: string
  /** 首页分组下的一句说明，控制在 20 字以内 */
  blurb: string
}

/** 内置兜底品类（仅当接口不可用时使用；顺序也是兜底顺序）。 */
export const SHOP_CATEGORIES: ShopCategoryMeta[] = [
  { value: 'gpt_topup', label: 'GPT 官方充值', blurb: '给已有账号续费升级' },
  { value: 'gpt_account', label: 'GPT 成品号', blurb: '开好即用的独享账号' },
  { value: 'gpt_rental', label: 'GPT 租号', blurb: '按月租赁，短期试用' },
  { value: 'gpt_usage', label: 'GPT 使用服务', blurb: 'Codex / API 额度等增值服务' },
  { value: 'x_premium', label: 'X 会员', blurb: 'X Premium 订阅开通' },
  { value: 'gemini', label: 'Gemini', blurb: 'Gemini Advanced 订阅' },
  { value: 'codex', label: 'Codex 额度', blurb: 'Codex / API 额度补充' },
  { value: 'other', label: '其他服务', blurb: '其余增值服务' },
]

const BUILTIN_MAP = new Map(SHOP_CATEGORIES.map((item) => [item.value, item]))

/** 把接口返回的品类数组转成查表用的 Map。 */
export function buildCategoryMap(categories: ShopCategoryMeta[]): Map<string, ShopCategoryMeta> {
  return new Map(categories.map((item) => [item.value, item]))
}

/**
 * 归一化品类 slug：只做 trim 与空值兜底，**不再把未知 slug 改写成 other**。
 * 否则后台新建的品类会被前端悄悄归到「其他」。
 */
export function normalizeShopCategory(value?: string | null): ShopCategory {
  const trimmed = String(value ?? '').trim()
  return trimmed || 'other'
}

export function shopCategoryLabel(value?: string | null, map?: Map<string, ShopCategoryMeta>): string {
  const slug = normalizeShopCategory(value)
  return map?.get(slug)?.label || BUILTIN_MAP.get(slug)?.label || slug
}

export function shopCategoryBlurb(value?: string | null, map?: Map<string, ShopCategoryMeta>): string {
  const slug = normalizeShopCategory(value)
  return map?.get(slug)?.blurb || BUILTIN_MAP.get(slug)?.blurb || ''
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
 * 后台自建品类没有专属清单时，按 slug 前缀归到最接近的家族（gpt_* → GPT 代充），仍是相关文案。
 */
export const SHOP_CATEGORY_BENEFITS: Record<string, string[]> = {
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
  gpt_rental: [
    '按月租赁的独享 ChatGPT 账号',
    '租期内独享，不与他人共用',
    '到期可续租，无需重新配置',
    '非人为中断 30 天质保',
  ],
  gpt_usage: [
    'Codex / API 额度补充',
    '按量计费，用完即止',
    '支持 CLI 与 API 调用',
    '额度实时到账',
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

/** 按前缀把未知 slug 归到最接近的内置家族，避免自建品类掉进「其他服务」的泛文案。 */
function resolveBenefitsFamily(slug: string): string {
  if (SHOP_CATEGORY_BENEFITS[slug]) return slug
  if (slug.startsWith('gpt_account')) return 'gpt_account'
  if (slug.startsWith('gpt')) return 'gpt_topup'
  if (slug.startsWith('x_')) return 'x_premium'
  if (slug.startsWith('gemini')) return 'gemini'
  if (slug.startsWith('codex')) return 'codex'
  return 'other'
}

export function shopCategoryBenefits(category?: string | null): string[] {
  const slug = normalizeShopCategory(category)
  return SHOP_CATEGORY_BENEFITS[resolveBenefitsFamily(slug)] || SHOP_CATEGORY_BENEFITS.other
}
