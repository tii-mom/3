import { apiClient } from './client'
import type { BasePaginationResponse } from '@/types'
import type { CreateOrderResult } from '@/types/payment'

export type ShopProductType = 'virtual' | 'platform_usd_balance'
export type ShopProductStatus = 'draft' | 'published' | 'archived'

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
  snapshot_grant_usd_amount: string
  snapshot_commission_bps: number
  created_at: string
  paid_at?: string
  fulfilled_at?: string
}

export interface CreateShopOrderResult {
  shop_order_id: number
  payment: CreateOrderResult
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

export const shopAPI = {
  listBanners() {
    return apiClient.get<ShopBanner[]>('/shop/banners')
  },
  listProducts() {
    return apiClient.get<ShopProduct[]>('/shop/products')
  },
  createOrder(data: { product_id: number; payment_type: string; return_url?: string; is_mobile?: boolean }) {
    return apiClient.post<CreateShopOrderResult>('/shop/orders', data)
  },
  myOrders(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<ShopOrder>>('/shop/orders/my', { params })
  }
}

export const adminShopAPI = {
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
  uploadAsset(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return apiClient.post<{ url: string }>('/admin/shop/assets', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  }
}
