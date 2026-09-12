import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores'
import {
  publicShopAPI,
  rememberGuestOrder,
  type FulfillmentMode,
  type PublicPaymentMethod,
  type PublicProduct
} from '@/api/publicShop'
import { shopAPI } from '@/api/shop'
import { decidePaymentLaunch } from '@/components/payment/paymentFlow'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import type { OrderType } from '@/types/payment'

/**
 * 首页免登录下单：选套餐 → 填联系方式 → 拉起支付。
 * 与登录态无关，支付成功后跳转查单页补交交付资料（Session / 收货邮箱）。
 *
 * 支付渠道与控制台商城同源（GET /public/shop/payment-methods），后台开关与单笔限额即时生效；
 * 首页只启用适合免登录冲动消费的即时到账渠道，其余渠道引导到控制台商城使用。
 */

/** 免登录场景支持的即时到账渠道 */
const GUEST_METHOD_WHITELIST = ['alipay', 'wxpay', 'easypay']

const PAYMENT_LABELS: Record<string, string> = {
  alipay: '支付宝',
  wxpay: '微信支付',
  stripe: '银行卡',
  airwallex: '空中云汇',
  easypay: '聚合支付'
}

export interface GuestPaymentOption {
  value: string
  label: string
  single_min: number
  single_max: number
}

function isMobileDevice(): boolean {
  if (typeof window === 'undefined') return false
  return /Mobile|Android|iPhone|iPad/i.test(window.navigator.userAgent)
}

function isWechatBrowser(): boolean {
  if (typeof window === 'undefined') return false
  return /MicroMessenger/i.test(window.navigator.userAgent)
}

export function useGuestCheckout() {
  const appStore = useAppStore()
  const authStore = useAuthStore()
  const router = useRouter()

  // 已登录用户走控制台商城同一下单通道，订单归属本人，控制台「我的商城订单」立即可见；
  // 未登录访客走免登录通道，凭订单号 + 联系方式查单，登录后可认领。
  const isLoggedIn = computed(() => !!authStore.isAuthenticated)

  const submitting = ref(false)
  const payDialog = reactive({
    open: false,
    orderId: 0,
    qrCode: '',
    expiresAt: '',
    paymentType: '',
    payUrl: ''
  })

  // 支付渠道来自后端配置，不再前端写死
  const paymentMethodLimits = ref<Record<string, PublicPaymentMethod>>({})
  const alipayForceQRCode = ref(false)
  const methodsLoaded = ref(false)

  /** 首页可用支付方式（免登录渠道白名单 ∩ 后台已开启） */
  const guestPaymentMethods = computed<GuestPaymentOption[]>(() =>
    GUEST_METHOD_WHITELIST
      .filter((value) => {
        const limit = paymentMethodLimits.value[value]
        return !!limit && limit.available !== false
      })
      .map((value) => {
        const limit = paymentMethodLimits.value[value]
        return {
          value,
          label: limit.display_name || PAYMENT_LABELS[value] || value,
          single_min: Number(limit.single_min) || 0,
          single_max: Number(limit.single_max) || 0
        }
      })
  )

  async function loadPaymentMethods(): Promise<void> {
    try {
      const response = await publicShopAPI.listPaymentMethods()
      const data = response.data
      paymentMethodLimits.value = data?.methods || {}
      alipayForceQRCode.value = !!data?.alipay_force_qrcode
    } catch {
      paymentMethodLimits.value = {}
      alipayForceQRCode.value = false
    } finally {
      methodsLoaded.value = true
    }
  }

  /** 按商品金额过滤出可用的支付方式（后台配置了单笔限额时以限额为准） */
  function methodsForAmount(priceMinor: number): GuestPaymentOption[] {
    const amount = Number(priceMinor || 0) / 100
    return guestPaymentMethods.value.filter((method) => {
      if (method.single_min > 0 && amount < method.single_min) return false
      if (method.single_max > 0 && amount > method.single_max) return false
      return true
    })
  }

  // 当前下单上下文，支付成功后用于跳转与本地留存
  let pendingContext: {
    orderNo: string
    guestToken: string
    contact: string
    productName: string
    fulfillmentMode: FulfillmentMode
  } | null = null

  async function startCheckout(product: PublicProduct, contact: string, paymentType: string): Promise<boolean> {
    if (submitting.value) return false
    const trimmedContact = String(contact || '').trim()
    // 登录用户无需联系方式：订单直接归属本人，可在控制台商城订单里查看
    if (!isLoggedIn.value && !trimmedContact) {
      appStore.showToast('error', '请填写用于查单的手机号或邮箱', 3000)
      return false
    }
    // 支付方式必须来自后台配置，且金额在其单笔限额内
    const option = methodsForAmount(product.price_cny_minor).find((item) => item.value === paymentType)
    if (!option) {
      appStore.showToast('warning', '当前商品暂无可用支付方式，请联系客服处理', 3500)
      return false
    }
    submitting.value = true
    try {
      let payment
      if (isLoggedIn.value) {
        const response = await shopAPI.createOrder({
          product_id: product.id,
          payment_type: option.value,
          return_url: typeof window !== 'undefined' ? window.location.origin + '/shop' : undefined,
          is_mobile: isMobileDevice()
        })
        payment = response.data.payment
        pendingContext = null
      } else {
        const response = await publicShopAPI.createOrder({
          product_id: product.id,
          payment_type: option.value,
          contact: trimmedContact,
          return_url: typeof window !== 'undefined' ? window.location.origin + '/order' : undefined,
          is_mobile: isMobileDevice()
        })
        const result = response.data
        payment = result.payment
        pendingContext = {
          orderNo: result.order_no,
          guestToken: result.guest_token,
          contact: trimmedContact,
          productName: product.name,
          fulfillmentMode: product.fulfillment_mode
        }
        rememberGuestOrder({
          order_no: result.order_no,
          guest_token: result.guest_token,
          contact: trimmedContact,
          product_name: product.name,
          fulfillment_mode: product.fulfillment_mode,
          created_at: Date.now()
        })
      }

      // 下单接口异常时可能不返回 payment，提前兜底，避免后续取 decision.paymentState 抛错
      if (!payment) {
        appStore.showToast('error', '下单未返回支付信息，请稍后重试', 3000)
        return false
      }

      const decision = decidePaymentLaunch(payment, {
        visibleMethod: option.value,
        orderType: 'shop' as OrderType,
        isMobile: isMobileDevice(),
        isWechatBrowser: isWechatBrowser(),
        forceQRCode: alipayForceQRCode.value && option.value === 'alipay',
        stripePopupUrl: '',
        stripeRouteUrl: '',
        airwallexRouteUrl: ''
      })

      if (decision.kind === 'unhandled') {
        appStore.showToast('error', '支付方式暂不可用，请换一个支付方式再试', 3000)
        return false
      }
      if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
        window.location.href = decision.oauth.authorize_url
        return true
      }
      if (decision.kind === 'wechat_jsapi') {
        appStore.showToast('warning', '微信内支付需要跳转处理，请换用支付宝或在浏览器中打开', 4000)
        return false
      }

      payDialog.orderId = decision.paymentState.orderId
      payDialog.qrCode = decision.paymentState.qrCode
      payDialog.expiresAt = decision.paymentState.expiresAt
      payDialog.paymentType = decision.paymentState.paymentType
      payDialog.payUrl = decision.paymentState.payUrl
      payDialog.open = true

      if (decision.kind === 'stripe_popup' || decision.kind === 'redirect_waiting') {
        const win = window.open(decision.paymentState.payUrl, 'paymentPopup', getPaymentPopupFeatures())
        if (!win || win.closed) {
          window.location.href = decision.paymentState.payUrl
        }
      }
      return true
    } catch (error: any) {
      appStore.showToast('error', error?.message || '创建订单失败，请稍后重试', 3000)
      return false
    } finally {
      submitting.value = false
    }
  }

  /** 支付成功后：登录用户去控制台商城订单（可直接提交交付资料），游客去查单页 */
  function goToOrderPage(): void {
    payDialog.open = false
    if (isLoggedIn.value) {
      void router.push('/shop')
      return
    }
    if (!pendingContext) {
      void router.push('/order')
      return
    }
    void router.push({
      path: '/order',
      query: { no: pendingContext.orderNo, contact: pendingContext.contact }
    })
  }

  function closePayDialog(): void {
    payDialog.open = false
    payDialog.qrCode = ''
    payDialog.payUrl = ''
  }

  return {
    isLoggedIn,
    submitting,
    payDialog,
    startCheckout,
    goToOrderPage,
    closePayDialog,
    paymentMethods: guestPaymentMethods,
    methodsLoaded,
    loadPaymentMethods,
    methodsForAmount
  }
}
