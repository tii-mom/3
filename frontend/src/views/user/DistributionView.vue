<template>
  <AppLayout>
    <div class="business-page-redesign mx-auto max-w-7xl space-y-6">
      <section class="flex flex-col gap-3 border-b border-gray-200 pb-5 sm:flex-row sm:items-end sm:justify-between dark:border-dark-700">
        <div><h1 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('finance.distribution.title') }}</h1><p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('finance.distribution.subtitle') }}</p></div>
        <div class="inline-flex h-9 self-start border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-700 dark:bg-dark-800">
          <button v-for="tab in tabs" :key="tab.id" class="px-3 text-sm" :class="activeTab === tab.id ? 'bg-white font-medium text-gray-950 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500'" @click="activeTab = tab.id">{{ tab.label }}</button>
        </div>
      </section>

      <template v-if="dashboard">
        <section class="grid gap-px overflow-hidden border border-gray-200 bg-gray-200 sm:grid-cols-2 lg:grid-cols-4 dark:border-dark-700 dark:bg-dark-700">
          <div v-for="stat in stats" :key="stat.label" class="bg-white p-5 dark:bg-dark-900">
            <p class="text-xs font-medium uppercase text-gray-500">{{ stat.label }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-950 dark:text-white font-mono">{{ stat.value }}</p>
          </div>
        </section>

        <section v-if="activeTab === 'overview'" class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
          <div>
            <h2 class="mb-3 text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.tiers') }}</h2>
            <div class="overflow-x-auto border border-gray-200 dark:border-dark-700">
              <table class="w-full min-w-[620px] text-sm">
                <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800">
                  <tr>
                    <th class="px-4 py-3 text-left">{{ t('finance.distribution.tier') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.distribution.threshold') }}</th>
                    <th v-for="level in 5" :key="level" class="px-4 py-3 text-right">L{{ level }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="tier in dashboard.tiers" :key="tier.tier" class="border-t border-gray-100 dark:border-dark-700">
                    <td class="px-4 py-3 font-mono">T{{ tier.tier }}</td>
                    <td class="px-4 py-3 text-right font-mono">{{ cny(tier.threshold_cny_minor) }}</td>
                    <td v-for="rate in tier.rates_bps" :key="rate" class="px-4 py-3 text-right font-mono">{{ rate / 100 }}%</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="border border-gray-200 p-5 dark:border-dark-700">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.levels') }}</h2>
            <div class="mt-4 space-y-3">
              <div v-for="level in 5" :key="level" class="flex items-center justify-between border-b border-gray-100 pb-3 last:border-0 dark:border-dark-700">
                <span class="text-sm text-gray-500">L{{ level }}</span>
                <strong class="text-gray-900 dark:text-white font-mono">{{ dashboard.level_counts[level] || 0 }}</strong>
              </div>
            </div>
          </div>
        </section>

        <section v-if="activeTab === 'team'" class="space-y-4">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_340px]">
            <div ref="graphContainer" class="h-[420px] min-w-0 border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900"></div>
            <div class="border border-gray-200 p-4 dark:border-dark-700">
              <input v-model="search" class="input w-full" :placeholder="t('common.search')" @keyup.enter="() => loadTeam()" />
              <div class="mt-4 space-y-2">
                <button v-for="node in team" :key="node.user_id" class="flex w-full items-center justify-between border-b border-gray-100 px-2 py-3 text-left dark:border-dark-700" @click="expandNode(node)">
                  <span>
                    <b class="block text-sm text-gray-900 dark:text-white">{{ node.username || node.email_masked }}</b>
                    <small class="text-gray-500 font-mono">{{ node.email_masked }}</small>
                  </span>
                  <span class="text-xs text-gray-500 font-mono">{{ node.direct_children }}</span>
                </button>
                <p v-if="team.length === 0" class="py-8 text-center text-sm text-gray-500">{{ t('common.noData') }}</p>
              </div>
            </div>
          </div>
        </section>

        <section v-if="activeTab === 'ledger'" class="overflow-x-auto border border-gray-200 dark:border-dark-700">
          <table class="w-full min-w-[760px] text-sm">
            <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left">{{ t('finance.distribution.order') }}</th>
                <th class="px-4 py-3 text-right">{{ t('finance.distribution.depth') }}</th>
                <th class="px-4 py-3 text-right">{{ t('finance.distribution.rate') }}</th>
                <th class="px-4 py-3 text-right">{{ t('finance.distribution.commission') }}</th>
                <th class="px-4 py-3 text-left">{{ t('common.status') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in ledger" :key="item.id" class="border-t border-gray-100 dark:border-dark-700">
                <td class="px-4 py-3 font-mono">#{{ item.source_order_id }}</td>
                <td class="px-4 py-3 text-right font-mono">L{{ item.depth }}</td>
                <td class="px-4 py-3 text-right font-mono">{{ item.rate_bps / 100 }}%</td>
                <td class="px-4 py-3 text-right font-mono font-medium text-emerald-600">{{ cny(item.amount_cny_minor) }}</td>
                <td class="px-4 py-3">{{ item.status }}</td>
              </tr>
            </tbody>
          </table>
        </section>

        <section v-if="activeTab === 'withdraw'" class="space-y-6">
          <div class="grid gap-6 lg:grid-cols-2">
            <form class="border border-gray-200 p-5 dark:border-dark-700" @submit.prevent="saveAccount">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.payout') }}</h2>
              <div v-if="payout" class="mt-4 text-sm text-gray-600 dark:text-gray-300 font-mono">{{ payout.real_name_mask }} · {{ payout.account_mask }}</div>
              <div class="mt-4 grid gap-3 sm:grid-cols-2">
                <input v-model="realName" class="input" :placeholder="t('finance.distribution.realName')" />
                <input v-model="alipay" class="input" :placeholder="t('finance.distribution.alipay')" />
              </div>
              <button class="btn btn-primary mt-4">{{ t('common.save') }}</button>
            </form>
            <form class="border border-gray-200 p-5 dark:border-dark-700" @submit.prevent="withdraw">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.withdraw') }}</h2>
              <p class="mt-3 text-sm text-gray-500 font-mono">{{ t('common.available') }}: {{ cny(dashboard.available_cny_minor) }}</p>
              <input v-model="withdrawAmount" class="input mt-4 w-full font-mono" inputmode="decimal" placeholder="100.00" />
              <button class="btn btn-primary mt-4">{{ t('finance.distribution.submitWithdrawal') }}</button>
            </form>
          </div>
          <div class="overflow-x-auto border border-gray-200 dark:border-dark-700">
            <table class="w-full min-w-[760px] text-sm">
              <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800">
                <tr>
                  <th class="px-4 py-3 text-left">ID</th>
                  <th class="px-4 py-3 text-right">{{ t('finance.distribution.withdraw') }}</th>
                  <th class="px-4 py-3 text-right">{{ t('finance.vouchers.fee') }}</th>
                  <th class="px-4 py-3 text-right">{{ t('finance.distribution.net') }}</th>
                  <th class="px-4 py-3 text-left">{{ t('common.status') }}</th>
                  <th class="px-4 py-3 text-left">{{ t('finance.admin.reference') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in withdrawals" :key="item.id" class="border-t border-gray-100 dark:border-dark-700">
                  <td class="px-4 py-3 font-mono">#{{ item.id }}</td>
                  <td class="px-4 py-3 text-right font-mono">{{ cny(item.amount_cny_minor) }}</td>
                  <td class="px-4 py-3 text-right font-mono">{{ cny(item.fee_cny_minor) }}</td>
                  <td class="px-4 py-3 text-right font-mono font-medium text-gray-900 dark:text-white">{{ cny(item.amount_cny_minor - item.fee_cny_minor) }}</td>
                  <td class="px-4 py-3">
                    {{ item.status }}
                    <p v-if="item.reject_reason" class="mt-1 text-xs text-rose-600">{{ item.reject_reason }}</p>
                  </td>
                  <td class="px-4 py-3 font-mono">{{ item.payment_reference || '-' }}</td>
                </tr>
                <tr v-if="withdrawals.length === 0">
                  <td colspan="6" class="px-4 py-8 text-center text-gray-500">{{ t('common.noData') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { createWithdrawal, getDistributionDashboard, getDistributionLedger, getDistributionTree, getPayoutAccount, listWithdrawals, savePayoutAccount, type Commission, type DistributionDashboard, type PayoutAccount, type TeamNode, type Withdrawal } from '@/api/financial'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n(); const app = useAppStore(); const auth = useAuthStore()
const activeTab = ref('overview'); const dashboard = ref<DistributionDashboard>(); const team = ref<TeamNode[]>([]); const ledger = ref<Commission[]>([]); const withdrawals = ref<Withdrawal[]>([]); const payout = ref<PayoutAccount>(); const search = ref(''); const graphContainer = ref<HTMLElement>(); const realName = ref(''); const alipay = ref(''); const withdrawAmount = ref(''); let graph: any
const tabs = computed(() => [{ id: 'overview', label: t('finance.distribution.overview') }, { id: 'team', label: t('finance.distribution.team') }, { id: 'ledger', label: t('finance.distribution.ledger') }, { id: 'withdraw', label: t('finance.distribution.withdraw') }])
const stats = computed(() => dashboard.value ? [{ label: t('finance.distribution.teamVolume'), value: cny(dashboard.value.team_volume_cny_minor) }, { label: t('finance.distribution.currentTier'), value: `T${dashboard.value.current_tier}` }, { label: t('common.available'), value: cny(dashboard.value.available_cny_minor) }, { label: t('common.frozenBalance'), value: cny(dashboard.value.frozen_cny_minor) }, ...(dashboard.value.debt_cny_minor > 0 ? [{ label: t('finance.distribution.debt'), value: cny(dashboard.value.debt_cny_minor) }] : [])] : [])
function cny(minor: number) { return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY' }).format(minor / 100) }
async function load() {
  try {
    dashboard.value = await getDistributionDashboard()
  } catch (e) {
    app.showError(extractApiErrorMessage(e))
    return
  }

  const [ledgerResult, withdrawalsResult, payoutResult] = await Promise.allSettled([
    getDistributionLedger(),
    listWithdrawals(),
    getPayoutAccount(),
  ])
  if (ledgerResult.status === 'fulfilled') ledger.value = ledgerResult.value.items
  if (withdrawalsResult.status === 'fulfilled') withdrawals.value = withdrawalsResult.value.items
  if (payoutResult.status === 'fulfilled') payout.value = payoutResult.value
  await loadTeam()
}
async function loadTeam(parent?: number) { try { team.value = (await getDistributionTree(parent, search.value)).items; await renderGraph(parent) } catch (e) { app.showError(extractApiErrorMessage(e)) } }
async function expandNode(node: TeamNode) { if (node.direct_children) await loadTeam(node.user_id) }
async function renderGraph(parent?: number) {
  await nextTick()
  if (!graphContainer.value) return
  graph?.destroy?.()
  const G6: any = await import('@antv/g6')
  const root = String(parent || auth.user?.id || 'me')
  
  const isDark = document.documentElement.classList.contains('dark')
  const brandColor = isDark ? '#E46B36' : '#D85A28'
  const surfaceColor = isDark ? '#181B1E' : '#FFFEFA'
  const textColor = isDark ? '#F5F3ED' : '#1B1D20'
  const lineColor = isDark ? 'rgba(255, 255, 255, 0.2)' : '#D9D8D2'

  const nodes = [
    {
      id: root,
      style: {
        labelText: parent ? `#${parent}` : t('finance.distribution.me'),
        fill: brandColor,
        stroke: brandColor,
        labelFill: textColor,
        labelFontFamily: 'Outfit, sans-serif',
        size: 38,
        labelPlacement: 'bottom'
      }
    },
    ...team.value.map(node => ({
      id: String(node.user_id),
      style: {
        labelText: node.username || node.email_masked,
        fill: surfaceColor,
        stroke: brandColor,
        labelFill: textColor,
        labelFontFamily: 'Outfit, sans-serif',
        size: 34,
        labelPlacement: 'bottom'
      }
    }))
  ]
  const edges = team.value.map(node => ({
    source: root,
    target: String(node.user_id)
  }))
  
  graph = new G6.Graph({
    container: graphContainer.value,
    data: { nodes, edges },
    autoFit: 'view',
    layout: { type: 'dagre', rankdir: 'LR' },
    behaviors: ['drag-canvas', 'zoom-canvas', 'drag-element'],
    node: {
      style: {
        lineWidth: 2,
      }
    },
    edge: {
      style: {
        stroke: lineColor,
        lineWidth: 1.5,
      }
    }
  })
  await graph.render()
}
async function saveAccount() { try { payout.value = await savePayoutAccount(alipay.value, realName.value); alipay.value = ''; realName.value = ''; app.showSuccess(t('common.saved')) } catch (e) { app.showError(extractApiErrorMessage(e)) } }
async function withdraw() { try { const minor = Math.round(Number(withdrawAmount.value) * 100); await createWithdrawal(minor); withdrawAmount.value = ''; dashboard.value = await getDistributionDashboard(); withdrawals.value = (await listWithdrawals()).items; app.showSuccess(t('common.success')) } catch (e) { app.showError(extractApiErrorMessage(e)) } }
watch(activeTab, value => { if (value === 'team') void renderGraph() })
onMounted(load); onBeforeUnmount(() => graph?.destroy?.())
</script>
