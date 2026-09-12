<script setup lang="ts">
/**
 * 商品品牌图标方块。
 * 优先用后台 image_url（真实商品图），否则按品类渲染品牌渐变 + 内联 SVG 字形。
 * 用于 PlanCard 与 HeroFloatCards，统一替换原先突兀的文字 chip / 「3」回退。
 */
import { computed } from 'vue'
import type { PublicProduct } from '@/api/publicShop'
import { resolveShopAssetUrl } from '@/api/shop'
import { resolveBrandDisplay } from '@/utils/productBrand'

const props = withDefaults(
  defineProps<{
    product: PublicProduct
    size?: number
  }>(),
  { size: 44 }
)

// 每个实例一个稳定唯一 id，给 SVG 渐变 def 用（避免同文档内多实例 id 冲突）。
const gradId = `brand-grad-${Math.random().toString(36).slice(2, 9)}`

const decision = computed(() => resolveBrandDisplay(props.product))
const meta = computed(() => decision.value.meta)
const productImgUrl = computed(() => resolveShopAssetUrl(props.product.image_url))
const dim = computed(() => `${props.size}px`)

// 多彩字形（如 Gemini 四角星）用渐变填充，否则用 currentColor / 显式 glyphColor 覆盖。
const glyphFill = computed(() => {
  if (meta.value.glyph.mode !== 'fill') return 'none'
  if (meta.value.glyph.gradient) return `url(#${gradId})`
  return meta.value.glyphColor || 'currentColor'
})
const glyphStroke = computed(() => {
  if (meta.value.glyph.mode !== 'stroke') return 'none'
  if (meta.value.glyph.gradient) return `url(#${gradId})`
  return meta.value.glyphColor || 'currentColor'
})
const strokeWidth = computed(() => meta.value.glyph.strokeWidth || 1.6)
</script>

<template>
  <span
    class="brand-logo"
    :class="{ 'brand-logo--light': meta.tileLight }"
    :style="{ width: dim, height: dim, background: meta.gradient, color: meta.glyphColor || undefined }"
  >
    <img v-if="decision.useProductImage" :src="productImgUrl" :alt="product.name" class="brand-logo__img" />
    <img
      v-else-if="decision.useBrandImage"
      :src="meta.image"
      :alt="meta.label"
      class="brand-logo__img brand-logo__img--brand"
    />
    <svg
      v-else
      viewBox="0 0 24 24"
      :fill="glyphFill"
      :stroke="glyphStroke"
      :stroke-width="strokeWidth"
      stroke-linecap="round"
      stroke-linejoin="round"
      class="brand-logo__glyph"
      aria-hidden="true"
    >
      <defs v-if="meta.glyph.gradient">
        <linearGradient
          :id="gradId"
          :x1="meta.glyph.gradient.x1"
          :y1="meta.glyph.gradient.y1"
          :x2="meta.glyph.gradient.x2"
          :y2="meta.glyph.gradient.y2"
        >
          <stop
            v-for="(s, i) in meta.glyph.gradient.stops"
            :key="i"
            :offset="s.offset"
            :stop-color="s.color"
          />
        </linearGradient>
      </defs>
      <path v-for="(d, i) in meta.glyph.paths" :key="i" :d="d" />
    </svg>
  </span>
</template>

<style scoped>
.brand-logo {
  display: inline-grid;
  place-items: center;
  flex: none;
  border-radius: 12px;
  color: #fff;
  overflow: hidden;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.25),
    0 6px 16px -8px rgba(0, 0, 0, 0.5);
}

/* 浅底方块（如 Gemini 浅色 tile）：补一层内描边 + 浅阴影，才能在白色卡片上看出边界。 */
.brand-logo--light {
  box-shadow:
    inset 0 0 0 1px rgba(15, 23, 42, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.7),
    0 6px 16px -8px rgba(15, 23, 42, 0.28);
}

.brand-logo__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 品牌线稿图（OpenAI 结）：图形本身的留白已烘焙进 PNG（内容约占画布 62%），
   这里不再叠加 padding，否则 46px 方块里的图标会被压得过小。 */
.brand-logo__img--brand {
  box-sizing: border-box;
  object-fit: contain;
  padding: 0;
}

.brand-logo__glyph {
  width: 62%;
  height: 62%;
}
</style>
