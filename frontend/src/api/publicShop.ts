import { apiClient } from './client'
import type { CreateOrderResult } from '@/types/payment'
import type { ShopCategory } from '@/constants/shop'

/**
 * 官网首页「免登录直充」链路的数据层。
 * 与需要登录的 shopAPI 分离：下单只要求联系方式，支付后凭订单号 + 联系方式查单。
 */

export type FulfillmentMode = 'manual' | 'session_topup' | 'account_delivery' | 'rental'

export type { ShopCategory as ProductCategory } from '@/constants/shop'
export {
  SHOP_CATEGORIES,
  normalizeShopCategory,
  shopCategoryBlurb,
  shopCategoryLabel,
} from '@/constants/shop'

export interface PublicProduct {
  id: number
  name: string
  description: string
  image_url: string
  product_type: string
  price_cny_minor: number
  original_price_cny_minor: number
  stock_quantity?: number | null
  sold_count: number
  commission_bps?: number
  fulfillment_mode: FulfillmentMode
  delivery_form_hint: string
  badge_text: string
  spec_label: string
  highlight: boolean
  category: ShopCategory
  gallery: string[]
  sort_order: number
}

export interface PublicBanner {
  id: number
  title: string
  subtitle: string
  image_url: string
  button_text: string
  product_id?: number | null
  enabled: boolean
  sort_order: number
}

export interface GuestOrderResult {
  shop_order_id: number
  order_no: string
  guest_token: string
  payment: CreateOrderResult
}

export interface GuestOrderLookup {
  order_no: string
  status: string
  fulfillment_status: string
  snapshot_name: string
  snapshot_description: string
  snapshot_image_url: string
  snapshot_price_cny_minor: number
  snapshot_fulfillment_mode: FulfillmentMode
  fulfillment_note: string
  delivery_hint?: string
  delivery_submitted: boolean
  rental_duration?: string
  created_at: string
  paid_at?: string
}

export interface PublicPaymentMethod {
  payment_type?: string
  display_name?: string
  currency?: string
  fee_rate?: number
  daily_limit?: number
  single_min: number
  single_max: number
  available?: boolean
}

/** 与控制台商城同源的支付渠道与限额（登录态接口的公开精简版） */
export interface PublicPaymentMethodsResult {
  methods: Record<string, PublicPaymentMethod>
  global_min: number
  global_max: number
  alipay_force_qrcode: boolean
}

export const publicShopAPI = {
  listProducts() {
    return apiClient.get<PublicProduct[]>('/public/shop/products')
  },
  listBanners() {
    return apiClient.get<PublicBanner[]>('/public/shop/banners')
  },
  listPaymentMethods() {
    return apiClient.get<PublicPaymentMethodsResult>('/public/shop/payment-methods')
  },
  createOrder(data: { product_id: number; payment_type: string; contact: string; return_url?: string; is_mobile?: boolean }) {
    return apiClient.post<GuestOrderResult>('/public/shop/orders', data)
  },
  submitDelivery(data: { order_no: string; contact: string; payload: string; rental_duration?: string }) {
    return apiClient.post<{ submitted: boolean }>('/public/shop/orders/delivery', data)
  },
  lookupOrder(params: { order_no: string; contact: string }) {
    return apiClient.get<GuestOrderLookup>('/public/shop/orders/lookup', { params })
  }
}

// --- 本机订单凭据 ---
// guest_token 只保存在用户本机，用于支付后提交资料与查单；
// 服务端校验始终要求 订单号 + 联系方式，token 泄露也无法拿到他人订单。

const TRACK_STORAGE_KEY = '3api.guest.orders'

export interface GuestOrderCredential {
  order_no: string
  guest_token: string
  contact: string
  product_name: string
  fulfillment_mode: FulfillmentMode
  created_at: number
}

function readCredentials(): GuestOrderCredential[] {
  try {
    const raw = localStorage.getItem(TRACK_STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export function rememberGuestOrder(credential: GuestOrderCredential): void {
  const list = readCredentials().filter((item) => item.order_no !== credential.order_no)
  list.unshift(credential)
  try {
    localStorage.setItem(TRACK_STORAGE_KEY, JSON.stringify(list.slice(0, 20)))
  } catch {
    // 存储不可用时忽略，用户仍可凭订单号 + 联系方式查单
  }
}

export function findGuestOrder(orderNo: string): GuestOrderCredential | undefined {
  return readCredentials().find((item) => item.order_no === orderNo)
}

export function listGuestOrders(): GuestOrderCredential[] {
  return readCredentials()
}

// --- 展示辅助 ---

export function formatCNY(minor: number): string {
  const value = Number(minor || 0) / 100
  return value.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

export const FULFILLMENT_LABELS: Record<FulfillmentMode, string> = {
  manual: '人工处理',
  session_topup: '代充值',
  account_delivery: '成品号',
  rental: '租号'
}

/** 该交付模式下，用户需要提交什么资料 */
export function deliveryRequirement(mode: FulfillmentMode): {
  needsSession: boolean
  needsEmail: boolean
  needsDuration: boolean
} {
  return {
    needsSession: mode === 'session_topup',
    needsEmail: mode === 'account_delivery' || mode === 'rental',
    needsDuration: mode === 'rental'
  }
}

/** 从粘贴内容中提取 ChatGPT session 的 accessToken，兼容整段 JSON 与裸 token */
export function extractSessionToken(raw: string): string {
  const text = String(raw || '').trim()
  if (!text) return ''
  if (text.startsWith('{')) {
    try {
      const parsed = JSON.parse(text)
      const token = parsed?.accessToken
      return typeof token === 'string' ? token.trim() : ''
    } catch {
      return ''
    }
  }
  return text
}
