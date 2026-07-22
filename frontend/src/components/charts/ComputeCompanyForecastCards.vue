<template>
  <div class="grid gap-3 sm:grid-cols-2">
    <article v-for="item in items" :key="item.label" class="border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
      <div class="flex items-center justify-between gap-3">
        <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ item.label }}</p>
        <Icon name="trendingUp" size="sm" class="text-primary-600 dark:text-primary-400" />
      </div>
      <template v-if="item.forecast?.eligible">
        <p class="mt-3 font-mono text-xl font-semibold tabular-nums text-gray-950 dark:text-white">{{ cny(item.forecast.estimated_commission_cny_minor) }}</p>
        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.distribution.commission') }} · {{ t('finance.distribution.growth') }} <span :class="item.forecast.commission_growth_percent >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">{{ signedPercent(item.forecast.commission_growth_percent) }}</span></p>
        <p class="mt-3 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.distribution.recharge') }} {{ cny(item.forecast.estimated_recharge_cny_minor) }}</p>
      </template>
      <p v-else class="mt-4 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ reason(item.forecast?.reason) }}</p>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { DistributionAnalytics, DistributionForecastHorizon } from '@/api/financial'

const props = defineProps<{ forecast?: DistributionAnalytics['forecast'] }>()
const { t } = useI18n()
const items = computed(() => [
  { label: t('finance.distribution.nextWeek'), forecast: props.forecast?.seven_days },
  { label: t('finance.distribution.nextMonth'), forecast: props.forecast?.thirty_days },
])
function cny(minor: number) { return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY' }).format(minor / 100) }
function signedPercent(value: number) { return `${value >= 0 ? '+' : ''}${value.toFixed(1)}%` }
function reason(value?: DistributionForecastHorizon['reason']) {
  if (value === 'insufficient_history') return t('finance.distribution.forecastInsufficientHistory')
  if (value === 'insufficient_activity') return t('finance.distribution.forecastInsufficientActivity')
  return t('finance.distribution.forecastUnavailable')
}
</script>
