<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useAuthStore } from '@/stores'
import { formatCNY, FULFILLMENT_LABELS, deliveryRequirement, type PublicProduct } from '@/api/publicShop'
import { resolveShopAssetUrl } from '@/api/shop'
import type { GuestPaymentOption } from '@/composables/useGuestCheckout'

const props = defineProps<{
  open: boolean
  product: PublicProduct | null
  /** 可用支付方式，来自后台配置（这里按「抵扣后的应付金额」再过滤一次） */
  paymentMethods?: GuestPaymentOption[]
  submitting?: boolean
  /** 售后 QQ 号，用于「暂无可用支付方式」时的联系入口 */
  supportQQ?: string
  /** 推广计划已开启且当前登录用户有返点余额，用于展示「返点余额抵扣」入口 */
  walletEligible?: boolean
  /** 可用于抵扣的人民币返点余额（分） */
  walletAvailableMinor?: number
}>()

const authStore = useAuthStore()
// 已登录用户订单直接归属本人，无需再填联系方式
const isLoggedIn = computed(() => !!authStore.isAuthenticated)

const emit = defineEmits<{
  close: []
  submit: [payload: { contact: string; paymentType: string; useWallet: boolean }]
}>()

const contact = ref('')
const paymentType = ref('')
const touched = ref(false)
const useWallet = ref(false)

const priceMinor = computed(() => Number(props.product?.price_cny_minor || 0))
const walletAvailable = computed(() => Number(props.walletAvailableMinor || 0))
// 抵扣只在登录态可用：访客没有返点钱包；推广计划关闭时后端也不返回余额入口。
const canUseWallet = computed(() => isLoggedIn.value && !!props.walletEligible && walletAvailable.value > 0)
const walletAppliedMinor = computed(() => {
  if (!useWallet.value || !canUseWallet.value) return 0
  return Math.min(walletAvailable.value, priceMinor.value)
})
const payableMinor = computed(() => Math.max(priceMinor.value - walletAppliedMinor.value, 0))
/** 全额抵扣：没有外部支付环节，也就无需选择支付方式 */
const fullyPaidByWallet = computed(() => priceMinor.value > 0 && payableMinor.value <= 0)

/** 主图 + 图廊；切换商品时回到第一张 */
const images = computed(() =>
  [props.product?.image_url, ...(props.product?.gallery || [])]
    .map((url) => resolveShopAssetUrl(url))
    .filter(Boolean)
)
const activeImage = ref(0)
watch(() => props.product?.id, () => {
  activeImage.value = 0
})

/** 渠道单笔限额以「实际走渠道的金额」为准：只有抵扣后的应付金额才受限额约束 */
function amountFitsMethod(method: GuestPaymentOption, amountMinor: number): boolean {
  const amount = amountMinor / 100
  if (method.single_min > 0 && amount < method.single_min) return false
  if (method.single_max > 0 && amount > method.single_max) return false
  return true
}

const methods = computed(() => {
  if (fullyPaidByWallet.value) return []
  return (props.paymentMethods || []).filter((method) => amountFitsMethod(method, payableMinor.value))
})
const noMethodAvailable = computed(() => !fullyPaidByWallet.value && methods.value.length === 0)

const requirement = computed(() => deliveryRequirement(props.product?.fulfillment_mode || 'manual'))

const contactError = computed(() => {
  if (isLoggedIn.value) return ''
  if (!touched.value) return ''
  const value = contact.value.trim()
  if (!value) return '请填写手机号或邮箱，用于查单与售后'
  if (value.includes('@')) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value) ? '' : '邮箱格式不正确'
  }
  return /^[+\d][\d\s-]{5,20}$/.test(value) ? '' : '手机号格式不正确'
})

const canSubmit = computed(() =>
  (fullyPaidByWallet.value || !noMethodAvailable.value) &&
  (isLoggedIn.value || (contact.value.trim() !== '' && !contactError.value)) &&
  !props.submitting
)

const deliveryNotice = computed(() => {
  if (requirement.value.needsSession) {
    return '支付完成后，需要你粘贴 ChatGPT 登录凭证（Session）用于充值。我们不会索要账号密码。'
  }
  if (requirement.value.needsEmail) {
    return '支付完成后，需要你留一个接收账号的邮箱，管理员人工发货。'
  }
  return '支付完成后，管理员会按订单信息人工处理发货。'
})

// 打开时默认勾选抵扣（防呆：多数用户的诉求就是「把返点花掉」），并重置联系方式与渠道
watch(() => props.open, (open) => {
  if (!open) return
  contact.value = ''
  touched.value = false
  useWallet.value = canUseWallet.value
  paymentType.value = methods.value[0]?.value || ''
})

// 抵扣额度变化会让原选中的渠道不再适用（金额落到限额之外），这里自动纠正，
// 保证「确认下单」时选中的渠道一定可用。
watch([methods, payableMinor], () => {
  if (fullyPaidByWallet.value) {
    paymentType.value = ''
    return
  }
  if (!methods.value.some((item) => item.value === paymentType.value)) {
    paymentType.value = methods.value[0]?.value || ''
  }
})

function submit() {
  touched.value = true
  if (!canSubmit.value) return
  emit('submit', {
    contact: contact.value.trim(),
    // 全额抵扣没有外部支付环节，不传渠道，避免后端把它当作「合并支付」
    paymentType: fullyPaidByWallet.value ? '' : paymentType.value,
    useWallet: fullyPaidByWallet.value ? true : (useWallet.value && canUseWallet.value)
  })
}
</script>

<template>
  <Transition name="sheet">
    <div v-if="open && product" class="sheet-root" @click.self="emit('close')">
      <section class="sheet" role="dialog" aria-modal="true">
        <header class="sheet__head">
          <div>
            <p class="sheet__eyebrow">{{ FULFILLMENT_LABELS[product.fulfillment_mode] }}</p>
            <h2 class="sheet__title">{{ product.name }}</h2>
          </div>
          <button type="button" class="sheet__close" aria-label="关闭" @click="emit('close')">
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M4 4l10 10M14 4L4 14" stroke-linecap="round" />
            </svg>
          </button>
        </header>

        <div v-if="images.length" class="sheet__gallery">
          <img class="sheet__gallery-main" :src="images[activeImage]" :alt="product.name">
          <div v-if="images.length > 1" class="sheet__thumbs">
            <button
              v-for="(url, index) in images"
              :key="`${index}-${url}`"
              type="button"
              class="sheet__thumb"
              :class="{ 'sheet__thumb--active': activeImage === index }"
              :aria-label="`查看第 ${index + 1} 张图`"
              @click="activeImage = index"
            >
              <img :src="url" alt="">
            </button>
          </div>
        </div>

        <p v-if="product.description" class="sheet__desc">{{ product.description }}</p>

        <div class="sheet__summary">
          <div v-if="walletAppliedMinor > 0" class="sheet__summary-breakdown">
            <span class="sheet__summary-row">
              <span>商品金额</span><span>¥{{ formatCNY(priceMinor) }}</span>
            </span>
            <span class="sheet__summary-row sheet__summary-row--credit">
              <span>返点余额抵扣</span><span>-¥{{ formatCNY(walletAppliedMinor) }}</span>
            </span>
          </div>
          <div class="sheet__summary-total">
            <span class="sheet__summary-label">应付金额</span>
            <span class="sheet__summary-price">¥{{ formatCNY(payableMinor) }}</span>
          </div>
        </div>

        <div v-if="isLoggedIn" class="sheet__field">
          <span class="sheet__label">订单归属</span>
          <p class="sheet__notice">订单将归入你当前登录的账号，支付后可在「控制台 · 我的商城订单」里查看进度并提交交付资料。</p>
        </div>

        <div v-else class="sheet__field">
          <label class="sheet__label" for="guest-contact">联系方式</label>
          <input
            id="guest-contact"
            v-model="contact"
            class="sheet__input"
            :class="{ 'sheet__input--error': contactError }"
            type="text"
            inputmode="text"
            autocomplete="email"
            placeholder="手机号或邮箱，用于查单和售后"
            @blur="touched = true"
          >
          <p v-if="contactError" class="sheet__error">{{ contactError }}</p>
          <p v-else class="sheet__hint">只用于核对订单与质保，与充值账号无关</p>
        </div>

        <div v-if="canUseWallet" class="sheet__field">
          <button
            type="button"
            class="sheet__wallet"
            :class="{ 'sheet__wallet--active': useWallet }"
            :aria-pressed="useWallet"
            @click="useWallet = !useWallet"
          >
            <span class="sheet__wallet-check" :class="{ 'sheet__wallet-check--on': useWallet }" aria-hidden="true">
              <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M2.5 6.4l2.4 2.4 4.6-5" stroke-linecap="round" stroke-linejoin="round" />
              </svg>
            </span>
            <span class="sheet__wallet-text">
              <span class="sheet__wallet-title">使用返点余额抵扣</span>
              <span class="sheet__wallet-hint">可用 ¥{{ formatCNY(walletAvailable) }}</span>
            </span>
            <span v-if="useWallet" class="sheet__wallet-amount">-¥{{ formatCNY(walletAppliedMinor) }}</span>
          </button>
          <p v-if="fullyPaidByWallet" class="sheet__wallet-full">返点余额足够全额抵扣，下单后无需再付款。</p>
        </div>

        <div v-if="!fullyPaidByWallet" class="sheet__field">
          <span class="sheet__label">支付方式</span>
          <div v-if="noMethodAvailable" class="sheet__no-method">
            <p>
              当前商品暂无可用的支付方式。<template v-if="supportQQ">可加 QQ 客服 {{ supportQQ }} 协助处理。</template><template v-else>请联系客服处理。</template>
            </p>
          </div>
          <div v-else class="sheet__methods">
            <button
              v-for="method in methods"
              :key="method.value"
              type="button"
              class="sheet__method"
              :class="{ 'sheet__method--active': paymentType === method.value }"
              @click="paymentType = method.value"
            >
              {{ method.label }}
            </button>
          </div>
        </div>

        <div class="sheet__notice">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4">
            <circle cx="8" cy="8" r="6.4" />
            <path d="M8 7.2v4M8 5.1v.7" stroke-linecap="round" />
          </svg>
          <p>{{ deliveryNotice }}</p>
        </div>

        <button type="button" class="sheet__submit" :disabled="!canSubmit" @click="submit">
          {{ submitting ? '正在创建订单…' : (fullyPaidByWallet ? '确认抵扣下单' : `去支付 ¥${formatCNY(payableMinor)}`) }}
        </button>

        <p class="sheet__foot">
          {{ fullyPaidByWallet ? '已用返点余额全额抵扣，下单后可在「控制台 · 我的商城订单」查看进度' : '支付后凭订单号 + 联系方式可在「查订单」页查看进度' }}
        </p>
      </section>
    </div>
  </Transition>
</template>

<style scoped>
/* 令牌继承自 HomeView 的 .sales-home，明暗主题自动切换 */
.sheet-root {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: flex;
  justify-content: flex-end;
  background: var(--sh-scrim, rgba(9, 9, 11, 0.55));
}

.sheet {
  width: min(420px, 100%);
  height: 100%;
  overflow-y: auto;
  padding: 26px 26px 30px;
  background: var(--sh-bg, #fff);
  border-left: 1px solid var(--sh-border, rgba(9, 9, 11, 0.08));
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.sheet__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.sheet__eyebrow {
  font-size: 12px;
  letter-spacing: 0.02em;
  color: var(--sh-text-3, #8a8a93);
}

.sheet__title {
  margin-top: 6px;
  font-size: 19px;
  font-weight: 600;
  letter-spacing: -0.015em;
  color: var(--sh-text, #09090b);
}

.sheet__close {
  width: 30px;
  height: 30px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  border: 1px solid var(--sh-border, rgba(9, 9, 11, 0.08));
  background: transparent;
  color: var(--sh-text-3, #8a8a93);
  cursor: pointer;
  transition: background 160ms ease-out, color 160ms ease-out;
}

.sheet__close:hover {
  background: var(--sh-surface-2, #f6f6f7);
  color: var(--sh-text, #09090b);
}

/* 金额条：中性表面 + hairline，不再用橙色块 */
/* 图廊：主图 + 缩略图，点击切换，不做自动轮播 */
.sheet__gallery {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sheet__gallery-main {
  width: 100%;
  max-height: 240px;
  object-fit: contain;
  border-radius: 11px;
  border: 1px solid var(--sh-border, rgba(9, 9, 11, 0.08));
  background: var(--sh-surface-2, #f6f6f7);
}

.sheet__thumbs {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  scrollbar-width: none;
}

.sheet__thumbs::-webkit-scrollbar {
  display: none;
}

.sheet__thumb {
  flex: 0 0 auto;
  width: 56px;
  height: 56px;
  padding: 0;
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid var(--sh-border, rgba(9, 9, 11, 0.08));
  background: var(--sh-surface-2, #f6f6f7);
  cursor: pointer;
  transition: border-color 160ms ease-out;
}

.sheet__thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.sheet__thumb--active {
  border-color: var(--sh-accent, #d85a28);
}

.sheet__desc {
  margin-top: -4px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--sh-text-2, #56565f);
  white-space: pre-line;
}

.sheet__summary {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 15px 16px;
  border-radius: 11px;
  background: var(--sh-surface-2, #f6f6f7);
  border: 1px solid var(--sh-border, rgba(9, 9, 11, 0.08));
}

/* 抵扣发生时才展开明细：商品金额 → 返点抵扣 → 应付，让「组合支付」一眼看清 */
.sheet__summary-breakdown {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-bottom: 10px;
  border-bottom: 1px dashed var(--sh-border-strong, rgba(9, 9, 11, 0.16));
  font-size: 13px;
  color: var(--sh-text-2, #56565f);
  font-variant-numeric: tabular-nums;
}

.sheet__summary-row {
  display: flex;
  justify-content: space-between;
}

.sheet__summary-row--credit {
  color: var(--sh-accent-text, #b34b1f);
}

.sheet__summary-total {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.sheet__summary-label {
  font-size: 13px;
  color: var(--sh-text-2, #56565f);
}

.sheet__summary-price {
  font-size: 24px;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--sh-text, #09090b);
  font-variant-numeric: tabular-nums;
}

.sheet__field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sheet__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--sh-text, #09090b);
}

.sheet__payment-hint {
  font-size: 12px;
  color: var(--sh-text-3, #8a8a93);
}

.sheet__input {
  height: 44px;
  padding: 0 13px;
  border-radius: 9px;
  border: 1px solid var(--sh-border-strong, rgba(9, 9, 11, 0.16));
  background: var(--sh-surface, #fff);
  color: var(--sh-text, #09090b);
  font-size: 15px;
  font-family: inherit;
  outline: none;
  transition: border-color 160ms ease-out, box-shadow 160ms ease-out;
}

.sheet__input::placeholder {
  color: var(--sh-text-3, #8a8a93);
}

.sheet__input:focus {
  border-color: var(--sh-accent, #d85a28);
  box-shadow: 0 0 0 3px var(--sh-accent-soft, rgba(216, 90, 40, 0.08));
}

.sheet__input--error {
  border-color: rgba(214, 69, 69, 0.7);
}

.sheet__error {
  font-size: 12px;
  color: #d64545;
}

.sheet__hint {
  font-size: 12px;
  color: var(--sh-text-3, #8a8a93);
}

.sheet__methods {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.sheet__method {
  height: 44px;
  border-radius: 9px;
  border: 1px solid var(--sh-border-strong, rgba(9, 9, 11, 0.16));
  background: transparent;
  color: var(--sh-text-2, #56565f);
  font-size: 14px;
  cursor: pointer;
  transition: border-color 160ms ease-out, color 160ms ease-out, background 160ms ease-out;
}

.sheet__method:hover {
  color: var(--sh-text, #09090b);
}

.sheet__method--active {
  border-color: var(--sh-accent, #d85a28);
  color: var(--sh-accent-text, #b34b1f);
  background: var(--sh-accent-soft, rgba(216, 90, 40, 0.08));
  font-weight: 500;
}

/* 安全说明：中性信息块，不用红色，避免吓退用户 */
.sheet__notice {
  display: flex;
  gap: 10px;
  padding: 13px 15px;
  border-radius: 10px;
  background: var(--sh-surface-2, #f6f6f7);
  border: 1px solid var(--sh-border, rgba(9, 9, 11, 0.08));
  color: var(--sh-text-2, #56565f);
  font-size: 13px;
  line-height: 1.7;
}

.sheet__notice svg {
  flex-shrink: 0;
  margin-top: 3px;
  color: var(--sh-text-3, #8a8a93);
}

/* 异常提示：唯一的琥珀色块 */
.sheet__no-method {
  padding: 12px 14px;
  border-radius: 10px;
  border: 1px solid rgba(194, 131, 31, 0.32);
  background: rgba(194, 131, 31, 0.08);
  font-size: 13px;
  line-height: 1.7;
  color: #96650f;
}


.sheet__submit {
  height: 48px;
  border-radius: 9px;
  border: none;
  background: var(--sh-accent, #d85a28);
  color: var(--sh-accent-fg, #fff);
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: background 160ms ease-out;
}

.sheet__submit:hover:not(:disabled) {
  background: var(--sh-accent-hover, #c04f21);
}

.sheet__submit:disabled {
  background: var(--sh-surface-2, #f6f6f7);
  color: var(--sh-text-3, #8a8a93);
  cursor: not-allowed;
}

.sheet__foot {
  font-size: 12px;
  text-align: center;
  color: var(--sh-text-3, #8a8a93);
}

.sheet-enter-active,
.sheet-leave-active {
  transition: opacity 200ms ease-out;
}

.sheet-enter-from,
.sheet-leave-to {
  opacity: 0;
}

.sheet-enter-active .sheet,
.sheet-leave-active .sheet {
  transition: transform 220ms ease-out;
}

.sheet-enter-from .sheet,
.sheet-leave-to .sheet {
  transform: translateX(24px);
}

@media (max-width: 640px) {
  .sheet-root {
    align-items: flex-end;
  }

  .sheet {
    width: 100%;
    height: auto;
    max-height: 92vh;
    border-left: none;
    border-top: 1px solid var(--sh-border, rgba(9, 9, 11, 0.08));
    border-radius: 16px 16px 0 0;
    padding: 22px 20px calc(26px + env(safe-area-inset-bottom));
  }

  .sheet-enter-from .sheet,
  .sheet-leave-to .sheet {
    transform: translateY(32px);
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
.dark .sheet__no-method {
  color: #e0ac52;
}
</style>
