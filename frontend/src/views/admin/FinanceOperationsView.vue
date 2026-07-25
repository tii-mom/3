<template>
  <AppLayout>
    <div class="space-y-5">
      <header class="flex flex-col gap-4 border-b border-gray-200 pb-5 lg:flex-row lg:items-end lg:justify-between dark:border-dark-700">
        <div>
          <h1 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('finance.admin.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('finance.distribution.subtitle') }}</p>
        </div>
        <div class="grid gap-2 sm:grid-cols-2 lg:min-w-[520px]">
          <button type="button" class="group flex items-center justify-between gap-4 border px-4 py-3 text-left transition-colors" :class="featureToggleClass(distributionEnabled)" @click="toggleDistribution">
            <span class="flex items-center gap-3">
              <span class="flex h-9 w-9 items-center justify-center border" :class="featureIconClass(distributionEnabled)"><Icon name="cpu" size="sm" /></span>
              <span>
                <span class="block text-sm font-semibold text-gray-900 dark:text-white">{{ t('nav.distribution') }}</span>
                <span class="mt-0.5 flex items-center gap-1.5 text-xs text-gray-500 dark:text-dark-400"><span class="h-1.5 w-1.5 rounded-full" :class="distributionEnabled ? 'bg-emerald-500' : 'bg-gray-400'"></span>{{ distributionEnabled ? t('common.enabled') : t('common.disabled') }}</span>
              </span>
            </span>
            <span class="text-xs font-semibold" :class="distributionEnabled ? 'text-rose-600 dark:text-rose-400' : 'text-emerald-600 dark:text-emerald-400'">{{ distributionEnabled ? t('finance.admin.turnOff') : t('finance.admin.turnOn') }}</span>
          </button>
          <button type="button" class="group flex items-center justify-between gap-4 border px-4 py-3 text-left transition-colors" :class="featureToggleClass(vouchersEnabled)" @click="toggleVouchers">
            <span class="flex items-center gap-3">
              <span class="flex h-9 w-9 items-center justify-center border" :class="featureIconClass(vouchersEnabled)"><Icon name="gift" size="sm" /></span>
              <span>
                <span class="block text-sm font-semibold text-gray-900 dark:text-white">{{ t('nav.balanceVouchers') }}</span>
                <span class="mt-0.5 flex items-center gap-1.5 text-xs text-gray-500 dark:text-dark-400"><span class="h-1.5 w-1.5 rounded-full" :class="vouchersEnabled ? 'bg-emerald-500' : 'bg-gray-400'"></span>{{ vouchersEnabled ? t('common.enabled') : t('common.disabled') }}</span>
              </span>
            </span>
            <span class="text-xs font-semibold" :class="vouchersEnabled ? 'text-rose-600 dark:text-rose-400' : 'text-emerald-600 dark:text-emerald-400'">{{ vouchersEnabled ? t('finance.admin.turnOff') : t('finance.admin.turnOn') }}</span>
          </button>
        </div>
      </header>

      <section class="grid gap-px overflow-hidden border border-gray-200 bg-gray-200 sm:grid-cols-2 xl:grid-cols-4 dark:border-dark-700 dark:bg-dark-700" :aria-label="t('finance.admin.operationsSnapshot')">
        <article v-for="item in operationsSnapshot" :key="item.label" class="bg-white p-4 dark:bg-dark-900">
          <div class="flex items-center justify-between gap-3">
            <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ item.label }}</p>
            <Icon :name="item.icon" size="sm" class="text-primary-600 dark:text-primary-400" />
          </div>
          <p class="mt-2 font-mono text-xl font-semibold tabular-nums text-gray-950 dark:text-white">{{ item.value }}</p>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ item.hint }}</p>
        </article>
      </section>

      <section class="grid gap-3 border border-gray-200 bg-white/70 p-4 sm:grid-cols-3 dark:border-dark-700 dark:bg-dark-900/60">
        <div class="sm:col-span-3">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('finance.admin.securityFields') }}</p>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.admin.securityHint') }}</p>
        </div>
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">TOTP<input v-model="security.totp" class="input mt-1.5 w-full" inputmode="numeric" maxlength="6" placeholder="000000" /></label>
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('finance.admin.reference') }}<input v-model="security.reference" class="input mt-1.5 w-full" /></label>
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('finance.admin.reason') }}<input v-model="security.reason" class="input mt-1.5 w-full" /></label>
        <div v-if="payoutDetails" class="sm:col-span-3 border-l-4 border-emerald-500 bg-emerald-50 px-4 py-3 text-sm text-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-200">
          {{ payoutDetails.real_name }} · {{ payoutDetails.account_type }} · {{ payoutDetails.account }}
        </div>
      </section>

      <form class="space-y-4 border border-gray-200 bg-white/70 p-4 dark:border-dark-700 dark:bg-dark-900/60" @submit.prevent="publishPolicy">
        <div class="flex flex-wrap items-start justify-between gap-3"><div><h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.admin.policy') }}</h2><p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.admin.policyHint') }}</p></div><span class="border border-gray-200 px-3 py-1 text-sm text-gray-500 dark:border-dark-700">{{ t('finance.admin.version') }} {{ policyVersion }}</span></div>
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.freezeHours') }}<input v-model.number="policy.commission_freeze_hours" class="input mt-1 w-full" type="number" min="0" /></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.minimumWithdrawal') }}<div class="relative mt-1"><input :value="minorToYuanInput(policy.withdrawal_min_cny_minor)" class="input w-full pr-10 text-right" type="number" min="0.01" step="0.01" inputmode="decimal" @input="setWithdrawalMin" /><span class="pointer-events-none absolute inset-y-0 right-3 flex items-center text-xs text-gray-500">{{ t('finance.admin.yuanUnit') }}</span></div></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.dailyLimit') }}<input v-model.number="policy.withdrawal_daily_limit" class="input mt-1 w-full text-right" type="number" min="1" /></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.bonusCap') }}<input v-model="policy.first_recharge_bonus_cap_usd" class="input mt-1 w-full text-right" inputmode="decimal" /></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.feeBps') }}<div class="relative mt-1"><input :value="bpsToPercentInput(policy.withdrawal_fee_bps)" class="input w-full pr-9 text-right" type="number" min="0" max="99.99" step="0.01" inputmode="decimal" @input="setWithdrawalFee" /><span class="pointer-events-none absolute inset-y-0 right-3 flex items-center text-xs text-gray-500">%</span></div></label>
          <label class="text-xs font-medium text-gray-500">{{ t('finance.admin.bonusBps') }}<div class="relative mt-1"><input :value="bpsToPercentInput(policy.first_recharge_bonus_bps)" class="input w-full pr-9 text-right" type="number" min="0" max="100" step="0.01" inputmode="decimal" @input="setBonusBps" /><span class="pointer-events-none absolute inset-y-0 right-3 flex items-center text-xs text-gray-500">%</span></div></label>
          <div class="border border-primary-200 bg-primary-50 px-3 py-2 dark:border-primary-900/60 dark:bg-primary-950/20"><p class="text-xs font-medium text-primary-700 dark:text-primary-300">{{ t('finance.admin.purchaseRatio') }}</p><p class="mt-1 font-mono text-sm font-semibold text-gray-900 dark:text-white">1 CNY = {{ purchaseMultiplier }} USD</p></div>
          <div class="border border-gray-200 px-3 py-2 dark:border-dark-700"><p class="text-xs font-medium text-gray-500">{{ t('finance.admin.currentWithdrawalRule') }}</p><p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ cny(policy.withdrawal_min_cny_minor) }} · {{ bpsToPercentInput(policy.withdrawal_fee_bps) }}%</p></div>
        </div>
        <div class="overflow-x-auto"><table class="w-full min-w-[1060px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-3 py-2 text-left">{{ t('finance.distribution.tier') }}</th><th class="px-3 py-2 text-right">{{ t('finance.distribution.threshold') }}</th><th v-for="(unit, unitIndex) in companyUnits" :key="unitIndex" class="px-3 py-2 text-right">{{ unit }}</th></tr></thead><tbody><tr v-for="tier in policy.tiers" :key="tier.tier" class="border-t border-gray-100 dark:border-dark-700"><td class="px-3 py-2"><span class="font-mono font-semibold">{{ tierDisplay(tier.tier, tier.threshold_cny_minor) }}</span><span class="ml-2 text-xs text-gray-500">{{ compactAmount(tier.threshold_cny_minor) }}</span></td><td class="px-3 py-2"><div class="relative ml-auto w-36"><input :value="minorToYuanInput(tier.threshold_cny_minor)" class="input w-full pr-10 text-right" type="number" :min="tier.tier === 0 ? 0 : 0.01" step="0.01" inputmode="decimal" :aria-label="`${tierDisplay(tier.tier, tier.threshold_cny_minor)} ${t('finance.distribution.threshold')}`" @input="setTierThreshold(tier, $event)" /><span class="pointer-events-none absolute inset-y-0 right-3 flex items-center text-xs text-gray-500">{{ t('finance.admin.yuanUnit') }}</span></div></td><td v-for="(_, rateIndex) in tier.rates_bps" :key="rateIndex" class="px-3 py-2"><div class="relative ml-auto w-24"><input :value="bpsToPercentInput(tier.rates_bps[rateIndex])" class="input w-full pr-8 text-right" type="number" min="0" max="100" step="0.01" inputmode="decimal" :aria-label="`${companyUnits[rateIndex]} ${t('finance.distribution.rate')}`" @input="setTierRate(tier, rateIndex, $event)" /><span class="pointer-events-none absolute inset-y-0 right-3 flex items-center text-xs text-gray-500">%</span></div></td></tr></tbody></table></div>
        <div class="flex flex-wrap items-center justify-between gap-3"><p class="text-xs text-gray-500">{{ t('finance.admin.purchaseRatioHint', { rate: purchaseMultiplier }) }}</p><button class="btn btn-primary" :disabled="publishingPolicy">{{ t('finance.admin.publishPolicy') }}</button></div>
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
        <div class="overflow-x-auto border border-gray-200 dark:border-dark-700"><table class="w-full min-w-[1080px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">{{ t('finance.admin.user') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.teamVolume') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.autoTier') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.manualTier') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.threshold') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.effectiveTier') }}</th><th class="px-4 py-3 text-left">{{ t('common.actions') }}</th></tr></thead><tbody><tr v-for="item in tierMembers" :key="item.user_id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3"><span class="block font-medium text-gray-900 dark:text-white">{{ item.email }}</span><span v-if="item.username" class="text-xs text-gray-500">{{ item.username }} · #{{ item.user_id }}</span></td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(item.team_volume_cny_minor) }}</td><td class="px-4 py-3 text-right">{{ tierDisplay(item.auto_tier, tierThreshold(item.auto_tier)) }}</td><td class="px-4 py-3 text-right"><select v-model="tierDrafts[item.user_id]" class="input min-w-28" :aria-label="`${item.email} tier`"><option :value="null">{{ t('finance.admin.automatic') }}</option><option v-for="tier in policy.tiers" :key="tier.tier" :value="tier.tier">{{ tierDisplay(tier.tier, tier.threshold_cny_minor) }}</option></select></td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ compactAmount(tierThreshold(tierDrafts[item.user_id] ?? item.effective_tier)) }}</td><td class="px-4 py-3 text-right">{{ tierDisplay(item.effective_tier, tierThreshold(item.effective_tier)) }}<span v-if="item.tier_override !== undefined" class="ml-1 text-xs text-amber-600">{{ t('finance.distribution.manual') }}</span></td><td class="px-4 py-3"><button class="btn btn-primary btn-sm" :disabled="tierSaving === item.user_id" @click="saveTier(item.user_id)">{{ t('common.save') }}</button></td></tr><tr v-if="tierMembers.length === 0"><td colspan="7" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td></tr></tbody></table></div>
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

      <section v-else-if="active === 'relations'" class="overflow-x-auto border border-gray-200 dark:border-dark-700">
        <table class="w-full min-w-[680px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">Ancestor</th><th class="px-4 py-3 text-left">Descendant</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.companyUnit') }}</th><th class="px-4 py-3 text-left">Created</th></tr></thead><tbody>
          <tr v-for="(item, index) in relations" :key="`${item.ancestor_user_id}-${item.descendant_user_id}-${index}`" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">#{{ item.ancestor_user_id }}</td><td class="px-4 py-3">#{{ item.descendant_user_id }}</td><td class="px-4 py-3 text-right">{{ companyUnitName(item.depth) }}</td><td class="px-4 py-3">{{ date(item.created_at) }}</td></tr>
        </tbody></table>
      </section>

      <section v-else-if="active === 'conversions'" class="overflow-x-auto border border-gray-200 dark:border-dark-700">
        <table class="w-full min-w-[1040px] text-sm"><thead class="bg-gray-50 text-gray-500 dark:bg-dark-800"><tr><th class="px-4 py-3 text-left">ID</th><th class="px-4 py-3 text-right">{{ t('finance.admin.user') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.cnyAmount') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.usdAmount') }}</th><th class="px-4 py-3 text-right">{{ t('finance.distribution.exchangeRate') }}</th><th class="px-4 py-3 text-left">{{ t('finance.admin.rateSource') }}</th><th class="px-4 py-3 text-right">{{ t('finance.admin.version') }}</th><th class="px-4 py-3 text-left">{{ t('finance.distribution.createdAt') }}</th></tr></thead><tbody><tr v-for="item in conversions" :key="item.id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">#{{ item.id }}</td><td class="px-4 py-3 text-right">#{{ item.user_id }}</td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(item.amount_cny_minor) }}</td><td class="px-4 py-3 text-right font-mono tabular-nums">${{ item.usd_amount }}</td><td class="px-4 py-3 text-right font-mono tabular-nums">{{ item.cny_to_usd_rate ? `1 CNY = ${item.cny_to_usd_rate} USD` : `1 USD = ¥${item.usd_to_cny_rate}` }}</td><td class="px-4 py-3 text-left text-xs text-gray-500">{{ item.rate_source || 'legacy' }}</td><td class="px-4 py-3 text-right">v{{ item.config_version }}</td><td class="px-4 py-3">{{ date(item.created_at) }}</td></tr><tr v-if="conversions.length === 0"><td colspan="8" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td></tr></tbody></table>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { createDistributionPolicyVersion, getDistributionConfig, getVoucherConfig, getWithdrawalPayoutDetails, listAdminCommissions, listAdminVouchers, listAdminWithdrawals, listDistributionConversions, listDistributionRelations, listDistributionTierMembers, listRechargeEvents, reverseRechargeEvent, setDistributionTierOverride, setVoucherRiskLock, transitionWithdrawal, updateDistributionConfig, updateVoucherConfig, type AdminCommission, type DistributionConfig, type DistributionConversionAudit, type DistributionPolicyInput, type DistributionRelation, type DistributionTierMember, type PayoutDetails, type RechargeEvent } from '@/api/admin/finance'
import type { Voucher, Withdrawal } from '@/api/financial'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { COMPUTE_COMPANY_UNIT_KEYS } from '@/constants/distribution'

const { t } = useI18n()
const app = useAppStore()
const tabs = ['tierAssignments', 'withdrawals', 'vouchers', 'commissions', 'recharges', 'relations', 'conversions'] as const
const active = ref<(typeof tabs)[number]>('tierAssignments')
const withdrawals = ref<Withdrawal[]>([])
const vouchers = ref<Voucher[]>([])
const commissions = ref<AdminCommission[]>([])
const recharges = ref<RechargeEvent[]>([])
const relations = ref<DistributionRelation[]>([])
const conversions = ref<DistributionConversionAudit[]>([])
const tierMembers = ref<DistributionTierMember[]>([])
const tierSearch = ref('')
const tierLoading = ref(false)
const tierSaving = ref<number>()
const tierDrafts = reactive<Record<number, number | null>>({})
const security = reactive({ totp: '', reference: '', reason: '' })
const payoutDetails = ref<PayoutDetails>()
const distributionEnabled = ref(false)
const vouchersEnabled = ref(false)
const policyVersion = ref(1)
const publishingPolicy = ref(false)
const policy = reactive<Omit<DistributionPolicyInput, 'totp_code'>>({ commission_freeze_hours: 168, withdrawal_min_cny_minor: 2000, withdrawal_daily_limit: 1, withdrawal_fee_bps: 0, first_recharge_bonus_bps: 1000, first_recharge_bonus_cap_usd: '10000', tiers: [] })
const purchaseMultiplier = ref('1')
const companyUnitKeys = COMPUTE_COMPANY_UNIT_KEYS
const companyUnits = computed(() => companyUnitKeys.map(key => t(`finance.distribution.companyUnits.${key}`)))
const cny = (minor: number) => new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY' }).format(minor / 100)
const date = (value: string) => new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
const tierThreshold = (tier: number) => policy.tiers.find(item => item.tier === tier)?.threshold_cny_minor || 0
const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value))
const eventNumber = (event: Event) => {
  const value = Number((event.target as HTMLInputElement).value)
  return Number.isFinite(value) ? value : 0
}
const minorToYuanInput = (minor: number) => String(Number(((minor || 0) / 100).toFixed(2)))
const bpsToPercentInput = (bps: number) => String(Number(((bps || 0) / 100).toFixed(2)))
const yuanToMinor = (value: number) => Math.round(Math.max(0, value) * 100)
const percentToBps = (value: number, max = 100) => Math.round(clamp(value, 0, max) * 100)
const compactAmount = (minor: number) => {
  const amount = minor / 100
  if (amount >= 1_000_000) return `${amount / 1_000_000}M`
  if (amount >= 1_000) return `${amount / 1_000}K`
  return `${amount}`
}
const tierDisplay = (tier: number, _thresholdMinor: number) => `T${tier}`
const companyUnitName = (depth: number) => companyUnits.value[depth - 1] || `${t('finance.distribution.companyUnit')} ${depth}`
const setWithdrawalMin = (event: Event) => { policy.withdrawal_min_cny_minor = yuanToMinor(eventNumber(event)) }
const setWithdrawalFee = (event: Event) => { policy.withdrawal_fee_bps = percentToBps(eventNumber(event), 99.99) }
const setBonusBps = (event: Event) => { policy.first_recharge_bonus_bps = percentToBps(eventNumber(event)) }
const setTierThreshold = (tier: DistributionConfig['tiers'][number], event: Event) => { tier.threshold_cny_minor = yuanToMinor(eventNumber(event)) }
const setTierRate = (tier: DistributionConfig['tiers'][number], rateIndex: number, event: Event) => { tier.rates_bps[rateIndex] = percentToBps(eventNumber(event)) }
const featureToggleClass = (enabled: boolean) => enabled
  ? 'border-emerald-200 bg-emerald-50/70 hover:border-emerald-300 dark:border-emerald-900/60 dark:bg-emerald-950/20'
  : 'border-gray-200 bg-white hover:border-primary-200 dark:border-dark-700 dark:bg-dark-900 dark:hover:border-primary-900/70'
const featureIconClass = (enabled: boolean) => enabled
  ? 'border-emerald-200 bg-white text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-950/40 dark:text-emerald-300'
  : 'border-gray-200 bg-gray-50 text-gray-500 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300'
const pendingWithdrawals = computed(() => withdrawals.value.filter(item => item.status === 'SUBMITTED' || item.status === 'APPROVED').length)
const operationsSnapshot = computed(() => [
  { label: t('finance.admin.policyVersion'), value: `v${policyVersion.value}`, hint: t('finance.admin.policyVersionHint'), icon: 'badge' as const },
  { label: t('finance.admin.pendingWithdrawals'), value: String(pendingWithdrawals.value), hint: pendingWithdrawals.value > 0 ? t('finance.admin.pendingWithdrawalsHint') : t('finance.admin.noPendingWithdrawals'), icon: 'creditCard' as const },
  { label: t('finance.admin.minimumWithdrawalShort'), value: cny(policy.withdrawal_min_cny_minor), hint: `${t('finance.admin.dailyLimit')}: ${policy.withdrawal_daily_limit}`, icon: 'shield' as const },
  { label: t('finance.admin.purchaseRatio'), value: `1 CNY = ${purchaseMultiplier.value} USD`, hint: t('finance.admin.purchaseRatioShortHint'), icon: 'swap' as const }
])

async function load() {
  try {
    const [withdrawalPage, voucherPage, commissionPage, rechargePage, relationPage, tierPage, distributionConfig, voucherConfig] = await Promise.all([listAdminWithdrawals(), listAdminVouchers(), listAdminCommissions(), listRechargeEvents(), listDistributionRelations(), listDistributionTierMembers(), getDistributionConfig(), getVoucherConfig()])
    withdrawals.value = withdrawalPage.items
    vouchers.value = voucherPage.items
    commissions.value = commissionPage.items
    recharges.value = rechargePage.items
    relations.value = relationPage.items
    try { conversions.value = (await listDistributionConversions()).items } catch { conversions.value = [] }
    tierMembers.value = tierPage.items
    tierPage.items.forEach(item => { tierDrafts[item.user_id] = item.tier_override ?? null })
    distributionEnabled.value = distributionConfig.enabled
    vouchersEnabled.value = voucherConfig.enabled
    policyVersion.value = distributionConfig.current_config_version
    purchaseMultiplier.value = distributionConfig.balance_recharge_multiplier || '1'
    Object.assign(policy, { commission_freeze_hours: distributionConfig.commission_freeze_hours, withdrawal_min_cny_minor: distributionConfig.withdrawal_min_cny_minor, withdrawal_daily_limit: distributionConfig.withdrawal_daily_limit, withdrawal_fee_bps: distributionConfig.withdrawal_fee_bps, first_recharge_bonus_bps: distributionConfig.first_recharge_bonus_bps, first_recharge_bonus_cap_usd: distributionConfig.first_recharge_bonus_cap_usd, tiers: distributionConfig.tiers.map(tier => ({ ...tier, rates_bps: [...tier.rates_bps] as [number, number, number, number, number] })) })
  } catch (error) { app.showError(extractApiErrorMessage(error)) }
}
async function loadTierMembers() { tierLoading.value = true; try { const result = await listDistributionTierMembers(tierSearch.value); tierMembers.value = result.items; result.items.forEach(item => { tierDrafts[item.user_id] = item.tier_override ?? null }) } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { tierLoading.value = false } }
async function saveTier(userId: number) { if (!security.totp) return app.showError('TOTP required'); tierSaving.value = userId; try { const item = await setDistributionTierOverride(userId, tierDrafts[userId] ?? null, security.reason, security.totp); const index = tierMembers.value.findIndex(candidate => candidate.user_id === userId); if (index >= 0) tierMembers.value[index] = item; tierDrafts[userId] = item.tier_override ?? null; app.showSuccess(t('common.saved')) } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { tierSaving.value = undefined } }
async function toggleDistribution() { if (!security.totp) return app.showError('TOTP required'); try { await updateDistributionConfig(!distributionEnabled.value, security.totp); distributionEnabled.value = !distributionEnabled.value } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function toggleVouchers() { if (!security.totp) return app.showError('TOTP required'); try { await updateVoucherConfig(!vouchersEnabled.value, security.totp); vouchersEnabled.value = !vouchersEnabled.value } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function publishPolicy() { if (!security.totp) return app.showError('TOTP required'); publishingPolicy.value = true; try { const result = await createDistributionPolicyVersion({ ...policy, tiers: policy.tiers.map(tier => ({ ...tier, rates_bps: [...tier.rates_bps] as [number, number, number, number, number] })), totp_code: security.totp }); policyVersion.value = result.config_version; app.showSuccess(t('common.saved')); await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { publishingPolicy.value = false } }
async function transition(id: number, status: string) { try { await transitionWithdrawal(id, { status, reason: security.reason, payment_reference: security.reference, totp_code: security.totp }); payoutDetails.value = undefined; await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function risk(id: number, locked: boolean) { if (!security.totp) return app.showError('TOTP required'); try { await setVoucherRiskLock(id, locked, security.reason, security.totp); await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function showPayout(id: number) { try { payoutDetails.value = await getWithdrawalPayoutDetails(id, security.totp) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function reverseRecharge(id: number) { if (!security.totp || !security.reason.trim()) return app.showError(t('finance.admin.reversalSecurityRequired')); if (!window.confirm(t('finance.admin.reverseChargebackConfirm'))) return; try { await reverseRechargeEvent(id, { reversal_type: 'CHARGEBACK', reason: security.reason.trim(), totp_code: security.totp }); await load(); app.showSuccess(t('common.success')) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
onMounted(load)
</script>
