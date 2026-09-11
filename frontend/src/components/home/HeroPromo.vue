<script setup lang="ts">
/**
 * 首屏促销条。
 *
 * 视觉配方：近黑底 + 三处径向光晕（蓝 / 琥珀 / 洋红）撑起「色彩丰富」，
 * 再加一道 5s 循环的扫光。全部是 CSS 背景与 transform，零 JS、零图片。
 * 点击直接进入购买环节（由父组件打开对应商品的下单抽屉）。
 */
defineProps<{
  title: string
  subtitle: string
  /** 右侧白色胶囊，如「立减 ¥20」 */
  badge: string
  actionText: string
}>()

defineEmits<{ action: [] }>()
</script>

<template>
  <button type="button" class="promo" @click="$emit('action')">
    <span class="promo__mark" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <path d="M13 2L4.5 13.5h6L11 22l8.5-11.5h-6z" />
      </svg>
    </span>

    <span class="promo__copy">
      <strong class="promo__title">{{ title }}</strong>
      <span class="promo__subtitle">{{ subtitle }}</span>
    </span>

    <span class="promo__badge">{{ badge }}</span>

    <span class="promo__action">
      {{ actionText }}
      <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M6 3.5L10.5 8L6 12.5" />
      </svg>
    </span>
  </button>
</template>

<style scoped>
.promo {
  position: relative;
  isolation: isolate;
  overflow: hidden;
  width: min(46rem, 100%);
  min-height: 5.5rem;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 18px;
  color: #fff;
  text-align: left;
  cursor: pointer;
  /* 三处角落光晕 + 近黑线性底：agent.space 让促销条「有颜色」的关键 */
  background:
    radial-gradient(48% 180% at 8% 120%, rgba(0, 153, 255, 0.7) 0%, transparent 72%),
    radial-gradient(42% 180% at 72% -70%, rgba(255, 177, 0, 0.58) 0%, transparent 74%),
    radial-gradient(46% 180% at 96% 125%, rgba(255, 31, 184, 0.54) 0%, transparent 72%),
    linear-gradient(110deg, #0b0d12, #161922 55%, #0b0d12);
  box-shadow: 0 28px 70px -34px rgba(0, 0, 0, 0.92);
  transition: transform 200ms ease-out, border-color 200ms ease-out;
}

.promo:hover {
  transform: translateY(-2px);
  border-color: rgba(255, 255, 255, 0.24);
}

/* 扫光：一道窄高光每 5 秒从左掠过，只动 transform（GPU 合成） */
.promo::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  z-index: 1;
  width: 45%;
  pointer-events: none;
  background: linear-gradient(100deg, transparent, rgba(225, 239, 255, 0.24), transparent);
  animation: promo-sweep 5s ease-in-out infinite;
}

@keyframes promo-sweep {
  0% {
    transform: translate3d(-120%, 0, 0);
  }
  55%,
  100% {
    transform: translate3d(330%, 0, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .promo::after {
    animation: none;
    opacity: 0;
  }
}

.promo__mark {
  flex: none;
  width: 44px;
  height: 44px;
  display: grid;
  place-items: center;
  border-radius: 13px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.1);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.promo__mark svg {
  width: 20px;
  height: 20px;
}

.promo__copy {
  flex: 1 1 160px;
  min-width: 0;
  display: grid;
  gap: 5px;
}

.promo__title {
  font-size: 17px;
  font-weight: 600;
  letter-spacing: -0.01em;
}

.promo__subtitle {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.62);
}

.promo__badge {
  flex: none;
  margin-left: 8px;
  padding: 7px 12px;
  border-radius: 10px;
  background: #fff;
  color: #050505;
  font-size: 13px;
  font-weight: 650;
  white-space: nowrap;
}

.promo__action {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}

.promo__action svg {
  width: 14px;
  height: 14px;
}

@media (max-width: 720px) {
  .promo {
    flex-wrap: wrap;
    gap: 12px;
    padding: 16px 18px;
    border-radius: 16px;
  }

  .promo__mark {
    width: 38px;
    height: 38px;
  }

  .promo__badge {
    margin-left: 0;
  }

  .promo__action {
    margin-left: auto;
  }
}
</style>
