<template>
  <AppLayout>
    <div class="business-page-redesign mx-auto max-w-7xl space-y-6">
      <section class="flex flex-col gap-3 border-b border-gray-200 pb-5 sm:flex-row sm:items-end sm:justify-between dark:border-dark-700">
        <div>
          <h1 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('finance.distribution.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('finance.distribution.subtitle') }}</p>
        </div>
        <div class="inline-flex h-9 self-start border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-700 dark:bg-dark-800">
          <button v-for="tab in tabs" :key="tab.id" class="px-3 text-sm" :class="activeTab === tab.id ? 'bg-white font-medium text-gray-950 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500'" @click="activeTab = tab.id">{{ tab.label }}</button>
        </div>
      </section>

      <template v-if="dashboard">
        <section class="grid gap-px overflow-hidden border border-gray-200 bg-gray-200 sm:grid-cols-2 lg:grid-cols-4 dark:border-dark-700 dark:bg-dark-700">
          <div v-for="stat in stats" :key="stat.label" class="bg-white p-5 dark:bg-dark-900">
            <p class="text-xs font-medium uppercase text-gray-500">{{ stat.label }}</p>
            <p class="mt-2 font-mono text-2xl font-semibold tabular-nums text-gray-950 dark:text-white">{{ stat.value }}</p>
          </div>
        </section>

        <section v-if="activeTab === 'overview'" class="space-y-6">
          <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
            <div>
              <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.companySummary') }}</h2>
                <span v-if="dashboard.tier_override" class="text-xs font-medium text-amber-700 dark:text-amber-300">{{ t('finance.distribution.manualTier') }} · {{ tierDisplay(dashboard.tier_override, tierThreshold(dashboard.tier_override)) }}</span>
              </div>
              <div class="overflow-x-auto border border-gray-200 dark:border-dark-700">
                <table class="w-full min-w-[760px] text-sm">
                  <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">{{ t('finance.distribution.companyUnit') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.memberCount') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.recharge') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.commission') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.availableCommission') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.frozenCommission') }}</th></tr></thead>
                  <tbody><tr v-for="summary in companySummaries" :key="summary.depth" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3 font-medium">{{ companyUnitName(summary.depth) }}</td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ summary.member_count }}</td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(summary.recharge_cny_minor) }}</td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(summary.commission_cny_minor) }}</td><td class="px-4 py-3 text-right font-mono tabular-nums text-emerald-600">{{ cny(summary.available_cny_minor) }}</td><td class="px-4 py-3 text-right font-mono tabular-nums text-amber-600">{{ cny(summary.frozen_cny_minor) }}</td></tr></tbody>
                </table>
              </div>
            </div>
            <div class="border border-gray-200 p-5 dark:border-dark-700">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.tiers') }}</h2>
              <div class="mt-4 overflow-x-auto"><table class="w-full min-w-[520px] text-sm"><thead class="text-gray-500"><tr><th class="py-2 text-left">{{ t('finance.distribution.tier') }}</th><th class="py-2 text-right">{{ t('finance.distribution.threshold') }}</th><th v-for="(unit, unitIndex) in companyUnits" :key="unitIndex" class="py-2 text-right">{{ unit }}</th></tr></thead><tbody><tr v-for="tier in dashboard.tiers" :key="tier.tier" class="border-t border-gray-100 dark:border-dark-700"><td class="py-2 font-mono">{{ tierDisplay(tier.tier, tier.threshold_cny_minor) }}</td><td class="py-2 text-right font-mono tabular-nums">{{ compactAmount(tier.threshold_cny_minor) }}</td><td v-for="(rate, index) in tier.rates_bps" :key="`${tier.tier}-${index}`" class="py-2 text-right">{{ rate / 100 }}%</td></tr></tbody></table></div>
              <p class="mt-4 text-xs leading-5 text-gray-500">{{ t('finance.distribution.tierStatus', { tier: effectiveTier }) }}</p>
            </div>
          </div>
        </section>

        <section v-if="activeTab === 'team'" class="space-y-4">
          <div class="flex flex-wrap items-center gap-3"><input v-model="search" class="input min-w-[240px] flex-1" :placeholder="t('common.search')" @keyup.enter="() => loadTeam()" /><button v-if="teamParent" class="btn btn-secondary btn-sm" @click="loadTeam()">{{ t('finance.distribution.backToTeam') }}</button><button class="btn btn-secondary btn-sm" @click="loadTeam(teamParent)">{{ t('common.search') }}</button></div>
          <div class="overflow-x-auto border border-gray-200 dark:border-dark-700"><table class="w-full min-w-[820px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">{{ t('finance.distribution.member') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.directChildren') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.teamVolume') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.autoTier') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.effectiveTier') }}</th><th class="px-4 py-3 text-left">{{ t('common.actions') }}</th></tr></thead><tbody><tr v-for="node in team" :key="node.user_id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3"><span class="block font-medium text-gray-900 dark:text-white">{{ node.username || node.email_masked }}</span><span class="text-xs text-gray-500">{{ node.email_masked }}</span></td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ node.direct_children }}</td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(node.team_volume_cny_minor) }}</td><td class="px-4 py-3 text-right">{{ tierDisplay(node.auto_tier, tierThreshold(node.auto_tier)) }}</td><td class="px-4 py-3 text-right">{{ tierDisplay(node.effective_tier, tierThreshold(node.effective_tier)) }}<span v-if="node.tier_override" class="ml-1 text-xs text-amber-600">{{ t('finance.distribution.manual') }}</span></td><td class="px-4 py-3"><button v-if="node.direct_children" class="font-medium text-gray-700 underline decoration-gray-300 underline-offset-4 dark:text-gray-200" @click="expandNode(node)">{{ t('finance.distribution.viewTeam') }}</button><span v-else class="text-gray-400">-</span></td></tr><tr v-if="team.length === 0"><td colspan="6" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td></tr></tbody></table></div>
        </section>

        <section v-if="activeTab === 'ledger'" class="overflow-x-auto border border-gray-200 dark:border-dark-700"><table class="w-full min-w-[760px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">{{ t('finance.distribution.order') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.companyUnit') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.rate') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.commission') }}</th><th class="px-4 py-3 text-left">{{ t('common.status') }}</th></tr></thead><tbody><tr v-for="item in ledger" :key="item.id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">#{{ item.source_order_id }}</td><td class="px-4 py-3 text-right">{{ companyUnitName(item.depth) }}</td><td class="px-4 py-3 text-right">{{ item.rate_bps / 100 }}%</td><td class="px-4 py-3 text-right font-mono font-medium tabular-nums text-emerald-600">{{ cny(item.amount_cny_minor) }}</td><td class="px-4 py-3">{{ item.status }}</td></tr><tr v-if="ledger.length === 0"><td colspan="5" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td></tr></tbody></table></section>

        <section v-if="activeTab === 'withdraw'" class="space-y-6">
          <div class="grid gap-6 lg:grid-cols-2"><form class="border border-gray-200 p-5 dark:border-dark-700" @submit.prevent="saveAccount"><h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.payout') }}</h2><div v-if="payout" class="mt-4 text-sm text-gray-600 dark:text-gray-300">{{ payout.real_name_mask }} · {{ payout.account_mask }}</div><div class="mt-4 grid gap-3 sm:grid-cols-2"><input v-model="realName" class="input" :placeholder="t('finance.distribution.realName')" /><input v-model="alipay" class="input" :placeholder="t('finance.distribution.alipay')" /></div><button class="btn btn-primary mt-4">{{ t('common.save') }}</button></form><form class="border border-gray-200 p-5 dark:border-dark-700" @submit.prevent="withdraw"><h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.withdraw') }}</h2><p class="mt-3 text-sm text-gray-500">{{ t('common.available') }}: {{ cny(dashboard.available_cny_minor) }}</p><input v-model="withdrawAmount" class="input mt-4 w-full" inputmode="decimal" placeholder="100.00" /><button class="btn btn-primary mt-4">{{ t('finance.distribution.submitWithdrawal') }}</button></form></div>
          <div class="overflow-x-auto border border-gray-200 dark:border-dark-700"><table class="w-full min-w-[760px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">ID</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.withdraw') }}</th><th class="px-4 py-3 text-right">{{ t('finance.vouchers.fee') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.net') }}</th><th class="px-4 py-3 text-left">{{ t('common.status') }}</th><th class="px-4 py-3 text-left">{{ t('finance.admin.reference') }}</th></tr></thead><tbody><tr v-for="item in withdrawals" :key="item.id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">#{{ item.id }}</td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(item.amount_cny_minor) }}</td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(item.fee_cny_minor) }}</td><td class="px-4 py-3 text-right font-mono font-medium tabular-nums">{{ cny(item.amount_cny_minor - item.fee_cny_minor) }}</td><td class="px-4 py-3">{{ item.status }}<p v-if="item.reject_reason" class="mt-1 text-xs text-rose-600">{{ item.reject_reason }}</p></td><td class="px-4 py-3">{{ item.payment_reference || '-' }}</td></tr><tr v-if="withdrawals.length === 0"><td colspan="6" class="px-4 py-8 text-center text-gray-500">{{ t('common.noData') }}</td></tr></tbody></table></div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { createWithdrawal, getDistributionDashboard, getDistributionLedger, getDistributionTree, getPayoutAccount, listWithdrawals, savePayoutAccount, type Commission, type DistributionDashboard, type PayoutAccount, type TeamNode, type Withdrawal } from '@/api/financial'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { COMPUTE_COMPANY_UNIT_KEYS } from '@/constants/distribution'

const { t } = useI18n()
const app = useAppStore()
const activeTab = ref('overview')
const dashboard = ref<DistributionDashboard>()
const team = ref<TeamNode[]>([])
const ledger = ref<Commission[]>([])
const withdrawals = ref<Withdrawal[]>([])
const payout = ref<PayoutAccount>()
const search = ref('')
const teamParent = ref<number>()
const realName = ref('')
const alipay = ref('')
const withdrawAmount = ref('')
const companyUnitKeys = COMPUTE_COMPANY_UNIT_KEYS
const companyUnits = computed(() => companyUnitKeys.map(key => t(`finance.distribution.companyUnits.${key}`)))
const tabs = computed(() => [{ id: 'overview', label: t('finance.distribution.overview') }, { id: 'team', label: t('finance.distribution.team') }, { id: 'ledger', label: t('finance.distribution.ledger') }, { id: 'withdraw', label: t('finance.distribution.withdraw') }])
const companySummaries = computed(() => dashboard.value?.levels?.length ? dashboard.value.levels : Array.from({ length: companyUnitKeys.length }, (_, index) => ({ depth: index + 1, member_count: dashboard.value?.level_counts[index + 1] || 0, recharge_cny_minor: 0, commission_cny_minor: 0, available_cny_minor: 0, frozen_cny_minor: 0 })))
const effectiveTier = computed(() => dashboard.value?.current_tier || 0)
const companyMemberCount = computed(() => companySummaries.value.reduce((total, summary) => total + summary.member_count, 0))
const activeCompanyUnitCount = computed(() => companySummaries.value.filter(summary => summary.member_count > 0 || summary.recharge_cny_minor > 0).length)
const stats = computed(() => dashboard.value ? [{ label: t('finance.distribution.companyMembers'), value: formatCount(companyMemberCount.value) }, { label: t('finance.distribution.activeUnits'), value: `${activeCompanyUnitCount.value}/${companyUnitKeys.length}` }, { label: t('finance.distribution.teamVolume'), value: cny(dashboard.value.team_volume_cny_minor) }, { label: t('finance.distribution.effectiveTier'), value: tierDisplay(effectiveTier.value, tierThreshold(effectiveTier.value)) }, { label: t('finance.distribution.lifetimeEarned'), value: cny(dashboard.value.lifetime_earned_cny_minor) }, { label: t('common.available'), value: cny(dashboard.value.available_cny_minor) }, { label: t('common.frozenBalance'), value: cny(dashboard.value.frozen_cny_minor) }, { label: t('finance.distribution.withdrawing'), value: cny(dashboard.value.withdrawing_cny_minor) }, ...(dashboard.value.debt_cny_minor > 0 ? [{ label: t('finance.distribution.debt'), value: cny(dashboard.value.debt_cny_minor) }] : [])] : [])
function cny(minor: number) { return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY' }).format(minor / 100) }
function formatCount(value: number) { return new Intl.NumberFormat().format(value) }
function compactAmount(minor: number) {
  const amount = minor / 100
  if (amount >= 1_000_000) return `${amount / 1_000_000}M`
  if (amount >= 1_000) return `${amount / 1_000}K`
  return `${amount}`
}
function tierThreshold(tier: number) { return dashboard.value?.tiers.find(candidate => candidate.tier === tier)?.threshold_cny_minor || 0 }
function tierDisplay(tier: number, _thresholdMinor: number) { return `T${tier}` }
function companyUnitName(depth: number) { return companyUnits.value[depth - 1] || `${t('finance.distribution.companyUnit')} ${depth}` }
async function load() {
  try { dashboard.value = await getDistributionDashboard() } catch (error) { app.showError(extractApiErrorMessage(error)); return }
  const [ledgerResult, withdrawalsResult, payoutResult] = await Promise.allSettled([getDistributionLedger(), listWithdrawals(), getPayoutAccount()])
  if (ledgerResult.status === 'fulfilled') ledger.value = ledgerResult.value.items
  if (withdrawalsResult.status === 'fulfilled') withdrawals.value = withdrawalsResult.value.items
  if (payoutResult.status === 'fulfilled') payout.value = payoutResult.value
  await loadTeam()
}
async function loadTeam(parent?: number) { try { teamParent.value = parent; team.value = (await getDistributionTree(parent, search.value)).items } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function expandNode(node: TeamNode) { if (node.direct_children) await loadTeam(node.user_id) }
async function saveAccount() { try { payout.value = await savePayoutAccount(alipay.value, realName.value); alipay.value = ''; realName.value = ''; app.showSuccess(t('common.saved')) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function withdraw() { try { const minor = Math.round(Number(withdrawAmount.value) * 100); await createWithdrawal(minor); withdrawAmount.value = ''; dashboard.value = await getDistributionDashboard(); withdrawals.value = (await listWithdrawals()).items; app.showSuccess(t('common.success')) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
onMounted(load)
</script>
