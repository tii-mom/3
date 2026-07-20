<template>
  <AppLayout>
    <div class="space-y-5">
      <header class="flex flex-col gap-4 border-b border-gray-200 pb-5 lg:flex-row lg:items-end lg:justify-between dark:border-dark-700">
        <div>
          <h1 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('finance.admin.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('finance.distribution.subtitle') }}</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <button class="btn btn-secondary btn-sm" @click="toggleDistribution">{{ t('nav.distribution') }}: {{ distributionEnabled ? t('common.enabled') : t('common.disabled') }}</button>
          <button class="btn btn-secondary btn-sm" @click="toggleLegacyStack">{{ t('finance.admin.legacyStack') }}: {{ stackLegacy ? t('common.enabled') : t('common.disabled') }}</button>
          <button class="btn btn-secondary btn-sm" @click="toggleVouchers">{{ t('nav.balanceVouchers') }}: {{ vouchersEnabled ? t('common.enabled') : t('common.disabled') }}</button>
          <button class="btn btn-secondary btn-sm" @click="toggleBucketEnforcement">{{ bucketEnforced ? t('finance.admin.bucketEnforced') : t('finance.admin.bucketShadow') }}</button>
        </div>
      </header>

      <section class="grid gap-3 border border-gray-200 p-4 sm:grid-cols-3 dark:border-dark-700">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">TOTP<input v-model="security.totp" class="input mt-1.5 w-full" inputmode="numeric" maxlength="6" placeholder="000000" /></label>
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('finance.admin.reference') }}<input v-model="security.reference" class="input mt-1.5 w-full" /></label>
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('finance.admin.reason') }}<input v-model="security.reason" class="input mt-1.5 w-full" /></label>
        <div v-if="payoutDetails" class="sm:col-span-3 border-l-4 border-emerald-500 bg-emerald-50 px-4 py-3 text-sm text-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-200">
          {{ payoutDetails.real_name }} · {{ payoutDetails.account_type }} · {{ payoutDetails.account }}
        </div>
      </section>

      <form class="space-y-4 border border-gray-200 p-4 dark:border-dark-700" @submit.prevent="publishPolicy">
        <div class="flex flex-wrap items-center justify-between gap-3"><h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.admin.policy') }}</h2><span class="text-sm text-gray-500">{{ t('finance.admin.version') }} {{ policyVersion }}</span></div>
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-6">
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.freezeHours') }}<input v-model.number="policy.commission_freeze_hours" class="input mt-1 w-full" type="number" min="0" /></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.distribution.threshold') }}<input v-model.number="policy.withdrawal_min_cny_minor" class="input mt-1 w-full" type="number" min="1" /></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.dailyLimit') }}<input v-model.number="policy.withdrawal_daily_limit" class="input mt-1 w-full" type="number" min="1" /></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.feeBps') }}<input v-model.number="policy.withdrawal_fee_bps" class="input mt-1 w-full" type="number" min="0" max="9999" /></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.bonusBps') }}<input v-model.number="policy.first_recharge_bonus_bps" class="input mt-1 w-full" type="number" min="0" max="10000" /></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.bonusCap') }}<input v-model="policy.first_recharge_bonus_cap_usd" class="input mt-1 w-full" inputmode="decimal" /></label>
        </div>
        <div class="overflow-x-auto"><table class="w-full min-w-[1060px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-3 py-2 text-left">{{ t('finance.distribution.tier') }}</th><th class="px-3 py-2 text-right">{{ t('finance.distribution.threshold') }}</th><th v-for="(unit, unitIndex) in companyUnits" :key="unitIndex" class="px-3 py-2 text-right">{{ unit }} bps</th></tr></thead><tbody><tr v-for="tier in policy.tiers" :key="tier.tier" class="border-t border-gray-100 dark:border-dark-700"><td class="px-3 py-2 font-mono">{{ tierDisplay(tier.tier, tier.threshold_cny_minor) }}</td><td class="px-3 py-2"><input v-model.number="tier.threshold_cny_minor" class="input ml-auto w-32 text-right" type="number" min="1" :aria-label="`${tierDisplay(tier.tier, tier.threshold_cny_minor)} ${t('finance.distribution.threshold')}`" /></td><td v-for="(_, rateIndex) in tier.rates_bps" :key="rateIndex" class="px-3 py-2"><input v-model.number="tier.rates_bps[rateIndex]" class="input ml-auto w-24 text-right" type="number" min="0" max="10000" :aria-label="`${companyUnits[rateIndex]} bps`" /></td></tr></tbody></table></div>
        <div class="flex justify-end"><button class="btn btn-primary" :disabled="publishingPolicy">{{ t('finance.admin.publishPolicy') }}</button></div>
      </form>

      <nav class="overflow-x-auto">
        <div class="inline-flex h-9 min-w-max border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-700 dark:bg-dark-800">
          <button v-for="tab in tabs" :key="tab" class="px-4 text-sm" :class="active === tab ? 'bg-white font-medium shadow-sm dark:bg-dark-700' : 'text-gray-500'" @click="active = tab">{{ t(`finance.admin.${tab}`) }}</button>
        </div>
      </nav>

      <section v-if="active === 'tierAssignments'" class="space-y-4">
        <div class="flex flex-wrap items-center gap-3 border border-gray-200 p-4 dark:border-dark-700">
          <input v-model="tierSearch" class="input min-w-[260px] flex-1" :placeholder="t('finance.admin.tierSearchPlaceholder')" @keyup.enter="loadTierMembers" />
          <button class="btn btn-secondary btn-sm" :disabled="tierLoading" @click="loadTierMembers">{{ t('common.search') }}</button>
          <p class="w-full text-xs text-gray-500">{{ t('finance.admin.tierAssignmentHint') }}</p>
        </div>
        <div class="overflow-x-auto border border-gray-200 dark:border-dark-700"><table class="w-full min-w-[1080px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">{{ t('finance.admin.user') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.teamVolume') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.autoTier') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.manualTier') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.threshold') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.effectiveTier') }}</th><th class="px-4 py-3 text-left">{{ t('common.actions') }}</th></tr></thead><tbody><tr v-for="item in tierMembers" :key="item.user_id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3"><span class="block font-medium text-gray-900 dark:text-white">{{ item.email }}</span><span v-if="item.username" class="text-xs text-gray-500">{{ item.username }} · #{{ item.user_id }}</span></td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(item.team_volume_cny_minor) }}</td><td class="px-4 py-3 text-right">{{ tierDisplay(item.auto_tier, tierThreshold(item.auto_tier)) }}</td><td class="px-4 py-3 text-right"><select v-model="tierDrafts[item.user_id]" class="input min-w-28" :aria-label="`${item.email} tier`"><option :value="null">{{ t('finance.admin.automatic') }}</option><option :value="1">{{ tierDisplay(1, tierThreshold(1)) }}</option><option :value="2">{{ tierDisplay(2, tierThreshold(2)) }}</option><option :value="3">{{ tierDisplay(3, tierThreshold(3)) }}</option></select></td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ compactAmount(tierThreshold(tierDrafts[item.user_id] ?? item.effective_tier)) }}</td><td class="px-4 py-3 text-right">{{ tierDisplay(item.effective_tier, tierThreshold(item.effective_tier)) }}<span v-if="item.tier_override" class="ml-1 text-xs text-amber-600">{{ t('finance.distribution.manual') }}</span></td><td class="px-4 py-3"><button class="btn btn-primary btn-sm" :disabled="tierSaving === item.user_id" @click="saveTier(item.user_id)">{{ t('common.save') }}</button></td></tr><tr v-if="tierMembers.length === 0"><td colspan="7" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td></tr></tbody></table></div>
      </section>

      <section v-else-if="active === 'withdrawals'" class="overflow-x-auto border border-gray-200 dark:border-dark-700">
        <table class="w-full min-w-[900px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">ID</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.commission') }}</th><th class="px-4 py-3 text-left">{{ t('common.status') }}</th><th class="px-4 py-3 text-left">{{ t('common.actions') }}</th></tr></thead><tbody>
          <tr v-for="item in withdrawals" :key="item.id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">#{{ item.id }}</td><td class="px-4 py-3 text-right font-medium">{{ cny(item.amount_cny_minor) }}</td><td class="px-4 py-3">{{ item.status }}</td><td class="px-4 py-3"><div class="flex flex-wrap gap-2"><button class="btn btn-secondary btn-sm" @click="showPayout(item.id)">{{ t('finance.admin.payoutDetails') }}</button><button v-if="item.status === 'SUBMITTED'" class="btn btn-secondary btn-sm" @click="transition(item.id, 'APPROVED')">{{ t('finance.admin.approve') }}</button><button v-if="item.status === 'APPROVED'" class="btn btn-primary btn-sm" @click="transition(item.id, 'PAID')">{{ t('finance.admin.paid') }}</button><button v-if="item.status === 'SUBMITTED' || item.status === 'APPROVED'" class="btn btn-danger btn-sm" @click="transition(item.id, 'REJECTED')">{{ t('finance.admin.reject') }}</button></div></td></tr>
          <tr v-if="withdrawals.length === 0"><td colspan="4" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td></tr>
        </tbody></table>
      </section>

      <section v-else-if="active === 'vouchers'" class="overflow-x-auto border border-gray-200 dark:border-dark-700">
        <table class="w-full min-w-[760px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">ID</th><th class="px-4 py-3 text-left">{{ t('finance.vouchers.code') }}</th><th class="px-4 py-3 text-right">{{ t('finance.vouchers.amount') }}</th><th class="px-4 py-3 text-left">{{ t('common.status') }}</th><th class="px-4 py-3 text-left">{{ t('common.actions') }}</th></tr></thead><tbody>
          <tr v-for="item in vouchers" :key="item.id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">#{{ item.id }}</td><td class="px-4 py-3 font-mono">•••• {{ item.code_last4 }}</td><td class="px-4 py-3 text-right">${{ item.face_value }}</td><td class="px-4 py-3">{{ item.status }}</td><td class="px-4 py-3"><button v-if="item.status === 'ISSUED'" class="font-medium text-amber-600" @click="risk(item.id, true)">{{ t('finance.admin.lock') }}</button><button v-if="item.status === 'RISK_LOCKED'" class="font-medium text-emerald-600" @click="risk(item.id, false)">{{ t('finance.admin.unlock') }}</button></td></tr>
        </tbody></table>
      </section>

      <section v-else-if="active === 'commissions'" class="overflow-x-auto border border-gray-200 dark:border-dark-700">
        <table class="w-full min-w-[980px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">ID</th><th class="px-4 py-3 text-left">{{ t('finance.distribution.order') }}</th><th class="px-4 py-3 text-left">User</th><th class="px-4 py-3 text-left">Beneficiary</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.companyUnit') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.rate') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.commission') }}</th><th class="px-4 py-3 text-left">{{ t('common.status') }}</th></tr></thead><tbody>
          <tr v-for="item in commissions" :key="item.id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">#{{ item.id }}</td><td class="px-4 py-3">#{{ item.source_order_id }}</td><td class="px-4 py-3">#{{ item.source_user_id }}</td><td class="px-4 py-3">#{{ item.beneficiary_user_id }}</td><td class="px-4 py-3 text-right">{{ companyUnitName(item.depth) }}</td><td class="px-4 py-3 text-right">{{ item.rate_bps / 100 }}%</td><td class="px-4 py-3 text-right font-medium text-emerald-600">{{ cny(item.amount_cny_minor) }}</td><td class="px-4 py-3">{{ item.status }}</td></tr>
        </tbody></table>
      </section>

      <section v-else-if="active === 'recharges'" class="overflow-x-auto border border-gray-200 dark:border-dark-700">
        <table class="w-full min-w-[1080px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">ID</th><th class="px-4 py-3 text-left">{{ t('finance.distribution.order') }}</th><th class="px-4 py-3 text-left">User</th><th class="px-4 py-3 text-right">CNY</th><th class="px-4 py-3 text-right">USD</th><th class="px-4 py-3 text-right">Bonus USD</th><th class="px-4 py-3 text-right">Config</th><th class="px-4 py-3 text-left">{{ t('common.status') }}</th><th class="px-4 py-3 text-left">{{ t('common.actions') }}</th></tr></thead><tbody>
          <tr v-for="item in recharges" :key="item.id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">#{{ item.id }}</td><td class="px-4 py-3">#{{ item.source_order_id }}</td><td class="px-4 py-3">#{{ item.user_id }}</td><td class="px-4 py-3 text-right">{{ cny(item.base_cny_minor) }}</td><td class="px-4 py-3 text-right">${{ item.credited_usd }}</td><td class="px-4 py-3 text-right font-medium text-emerald-600">${{ item.first_recharge_bonus_usd }}</td><td class="px-4 py-3 text-right">v{{ item.config_version }}</td><td class="px-4 py-3">{{ item.status }}<p v-if="item.reversal_reason" class="mt-1 max-w-56 text-xs text-gray-500">{{ item.reversal_reason }}</p></td><td class="px-4 py-3"><button v-if="item.status === 'APPLIED'" class="btn btn-danger btn-sm" @click="reverseRecharge(item.id)">{{ t('finance.admin.reverseChargeback') }}</button></td></tr>
        </tbody></table>
      </section>

      <section v-else class="overflow-x-auto border border-gray-200 dark:border-dark-700">
        <table class="w-full min-w-[680px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">Ancestor</th><th class="px-4 py-3 text-left">Descendant</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.companyUnit') }}</th><th class="px-4 py-3 text-left">Created</th></tr></thead><tbody>
          <tr v-for="(item, index) in relations" :key="`${item.ancestor_user_id}-${item.descendant_user_id}-${index}`" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">#{{ item.ancestor_user_id }}</td><td class="px-4 py-3">#{{ item.descendant_user_id }}</td><td class="px-4 py-3 text-right">{{ companyUnitName(item.depth) }}</td><td class="px-4 py-3">{{ date(item.created_at) }}</td></tr>
        </tbody></table>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { createDistributionPolicyVersion, getDistributionConfig, getFinancialRuntimeConfig, getVoucherConfig, getWithdrawalPayoutDetails, listAdminCommissions, listAdminVouchers, listAdminWithdrawals, listDistributionRelations, listDistributionTierMembers, listRechargeEvents, reverseRechargeEvent, setDistributionTierOverride, setVoucherRiskLock, transitionWithdrawal, updateDistributionConfig, updateFinancialRuntimeConfig, updateVoucherConfig, type AdminCommission, type DistributionPolicyInput, type DistributionRelation, type DistributionTierMember, type PayoutDetails, type RechargeEvent } from '@/api/admin/finance'
import type { Voucher, Withdrawal } from '@/api/financial'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { COMPUTE_COMPANY_UNIT_KEYS } from '@/constants/distribution'

const { t } = useI18n()
const app = useAppStore()
const tabs = ['tierAssignments', 'withdrawals', 'vouchers', 'commissions', 'recharges', 'relations'] as const
const active = ref<(typeof tabs)[number]>('withdrawals')
const withdrawals = ref<Withdrawal[]>([])
const vouchers = ref<Voucher[]>([])
const commissions = ref<AdminCommission[]>([])
const recharges = ref<RechargeEvent[]>([])
const relations = ref<DistributionRelation[]>([])
const tierMembers = ref<DistributionTierMember[]>([])
const tierSearch = ref('')
const tierLoading = ref(false)
const tierSaving = ref<number>()
const tierDrafts = reactive<Record<number, number | null>>({})
const security = reactive({ totp: '', reference: '', reason: '' })
const payoutDetails = ref<PayoutDetails>()
const distributionEnabled = ref(false)
const stackLegacy = ref(false)
const vouchersEnabled = ref(false)
const bucketEnforced = ref(false)
const policyVersion = ref(1)
const publishingPolicy = ref(false)
const policy = reactive<Omit<DistributionPolicyInput, 'totp_code'>>({ commission_freeze_hours: 168, withdrawal_min_cny_minor: 10000, withdrawal_daily_limit: 1, withdrawal_fee_bps: 0, first_recharge_bonus_bps: 1000, first_recharge_bonus_cap_usd: '10000', tiers: [] })
const companyUnitKeys = COMPUTE_COMPANY_UNIT_KEYS
const companyUnits = computed(() => companyUnitKeys.map(key => t(`finance.distribution.companyUnits.${key}`)))
const cny = (minor: number) => new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY' }).format(minor / 100)
const date = (value: string) => new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
const tierThreshold = (tier: number) => policy.tiers.find(item => item.tier === tier)?.threshold_cny_minor || 0
const compactAmount = (minor: number) => {
  const amount = minor / 100
  if (amount >= 1_000_000) return `${amount / 1_000_000}M`
  if (amount >= 1_000) return `${amount / 1_000}K`
  return `${amount}`
}
const tierDisplay = (tier: number, _thresholdMinor: number) => `T${tier}`
const companyUnitName = (depth: number) => companyUnits.value[depth - 1] || `${t('finance.distribution.companyUnit')} ${depth}`

async function load() {
  try {
    const [withdrawalPage, voucherPage, commissionPage, rechargePage, relationPage, tierPage, distributionConfig, voucherConfig, runtimeConfig] = await Promise.all([listAdminWithdrawals(), listAdminVouchers(), listAdminCommissions(), listRechargeEvents(), listDistributionRelations(), listDistributionTierMembers(), getDistributionConfig(), getVoucherConfig(), getFinancialRuntimeConfig()])
    withdrawals.value = withdrawalPage.items
    vouchers.value = voucherPage.items
    commissions.value = commissionPage.items
    recharges.value = rechargePage.items
    relations.value = relationPage.items
    tierMembers.value = tierPage.items
    tierPage.items.forEach(item => { tierDrafts[item.user_id] = item.tier_override ?? null })
    distributionEnabled.value = distributionConfig.enabled
    stackLegacy.value = distributionConfig.stack_with_legacy
    vouchersEnabled.value = voucherConfig.enabled
    bucketEnforced.value = runtimeConfig.credit_bucket_enforce_enabled
    policyVersion.value = distributionConfig.current_config_version
    Object.assign(policy, { commission_freeze_hours: distributionConfig.commission_freeze_hours, withdrawal_min_cny_minor: distributionConfig.withdrawal_min_cny_minor, withdrawal_daily_limit: distributionConfig.withdrawal_daily_limit, withdrawal_fee_bps: distributionConfig.withdrawal_fee_bps, first_recharge_bonus_bps: distributionConfig.first_recharge_bonus_bps, first_recharge_bonus_cap_usd: distributionConfig.first_recharge_bonus_cap_usd, tiers: distributionConfig.tiers.map(tier => ({ ...tier, rates_bps: [...tier.rates_bps] as [number, number, number, number, number] })) })
  } catch (error) { app.showError(extractApiErrorMessage(error)) }
}
async function loadTierMembers() { tierLoading.value = true; try { const result = await listDistributionTierMembers(tierSearch.value); tierMembers.value = result.items; result.items.forEach(item => { tierDrafts[item.user_id] = item.tier_override ?? null }) } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { tierLoading.value = false } }
async function saveTier(userId: number) { if (!security.totp) return app.showError('TOTP required'); tierSaving.value = userId; try { const item = await setDistributionTierOverride(userId, tierDrafts[userId] ?? null, security.reason, security.totp); const index = tierMembers.value.findIndex(candidate => candidate.user_id === userId); if (index >= 0) tierMembers.value[index] = item; tierDrafts[userId] = item.tier_override ?? null; app.showSuccess(t('common.saved')) } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { tierSaving.value = undefined } }
async function toggleDistribution() { if (!security.totp) return app.showError('TOTP required'); try { await updateDistributionConfig(!distributionEnabled.value, stackLegacy.value, security.totp); distributionEnabled.value = !distributionEnabled.value } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function toggleLegacyStack() { if (!security.totp) return app.showError('TOTP required'); try { await updateDistributionConfig(distributionEnabled.value, !stackLegacy.value, security.totp); stackLegacy.value = !stackLegacy.value } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function toggleVouchers() { if (!security.totp) return app.showError('TOTP required'); try { await updateVoucherConfig(!vouchersEnabled.value, security.totp); vouchersEnabled.value = !vouchersEnabled.value } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function toggleBucketEnforcement() { if (!security.totp) return app.showError('TOTP required'); try { await updateFinancialRuntimeConfig(!bucketEnforced.value, security.totp); bucketEnforced.value = !bucketEnforced.value } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function publishPolicy() { if (!security.totp) return app.showError('TOTP required'); publishingPolicy.value = true; try { const result = await createDistributionPolicyVersion({ ...policy, tiers: policy.tiers.map(tier => ({ ...tier, rates_bps: [...tier.rates_bps] as [number, number, number, number, number] })), totp_code: security.totp }); policyVersion.value = result.config_version; app.showSuccess(t('common.saved')); await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { publishingPolicy.value = false } }
async function transition(id: number, status: string) { try { await transitionWithdrawal(id, { status, reason: security.reason, payment_reference: security.reference, totp_code: security.totp }); payoutDetails.value = undefined; await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function risk(id: number, locked: boolean) { if (!security.totp) return app.showError('TOTP required'); try { await setVoucherRiskLock(id, locked, security.reason, security.totp); await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function showPayout(id: number) { try { payoutDetails.value = await getWithdrawalPayoutDetails(id, security.totp) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function reverseRecharge(id: number) { if (!security.totp || !security.reason.trim()) return app.showError(t('finance.admin.reversalSecurityRequired')); if (!window.confirm(t('finance.admin.reverseChargebackConfirm'))) return; try { await reverseRechargeEvent(id, { reversal_type: 'CHARGEBACK', reason: security.reason.trim(), totp_code: security.totp }); await load(); app.showSuccess(t('common.success')) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
onMounted(load)
</script>
