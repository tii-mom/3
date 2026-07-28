<template>
  <AppLayout>
  <div class="space-y-6 p-4 sm:p-6 lg:p-8">
    <section class="rounded-3xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p class="text-sm font-semibold text-primary-600 dark:text-primary-300">商城管理</p>
          <h1 class="mt-1 text-2xl font-bold text-gray-950 dark:text-white">商品上架、轮播展示、订单查看</h1>
          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">适合售卖平台商品；推广佣金自动进入算力公司钱包。</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="goBack">返回后台</button>
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="router.push('/shop')">预览商城</button>
          <button v-if="tab === 'products'" class="btn-primary rounded-2xl px-4 py-2.5" @click="openProductDialog()">新增商品</button>
          <button v-else-if="tab === 'banners'" class="btn-primary rounded-2xl px-4 py-2.5" @click="openBannerDialog()">新增轮播</button>
        </div>
      </div>
      <div class="mt-6 flex gap-2 overflow-x-auto">
        <button v-for="item in tabs" :key="item.key" class="rounded-2xl px-4 py-2 text-sm font-semibold transition" :class="tab === item.key ? 'bg-gray-950 text-white dark:bg-white dark:text-gray-950' : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-800 dark:text-gray-300'" @click="tab = item.key">
          {{ item.label }}
        </button>
      </div>
    </section>

    <section v-if="tab === 'products'" class="rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
          <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
            <tr>
              <th class="px-5 py-3">商品</th>
              <th class="px-5 py-3">价格</th>
              <th class="px-5 py-3">类型</th>
              <th class="px-5 py-3">佣金</th>
              <th class="px-5 py-3">状态</th>
              <th class="px-5 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="product in products" :key="product.id">
              <td class="px-5 py-4">
                <div class="flex items-center gap-3">
                  <img :src="product.image_url || defaultProductImage" class="h-12 w-12 rounded-2xl object-cover" alt="">
                  <div>
                    <div class="font-semibold text-gray-950 dark:text-white">{{ product.name }}</div>
                    <div class="line-clamp-1 text-xs text-gray-500">{{ product.description }}</div>
                  </div>
                </div>
              </td>
              <td class="px-5 py-4 font-semibold">¥{{ money(product.price_cny_minor) }}</td>
              <td class="px-5 py-4">{{ typeLabel(product.product_type) }}</td>
              <td class="px-5 py-4">{{ (product.commission_bps / 100).toFixed(0) }}%</td>
              <td class="px-5 py-4"><span :class="statusClass(product.status)" class="rounded-full px-2.5 py-1 text-xs font-semibold">{{ statusLabel(product.status) }}</span></td>
              <td class="px-5 py-4 text-right">
                <button class="btn-secondary mr-2 rounded-xl px-3 py-1.5 text-xs" @click="openProductDialog(product)">编辑</button>
                <button class="rounded-xl px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 dark:hover:bg-red-500/10" @click="removeProduct(product)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section v-else-if="tab === 'banners'" class="grid gap-4 lg:grid-cols-2">
      <article v-for="banner in banners" :key="banner.id" class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <img :src="banner.image_url || defaultBannerImage" class="h-40 w-full object-cover" alt="">
        <div class="p-5">
          <div class="flex items-start justify-between gap-3">
            <div>
              <h3 class="font-bold text-gray-950 dark:text-white">{{ banner.title }}</h3>
              <p class="mt-1 text-sm text-gray-500">{{ banner.subtitle }}</p>
            </div>
            <span class="rounded-full px-2.5 py-1 text-xs font-semibold" :class="banner.enabled ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300' : 'bg-gray-100 text-gray-500 dark:bg-dark-800'">
              {{ banner.enabled ? '展示中' : '已关闭' }}
            </span>
          </div>
          <div class="mt-4 flex justify-end gap-2">
            <button class="btn-secondary rounded-xl px-3 py-1.5 text-xs" @click="openBannerDialog(banner)">编辑</button>
            <button class="rounded-xl px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 dark:hover:bg-red-500/10" @click="removeBanner(banner)">删除</button>
          </div>
        </div>
      </article>
    </section>

    <section v-else class="rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
          <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
            <tr>
              <th class="px-5 py-3">订单</th>
              <th class="px-5 py-3">用户</th>
              <th class="px-5 py-3">金额</th>
              <th class="px-5 py-3">状态</th>
              <th class="px-5 py-3">佣金</th>
              <th class="px-5 py-3">时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="order in orders" :key="order.id">
              <td class="px-5 py-4">
                <div class="font-semibold text-gray-950 dark:text-white">#{{ order.id }} {{ order.snapshot_name }}</div>
                <div class="text-xs text-gray-500">支付订单：{{ order.payment_order_id || '-' }}</div>
              </td>
              <td class="px-5 py-4">{{ order.user_email || order.user_id }}</td>
              <td class="px-5 py-4 font-semibold">¥{{ money(order.snapshot_price_cny_minor) }}</td>
              <td class="px-5 py-4">{{ orderStatusLabel(order.status) }}</td>
              <td class="px-5 py-4">{{ order.snapshot_commission_bps > 0 ? `${(order.snapshot_commission_bps / 100).toFixed(0)}%` : '无' }}</td>
              <td class="px-5 py-4 text-xs text-gray-500">{{ formatDate(order.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <div v-if="productDialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/55 p-4 backdrop-blur-sm">
      <form class="shop-modal max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-3xl bg-white p-6 shadow-2xl dark:bg-dark-900" @submit.prevent="saveProduct">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-sm font-semibold text-primary-600 dark:text-primary-300">商品管理</p>
            <h3 class="mt-1 text-xl font-bold text-gray-950 dark:text-white">{{ productDialog.id ? '编辑商品' : '新增商品' }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">填写商品展示、价格、库存和推广佣金，保存后即可在商城展示。</p>
          </div>
          <button type="button" class="rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-800 dark:hover:text-white" aria-label="关闭" @click="productDialog.open = false">✕</button>
        </div>

        <div class="mt-6 grid gap-5 lg:grid-cols-[1fr_18rem]">
          <div class="space-y-4">
            <label class="shop-field">
              <span>商品名称</span>
              <input v-model="productForm.name" class="input-field w-full" placeholder="例如：ChatGPT 桌面端安装服务" required>
            </label>
            <label class="shop-field">
              <span>商品描述</span>
              <textarea v-model="productForm.description" class="input-field min-h-28 w-full" placeholder="写清楚用户购买后能获得什么服务或商品。"></textarea>
            </label>
            <div class="grid gap-4 sm:grid-cols-2">
              <label class="shop-field">
                <span>售卖金额（元）</span>
                <input v-model.number="productPrice" type="number" min="0.01" step="0.01" class="input-field w-full" required>
              </label>
              <label class="shop-field">
                <span>划线原价（元）</span>
                <input v-model.number="productOriginalPrice" type="number" min="0" step="0.01" class="input-field w-full" placeholder="可不填">
              </label>
              <label class="shop-field">
                <span>推广佣金（%）</span>
                <input v-model.number="productCommissionPercent" type="number" min="0" max="100" step="1" class="input-field w-full">
              </label>
              <label class="shop-field">
                <span>库存（留空不限）</span>
                <input v-model.number="productForm.stock_quantity" type="number" min="0" class="input-field w-full" placeholder="不限">
              </label>
              <label class="shop-field">
                <span>状态</span>
                <select v-model="productForm.status" class="input-field w-full">
                  <option value="draft">草稿</option>
                  <option value="published">上架</option>
                  <option value="archived">归档</option>
                </select>
              </label>
            </div>
          </div>

          <aside class="rounded-3xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">商品图片</p>
            <div class="mt-3 overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
              <img :src="productForm.image_url || defaultProductImage" alt="商品预览图" class="h-44 w-full object-cover">
            </div>
            <input ref="productFileInput" type="file" accept="image/jpeg,image/png,image/webp,image/gif" class="hidden" @change="handleProductImageChange">
            <button type="button" class="btn-secondary mt-3 w-full justify-center rounded-2xl px-4 py-2.5" :disabled="uploadingProductImage" @click="productFileInput?.click()">
              {{ uploadingProductImage ? '上传中...' : '本地上传图片' }}
            </button>
            <label class="shop-field mt-3">
              <span>或粘贴图片地址</span>
              <input v-model="productForm.image_url" class="input-field w-full" placeholder="https://...">
            </label>
            <p class="mt-3 text-xs leading-5 text-gray-500 dark:text-gray-400">建议使用 16:9 或 4:3 图片，支持 JPG、PNG、WebP、GIF，最大 5MB。</p>
          </aside>
        </div>
        <div class="mt-6 flex justify-end gap-3">
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="productDialog.open = false">取消</button>
          <button class="btn-primary rounded-2xl px-4 py-2.5">保存</button>
        </div>
      </form>
    </div>

    <div v-if="bannerDialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/55 p-4 backdrop-blur-sm">
      <form class="shop-modal w-full max-w-3xl rounded-3xl bg-white p-6 shadow-2xl dark:bg-dark-900" @submit.prevent="saveBanner">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-sm font-semibold text-primary-600 dark:text-primary-300">轮播管理</p>
            <h3 class="mt-1 text-xl font-bold text-gray-950 dark:text-white">{{ bannerDialog.id ? '编辑轮播' : '新增轮播' }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">设置商城顶部展示图，可关联某个商品。</p>
          </div>
          <button type="button" class="rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-800 dark:hover:text-white" aria-label="关闭" @click="bannerDialog.open = false">✕</button>
        </div>
        <div class="mt-6 grid gap-5 md:grid-cols-[1fr_16rem]">
          <div class="space-y-4">
            <label class="shop-field"><span>标题</span><input v-model="bannerForm.title" class="input-field w-full" required></label>
            <label class="shop-field"><span>副标题</span><textarea v-model="bannerForm.subtitle" class="input-field min-h-20 w-full"></textarea></label>
            <label class="shop-field"><span>按钮文字</span><input v-model="bannerForm.button_text" class="input-field w-full"></label>
            <label class="shop-field"><span>关联商品</span><select v-model.number="bannerForm.product_id" class="input-field w-full"><option :value="null">不关联</option><option v-for="product in products" :key="product.id" :value="product.id">{{ product.name }}</option></select></label>
            <label class="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-200"><input v-model="bannerForm.enabled" type="checkbox" class="rounded"> 展示轮播</label>
          </div>
          <aside class="rounded-3xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800/70">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">轮播图片</p>
            <div class="mt-3 overflow-hidden rounded-2xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
              <img :src="bannerForm.image_url || defaultBannerImage" alt="轮播预览图" class="h-36 w-full object-cover">
            </div>
            <input ref="bannerFileInput" type="file" accept="image/jpeg,image/png,image/webp,image/gif" class="hidden" @change="handleBannerImageChange">
            <button type="button" class="btn-secondary mt-3 w-full justify-center rounded-2xl px-4 py-2.5" :disabled="uploadingBannerImage" @click="bannerFileInput?.click()">
              {{ uploadingBannerImage ? '上传中...' : '本地上传图片' }}
            </button>
            <label class="shop-field mt-3"><span>或粘贴图片地址</span><input v-model="bannerForm.image_url" class="input-field w-full" placeholder="https://..."></label>
          </aside>
        </div>
        <div class="mt-6 flex justify-end gap-3">
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="bannerDialog.open = false">取消</button>
          <button class="btn-primary rounded-2xl px-4 py-2.5">保存</button>
        </div>
      </form>
    </div>
  </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { adminShopAPI, type ShopBanner, type ShopOrder, type ShopProduct, type ShopProductPayload } from '@/api/shop'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'

type TabKey = 'products' | 'banners' | 'orders'

const appStore = useAppStore()
const router = useRouter()
const tab = ref<TabKey>('products')
const tabs: Array<{ key: TabKey; label: string }> = [
  { key: 'products', label: '商品管理' },
  { key: 'banners', label: '轮播管理' },
  { key: 'orders', label: '商城订单' },
]
const products = ref<ShopProduct[]>([])
const banners = ref<ShopBanner[]>([])
const orders = ref<ShopOrder[]>([])
const productFileInput = ref<HTMLInputElement | null>(null)
const bannerFileInput = ref<HTMLInputElement | null>(null)
const uploadingProductImage = ref(false)
const uploadingBannerImage = ref(false)
const defaultProductImage = 'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=600&q=80'
const defaultBannerImage = 'https://images.unsplash.com/photo-1558494949-ef010cbdcc31?auto=format&fit=crop&w=900&q=80'

const productDialog = reactive({ open: false, id: 0 })
const bannerDialog = reactive({ open: false, id: 0 })
const productForm = reactive<ShopProductPayload>({
  name: '',
  description: '',
  image_url: '',
  product_type: 'virtual',
  price_cny_minor: 1000,
  original_price_cny_minor: 0,
  grant_usd_amount: '0',
  stock_quantity: null,
  commission_bps: 1000,
  status: 'draft',
  sort_order: 0,
})
const bannerForm = reactive({
  title: '',
  subtitle: '',
  image_url: '',
  button_text: '立即查看',
  product_id: null as number | null,
  enabled: true,
  sort_order: 0,
})

const productPrice = computed({
  get: () => productForm.price_cny_minor / 100,
  set: (value: number) => { productForm.price_cny_minor = Math.round(Number(value || 0) * 100) }
})
const productOriginalPrice = computed({
  get: () => (productForm.original_price_cny_minor || 0) / 100,
  set: (value: number) => { productForm.original_price_cny_minor = Math.round(Number(value || 0) * 100) }
})
const productCommissionPercent = computed({
  get: () => (productForm.commission_bps || 0) / 100,
  set: (value: number) => { productForm.commission_bps = Math.round(Number(value || 0) * 100) }
})

watch(tab, () => {
  if (tab.value === 'products') void loadProducts()
  if (tab.value === 'banners') void loadBanners()
  if (tab.value === 'orders') void loadOrders()
})

function money(minor: number) {
  return (minor / 100).toFixed(2)
}

function typeLabel(type: string) {
  return type === 'platform_usd_balance' ? '额度商品' : '平台商品'
}

function statusLabel(status: string) {
  return ({ draft: '草稿', published: '上架中', archived: '已归档' } as Record<string, string>)[status] || status
}

function statusClass(status: string) {
  if (status === 'published') return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
  if (status === 'draft') return 'bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-gray-300'
}

function orderStatusLabel(status: string) {
  return ({ pending: '待支付', paid: '已支付', fulfilled: '已完成', cancelled: '已取消', refunded: '已退款', failed: '失败' } as Record<string, string>)[status] || status
}

function formatDate(value: string) {
  return value ? new Date(value).toLocaleString('zh-CN') : '-'
}

function goBack() {
  void router.push('/admin/dashboard')
}

async function loadProducts() {
  const res = await adminShopAPI.listProducts()
  products.value = res.data || []
}

async function loadBanners() {
  const res = await adminShopAPI.listBanners()
  banners.value = res.data || []
}

async function loadOrders() {
  const res = await adminShopAPI.listOrders({ page: 1, page_size: 50 })
  orders.value = res.data.items || []
}

function openProductDialog(product?: ShopProduct) {
  productDialog.open = true
  productDialog.id = product?.id || 0
  Object.assign(productForm, product ? {
    name: product.name,
    description: product.description,
    image_url: product.image_url,
    product_type: product.product_type,
    price_cny_minor: product.price_cny_minor,
    original_price_cny_minor: product.original_price_cny_minor,
    grant_usd_amount: product.grant_usd_amount,
    stock_quantity: product.stock_quantity ?? null,
    commission_bps: product.commission_bps,
    status: product.status,
    sort_order: product.sort_order,
  } : {
    name: '',
    description: '',
    image_url: '',
    product_type: 'virtual',
    price_cny_minor: 1000,
    original_price_cny_minor: 0,
    grant_usd_amount: '0',
    stock_quantity: null,
    commission_bps: 1000,
    status: 'draft',
    sort_order: 0,
  })
}

async function saveProduct() {
  try {
    const payload: ShopProductPayload = {
      ...productForm,
      product_type: 'virtual',
      grant_usd_amount: '0',
    }
    if (productDialog.id) await adminShopAPI.updateProduct(productDialog.id, payload)
    else await adminShopAPI.createProduct(payload)
    productDialog.open = false
    appStore.showToast('success', '商品已保存', 2500)
    await loadProducts()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '保存商品失败', 3000)
  }
}

async function uploadImage(file: File, target: 'product' | 'banner') {
  if (!file.type.startsWith('image/')) {
    appStore.showToast('warning', '请选择图片文件', 2500)
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    appStore.showToast('warning', '图片不能超过 5MB', 2500)
    return
  }
  if (target === 'product') uploadingProductImage.value = true
  else uploadingBannerImage.value = true
  try {
    const res = await adminShopAPI.uploadAsset(file)
    if (target === 'product') productForm.image_url = res.data.url
    else bannerForm.image_url = res.data.url
    appStore.showToast('success', '图片已上传', 2200)
  } catch (error: any) {
    appStore.showToast('error', error?.message || '图片上传失败', 3000)
  } finally {
    if (target === 'product') uploadingProductImage.value = false
    else uploadingBannerImage.value = false
  }
}

function handleProductImageChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) void uploadImage(file, 'product')
  input.value = ''
}

function handleBannerImageChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) void uploadImage(file, 'banner')
  input.value = ''
}

async function removeProduct(product: ShopProduct) {
  if (!confirm(`确认删除商品「${product.name}」？`)) return
  await adminShopAPI.deleteProduct(product.id)
  appStore.showToast('success', '商品已删除', 2500)
  await loadProducts()
}

function openBannerDialog(banner?: ShopBanner) {
  bannerDialog.open = true
  bannerDialog.id = banner?.id || 0
  Object.assign(bannerForm, banner ? {
    title: banner.title,
    subtitle: banner.subtitle,
    image_url: banner.image_url,
    button_text: banner.button_text,
    product_id: banner.product_id ?? null,
    enabled: banner.enabled,
    sort_order: banner.sort_order,
  } : {
    title: '',
    subtitle: '',
    image_url: '',
    button_text: '立即查看',
    product_id: null,
    enabled: true,
    sort_order: 0,
  })
}

async function saveBanner() {
  try {
    const payload = { ...bannerForm }
    if (bannerDialog.id) await adminShopAPI.updateBanner(bannerDialog.id, payload)
    else await adminShopAPI.createBanner(payload)
    bannerDialog.open = false
    appStore.showToast('success', '轮播已保存', 2500)
    await loadBanners()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '保存轮播失败', 3000)
  }
}

async function removeBanner(banner: ShopBanner) {
  if (!confirm(`确认删除轮播「${banner.title}」？`)) return
  await adminShopAPI.deleteBanner(banner.id)
  appStore.showToast('success', '轮播已删除', 2500)
  await loadBanners()
}

onMounted(async () => {
  await Promise.all([loadProducts(), loadBanners(), loadOrders()])
})
</script>

<style scoped>
.shop-modal {
  color: rgb(17 24 39);
}

:global(.dark) .shop-modal {
  color: rgb(243 244 246);
}

.shop-field {
  display: block;
}

.shop-field > span {
  margin-bottom: 0.35rem;
  display: block;
  font-size: 0.875rem;
  font-weight: 700;
  color: rgb(55 65 81);
}

:global(.dark) .shop-field > span {
  color: rgb(229 231 235);
}

.shop-modal :deep(.input-field) {
  min-height: 2.75rem;
  border-color: rgb(209 213 219);
  background-color: white;
  color: rgb(17 24 39);
}

.shop-modal :deep(.input-field::placeholder) {
  color: rgb(156 163 175);
}

:global(.dark) .shop-modal :deep(.input-field) {
  border-color: rgb(55 65 81);
  background-color: rgb(17 24 39);
  color: white;
}
</style>
