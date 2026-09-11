<script setup lang="ts">
import { computed, ref } from 'vue'
import { formatCNY, type PublicProduct } from '@/api/publicShop'
import { resolveShopAssetUrl } from '@/api/shop'

const props = defineProps<{
  product: PublicProduct
  active?: boolean
  /** 登录后展示推广奖励与推广链接 */
  showCommission?: boolean
  affCode?: string
}>()

defineEmits<{ select: [product: PublicProduct] }>()

/** 主图 + 图廊合成一组；hover 时切到第二张，移开回到主图（不做自动轮播） */
const images = computed(() =>
  [props.product.image_url, ...(props.product.gallery || [])]
    .map((url) => resolveShopAssetUrl(url))
    .filter(Boolean)
)
const hovered = ref(false)
const displayImage = computed(() => (hovered.value ? images.value[1] || images.value[0] : images.value[0]) || '')

const soldOut = computed(() => props.product.stock_quantity !== null && props.product.stock_quantity !== undefined && props.product.stock_quantity <= 0)
const lowStock = computed(() => !soldOut.value && props.product.stock_quantity !== null && props.product.stock_quantity !== undefined && props.product.stock_quantity <= 5)
const hasDiscount = computed(() => props.product.original_price_cny_minor > props.product.price_cny_minor)
const hint = computed(() => props.product.delivery_form_hint?.trim() || '')
const commissionPercent = computed(() => Math.round((props.product.commission_bps || 0) / 100))
const showCommission = computed(() => !!props.showCommission && (props.product.commission_bps || 0) > 0)
const stockText = computed(() => {
  if (soldOut.value) return '暂时缺货'
  if (props.product.stock_quantity === null || props.product.stock_quantity === undefined) return ''
  return props.product.stock_quantity <= 5 ? `仅剩 ${props.product.stock_quantity} 件` : `库存 ${props.product.stock_quantity}`
})

async function copyPromotionLink() {
  if (!props.affCode) return
  const url = new URL('/shop', window.location.origin)
  url.searchParams.set('product', String(props.product.id))
  url.searchParams.set('aff', props.affCode)
  try {
    await navigator.clipboard.writeText(url.toString())
  } catch {
    // 复制失败忽略，用户可手动复制
  }
}
</script>

<template>
  <article
    class="plan-card"
    :class="{ 'plan-card--featured': product.highlight, 'plan-card--soldout': soldOut }"
  >
    <div
      v-if="images.length"
      class="plan-card__media"
      @mouseenter="hovered = true"
      @mouseleave="hovered = false"
    >
      <img :src="displayImage" :alt="product.name" loading="lazy">
      <span v-if="images.length > 1" class="plan-card__media-count">{{ images.length }} 图</span>
    </div>

    <header class="plan-card__head">
      <!-- 角标放在标题行内：绝对定位到右上角会和商品图重叠，导致文字压在图上读不清 -->
      <div class="plan-card__title-line">
        <h3 class="plan-card__title">{{ product.name }}</h3>
        <span v-if="product.badge_text" class="plan-card__badge">{{ product.badge_text }}</span>
      </div>
      <p v-if="product.description" class="plan-card__desc">{{ product.description }}</p>
    </header>

    <div class="plan-card__price">
      <span class="plan-card__currency">¥</span>
      <span class="plan-card__amount">{{ formatCNY(product.price_cny_minor) }}</span>
      <span v-if="hasDiscount" class="plan-card__origin">¥{{ formatCNY(product.original_price_cny_minor) }}</span>
    </div>

    <ul class="plan-card__meta">
      <li v-if="product.spec_label">{{ product.spec_label }}</li>
      <li v-if="hint">{{ hint }}</li>
      <li v-if="product.sold_count > 0">已售 {{ product.sold_count }} 份</li>
      <li v-if="stockText" :class="{ 'plan-card__meta--warn': lowStock }">{{ stockText }}</li>
    </ul>

    <div v-if="showCommission" class="plan-card__commission">
      <span>推广奖励 <b>{{ commissionPercent }}%</b></span>
      <button v-if="affCode" type="button" class="plan-card__commission-copy" @click.stop="copyPromotionLink">复制链接</button>
    </div>

    <button
      type="button"
      class="plan-card__cta"
      :disabled="soldOut"
      @click="$emit('select', product)"
    >
      {{ soldOut ? '暂时缺货' : '立即下单' }}
    </button>
  </article>
</template>

<style scoped>
/* 令牌继承自 HomeView 的 .sales-home，明暗主题自动切换 */
.plan-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 26px 22px 22px;
  border: 1px solid var(--sh-border, rgba(9, 9, 11, 0.08));
  border-radius: 14px;
  background: var(--sh-surface, #fff);
  box-shadow: var(--sh-shadow, none);
  transition: border-color 180ms ease-out, transform 180ms ease-out;
}

.plan-card:hover {
  border-color: var(--sh-border-strong, rgba(9, 9, 11, 0.16));
  transform: translateY(-2px);
}

/* 推荐项：只靠描边与顶部分隔条区分，不做放大变形 */
.plan-card--featured {
  border-color: color-mix(in srgb, var(--sh-accent, #d85a28) 42%, transparent);
}

.plan-card--soldout {
  opacity: 0.6;
}

.plan-card__title-line {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.plan-card__badge {
  flex: 0 0 auto;
  margin-top: 2px;
  padding: 3px 9px;
  border-radius: 6px;
  background: var(--sh-accent-soft, rgba(216, 90, 40, 0.08));
  color: var(--sh-accent-text, #b34b1f);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.02em;
  white-space: nowrap;
}

.plan-card__title {
  flex: 1 1 auto;
  min-width: 0;
  font-size: 18px;
  font-weight: 600;
  letter-spacing: -0.015em;
  color: var(--sh-text, #09090b);
}

.plan-card__desc {
  margin-top: 8px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--sh-text-2, #56565f);
  min-height: 44px;
  /* 后台介绍支持换行，这里保留换行与多余空格的语义 */
  white-space: pre-line;
  /* 统一截到两行：同组卡片的图片/价格才能横向对齐，完整介绍在下单抽屉里看 */
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.plan-card__price {
  display: flex;
  align-items: baseline;
  gap: 6px;
  padding-top: 2px;
}

.plan-card__currency {
  font-size: 15px;
  color: var(--sh-text-3, #8a8a93);
}

.plan-card__amount {
  font-size: 34px;
  font-weight: 600;
  letter-spacing: -0.03em;
  color: var(--sh-text, #09090b);
  font-variant-numeric: tabular-nums;
}

.plan-card__origin {
  font-size: 13px;
  color: var(--sh-text-3, #8a8a93);
  text-decoration: line-through;
  font-variant-numeric: tabular-nums;
}

.plan-card__meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0;
  margin: 0;
  list-style: none;
  font-size: 13px;
  line-height: 1.6;
  color: var(--sh-text-2, #56565f);
}

.plan-card__meta li {
  position: relative;
  padding-left: 18px;
}

/* 勾选式标记，比圆点更有"规格清单"的语义 */
.plan-card__meta li::before {
  content: '';
  position: absolute;
  left: 2px;
  top: 7px;
  width: 8px;
  height: 4px;
  border-left: 1.5px solid var(--sh-text-3, #8a8a93);
  border-bottom: 1.5px solid var(--sh-text-3, #8a8a93);
  transform: rotate(-45deg);
}

.plan-card__media {
  position: relative;
  height: 128px;
  margin: -2px 0 2px;
  overflow: hidden;
  border-radius: 10px;
  border: 1px solid var(--sh-border, rgba(9, 9, 11, 0.08));
  background: var(--sh-surface-2, #f6f6f7);
}

.plan-card__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 图廊角标：告诉用户还有更多图，hover 可切换 */
.plan-card__media-count {
  position: absolute;
  right: 8px;
  bottom: 8px;
  padding: 2px 7px;
  border-radius: 999px;
  background: rgba(9, 9, 11, 0.62);
  color: #fff;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.02em;
}

.plan-card__meta--warn {
  color: #c2831f;
}

.plan-card__commission {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 9px 12px;
  border-radius: 8px;
  border: 1px solid rgba(26, 158, 111, 0.24);
  background: rgba(26, 158, 111, 0.07);
  font-size: 12px;
  color: #14795a;
}


.plan-card__commission b {
  font-variant-numeric: tabular-nums;
}

.plan-card__commission-copy {
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid rgba(26, 158, 111, 0.35);
  background: transparent;
  color: inherit;
  font-size: 12px;
  cursor: pointer;
}

/* CTA：仅推荐项用实心强调色，其余走描边，避免一屏全橙 */
.plan-card__cta {
  margin-top: auto;
  height: 44px;
  border-radius: 9px;
  border: 1px solid var(--sh-border-strong, rgba(9, 9, 11, 0.16));
  background: transparent;
  color: var(--sh-text, #09090b);
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: background 160ms ease-out, border-color 160ms ease-out;
}

.plan-card__cta:hover:not(:disabled) {
  background: var(--sh-surface-2, #f6f6f7);
  border-color: var(--sh-text-3, #8a8a93);
}

.plan-card--featured .plan-card__cta {
  border-color: var(--sh-accent, #d85a28);
  background: var(--sh-accent, #d85a28);
  color: var(--sh-accent-fg, #fff);
}

.plan-card--featured .plan-card__cta:hover:not(:disabled) {
  background: var(--sh-accent-hover, #c04f21);
  border-color: var(--sh-accent-hover, #c04f21);
}

.plan-card__cta:disabled {
  cursor: not-allowed;
  background: var(--sh-surface-2, #f6f6f7);
  border-color: var(--sh-border, rgba(9, 9, 11, 0.08));
  color: var(--sh-text-3, #8a8a93);
}
</style>

<style>
/* 深色覆盖放在**非 scoped** 块中：Vue 的 scoped-CSS 编译器会把
   `:global(.dark) X` 编译成只有 `.dark`，X 被丢弃，导致生产构建里
   深色规则整体失效。与 SettingsView 的处理方式保持一致。 */
.dark .plan-card__commission {
  color: #6fd8b0;
}
</style>
