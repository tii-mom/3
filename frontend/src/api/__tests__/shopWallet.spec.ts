import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    get,
    post,
  },
}))

import { shopAPI } from '@/api/shop'

describe('shop wallet deduction api', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    get.mockResolvedValue({ data: {} })
    post.mockResolvedValue({ data: {} })
  })

  it('reads the RMB rebate balance from the dedicated endpoint', async () => {
    await shopAPI.walletBalance()

    expect(get).toHaveBeenCalledWith('/shop/wallet-balance')
  })

  it('forwards the use_wallet flag so checkout can deduct the rebate balance', async () => {
    await shopAPI.createOrder({
      product_id: 7,
      payment_type: 'alipay',
      return_url: 'https://3api.shop/payment/result',
      is_mobile: false,
      use_wallet: true,
    })

    expect(post).toHaveBeenCalledWith('/shop/orders', {
      product_id: 7,
      payment_type: 'alipay',
      return_url: 'https://3api.shop/payment/result',
      is_mobile: false,
      use_wallet: true,
    })
  })
})
