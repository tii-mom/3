<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <!-- Current Balance Card -->
      <div class="card p-6 border-l-4 border-l-primary-500 dark:border-l-primary-500 bg-white dark:bg-dark-800">
        <div class="flex flex-col md:flex-row items-center justify-between gap-6">
          <div class="flex items-center gap-4">
            <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400">
              <Icon name="creditCard" size="lg" />
            </div>
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-gray-500">{{ t('redeem.currentBalance') }}</p>
              <p class="text-3xl font-extrabold text-gray-900 dark:text-white font-mono mt-1">
                ${{ user?.balance?.toFixed(2) || '0.00' }}
              </p>
            </div>
          </div>
          <div class="text-right text-xs text-gray-500 dark:text-dark-400 font-mono">
            {{ t('redeem.concurrency') }}: <span class="font-bold text-gray-900 dark:text-white">{{ user?.concurrency || 0 }}</span> {{ t('redeem.requests') }}
          </div>
        </div>
      </div>

      <!-- Redeem Form -->
      <div class="card">
        <div class="p-6">
          <form @submit.prevent="handleRedeem" class="space-y-5">
            <div>
              <label for="code" class="input-label">
                {{ t('redeem.redeemCodeLabel') }}
              </label>
              <div class="relative mt-1">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
                  <Icon name="gift" size="md" class="text-gray-400 dark:text-dark-500" />
                </div>
                <input
                  id="code"
                  v-model="redeemCode"
                  type="text"
                  required
                  :placeholder="t('redeem.redeemCodePlaceholder')"
                  :disabled="submitting"
                  class="input py-3 pl-12 text-lg"
                />
              </div>
              <p class="input-hint">
                {{ t('redeem.redeemCodeHint') }}
              </p>
            </div>

            <button
              type="submit"
              :disabled="!redeemCode || submitting"
              class="btn btn-primary w-full py-3"
            >
              <svg
                v-if="submitting"
                class="-ml-1 mr-2 h-5 w-5 animate-spin"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <Icon v-else name="checkCircle" size="md" class="mr-2" />
              {{ submitting ? t('redeem.redeeming') : t('redeem.redeemButton') }}
            </button>
          </form>
        </div>
      </div>

      <!-- User Balance Voucher Creation -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('finance.vouchers.createTitle') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ t('finance.vouchers.createSubtitle') }}
              </p>
            </div>
            <button type="button" class="btn btn-secondary btn-sm" @click="loadVoucherData">
              {{ t('common.refresh') }}
            </button>
          </div>
        </div>

        <div class="space-y-5 p-6">
          <div v-if="voucherAvailabilityLoading" class="flex items-center justify-center py-6">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>

          <template v-else>
            <div
              v-if="!voucherAvailability?.enabled || !voucherAvailability?.credit_buckets_enforced"
              class="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900 dark:border-amber-800/60 dark:bg-amber-950/30 dark:text-amber-200"
            >
              {{ t('finance.vouchers.unavailable') }}
            </div>

            <template v-else>
              <div class="grid gap-px overflow-hidden rounded-lg border border-gray-200 bg-gray-200 sm:grid-cols-3 dark:border-dark-700 dark:bg-dark-700">
                <div class="bg-white p-4 dark:bg-dark-900">
                  <p class="text-xs font-medium text-gray-500">{{ t('finance.vouchers.maxAvailable') }}</p>
                  <p class="mt-1 text-xl font-semibold text-gray-950 dark:text-white font-mono">{{ formatUSD(maximumVoucherAmount) }}</p>
                </div>
                <div class="bg-white p-4 dark:bg-dark-900">
                  <p class="text-xs font-medium text-gray-500">{{ t('finance.vouchers.transferable') }}</p>
                  <p class="mt-1 text-xl font-semibold text-gray-950 dark:text-white font-mono">{{ formatUSD(voucherAvailability.transferable_credit) }}</p>
                </div>
                <div class="bg-white p-4 dark:bg-dark-900">
                  <p class="text-xs font-medium text-gray-500">{{ t('finance.vouchers.dailyRemaining') }}</p>
                  <p class="mt-1 text-xl font-semibold text-gray-950 dark:text-white font-mono">{{ voucherAvailability.daily_remaining_count }} / {{ voucherAvailability.daily_count }}</p>
                </div>
              </div>

              <form class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_180px_auto] lg:items-end" @submit.prevent="handleCreateVoucher">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('finance.vouchers.amount') }}
                  <div class="mt-2 flex gap-2">
                    <input
                      v-model="voucherAmount"
                      class="input min-w-0 flex-1 font-mono"
                      inputmode="decimal"
                      :placeholder="formatPlainUSD(voucherAvailability.minimum_usd)"
                    />
                    <button type="button" class="btn btn-secondary whitespace-nowrap" @click="fillMaximumVoucherAmount">
                      {{ t('finance.vouchers.fillMaximum') }}
                    </button>
                  </div>
                  <span class="mt-1 block text-xs text-gray-500 dark:text-dark-400">
                    {{ t('finance.vouchers.amountHelp', { min: formatUSD(voucherAvailability.minimum_usd), max: formatUSD(maximumVoucherAmount) }) }}
                  </span>
                </label>

                <label v-if="requiresVoucherTotp" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('finance.vouchers.validatorCode') }}
                  <input v-model="voucherTotpCode" class="input mt-2 w-full" inputmode="numeric" maxlength="6" placeholder="000000" />
                </label>
                <div v-else class="rounded-lg bg-gray-50 p-3 text-xs text-gray-500 dark:bg-dark-800 dark:text-dark-400">
                  {{ t('finance.vouchers.totpHint', { amount: formatUSD(voucherAvailability.step_up_minimum_usd) }) }}
                </div>

                <button type="submit" class="btn btn-primary h-10" :disabled="!canCreateVoucher">
                  {{ voucherCreating ? t('common.processing') : t('finance.vouchers.generate') }}
                </button>
              </form>

              <div class="rounded-lg bg-gray-50 p-4 text-sm text-gray-600 dark:bg-dark-800 dark:text-dark-300">
                {{ t('finance.vouchers.reservePreview', { fee: formatUSD(estimatedVoucherFee), total: formatUSD(estimatedVoucherTotal), days: voucherAvailability.expiry_days }) }}
              </div>
            </template>

            <section v-if="voucherIssuedCode" class="rounded-lg border border-emerald-300 bg-emerald-50 p-4 dark:border-emerald-800 dark:bg-emerald-950/30">
              <div class="flex items-start justify-between gap-4">
                <div class="min-w-0">
                  <p class="text-sm font-medium text-emerald-800 dark:text-emerald-200">{{ t('finance.vouchers.createdCode') }}</p>
                  <code class="mt-2 block break-all text-base font-semibold text-emerald-950 dark:text-emerald-100 font-mono">{{ voucherIssuedCode }}</code>
                </div>
                <button type="button" class="btn btn-secondary btn-sm" @click="copyVoucherCode">{{ t('common.copy') }}</button>
              </div>
              <p class="mt-3 text-xs text-emerald-700 dark:text-emerald-300">{{ t('finance.vouchers.securityNote') }}</p>
            </section>

            <div class="overflow-x-auto border border-gray-200 dark:border-dark-700">
              <table class="w-full min-w-[760px] text-left text-sm">
                <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800 dark:text-dark-400">
                  <tr>
                    <th class="px-4 py-3">{{ t('finance.vouchers.code') }}</th>
                    <th class="px-4 py-3">{{ t('finance.vouchers.amount') }}</th>
                    <th class="px-4 py-3">{{ t('finance.vouchers.fee') }}</th>
                    <th class="px-4 py-3">{{ t('common.status') }}</th>
                    <th class="px-4 py-3">{{ t('finance.vouchers.expiresAt') }}</th>
                    <th class="px-4 py-3">{{ t('common.actions') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in voucherItems" :key="item.id" class="border-t border-gray-100 dark:border-dark-700">
                    <td class="px-4 py-3 font-mono">•••• {{ item.code_last4 }}</td>
                    <td class="px-4 py-3 font-mono">{{ formatUSD(item.face_value) }}</td>
                    <td class="px-4 py-3 font-mono">{{ formatUSD(item.fee_amount) }}</td>
                    <td class="px-4 py-3">{{ item.status }}</td>
                    <td class="px-4 py-3 font-mono">{{ formatDateTime(item.expires_at) }}</td>
                    <td class="px-4 py-3">
                      <button v-if="item.status === 'ISSUED'" type="button" class="text-sm font-medium text-red-600" @click="cancelUserVoucher(item.id)">
                        {{ t('common.cancel') }}
                      </button>
                    </td>
                  </tr>
                  <tr v-if="voucherItems.length === 0">
                    <td colspan="6" class="px-4 py-8 text-center text-gray-500">{{ t('finance.vouchers.noVouchers') }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </template>
        </div>
      </div>

      <!-- Success Message -->
      <transition name="fade">
        <div
          v-if="redeemResult"
          class="card border-emerald-200 bg-emerald-50 dark:border-emerald-800/50 dark:bg-emerald-900/20"
        >
          <div class="p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-900/30"
              >
                <Icon name="checkCircle" size="md" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-emerald-800 dark:text-emerald-300">
                  {{ t('redeem.redeemSuccess') }}
                </h3>
                <div class="mt-2 text-sm text-emerald-700 dark:text-emerald-400">
                  <p>{{ redeemResult.message }}</p>
                  <div class="mt-3 space-y-1">
                    <p v-if="redeemResult.type === 'balance'" class="font-medium">
                      {{ t('redeem.added') }}: ${{ redeemResult.value.toFixed(2) }}
                    </p>
                    <p v-else-if="redeemResult.type === 'concurrency'" class="font-medium">
                      {{ t('redeem.added') }}: {{ redeemResult.value }}
                      {{ t('redeem.concurrentRequests') }}
                    </p>
                    <p v-else-if="redeemResult.type === 'subscription'" class="font-medium">
                      {{ t('redeem.subscriptionAssigned') }}
                      <span v-if="redeemResult.group_name"> - {{ redeemResult.group_name }}</span>
                      <span v-if="redeemResult.validity_days">
                        ({{
                          t('redeem.subscriptionDays', { days: redeemResult.validity_days })
                        }})</span
                      >
                    </p>
                    <p v-if="redeemResult.new_balance !== undefined">
                      {{ t('redeem.newBalance') }}:
                      <span class="font-semibold">${{ redeemResult.new_balance.toFixed(2) }}</span>
                    </p>
                    <p v-if="redeemResult.new_concurrency !== undefined">
                      {{ t('redeem.newConcurrency') }}:
                      <span class="font-semibold"
                        >{{ redeemResult.new_concurrency }} {{ t('redeem.requests') }}</span
                      >
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Error Message -->
      <transition name="fade">
        <div
          v-if="errorMessage"
          class="card border-red-200 bg-red-50 dark:border-red-800/50 dark:bg-red-900/20"
        >
          <div class="p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-red-100 dark:bg-red-900/30"
              >
                <Icon
                  name="exclamationCircle"
                  size="md"
                  class="text-red-600 dark:text-red-400"
                />
              </div>
              <div class="flex-1">
                <h3 class="text-sm font-semibold text-red-800 dark:text-red-300">
                  {{ t('redeem.redeemFailed') }}
                </h3>
                <p class="mt-2 text-sm text-red-700 dark:text-red-400">
                  {{ errorMessage }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Information Card -->
      <div
        class="card border-primary-200 bg-primary-50 dark:border-primary-800/50 dark:bg-primary-900/20"
      >
        <div class="p-6">
          <div class="flex items-start gap-4">
            <div
              class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30"
            >
              <Icon name="infoCircle" size="md" class="text-primary-600 dark:text-primary-400" />
            </div>
            <div class="flex-1">
              <h3 class="text-sm font-semibold text-primary-800 dark:text-primary-300">
                {{ t('redeem.aboutCodes') }}
              </h3>
              <ul
                class="mt-2 list-inside list-disc space-y-1 text-sm text-primary-700 dark:text-primary-400"
              >
                <li>{{ t('redeem.codeRule1') }}</li>
                <li>{{ t('redeem.codeRule2') }}</li>
                <li>
                  {{ t('redeem.codeRule3') }}
                  <span
                    v-if="contactInfo"
                    class="ml-1.5 inline-flex items-center rounded-md bg-primary-200/50 px-2 py-0.5 text-xs font-medium text-primary-800 dark:bg-primary-800/40 dark:text-primary-200"
                  >
                    {{ contactInfo }}
                  </span>
                </li>
                <li>{{ t('redeem.codeRule4') }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Activity -->
      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('redeem.recentActivity') }}
          </h2>
        </div>
        <div class="p-6">
          <!-- Loading State -->
          <div v-if="loadingHistory" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
          </div>

          <!-- History List -->
          <div v-else-if="history.length > 0" class="space-y-3">
            <div
              v-for="item in history"
              :key="item.id"
              class="flex items-center justify-between rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
            >
              <div class="flex items-center gap-4">
                <div
                  :class="[
                    'flex h-10 w-10 items-center justify-center rounded-xl',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'bg-emerald-100 dark:bg-emerald-900/30'
                        : 'bg-red-100 dark:bg-red-900/30'
                      : isSubscriptionType(item.type)
                        ? 'bg-purple-100 dark:bg-purple-900/30'
                        : item.value >= 0
                          ? 'bg-blue-100 dark:bg-blue-900/30'
                          : 'bg-orange-100 dark:bg-orange-900/30'
                  ]"
                >
                  <!-- 余额类型图标 -->
                  <Icon
                    v-if="isBalanceType(item.type)"
                    name="dollar"
                    size="md"
                    :class="
                      item.value >= 0
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-red-600 dark:text-red-400'
                    "
                  />
                  <!-- 订阅类型图标 -->
                  <Icon
                    v-else-if="isSubscriptionType(item.type)"
                    name="badge"
                    size="md"
                    class="text-purple-600 dark:text-purple-400"
                  />
                  <!-- 并发类型图标 -->
                  <Icon
                    v-else
                    name="bolt"
                    size="md"
                    :class="
                      item.value >= 0
                        ? 'text-blue-600 dark:text-blue-400'
                        : 'text-orange-600 dark:text-orange-400'
                    "
                  />
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ getHistoryItemTitle(item) }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-dark-400">
                    {{ formatDateTime(item.used_at) }}
                  </p>
                </div>
              </div>
              <div class="text-right">
                <p
                  :class="[
                    'text-sm font-semibold',
                    isBalanceType(item.type)
                      ? item.value >= 0
                        ? 'text-emerald-600 dark:text-emerald-400'
                        : 'text-red-600 dark:text-red-400'
                      : isSubscriptionType(item.type)
                        ? 'text-purple-600 dark:text-purple-400'
                        : item.value >= 0
                          ? 'text-blue-600 dark:text-blue-400'
                          : 'text-orange-600 dark:text-orange-400'
                  ]"
                >
                  {{ formatHistoryValue(item) }}
                </p>
                <p
                  v-if="!isAdminAdjustment(item.type)"
                  class="font-mono text-xs text-gray-400 dark:text-dark-500"
                >
                  {{ item.code.slice(0, 8) }}...
                </p>
                <p v-else class="text-xs text-gray-400 dark:text-dark-500">
                  {{ t('redeem.adminAdjustment') }}
                </p>
                <!-- Display notes for admin adjustments -->
                <p
                  v-if="item.notes"
                  class="mt-1 text-xs text-gray-500 dark:text-dark-400 italic max-w-[200px] truncate"
                  :title="item.notes"
                >
                  {{ item.notes }}
                </p>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="empty-state py-8">
            <div
              class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800"
            >
              <Icon name="clock" size="xl" class="text-gray-400 dark:text-dark-500" />
            </div>
            <p class="text-sm text-gray-500 dark:text-dark-400">
              {{ t('redeem.historyWillAppear') }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { redeemAPI, authAPI, type RedeemHistoryItem } from '@/api'
import {
  cancelVoucher,
  createVoucher,
  getVoucherAvailability,
  listVouchers,
  type Voucher,
  type VoucherAvailability,
} from '@/api/financial'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()

const user = computed(() => authStore.user)

const redeemCode = ref('')
const submitting = ref(false)
const redeemResult = ref<{
  message: string
  type: string
  value: number
  new_balance?: number
  new_concurrency?: number
  group_name?: string
  validity_days?: number
} | null>(null)
const errorMessage = ref('')

// History data
const history = ref<RedeemHistoryItem[]>([])
const loadingHistory = ref(false)
const contactInfo = ref('')
const voucherAmount = ref('')
const voucherTotpCode = ref('')
const voucherCreating = ref(false)
const voucherAvailabilityLoading = ref(false)
const voucherAvailability = ref<VoucherAvailability | null>(null)
const voucherItems = ref<Voucher[]>([])
const voucherIssuedCode = ref('')

const parseMoney = (value: string | number | undefined | null): number => {
  const parsed = Number(value ?? 0)
  return Number.isFinite(parsed) ? parsed : 0
}

const formatPlainUSD = (value: string | number | undefined | null): string => {
  const parsed = parseMoney(value)
  return parsed > 0 ? parsed.toFixed(2) : '0.00'
}

const formatUSD = (value: string | number | undefined | null): string => {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD' }).format(parseMoney(value))
}

const maximumVoucherAmount = computed(() => parseMoney(voucherAvailability.value?.maximum_face_value_usd))
const enteredVoucherAmount = computed(() => parseMoney(voucherAmount.value))
const estimatedVoucherFee = computed(() => {
  const feeBps = voucherAvailability.value?.fee_bps ?? 0
  return enteredVoucherAmount.value > 0 ? enteredVoucherAmount.value * feeBps / 10000 : 0
})
const estimatedVoucherTotal = computed(() => enteredVoucherAmount.value + estimatedVoucherFee.value)
const requiresVoucherTotp = computed(() => {
  const threshold = parseMoney(voucherAvailability.value?.step_up_minimum_usd)
  return threshold > 0 && enteredVoucherAmount.value >= threshold
})
const canCreateVoucher = computed(() => {
  if (voucherCreating.value) return false
  if (!voucherAvailability.value?.enabled || !voucherAvailability.value.credit_buckets_enforced) return false
  if (enteredVoucherAmount.value <= 0) return false
  if (enteredVoucherAmount.value < parseMoney(voucherAvailability.value.minimum_usd)) return false
  if (enteredVoucherAmount.value > maximumVoucherAmount.value) return false
  if (requiresVoucherTotp.value && !/^\d{6}$/.test(voucherTotpCode.value.trim())) return false
  return true
})

// Helper functions for history display
const isBalanceType = (type: string) => {
  return type === 'balance' || type === 'admin_balance'
}

const isSubscriptionType = (type: string) => {
  return type === 'subscription'
}

const isAdminAdjustment = (type: string) => {
  return type === 'admin_balance' || type === 'admin_concurrency'
}

const getHistoryItemTitle = (item: RedeemHistoryItem) => {
  if (item.type === 'balance') {
    return t('redeem.balanceAddedRedeem')
  } else if (item.type === 'admin_balance') {
    return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
  } else if (item.type === 'concurrency') {
    return t('redeem.concurrencyAddedRedeem')
  } else if (item.type === 'admin_concurrency') {
    return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
  } else if (item.type === 'subscription') {
    return t('redeem.subscriptionAssigned')
  }
  return t('common.unknown')
}

const formatHistoryValue = (item: RedeemHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  } else if (isSubscriptionType(item.type)) {
    // 订阅类型显示有效天数和分组名称
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}${t('redeem.days')} - ${groupName}` : `${days}${t('redeem.days')}`
  } else {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}${item.value} ${t('redeem.requests')}`
  }
}

const fetchHistory = async () => {
  loadingHistory.value = true
  try {
    history.value = await redeemAPI.getHistory()
  } catch (error) {
    console.error('Failed to fetch history:', error)
  } finally {
    loadingHistory.value = false
  }
}

const loadVoucherData = async () => {
  voucherAvailabilityLoading.value = true
  try {
    const [availability, voucherPage] = await Promise.all([
      getVoucherAvailability(),
      listVouchers().catch(() => ({ items: [] as Voucher[], total: 0, page: 1, page_size: 20, pages: 0 })),
    ])
    voucherAvailability.value = availability
    voucherItems.value = voucherPage.items
  } catch (error) {
    console.error('Failed to load voucher availability:', error)
  } finally {
    voucherAvailabilityLoading.value = false
  }
}

const fillMaximumVoucherAmount = () => {
  voucherAmount.value = formatPlainUSD(maximumVoucherAmount.value)
}

const handleCreateVoucher = async () => {
  if (!canCreateVoucher.value) return
  voucherCreating.value = true
  try {
    const result = await createVoucher(voucherAmount.value.trim(), requiresVoucherTotp.value ? voucherTotpCode.value.trim() : '')
    voucherIssuedCode.value = result.code || ''
    voucherAmount.value = ''
    voucherTotpCode.value = ''
    await Promise.all([loadVoucherData(), authStore.refreshUser()])
    appStore.showSuccess(t('finance.vouchers.createdCode'))
  } catch (error: any) {
    errorMessage.value = error.response?.data?.detail || error.message || t('finance.vouchers.createFailed')
    appStore.showError(errorMessage.value)
  } finally {
    voucherCreating.value = false
  }
}

const cancelUserVoucher = async (id: number) => {
  if (!window.confirm(t('finance.vouchers.cancelConfirm'))) return
  try {
    await cancelVoucher(id)
    await Promise.all([loadVoucherData(), authStore.refreshUser()])
    appStore.showSuccess(t('common.success'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || error.message || t('common.error'))
  }
}

const copyVoucherCode = async () => {
  if (!voucherIssuedCode.value) return
  await navigator.clipboard.writeText(voucherIssuedCode.value)
  appStore.showSuccess(t('common.copied'))
}

const handleRedeem = async () => {
  if (!redeemCode.value.trim()) {
    appStore.showError(t('redeem.pleaseEnterCode'))
    return
  }

  submitting.value = true
  errorMessage.value = ''
  redeemResult.value = null

  try {
    const result = await redeemAPI.redeem(redeemCode.value.trim())

    redeemResult.value = result

    // Refresh user data to get updated balance/concurrency
    await authStore.refreshUser()

    // If subscription type, immediately refresh subscription status
    if (result.type === 'subscription') {
      try {
        await subscriptionStore.fetchActiveSubscriptions(true) // force refresh
      } catch (error) {
        console.error('Failed to refresh subscriptions after redeem:', error)
        appStore.showWarning(t('redeem.subscriptionRefreshFailed'))
      }
    }

    // Clear the input
    redeemCode.value = ''

    // Refresh history
    await fetchHistory()

    // Show success toast
    appStore.showSuccess(t('redeem.codeRedeemSuccess'))
  } catch (error: any) {
    errorMessage.value = error.response?.data?.detail || t('redeem.failedToRedeem')

    appStore.showError(t('redeem.redeemFailed'))
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  fetchHistory()
  loadVoucherData()
  try {
    const settings = await authAPI.getPublicSettings()
    contactInfo.value = settings.contact_info || ''
  } catch (error) {
    console.error('Failed to load contact info:', error)
  }
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
