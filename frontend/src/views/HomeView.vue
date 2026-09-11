<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import FAQ_ITEMS from '@/content/home-faq.json'
import { useAppStore } from '@/stores/app'
import { publicShopAPI, type FulfillmentMode, type PublicBanner, type PublicProduct } from '@/api/publicShop'
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
const activeTab = ref<FulfillmentMode>('session_topup')
const sheetOpen = ref(false)
const selectedProduct = ref<PublicProduct | null>(null)
const openFaq = ref<number | null>(0)

const TABS: { value: FulfillmentMode; label: string; blurb: string }[] = [
  { value: 'session_topup', label: '代充值', blurb: '给自己已有的 ChatGPT 账号续费升级' },
  { value: 'account_delivery', label: '成品号', blurb: '直接拿一个开通好的独享账号' },
  { value: 'rental', label: '租号', blurb: '按周或按月短期使用，成本更低' }
]

const availableTabs = computed(() => TABS.filter((tab) => products.value.some((item) => item.fulfillment_mode === tab.value)))
const visibleProducts = computed(() => products.value.filter((item) => item.fulfillment_mode === activeTab.value))
const activeTabBlurb = computed(() => availableTabs.value.find((tab) => tab.value === activeTab.value)?.blurb || '')
const contactInfo = computed(() => appStore.contactInfo?.trim() || '')
// 下单抽屉里的支付方式：后台配置 ∩ 当前商品金额在单笔限额内
const sheetPaymentMethods = computed(() =>
  selectedProduct.value ? methodsForAmount(selectedProduct.value.price_cny_minor) : []
)

const ASSURANCE = [
  { title: '官方渠道充值', desc: '按官方订阅流程开通，不开后门、不刷额度' },
  { title: '1-3 分钟到账', desc: '凭证提交后自动进入处理队列' },
  { title: '30 天质保', desc: '非用户原因中断，按条款处理' },
  { title: '订单可查', desc: '凭订单号 + 联系方式随时看进度' },
  { title: '正规发票', desc: '按实际金额开票，报销无忧' },
  { title: '边界透明', desc: '第三方独立服务，非 OpenAI 官方' }
]

const METRICS = [
  { value: '1-3', unit: '分钟', label: '平均到账时间' },
  { value: '30', unit: '天', label: '非人为中断质保' },
  { value: '0', unit: '个', label: '索要的账号密码' },
  { value: '24', unit: '小时', label: '售后响应窗口' }
]

const COMPARISON = [
  { label: '适合谁', topup: '已有账号要续费', account: '想换个新号', rental: '短期试用' },
  { label: '需要提供', topup: '登录凭证', account: '接收邮箱', rental: '接收邮箱' },
  { label: '账号归属', topup: '你自己的号', account: '交付给你', rental: '租期内使用' },
  { label: '交付时效', topup: '1-3 分钟', account: '人工发货', rental: '人工发货' },
  { label: '成本', topup: '低', account: '中', rental: '最低' }
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
    const first = availableTabs.value[0]
    if (first && !products.value.some((item) => item.fulfillment_mode === activeTab.value)) {
      activeTab.value = first.value
    }
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
  activeTab.value = product.fulfillment_mode
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
        <div class="hero__inner">
          <div class="hero__copy">
            <p class="eyebrow">ChatGPT Plus / Pro 充值 · 成品号直供</p>
            <h1 class="hero__title">给你的 ChatGPT 续上<br>官方会员与独享账号</h1>
            <p class="hero__desc">
              官方渠道代充值，1-3 分钟到账，30 天质保，可开发票。
              不想动自己的账号？成品号与租号即买即用，全程不需要账号密码。
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

          <div v-if="availableTabs.length" class="sh-tabs">
            <div class="sh-tabs__track" role="tablist">
              <button
                v-for="tab in availableTabs"
                :key="tab.value"
                type="button"
                role="tab"
                class="sh-tabs__item"
                :class="{ 'sh-tabs__item--active': activeTab === tab.value }"
                :aria-selected="activeTab === tab.value"
                @click="activeTab = tab.value"
              >
                {{ tab.label }}
              </button>
            </div>
            <p v-if="activeTabBlurb" class="sh-tabs__blurb">{{ activeTabBlurb }}</p>
          </div>

          <div v-if="loading" class="plans__grid">
            <div v-for="i in 3" :key="i" class="sh-skeleton" />
          </div>

          <div v-else-if="visibleProducts.length" class="plans__grid">
            <PlanCard
              v-for="product in visibleProducts"
              :key="product.id"
              :product="product"
              :show-commission="isLoggedIn"
              :aff-code="affCode"
              @select="openSheet"
            />
          </div>

          <div v-else class="empty">
            <p class="empty__title">这个分类的套餐正在准备中</p>
            <p class="empty__desc">可以先看看其他分类，或联系客服咨询</p>
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
            <p class="section__desc">一张表看懂差异</p>
          </header>
          <div class="sh-table-wrap">
            <table class="sh-table">
              <thead>
                <tr>
                  <th />
                  <th>代充值</th>
                  <th>成品号</th>
                  <th>租号</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in COMPARISON" :key="row.label">
                  <th scope="row">{{ row.label }}</th>
                  <td>{{ row.topup }}</td>
                  <td>{{ row.account }}</td>
                  <td>{{ row.rental }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section class="assurance">
        <div class="section__inner">
          <header class="section__head">
            <h2 class="section__title">服务保障</h2>
            <p class="section__desc">流程写清楚，售后才有据可依</p>
          </header>
          <div class="assurance__grid">
            <article v-for="item in ASSURANCE" :key="item.title" class="assurance__card">
              <h3>{{ item.title }}</h3>
              <p>{{ item.desc }}</p>
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
        <div class="section__inner section__inner--narrow closing__inner">
          <h2>准备好开始了吗</h2>
          <p>选择套餐后支付，1-3 分钟内完成。需要帮助随时联系客服。</p>
          <div class="closing__actions">
            <a class="sh-btn sh-btn--primary" href="#plans">选择套餐</a>
            <RouterLink class="sh-btn sh-btn--ghost" to="/order">查询订单</RouterLink>
          </div>
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

/* 分段控件：不是卡片，是一个整体 */
.sh-tabs {
  margin-bottom: 32px;
}

.sh-tabs__track {
  display: inline-flex;
  padding: 4px;
  gap: 2px;
  border-radius: 11px;
  border: 1px solid var(--sh-border);
  background: var(--sh-surface-2);
}

.sh-tabs__item {
  padding: 8px 20px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--sh-text-2);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 160ms ease-out, color 160ms ease-out;
}

.sh-tabs__item:hover {
  color: var(--sh-text);
}

.sh-tabs__item--active {
  background: var(--sh-surface);
  color: var(--sh-text);
  box-shadow: var(--sh-shadow);
}

.sh-tabs__blurb {
  margin-top: 12px;
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

.sh-table-wrap {
  overflow-x: auto;
  border-radius: 14px;
  border: 1px solid var(--sh-border);
}

.sh-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.sh-table th,
.sh-table td {
  padding: 16px 20px;
  text-align: left;
  border-bottom: 1px solid var(--sh-border);
}

.sh-table thead th {
  font-weight: 500;
  color: var(--sh-text);
  background: var(--sh-surface-2);
}

.sh-table tbody th {
  font-weight: 400;
  color: var(--sh-text-3);
  white-space: nowrap;
}

.sh-table td {
  color: var(--sh-text);
}

.sh-table tbody tr:last-child th,
.sh-table tbody tr:last-child td {
  border-bottom: none;
}

/* ---------- assurance ---------- */
.assurance {
  padding: 96px 0;
  border-top: 1px solid var(--sh-border);
}

.assurance__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1px;
  background: var(--sh-border);
  border: 1px solid var(--sh-border);
  border-radius: 14px;
  overflow: hidden;
}

.assurance__card {
  padding: 24px;
  background: var(--sh-surface);
}

.assurance__card h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--sh-text);
}

.assurance__card p {
  margin-top: 8px;
  font-size: 14px;
  line-height: 1.7;
  color: var(--sh-text-2);
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
  padding: 96px 0;
  border-top: 1px solid var(--sh-border);
}

.closing__inner {
  text-align: center;
}

.closing h2 {
  font-size: 30px;
  font-weight: 600;
  letter-spacing: -0.025em;
  color: var(--sh-text);
}

.closing p {
  margin-top: 12px;
  font-size: 15px;
  color: var(--sh-text-2);
}

.closing__actions {
  margin-top: 32px;
  display: flex;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;
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
  .assurance__grid {
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
  .assurance,
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

  /* 分段控件横向滚动，避免竖排堆叠 */
  .sh-tabs__track {
    display: flex;
    width: 100%;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .sh-tabs__track::-webkit-scrollbar {
    display: none;
  }

  .sh-tabs__item {
    flex: 1 0 auto;
    text-align: center;
    padding: 8px 16px;
  }

  .plans__grid {
    grid-template-columns: 1fr;
  }

  .sh-table th,
  .sh-table td {
    padding: 13px 14px;
    font-size: 13px;
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
</style>
