<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import FAQ_ITEMS from '@/content/home-faq.json'
import { useAppStore } from '@/stores/app'
import { formatCNY, publicShopAPI, type PublicBanner, type PublicProduct } from '@/api/publicShop'
import { SHOP_CATEGORIES, normalizeShopCategory, type ShopCategory } from '@/constants/shop'
import userAPI from '@/api/user'
import { resolveShopAssetUrl } from '@/api/shop'
import { planHeroPromo } from '@/utils/heroPromo'
import { useGuestCheckout } from '@/composables/useGuestCheckout'
import { useRevealOnScroll } from '@/composables/useRevealOnScroll'
import PlanCard from '@/components/home/PlanCard.vue'
import OrderSheet from '@/components/home/OrderSheet.vue'
import SessionGuide from '@/components/home/SessionGuide.vue'
import HeroPromo from '@/components/home/HeroPromo.vue'
import HeroFloatCards from '@/components/home/HeroFloatCards.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'

const appStore = useAppStore()
const { vReveal } = useRevealOnScroll()
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

/** 按展示顺序（推荐位 → sort_order → id）排好的商品，首屏所有选品都从这里取 */
const orderedProducts = computed(() => products.value.slice().sort(compareProducts))

/**
 * 首屏右侧悬浮卡取前三个商品。
 * 用真实商品而不是装饰卡——悬浮元素本身就是购买入口，点一下直接开下单抽屉。
 * 想让某个商品出现在这里，在后台把它勾上「推荐位」或把排序值调小。
 */
const heroProducts = computed(() => orderedProducts.value.slice(0, 3))

const heroPromoPlan = computed(() => planHeroPromo(orderedProducts.value))

/**
 * 首屏促销条的文案全部由真实商品数据推导。选品规则见 @/utils/heroPromo：
 * 只主推后台勾了「推荐」的商品，没有任何推荐位时退化成不含商品名的服务承诺——
 * 避免自动挑中「XX（无质保）」这类名字，把首屏最响的位置浪费掉。
 */
const heroPromo = computed(() => {
  const { product, savingMinor, minPriceMinor } = heroPromoPlan.value
  if (product) {
    return {
      title: product.name,
      subtitle: savingMinor > 0 ? '官方渠道代充值 · 下单立减' : '官方渠道直供 · 1-3 分钟到账',
      badge:
        savingMinor > 0
          ? `省 ¥${formatCNY(savingMinor)}`
          : `¥${formatCNY(product.price_cny_minor)}`,
      actionText: '立即下单'
    }
  }
  return {
    title: 'ChatGPT Plus / Pro 会员代充',
    subtitle: '1-3 分钟到账 · 30 天质保 · 不需要账号密码',
    badge: minPriceMinor !== null ? `最低 ¥${formatCNY(minPriceMinor)} 起` : '查看套餐',
    actionText: '查看套餐'
  }
})

/** 首屏促销条：点击直接进入购买环节 */
function onPromoAction() {
  const target = heroPromoPlan.value.product || heroProducts.value[0]
  if (target) {
    openSheet(target)
    return
  }
  document.getElementById('plans')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const HERO_FEATURES = [
  {
    label: '不需要账号密码',
    path: 'M12 3l7 3v6c0 4.2-2.9 7.4-7 9-4.1-1.6-7-4.8-7-9V6z'
  },
  {
    label: '1-3 分钟到账',
    path: 'M12 7v5l3.2 1.9M12 3a9 9 0 100 18 9 9 0 000-18z'
  },
  {
    label: '30 天质保',
    path: 'M4.5 12.5l4.5 4.5L19.5 6.5'
  }
]

/** 跑马灯：深色段与浅色段之间的过渡带，一列服务承诺横向滚动 */
const MARQUEE: string[] = [
  'ChatGPT Plus / Pro 官方渠道代充',
  '成品号独享交付',
  'X 蓝V 认证',
  'Gemini Advanced 会员',
  '支付宝 / 微信支付',
  '支付后凭订单号查进度',
  '非人为中断 30 天质保',
  '全程不需要账号密码'
]

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
        <!-- 五层叠加：极光 → 透镜 → 内容 → 颗粒。纯 CSS，无 JS / canvas / 视频 -->
        <span class="hero__glow hero__glow--a" aria-hidden="true" />
        <span class="hero__glow hero__glow--b" aria-hidden="true" />
        <span class="hero__glow hero__glow--c" aria-hidden="true" />
        <span class="hero__glow hero__glow--d" aria-hidden="true" />
        <span class="hero__lens" aria-hidden="true" />
        <span class="hero__grain" aria-hidden="true" />

        <div class="hero__inner">
          <div class="hero__copy">
            <p class="hero__eyebrow">
              <span class="hero__eyebrow-dot" aria-hidden="true" />
              ChatGPT Plus / Pro 充值 · 成品号直供
            </p>
            <h1 class="hero__title">
              <span class="hero__title-line hero__title-line--lead">把 ChatGPT 会员这件小事</span>
              <span class="hero__title-line hero__title-line--emphasis">一分钟交给我们</span>
            </h1>
            <p class="hero__desc">
              官方渠道代充值，1-3 分钟到账，30 天质保。成品号与租号即买即用，全程不需要账号密码。
            </p>
            <div class="hero__actions">
              <a class="sh-btn sh-btn--primary" href="#plans">选择套餐</a>
              <a class="sh-btn sh-btn--dark-ghost" href="#workflow">查看充值流程</a>
            </div>
            <ul class="hero__features">
              <li v-for="feature in HERO_FEATURES" :key="feature.label" class="hero__feature">
                <span class="hero__feature-icon" aria-hidden="true">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
                    <path :d="feature.path" />
                  </svg>
                </span>
                <span>{{ feature.label }}</span>
              </li>
            </ul>
          </div>

          <div class="hero__float">
            <HeroFloatCards v-if="heroProducts.length" :products="heroProducts" @select="openSheet" />
          </div>
        </div>

        <div class="hero__wide">
          <HeroPromo
            :title="heroPromo.title"
            :subtitle="heroPromo.subtitle"
            :badge="heroPromo.badge"
            :action-text="heroPromo.actionText"
            @action="onPromoAction"
          />
        </div>
      </section>

      <div class="marquee" aria-hidden="true">
        <div class="marquee__track">
          <div v-for="pass in 2" :key="pass" class="marquee__item">
            <template v-for="(item, index) in MARQUEE" :key="`${pass}-${index}`">
              <span>{{ item }}</span>
              <span class="marquee__dot" />
            </template>
          </div>
        </div>
      </div>

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
          <header class="section__head" v-reveal>
            <span class="section__no">01 — 选择套餐</span>
            <h2 class="section__title">按你要的服务挑一个</h2>
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
          <header class="section__head" v-reveal>
            <span class="section__no">02 — 下单流程</span>
            <h2 class="section__title">三步完成充值</h2>
            <p class="section__desc">先选套餐付款，再补交凭证，全程不需要账号密码</p>
          </header>

          <div class="workflow__grid">
            <article class="step" v-reveal="0">
              <span class="step__num">01</span>
              <h3>选择套餐</h3>
              <p>在上方选择代充值、成品号或租号，确认价格与服务说明</p>
            </article>
            <article class="step" v-reveal="90">
              <span class="step__num">02</span>
              <h3>留联系方式并支付</h3>
              <p>填写手机号或邮箱用于查单与售后，选择支付宝或微信完成支付</p>
            </article>
            <article class="step" v-reveal="180">
              <span class="step__num">03</span>
              <h3>提交信息等待到账</h3>
              <p>代充值粘贴登录凭证，成品号与租号留接收邮箱，随后等待交付</p>
            </article>
          </div>

          <div class="guide-wrap" v-reveal>
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
          <header class="section__head" v-reveal>
            <span class="section__no">03 — 怎么选</span>
            <h2 class="section__title">三种方式怎么选</h2>
            <p class="section__desc">按你现在的情况对号入座</p>
          </header>
          <div class="compare__grid">
            <article v-for="(col, i) in COMPARISON" :key="col.name" class="compare__card" v-reveal="i * 90">
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
          <header class="section__head" v-reveal>
            <span class="section__no">04 — 常见问题</span>
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

      <section class="closing" v-reveal>
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
  --sh-shadow: 0 1px 2px rgba(9, 9, 11, 0.04), 0 1px 3px rgba(9, 9, 11, 0.04);

  /* ---------- 深色段令牌 ----------
     导航 / 首屏 / 跑马灯这一段永远是深色（不跟随明暗主题），
     与下方浅色转化段靠 --sh-* 各管各的，两套互不干扰。 */
  --dk-bg: #0b0d12;
  --dk-fg: #f6f7fb;
  --dk-fg-2: rgba(246, 247, 251, 0.64);
  --dk-fg-3: rgba(246, 247, 251, 0.4);
  --dk-border: rgba(255, 255, 255, 0.1);
  --dk-border-strong: rgba(255, 255, 255, 0.22);
  /* 极光三色：暖橙（品牌）＋蓝紫＋青，比纯冷色更有温度。
     取值刻意偏高——光斑本身是极低饱和的宽渐变，叠两层遮罩后实际观感会再降一档。 */
  --dk-glow-1: rgba(255, 122, 69, 0.52);
  --dk-glow-2: rgba(124, 92, 255, 0.48);
  --dk-glow-3: rgba(34, 211, 238, 0.4);
  --dk-glow-4: rgba(255, 77, 109, 0.34);
  --dk-accent: #ff6b35;
  --dk-accent-2: #ff4d6d;

  min-height: 100vh;
  background: var(--sh-bg);
  color: var(--sh-text);
  font-family: inherit;
  -webkit-font-smoothing: antialiased;
}


/* 吸顶导航高 60px：锚点跳转时留出余量，否则标题会被导航盖住 */
section[id],
.plans__group[id] {
  scroll-margin-top: 76px;
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

/* ---------- 滚动入场 ----------
   .reveal 是初始态（透明 + 下移 14px），.is-revealed 由 v-reveal 指令在进入视口时加上。
   只动 opacity / transform，不动 margin / height，避免触发 layout。 */
.reveal {
  opacity: 0;
  transform: translate3d(0, 14px, 0);
  transition: opacity 620ms cubic-bezier(0.22, 0.61, 0.36, 1), transform 620ms cubic-bezier(0.22, 0.61, 0.36, 1);
}

.reveal.is-revealed {
  opacity: 1;
  transform: translate3d(0, 0, 0);
}

/* 分段编号：01 / 02 / 03 —— 极细的排版锚点，给浅色段一点 editorial 感 */
.section__no {
  display: block;
  margin-bottom: 14px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.16em;
  color: var(--sh-text-3);
  font-variant-numeric: tabular-nums;
}

@media (prefers-reduced-motion: reduce) {
  .reveal {
    opacity: 1;
    transform: none;
    transition: none;
  }
}

/* ---------- nav（深色常驻，与首屏共用一套令牌） ---------- */
.nav {
  position: sticky;
  top: 0;
  z-index: 40;
  border-bottom: 1px solid var(--dk-border);
  background: color-mix(in srgb, var(--dk-bg) 76%, transparent);
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
  background: linear-gradient(140deg, var(--dk-accent), var(--dk-accent-2));
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}

.nav__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--dk-fg);
  letter-spacing: -0.01em;
}

.nav__links {
  display: flex;
  gap: 28px;
  font-size: 14px;
}

.nav__links a {
  color: var(--dk-fg-2);
  text-decoration: none;
  transition: color 160ms ease-out;
}

.nav__links a:hover {
  color: var(--dk-fg);
}

.nav__actions {
  display: flex;
  align-items: center;
  gap: 14px;
}

.nav__ghost {
  font-size: 14px;
  color: var(--dk-fg-2);
  text-decoration: none;
  transition: color 160ms ease-out;
}

.nav__ghost:hover {
  color: var(--dk-fg);
}

.nav__cta {
  padding: 8px 15px;
  border-radius: 8px;
  background: linear-gradient(100deg, var(--dk-accent), var(--dk-accent-2));
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  transition: filter 160ms ease-out;
}

.nav__cta:hover {
  filter: brightness(1.08);
}

/* ---------- hero：深色沉浸首屏 ----------
   五层叠加：极光（唯一色彩来源）→ 透镜（景深）→ 内容 → 颗粒（质感）→ 悬浮商品卡。
   全部纯 CSS —— 无 JS、无 canvas、无视频；动画只动 opacity / transform（GPU 合成）。 */
.hero {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  background: var(--dk-bg);
  color: var(--dk-fg);
  padding: 88px 0 96px;
}

/* 1) 极光层：三团光斑极缓呼吸。这是整段唯一的「彩色」来源，其余一律中性灰阶。 */
.hero__glow {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
  will-change: opacity, transform;
}

.hero__glow--a {
  top: -22%;
  left: -12%;
  width: 58vw;
  height: 58vw;
  background: radial-gradient(circle at 50% 50%, var(--dk-glow-1) 0%, transparent 66%);
  animation: dk-glow-a 34s ease-in-out infinite alternate;
}

.hero__glow--b {
  top: -14%;
  right: -14%;
  width: 54vw;
  height: 54vw;
  background: radial-gradient(circle at 50% 50%, var(--dk-glow-2) 0%, transparent 66%);
  animation: dk-glow-b 44s ease-in-out infinite alternate;
}

.hero__glow--c {
  bottom: -30%;
  left: 24%;
  width: 48vw;
  height: 48vw;
  background: radial-gradient(circle at 50% 50%, var(--dk-glow-3) 0%, transparent 68%);
  animation: dk-glow-c 38s ease-in-out infinite alternate;
}

/* 第四团暖粉：补在右侧中段，把「冷→暖」的过渡补齐，避免整屏只剩蓝紫 */
.hero__glow--d {
  top: 22%;
  right: 2%;
  width: 34vw;
  height: 34vw;
  background: radial-gradient(circle at 50% 50%, var(--dk-glow-4) 0%, transparent 70%);
  animation: dk-glow-d 41s ease-in-out infinite alternate;
}

@keyframes dk-glow-a {
  from {
    opacity: 0.6;
    transform: translate3d(0, 0, 0) scale(1);
  }
  to {
    opacity: 1;
    transform: translate3d(3%, -2%, 0) scale(1.08);
  }
}

@keyframes dk-glow-b {
  from {
    opacity: 0.5;
    transform: translate3d(0, 0, 0) scale(1.05);
  }
  to {
    opacity: 0.92;
    transform: translate3d(-3%, 3%, 0) scale(1);
  }
}

@keyframes dk-glow-c {
  from {
    opacity: 0.36;
    transform: translate3d(0, 2%, 0) scale(1.02);
  }
  to {
    opacity: 0.72;
    transform: translate3d(2%, -3%, 0) scale(1.1);
  }
}

@keyframes dk-glow-d {
  from {
    opacity: 0.42;
    transform: translate3d(0, 0, 0) scale(1.04);
  }
  to {
    opacity: 0.86;
    transform: translate3d(-4%, 4%, 0) scale(0.98);
  }
}

/* 2) 透镜层：一圈椭圆描边 + 三重 inset 阴影 + 上下渐隐，制造「往里看」的景深 */
.hero__lens {
  position: absolute;
  inset: -14% -6% 8%;
  border-radius: 48% / 42%;
  border: 1px solid rgba(255, 255, 255, 0.05);
  box-shadow:
    inset 7rem 0 12rem -9rem rgba(255, 255, 255, 0.18),
    inset -7rem 0 12rem -9rem rgba(255, 255, 255, 0.18),
    inset 0 -5rem 10rem -8rem rgba(255, 255, 255, 0.12);
  /* 上下都要完全淡出：椭圆下弧如果只淡到一半，会在促销条上方留一道突兀的白线 */
  -webkit-mask-image: linear-gradient(transparent 4%, #000 22% 68%, transparent 86%);
  mask-image: linear-gradient(transparent 4%, #000 22% 68%, transparent 86%);
  pointer-events: none;
}

/* 3) 颗粒层：一张 180×180 的静态 SVG 噪声平铺，去掉数字塑料感。
   它是静态图片，没有运行时开销——整套视觉里性价比最高的一层。 */
.hero__grain {
  position: absolute;
  inset: 0;
  opacity: 0.16;
  mix-blend-mode: overlay;
  pointer-events: none;
  background-repeat: repeat;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='180' height='180'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.82' numOctaves='3' stitchTiles='stitch'/%3E%3CfeColorMatrix type='saturate' values='0'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='.55'/%3E%3C/svg%3E");
}

/* 系统开启「减弱动态效果」时退化为静态光斑 */
@media (prefers-reduced-motion: reduce) {
  .hero__glow {
    animation: none;
    opacity: 0.7;
  }
}

.hero__inner {
  position: relative;
  z-index: 2;
  max-width: 1120px;
  margin: 0 auto;
  padding: 0 24px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 384px;
  gap: 56px;
  align-items: center;
}

/* 右侧悬浮卡容器：桌面端给足 3D 位移的呼吸空间，窄屏由媒体查询收窄 */
.hero__float {
  position: relative;
  min-width: 0;
}

.hero__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 999px;
  border: 1px solid var(--dk-border);
  background: rgba(255, 255, 255, 0.04);
  font-size: 13px;
  color: var(--dk-fg-2);
}

.hero__eyebrow-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #34d399;
  box-shadow: 0 0 0 3px rgba(52, 211, 153, 0.16);
}

.hero__title {
  margin-top: 24px;
  font-size: clamp(2.5rem, 4.4vw, 3.6rem);
  line-height: 1.06;
  font-weight: 600;
  letter-spacing: -0.045em;
  text-wrap: balance;
  color: var(--dk-fg);
}

.hero__title-line {
  display: block;
}

.hero__title-line--lead {
  font-size: 0.78em;
  font-weight: 500;
  letter-spacing: -0.035em;
  color: var(--dk-fg-2);
}

/* 第二行做渐变文字：纯白 → 暖橙 → 纯白，把视觉重心压在这一句上 */
.hero__title-line--emphasis {
  margin-top: 0.08em;
  font-size: 1.04em;
  font-weight: 700;
  letter-spacing: -0.025em;
  background: linear-gradient(104deg, #ffffff 6%, #ffb08a 54%, #ffffff 96%);
  -webkit-text-fill-color: transparent;
  background-clip: text;
  -webkit-background-clip: text;
}

.hero__desc {
  margin-top: 20px;
  max-width: 520px;
  font-size: 15px;
  line-height: 1.75;
  color: var(--dk-fg-2);
}

.hero__actions {
  margin-top: 30px;
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.sh-btn--dark-ghost {
  border: 1px solid var(--dk-border-strong);
  color: var(--dk-fg);
  background: transparent;
}

.sh-btn--dark-ghost:hover {
  background: rgba(255, 255, 255, 0.07);
}

/* 促销条容器：跟在两栏内容下方，整幅居中，抢到最大视觉宽度 */
.hero__wide {
  position: relative;
  z-index: 2;
  max-width: 1120px;
  margin: 52px auto 0;
  padding: 0 24px;
  display: flex;
  justify-content: center;
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

/* 特性胶囊：小圆图标 + 短语，替掉原来那串带圆点的文字列表 */
.hero__features {
  margin: 26px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  max-width: 520px;
}

.hero__feature {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  line-height: 1.3;
  color: var(--dk-fg-2);
}

.hero__feature-icon {
  flex: none;
  width: 26px;
  height: 26px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  border: 1px solid var(--dk-border);
  background: rgba(255, 255, 255, 0.06);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.16);
  color: var(--dk-fg);
}

.hero__feature-icon svg {
  width: 13px;
  height: 13px;
}

/* ---------- 跑马灯：深色段与浅色段之间的过渡带 ---------- */
.marquee {
  background: var(--dk-bg);
  border-top: 1px solid var(--dk-border);
  overflow: hidden;
  padding: 20px 0;
  -webkit-mask-image: linear-gradient(90deg, transparent, #000 8% 92%, transparent);
  mask-image: linear-gradient(90deg, transparent, #000 8% 92%, transparent);
}

/* 轨道里放两份同样的列表，位移 50% 即可无缝衔接。
   注意：轨道本身不能有 gap（否则 -50% 会差半个间距，循环时出现跳动），
   间距一律放在 .marquee__item 内部 + padding-right，这样两份列表宽度严格相等。 */
.marquee__track {
  display: flex;
  width: max-content;
  animation: dk-marquee 36s linear infinite;
}

.marquee__item {
  display: flex;
  align-items: center;
  gap: 40px;
  padding-right: 40px;
  font-size: 14px;
  color: rgba(246, 247, 251, 0.52);
  white-space: nowrap;
}

.marquee__dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--dk-border-strong);
}

@keyframes dk-marquee {
  from {
    transform: translate3d(0, 0, 0);
  }
  to {
    transform: translate3d(-50%, 0, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .marquee__track {
    animation: none;
  }
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

  /* 悬浮卡在窄屏改为横向滑动条（组件内部已处理），这里只需让它占满栏宽 */
  .hero__float {
    max-width: 560px;
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

  .hero__features {
    margin-top: 22px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px 12px;
  }

  .hero__wide {
    margin-top: 36px;
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
}

/* 锚点平滑滚动（导航「套餐 / 流程 / 常见问题」）。放在非 scoped 块里才能作用到 html。 */
html:has(.sales-home) {
  scroll-behavior: smooth;
}

@media (prefers-reduced-motion: reduce) {
  html:has(.sales-home) {
    scroll-behavior: auto;
  }
}
</style>
