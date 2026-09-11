<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import FAQ_ITEMS from '@/content/home-faq.json'
import { useAppStore } from '@/stores/app'
import { publicShopAPI, type PublicBanner, type PublicProduct } from '@/api/publicShop'
import { SHOP_CATEGORIES, normalizeShopCategory, type ShopCategory } from '@/constants/shop'
import userAPI from '@/api/user'
import { resolveShopAssetUrl } from '@/api/shop'
import { useGuestCheckout } from '@/composables/useGuestCheckout'
import PlanCard from '@/components/home/PlanCard.vue'
import OrderSheet from '@/components/home/OrderSheet.vue'
import SessionGuide from '@/components/home/SessionGuide.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'

const appStore = useAppStore()
const {
  isLoggedIn,
  submitting,
  payDialog,
  startCheckout,
  goToOrderPage,
  closePayDialog,
  methodsForAmount,
  loadPaymentMethods
} = useGuestCheckout()

const products = ref<PublicProduct[]>([])
const banners = ref<PublicBanner[]>([])
const affCode = ref('')
const loading = ref(true)
const sheetOpen = ref(false)
const selectedProduct = ref<PublicProduct | null>(null)
const openFaq = ref<number | null>(0)

/**
 * 商品按品类分组（而不是按交付模式 tab）。
 * 品类回答「卖什么」（GPT 代充 / 成品号 / X 蓝V / Gemini ...），所有分组都直接渲染到 DOM，
 * 不像 tab 那样只渲染当前激活的一组——既方便横向比价，也让爬虫能抓到全部商品。
 */
function compareProducts(a: PublicProduct, b: PublicProduct) {
  if (a.highlight !== b.highlight) return a.highlight ? -1 : 1
  if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
  return a.id - b.id
}

const categoryGroups = computed(() => {
  const buckets = new Map<ShopCategory, PublicProduct[]>()
  for (const product of products.value) {
    const key = normalizeShopCategory(product.category)
    const bucket = buckets.get(key)
    if (bucket) bucket.push(product)
    else buckets.set(key, [product])
  }
  // 按 SHOP_CATEGORIES 的固定顺序输出，空品类不渲染
  return SHOP_CATEGORIES.filter((meta) => (buckets.get(meta.value)?.length ?? 0) > 0).map((meta) => ({
    value: meta.value,
    label: meta.label,
    blurb: meta.blurb,
    items: (buckets.get(meta.value) || []).slice().sort(compareProducts)
  }))
})

const contactInfo = computed(() => appStore.contactInfo?.trim() || '')
// 下单抽屉里的支付方式：后台配置 ∩ 当前商品金额在单笔限额内
const sheetPaymentMethods = computed(() =>
  selectedProduct.value ? methodsForAmount(selectedProduct.value.price_cny_minor) : []
)

const METRICS = [
  { value: '1-3', unit: '分钟', label: '平均到账时间' },
  { value: '30', unit: '天', label: '非人为中断质保' },
  { value: '0', unit: '个', label: '索要的账号密码' },
  { value: '24', unit: '小时', label: '售后响应窗口' }
]

/** 三种交付方式对比：按「方式」分组而不是按维度分行，移动端更好读 */
const COMPARISON: { name: string; for: string; rows: [string, string][] }[] = [
  {
    name: '代充值',
    for: '已有账号要续费',
    rows: [
      ['需要提供', '登录凭证'],
      ['账号归属', '你自己的号'],
      ['交付时效', '1-3 分钟'],
      ['成本', '低']
    ]
  },
  {
    name: '成品号',
    for: '想换个新号',
    rows: [
      ['需要提供', '接收邮箱'],
      ['账号归属', '交付给你'],
      ['交付时效', '人工发货'],
      ['成本', '中']
    ]
  },
  {
    name: '租号',
    for: '短期试用',
    rows: [
      ['需要提供', '接收邮箱'],
      ['账号归属', '租期内使用'],
      ['交付时效', '人工发货'],
      ['成本', '最低']
    ]
  }
]

onMounted(async () => {
  // 支付渠道与控制台商城同源，后台开关与限额即时生效
  void loadPaymentMethods()
  try {
    const [productRes, bannerRes] = await Promise.all([
      publicShopAPI.listProducts(),
      publicShopAPI.listBanners().catch(() => null)
    ])
    products.value = productRes.data || []
    const list = bannerRes?.data || []
    banners.value = list.filter((item) => item.enabled)
  } catch {
    products.value = []
  } finally {
    loading.value = false
  }
  // 推广奖励只在登录后展示
  if (isLoggedIn.value) {
    try {
      const detail = await userAPI.getAffiliateDetail()
      affCode.value = detail?.aff_code?.trim() || ''
    } catch {
      affCode.value = ''
    }
  }
})

function openSheet(product: PublicProduct) {
  selectedProduct.value = product
  sheetOpen.value = true
}

function openBannerProduct(productID: number) {
  const product = products.value.find((item) => item.id === productID)
  if (!product) return
  openSheet(product)
}

async function handleSubmit(payload: { contact: string; paymentType: string }) {
  if (!selectedProduct.value) return
  const ok = await startCheckout(selectedProduct.value, payload.contact, payload.paymentType)
  if (ok) {
    sheetOpen.value = false
  }
}

function handlePaymentSuccess() {
  goToOrderPage()
}
</script>

<template>
  <div class="sales-home">
    <header class="nav">
      <div class="nav__inner">
        <a class="nav__brand" href="/">
          <span class="nav__mark">3</span>
          <span class="nav__name">3API</span>
        </a>
        <nav class="nav__links">
          <a href="#plans">套餐</a>
          <a href="#workflow">流程</a>
          <a href="#faq">常见问题</a>
          <RouterLink to="/order">查订单</RouterLink>
        </nav>
        <div class="nav__actions">
          <RouterLink class="nav__ghost" to="/login">登录</RouterLink>
          <a class="nav__cta" href="#plans">立即充值</a>
        </div>
      </div>
    </header>

    <main>
      <section class="hero">
        <div class="hero__grid" aria-hidden="true" />
        <div class="hero__aura" aria-hidden="true">
          <span class="hero__orb hero__orb--a" />
          <span class="hero__orb hero__orb--b" />
        </div>
        <div class="hero__inner">
          <div class="hero__copy">
            <p class="eyebrow">ChatGPT Plus / Pro 充值 · 成品号直供</p>
            <h1 class="hero__title">给你的 ChatGPT 续上<br>官方会员与独享账号</h1>
            <p class="hero__desc">
              官方渠道代充值，1-3 分钟到账，30 天质保。成品号与租号即买即用，全程不需要账号密码。
            </p>
            <div class="hero__actions">
              <a class="sh-btn sh-btn--primary" href="#plans">选择套餐</a>
              <a class="sh-btn sh-btn--ghost" href="#workflow">查看充值流程</a>
            </div>
            <ul class="hero__trust">
              <li>不需要账号密码</li>
              <li>支付后凭订单号查进度</li>
              <li>独立第三方服务</li>
            </ul>
          </div>

          <aside class="hero__panel" aria-hidden="true">
            <div class="panel">
              <div class="panel__head">
                <span class="panel__pulse" />
                <span>充值进度</span>
                <span class="panel__tag">实时</span>
              </div>
              <ol class="panel__steps">
                <li class="panel__step is-done">
                  <span class="panel__node" />
                  <div>
                    <p>选择套餐</p>
                    <small>确认价格与服务说明</small>
                  </div>
                </li>
                <li class="panel__step is-done">
                  <span class="panel__node" />
                  <div>
                    <p>填写联系方式</p>
                    <small>仅用于查单与售后</small>
                  </div>
                </li>
                <li class="panel__step is-active">
                  <span class="panel__node" />
                  <div>
                    <p>提交登录凭证</p>
                    <small>支付后随时可补交</small>
                  </div>
                </li>
                <li class="panel__step">
                  <span class="panel__node" />
                  <div>
                    <p>到账完成</p>
                    <small>平均 1-3 分钟</small>
                  </div>
                </li>
              </ol>
            </div>
          </aside>
        </div>
      </section>

      <section class="metrics">
        <div class="metrics__inner">
          <div v-for="item in METRICS" :key="item.label" class="metric">
            <p class="metric__value">
              {{ item.value }}<span>{{ item.unit }}</span>
            </p>
            <p class="metric__label">{{ item.label }}</p>
          </div>
        </div>
      </section>

      <section v-if="banners.length" class="banners">
        <div class="section__inner">
          <div class="banners__track">
            <article v-for="banner in banners" :key="banner.id" class="banner">
              <div v-if="banner.image_url" class="banner__media">
                <img :src="resolveShopAssetUrl(banner.image_url)" :alt="banner.title" loading="lazy">
              </div>
              <div class="banner__body">
                <h3 class="banner__title">{{ banner.title }}</h3>
                <p v-if="banner.subtitle" class="banner__subtitle">{{ banner.subtitle }}</p>
                <button
                  v-if="banner.product_id"
                  type="button"
                  class="banner__cta"
                  @click="openBannerProduct(banner.product_id)"
                >
                  {{ banner.button_text || '立即查看' }}
                </button>
              </div>
            </article>
          </div>
        </div>
      </section>

      <section id="plans" class="plans">
        <div class="section__inner">
          <header class="section__head">
            <h2 class="section__title">选择套餐</h2>
            <p class="section__desc">价格以支付前页面显示为准，支持支付宝与微信支付</p>
          </header>

          <div v-if="loading" class="plans__grid">
            <div v-for="i in 3" :key="i" class="sh-skeleton" />
          </div>

          <template v-else-if="categoryGroups.length">
            <nav v-if="categoryGroups.length > 1" class="plans__nav" aria-label="商品分类">
              <a v-for="group in categoryGroups" :key="group.value" class="plans__nav-item" :href="`#cat-${group.value}`">
                {{ group.label }}
              </a>
            </nav>

            <section v-for="group in categoryGroups" :id="`cat-${group.value}`" :key="group.value" class="plans__group">
              <header class="plans__group-head">
                <h3 class="plans__group-title">{{ group.label }}</h3>
                <p v-if="group.blurb" class="plans__group-blurb">{{ group.blurb }}</p>
              </header>
              <div class="plans__grid">
                <PlanCard
                  v-for="product in group.items"
                  :key="product.id"
                  :product="product"
                  :show-commission="isLoggedIn"
                  :aff-code="affCode"
                  @select="openSheet"
                />
              </div>
            </section>
          </template>

          <div v-else class="empty">
            <p class="empty__title">套餐正在准备中</p>
            <p class="empty__desc">可以先联系客服咨询，或稍后再来看看</p>
          </div>
        </div>
      </section>

      <section id="workflow" class="workflow">
        <div class="section__inner">
          <header class="section__head">
            <h2 class="section__title">三步完成充值</h2>
            <p class="section__desc">先选套餐付款，再补交凭证，全程不需要账号密码</p>
          </header>

          <div class="workflow__grid">
            <article class="step">
              <span class="step__num">01</span>
              <h3>选择套餐</h3>
              <p>在上方选择代充值、成品号或租号，确认价格与服务说明</p>
            </article>
            <article class="step">
              <span class="step__num">02</span>
              <h3>留联系方式并支付</h3>
              <p>填写手机号或邮箱用于查单与售后，选择支付宝或微信完成支付</p>
            </article>
            <article class="step">
              <span class="step__num">03</span>
              <h3>提交信息等待到账</h3>
              <p>代充值粘贴登录凭证，成品号与租号留接收邮箱，随后等待交付</p>
            </article>
          </div>

          <div class="guide-wrap">
            <div class="guide-wrap__head">
              <h3>怎么拿到登录凭证</h3>
              <p>四步搞定，复制粘贴即可，不需要懂技术</p>
            </div>
            <SessionGuide />
          </div>
        </div>
      </section>

      <section class="compare">
        <div class="section__inner">
          <header class="section__head">
            <h2 class="section__title">三种方式怎么选</h2>
            <p class="section__desc">按你现在的情况对号入座</p>
          </header>
          <div class="compare__grid">
            <article v-for="col in COMPARISON" :key="col.name" class="compare__card">
              <header class="compare__head">
                <h3 class="compare__name">{{ col.name }}</h3>
                <p class="compare__for">{{ col.for }}</p>
              </header>
              <dl class="compare__rows">
                <div v-for="row in col.rows" :key="row[0]" class="compare__row">
                  <dt>{{ row[0] }}</dt>
                  <dd>{{ row[1] }}</dd>
                </div>
              </dl>
            </article>
          </div>
        </div>
      </section>

      <section id="faq" class="faq">
        <div class="section__inner section__inner--narrow">
          <header class="section__head">
            <h2 class="section__title">常见问题</h2>
            <p class="section__desc">下单前最常被问到的几件事</p>
          </header>
          <div class="faq__list">
            <div
              v-for="(item, index) in FAQ_ITEMS"
              :key="item.q"
              class="faq__item"
              :class="{ 'faq__item--open': openFaq === index }"
            >
              <button type="button" class="faq__q" @click="openFaq = openFaq === index ? null : index">
                <span>{{ item.q }}</span>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
                  <path d="M4 6.5l4 4 4-4" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </button>
              <p v-if="openFaq === index" class="faq__a">{{ item.a }}</p>
            </div>
          </div>
        </div>
      </section>

      <section class="closing">
        <div class="section__inner closing__inner">
          <h2 class="closing__title">准备好开始了吗</h2>
          <a class="sh-btn sh-btn--primary" href="#plans">选择套餐</a>
        </div>
      </section>
    </main>

    <footer class="footer">
      <div class="footer__inner">
        <div class="footer__brand">
          <span class="nav__mark">3</span>
          <p>3API · ChatGPT 充值与成品号服务</p>
        </div>
        <div class="footer__cols">
          <div>
            <h4>服务</h4>
            <a href="#plans">代充值</a>
            <a href="#plans">成品号</a>
            <a href="#plans">租号</a>
          </div>
          <div>
            <h4>帮助</h4>
            <RouterLink to="/order">查订单</RouterLink>
            <a href="#workflow">充值流程</a>
            <a href="#faq">常见问题</a>
          </div>
          <div>
            <h4>开发者</h4>
            <RouterLink to="/api-relay">API 中转</RouterLink>
            <RouterLink to="/openai-api">OpenAI 接入</RouterLink>
            <RouterLink to="/login">控制台登录</RouterLink>
          </div>
          <div v-if="contactInfo">
            <h4>客服</h4>
            <p class="footer__contact">{{ contactInfo }}</p>
          </div>
        </div>
      </div>
      <p class="footer__legal">
        本站为独立第三方服务平台，非 OpenAI 或 ChatGPT 官方网站，与相关官方主体不存在授权、代理或合作关系。
      </p>
    </footer>

    <div class="mobile-cta">
      <a href="#plans">立即充值</a>
    </div>

    <OrderSheet
      :open="sheetOpen"
      :product="selectedProduct"
      :payment-methods="sheetPaymentMethods"
      :submitting="submitting"
      @close="sheetOpen = false"
      @submit="handleSubmit"
    />

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
      />
    </BaseDialog>
  </div>
</template>

<style scoped>
/* ============================================================
   设计令牌 —— 明暗双主题
   浅色：白底 + 中性灰阶 + 极低饱和描边（Stripe 式克制）
   深色：近黑表面 + hairline 描边（Linear 式）
   强调色全站仅一个：品牌橙，只用于 CTA / 激活态 / 小标记
   ============================================================ */
.sales-home {
  --sh-bg: #ffffff;
  --sh-surface: #ffffff;
  --sh-surface-2: #f6f6f7;
  --sh-border: rgba(9, 9, 11, 0.08);
  --sh-border-strong: rgba(9, 9, 11, 0.16);
  --sh-text: #09090b;
  --sh-text-2: #56565f;
  --sh-text-3: #8a8a93;
  --sh-accent: #d85a28;
  --sh-accent-hover: #c04f21;
  --sh-accent-fg: #ffffff;
  --sh-accent-soft: rgba(216, 90, 40, 0.08);
  --sh-accent-text: #b34b1f;
  --sh-grid: rgba(9, 9, 11, 0.05);
  --sh-scrim: rgba(9, 9, 11, 0.55);
  /* 首屏光斑：品牌橙 + 少量冷色，饱和度压到最低，只做氛围不抢内容 */
  --sh-aura-1: rgba(216, 90, 40, 0.15);
  --sh-aura-2: rgba(58, 118, 240, 0.11);
  --sh-shadow: 0 1px 2px rgba(9, 9, 11, 0.04), 0 1px 3px rgba(9, 9, 11, 0.04);

  min-height: 100vh;
  background: var(--sh-bg);
  color: var(--sh-text);
  font-family: inherit;
  -webkit-font-smoothing: antialiased;
}


.section__inner {
  max-width: 1120px;
  margin: 0 auto;
  padding: 0 24px;
}

.section__inner--narrow {
  max-width: 760px;
}

.section__head {
  margin-bottom: 40px;
  max-width: 640px;
}

.section__title {
  font-size: 30px;
  font-weight: 600;
  letter-spacing: -0.025em;
  color: var(--sh-text);
}

.section__desc {
  margin-top: 10px;
  font-size: 15px;
  line-height: 1.7;
  color: var(--sh-text-2);
}

.eyebrow {
  font-size: 13px;
  letter-spacing: 0.02em;
  color: var(--sh-text-3);
}

/* ---------- nav ---------- */
.nav {
  position: sticky;
  top: 0;
  z-index: 40;
  border-bottom: 1px solid var(--sh-border);
  background: color-mix(in srgb, var(--sh-bg) 82%, transparent);
  backdrop-filter: saturate(180%) blur(14px);
}

.nav__inner {
  max-width: 1120px;
  margin: 0 auto;
  padding: 0 24px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

.nav__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
}

.nav__mark {
  width: 26px;
  height: 26px;
  display: grid;
  place-items: center;
  border-radius: 7px;
  background: var(--sh-text);
  color: var(--sh-bg);
  font-size: 14px;
  font-weight: 600;
}

.nav__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--sh-text);
  letter-spacing: -0.01em;
}

.nav__links {
  display: flex;
  gap: 28px;
  font-size: 14px;
}

.nav__links a {
  color: var(--sh-text-2);
  text-decoration: none;
  transition: color 160ms ease-out;
}

.nav__links a:hover {
  color: var(--sh-text);
}

.nav__actions {
  display: flex;
  align-items: center;
  gap: 14px;
}

.nav__ghost {
  font-size: 14px;
  color: var(--sh-text-2);
  text-decoration: none;
  transition: color 160ms ease-out;
}

.nav__ghost:hover {
  color: var(--sh-text);
}

.nav__cta {
  padding: 8px 15px;
  border-radius: 8px;
  background: var(--sh-accent);
  color: var(--sh-accent-fg);
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  transition: background 160ms ease-out;
}

.nav__cta:hover {
  background: var(--sh-accent-hover);
}

/* ---------- hero ---------- */
.hero {
  position: relative;
  padding: 104px 0 80px;
  overflow: hidden;
}

.hero__grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(var(--sh-grid) 1px, transparent 1px),
    linear-gradient(90deg, var(--sh-grid) 1px, transparent 1px);
  background-size: 68px 68px;
  -webkit-mask-image: radial-gradient(ellipse 90% 62% at 50% 0%, #000 0%, transparent 72%);
  mask-image: radial-gradient(ellipse 90% 62% at 50% 0%, #000 0%, transparent 72%);
  pointer-events: none;
}

/* 动态背景：两团渐变光斑做极缓呼吸（34s / 44s 交替）。
   只动 opacity 与 transform —— 两者都走 GPU 合成，不触发重排重绘；
   没有 JS、没有 canvas、没有视频，常驻内存开销约等于 0。
   光斑本身用 radial-gradient 自带柔边，不需要 filter: blur（那才真的吃性能）。 */
.hero__aura {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}

.hero__orb {
  position: absolute;
  border-radius: 50%;
  will-change: opacity, transform;
}

.hero__orb--a {
  top: -22%;
  left: -8%;
  width: 52vw;
  height: 52vw;
  background: radial-gradient(circle at 50% 50%, var(--sh-aura-1) 0%, transparent 66%);
  animation: sh-orb-a 34s ease-in-out infinite alternate;
}

.hero__orb--b {
  top: -12%;
  right: -12%;
  width: 44vw;
  height: 44vw;
  background: radial-gradient(circle at 50% 50%, var(--sh-aura-2) 0%, transparent 66%);
  animation: sh-orb-b 44s ease-in-out infinite alternate;
}

@keyframes sh-orb-a {
  from {
    opacity: 0.5;
    transform: translate3d(0, 0, 0) scale(1);
  }
  to {
    opacity: 0.85;
    transform: translate3d(2%, -2%, 0) scale(1.07);
  }
}

@keyframes sh-orb-b {
  from {
    opacity: 0.38;
    transform: translate3d(0, 0, 0) scale(1.04);
  }
  to {
    opacity: 0.68;
    transform: translate3d(-2%, 2%, 0) scale(1);
  }
}

/* 系统开启「减弱动态效果」时退化为静态光斑 */
@media (prefers-reduced-motion: reduce) {
  .hero__orb {
    animation: none;
    opacity: 0.6;
  }
}

.hero__inner {
  position: relative;
  max-width: 1120px;
  margin: 0 auto;
  padding: 0 24px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 372px;
  gap: 64px;
  align-items: center;
}

.hero__title {
  margin-top: 18px;
  font-size: 48px;
  line-height: 1.12;
  font-weight: 600;
  letter-spacing: -0.035em;
  color: var(--sh-text);
}

.hero__desc {
  margin-top: 20px;
  max-width: 560px;
  font-size: 16px;
  line-height: 1.75;
  color: var(--sh-text-2);
}

.hero__actions {
  margin-top: 32px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.sh-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 46px;
  padding: 0 22px;
  border-radius: 9px;
  font-size: 15px;
  font-weight: 500;
  text-decoration: none;
  transition: background 160ms ease-out, border-color 160ms ease-out;
}

.sh-btn--primary {
  background: var(--sh-accent);
  color: var(--sh-accent-fg);
}

.sh-btn--primary:hover {
  background: var(--sh-accent-hover);
}

.sh-btn--ghost {
  border: 1px solid var(--sh-border-strong);
  color: var(--sh-text);
}

.sh-btn--ghost:hover {
  background: var(--sh-surface-2);
}

.hero__trust {
  margin-top: 28px;
  padding: 0;
  list-style: none;
  display: flex;
  flex-wrap: wrap;
  gap: 8px 20px;
  font-size: 13px;
  color: var(--sh-text-3);
}

.hero__trust li {
  display: flex;
  align-items: center;
  gap: 8px;
}

.hero__trust li::before {
  content: '';
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--sh-border-strong);
}

/* ---------- hero panel ---------- */
.panel {
  padding: 20px;
  border-radius: 14px;
  border: 1px solid var(--sh-border);
  background: var(--sh-surface);
  box-shadow: var(--sh-shadow);
}

.panel__head {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--sh-text-2);
  padding-bottom: 16px;
  border-bottom: 1px solid var(--sh-border);
}

.panel__pulse {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #1a9e6f;
  box-shadow: 0 0 0 3px rgba(26, 158, 111, 0.14);
}

.panel__tag {
  margin-left: auto;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid var(--sh-border);
  font-size: 11px;
  color: var(--sh-text-3);
}

.panel__steps {
  margin: 0;
  padding: 0;
  list-style: none;
}

.panel__step {
  display: flex;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid var(--sh-border);
}

.panel__step:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.panel__node {
  flex-shrink: 0;
  width: 9px;
  height: 9px;
  margin-top: 6px;
  border-radius: 50%;
  border: 1.5px solid var(--sh-border-strong);
}

.panel__step.is-done .panel__node {
  background: var(--sh-text-3);
  border-color: var(--sh-text-3);
}

.panel__step.is-active .panel__node {
  background: var(--sh-accent);
  border-color: var(--sh-accent);
  box-shadow: 0 0 0 4px var(--sh-accent-soft);
}

.panel__step p {
  font-size: 14px;
  color: var(--sh-text-3);
}

.panel__step.is-done p,
.panel__step.is-active p {
  color: var(--sh-text);
}

.panel__step small {
  display: block;
  margin-top: 3px;
  font-size: 12px;
  color: var(--sh-text-3);
}

/* ---------- metrics ---------- */
.metrics {
  border-top: 1px solid var(--sh-border);
  border-bottom: 1px solid var(--sh-border);
}

.metrics__inner {
  max-width: 1120px;
  margin: 0 auto;
  padding: 40px 24px;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
}

.metric__value {
  font-size: 32px;
  font-weight: 600;
  color: var(--sh-text);
  letter-spacing: -0.02em;
  font-variant-numeric: tabular-nums;
}

.metric__value span {
  margin-left: 4px;
  font-size: 14px;
  font-weight: 400;
  color: var(--sh-text-3);
}

.metric__label {
  margin-top: 6px;
  font-size: 13px;
  color: var(--sh-text-3);
}

/* ---------- banners ---------- */
.banners {
  padding: 64px 0 0;
}

.banners__track {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
}

.banner {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 14px;
  border: 1px solid var(--sh-border);
  background: var(--sh-surface);
}

.banner__media {
  height: 128px;
  overflow: hidden;
  background: var(--sh-surface-2);
}

.banner__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.banner__body {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  padding: 18px 20px 20px;
}

.banner__title {
  font-size: 17px;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: var(--sh-text);
}

.banner__subtitle {
  font-size: 13px;
  line-height: 1.7;
  color: var(--sh-text-2);
}

.banner__cta {
  margin-top: 6px;
  padding: 9px 18px;
  border: 1px solid var(--sh-border-strong);
  border-radius: 9px;
  background: transparent;
  color: var(--sh-text);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 160ms ease-out;
}

.banner__cta:hover {
  background: var(--sh-surface-2);
}

/* ---------- plans ---------- */
.plans {
  padding: 96px 0;
}

/* 品类锚点导航：胶囊 chip，点击滚到对应分组 */
.plans__nav {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 40px;
}

.plans__nav-item {
  padding: 7px 16px;
  border-radius: 999px;
  border: 1px solid var(--sh-border);
  background: var(--sh-surface-2);
  color: var(--sh-text-2);
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  transition: color 160ms ease-out, border-color 160ms ease-out;
}

.plans__nav-item:hover {
  color: var(--sh-text);
  border-color: var(--sh-border-strong);
}

/* 分组之间留出比卡片间距更大的呼吸，避免所有商品糊成一片 */
.plans__group {
  scroll-margin-top: 88px;
}

.plans__group + .plans__group {
  margin-top: 56px;
}

.plans__group-head {
  margin-bottom: 20px;
}

.plans__group-title {
  font-size: 20px;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: var(--sh-text);
}

.plans__group-blurb {
  margin-top: 6px;
  font-size: 13px;
  color: var(--sh-text-3);
}

.plans__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.sh-skeleton {
  height: 320px;
  border-radius: 14px;
  border: 1px solid var(--sh-border);
  background: var(--sh-surface-2);
}

.empty {
  padding: 64px 24px;
  text-align: center;
  border: 1px dashed var(--sh-border-strong);
  border-radius: 14px;
}

.empty__title {
  font-size: 16px;
  color: var(--sh-text);
}

.empty__desc {
  margin-top: 8px;
  font-size: 14px;
  color: var(--sh-text-3);
}

/* ---------- workflow ---------- */
.workflow {
  padding: 96px 0;
  border-top: 1px solid var(--sh-border);
}

.workflow__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.step {
  padding: 24px;
  border-radius: 14px;
  border: 1px solid var(--sh-border);
  background: var(--sh-surface);
}

.step__num {
  display: inline-block;
  font-size: 12px;
  font-weight: 600;
  color: var(--sh-text-3);
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.04em;
}

.step h3 {
  margin-top: 14px;
  font-size: 17px;
  font-weight: 600;
  color: var(--sh-text);
}

.step p {
  margin-top: 10px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--sh-text-2);
}

.guide-wrap {
  margin-top: 32px;
  padding: 28px;
  border-radius: 14px;
  border: 1px solid var(--sh-border);
  background: var(--sh-surface);
}

.guide-wrap__head {
  margin-bottom: 22px;
}

.guide-wrap__head h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--sh-text);
}

.guide-wrap__head p {
  margin-top: 8px;
  font-size: 14px;
  color: var(--sh-text-2);
}

/* ---------- compare ---------- */
.compare {
  padding: 96px 0;
  border-top: 1px solid var(--sh-border);
}

/* 三列对比卡：比表格少一半文字，移动端也不用横向滚动 */
.compare__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.compare__card {
  padding: 24px 22px;
  border: 1px solid var(--sh-border);
  border-radius: 14px;
  background: var(--sh-surface);
}

.compare__name {
  font-size: 17px;
  font-weight: 600;
  color: var(--sh-text);
}

.compare__for {
  margin-top: 4px;
  font-size: 13px;
  color: var(--sh-text-3);
}

.compare__rows {
  margin: 20px 0 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.compare__row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
}

.compare__row dt {
  color: var(--sh-text-3);
}

.compare__row dd {
  color: var(--sh-text);
  text-align: right;
}

/* ---------- faq ---------- */
.faq {
  padding: 96px 0;
  border-top: 1px solid var(--sh-border);
}

.faq__list {
  border-top: 1px solid var(--sh-border);
}

.faq__item {
  border-bottom: 1px solid var(--sh-border);
}

.faq__q {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 0;
  background: transparent;
  border: none;
  color: var(--sh-text);
  font-size: 16px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
}

.faq__q svg {
  flex-shrink: 0;
  transition: transform 200ms ease-out;
  color: var(--sh-text-3);
}

.faq__item--open .faq__q svg {
  transform: rotate(180deg);
}

.faq__a {
  padding: 0 0 22px;
  font-size: 14px;
  line-height: 1.8;
  color: var(--sh-text-2);
}

/* ---------- closing ---------- */
.closing {
  padding: 72px 0;
  border-top: 1px solid var(--sh-border);
}

/* 一句话 + 一个按钮，横向排布；不再重复 hero 里已经说过的说明文字 */
.closing__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  flex-wrap: wrap;
}

.closing__title {
  font-size: 26px;
  font-weight: 600;
  letter-spacing: -0.025em;
  color: var(--sh-text);
}

/* ---------- footer ---------- */
.footer {
  border-top: 1px solid var(--sh-border);
  padding: 56px 0 32px;
}

.footer__inner {
  max-width: 1120px;
  margin: 0 auto;
  padding: 0 24px;
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 48px;
}

.footer__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: var(--sh-text-2);
}

.footer__cols {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
}

.footer__cols h4 {
  font-size: 13px;
  font-weight: 600;
  color: var(--sh-text);
  margin-bottom: 12px;
}

.footer__cols a,
.footer__cols p {
  display: block;
  margin-top: 8px;
  font-size: 13px;
  color: var(--sh-text-3);
  text-decoration: none;
}

.footer__cols a:hover {
  color: var(--sh-text);
}

.footer__contact {
  white-space: pre-wrap;
}

.footer__legal {
  max-width: 1120px;
  margin: 40px auto 0;
  padding: 0 24px;
  font-size: 12px;
  line-height: 1.8;
  color: var(--sh-text-3);
}

/* ---------- mobile ---------- */
.mobile-cta {
  display: none;
}

@media (max-width: 1024px) {
  .hero__inner {
    grid-template-columns: 1fr;
    gap: 40px;
  }

  .hero__panel {
    max-width: 460px;
  }

  .metrics__inner {
    grid-template-columns: repeat(2, 1fr);
  }

  .workflow__grid,
  .compare__grid {
    grid-template-columns: 1fr;
  }

  .footer__inner {
    grid-template-columns: 1fr;
    gap: 32px;
  }
}

@media (max-width: 768px) {
  .nav__links {
    display: none;
  }

  .nav__actions {
    gap: 14px;
  }

  .hero {
    padding: 56px 0 48px;
  }

  .hero__title {
    margin-top: 14px;
    font-size: 32px;
    line-height: 1.18;
    letter-spacing: -0.03em;
  }

  .hero__desc {
    margin-top: 16px;
    font-size: 15px;
  }

  .hero__actions {
    margin-top: 26px;
  }

  .hero__actions .sh-btn {
    flex: 1 1 auto;
    min-width: 0;
  }

  .hero__trust {
    margin-top: 22px;
    gap: 6px 16px;
  }

  .banners {
    padding-top: 48px;
  }

  .plans,
  .workflow,
  .compare,
  .faq,
  .closing {
    padding: 64px 0;
  }

  .section__head {
    margin-bottom: 28px;
  }

  .section__title {
    font-size: 24px;
  }

  /* 品类导航横向滚动，避免竖排堆叠占满首屏 */
  .plans__nav {
    flex-wrap: nowrap;
    overflow-x: auto;
    scrollbar-width: none;
    margin-bottom: 28px;
  }

  .plans__nav::-webkit-scrollbar {
    display: none;
  }

  .plans__nav-item {
    flex: 0 0 auto;
  }

  .plans__group + .plans__group {
    margin-top: 44px;
  }

  .plans__grid {
    grid-template-columns: 1fr;
  }

  .closing__inner {
    flex-direction: column;
    align-items: flex-start;
    gap: 20px;
  }

  .footer__cols {
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }

  .mobile-cta {
    position: fixed;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 30;
    display: block;
    padding: 10px 16px calc(10px + env(safe-area-inset-bottom));
    background: color-mix(in srgb, var(--sh-bg) 92%, transparent);
    border-top: 1px solid var(--sh-border);
    backdrop-filter: blur(12px);
  }

  .mobile-cta a {
    display: block;
    height: 46px;
    line-height: 46px;
    text-align: center;
    border-radius: 9px;
    background: var(--sh-accent);
    color: var(--sh-accent-fg);
    font-size: 16px;
    font-weight: 500;
    text-decoration: none;
  }

  .footer {
    padding-bottom: 104px;
  }
}

@media (prefers-reduced-motion: reduce) {
  * {
    transition-duration: 0.01ms !important;
  }
}
</style>

<style>
/* 深色覆盖放在**非 scoped** 块中：Vue 的 scoped-CSS 编译器会把
   `:global(.dark) X` 编译成只有 `.dark`，X 被丢弃，导致生产构建里
   深色规则整体失效。与 SettingsView 的处理方式保持一致。 */
.dark .sales-home {
  --sh-bg: #08090b;
  --sh-surface: #0e0f12;
  --sh-surface-2: #131418;
  --sh-border: rgba(255, 255, 255, 0.08);
  --sh-border-strong: rgba(255, 255, 255, 0.18);
  --sh-text: #f4f4f5;
  --sh-text-2: rgba(255, 255, 255, 0.6);
  --sh-text-3: rgba(255, 255, 255, 0.4);
  --sh-accent: #e0632f;
  --sh-accent-hover: #ef7440;
  --sh-accent-fg: #ffffff;
  --sh-accent-soft: rgba(224, 99, 47, 0.13);
  --sh-accent-text: #f0916a;
  --sh-grid: rgba(255, 255, 255, 0.045);
  --sh-scrim: rgba(0, 0, 0, 0.6);
  --sh-shadow: none;
  /* 深色下光斑更亮一点，否则近黑背景上看不见 */
  --sh-aura-1: rgba(255, 122, 69, 0.2);
  --sh-aura-2: rgba(88, 132, 255, 0.15);
}
</style>
