<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import {
  publicShopAPI,
  findGuestOrder,
  formatCNY,
  FULFILLMENT_LABELS,
  type GuestOrderLookup
} from '@/api/publicShop'
import DeliveryForm from '@/components/shop/DeliveryForm.vue'

const route = useRoute()
const appStore = useAppStore()

const orderNo = ref('')
const contact = ref('')
const loading = ref(false)
const searched = ref(false)
const order = ref<GuestOrderLookup | null>(null)

/** 安全格式化时间：空值/非法值兜底为「—」，避免渲染 Invalid Date */
function formatDateTime(value?: string | null): string {
  if (!value) return '—'
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleString('zh-CN')
}

const statusLabel = computed(() => {
  const labels: Record<string, string> = {
    pending: '待支付',
    paid: '已支付',
    fulfilled: '已完成',
    cancelled: '已取消',
    refunded: '已退款',
    failed: '失败'
  }
  return labels[order.value?.status || ''] || order.value?.status || '-'
})

const fulfilmentLabel = computed(() => {
  const labels: Record<string, string> = { pending: '待处理', fulfilled: '已交付', failed: '处理失败' }
  return labels[order.value?.fulfillment_status || ''] || '-'
})

const canEditDelivery = computed(() => {
  if (!order.value) return false
  if (order.value.status === 'cancelled' || order.value.status === 'refunded') return false
  return order.value.status !== 'pending'
})

onMounted(() => {
  const no = String(route.query.no || '')
  const ct = String(route.query.contact || '')
  if (no) {
    orderNo.value = no
    if (ct) contact.value = ct
    else {
      const saved = findGuestOrder(no)
      if (saved) contact.value = saved.contact
    }
    void search()
  }
})

async function search() {
  const no = orderNo.value.trim()
  const ct = contact.value.trim()
  if (!no || !ct) {
    appStore.showToast('error', '请填写订单号和联系方式', 3000)
    return
  }
  loading.value = true
  searched.value = true
  try {
    const response = await publicShopAPI.lookupOrder({ order_no: no, contact: ct })
    order.value = response.data || null
  } catch (error: any) {
    order.value = null
    appStore.showToast('error', error?.message || '没有找到这张订单，请核对订单号与联系方式', 3500)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="lookup">
    <header class="lookup__nav">
      <div class="lookup__nav-inner">
        <RouterLink class="lookup__brand" to="/">
          <span class="lookup__mark">3</span>
          <span>3API</span>
        </RouterLink>
        <RouterLink class="lookup__back" to="/">返回首页</RouterLink>
      </div>
    </header>

    <main class="lookup__main">
      <section class="lookup__head">
        <h1>查询订单</h1>
        <p>输入订单号和下单时填写的联系方式，查看进度并补交所需信息</p>
      </section>

      <section class="sh-card lookup__search">
        <div class="lookup__fields">
          <label class="sh-field">
            <span>订单号</span>
            <input v-model="orderNo" type="text" placeholder="例如 C20260909AB12CD" autocomplete="off">
          </label>
          <label class="sh-field">
            <span>联系方式</span>
            <input v-model="contact" type="text" placeholder="下单时填写的手机号或邮箱" autocomplete="off">
          </label>
        </div>
        <button type="button" class="sh-btn-primary" :disabled="loading" @click="search">
          {{ loading ? '查询中…' : '查询订单' }}
        </button>
      </section>

      <section v-if="order" class="sh-card">
        <header class="order__head">
          <div>
            <p class="order__no">{{ order.order_no }}</p>
            <p class="order__name">{{ order.snapshot_name }}</p>
          </div>
          <div class="order__status">
            <span class="sh-badge">{{ statusLabel }}</span>
            <span class="sh-badge sh-badge--muted">{{ FULFILLMENT_LABELS[order.snapshot_fulfillment_mode] }} · {{ fulfilmentLabel }}</span>
          </div>
        </header>

        <dl class="order__meta">
          <div>
            <dt>金额</dt>
            <dd>¥{{ formatCNY(order.snapshot_price_cny_minor) }}</dd>
          </div>
          <div>
            <dt>下单时间</dt>
            <dd>{{ formatDateTime(order.created_at) }}</dd>
          </div>
          <div v-if="order.paid_at">
            <dt>支付时间</dt>
            <dd>{{ formatDateTime(order.paid_at) }}</dd>
          </div>
          <div v-if="order.rental_duration">
            <dt>租期</dt>
            <dd>{{ order.rental_duration }}</dd>
          </div>
        </dl>

        <div v-if="order.fulfillment_note" class="note">
          <p class="note__title">处理备注</p>
          <p class="note__body">{{ order.fulfillment_note }}</p>
        </div>

        <div v-if="order.delivery_submitted" class="note note--ok">
          <p class="note__title">资料已提交</p>
          <p class="note__body">
            {{ order.delivery_hint ? `已收到：${order.delivery_hint}` : '管理员正在处理，请稍候' }}
            如信息填错，可以在下方重新提交。
          </p>
        </div>

        <div v-if="order.status === 'pending'" class="note note--warn">
          <p class="note__title">订单尚未支付</p>
          <p class="note__body">支付完成后才能提交充值所需的信息。</p>
        </div>

        <div v-else-if="canEditDelivery" class="deliver">
          <DeliveryForm
            mode="guest"
            :fulfillment-mode="order.snapshot_fulfillment_mode"
            :order-no="order.order_no"
            :contact="contact.trim()"
            :submitted="order.delivery_submitted"
            :hint="order.delivery_hint || ''"
            @submitted="search()"
          />
        </div>
      </section>

      <section v-else-if="searched && !loading" class="sh-card empty">
        <p>没有查到这张订单</p>
        <p class="empty__hint">请核对订单号与联系方式是否与下单时一致</p>
      </section>
    </main>
  </div>
</template>

<style scoped>
/* 与首页共用同一套明暗令牌 */
.lookup {
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
  --sh-accent-soft: rgba(216, 90, 40, 0.08);
  --sh-accent-text: #b34b1f;

  min-height: 100vh;
  background: var(--sh-bg);
  color: var(--sh-text);
  -webkit-font-smoothing: antialiased;
}


.lookup__nav {
  border-bottom: 1px solid var(--sh-border);
  background: color-mix(in srgb, var(--sh-bg) 82%, transparent);
  backdrop-filter: saturate(180%) blur(14px);
}

.lookup__nav-inner {
  max-width: 760px;
  margin: 0 auto;
  padding: 0 24px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.lookup__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--sh-text);
  text-decoration: none;
  font-size: 15px;
  font-weight: 600;
}

.lookup__mark {
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

.lookup__back {
  font-size: 14px;
  color: var(--sh-text-2);
  text-decoration: none;
  transition: color 160ms ease-out;
}

.lookup__back:hover {
  color: var(--sh-text);
}

.lookup__main {
  max-width: 760px;
  margin: 0 auto;
  padding: 48px 24px 96px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.lookup__head h1 {
  font-size: 28px;
  font-weight: 600;
  letter-spacing: -0.025em;
  color: var(--sh-text);
}

.lookup__head p {
  margin-top: 10px;
  font-size: 14px;
  line-height: 1.7;
  color: var(--sh-text-2);
}

.sh-card {
  padding: 24px;
  border-radius: 14px;
  border: 1px solid var(--sh-border);
  background: var(--sh-surface);
}

.lookup__search {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.lookup__fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.sh-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sh-field span {
  font-size: 13px;
  font-weight: 500;
  color: var(--sh-text);
}

.sh-field input {
  height: 44px;
  padding: 0 13px;
  border-radius: 9px;
  border: 1px solid var(--sh-border-strong);
  background: var(--sh-surface);
  color: var(--sh-text);
  font-size: 14px;
  font-family: inherit;
  outline: none;
  transition: border-color 160ms ease-out, box-shadow 160ms ease-out;
}

.sh-field input::placeholder {
  color: var(--sh-text-3);
}

.sh-field input:focus {
  border-color: var(--sh-accent);
  box-shadow: 0 0 0 3px var(--sh-accent-soft);
}

.sh-btn-primary {
  height: 44px;
  border: none;
  border-radius: 9px;
  background: var(--sh-accent);
  color: #fff;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: background 160ms ease-out;
}

.sh-btn-primary:hover:not(:disabled) {
  background: var(--sh-accent-hover);
}

.sh-btn-primary:disabled {
  background: var(--sh-surface-2);
  color: var(--sh-text-3);
  cursor: not-allowed;
}

.order__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--sh-border);
}

.order__no {
  font-size: 12px;
  color: var(--sh-text-3);
  font-variant-numeric: tabular-nums;
}

.order__name {
  margin-top: 6px;
  font-size: 19px;
  font-weight: 600;
  letter-spacing: -0.015em;
  color: var(--sh-text);
}

.order__status {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}

.sh-badge {
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid var(--sh-accent-soft);
  background: var(--sh-accent-soft);
  color: var(--sh-accent-text);
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.sh-badge--muted {
  border-color: var(--sh-border);
  background: var(--sh-surface-2);
  color: var(--sh-text-2);
}

.order__meta {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin: 20px 0 0;
}

.order__meta dt {
  font-size: 12px;
  color: var(--sh-text-3);
}

.order__meta dd {
  margin: 6px 0 0;
  font-size: 14px;
  color: var(--sh-text);
  font-variant-numeric: tabular-nums;
}

.note {
  margin-top: 20px;
  padding: 14px 16px;
  border-radius: 10px;
  border: 1px solid var(--sh-border);
  background: var(--sh-surface-2);
}

.note--ok {
  border-color: rgba(26, 158, 111, 0.32);
  background: rgba(26, 158, 111, 0.08);
}

.note--warn {
  border-color: rgba(194, 131, 31, 0.34);
  background: rgba(194, 131, 31, 0.08);
}

.note__title {
  font-size: 13px;
  font-weight: 500;
  color: var(--sh-text);
}

.note__body {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.75;
  color: var(--sh-text-2);
  white-space: pre-wrap;
}

.deliver {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--sh-border);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.empty {
  text-align: center;
  color: var(--sh-text-2);
}

.empty__hint {
  margin-top: 8px;
  font-size: 13px;
  color: var(--sh-text-3);
}

@media (max-width: 640px) {
  .lookup__main {
    padding: 32px 20px 72px;
  }

  .lookup__nav-inner {
    padding: 0 20px;
  }

  .lookup__fields,
  .order__meta {
    grid-template-columns: 1fr;
  }

  .order__head {
    flex-direction: column;
  }

  .order__status {
    align-items: flex-start;
  }
}
</style>

<style>
/* 深色覆盖放在**非 scoped** 块中：Vue 的 scoped-CSS 编译器会把
   `:global(.dark) X` 编译成只有 `.dark`，X 被丢弃，导致生产构建里
   深色规则整体失效。与 SettingsView 的处理方式保持一致。 */
.dark .lookup {
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
  --sh-accent-soft: rgba(224, 99, 47, 0.13);
  --sh-accent-text: #f0916a;
}
</style>
