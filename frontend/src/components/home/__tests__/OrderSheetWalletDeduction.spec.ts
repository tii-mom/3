import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'

// 登录态用一个可切换的共享标志：组件内部读的是 authStore.isAuthenticated。
const { auth } = vi.hoisted(() => ({ auth: { value: true } }))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    get isAuthenticated() {
      return auth.value
    }
  })
}))

vi.mock('@/api/shop', () => ({
  resolveShopAssetUrl: (url?: string) => url || ''
}))

vi.mock('@/api/publicShop', () => ({
  // 与线上一致：金额按分转元；这里固定两位小数，便于断言。
  formatCNY: (minor: number) => (Number(minor || 0) / 100).toFixed(2),
  FULFILLMENT_LABELS: { manual: '人工发货' },
  deliveryRequirement: () => ({ needsSession: false, needsEmail: false })
}))

import OrderSheet from '../OrderSheet.vue'

const product = {
  id: 1,
  name: 'GPT 充值',
  description: '',
  image_url: '',
  gallery: [],
  price_cny_minor: 10000,
  fulfillment_mode: 'manual'
} as never

const methods = [
  { value: 'alipay', label: '支付宝', single_min: 0, single_max: 0 },
  { value: 'wxpay', label: '微信支付', single_min: 0, single_max: 0 }
]

/** 打开抽屉：组件是常驻实例，抵扣默认勾选发生在 open 由 false 变 true 时。 */
async function openSheet(props: Record<string, unknown> = {}) {
  const wrapper = mount(OrderSheet, {
    props: { open: false, product, paymentMethods: methods, submitting: false, ...props }
  })
  await wrapper.setProps({ open: true })
  // 抵扣默认勾选 → 应付金额变化 → 渠道自动纠正，需要让 Vue 的调度器跑完
  await flushPromises()
  await nextTick()
  await flushPromises()
  return wrapper
}

function walletButton(wrapper: ReturnType<typeof mount>) {
  return wrapper.find('.sheet__wallet')
}

describe('OrderSheet 返点余额抵扣', () => {
  beforeEach(() => {
    auth.value = true
  })

  it('余额不足全额时按「抵扣后应付」结算，并展示抵扣明细', async () => {
    const wrapper = await openSheet({ walletEligible: true, walletAvailableMinor: 4000 })

    // 有余额时默认勾选抵扣（防呆：多数用户就是想把返点花掉）
    expect(walletButton(wrapper).classes()).toContain('sheet__wallet--active')

    // 抵扣明细：商品 ¥100 - 返点 ¥40 = 应付 ¥60
    const summary = wrapper.find('.sheet__summary').text()
    expect(summary).toContain('商品金额')
    expect(summary).toContain('100.00')
    expect(summary).toContain('返点余额抵扣')
    expect(summary).toContain('-¥40.00')
    expect(summary).toContain('应付金额')
    expect(summary).toContain('¥60.00')

    // 仍需外部支付 → 展示支付方式，按钮金额是抵扣后的 60
    expect(wrapper.text()).toContain('支付方式')
    expect(wrapper.find('.sheet__submit').text()).toContain('去支付 ¥60.00')
  })

  it('余额足够全额抵扣时不展示支付方式，提交不带渠道', async () => {
    const wrapper = await openSheet({ walletEligible: true, walletAvailableMinor: 10000 })

    expect(wrapper.find('.sheet__wallet-full').exists()).toBe(true)
    // 全额抵扣没有外部支付环节：支付方式整块隐藏
    expect(wrapper.find('.sheet__methods').exists()).toBe(false)
    expect(wrapper.find('.sheet__submit').text()).toContain('确认抵扣下单')

    await wrapper.find('.sheet__submit').trigger('click')
    const payload = wrapper.emitted('submit')?.[0]?.[0] as {
      paymentType: string
      useWallet: boolean
    }
    expect(payload).toEqual({ contact: '', paymentType: '', useWallet: true })
  })

  it('关掉抵扣后应付金额回到商品原价', async () => {
    const wrapper = await openSheet({ walletEligible: true, walletAvailableMinor: 4000 })

    await walletButton(wrapper).trigger('click')
    await flushPromises()

    const summary = wrapper.find('.sheet__summary').text()
    expect(summary).not.toContain('返点余额抵扣')
    expect(summary).toContain('¥100.00')
    expect(wrapper.find('.sheet__submit').text()).toContain('去支付 ¥100.00')

    await wrapper.find('.sheet__submit').trigger('click')
    const payload = wrapper.emitted('submit')?.[0]?.[0] as { useWallet: boolean }
    expect(payload.useWallet).toBe(false)
  })

  it('没有可用余额（如推广计划未开启）时不展示抵扣入口', async () => {
    const wrapper = await openSheet({ walletEligible: false, walletAvailableMinor: 0 })

    expect(wrapper.find('.sheet__wallet').exists()).toBe(false)
    expect(wrapper.find('.sheet__wallet-full').exists()).toBe(false)
    expect(wrapper.find('.sheet__submit').text()).toContain('去支付 ¥100.00')
  })

  it('访客没有返点钱包：即使传了余额也不展示抵扣入口', async () => {
    auth.value = false
    const wrapper = await openSheet({ walletEligible: true, walletAvailableMinor: 10000 })

    expect(wrapper.find('.sheet__wallet').exists()).toBe(false)
    expect(wrapper.find('.sheet__submit').text()).toContain('去支付 ¥100.00')
  })
})
