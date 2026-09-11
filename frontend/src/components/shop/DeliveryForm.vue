<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAppStore } from '@/stores/app'
import { shopAPI } from '@/api/shop'
import {
  publicShopAPI,
  extractSessionToken,
  deliveryRequirement,
  type FulfillmentMode
} from '@/api/publicShop'
import SessionGuide from '@/components/home/SessionGuide.vue'

/**
 * 商城交付资料提交表单，官网首页、查单页与控制台商城共用。
 * 按商品交付模式自动切换需要提交的内容：
 *   session_topup     → 粘贴 ChatGPT 登录凭证
 *   account_delivery  → 填写接收账号的邮箱
 *   rental            → 邮箱 + 租期
 *   manual            → 自由备注
 */

const props = withDefaults(defineProps<{
  /** user：登录态订单，凭订单 ID 提交；guest：免登录订单，凭订单号 + 联系方式提交 */
  mode: 'user' | 'guest'
  fulfillmentMode: FulfillmentMode
  orderId?: number
  orderNo?: string
  contact?: string
  submitted?: boolean
  hint?: string
  theme?: 'dark' | 'light'
}>(), { theme: 'light', submitted: false })

const emit = defineEmits<{ submitted: [] }>()

const appStore = useAppStore()
const payload = ref('')
const email = ref('')
const rentalDuration = ref('')
const submitting = ref(false)
const guideOpen = ref(false)

const requirement = computed(() => deliveryRequirement(props.fulfillmentMode))
const detectedToken = computed(() => (requirement.value.needsSession ? extractSessionToken(payload.value) : ''))

const canSubmit = computed(() => {
  if (props.submitted && !dirty.value) return false
  if (requirement.value.needsSession) return detectedToken.value.length >= 20
  if (requirement.value.needsEmail) return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim())
  return payload.value.trim().length > 0
})

// 已提交过资料时，只有用户重新填写内容才允许再次提交
const dirty = computed(() =>
  requirement.value.needsSession ? payload.value.trim().length > 0
    : requirement.value.needsEmail ? email.value.trim().length > 0
      : payload.value.trim().length > 0
)

const title = computed(() => {
  if (requirement.value.needsSession) return '提交登录凭证'
  if (requirement.value.needsEmail) return '提交接收邮箱'
  return '补充信息'
})

async function submit() {
  if (!canSubmit.value || submitting.value) return
  submitting.value = true
  try {
    const body = requirement.value.needsSession
      ? detectedToken.value
      : requirement.value.needsEmail
        ? email.value.trim()
        : payload.value.trim()
    if (props.mode === 'user') {
      if (!props.orderId) return
      await shopAPI.submitDelivery(props.orderId, {
        payload: body,
        rental_duration: rentalDuration.value.trim() || undefined
      })
    } else {
      if (!props.orderNo || !props.contact) return
      await publicShopAPI.submitDelivery({
        order_no: props.orderNo,
        contact: props.contact,
        payload: body,
        rental_duration: rentalDuration.value.trim() || undefined
      })
    }
    appStore.showToast('success', '已提交，正在为你处理', 3000)
    payload.value = ''
    email.value = ''
    emit('submitted')
  } catch (error: any) {
    appStore.showToast('error', error?.message || '提交失败，请稍后重试', 3500)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="delivery" :class="{ 'delivery--dark': theme === 'dark' }">
    <div v-if="submitted" class="delivery__done">
      <p class="delivery__done-title">资料已提交</p>
      <p class="delivery__done-body">
        {{ hint ? `已收到：${hint}` : '管理员正在处理，请稍候。' }}
        如信息填错，可以在下方重新提交。
      </p>
    </div>

    <h4 class="delivery__title">{{ title }}</h4>

    <template v-if="requirement.needsSession">
      <button type="button" class="delivery__toggle" @click="guideOpen = !guideOpen">
        {{ guideOpen ? '收起获取教程' : '怎么拿到登录凭证？' }}
      </button>
      <div v-if="guideOpen" class="delivery__guide">
        <SessionGuide :theme="theme === 'dark' ? 'dark' : 'auto'" />
      </div>

      <label class="delivery__field">
        <span>粘贴 Session</span>
        <textarea
          v-model="payload"
          rows="4"
          placeholder="粘贴 https://chatgpt.com/api/auth/session 打开后的整段内容，或直接粘贴 accessToken"
        />
      </label>
      <p v-if="detectedToken" class="delivery__hint delivery__hint--ok">
        已识别到凭证：{{ detectedToken.slice(0, 10) }}…（长度 {{ detectedToken.length }}）
      </p>
      <p v-else-if="payload.trim()" class="delivery__hint delivery__hint--warn">
        没能识别出凭证，请确认是否完整复制了页面内容
      </p>
      <p v-else class="delivery__hint">只读取登录凭证用于本次充值，不会索要密码</p>
    </template>

    <template v-else-if="requirement.needsEmail">
      <label class="delivery__field">
        <span>接收账号的邮箱</span>
        <input v-model="email" type="email" placeholder="用于接收开通好的账号">
      </label>
      <label v-if="requirement.needsDuration" class="delivery__field">
        <span>租期</span>
        <select v-model="rentalDuration">
          <option value="">请选择</option>
          <option value="7 天">7 天</option>
          <option value="1 个月">1 个月</option>
          <option value="3 个月">3 个月</option>
          <option value="12 个月">12 个月</option>
        </select>
      </label>
    </template>

    <template v-else>
      <label class="delivery__field">
        <span>补充信息</span>
        <textarea v-model="payload" rows="3" placeholder="备注或收货信息" />
      </label>
    </template>

    <button type="button" class="delivery__submit" :disabled="!canSubmit || submitting" @click="submit">
      {{ submitting ? '提交中…' : submitted ? '重新提交' : '提交' }}
    </button>
  </div>
</template>

<style scoped>
.delivery {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.delivery__done {
  padding: 12px 14px;
  border-radius: 10px;
  border: 1px solid rgba(29, 158, 117, 0.35);
  background: rgba(29, 158, 117, 0.08);
}

.delivery__done-title {
  font-size: 13px;
  font-weight: 500;
  color: #065f46;
}

.delivery__done-body {
  margin-top: 4px;
  font-size: 13px;
  line-height: 1.7;
  color: #047857;
}

.delivery__title {
  font-size: 15px;
  font-weight: 500;
  color: #111827;
}

.delivery__toggle {
  align-self: flex-start;
  padding: 6px 0;
  border: none;
  background: transparent;
  color: #c2410c;
  font-size: 13px;
  cursor: pointer;
}

.delivery__guide {
  padding: 12px 14px;
  border-radius: 10px;
  border: 1px solid rgba(17, 24, 39, 0.1);
  background: rgba(17, 24, 39, 0.02);
}

.delivery__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.delivery__field span {
  font-size: 13px;
  color: #4b5563;
}

.delivery__field input,
.delivery__field textarea,
.delivery__field select {
  padding: 10px 12px;
  border-radius: 9px;
  border: 1px solid rgba(17, 24, 39, 0.14);
  background: #fff;
  color: #111827;
  font-size: 14px;
  font-family: inherit;
  outline: none;
  resize: vertical;
}

.delivery__field input:focus,
.delivery__field textarea:focus,
.delivery__field select:focus {
  border-color: rgba(216, 90, 40, 0.6);
}

.delivery__hint {
  font-size: 12px;
  color: #6b7280;
}

.delivery__hint--ok {
  color: #047857;
}

.delivery__hint--warn {
  color: #b45309;
}

.delivery__submit {
  height: 40px;
  border: none;
  border-radius: 9px;
  background: #d85a28;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
}

.delivery__submit:disabled {
  background: rgba(17, 24, 39, 0.1);
  color: rgba(17, 24, 39, 0.4);
  cursor: not-allowed;
}

.delivery--dark .delivery__title {
  color: rgba(255, 255, 255, 0.9);
}

.delivery--dark .delivery__toggle {
  color: #e88a5c;
}

.delivery--dark .delivery__guide {
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.02);
}

.delivery--dark .delivery__field span {
  color: rgba(255, 255, 255, 0.62);
}

.delivery--dark .delivery__field input,
.delivery--dark .delivery__field textarea,
.delivery--dark .delivery__field select {
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.04);
  color: #fff;
}

.delivery--dark .delivery__hint {
  color: rgba(255, 255, 255, 0.45);
}

.delivery--dark .delivery__hint--ok {
  color: #5dcaa5;
}

.delivery--dark .delivery__hint--warn {
  color: #e8a33c;
}

.delivery--dark .delivery__submit:disabled {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.35);
}
</style>

<style>
/* 深色覆盖放在**非 scoped** 块中：Vue 的 scoped-CSS 编译器会把
   `:global(.dark) X` 编译成只有 `.dark`，X 被丢弃，导致生产构建里
   深色规则整体失效。与 SettingsView 的处理方式保持一致。 */
.dark .delivery__title {
  color: rgba(255, 255, 255, 0.9);
}

.dark .delivery__field span {
  color: rgba(255, 255, 255, 0.62);
}

.dark .delivery__field input,
.dark .delivery__field textarea,
.dark .delivery__field select {
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.04);
  color: #fff;
}

.dark .delivery__done {
  border-color: rgba(29, 158, 117, 0.4);
  background: rgba(29, 158, 117, 0.12);
}

.dark .delivery__done-title {
  color: #6ee7b7;
}

.dark .delivery__done-body {
  color: #a7f3d0;
}

.dark .delivery__guide {
  border-color: rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.02);
}

.dark .delivery__submit:disabled {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.35);
}
</style>
