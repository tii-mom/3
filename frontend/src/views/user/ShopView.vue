<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6 p-4 sm:p-6 lg:p-8">
      <section class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div v-if="banners.length" class="relative isolate overflow-hidden bg-slate-950">
          <img
            :src="activeBanner.image_url || defaultBannerImage"
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
          <div>
            <p class="mb-3 text-sm font-semibold text-primary-600 dark:text-primary-300">3API 商城</p>
            <h1 class="text-2xl font-bold text-gray-950 dark:text-white sm:text-4xl">购买平台商品，付款后自动发放</h1>
            <p class="mt-3 max-w-2xl text-sm text-gray-600 dark:text-gray-300">精选平台商品，支付成功后自动生成订单记录；推广奖励进入算力公司钱包。</p>
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
          支付暂未开启
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
            <img :src="product.image_url || defaultProductImage" :alt="product.name" class="h-full w-full object-cover transition duration-500 group-hover:scale-105">
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
              <div v-if="product.commission_bps > 0" class="rounded-2xl bg-emerald-50 px-3 py-2 text-right text-xs text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300">
                推广奖励<br><b>{{ (product.commission_bps / 100).toFixed(0) }}%</b>
              </div>
            </div>
            <button
              class="btn-primary w-full justify-center rounded-2xl py-3"
              :disabled="creatingProductID === product.id || !paymentType"
              @click="buy(product)"
            >
              {{ !paymentType ? '支付暂未开启' : creatingProductID === product.id ? '正在创建订单...' : '立即购买' }}
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
          <div v-for="order in orders" :key="order.id" class="flex flex-col gap-2 py-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <div class="font-semibold text-gray-950 dark:text-white">{{ order.snapshot_name }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400">订单 #{{ order.id }} · {{ statusLabel(order.status) }}</div>
            </div>
            <div class="text-sm font-bold text-gray-950 dark:text-white">¥{{ formatMoney(order.snapshot_price_cny_minor) }}</div>
          </div>
        </div>
      </section>

      <div v-if="payDialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
        <div class="w-full max-w-md rounded-3xl bg-white p-6 shadow-2xl dark:bg-dark-900">
          <h3 class="text-lg font-bold text-gray-950 dark:text-white">请完成支付</h3>
          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">支付成功后订单会自动处理，你也可以回到订单列表刷新查看。</p>
          <div v-if="payDialog.qr" class="mt-5 rounded-2xl bg-gray-50 p-4 text-center dark:bg-dark-800">
            <img :src="payDialog.qr" alt="支付二维码" class="mx-auto h-56 w-56 rounded-xl bg-white p-2">
          </div>
          <a v-if="payDialog.url" :href="payDialog.url" target="_blank" rel="noopener noreferrer" class="btn-primary mt-5 w-full justify-center rounded-2xl py-3">打开支付页面</a>
          <button class="btn-secondary mt-3 w-full justify-center rounded-2xl py-3" @click="closePayDialog">我已完成 / 稍后查看</button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { shopAPI, type ShopBanner, type ShopOrder, type ShopProduct } from '@/api/shop'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'

const appStore = useAppStore()
const loading = ref(true)
const banners = ref<ShopBanner[]>([])
const products = ref<ShopProduct[]>([])
const orders = ref<ShopOrder[]>([])
const paymentMethods = ref<string[]>([])
const paymentType = ref('')
const creatingProductID = ref<number | null>(null)
const activeBannerIndex = ref(0)
const payDialog = reactive({ open: false, url: '', qr: '' })

const defaultBannerImage = 'https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=1600&q=80'
const defaultProductImage = 'data:image/svg+xml;utf8,%3Csvg xmlns="http://www.w3.org/2000/svg" width="900" height="600" viewBox="0 0 900 600"%3E%3Cdefs%3E%3ClinearGradient id="g" x1="0" y1="0" x2="1" y2="1"%3E%3Cstop offset="0" stop-color="%23fff7ed"/%3E%3Cstop offset="1" stop-color="%23fed7aa"/%3E%3C/linearGradient%3E%3C/defs%3E%3Crect width="900" height="600" fill="url(%23g)"/%3E%3Ccircle cx="702" cy="108" r="92" fill="%23fb923c" opacity=".20"/%3E%3Ccircle cx="152" cy="480" r="132" fill="%230ea5e9" opacity=".12"/%3E%3Ctext x="72" y="292" font-family="Arial, sans-serif" font-size="64" font-weight="800" fill="%237c2d12"%3E3API%3C/text%3E%3Ctext x="72" y="350" font-family="Arial, sans-serif" font-size="32" fill="%239a3412"%3E%E5%B9%B3%E5%8F%B0%E5%95%86%E5%93%81%3C/text%3E%3C/svg%3E'

const activeBanner = computed(() => banners.value[activeBannerIndex.value] || banners.value[0])

function formatMoney(minor: number): string {
  return (minor / 100).toFixed(2)
}

function productTypeLabel(type: string): string {
  return type === 'platform_usd_balance' ? '平台权益' : '平台商品'
}

function paymentLabel(method: string): string {
  const labels: Record<string, string> = { alipay: '支付宝', wxpay: '微信支付', stripe: '银行卡', easypay: '聚合支付' }
  return labels[method] || method
}

function statusLabel(status: string): string {
  const labels: Record<string, string> = { pending: '待支付', paid: '已支付', fulfilled: '已完成', cancelled: '已取消', refunded: '已退款', failed: '失败' }
  return labels[status] || status
}

function scrollToProduct(productID: number) {
  document.getElementById(`shop-product-${productID}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

async function loadCheckoutMethods() {
  try {
    const res = await paymentAPI.getCheckoutInfo()
    const methods = Object.entries(res.data.methods || {})
      .filter(([, limits]) => limits.available)
      .map(([key]) => key)
    paymentMethods.value = methods
    paymentType.value = methods[0] || ''
  } catch {
    paymentMethods.value = []
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
  }
}

async function buy(product: ShopProduct) {
  if (!paymentType.value) {
    appStore.showToast('warning', '请先选择支付方式', 2500)
    return
  }
  creatingProductID.value = product.id
  try {
    const res = await shopAPI.createOrder({
      product_id: product.id,
      payment_type: paymentType.value,
      return_url: `${window.location.origin}/shop`,
      is_mobile: /Mobile|Android|iPhone|iPad/i.test(navigator.userAgent),
    })
    payDialog.url = res.data.payment.pay_url || ''
    payDialog.qr = res.data.payment.qr_code || ''
    payDialog.open = true
    await loadOrders()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '创建订单失败', 3000)
  } finally {
    creatingProductID.value = null
  }
}

function closePayDialog() {
  payDialog.open = false
  payDialog.url = ''
  payDialog.qr = ''
  void loadOrders()
}

onMounted(loadData)
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
    radial-gradient(circle at top left, rgba(16, 185, 129, 0.24), transparent 34rem),
    linear-gradient(135deg, rgba(15,23,42,1), rgba(2,6,23,1));
}
</style>
