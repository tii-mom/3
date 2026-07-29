<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6 p-4 sm:p-6 lg:p-8">
      <section class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div v-if="banners.length" class="relative isolate overflow-hidden bg-slate-950">
          <img
            :src="shopImage(activeBanner.image_url) || defaultBannerImage"
            :alt="activeBanner.title"
            class="h-56 w-full object-cover opacity-45 sm:h-72"
          >
          <div class="absolute inset-0 bg-gradient-to-r from-slate-950 via-slate-950/90 to-slate-950/55"></div>
          <div class="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(251,146,60,0.28),transparent_28rem)]"></div>
          <div class="absolute inset-0 flex items-center p-6 sm:p-10">
            <div class="max-w-xl rounded-3xl border border-white/15 bg-slate-950/55 p-5 text-white shadow-2xl backdrop-blur-md sm:p-6">
              <p class="mb-3 inline-flex rounded-full bg-white/15 px-3 py-1 text-xs font-medium text-orange-100 backdrop-blur">3API 商城</p>
              <h1 class="text-2xl font-bold tracking-tight sm:text-4xl">{{ activeBanner.title }}</h1>
              <p class="mt-3 text-sm leading-6 text-white/90 sm:text-base">{{ activeBanner.subtitle }}</p>
              <button
                v-if="activeBanner.product_id"
                class="mt-5 rounded-full bg-white px-5 py-2.5 text-sm font-semibold text-gray-950 shadow-lg transition hover:-translate-y-0.5 hover:bg-orange-50"
                @click="scrollToProduct(activeBanner.product_id)"
              >
                {{ activeBanner.button_text || '立即查看' }}
              </button>
            </div>
          </div>
        </div>
        <div v-else class="shop-hero-empty">
          <div class="max-w-2xl">
            <p class="mb-3 inline-flex rounded-full bg-orange-100 px-3 py-1 text-sm font-semibold text-orange-700 dark:bg-orange-400/15 dark:text-orange-100">3API 商城</p>
            <h1 class="text-2xl font-bold text-gray-950 dark:text-white sm:text-4xl">购买平台商品，付款后自动发放</h1>
            <p class="mt-3 text-sm font-medium leading-6 text-gray-700 dark:text-gray-100">精选平台商品，支付成功后自动生成订单记录；推广奖励进入算力公司钱包。</p>
          </div>
        </div>
      </section>

      <section class="grid gap-4 md:grid-cols-[1fr_auto] md:items-end">
        <div>
          <h2 class="text-xl font-bold text-gray-950 dark:text-white">精选商品</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">选择商品后直接付款，支付成功后自动到账或完成记录。</p>
        </div>
        <select v-if="paymentMethods.length > 0" v-model="paymentType" class="input-field min-w-40">
          <option v-for="method in paymentMethods" :key="method" :value="method">{{ paymentLabel(method) }}</option>
        </select>
        <div v-else class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-2 text-sm font-semibold text-amber-700 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-300">
          支付暂未开启，请联系管理员
        </div>
      </section>

      <div v-if="loading" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <div v-for="i in 6" :key="i" class="h-72 animate-pulse rounded-3xl bg-gray-100 dark:bg-dark-800"></div>
      </div>

      <div v-else-if="products.length === 0" class="rounded-3xl border border-dashed border-gray-300 bg-white p-10 text-center shadow-sm dark:border-dark-600 dark:bg-dark-900">
        <p class="text-base font-semibold text-gray-900 dark:text-white">商城商品正在准备中</p>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">管理员上架商品后会显示在这里。</p>
      </div>

      <div v-else class="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="product in products"
          :id="`shop-product-${product.id}`"
          :key="product.id"
          class="group overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm transition hover:-translate-y-1 hover:shadow-xl dark:border-dark-700 dark:bg-dark-900"
        >
          <div class="relative h-44 overflow-hidden bg-gradient-to-br from-orange-50 via-white to-slate-100 dark:from-dark-800 dark:via-dark-800 dark:to-dark-900">
            <img :src="shopImage(product.image_url) || defaultProductImage" :alt="product.name" class="h-full w-full object-cover transition duration-500 group-hover:scale-105">
            <span class="absolute left-4 top-4 rounded-full bg-white/90 px-3 py-1 text-xs font-semibold text-gray-900 shadow-sm dark:bg-dark-950/80 dark:text-white">
              {{ productTypeLabel(product.product_type) }}
            </span>
          </div>
          <div class="space-y-4 p-5">
            <div>
              <h3 class="line-clamp-1 text-lg font-bold text-gray-950 dark:text-white">{{ product.name }}</h3>
              <p class="mt-2 line-clamp-2 min-h-[2.5rem] text-sm leading-5 text-gray-500 dark:text-gray-400">{{ product.description || '付款后自动处理。' }}</p>
            </div>
            <div class="flex items-end justify-between gap-3">
              <div>
                <div class="text-2xl font-black text-gray-950 dark:text-white">¥{{ formatMoney(product.price_cny_minor) }}</div>
                <div v-if="product.original_price_cny_minor > product.price_cny_minor" class="text-xs text-gray-400 line-through">¥{{ formatMoney(product.original_price_cny_minor) }}</div>
              </div>
              <div v-if="product.commission_bps > 0" class="flex items-center gap-2 rounded-2xl bg-emerald-50 px-3 py-2 text-xs text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-200">
                <div class="text-right">
                  推广奖励<br><b>{{ (product.commission_bps / 100).toFixed(0) }}%</b>
                </div>
                <button
                  type="button"
                  class="rounded-xl bg-white px-2.5 py-1.5 text-xs font-bold text-emerald-700 shadow-sm transition hover:bg-emerald-100 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-emerald-400/15 dark:text-emerald-100 dark:hover:bg-emerald-400/25"
                  :disabled="!inviteDetail?.aff_code"
                  @click.stop="copyProductPromotionLink(product)"
                >
                  复制链接
                </button>
              </div>
            </div>
            <button
              class="btn-primary w-full justify-center rounded-2xl py-3"
              :disabled="creatingProductID === product.id || !availablePaymentType(product)"
              @click="buy(product)"
            >
              {{ !availablePaymentType(product) ? '支付暂未开启' : creatingProductID === product.id ? '正在创建订单...' : '立即购买' }}
            </button>
          </div>
        </article>
      </div>

      <section class="rounded-3xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-bold text-gray-950 dark:text-white">我的商城订单</h2>
          <button class="btn-secondary rounded-xl px-3 py-2 text-sm" @click="loadOrders">刷新</button>
        </div>
        <div v-if="orders.length === 0" class="py-6 text-center text-sm text-gray-500 dark:text-gray-400">暂无商城订单</div>
        <div v-else class="divide-y divide-gray-100 dark:divide-dark-700">
          <div v-for="order in orders" :key="order.id" class="flex flex-col gap-3 py-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <div class="font-semibold text-gray-950 dark:text-white">{{ order.snapshot_name }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400">订单 #{{ order.id }} · {{ statusLabel(order.status) }}</div>
            </div>
            <div class="flex items-center gap-3 sm:justify-end">
              <div class="text-sm font-bold text-gray-950 dark:text-white">¥{{ formatMoney(order.snapshot_price_cny_minor) }}</div>
              <button
                v-if="canCancelOrder(order)"
                type="button"
                class="rounded-xl border border-gray-200 px-3 py-2 text-xs font-semibold text-gray-600 transition hover:border-red-200 hover:bg-red-50 hover:text-red-600 disabled:cursor-not-allowed disabled:opacity-60 dark:border-dark-600 dark:text-gray-300 dark:hover:border-red-500/30 dark:hover:bg-red-500/10 dark:hover:text-red-200"
                :disabled="cancellingOrderID === order.id"
                @click="cancelShopPaymentOrder(order)"
              >
                {{ cancellingOrderID === order.id ? '取消中...' : '取消订单' }}
              </button>
            </div>
          </div>
        </div>
      </section>

      <BaseDialog :show="payDialog.open" title="请完成支付" width="narrow" @close="closePayDialog">
        <PaymentStatusPanel
          v-if="payDialog.open"
          :order-id="payDialog.orderId"
          :qr-code="payDialog.qrCode"
          :expires-at="payDialog.expiresAt"
          :payment-type="payDialog.paymentType"
          :pay-url="payDialog.payUrl"
          order-type="shop"
          @done="closePayDialog"
          @success="handlePaymentSuccess"
          @settled="handlePaymentSettled"
        />
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { resolveShopAssetUrl, shopAPI, type ShopBanner, type ShopOrder, type ShopProduct } from '@/api/shop'
import { paymentAPI } from '@/api/payment'
import userAPI from '@/api/user'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import { decidePaymentLaunch, getVisibleMethods, normalizeVisibleMethod } from '@/components/payment/paymentFlow'
import { useClipboard } from '@/composables/useClipboard'
import type { MethodLimit, OrderType } from '@/types/payment'
import type { UserAffiliateDetail } from '@/types'

const appStore = useAppStore()
const route = useRoute()
const router = useRouter()
const { copyToClipboard } = useClipboard()
const loading = ref(true)
const banners = ref<ShopBanner[]>([])
const products = ref<ShopProduct[]>([])
const orders = ref<ShopOrder[]>([])
const paymentMethodLimits = ref<Record<string, MethodLimit>>({})
const paymentType = ref('')
const creatingProductID = ref<number | null>(null)
const cancellingOrderID = ref<number | null>(null)
const activeBannerIndex = ref(0)
const inviteDetail = ref<UserAffiliateDetail | null>(null)
const alipayForceQRCode = ref(false)
const payDialog = reactive({
  open: false,
  orderId: 0,
  qrCode: '',
  expiresAt: '',
  paymentType: '',
  payUrl: '',
})

const defaultBannerImage = 'https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=1600&q=80'
const defaultProductImage = 'data:image/svg+xml;utf8,%3Csvg xmlns="http://www.w3.org/2000/svg" width="900" height="600" viewBox="0 0 900 600"%3E%3Cdefs%3E%3ClinearGradient id="g" x1="0" y1="0" x2="1" y2="1"%3E%3Cstop offset="0" stop-color="%23fff7ed"/%3E%3Cstop offset="1" stop-color="%23fed7aa"/%3E%3C/linearGradient%3E%3C/defs%3E%3Crect width="900" height="600" fill="url(%23g)"/%3E%3Ccircle cx="702" cy="108" r="92" fill="%23fb923c" opacity=".20"/%3E%3Ccircle cx="152" cy="480" r="132" fill="%230ea5e9" opacity=".12"/%3E%3Ctext x="72" y="292" font-family="Arial, sans-serif" font-size="64" font-weight="800" fill="%237c2d12"%3E3API%3C/text%3E%3Ctext x="72" y="350" font-family="Arial, sans-serif" font-size="32" fill="%239a3412"%3E%E5%B9%B3%E5%8F%B0%E5%95%86%E5%93%81%3C/text%3E%3C/svg%3E'

const activeBanner = computed(() => banners.value[activeBannerIndex.value] || banners.value[0])
const paymentMethods = computed(() => Object.keys(paymentMethodLimits.value))

function formatMoney(minor: number): string {
  return (minor / 100).toFixed(2)
}

function shopImage(url?: string | null): string {
  return resolveShopAssetUrl(url)
}

function productTypeLabel(type: string): string {
  return type === 'platform_usd_balance' ? '平台权益' : '平台商品'
}

function paymentLabel(method: string): string {
  const label = paymentMethodLimits.value[method]?.display_name
  if (label) return label
  const labels: Record<string, string> = { alipay: '支付宝', wxpay: '微信支付', stripe: '银行卡', airwallex: '空中云汇', easypay: '聚合支付' }
  return labels[method] || method
}

function isMobileDevice(): boolean {
  if (typeof window === 'undefined') return false
  return /Mobile|Android|iPhone|iPad/i.test(window.navigator.userAgent)
}

function isWechatBrowser(): boolean {
  if (typeof window === 'undefined') return false
  return /MicroMessenger/i.test(window.navigator.userAgent)
}

function statusLabel(status: string): string {
  const labels: Record<string, string> = { pending: '待支付', paid: '已支付', fulfilled: '已完成', cancelled: '已取消', refunded: '已退款', failed: '失败' }
  return labels[status] || status
}

function canCancelOrder(order: ShopOrder): boolean {
  return order.status === 'pending' && !!order.payment_order_id
}

function scrollToProduct(productID: number) {
  document.getElementById(`shop-product-${productID}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

function amountFitsMethod(product: ShopProduct, method: string): boolean {
  const limits = paymentMethodLimits.value[method]
  if (!limits || limits.available === false) return false
  const amount = product.price_cny_minor / 100
  if (limits.single_min > 0 && amount < limits.single_min) return false
  if (limits.single_max > 0 && amount > limits.single_max) return false
  return true
}

function availablePaymentType(product: ShopProduct): string {
  if (paymentType.value && amountFitsMethod(product, paymentType.value)) {
    return paymentType.value
  }
  return paymentMethods.value.find(method => amountFitsMethod(product, method)) || ''
}

function buildProductPromotionLink(product: ShopProduct): string {
  const code = inviteDetail.value?.aff_code?.trim()
  if (!code || typeof window === 'undefined') return ''
  const url = new URL('/shop', window.location.origin || 'https://3api.shop')
  url.searchParams.set('product', String(product.id))
  url.searchParams.set('aff', code)
  return url.toString()
}

async function copyProductPromotionLink(product: ShopProduct) {
  const link = buildProductPromotionLink(product)
  if (!link) {
    appStore.showToast('warning', '推广码暂未获取，请稍后再试', 2500)
    return
  }
  await copyToClipboard(link, '推广链接已复制')
}

async function scrollToQueryProduct() {
  const raw = Array.isArray(route.query.product) ? route.query.product[0] : route.query.product
  const id = Number(raw)
  if (!Number.isFinite(id) || id <= 0) return
  await nextTick()
  scrollToProduct(id)
}

async function loadCheckoutMethods() {
  try {
    const res = await paymentAPI.getCheckoutInfo()
    const visible = getVisibleMethods(res.data.methods || {})
    alipayForceQRCode.value = !!res.data.alipay_force_qrcode
    paymentMethodLimits.value = Object.fromEntries(
      Object.entries(visible).filter(([, limits]) => limits?.available !== false)
    )
    if (!paymentType.value || !paymentMethodLimits.value[paymentType.value]) {
      paymentType.value = paymentMethods.value[0] || ''
    }
  } catch {
    alipayForceQRCode.value = false
    paymentMethodLimits.value = {}
    paymentType.value = ''
  }
}

async function loadOrders() {
  try {
    const res = await shopAPI.myOrders({ page: 1, page_size: 10 })
    orders.value = res.data.items || []
  } catch {
    orders.value = []
  }
}

async function loadData() {
  loading.value = true
  try {
    const [bannerRes, productRes] = await Promise.all([shopAPI.listBanners(), shopAPI.listProducts(), loadCheckoutMethods(), loadOrders()])
    banners.value = bannerRes.data || []
    products.value = productRes.data || []
  } catch (error: any) {
    appStore.showToast('error', error?.message || '商城加载失败', 3000)
  } finally {
    loading.value = false
    await scrollToQueryProduct()
  }
}

async function buy(product: ShopProduct) {
  const method = availablePaymentType(product)
  if (!method) {
    appStore.showToast('warning', '当前商品暂无可用支付方式，请联系管理员开启支付渠道', 3000)
    return
  }
  creatingProductID.value = product.id
  const visibleMethod = normalizeVisibleMethod(method) || method
  const forceQRCode = !!(alipayForceQRCode.value && visibleMethod === 'alipay')
  const orderType: OrderType = 'shop'
  try {
    const res = await shopAPI.createOrder({
      product_id: product.id,
      payment_type: visibleMethod,
      return_url: `${window.location.origin}/payment/result`,
      is_mobile: forceQRCode ? false : isMobileDevice(),
    })
    const payment = res.data.payment
    const stripeMethod = visibleMethod === 'stripe'
      ? ''
      : visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = payment.client_secret && visibleMethod !== 'airwallex'
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(payment.order_id),
          client_secret: payment.client_secret,
          method: stripeMethod || undefined,
          resume_token: payment.resume_token || undefined,
        },
      }).href
      : ''
    const airwallexRouteUrl = payment.client_secret && payment.intent_id
      ? router.resolve({
        path: '/payment/airwallex',
        query: {
          order_id: String(payment.order_id),
          out_trade_no: payment.out_trade_no || undefined,
          resume_token: payment.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(payment, {
      visibleMethod,
      orderType,
      isMobile: isMobileDevice(),
      isWechatBrowser: isWechatBrowser(),
      forceQRCode,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
      airwallexRouteUrl,
    })

    if (decision.kind === 'unhandled') {
      appStore.showToast('error', '支付方式暂不可用，请换一个支付方式再试', 3000)
      return
    }
    if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
      window.location.href = decision.oauth.authorize_url
      return
    }
    if (decision.kind === 'wechat_jsapi') {
      appStore.showToast('warning', '当前微信内支付需要跳转处理，请换用支付宝或在浏览器中打开商城', 4000)
      return
    }

    payDialog.orderId = decision.paymentState.orderId
    payDialog.qrCode = decision.paymentState.qrCode
    payDialog.expiresAt = decision.paymentState.expiresAt
    payDialog.paymentType = decision.paymentState.paymentType
    payDialog.payUrl = decision.paymentState.payUrl
    payDialog.open = true

    if (decision.kind === 'stripe_route' || decision.kind === 'airwallex_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'stripe_popup' || decision.kind === 'redirect_waiting') {
      const win = window.open(decision.paymentState.payUrl, 'paymentPopup', getPaymentPopupFeatures())
      if (!win || win.closed) {
        window.location.href = decision.paymentState.payUrl
        return
      }
    }
    await loadOrders()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '创建订单失败', 3000)
  } finally {
    creatingProductID.value = null
  }
}

async function loadAffiliateDetail() {
  try {
    inviteDetail.value = await userAPI.getAffiliateDetail()
  } catch {
    inviteDetail.value = null
  }
}

async function cancelShopPaymentOrder(order: ShopOrder) {
  if (!order.payment_order_id || cancellingOrderID.value) return
  cancellingOrderID.value = order.id
  try {
    await paymentAPI.cancelOrder(order.payment_order_id)
    appStore.showToast('success', '订单已取消', 2200)
    await loadOrders()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '取消订单失败，请刷新后再试', 3000)
  } finally {
    cancellingOrderID.value = null
  }
}

function closePayDialog() {
  payDialog.open = false
  payDialog.orderId = 0
  payDialog.qrCode = ''
  payDialog.expiresAt = ''
  payDialog.paymentType = ''
  payDialog.payUrl = ''
  void loadOrders()
}

function handlePaymentSuccess() {
  void loadOrders()
}

function handlePaymentSettled() {
  void loadOrders()
}

onMounted(() => {
  void Promise.all([loadData(), loadAffiliateDetail()])
})
</script>

<style scoped>
.shop-hero-empty {
  display: flex;
  min-height: 14rem;
  align-items: center;
  background:
    radial-gradient(circle at top left, rgba(16, 185, 129, 0.18), transparent 34rem),
    linear-gradient(135deg, rgba(255,255,255,1), rgba(243,244,246,1));
  padding: 2rem;
}

:global(.dark) .shop-hero-empty {
  background:
    radial-gradient(circle at top left, rgba(251, 146, 60, 0.22), transparent 30rem),
    linear-gradient(135deg, rgb(15, 23, 42), rgb(2, 6, 23));
}
</style>
