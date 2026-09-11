import type { Directive, DirectiveBinding } from 'vue'

/**
 * 滚动入场：元素进入视口时加一次 .is-revealed，触发 CSS 的 opacity / translateY 过渡。
 *
 * 设计约束（跟首屏动态背景一致）：
 * - 用 IntersectionObserver，不监听 scroll 事件，避免每帧回调；
 * - 只做一次，进场后立刻 unobserve，长页面不会堆积观察者；
 * - 只动 opacity / transform，走 GPU 合成，不触发 layout；
 * - 系统开启「减弱动态效果」时直接置为可见，不做动画；
 * - 浏览器不支持 IntersectionObserver 时直接显示，绝不把内容藏起来。
 */

const REVEALED = 'is-revealed'
const PENDING = 'reveal'

function reduceMotion() {
  return typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

let observer: IntersectionObserver | null = null

function getObserver(): IntersectionObserver | null {
  if (typeof IntersectionObserver === 'undefined') return null
  if (!observer) {
    observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue
          entry.target.classList.add(REVEALED)
          observer?.unobserve(entry.target)
        }
      },
      { rootMargin: '0px 0px -10% 0px', threshold: 0.05 }
    )
  }
  return observer
}

function apply(el: HTMLElement, binding: DirectiveBinding) {
  const delay = typeof binding.value === 'number' ? binding.value : 0

  if (reduceMotion() || typeof IntersectionObserver === 'undefined') {
    el.classList.add(REVEALED)
    return
  }

  el.classList.add(PENDING)
  if (delay > 0) el.style.transitionDelay = `${delay}ms`

  // 已经在视口内的元素（首屏下方一点点的区块）不应该等滚动才出现，
  // requestAnimationFrame 让 IntersectionObserver 有机会先回调一次。
  requestAnimationFrame(() => {
    getObserver()?.observe(el)
  })
}

export const vReveal: Directive<HTMLElement, number | undefined> = {
  mounted(el, binding) {
    apply(el, binding)
  },
  unmounted(el) {
    getObserver()?.unobserve(el)
  }
}

/**
 * 在 <script setup> 里：
 *   const { vReveal } = useRevealOnScroll()
 * 然后模板里用 v-reveal / v-reveal="120"。
 */
export function useRevealOnScroll() {
  return { vReveal }
}
