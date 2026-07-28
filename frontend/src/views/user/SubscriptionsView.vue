<template>
  <AppLayout>
    <div class="space-y-6">
      <section
        class="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900"
      >
        <div class="grid gap-0 xl:grid-cols-[1.25fr_0.75fr]">
          <div class="p-6 sm:p-8">
            <div class="flex flex-wrap items-center gap-2">
              <span class="rounded-full bg-primary-500/10 px-3 py-1 text-xs font-medium text-primary-700 dark:bg-primary-500/15 dark:text-primary-300">
                订阅概览
              </span>
              <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600 dark:bg-dark-800 dark:text-dark-300">
                {{ loading ? '加载中' : `${activeSubscriptions.length} 个有效订阅` }}
              </span>
              <span
                v-if="!loading && nearestExpirationLabel"
                class="rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300"
              >
                {{ nearestExpirationLabel }}
              </span>
            </div>

            <div class="mt-4 flex flex-wrap items-end justify-between gap-4">
              <div class="space-y-2">
                <h1 class="text-2xl font-semibold tracking-tight text-gray-950 dark:text-white sm:text-3xl">
                  {{ t('userSubscriptions.title') }}
                </h1>
                <p class="max-w-2xl text-sm leading-6 text-gray-500 dark:text-dark-400">
                  {{ t('userSubscriptions.description') }}
                </p>
              </div>

              <div class="flex flex-wrap gap-2">
                <button
                  type="button"
                  class="inline-flex items-center justify-center rounded-xl border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50 active:translate-y-px disabled:cursor-wait disabled:opacity-70 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200 dark:hover:border-dark-600 dark:hover:bg-dark-700"
                  :disabled="loading"
                  @click="loadSubscriptions"
                >
                  <Icon name="refresh" size="sm" :class="['mr-2', loading ? 'animate-spin' : '']" />
                  刷新
                </button>
                <button
                  type="button"
                  class="inline-flex items-center justify-center rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-primary-700 active:translate-y-px dark:bg-primary-500 dark:hover:bg-primary-400"
                  @click="router.push({ path: '/purchase', query: { tab: 'subscription' } })"
                >
                  前往订阅 / 续费
                </button>
              </div>
            </div>

            <div class="mt-6 grid gap-3 sm:grid-cols-3">
              <div class="rounded-2xl border border-slate-200 bg-slate-50/80 p-4 dark:border-dark-700 dark:bg-dark-800/60">
                <div class="flex items-center gap-2 text-xs font-medium text-slate-500 dark:text-dark-400">
                  <Icon name="checkCircle" size="xs" class="text-emerald-500" />
                  有效订阅
                </div>
                <div class="mt-2 text-2xl font-semibold text-gray-950 dark:text-white">
                  {{ loading ? '...' : activeSubscriptions.length }}
                </div>
                <p class="mt-1 text-xs text-slate-500 dark:text-dark-400">当前正在生效的套餐数量</p>
              </div>

              <div class="rounded-2xl border border-slate-200 bg-slate-50/80 p-4 dark:border-dark-700 dark:bg-dark-800/60">
                <div class="flex items-center gap-2 text-xs font-medium text-slate-500 dark:text-dark-400">
                  <Icon name="creditCard" size="xs" class="text-primary-500" />
                  额度窗口
                </div>
                <div class="mt-2 text-2xl font-semibold text-gray-950 dark:text-white">
                  {{ loading ? '...' : totalWindowCount }}
                </div>
                <p class="mt-1 text-xs text-slate-500 dark:text-dark-400">每日 / 每周 / 每月额度合计</p>
              </div>

              <div class="rounded-2xl border border-slate-200 bg-slate-50/80 p-4 dark:border-dark-700 dark:bg-dark-800/60">
                <div class="flex items-center gap-2 text-xs font-medium text-slate-500 dark:text-dark-400">
                  <Icon name="clock" size="xs" class="text-amber-500" />
                  最近到期
                </div>
                <div class="mt-2 text-xl font-semibold text-gray-950 dark:text-white">
                  {{ loading ? '加载中' : nearestExpirationValue }}
                </div>
                <p class="mt-1 text-xs text-slate-500 dark:text-dark-400">建议在到期前完成续费</p>
              </div>
            </div>
          </div>

          <div
            class="border-t border-slate-200 bg-slate-50/80 p-6 sm:p-8 xl:border-l xl:border-t-0 dark:border-dark-700 dark:bg-dark-800/60"
          >
            <div class="flex h-full flex-col justify-between gap-4">
              <div>
                <p class="text-sm font-medium text-slate-500 dark:text-dark-400">当前状态</p>
                <div class="mt-3 flex flex-wrap items-center gap-2">
                  <span
                    class="rounded-full px-3 py-1 text-sm font-medium"
                    :class="activeSubscriptions.length > 0 ? 'bg-emerald-500/10 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300' : 'bg-slate-200 text-slate-600 dark:bg-dark-700 dark:text-dark-300'"
                  >
                    {{ loading ? '正在加载' : activeSubscriptions.length > 0 ? '订阅已生效' : '暂无有效订阅' }}
                  </span>
                  <span class="rounded-full bg-slate-200 px-3 py-1 text-sm font-medium text-slate-600 dark:bg-dark-700 dark:text-dark-300">
                    {{ loading ? '...' : totalWindowCount }} 个额度窗口
                  </span>
                </div>
                <p class="mt-4 text-sm leading-6 text-slate-500 dark:text-dark-400">
                  这里展示的是当前账号的订阅额度和重置时间。已开通的套餐会自动显示每日、每周、每月的剩余额度。
                </p>
              </div>

              <div class="rounded-2xl border border-slate-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
                <div class="flex items-center gap-3">
                  <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary-500/10 text-primary-600 dark:bg-primary-500/15 dark:text-primary-300">
                    <Icon name="sparkles" size="sm" />
                  </div>
                  <div>
                    <p class="text-sm font-medium text-gray-950 dark:text-white">订阅使用提示</p>
                    <p class="text-xs leading-5 text-slate-500 dark:text-dark-400">
                      如果额度页面看起来和预期不同，先刷新当前页面查看最新状态。
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div v-if="loading" class="space-y-5">
        <div class="grid gap-5 xl:grid-cols-2">
          <div v-for="idx in 2" :key="idx" class="animate-pulse rounded-3xl border border-slate-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
            <div class="h-5 w-1/2 rounded bg-slate-200 dark:bg-dark-700" />
            <div class="mt-4 h-4 w-2/3 rounded bg-slate-200 dark:bg-dark-700" />
            <div class="mt-6 h-36 rounded-2xl bg-slate-100 dark:bg-dark-800" />
          </div>
        </div>
      </div>

      <div v-else-if="subscriptions.length === 0" class="overflow-hidden rounded-3xl border border-slate-200 bg-white p-12 text-center shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary-500/10 text-primary-600 dark:bg-primary-500/15 dark:text-primary-300">
          <Icon name="creditCard" size="xl" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-950 dark:text-white">
          {{ t('userSubscriptions.noActiveSubscriptions') }}
        </h3>
        <p class="mx-auto max-w-lg text-sm leading-6 text-slate-500 dark:text-dark-400">
          {{ t('userSubscriptions.noActiveSubscriptionsDesc') }}
        </p>
        <div class="mt-6 flex justify-center">
          <button
            type="button"
            class="inline-flex items-center justify-center rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-primary-700 active:translate-y-px dark:bg-primary-500 dark:hover:bg-primary-400"
            @click="router.push({ path: '/purchase', query: { tab: 'subscription' } })"
          >
            前往订阅 / 续费
          </button>
        </div>
      </div>

      <div v-else class="space-y-5">
        <div class="flex items-end justify-between gap-4">
          <div>
            <h2 class="text-lg font-semibold text-gray-950 dark:text-white">订阅明细</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-dark-400">
              按套餐查看每个额度窗口的剩余、已用和重置时间
            </p>
          </div>
          <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600 dark:bg-dark-800 dark:text-dark-300">
            共 {{ subscriptions.length }} 个套餐
          </span>
        </div>

        <div :class="subscriptionGridShellClass">
          <div
            v-for="subscription in subscriptions"
            :key="subscription.id"
            class="overflow-hidden rounded-3xl border bg-white shadow-sm dark:bg-dark-900"
            :class="platformBorderClass(subscription.group?.platform || '')"
          >
            <div :class="['h-1.5 w-full', platformAccentBarClass(subscription.group?.platform || '')]" />

            <div class="space-y-5 p-5 sm:p-6">
              <div class="flex flex-wrap items-start justify-between gap-4">
                <div class="min-w-0 space-y-3">
                  <div class="flex flex-wrap items-center gap-2">
                    <h3 class="truncate text-xl font-semibold tracking-tight text-gray-950 dark:text-white">
                      {{ subscription.group?.name || `Group #${subscription.group_id}` }}
                    </h3>
                    <span
                      :class="[
                        'rounded-full border px-2.5 py-1 text-xs font-medium',
                        platformBadgeClass(subscription.group?.platform || '')
                      ]"
                    >
                      {{ platformLabel(subscription.group?.platform || '') }}
                    </span>
                  </div>

                  <p
                    v-if="subscription.group?.description"
                    class="max-w-3xl text-sm leading-6 text-slate-500 dark:text-dark-400"
                  >
                    {{ subscription.group.description }}
                  </p>

                  <div class="flex flex-wrap gap-2 text-xs text-slate-500 dark:text-dark-400">
                    <span class="rounded-full bg-slate-100 px-3 py-1 dark:bg-dark-800">
                      倍率 ×{{ subscription.group?.rate_multiplier ?? 1 }}
                    </span>
                    <span
                      v-if="subscriptionHasPeakRate(subscription)"
                      class="rounded-full bg-amber-500/10 px-3 py-1 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300"
                    >
                      {{ t('payment.planCard.peakRate') }}: {{ subscriptionPeakRateLabel(subscription) }}
                    </span>
                  </div>
                </div>

                <div class="flex flex-col items-end gap-2">
                  <span
                    :class="[
                      'rounded-full px-3 py-1 text-xs font-medium',
                      subscriptionStatusClass(subscription.status)
                    ]"
                  >
                    {{ subscriptionStatusLabel(subscription.status) }}
                  </span>
                  <button
                    v-if="subscription.status === 'active'"
                    type="button"
                    :class="[
                      'inline-flex items-center justify-center rounded-xl px-4 py-2 text-sm font-semibold text-white transition active:translate-y-px',
                      platformButtonClass(subscription.group?.platform || '')
                    ]"
                    @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(subscription.group_id) } })"
                  >
                    续费
                  </button>
                </div>
              </div>

              <div class="grid gap-4 border-t border-slate-200 pt-5 dark:border-dark-700 sm:grid-cols-2">
                <div class="rounded-2xl border border-slate-200 bg-slate-50/80 p-4 dark:border-dark-700 dark:bg-dark-800/60">
                  <div class="flex items-center justify-between gap-3 text-sm">
                    <span class="text-slate-500 dark:text-dark-400">到期时间</span>
                    <span v-if="subscription.expires_at" :class="getExpirationClass(subscription.expires_at)">
                      {{ formatExpirationDate(subscription.expires_at) }}
                    </span>
                    <span v-else class="text-slate-700 dark:text-dark-200">
                      {{ t('userSubscriptions.noExpiration') }}
                    </span>
                  </div>
                  <div class="mt-4 flex items-center justify-between gap-3">
                    <div>
                      <p class="text-xs font-medium text-slate-500 dark:text-dark-400">
                        当前状态
                      </p>
                      <p class="mt-1 text-sm font-medium text-gray-950 dark:text-white">
                        {{ subscription.status === 'active' ? '可以正常使用' : '订阅不可用' }}
                      </p>
                    </div>
                    <Icon name="clock" size="lg" class="text-primary-500 dark:text-primary-400" />
                  </div>
                </div>

                <div class="rounded-2xl border border-slate-200 bg-slate-50/80 p-4 dark:border-dark-700 dark:bg-dark-800/60">
                  <div class="flex items-center justify-between gap-3 text-sm">
                    <span class="text-slate-500 dark:text-dark-400">额度窗口</span>
                    <span class="rounded-full bg-white px-3 py-1 text-xs font-medium text-slate-500 shadow-sm dark:bg-dark-900 dark:text-dark-300">
                      {{ usageWindows(subscription).length }} 个
                    </span>
                  </div>
                  <div class="mt-4 text-sm text-slate-500 dark:text-dark-400">
                    {{ t('userSubscriptions.usage') }}：
                    {{ subscription.group?.daily_limit_usd || subscription.group?.weekly_limit_usd || subscription.group?.monthly_limit_usd ? '按窗口分开显示' : t('userSubscriptions.unlimitedDesc') }}
                  </div>
                </div>
              </div>

              <div
                v-if="
                  !subscription.group?.daily_limit_usd &&
                  !subscription.group?.weekly_limit_usd &&
                  !subscription.group?.monthly_limit_usd
                "
                class="flex items-center justify-center rounded-2xl border border-emerald-200 bg-emerald-50/80 px-4 py-6 text-center dark:border-emerald-500/20 dark:bg-emerald-500/10"
              >
                <div class="flex items-center gap-3">
                  <span class="text-4xl font-light leading-none text-emerald-600 dark:text-emerald-400">∞</span>
                  <div class="text-left">
                    <p class="text-sm font-semibold text-emerald-700 dark:text-emerald-300">
                      {{ t('userSubscriptions.unlimited') }}
                    </p>
                    <p class="text-xs text-emerald-600/80 dark:text-emerald-400/80">
                      {{ t('userSubscriptions.unlimitedDesc') }}
                    </p>
                  </div>
                </div>
              </div>

              <div v-else class="space-y-4">
                <div
                  v-for="window in usageWindows(subscription)"
                  :key="window.key"
                  class="rounded-2xl border border-slate-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900/50"
                >
                  <div class="flex items-start justify-between gap-4">
                    <div>
                      <p class="text-sm font-medium text-slate-500 dark:text-dark-400">
                        {{ window.title }}
                      </p>
                      <p class="mt-1 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white">
                        {{ formatUsd(window.remaining) }}
                        <span class="text-sm font-medium text-slate-400 dark:text-dark-500">
                          / {{ formatUsd(window.limit) }}
                        </span>
                      </p>
                    </div>
                    <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600 dark:bg-dark-800 dark:text-dark-300">
                      {{ t('subscriptionProgress.remaining') }}
                    </span>
                  </div>

                  <div class="mt-4 h-2 overflow-hidden rounded-full bg-slate-200 dark:bg-dark-700">
                    <div
                      class="h-full rounded-full transition-all duration-300"
                      :class="getProgressBarClass(window.used, window.limit)"
                      :style="{ width: getProgressWidth(window.used, window.limit) }"
                    />
                  </div>

                  <div class="mt-3 flex flex-wrap items-center justify-between gap-2 text-xs text-slate-500 dark:text-dark-400">
                    <span>{{ t('subscriptionProgress.usedAmount', { amount: formatUsd(window.used) }) }}</span>
                    <span class="font-mono">{{ window.footer }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import subscriptionsAPI from '@/api/subscriptions'
import type { UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateOnly } from '@/utils/format'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import { platformAccentBarClass, platformBorderClass, platformBadgeClass, platformButtonClass, platformLabel } from '@/utils/platformColors'
import { getRemainingDurationParts, isOneTimeDailyQuota, type RemainingDurationParts } from '@/utils/subscriptionQuota'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)

type UsageWindowView = {
  key: 'daily' | 'weekly' | 'monthly'
  title: string
  used: number
  limit: number
  remaining: number
  footer: string
}

const activeSubscriptions = computed(() =>
  subscriptions.value.filter((subscription) => subscription.status === 'active')
)

const totalWindowCount = computed(() =>
  subscriptions.value.reduce((total, subscription) => total + usageWindows(subscription).length, 0)
)

const subscriptionGridShellClass = computed(() =>
  subscriptions.value.length === 1
    ? 'mx-auto max-w-4xl'
    : 'grid gap-5 xl:grid-cols-2'
)

const nearestExpirationDates = computed(() =>
  activeSubscriptions.value
    .map((subscription) => subscription.expires_at)
    .filter((value): value is string => Boolean(value))
    .sort((a, b) => new Date(a).getTime() - new Date(b).getTime())
)

const nearestExpirationValue = computed(() => {
  if (nearestExpirationDates.value.length === 0) {
    return '暂无到期'
  }

  return formatExpirationDate(nearestExpirationDates.value[0])
})

const nearestExpirationLabel = computed(() => {
  if (nearestExpirationDates.value.length === 0) return ''
  return `最早到期：${formatDateOnly(nearestExpirationDates.value[0])}`
})

function subscriptionHasPeakRate(subscription: UserSubscription): boolean {
  return hasPeakRate(subscription.group)
}

function subscriptionPeakRateLabel(subscription: UserSubscription): string {
  return formatPeakRateWindow(subscription.group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

async function loadSubscriptions() {
  try {
    loading.value = true
    subscriptions.value = await subscriptionsAPI.getMySubscriptions()
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'bg-gray-400'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function normalizeUsageAmount(value: number | null | undefined): number {
  if (!Number.isFinite(value)) return 0
  return Math.max(value ?? 0, 0)
}

function remainingUsage(used: number | null | undefined, limit: number | null | undefined): number {
  const safeLimit = normalizeUsageAmount(limit)
  const safeUsed = normalizeUsageAmount(used)
  return Math.max(safeLimit - safeUsed, 0)
}

function formatUsd(value: number | null | undefined): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(normalizeUsageAmount(value))
}

function subscriptionStatusLabel(status: UserSubscription['status']): string {
  switch (status) {
    case 'active':
      return t('userSubscriptions.status.active')
    case 'expired':
      return t('userSubscriptions.status.expired')
    case 'revoked':
      return t('userSubscriptions.status.revoked')
    case 'suspended':
      return t('userSubscriptions.status.suspended')
    default:
      return status
  }
}

function subscriptionStatusClass(status: UserSubscription['status']): string {
  switch (status) {
    case 'active':
      return 'bg-emerald-500/10 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'expired':
      return 'bg-slate-100 text-slate-600 dark:bg-dark-700 dark:text-dark-300'
    case 'revoked':
      return 'bg-red-500/10 text-red-700 dark:bg-red-500/15 dark:text-red-300'
    case 'suspended':
      return 'bg-amber-500/10 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
    default:
      return 'bg-slate-100 text-slate-600 dark:bg-dark-700 dark:text-dark-300'
  }
}

function usageWindows(subscription: UserSubscription): UsageWindowView[] {
  const group = subscription.group
  if (!group) return []

  const windows: UsageWindowView[] = []
  if (group.daily_limit_usd) {
    const used = normalizeUsageAmount(subscription.daily_usage_usd)
    const limit = normalizeUsageAmount(group.daily_limit_usd)
    windows.push({
      key: 'daily',
      title: t('subscriptionProgress.dailyRemaining'),
      used,
      limit,
      remaining: remainingUsage(used, limit),
      footer: subscription.daily_window_start ? formatDailyUsageWindow(subscription) : t('userSubscriptions.windowNotActive'),
    })
  }

  if (group.weekly_limit_usd) {
    const used = normalizeUsageAmount(subscription.weekly_usage_usd)
    const limit = normalizeUsageAmount(group.weekly_limit_usd)
    windows.push({
      key: 'weekly',
      title: t('subscriptionProgress.weeklyRemaining'),
      used,
      limit,
      remaining: remainingUsage(used, limit),
      footer: subscription.weekly_window_start
        ? t('userSubscriptions.resetIn', { time: formatResetTime(subscription.weekly_window_start, 168) })
        : t('userSubscriptions.windowNotActive'),
    })
  }

  if (group.monthly_limit_usd) {
    const used = normalizeUsageAmount(subscription.monthly_usage_usd)
    const limit = normalizeUsageAmount(group.monthly_limit_usd)
    windows.push({
      key: 'monthly',
      title: t('subscriptionProgress.monthlyRemaining'),
      used,
      limit,
      remaining: remainingUsage(used, limit),
      footer: subscription.monthly_window_start
        ? t('userSubscriptions.resetIn', { time: formatResetTime(subscription.monthly_window_start, 720) })
        : t('userSubscriptions.windowNotActive'),
    })
  }

  return windows
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days < 0) {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateOnly(expires)

  if (days === 0) {
    return `${dateStr} (${t('common.today')})`
  }
  if (days === 1) {
    return `${dateStr} (${t('common.tomorrow')})`
  }

  return t('userSubscriptions.daysRemaining', { days }) + ` (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days <= 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-700 dark:text-gray-300'
}

function formatDurationParts(parts: RemainingDurationParts): string {
  if (parts.days > 0) {
    return `${parts.days}d ${parts.hours}h`
  }

  if (parts.hours > 0) {
    return `${parts.hours}h ${parts.minutes}m`
  }

  return `${parts.minutes}m`
}

function formatDailyUsageWindow(subscription: UserSubscription): string {
  if (isOneTimeDailyQuota(subscription) && subscription.expires_at) {
    const parts = getRemainingDurationParts(subscription.expires_at)
    if (!parts) return t('userSubscriptions.windowNotActive')
    return t('userSubscriptions.quotaEndsIn', { time: formatDurationParts(parts) })
  }

  return t('userSubscriptions.resetIn', {
    time: formatResetTime(subscription.daily_window_start, 24)
  })
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const parts = getRemainingDurationParts(end)

  return parts ? formatDurationParts(parts) : t('userSubscriptions.windowNotActive')
}

onMounted(() => {
  loadSubscriptions()
})
</script>
