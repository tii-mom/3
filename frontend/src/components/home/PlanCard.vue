<script setup lang="ts">
import { computed } from 'vue'
import { formatCNY, type PublicProduct } from '@/api/publicShop'
import { shopCategoryLabel, shopCategoryBenefits, shopModeFact } from '@/constants/shop'
import BrandLogo from '@/components/home/BrandLogo.vue'

const props = defineProps<{
  product: PublicProduct
  active?: boolean
  /** 登录后展示推广奖励 */
  showCommission?: boolean
  affCode?: string
}>()

defineEmits<{ select: [product: PublicProduct] }>()

const featured = computed(() => props.product.highlight)
const soldOut = computed(() => props.product.stock_quantity !== null && props.product.stock_quantity !== undefined && props.product.stock_quantity <= 0)
const hasDiscount = computed(() => props.product.original_price_cny_minor > props.product.price_cny_minor)
const saveAmount = computed(() => Math.round((props.product.original_price_cny_minor - props.product.price_cny_minor) / 100))

const categoryLabel = computed(() => shopCategoryLabel(props.product.category))
const benefits = computed(() => shopCategoryBenefits(props.product.category))
const modeFact = computed(() => shopModeFact(props.product.fulfillment_mode))

/** 价格右侧单位：从名称/规格里识别 月/季/年/周/天/次，识别不到就不显示 */
const priceUnit = computed(() => {
  const hay = `${props.product.spec_label} ${props.product.name}`.toLowerCase()
  if (hay.includes('月')) return '/月'
  if (hay.includes('季')) return '/季'
  if (hay.includes('年')) return '/年'
  if (hay.includes('周')) return '/周'
  if (hay.includes('天')) return '/天'
  if (hay.includes('次')) return '/次'
  return ''
})

const commissionPercent = computed(() => Math.round((props.product.commission_bps || 0) / 100))
const showCommission = computed(() => !!props.showCommission && (props.product.commission_bps || 0) > 0)

/** 角标优先级：有折扣显示「立省 ¥X」（按金额，不按百分比，避免误导），否则显示后台填的角标文字 */
const badgeText = computed(() => {
  if (hasDiscount.value) return `立省 ¥${saveAmount.value}`
  return props.product.badge_text || ''
})
</script>

<template>
  <article
    class="tier-card"
    :class="{ 'tier-card--featured': featured, 'tier-card--soldout': soldOut }"
  >
    <div v-if="featured" class="tier-card__ribbon">最受欢迎</div>

    <div class="tier-card__head">
      <BrandLogo :product="product" :size="46" />
      <div class="tier-card__heading">
        <h3 class="tier-card__name">{{ product.name }}</h3>
        <p class="tier-card__cat">{{ categoryLabel }}</p>
      </div>
      <span v-if="badgeText" class="tier-card__badge">{{ badgeText }}</span>
    </div>

    <p v-if="product.description" class="tier-card__desc">{{ product.description }}</p>

    <div class="tier-card__price">
      <div class="tier-card__price-now">
        <span class="tier-card__currency">¥</span>
        <span class="tier-card__amount">{{ formatCNY(product.price_cny_minor) }}</span>
        <span v-if="priceUnit" class="tier-card__unit">{{ priceUnit }}</span>
      </div>
    </div>

    <button
      type="button"
      class="tier-card__cta"
      :disabled="soldOut"
      @click="$emit('select', product)"
    >
      {{ soldOut ? '暂时缺货' : '立即下单' }}
    </button>

    <div class="tier-card__panel">
      <p class="tier-card__panel-label">
        {{ modeFact.label }}<template v-if="product.spec_label"> · {{ product.spec_label }}</template>
      </p>
      <div class="tier-card__panel-divider" />
      <div class="tier-card__panel-row"><span>到账时间</span><span>{{ modeFact.turnaround }}</span></div>
      <div class="tier-card__panel-row"><span>账号归属</span><span>{{ modeFact.belongs }}</span></div>
      <div v-if="product.delivery_form_hint" class="tier-card__panel-hint">{{ product.delivery_form_hint }}</div>
    </div>

    <div class="tier-card__save">
      <span v-if="product.sold_count > 0" class="tier-card__sold">已售 {{ product.sold_count }} 份</span>
    </div>

    <div class="tier-card__divider" />

    <div class="tier-card__benefits">
      <p class="tier-card__benefits-label">套餐包含</p>
      <ul>
        <li v-for="(b, i) in benefits" :key="i">{{ b }}</li>
      </ul>
    </div>

    <div v-if="showCommission" class="tier-card__footer">
      <span>推广奖励</span>
      <span class="tier-card__footer-amt">{{ commissionPercent }}%</span>
    </div>
  </article>
</template>

<style scoped>
/* 自包含令牌：卡片在任何页面（首页 / 商城）都一致，不依赖父级 --sh-* 变量 */
.tier-card {
  --tc-bg: #ffffff;
  --tc-bg-soft: #f6f6f7;
  --tc-border: rgba(9, 9, 11, 0.12);
  --tc-border-strong: rgba(9, 9, 11, 0.2);
  --tc-text: #09090b;
  --tc-text-2: #56565f;
  --tc-text-3: #8a8a93;
  --tc-accent: #d85a30;
  --tc-accent-soft: rgba(216, 90, 48, 0.08);
  --tc-accent-text: #b34b1f;
  --tc-accent-fg: #ffffff;
  --tc-accent-hover: #c04f21;
  --tc-save: #993c1d;
  --tc-save-soft: #faece7;
  --tc-commission: #0f6e56;
  --tc-commission-soft: rgba(15, 110, 86, 0.08);

  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  border: 1px solid var(--tc-border);
  border-radius: 14px;
  background: var(--tc-bg);
  color: var(--tc-text);
  transition: border-color 180ms ease-out, transform 180ms ease-out;
}

.tier-card:hover {
  border-color: var(--tc-border-strong);
  transform: translateY(-2px);
}

.tier-card--soldout {
  opacity: 0.6;
}

/* 推荐位：仅它用品牌橙描边 + 浅橙底 + 实心 CTA，全站唯一强调色 */
.tier-card--featured {
  border-color: var(--tc-accent);
  background: #fdf7f4;
}

.tier-card__ribbon {
  position: absolute;
  top: -11px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--tc-text);
  color: #fff;
  font-size: 11px;
  font-weight: 500;
  padding: 3px 11px;
  border-radius: 999px;
  white-space: nowrap;
}

.tier-card__head {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 0;
}

.tier-card__heading {
  min-width: 0;
  flex: 1 1 auto;
}

.tier-card__cat {
  margin: 3px 0 0;
  font-size: 12px;
  color: var(--tc-text-3);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tier-card__badge {
  flex: none;
  margin-left: auto;
  font-size: 11px;
  background: var(--tc-save-soft);
  color: var(--tc-save);
  padding: 2px 7px;
  border-radius: 6px;
  white-space: nowrap;
}

.tier-card--featured .tier-card__badge {
  background: var(--tc-accent);
  color: var(--tc-accent-fg);
}

.tier-card__head {
  min-height: 0;
}

.tier-card__name {
  margin: 0;
  font-size: 20px;
  font-weight: 500;
  letter-spacing: -0.015em;
  line-height: 1.35;
  color: var(--tc-text);
}

.tier-card__desc {
  margin: 4px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--tc-text-2);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.tier-card__price {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.tier-card__price-now {
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.tier-card__currency {
  font-size: 16px;
  color: var(--tc-text-3);
}

.tier-card__amount {
  font-size: 32px;
  font-weight: 500;
  letter-spacing: -0.03em;
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
  color: var(--tc-text);
}

.tier-card__unit {
  font-size: 13px;
  color: var(--tc-text-3);
  font-weight: 400;
}

.tier-card__cta {
  height: 40px;
  border-radius: 9px;
  border: 1px solid var(--tc-border-strong);
  background: transparent;
  color: var(--tc-text);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 160ms ease-out, border-color 160ms ease-out;
}

.tier-card__cta:hover:not(:disabled) {
  background: var(--tc-bg-soft);
}

.tier-card--featured .tier-card__cta {
  border-color: var(--tc-accent);
  background: var(--tc-accent);
  color: var(--tc-accent-fg);
}

.tier-card--featured .tier-card__cta:hover:not(:disabled) {
  background: var(--tc-accent-hover);
  border-color: var(--tc-accent-hover);
}

.tier-card__cta:disabled {
  cursor: not-allowed;
  background: var(--tc-bg-soft);
  border-color: var(--tc-border);
  color: var(--tc-text-3);
}

/* 着色权益面板：把交付模式事实 + 规格说明收进一处 */
.tier-card__panel {
  background: var(--tc-bg-soft);
  border-radius: 10px;
  padding: 10px 11px;
}

.tier-card--featured .tier-card__panel {
  background: var(--tc-save-soft);
}

.tier-card__panel-label {
  margin: 0;
  font-size: 11px;
  color: var(--tc-text-2);
}

.tier-card--featured .tier-card__panel-label {
  color: var(--tc-accent-text);
}

.tier-card__panel-divider {
  height: 1px;
  background: var(--tc-border);
  margin: 8px 0;
}

.tier-card--featured .tier-card__panel-divider {
  background: rgba(216, 90, 48, 0.22);
}

.tier-card__panel-row {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  line-height: 1.5;
}

.tier-card__panel-row > span:first-child {
  color: var(--tc-text-3);
}

.tier-card__panel-row > span:last-child {
  color: var(--tc-text);
}

.tier-card__panel-hint {
  margin-top: 7px;
  font-size: 11px;
  line-height: 1.5;
  color: var(--tc-text-2);
}

.tier-card__save {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: 11px;
  min-height: 14px;
}

.tier-card__sold {
  color: var(--tc-text-3);
}

.tier-card__divider {
  height: 1px;
  background: var(--tc-border);
}

.tier-card__benefits-label {
  margin: 0 0 7px;
  font-size: 11px;
  color: var(--tc-text-3);
}

.tier-card--featured .tier-card__benefits-label {
  color: var(--tc-accent-text);
}

.tier-card__benefits ul {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* 勾选式标记，比圆点更有「规格清单」语义（与旧 PlanCard 一致） */
.tier-card__benefits li {
  position: relative;
  padding-left: 18px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--tc-text-2);
}

.tier-card__benefits li::before {
  content: '';
  position: absolute;
  left: 2px;
  top: 6px;
  width: 8px;
  height: 4px;
  border-left: 1.5px solid var(--tc-text-3);
  border-bottom: 1.5px solid var(--tc-text-3);
  transform: rotate(-45deg);
}

/* 虚线 bonus 页脚：推广奖励 */
.tier-card__footer {
  margin-top: auto;
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  border: 1px dashed var(--tc-border-strong);
  border-radius: 9px;
  padding: 9px 11px;
  font-size: 11px;
  color: var(--tc-text-2);
}

.tier-card--featured .tier-card__footer {
  border-color: rgba(216, 90, 48, 0.4);
}

.tier-card__footer-amt {
  font-size: 15px;
  font-weight: 500;
  color: var(--tc-commission);
}
</style>

<style>
/* 深色覆盖放在**非 scoped** 块中：Vue 的 scoped-CSS 编译器会把
   `:global(.dark) X` 编译成只有 `.dark`，X 被丢弃，导致生产构建深色规则失效。 */
.dark .tier-card {
  --tc-bg: #18181b;
  --tc-bg-soft: #232429;
  --tc-border: rgba(255, 255, 255, 0.12);
  --tc-border-strong: rgba(255, 255, 255, 0.22);
  --tc-text: #f4f4f5;
  --tc-text-2: #a1a1aa;
  --tc-text-3: #71717a;
  --tc-accent: #e8743f;
  --tc-accent-soft: rgba(232, 116, 63, 0.14);
  --tc-accent-text: #f0a574;
  --tc-accent-fg: #1a1206;
  --tc-accent-hover: #d8652f;
  --tc-save: #e0ac52;
  --tc-save-soft: rgba(224, 172, 82, 0.12);
  --tc-commission: #6fd8b0;
  --tc-commission-soft: rgba(111, 216, 176, 0.1);
}

.dark .tier-card--featured {
  background: #211a16;
}
</style>
