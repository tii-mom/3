<template>
  <div class="h-[250px] w-full sm:h-[290px]">
    <div v-if="loading" class="flex h-full items-center justify-center">
      <LoadingSpinner />
    </div>
    <Line v-else-if="series.length" :data="chartData" :options="chartOptions" />
    <div v-else class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-dark-400">
      {{ t('finance.distribution.analyticsNoData') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { DistributionAnalyticsPoint } from '@/api/financial'
import { useTheme } from '@/composables/useTheme'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const props = withDefaults(defineProps<{ series: DistributionAnalyticsPoint[]; loading?: boolean }>(), { loading: false })
const { t } = useI18n()
const { isDark } = useTheme()
const colors = computed(() => ({
  text: isDark.value ? '#c6cbd1' : '#68717c',
  grid: isDark.value ? '#2c3035' : '#e4e6e1',
  recharge: isDark.value ? '#e6a07b' : '#c45126',
  earnings: isDark.value ? '#83c4a1' : '#238653',
}))
const formatCNY = (minor: number) => new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY', maximumFractionDigits: 0 }).format(minor / 100)

const chartData = computed(() => ({
  labels: props.series.map(point => point.date.slice(5)),
  datasets: [
    {
      label: t('finance.distribution.recharge'),
      data: props.series.map(point => point.recharge_cny_minor / 100),
      borderColor: colors.value.recharge,
      backgroundColor: `${colors.value.recharge}18`,
      fill: true,
      tension: 0.32,
      pointRadius: 2,
      pointHoverRadius: 5,
      borderWidth: 2,
    },
    {
      label: t('finance.distribution.commission'),
      data: props.series.map(point => point.commission_cny_minor / 100),
      borderColor: colors.value.earnings,
      backgroundColor: `${colors.value.earnings}12`,
      fill: true,
      tension: 0.32,
      pointRadius: 2,
      pointHoverRadius: 5,
      borderWidth: 2,
    },
  ],
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { intersect: false, mode: 'index' as const },
  plugins: {
    legend: {
      position: 'top' as const,
      align: 'start' as const,
      labels: { color: colors.value.text, usePointStyle: true, pointStyle: 'circle', boxWidth: 7, padding: 18, font: { size: 11 } },
    },
    tooltip: {
      callbacks: {
        label: (context: { dataset: { label?: string }; raw: unknown }) => `${context.dataset.label || ''}: ${formatCNY(Number(context.raw) * 100)}`,
      },
    },
  },
  scales: {
    x: { grid: { color: colors.value.grid, drawTicks: false }, ticks: { color: colors.value.text, maxTicksLimit: 8, font: { size: 10 } } },
    y: { beginAtZero: true, grid: { color: colors.value.grid }, ticks: { color: colors.value.text, font: { size: 10 }, callback: (value: string | number) => `¥${Number(value).toLocaleString('zh-CN')}` } },
  },
}))
</script>
