import { apiClient } from './client'
import { buildApiUrl } from './url'
import type { BasePaginationResponse } from '@/types'
import type { CreateOrderResult } from '@/types/payment'
import { SHOP_CATEGORIES, type ShopCategory } from '@/constants/shop'

export type ShopProductType = 'virtual' | 'platform_usd_balance'
export type ShopProductStatus = 'draft' | 'published' | 'archived'
export type ShopFulfillmentMode = 'manual' | 'session_topup' | 'account_delivery' | 'rental'

/** 素材用途：后端按用途套用不同的体积/尺寸上限（对齐首页展示位）。 */
export type ShopAssetPurpose = 'product' | 'banner'

export type { ShopCategory }
export { SHOP_CATEGORIES }

export interface ShopProduct {
  id: number
  name: string
  description: string
  image_url: string
  product_type: ShopProductType
  price_cny_minor: number
  original_price_cny_minor: number
  grant_usd_amount: string
  stock_quantity?: number | null
  sold_count: number
  commission_bps: number
  status: ShopProductStatus
  sort_order: number
  fulfillment_mode: ShopFulfillmentMode
  delivery_form_hint: string
  badge_text: string
  spec_label: string
  highlight: boolean
  category: ShopCategory
  gallery: string[]
  created_at?: string
  updated_at?: string
}

export interface ShopBanner {
  id: number
  title: string
  subtitle: string
  image_url: string
  button_text: string
  product_id?: number | null
  enabled: boolean
  sort_order: number
  created_at?: string
  updated_at?: string
}

export interface ShopOrder {
  id: number
  user_id: number
  user_email?: string
  product_id: number
  payment_order_id?: number
  status: string
  fulfillment_status: string
  commission_status: string
  snapshot_name: string
  snapshot_description: string
  snapshot_image_url: string
  snapshot_product_type: ShopProductType
  snapshot_price_cny_minor: number
  /** 本单用人民币返点余额抵扣的金额（分） */
  wallet_applied_cny_minor: number
  /** 本单还需外部支付（微信/支付宝）的金额（分）；恒等于 价格 - 抵扣 */
  payable_cny_minor: number
  snapshot_grant_usd_amount: string
  snapshot_commission_bps: number
  fulfillment_note: string
  order_no?: string
  guest_token?: string
  guest_contact?: string
  snapshot_fulfillment_mode?: ShopFulfillmentMode
  delivery_hint?: string
  delivery_submitted_at?: string
  rental_duration?: string
  created_at: string
  paid_at?: string
  fulfilled_at?: string
}

export interface CreateShopOrderResult {
  shop_order_id: number
  /** 全额返点余额抵扣时没有外部支付单，此时 payment 缺失 */
  payment?: CreateOrderResult
  /** 本单抵扣的返点余额（分） */
  wallet_applied_cny_minor: number
  /** 本单还需外部支付的金额（分），为 0 表示全额抵扣 */
  payable_cny_minor: number
  /** 为 true 时订单已直接进入交付，前端不要再拉起支付渠道 */
  fully_paid_by_wallet: boolean
}

/**
 * 「我的返点余额」快照。人民币返点与 API 美金额度是两个独立的资金账户：
 * 这里只反映推广返点，和 user_credit_accounts 的平台额度无关。
 */
export interface ShopWalletBalance {
  /** 推广计划是否启用；未启用时前端不应展示抵扣入口 */
  enabled: boolean
  available_cny_minor: number
  frozen_cny_minor: number
}

/** 后台可管理的商城品类（迁移 208）。 */
export interface AdminShopCategory {
  id: number
  slug: string
  label: string
  blurb: string
  sort_order: number
  enabled: boolean
  /** 该品类下未删除商品数；>0 时后端拒绝删除 */
  product_count: number
  created_at?: string
  updated_at?: string
}

export interface ShopCategoryPayload {
  /** 仅创建时生效；更新时后端忽略（slug 不可变更） */
  slug: string
  label: string
  blurb?: string
  sort_order?: number
  enabled?: boolean
}

export interface ShopProductPayload {
  name: string
  description?: string
  image_url?: string
  product_type: ShopProductType
  price_cny_minor: number
  original_price_cny_minor?: number
  grant_usd_amount?: string
  stock_quantity?: number | null
  commission_bps?: number
  status: ShopProductStatus
  sort_order?: number
  fulfillment_mode?: ShopFulfillmentMode
  delivery_form_hint?: string
  badge_text?: string
  spec_label?: string
  highlight?: boolean
  category?: ShopCategory
  gallery?: string[]
}

export interface ShopBannerPayload {
  title: string
  subtitle?: string
  image_url?: string
  button_text?: string
  product_id?: number | null
  enabled: boolean
  sort_order?: number
}

export function resolveShopAssetUrl(url?: string | null): string {
  const raw = String(url || '').trim()
  if (!raw) return ''
  if (/^(data|blob):/i.test(raw)) return raw
  if (/^https?:\/\//i.test(raw) || raw.startsWith('//')) return raw

  const normalized = raw.startsWith('/') ? raw : `/${raw}`
  if (normalized.startsWith('/api/v1/shop/assets/')) {
    return buildApiUrl(normalized)
  }
  if (normalized.startsWith('/shop/assets/')) {
    return buildApiUrl(normalized)
  }

  return raw
}

export const shopAPI = {
  listBanners() {
    return apiClient.get<ShopBanner[]>('/shop/banners')
  },
  listProducts() {
    return apiClient.get<ShopProduct[]>('/shop/products')
  },
  /** 可用于商城抵扣的人民币返点余额（与 API 美金额度无关） */
  walletBalance() {
    return apiClient.get<ShopWalletBalance>('/shop/wallet-balance')
  },
  createOrder(data: {
    product_id: number
    payment_type: string
    return_url?: string
    is_mobile?: boolean
    /** true 时先扣人民币返点余额，差额再走微信/支付宝 */
    use_wallet?: boolean
  }) {
    return apiClient.post<CreateShopOrderResult>('/shop/orders', data)
  },
  myOrders(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<ShopOrder>>('/shop/orders/my', { params })
  },
  /** 登录用户为自己的订单补交交付资料（Session / 收货邮箱 / 租期） */
  submitDelivery(id: number, data: { payload: string; rental_duration?: string }) {
    return apiClient.post<{ submitted: boolean }>(`/shop/orders/${id}/delivery`, data)
  },
  /** 认领此前以免登录方式在首页下的订单，归入当前账号 */
  claimOrder(data: { order_no: string; contact: string }) {
    return apiClient.post<{ claimed: boolean }>('/shop/orders/claim', data)
  }
}

export const adminShopAPI = {
  listCategories() {
    return apiClient.get<AdminShopCategory[]>('/admin/shop/categories')
  },
  createCategory(data: ShopCategoryPayload) {
    return apiClient.post<AdminShopCategory>('/admin/shop/categories', data)
  },
  updateCategory(id: number, data: ShopCategoryPayload) {
    return apiClient.put<AdminShopCategory>(`/admin/shop/categories/${id}`, data)
  },
  deleteCategory(id: number) {
    return apiClient.delete(`/admin/shop/categories/${id}`)
  },
  listProducts() {
    return apiClient.get<ShopProduct[]>('/admin/shop/products')
  },
  createProduct(data: ShopProductPayload) {
    return apiClient.post<ShopProduct>('/admin/shop/products', data)
  },
  updateProduct(id: number, data: ShopProductPayload) {
    return apiClient.put<ShopProduct>(`/admin/shop/products/${id}`, data)
  },
  deleteProduct(id: number) {
    return apiClient.delete(`/admin/shop/products/${id}`)
  },
  listBanners() {
    return apiClient.get<ShopBanner[]>('/admin/shop/banners')
  },
  createBanner(data: ShopBannerPayload) {
    return apiClient.post<ShopBanner>('/admin/shop/banners', data)
  },
  updateBanner(id: number, data: ShopBannerPayload) {
    return apiClient.put<ShopBanner>(`/admin/shop/banners/${id}`, data)
  },
  deleteBanner(id: number) {
    return apiClient.delete(`/admin/shop/banners/${id}`)
  },
  listOrders(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<ShopOrder>>('/admin/shop/orders', { params })
  },
  fulfillOrder(id: number, data: { fulfillment_note: string }) {
    return apiClient.post(`/admin/shop/orders/${id}/fulfill`, data)
  },
  getOrderDelivery(id: number) {
    return apiClient.get<{ payload: string }>(`/admin/shop/orders/${id}/delivery`)
  },
  /**
   * 上传商城素材。`purpose` 决定后端按哪档体积/尺寸上限校验（首页商品图 vs 轮播图），
   * 与首页展示位一一对应。
   */
  uploadAsset(file: File, purpose: ShopAssetPurpose = 'product') {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('purpose', purpose)
    return apiClient.post<{ url: string }>('/admin/shop/assets', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  }
}
