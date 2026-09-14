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
            <p class="mt-3 max-w-xl text-sm font-medium leading-6 text-slate-200 sm:text-base">精选平台商品，支付成功后生成订单；管理员处理发货，推广返点进入你的推广钱包。</p>
          </div>
        </div>
      </section>

      <section class="grid gap-4 md:grid-cols-[1fr_auto] md:items-end">
        <div>
          <h2 class="text-xl font-bold text-gray-950 dark:text-white">精选商品</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">与官网首页同一套商品。选择商品后直接付款，支付成功后会进入订单，等待管理员处理发货。</p>
          <p
            v-if="walletEligible"
            class="mt-2 inline-flex items-center gap-1.5 rounded-full border border-orange-200 bg-orange-50 px-3 py-1 text-xs font-semibold text-orange-700 dark:border-orange-500/20 dark:bg-orange-500/10 dark:text-orange-300"
          >
            返点余额 ¥{{ formatMoney(walletAvailableMinor) }} 可在下单时抵扣
          </p>
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

      <div v-if="loading" class="shop-grid">
        <div v-for="i in 8" :key="i" class="h-80 animate-pulse rounded-2xl bg-gray-100 dark:bg-dark-800"></div>
      </div>

      <div v-else-if="products.length === 0" class="rounded-2xl border border-dashed border-gray-300 bg-white p-10 text-center shadow-sm dark:border-dark-600 dark:bg-dark-900">
        <p class="text-base font-semibold text-gray-900 dark:text-white">商城商品正在准备中</p>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">管理员上架商品后会显示在这里。</p>
      </div>

      <div v-else-if="visibleProducts.length === 0" class="rounded-2xl border border-dashed border-gray-300 bg-white p-10 text-center shadow-sm dark:border-dark-600 dark:bg-dark-900">
        <p class="text-base font-semibold text-gray-900 dark:text-white">该分类下暂无商品</p>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">切换到「全部」查看其它商品。</p>
      </div>

      <div v-else class="shop-grid">
        <div
          v-for="product in visibleProducts"
          :id="`shop-product-${product.id}`"
          :key="product.id"
          class="shop-grid__item"
        >
          <PlanCard
            :product="product"
            :show-commission="!!inviteDetail?.aff_code"
            @select="onPlanSelect"
          />
          <button
            v-if="product.commission_bps > 0 && inviteDetail?.aff_code"
            type="button"
            class="shop-copy-btn"
            @click.stop="copyProductPromotionLink(product)"
          >
            复制推广链接 · 返 {{ (product.commission_bps / 100).toFixed(0) }}%
          </button>
        </div>
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
                <div class="text-right">
                  <div class="text-sm font-bold text-gray-950 dark:text-white">¥{{ formatMoney(order.snapshot_price_cny_minor) }}</div>
                  <div v-if="order.wallet_applied_cny_minor > 0" class="text-xs text-orange-600 dark:text-orange-300">
                    返点抵扣 ¥{{ formatMoney(order.wallet_applied_cny_minor) }}<template v-if="order.payable_cny_minor > 0"> · 实付 ¥{{ formatMoney(order.payable_cny_minor) }}</template>
                  </div>
                </div>
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

      <BaseDialog :show="checkoutDialog.open" title="确认下单" width="narrow" @close="closeCheckout">
        <div v-if="checkoutDialog.open && checkoutProduct" class="space-y-4">
          <div class="flex items-center gap-3 rounded-2xl border border-gray-100 bg-gray-50/70 p-3 dark:border-dark-700 dark:bg-dark-800/60">
            <img
              v-if="shopImage(checkoutProduct.image_url)"
              :src="shopImage(checkoutProduct.image_url)"
              :alt="checkoutProduct.name"
              class="h-14 w-14 shrink-0 rounded-xl object-cover"
            >
            <div class="min-w-0 flex-1">
              <div class="truncate font-semibold text-gray-950 dark:text-white">{{ checkoutProduct.name }}</div>
              <div class="mt-0.5 text-sm font-bold text-gray-950 dark:text-white">¥{{ formatMoney(checkoutPriceMinor) }}</div>
            </div>
          </div>

          <button
            v-if="walletEligible"
            type="button"
            class="flex w-full items-center justify-between gap-3 rounded-2xl border px-4 py-3 text-left transition"
            :class="checkoutDialog.useWallet
              ? 'border-orange-300 bg-orange-50 dark:border-orange-500/30 dark:bg-orange-500/10'
              : 'border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800'"
            @click="checkoutDialog.useWallet = !checkoutDialog.useWallet"
          >
            <span class="flex items-center gap-2.5">
              <span
                class="inline-flex h-5 w-5 items-center justify-center rounded-md border text-xs font-bold"
                :class="checkoutDialog.useWallet ? 'border-orange-500 bg-orange-500 text-white' : 'border-gray-300 text-transparent dark:border-dark-500'"
              >✓</span>
              <span>
                <span class="block text-sm font-semibold text-gray-950 dark:text-white">使用返点余额抵扣</span>
                <span class="block text-xs text-gray-500 dark:text-gray-400">可用 ¥{{ formatMoney(walletAvailableMinor) }}</span>
              </span>
            </span>
            <span class="shrink-0 text-sm font-semibold" :class="checkoutDialog.useWallet ? 'text-orange-600 dark:text-orange-300' : 'text-gray-400 dark:text-gray-500'">
              <template v-if="checkoutDialog.useWallet">-¥{{ formatMoney(walletAppliedMinor) }}</template>
              <template v-else>不使用</template>
            </span>
          </button>

          <div class="space-y-2 rounded-2xl border border-gray-100 bg-white p-4 text-sm dark:border-dark-700 dark:bg-dark-900">
            <div class="flex items-center justify-between text-gray-500 dark:text-gray-400">
              <span>商品金额</span><span>¥{{ formatMoney(checkoutPriceMinor) }}</span>
            </div>
            <div v-if="walletAppliedMinor > 0" class="flex items-center justify-between text-orange-600 dark:text-orange-300">
              <span>返点余额抵扣</span><span>-¥{{ formatMoney(walletAppliedMinor) }}</span>
            </div>
            <div class="flex items-center justify-between border-t border-dashed border-gray-200 pt-2 text-base font-bold text-gray-950 dark:border-dark-700 dark:text-white">
              <span>需支付</span><span>¥{{ formatMoney(checkoutPayableMinor) }}</span>
            </div>
          </div>

          <label v-if="!checkoutFullyPaid" class="block">
            <span class="mb-1 block text-sm text-gray-600 dark:text-gray-300">支付方式</span>
            <select v-model="checkoutDialog.method" class="input-field w-full">
              <option v-for="method in checkoutMethods" :key="method" :value="method">{{ paymentLabel(method) }}</option>
            </select>
          </label>
          <div v-else class="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-semibold text-emerald-700 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-300">
            返点余额足够全额抵扣，下单后无需再付款。
          </div>

          <div class="flex justify-end gap-3">
            <button type="button" class="btn-secondary rounded-xl px-4 py-2" @click="closeCheckout">取消</button>
            <button type="button" class="btn-primary rounded-xl px-4 py-2" :disabled="creatingProductID !== null" @click="confirmCheckout">
              {{ creatingProductID !== null ? '处理中...' : (checkoutFullyPaid ? '确认抵扣下单' : '去支付 ¥' + formatMoney(checkoutPayableMinor)) }}
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
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { resolveShopAssetUrl, shopAPI, type ShopBanner, type ShopOrder, type ShopProduct, type ShopWalletBalance } from '@/api/shop'
import { type FulfillmentMode, publicShopAPI, type PublicCategory, type PublicProduct } from '@/api/publicShop'
import { SHOP_CATEGORIES, type ShopCategory } from '@/constants/shop'
import PlanCard from '@/components/home/PlanCard.vue'
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

// 「我的返点余额」= 推广返点结算进的人民币钱包，可在商城下单时抵扣。
// 与 API 美金额度（平台额度）是两个完全独立的账户，此处只用前者。
const walletBalance = ref<ShopWalletBalance | null>(null)

// 下单确认抽屉：展示可用余额 → 本单抵扣 → 剩余应付，让「组合支付」一眼看清。
const checkoutDialog = reactive<{
  open: boolean
  product: ShopProduct | null
  useWallet: boolean
  method: string
}>({ open: false, product: null, useWallet: false, method: '' })

// 商品分组与官网首页保持一致：按品类（category）划分，品类数据源同为后台配置
const catalogTab = ref<'all' | ShopCategory>('all')
const categories = ref<PublicCategory[]>([])
const catalogTabs = computed(() => {
  const present = new Set(products.value.map(item => (item.category || 'other') as ShopCategory))
  const meta = categories.value.length
    ? categories.value.map(item => ({ value: item.slug as ShopCategory, label: item.label }))
    : SHOP_CATEGORIES.map(item => ({ value: item.value, label: item.label }))

  const list: { value: 'all' | ShopCategory; label: string }[] = [{ value: 'all', label: '全部' }]
  meta.forEach(cat => {
    if (present.has(cat.value)) list.push(cat)
  })

  // 兜底：商品挂在一个已从后台删除的品类上时补一个 tab，
  // 否则这些商品只能在「全部」里看到，点分类反而找不到
  const known = new Set(list.map(item => item.value))
  present.forEach(slug => {
    if (slug !== 'other' && !known.has(slug)) list.push({ value: slug, label: slug })
  })
  return list
})
const visibleProducts = computed(() =>
  catalogTab.value === 'all' ? products.value : products.value.filter(item => (item.category || 'other') === catalogTab.value)
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

/** 渠道可用性以「实际走渠道的金额」为准：抵扣后的应付金额才受单笔限额约束。 */
function amountFitsMethodAmount(method: string, priceMinor: number): boolean {
  const limits = paymentMethodLimits.value[method]
  if (!limits || limits.available === false) return false
  const amount = priceMinor / 100
  if (limits.single_min > 0 && amount < limits.single_min) return false
  if (limits.single_max > 0 && amount > limits.single_max) return false
  return true
}

function amountFitsMethod(product: ShopProduct, method: string): boolean {
  return amountFitsMethodAmount(method, product.price_cny_minor)
}

function availablePaymentType(product: ShopProduct): string {
  if (paymentType.value && amountFitsMethod(product, paymentType.value)) {
    return paymentType.value
  }
  return paymentMethods.value.find(method => amountFitsMethod(product, method)) || ''
}

// ---- 下单确认（组合支付：先扣返点余额，差额走微信/支付宝）----

const checkoutProduct = computed(() => checkoutDialog.product)
const checkoutPriceMinor = computed(() => checkoutDialog.product?.price_cny_minor || 0)
const walletAvailableMinor = computed(() => walletBalance.value?.available_cny_minor || 0)
/** 有余额且推广计划开启时才展示抵扣入口，避免给用户「有返点却用不了」的错觉 */
const walletEligible = computed(() => !!walletBalance.value?.enabled && walletAvailableMinor.value > 0)
const walletAppliedMinor = computed(() => {
  if (!checkoutDialog.useWallet || !walletEligible.value) return 0
  return Math.min(walletAvailableMinor.value, checkoutPriceMinor.value)
})
const checkoutPayableMinor = computed(() => Math.max(checkoutPriceMinor.value - walletAppliedMinor.value, 0))
const checkoutFullyPaid = computed(() => checkoutPayableMinor.value <= 0)
/** 只保留对当前应付金额可用的渠道；全额抵扣时无需支付渠道 */
const checkoutMethods = computed(() => {
  if (checkoutFullyPaid.value) return []
  return paymentMethods.value.filter(method => amountFitsMethodAmount(method, checkoutPayableMinor.value))
})
const checkoutEffectiveMethod = computed(() => {
  if (checkoutFullyPaid.value) return ''
  if (checkoutDialog.method && amountFitsMethodAmount(checkoutDialog.method, checkoutPayableMinor.value)) {
    return checkoutDialog.method
  }
  return checkoutMethods.value[0] || ''
})

// 抵扣额度变化会让原选中的渠道不再适用（金额落到限额之外），这里自动纠正，
// 保证「确认下单」时选的渠道一定可用。
watch(checkoutPayableMinor, () => {
  if (checkoutFullyPaid.value) return
  if (!checkoutDialog.method || !amountFitsMethodAmount(checkoutDialog.method, checkoutPayableMinor.value)) {
    checkoutDialog.method = checkoutMethods.value[0] || ''
  }
})

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

/** 拉取可用于抵扣的人民币返点余额；拉不到时置空（不展示抵扣入口，不阻塞下单）。 */
async function loadWalletBalance() {
  try {
    const res = await shopAPI.walletBalance()
    walletBalance.value = res.data || null
  } catch {
    walletBalance.value = null
  }
}

async function loadData() {
  loading.value = true
  try {
    // 商品/轮播是页面渲染数据；支付方式与订单列表是独立副作用加载，拆开更清晰
    const [bannerRes, productRes, categoryRes] = await Promise.all([
      shopAPI.listBanners(),
      shopAPI.listProducts(),
      publicShopAPI.listCategories().catch(() => null)
    ])
    banners.value = bannerRes.data || []
    products.value = productRes.data || []
    categories.value = categoryRes?.data || []
    // 当前选中的品类若已从后台下线，回退到「全部」，避免停在空白分类上
    if (catalogTab.value !== 'all' && !catalogTabs.value.some(item => item.value === catalogTab.value)) {
      catalogTab.value = 'all'
    }
    await Promise.all([loadCheckoutMethods(), loadOrders(), loadWalletBalance()])
  } catch (error: any) {
    appStore.showToast('error', error?.message || '商城加载失败', 3000)
  } finally {
    loading.value = false
    await scrollToQueryProduct()
  }
}

/** PlanCard 的 select 事件回传 PublicProduct（首页/商城共用卡片的类型），
 *  而 buy() 需要 ShopProduct；运行时 product 本就来自商城列表，做一次收窄转换。 */
function onPlanSelect(product: PublicProduct) {
  openCheckout(product as unknown as ShopProduct)
}

/** 打开下单确认：先让用户看清「余额抵扣多少、还要付多少」，再决定支付方式。 */
function openCheckout(product: ShopProduct) {
  // 防重入：下单是异步的，快速连点会并发创建多笔订单（creatingProductID 非空即视为进行中）
  if (creatingProductID.value !== null) return
  checkoutDialog.product = product
  // 有余额时默认勾选抵扣（防呆：多数用户就是想把返点花掉）
  checkoutDialog.useWallet = walletEligible.value
  checkoutDialog.method = availablePaymentType(product) || paymentMethods.value[0] || ''
  checkoutDialog.open = true
}

function closeCheckout() {
  checkoutDialog.open = false
  checkoutDialog.product = null
  checkoutDialog.useWallet = false
  checkoutDialog.method = ''
}

async function confirmCheckout() {
  const product = checkoutDialog.product
  if (!product || creatingProductID.value !== null) return
  const method = checkoutEffectiveMethod.value
  if (!checkoutFullyPaid.value && !method) {
    appStore.showToast('warning', '当前应付金额暂无可用支付方式，请调整支付渠道或联系管理员', 3000)
    return
  }
  // 选了某个支付渠道但该渠道不适用于应付金额时显式提示，避免用户困惑
  if (!checkoutFullyPaid.value && paymentType.value && method !== paymentType.value) {
    appStore.showToast('warning', '所选支付方式不适用于该金额，已自动切换为可用方式', 3000)
  }
  creatingProductID.value = product.id
  const visibleMethod = method ? (normalizeVisibleMethod(method) || method) : ''
  const forceQRCode = !!(alipayForceQRCode.value && visibleMethod === 'alipay')
  const orderType: OrderType = 'shop'
  try {
    const res = await shopAPI.createOrder({
      product_id: product.id,
      payment_type: visibleMethod,
      return_url: `${window.location.origin}/payment/result`,
      is_mobile: forceQRCode ? false : isMobileDevice(),
      use_wallet: checkoutDialog.useWallet,
    })
    // 全额抵扣：没有外部支付环节，订单已进入交付，直接关抽屉刷新订单。
    if (res.data?.fully_paid_by_wallet) {
      closeCheckout()
      appStore.showToast('success', '已用返点余额全额抵扣，订单已提交', 3000)
      await Promise.all([loadOrders(), loadWalletBalance()])
      return
    }
    const payment = res.data?.payment
    if (!payment) {
      appStore.showToast('error', '下单未返回支付信息，请稍后重试', 3000)
      return
    }
    closeCheckout()
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
  void Promise.all([loadOrders(), loadWalletBalance()])
}

function handlePaymentSuccess() {
  void Promise.all([loadOrders(), loadWalletBalance()])
}

function handlePaymentSettled() {
  void Promise.all([loadOrders(), loadWalletBalance()])
}

onMounted(() => {
  void Promise.all([appStore.fetchPublicSettings(), loadData(), loadAffiliateDetail()])
})
</script>

<style scoped>
/* 商品网格：与官网首页 #plans 同一套 4 列自适配栅格（228px 起） */
.shop-grid {
  display: grid;
  gap: 20px;
  grid-template-columns: repeat(auto-fit, minmax(228px, 1fr));
}

.shop-grid__item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 卡片下方的推广复制入口：保留商城登录态的返佣能力，
   不在 PlanCard 内，避免污染首页的免登录卡片。 */
.shop-copy-btn {
  width: 100%;
  border-radius: 9px;
  border: 1px solid rgba(9, 9, 11, 0.12);
  background: #fff;
  padding: 9px 12px;
  font-size: 13px;
  font-weight: 500;
  color: #56565f;
  cursor: pointer;
  transition: background 160ms ease-out, border-color 160ms ease-out;
}

.shop-copy-btn:hover {
  background: #f6f6f7;
  border-color: rgba(9, 9, 11, 0.2);
}

.dark .shop-copy-btn {
  border-color: rgba(255, 255, 255, 0.14);
  background: #18181b;
  color: #a1a1aa;
}

.dark .shop-copy-btn:hover {
  background: #232429;
  border-color: rgba(255, 255, 255, 0.22);
}
</style>
