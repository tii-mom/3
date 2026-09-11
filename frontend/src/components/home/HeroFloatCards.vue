<script setup lang="ts">
import { formatCNY, type PublicProduct } from '@/api/publicShop'
import { resolveShopAssetUrl } from '@/api/shop'

/**
 * 首屏右侧的悬浮商品预览卡。
 *
 * 三张真实商品卡做 3D 错位排布 + 9s 极缓浮动，点任意一张直接打开该商品的下单抽屉
 * —— 悬浮元素本身就是购买入口，不只是装饰。
 *
 * 玻璃质感用「径向高光 + 半透明渐变 + 内阴影」三层模拟，
 * 刻意不用 backdrop-filter：多层实时模糊在中低端安卓上掉帧明显。
 */
const props = defineProps<{ products: PublicProduct[] }>()
const emit = defineEmits<{ select: [product: PublicProduct] }>()

function imageOf(product: PublicProduct) {
  return resolveShopAssetUrl(product.image_url)
}

function items() {
  return props.products.slice(0, 3)
}
</script>

<template>
  <div class="floats">
    <button
      v-for="(product, index) in items()"
      :key="product.id"
      type="button"
      class="floats__card"
      :class="`floats__card--${index}`"
      @click="emit('select', product)"
    >
      <span class="floats__media">
        <img v-if="imageOf(product)" :src="imageOf(product)" :alt="product.name" loading="lazy">
        <span v-else class="floats__media-fallback" aria-hidden="true">3</span>
      </span>
      <span class="floats__body">
        <span class="floats__name">{{ product.name }}</span>
        <span class="floats__row">
          <span class="floats__price">¥{{ formatCNY(product.price_cny_minor) }}</span>
          <span v-if="product.badge_text" class="floats__badge">{{ product.badge_text }}</span>
        </span>
      </span>
    </button>
  </div>
</template>

<style scoped>
.floats {
  position: relative;
  height: 400px;
  perspective: 1100px;
  transform-style: preserve-3d;
}

.floats__card {
  --float-rot: 0deg;
  position: absolute;
  width: 262px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 16px;
  border: 1px solid var(--dk-border, rgba(255, 255, 255, 0.1));
  color: var(--dk-fg, #f6f7fb);
  text-align: left;
  cursor: pointer;
  /* 玻璃质感：高处一抹径向高光 + 极淡渐变底 + 顶部内高光 + 大范围落地投影 */
  background:
    radial-gradient(120% 90% at 24% 6%, rgba(255, 255, 255, 0.2), transparent 48%),
    linear-gradient(158deg, rgba(255, 255, 255, 0.1), rgba(255, 255, 255, 0.026));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.2),
    0 26px 60px -26px rgba(0, 0, 0, 0.78);
  transform: rotate(var(--float-rot));
  will-change: transform;
  animation: float-bob 9s ease-in-out infinite;
  transition: border-color 200ms ease-out;
}

.floats__card:hover {
  border-color: var(--dk-border-strong, rgba(255, 255, 255, 0.22));
}

.floats__card--0 {
  top: 0;
  left: 34px;
  z-index: 3;
  --float-rot: -5deg;
}

.floats__card--1 {
  top: 132px;
  left: 0;
  z-index: 2;
  --float-rot: 4deg;
  animation-delay: -3s;
}

.floats__card--2 {
  top: 264px;
  left: 52px;
  z-index: 1;
  --float-rot: -2deg;
  animation-delay: -6s;
}

/* 旋转角从自定义属性读，这样关键帧不会把 rotate 覆盖掉 */
@keyframes float-bob {
  0%,
  100% {
    transform: rotate(var(--float-rot)) translate3d(0, 0, 0);
  }
  50% {
    transform: rotate(var(--float-rot)) translate3d(0, -12px, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .floats__card {
    animation: none;
  }
}

.floats__media {
  flex: none;
  width: 56px;
  height: 56px;
  overflow: hidden;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.06);
}

.floats__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.floats__media-fallback {
  display: grid;
  place-items: center;
  width: 100%;
  height: 100%;
  font-size: 18px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.5);
}

.floats__body {
  min-width: 0;
  display: grid;
  gap: 6px;
}

.floats__name {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  font-size: 13px;
  line-height: 1.35;
  color: rgba(246, 247, 251, 0.82);
}

.floats__row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.floats__price {
  font-size: 16px;
  font-weight: 600;
  letter-spacing: -0.02em;
  font-variant-numeric: tabular-nums;
}

.floats__badge {
  padding: 2px 7px;
  border-radius: 6px;
  background: rgba(255, 107, 53, 0.18);
  color: #ffb08a;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}

@media (max-width: 900px) {
  .floats {
    height: auto;
    perspective: none;
    display: flex;
    gap: 12px;
    overflow-x: auto;
    padding-bottom: 6px;
    scrollbar-width: none;
  }

  .floats::-webkit-scrollbar {
    display: none;
  }

  .floats__card {
    position: static;
    flex: 0 0 232px;
    width: auto;
    transform: none;
    animation: none;
  }
}
</style>
