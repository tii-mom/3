<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6 p-4 sm:p-6 lg:p-8">
      <section class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div v-if="banners.length" class="relative isolate min-h-56 overflow-hidden bg-slate-950 sm:min-h-72">
          <img
            v-if="shopImage(activeBanner.image_url)"
            :src="shopImage(activeBanner.image_url)"
            :alt="activeBanner.title"
            class="h-56 w-full object-cover opacity-40 sm:h-72"
          >
          <div class="absolute inset-0 bg-gradient-to-r from-slate-950 via-slate-950/92 to-slate-950/60"></div>
          <div class="absolute inset-0 flex items-center p-6 sm:p-10">
            <div class="max-w-xl rounded-2xl border border-white/15 bg-slate-950/55 p-5 text-white shadow-2xl backdrop-blur-md sm:p-6">
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
        <div v-else class="relative isolate flex min-h-56 items-center overflow-hidden bg-slate-950 p-6 text-white sm:p-10">
          <div class="absolute inset-0 bg-[radial-gradient(circle_at_22%_18%,rgba(251,146,60,0.16),transparent_26rem)]"></div>
          <div class="absolute inset-0 bg-[linear-gradient(135deg,rgba(15,23,42,0.96),rgba(2,6,23,0.88))]"></div>
          <div class="absolute inset-x-0 bottom-0 h-px bg-gradient-to-r from-transparent via-white/25 to-transparent"></div>
          <div class="relative max-w-2xl">
            <p class="mb-3 inline-flex rounded-full border border-orange-200/30 bg-orange-300/15 px-3 py-1 text-sm font-semibold text-orange-100 shadow-sm">3API 商城</p>
            <h1 class="text-2xl font-bold tracking-tight text-white sm:text-4xl">购买平台商品，付款后等待后台处理</h1>
            <p class="mt-3 max-w-xl text-sm font-medium leading-6 text-slate-200 sm:text-base">精选平台商品，支付成功后生成订单；管理员处理发货，推广奖励进入算力公司钱包。</p>
          </div>
        </div>
      </section>

      <section class="grid gap-4 md:grid-cols-[1fr_auto] md:items-end">
        <div>
          <h2 class="text-xl font-bold text-gray-950 dark:text-white">精选商品</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">与官网首页同一套商品。选择商品后直接付款，支付成功后会进入订单，等待管理员处理发货。</p>
        </div>
        <select v-if="paymentMethods.length > 0" v-model="paymentType" class="input-field min-w-40">
          <option v-for="method in paymentMethods" :key="method" :value="method">{{ paymentLabel(method) }}</option>
        </select>
        <div v-else class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-2 text-sm font-semibold text-amber-700 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-300">
          支付暂未开启，请联系管理员
        </div>
      </section>

      <div v-if="catalogTabs.length > 2" class="flex flex-wrap gap-2">
        <button
          v-for="tab in catalogTabs"
          :key="tab.value"
          type="button"
          class="rounded-full px-4 py-2 text-sm font-medium transition"
          :class="catalogTab === tab.value
            ? 'bg-orange-500 text-white shadow-sm'
            : 'border border-gray-200 bg-white text-gray-600 hover:border-orange-200 hover:text-orange-600 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300'"
          @click="catalogTab = tab.value"
        >
          {{ tab.label }}
        </button>
      </div>

      <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <p class="text-sm font-semibold text-gray-500 dark:text-gray-400">售后服务</p>
            <h2 class="mt-1 text-lg font-bold text-gray-950 dark:text-white">购买后需要帮助，可以联系平台客服</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">付款、发货、安装、使用问题，都可以通过下面的客服信息处理。</p>
          </div>
          <div class="rounded-2xl border border-orange-200 bg-white/85 px-4 py-3 text-sm shadow-sm dark:border-dark-700 dark:bg-dark-800">
            <div class="text-xs font-semibold text-gray-500 dark:text-gray-400">客服信息</div>
            <div class="mt-1 whitespace-pre-wrap break-words font-semibold text-gray-950 dark:text-white">{{ afterSalesContact || '管理员暂未填写客服联系方式，请前往后台系统设置补充。' }}</div>
          </div>
        </div>
      </section>

      <div v-if="loading" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <div v-for="i in 6" :key="i" class="h-72 animate-pulse rounded-2xl bg-gray-100 dark:bg-dark-800"></div>
      </div>

      <div v-else-if="products.length === 0" class="rounded-2xl border border-dashed border-gray-300 bg-white p-10 text-center shadow-sm dark:border-dark-600 dark:bg-dark-900">
        <p class="text-base font-semibold text-gray-900 dark:text-white">商城商品正在准备中</p>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">管理员上架商品后会显示在这里。</p>
      </div>

      <div v-else-if="visibleProducts.length === 0" class="rounded-2xl border border-dashed border-gray-300 bg-white p-10 text-center shadow-sm dark:border-dark-600 dark:bg-dark-900">
        <p class="text-base font-semibold text-gray-900 dark:text-white">该分类下暂无商品</p>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">切换到「全部」查看其它商品。</p>
      </div>

      <div v-else class="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="product in visibleProducts"
          :id="`shop-product-${product.id}`"
          :key="product.id"
          class="group relative overflow-hidden rounded-2xl border bg-white shadow-sm transition hover:-translate-y-1 hover:shadow-xl dark:bg-dark-900"
          :class="product.highlight ? 'border-orange-300 ring-1 ring-orange-200 dark:border-orange-500/40 dark:ring-orange-500/20' : 'border-gray-200 dark:border-dark-700'"
        >
          <div class="relative h-44 overflow-hidden">
            <img
              v-if="shopImage(product.image_url)"
              :src="shopImage(product.image_url)"
              :alt="product.name"
              class="h-full w-full object-cover transition duration-500 group-hover:scale-105"
            >
            <div v-else class="shop-placeholder" aria-hidden="true">
              <span class="shop-placeholder__glyph">3</span>
            </div>
            <div class="absolute left-4 top-4 flex flex-wrap gap-2">
              <span class="rounded-full bg-white/90 px-3 py-1 text-xs font-semibold text-gray-900 shadow-sm dark:bg-dark-950/80 dark:text-white">
                {{ fulfillmentLabel(product.fulfillment_mode) }}
              </span>
              <span v-if="product.badge_text" class="rounded-full bg-orange-500 px-3 py-1 text-xs font-semibold text-white shadow-sm">
                {{ product.badge_text }}
              </span>
            </div>
            <span v-if="isSoldOut(product)" class="absolute inset-0 flex items-center justify-center bg-slate-950/55 text-sm font-semibold text-white">暂时缺货</span>
          </div>
          <div class="space-y-4 p-5">
            <div>
              <h3 class="line-clamp-1 text-lg font-bold text-gray-950 dark:text-white">{{ product.name }}</h3>
              <p v-if="product.spec_label" class="mt-1 text-xs font-medium text-orange-600 dark:text-orange-300">{{ product.spec_label }}</p>
              <p class="mt-2 line-clamp-2 min-h-[2.5rem] text-sm leading-5 text-gray-500 dark:text-gray-400">{{ product.description || '付款后等待管理员发货。' }}</p>
              <p v-if="product.delivery_form_hint" class="mt-1 text-xs text-gray-400 dark:text-gray-500">{{ product.delivery_form_hint }}</p>
              <div class="mt-2 flex flex-wrap gap-3 text-xs text-gray-400 dark:text-gray-500">
                <span v-if="product.stock_quantity !== null && product.stock_quantity !== undefined">
                  {{ product.stock_quantity > 0 ? (product.stock_quantity <= 5 ? `仅剩 ${product.stock_quantity} 件` : `库存 ${product.stock_quantity}`) : '已售罄' }}
                </span>
                <span v-if="product.sold_count > 0">已售 {{ product.sold_count }} 份</span>
              </div>
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
              class="w-full justify-center rounded-2xl py-3"
              :class="product.highlight ? 'btn-primary' : 'btn-secondary'"
              :disabled="creatingProductID === product.id || !availablePaymentType(product) || isSoldOut(product)"
              @click="buy(product)"
            >
              {{ isSoldOut(product) ? '暂时缺货' : !availablePaymentType(product) ? '支付暂未开启' : creatingProductID === product.id ? '正在创建订单...' : '立即购买' }}
            </button>
          </div>
        </article>
      </div>

      <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-bold text-gray-950 dark:text-white">我的商城订单</h2>
          <button class="btn-secondary rounded-xl px-3 py-2 text-sm" @click="loadOrders">刷新</button>
        </div>
        <div v-if="orders.some(order => order.status === 'paid' && order.fulfillment_status === 'pending')" class="mb-4 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-200">
          你有待发货的商城订单，管理员处理后会在这里显示发货内容。
        </div>
        <div v-if="orders.length === 0" class="py-6 text-center text-sm text-gray-500 dark:text-gray-400">暂无商城订单</div>
        <div v-else class="divide-y divide-gray-100 dark:divide-dark-700">
          <div v-for="order in orders" :key="order.id" class="flex flex-col gap-3 py-4">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <div class="font-semibold text-gray-950 dark:text-white">{{ order.snapshot_name }}</div>
                <div class="text-xs text-gray-500 dark:text-gray-400">
                  订单 #{{ order.id }}<template v-if="order.order_no"> · {{ order.order_no }}</template> · {{ orderLabel(order) }}
                </div>
                <div v-if="needsDelivery(order)" class="mt-1 text-xs">
                  <span v-if="deliverySubmitted(order)" class="text-emerald-600 dark:text-emerald-300">
                    资料已提交<template v-if="order.delivery_hint">：{{ order.delivery_hint }}</template>
                  </span>
                  <span v-else-if="order.status === 'pending'" class="text-gray-400">支付完成后即可提交{{ deliveryRequirementLabel(order) }}</span>
                  <span v-else class="text-amber-600 dark:text-amber-300">待提交{{ deliveryRequirementLabel(order) }}，提交后开始处理</span>
                </div>
                <div v-if="order.fulfillment_note" class="mt-1 text-xs text-amber-600 dark:text-amber-300">发货内容：{{ order.fulfillment_note }}</div>
              </div>
              <div class="flex items-center gap-3 sm:justify-end">
                <div class="text-sm font-bold text-gray-950 dark:text-white">¥{{ formatMoney(order.snapshot_price_cny_minor) }}</div>
                <button
                  v-if="needsDelivery(order) && order.status !== 'pending'"
                  type="button"
                  class="rounded-xl bg-orange-500 px-3 py-2 text-xs font-semibold text-white transition hover:bg-orange-600"
                  @click="openDeliveryDialog(order)"
                >
                  {{ deliverySubmitted(order) ? '重新提交' : '提交资料' }}
                </button>
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
        </div>

        <div class="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-gray-100 pt-4 dark:border-dark-700">
          <p class="text-xs text-gray-500 dark:text-gray-400">在官网首页以免登录方式下过单？输入订单号和联系方式即可归入这里统一管理。</p>
          <button type="button" class="rounded-xl border border-gray-200 px-3 py-2 text-xs font-semibold text-gray-600 transition hover:border-orange-200 hover:text-orange-600 dark:border-dark-600 dark:text-gray-300" @click="claimDialog.open = true">
            认领订单
          </button>
        </div>
      </section>

      <BaseDialog :show="deliveryDialog.open" title="提交交付资料" @close="deliveryDialog.open = false">
        <DeliveryForm
          v-if="deliveryDialog.open && deliveryDialog.order"
          mode="user"
          :order-id="deliveryDialog.order.id"
          :fulfillment-mode="(deliveryDialog.order.snapshot_fulfillment_mode || 'manual') as FulfillmentMode"
          :submitted="deliverySubmitted(deliveryDialog.order)"
          :hint="deliveryDialog.order.delivery_hint || ''"
          @submitted="onDeliverySubmitted"
        />
      </BaseDialog>

      <BaseDialog :show="claimDialog.open" title="认领订单" width="narrow" @close="claimDialog.open = false">
        <div class="space-y-4">
          <p class="text-sm text-gray-500 dark:text-gray-400">填写在官网首页下单时的订单号与联系方式，订单会归入你的账号。</p>
          <label class="block">
            <span class="mb-1 block text-sm text-gray-600 dark:text-gray-300">订单号</span>
            <input v-model="claimDialog.orderNo" type="text" class="input-field w-full" placeholder="例如 C20260909AB12CD" autocomplete="off">
          </label>
          <label class="block">
            <span class="mb-1 block text-sm text-gray-600 dark:text-gray-300">联系方式</span>
            <input v-model="claimDialog.contact" type="text" class="input-field w-full" placeholder="下单时填写的手机号或邮箱" autocomplete="off">
          </label>
          <div class="flex justify-end gap-3">
            <button type="button" class="btn-secondary rounded-xl px-4 py-2" @click="claimDialog.open = false">取消</button>
            <button type="button" class="btn-primary rounded-xl px-4 py-2" :disabled="claimDialog.submitting" @click="claimOrder">
              {{ claimDialog.submitting ? '认领中...' : '确认认领' }}
            </button>
          </div>
        </div>
      </BaseDialog>

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
import { FULFILLMENT_LABELS, type FulfillmentMode } from '@/api/publicShop'
import { paymentAPI } from '@/api/payment'
import userAPI from '@/api/user'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import DeliveryForm from '@/components/shop/DeliveryForm.vue'
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

// 商品分组与官网首页保持一致：按交付模式划分
const catalogTab = ref<'all' | FulfillmentMode>('all')
const catalogTabs = computed(() => {
  const modes = new Set(products.value.map(item => item.fulfillment_mode || 'manual'))
  const list: { value: 'all' | FulfillmentMode; label: string }[] = [{ value: 'all', label: '全部' }]
  const order: FulfillmentMode[] = ['session_topup', 'account_delivery', 'rental', 'manual']
  order.forEach(mode => {
    if (modes.has(mode)) list.push({ value: mode, label: FULFILLMENT_LABELS[mode] })
  })
  return list
})
const visibleProducts = computed(() =>
  catalogTab.value === 'all' ? products.value : products.value.filter(item => (item.fulfillment_mode || 'manual') === catalogTab.value)
)

// 订单交付资料提交
const deliveryDialog = reactive<{ open: boolean; order: ShopOrder | null }>({ open: false, order: null })

function needsDelivery(order: ShopOrder): boolean {
  const mode = order.snapshot_fulfillment_mode || 'manual'
  return mode !== 'manual' && order.status !== 'cancelled' && order.status !== 'refunded'
}

function deliverySubmitted(order: ShopOrder): boolean {
  return !!order.delivery_submitted_at
}

function deliveryRequirementLabel(order: ShopOrder): string {
  switch (order.snapshot_fulfillment_mode) {
    case 'session_topup': return '登录凭证'
    case 'account_delivery': return '接收邮箱'
    case 'rental': return '接收邮箱与租期'
    default: return '补充信息'
  }
}

function openDeliveryDialog(order: ShopOrder) {
  deliveryDialog.order = order
  deliveryDialog.open = true
}

async function onDeliverySubmitted() {
  deliveryDialog.open = false
  deliveryDialog.order = null
  await loadOrders()
}

// 认领首页免登录订单
const claimDialog = reactive({ open: false, orderNo: '', contact: '', submitting: false })

async function claimOrder() {
  if (!claimDialog.orderNo.trim() || !claimDialog.contact.trim() || claimDialog.submitting) return
  claimDialog.submitting = true
  try {
    await shopAPI.claimOrder({ order_no: claimDialog.orderNo.trim(), contact: claimDialog.contact.trim() })
    appStore.showToast('success', '订单已归入你的账号', 3000)
    claimDialog.open = false
    claimDialog.orderNo = ''
    claimDialog.contact = ''
    await loadOrders()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '认领失败，请核对订单号与联系方式', 3500)
  } finally {
    claimDialog.submitting = false
  }
}

const activeBanner = computed(() => banners.value[activeBannerIndex.value] || banners.value[0])
const paymentMethods = computed(() => Object.keys(paymentMethodLimits.value))
const afterSalesContact = computed(() => appStore.contactInfo?.trim() || '')

function formatMoney(minor: number): string {
  const value = Number(minor || 0) / 100
  return value.toLocaleString('zh-CN', { minimumFractionDigits: 0, maximumFractionDigits: 2 })
}

function shopImage(url?: string | null): string {
  return resolveShopAssetUrl(url)
}

function fulfillmentLabel(mode?: string): string {
  return FULFILLMENT_LABELS[(mode || 'manual') as FulfillmentMode] || '商品'
}

function isSoldOut(product: ShopProduct): boolean {
  return product.stock_quantity !== null && product.stock_quantity !== undefined && product.stock_quantity <= 0
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

function orderLabel(order: ShopOrder): string {
  if (order.status === 'paid' && order.fulfillment_status === 'pending') return '待发货'
  if (order.fulfillment_status === 'fulfilled' || order.status === 'fulfilled') return '已发货'
  return statusLabel(order.status)
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
  void Promise.all([appStore.fetchPublicSettings(), loadData(), loadAffiliateDetail()])
})
</script>

<style scoped>
/* 无商品图时的占位：跟随明暗主题，避免深色模式出现刺眼亮块 */
.shop-placeholder {
  display: grid;
  place-items: center;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #f4f4f5 0%, #e8e8ea 100%);
}


.shop-placeholder__glyph {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: rgba(9, 9, 11, 0.06);
  color: rgba(9, 9, 11, 0.34);
  font-size: 20px;
  font-weight: 700;
}
</style>

<style>
/* 深色覆盖放在**非 scoped** 块中：Vue 的 scoped-CSS 编译器会把
   `:global(.dark) X` 编译成只有 `.dark`，X 被丢弃，导致生产构建里
   深色规则整体失效。与 SettingsView 的处理方式保持一致。 */
.dark .shop-placeholder {
  background: linear-gradient(135deg, #16171b 0%, #0f1013 100%);
}

.dark .shop-placeholder__glyph {
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.34);
}
</style>
