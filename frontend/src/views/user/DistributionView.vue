<template>
  <AppLayout>
    <div class="business-page-redesign mx-auto max-w-7xl space-y-6">
      <section class="flex flex-col gap-3 border-b border-gray-200 pb-5 sm:flex-row sm:items-end sm:justify-between dark:border-dark-700">
        <div>
          <h1 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('finance.distribution.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('finance.distribution.subtitle') }}</p>
        </div>
        <div class="relative max-w-full self-start after:pointer-events-none after:absolute after:inset-y-0 after:right-0 after:w-7 after:bg-gradient-to-l after:from-white after:to-transparent sm:after:hidden dark:after:from-dark-900" role="tablist" :aria-label="t('finance.distribution.title')">
          <div class="flex max-w-[calc(100vw-2rem)] snap-x snap-mandatory overflow-x-auto border border-gray-200 bg-gray-50 p-0.5 pr-6 sm:max-w-none sm:pr-0 dark:border-dark-700 dark:bg-dark-800">
            <button v-for="tab in tabs" :key="tab.id" type="button" role="tab" :aria-selected="activeTab === tab.id" :tabindex="activeTab === tab.id ? 0 : -1" class="inline-flex min-h-10 flex-none snap-start items-center gap-1.5 px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-primary-500" :class="activeTab === tab.id ? 'bg-white font-medium text-gray-950 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500 hover:text-gray-900 dark:hover:text-white'" @click="activeTab = tab.id">
              <Icon :name="tab.icon" size="xs" />
              {{ tab.label }}
            </button>
          </div>
        </div>
      </section>

      <template v-if="dashboard">
        <section v-if="activeTab === 'overview'" class="compute-overview-hero relative overflow-hidden border p-5 text-gray-950 shadow-sm sm:p-7 dark:text-white">
          <div class="pointer-events-none absolute inset-0 opacity-20" aria-hidden="true">
            <div class="absolute -right-20 -top-24 h-64 w-64 rounded-full border border-primary-300/50"></div>
            <div class="absolute bottom-0 left-0 h-px w-full bg-primary-300"></div>
            <div class="absolute inset-y-0 left-1/2 w-px bg-primary-300"></div>
          </div>
          <div class="relative grid gap-6 lg:grid-cols-[minmax(0,1fr)_340px] lg:items-end">
            <div class="max-w-3xl">
              <div class="inline-flex items-center gap-2 border border-primary-200/70 bg-primary-50 px-2.5 py-1 text-[11px] font-semibold tracking-[0.14em] text-primary-700 dark:border-white/10 dark:bg-white/5 dark:text-primary-200">
                <Icon name="sparkles" size="xs" />
                {{ t('finance.distribution.heroKicker') }}
              </div>
              <h2 class="mt-4 text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('finance.distribution.heroTitle') }}</h2>
              <p class="mt-3 max-w-2xl text-sm leading-6 text-gray-600 sm:text-base dark:text-gray-300">{{ t('finance.distribution.heroSubtitle') }}</p>
              <div class="mt-4 flex flex-wrap gap-2">
                <span class="inline-flex items-center gap-1.5 border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-xs font-medium text-emerald-700 dark:border-emerald-300/25 dark:bg-emerald-400/10 dark:text-emerald-100">
                  <Icon name="fire" size="xs" />
                  {{ t('finance.distribution.heroBadge') }}
                </span>
                <span class="inline-flex items-center gap-1.5 border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 dark:border-white/15 dark:bg-white/5 dark:text-gray-100">
                  <Icon name="users" size="xs" />
                  {{ t('finance.distribution.inviteeBadge', { count: formatCount(dashboard.invitee_count) }) }}
                </span>
                <span v-if="previewDataEnabled" class="inline-flex items-center gap-1.5 border border-amber-200 bg-amber-50 px-3 py-1.5 text-xs font-medium text-amber-700 dark:border-amber-200/30 dark:bg-amber-300/10 dark:text-amber-100">
                  <Icon name="eye" size="xs" />
                  {{ t('finance.distribution.previewDataBadge') }}
                </span>
              </div>
            </div>
            <div class="grid gap-2 sm:grid-cols-2 lg:grid-cols-1">
              <button type="button" class="btn btn-primary inline-flex items-center justify-center gap-2 whitespace-nowrap" :disabled="!inviteLink" @click="openSharePoster">
                <Icon name="grid" size="sm" />
                {{ t('finance.distribution.createShareCard') }}
              </button>
              <button type="button" class="btn btn-secondary inline-flex items-center justify-center gap-2 whitespace-nowrap dark:!border-white/20 dark:!bg-white/5 dark:!text-white dark:hover:!bg-white/10" :disabled="!inviteLink" @click="copyText(inviteLink)">
                <Icon name="link" size="sm" />
                {{ t('finance.distribution.copyLink') }}
              </button>
            </div>
          </div>
          <div class="relative mt-6 grid gap-px overflow-hidden border border-gray-200 bg-gray-200 sm:grid-cols-2 lg:grid-cols-4 dark:border-white/10 dark:bg-white/10">
            <article v-for="stat in stats" :key="stat.label" class="min-w-0 bg-white/80 p-4 dark:bg-white/[0.06]">
              <div class="flex items-center gap-2 text-gray-500 dark:text-gray-300">
                <Icon :name="stat.icon" size="sm" class="flex-none text-primary-600 dark:text-primary-200" />
                <p class="truncate text-xs font-medium" :title="stat.label">{{ stat.label }}</p>
              </div>
              <p class="mt-2 truncate font-mono text-xl font-semibold tabular-nums text-gray-950 dark:text-white" :title="stat.value">{{ stat.value }}</p>
            </article>
          </div>
        </section>

        <p v-if="previewDataEnabled" class="border border-amber-200 bg-amber-50 px-3 py-2 text-xs leading-5 text-amber-800 dark:border-amber-800/60 dark:bg-amber-950/30 dark:text-amber-200">
          {{ t('finance.distribution.previewDataNotice') }}
        </p>

        <section v-if="activeTab === 'overview'" class="space-y-6">
          <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_380px]">
            <div class="space-y-6">
              <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
                <div class="flex items-start justify-between gap-3">
                  <div>
                    <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.growthPathTitle') }}</h2>
                    <p class="mt-1 text-sm leading-6 text-gray-500 dark:text-dark-400">{{ t('finance.distribution.growthPathHint') }}</p>
                  </div>
                  <Icon name="trendingUp" size="sm" class="text-primary-600 dark:text-primary-400" />
                </div>
                <div class="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                  <article v-for="(step, index) in growthSteps" :key="step.label" class="border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800">
                    <div class="flex items-center justify-between gap-3">
                      <span class="flex h-10 w-10 items-center justify-center border border-primary-200 bg-white text-primary-700 dark:border-primary-900/70 dark:bg-dark-900 dark:text-primary-300">
                        <Icon :name="step.icon" size="md" />
                      </span>
                      <span class="font-mono text-xs text-gray-400">0{{ index + 1 }}</span>
                    </div>
                    <h3 class="mt-3 text-sm font-semibold text-gray-900 dark:text-white">{{ step.label }}</h3>
                    <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ step.hint }}</p>
                  </article>
                </div>
              </section>

              <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.rebateRules') }}</h2>
                    <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('finance.distribution.rebateRulesHint') }}</p>
                  </div>
                  <span class="inline-flex items-center gap-1.5 border border-gray-200 bg-gray-50 px-2.5 py-1 text-xs font-medium text-gray-600 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300">
                    <Icon name="badge" size="xs" />
                    {{ t('finance.distribution.rebateRulesBadge') }}
                  </span>
                </div>
                <div class="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                  <article v-for="rule in rebateRules" :key="rule.label" class="border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-800">
                    <span class="flex h-10 w-10 items-center justify-center border border-primary-200 bg-white text-primary-700 dark:border-primary-900/70 dark:bg-dark-900 dark:text-primary-300">
                      <Icon :name="rule.icon" size="md" />
                    </span>
                    <h3 class="mt-3 text-sm font-semibold text-gray-900 dark:text-white">{{ rule.label }}</h3>
                    <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ rule.hint }}</p>
                  </article>
                </div>
              </section>
            </div>

            <aside id="referral-share-panel" class="space-y-6">
              <ReferralShareCard ref="shareCardRef" :invite-link="inviteLink" :invite-code="inviteDetail?.aff_code || ''" />
              <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.walletRules') }}</h2>
                <dl class="mt-4 grid gap-3 text-sm">
                  <div class="flex items-center justify-between gap-3">
                    <dt class="text-gray-500">{{ t('finance.distribution.freezePeriod') }}</dt>
                    <dd class="font-mono font-medium text-gray-900 dark:text-white">{{ dashboard.commission_freeze_hours }}h</dd>
                  </div>
                  <div class="flex items-center justify-between gap-3">
                    <dt class="text-gray-500">{{ t('finance.distribution.minimumWithdrawal') }}</dt>
                    <dd class="font-mono font-medium text-gray-900 dark:text-white">{{ cny(dashboard.withdrawal_min_cny_minor) }}</dd>
                  </div>
                  <div class="flex items-center justify-between gap-3">
                    <dt class="text-gray-500">{{ t('finance.distribution.dailyWithdrawalLimit') }}</dt>
                    <dd class="font-mono font-medium text-gray-900 dark:text-white">{{ dashboard.withdrawal_daily_limit }}</dd>
                  </div>
                </dl>
                <p class="mt-4 border-t border-gray-100 pt-4 text-xs leading-5 text-gray-500 dark:border-dark-700">{{ t('finance.distribution.rateSnapshot', { rate: purchaseMultiplier }) }}</p>
              </section>
              <section v-if="inviteDetail && inviteDetail.aff_quota > 0" class="border border-amber-200 bg-amber-50 p-5 dark:border-amber-900/60 dark:bg-amber-950/20">
                <h2 class="text-base font-semibold text-amber-900 dark:text-amber-100">{{ t('finance.distribution.historicalBalanceTitle') }}</h2>
                <p class="mt-2 text-xs leading-5 text-amber-800 dark:text-amber-200">{{ t('finance.distribution.historicalBalance', { amount: historicalBalance(inviteDetail.aff_quota) }) }}</p>
                <button type="button" class="btn btn-secondary btn-sm mt-3" :disabled="transferringHistory" @click="transferHistorical">{{ transferringHistory ? t('common.saving') : t('finance.distribution.transferHistorical') }}</button>
              </section>
            </aside>
          </div>

          <details class="group border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900" open>
            <summary class="flex cursor-pointer list-none items-center justify-between gap-3 text-sm font-semibold text-gray-900 marker:hidden dark:text-white">
              <span class="inline-flex items-center gap-2">
                <Icon name="document" size="sm" class="text-primary-600 dark:text-primary-400" />
                {{ t('finance.distribution.sourceDetails') }}
              </span>
              <Icon name="chevronDown" size="sm" class="transition-transform group-open:rotate-180" />
            </summary>
            <p class="mt-3 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ t('finance.distribution.sourceDetailsHint') }}</p>
            <article class="mt-4 max-w-sm border border-gray-200 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-800">
              <span class="flex h-9 w-9 items-center justify-center border border-primary-200 bg-white text-primary-700 dark:border-primary-900/60 dark:bg-dark-900 dark:text-primary-300">
                <Icon name="userPlus" size="sm" />
              </span>
              <h3 class="mt-3 text-sm font-semibold leading-5 text-gray-900 dark:text-white">{{ t('finance.distribution.companyUnits.directInvite') }}</h3>
              <dl class="mt-3 space-y-2 text-xs">
                <div class="flex items-center justify-between gap-2">
                  <dt class="text-gray-500">{{ t('finance.distribution.memberCount') }}</dt>
                  <dd class="font-mono font-medium tabular-nums text-gray-900 dark:text-white">{{ formatCount(dashboard.invitee_count) }}</dd>
                </div>
                <div class="flex items-center justify-between gap-2">
                  <dt class="text-gray-500">{{ t('finance.distribution.inviteVolume') }}</dt>
                  <dd class="font-mono font-medium tabular-nums text-gray-900 dark:text-white">{{ cny(dashboard.team_volume_cny_minor) }}</dd>
                </div>
                <div class="flex items-center justify-between gap-2">
                  <dt class="text-gray-500">{{ t('finance.distribution.lifetimeEarned') }}</dt>
                  <dd class="font-mono font-medium tabular-nums text-emerald-600 dark:text-emerald-400">{{ cny(dashboard.lifetime_earned_cny_minor) }}</dd>
                </div>
              </dl>
            </article>
          </details>
        </section>

        <section v-if="activeTab === 'team'" class="space-y-4">
          <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div>
                <div class="inline-flex items-center gap-2 text-xs font-semibold tracking-[0.12em] text-primary-600 dark:text-primary-400">
                  <Icon name="users" size="xs" />
                  {{ t('finance.distribution.teamPanelKicker') }}
                </div>
                <h2 class="mt-2 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white">{{ t('finance.distribution.teamPanelTitle') }}</h2>
                <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-500 dark:text-dark-400">{{ t('finance.distribution.teamPanelHint') }}</p>
              </div>
              <button type="button" class="btn btn-primary inline-flex items-center justify-center gap-2 whitespace-nowrap" :disabled="!inviteLink" @click="copyText(inviteLink)">
                <Icon name="link" size="sm" />
                {{ t('finance.distribution.copyLink') }}
              </button>
            </div>
            <div class="mt-5 grid gap-px overflow-hidden border border-gray-200 bg-gray-200 sm:grid-cols-2 lg:grid-cols-4 dark:border-dark-700 dark:bg-dark-700">
              <article v-for="stat in partnerStats" :key="stat.label" class="min-w-0 bg-white p-4 dark:bg-dark-900">
                <div class="flex items-center gap-2 text-gray-500 dark:text-dark-400">
                  <Icon :name="stat.icon" size="sm" class="flex-none text-primary-600 dark:text-primary-400" />
                  <p class="truncate text-xs font-medium" :title="stat.label">{{ stat.label }}</p>
                </div>
                <p class="mt-2 truncate font-mono text-xl font-semibold tabular-nums text-gray-950 dark:text-white" :title="stat.value">{{ stat.value }}</p>
              </article>
            </div>
          </section>

          <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.companyMembersTitle') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.distribution.partnerHint') }}</p>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <input v-model="search" class="input min-w-[220px] flex-1 sm:flex-none" :placeholder="t('common.search')" @keyup.enter="loadTeam" />
                <button class="btn btn-secondary btn-sm" @click="loadTeam">{{ t('common.search') }}</button>
              </div>
            </div>
            <div class="mt-4 overflow-x-auto border border-gray-200 dark:border-dark-700">
              <table class="w-full min-w-[560px] text-sm">
                <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800">
                  <tr>
                    <th class="px-4 py-3 text-left">{{ t('finance.distribution.member') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.distribution.teamVolume') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.distribution.directChildren') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="node in team" :key="node.user_id" class="border-t border-gray-100 dark:border-dark-700">
                    <td class="px-4 py-3">
                      <span class="block font-medium text-gray-900 dark:text-white">{{ node.username || node.email_masked }}</span>
                      <span class="text-xs text-gray-500">{{ node.email_masked }}</span>
                    </td>
                    <td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(node.team_volume_cny_minor) }}</td>
                    <td class="px-4 py-3 text-right font-mono tabular-nums">{{ formatCount(node.direct_children) }}</td>
                  </tr>
                  <tr v-if="team.length === 0">
                    <td colspan="3" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </section>

        <section v-if="activeTab === 'ledger'" class="space-y-6">
          <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="inline-flex items-center gap-2 text-xs font-semibold tracking-[0.12em] text-primary-600 dark:text-primary-400">
                  <Icon name="document" size="xs" />
                  {{ t('finance.distribution.earningsPanelKicker') }}
                </div>
                <h2 class="mt-2 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white">{{ t('finance.distribution.earningsPanelTitle') }}</h2>
                <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-500 dark:text-dark-400">{{ t('finance.distribution.earningsPanelHint') }}</p>
              </div>
              <Icon name="trendingUp" size="lg" class="hidden text-primary-600 dark:text-primary-400 sm:block" />
            </div>
            <div class="mt-5 grid gap-px overflow-hidden border border-gray-200 bg-gray-200 sm:grid-cols-2 lg:grid-cols-4 dark:border-dark-700 dark:bg-dark-700">
              <article v-for="stat in earningsStats" :key="stat.label" class="min-w-0 bg-white p-4 dark:bg-dark-900">
                <div class="flex items-center gap-2 text-gray-500 dark:text-dark-400">
                  <Icon :name="stat.icon" size="sm" class="flex-none text-primary-600 dark:text-primary-400" />
                  <p class="truncate text-xs font-medium" :title="stat.label">{{ stat.label }}</p>
                </div>
                <p class="mt-2 truncate font-mono text-xl font-semibold tabular-nums text-gray-950 dark:text-white" :title="stat.value">{{ stat.value }}</p>
              </article>
            </div>
          </section>

          <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.analyticsTitle') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.distribution.analyticsHint') }}</p>
              </div>
              <div class="inline-flex border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-700 dark:bg-dark-800" role="group" :aria-label="t('finance.distribution.analyticsTitle')">
                <button v-for="range in analyticsRanges" :key="`ledger-${range}`" type="button" class="px-2.5 py-1 text-xs font-medium transition-colors" :class="analyticsRange === range ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white'" @click="changeAnalyticsRange(range)">
                  {{ t(`finance.distribution.range${range === '7d' ? '7d' : range === '90d' ? '90d' : '30d'}`) }}
                </button>
              </div>
            </div>
            <ComputeCompanyTrendChart class="mt-4" :series="analytics?.series || []" :loading="analyticsLoading" />
            <div v-if="analytics" class="mt-4 grid gap-3 border-t border-gray-100 pt-4 sm:grid-cols-2 dark:border-dark-700">
              <div>
                <p class="text-xs text-gray-500">{{ t('finance.distribution.inviteSpend') }} {{ t('finance.distribution.periodComparison') }}</p>
                <p class="mt-1 font-mono text-sm font-semibold tabular-nums" :class="analytics.summary.spend_growth_percent >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">{{ signedPercent(analytics.summary.spend_growth_percent) }}</p>
              </div>
              <div>
                <p class="text-xs text-gray-500">{{ t('finance.distribution.commission') }} {{ t('finance.distribution.periodComparison') }}</p>
                <p class="mt-1 font-mono text-sm font-semibold tabular-nums" :class="analytics.summary.commission_growth_percent >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">{{ signedPercent(analytics.summary.commission_growth_percent) }}</p>
              </div>
            </div>
            <div class="mt-5">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.analyticsForecast') }}</h3>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.distribution.analyticsForecastHint') }}</p>
            </div>
            <ComputeCompanyForecastCards class="mt-3" :forecast="analytics?.forecast" />
          </section>

          <section class="border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
            <div class="flex items-center justify-between gap-3 border-b border-gray-200 p-5 dark:border-dark-700">
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.ledgerTitle') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.distribution.ledgerHint') }}</p>
              </div>
              <Icon name="document" size="sm" class="text-primary-600 dark:text-primary-400" />
            </div>
            <div class="overflow-x-auto">
              <table class="w-full min-w-[680px] text-sm">
                <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800">
                  <tr>
                    <th class="px-4 py-3 text-left">{{ t('finance.distribution.order') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.distribution.rate') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.distribution.commission') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('common.status') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('finance.distribution.createdAt') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in ledger" :key="(item.source ?? 'distribution') + '-' + item.id" class="border-t border-gray-100 dark:border-dark-700">
                    <td class="px-4 py-3">#{{ item.source_order_id }}</td>
                    <td class="px-4 py-3 text-right font-mono tabular-nums">{{ percentFromBps(item.rate_bps) }}</td>
                    <td class="px-4 py-3 text-right font-mono font-medium tabular-nums text-emerald-600 dark:text-emerald-400">{{ cny(item.amount_cny_minor) }}</td>
                    <td class="px-4 py-3">
                      <span class="inline-flex items-center border px-2 py-0.5 text-xs font-medium" :class="statusClass(item.status)">{{ commissionStatusName(item.status) }}</span>
                    </td>
                    <td class="px-4 py-3 text-gray-500">{{ formatDateTime(item.created_at) }}</td>
                  </tr>
                  <tr v-if="ledger.length === 0">
                    <td colspan="5" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </section>

        <section v-if="activeTab === 'withdraw'" class="space-y-6">
          <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="inline-flex items-center gap-2 text-xs font-semibold tracking-[0.12em] text-primary-600 dark:text-primary-400">
                  <Icon name="arrowUp" size="xs" />
                  {{ t('finance.distribution.withdrawPanelKicker') }}
                </div>
                <h2 class="mt-2 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white">{{ t('finance.distribution.withdrawPanelTitle') }}</h2>
                <p class="mt-2 max-w-2xl text-sm leading-6 text-gray-500 dark:text-dark-400">{{ t('finance.distribution.withdrawPanelHint') }}</p>
              </div>
              <Icon name="creditCard" size="lg" class="hidden text-primary-600 dark:text-primary-400 sm:block" />
            </div>
            <div class="mt-5 grid gap-px overflow-hidden border border-gray-200 bg-gray-200 sm:grid-cols-2 lg:grid-cols-4 dark:border-dark-700 dark:bg-dark-700">
              <article v-for="stat in withdrawStats" :key="stat.label" class="min-w-0 bg-white p-4 dark:bg-dark-900">
                <div class="flex items-center gap-2 text-gray-500 dark:text-dark-400">
                  <Icon :name="stat.icon" size="sm" class="flex-none text-primary-600 dark:text-primary-400" />
                  <p class="truncate text-xs font-medium" :title="stat.label">{{ stat.label }}</p>
                </div>
                <p class="mt-2 truncate font-mono text-xl font-semibold tabular-nums text-gray-950 dark:text-white" :title="stat.value">{{ stat.value }}</p>
              </article>
            </div>
          </section>

          <!-- 严格分区：人民币返点余额 与 API 平台额度（美元）是两种钱 -->
          <section class="space-y-3">
            <h3 class="text-xs font-semibold uppercase tracking-[0.12em] text-gray-400 dark:text-dark-400">{{ t('finance.distribution.balanceTypesTitle') }}</h3>
            <div class="grid gap-4 sm:grid-cols-2">
              <article class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
                <div class="flex items-center gap-2 text-primary-600 dark:text-primary-400">
                  <Icon name="dollar" size="sm" />
                  <h4 class="text-sm font-semibold">{{ t('finance.distribution.rebateBalanceLabel') }}</h4>
                </div>
                <p class="mt-3 font-mono text-xl font-semibold tabular-nums text-gray-950 dark:text-white">{{ cny(dashboard.available_cny_minor) }}</p>
                <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ t('finance.distribution.rebateBalanceDesc') }}</p>
                <RouterLink to="/shop" class="btn btn-secondary mt-4 inline-flex">{{ t('finance.distribution.rebateBalanceCta') }}</RouterLink>
              </article>
              <article class="border border-gray-200 bg-gray-50 p-5 dark:border-dark-700 dark:bg-dark-800">
                <div class="flex items-center gap-2 text-gray-500 dark:text-dark-400">
                  <Icon name="cloud" size="sm" />
                  <h4 class="text-sm font-semibold">{{ t('finance.distribution.platformQuotaLabel') }}</h4>
                </div>
                <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ t('finance.distribution.platformQuotaDesc') }}</p>
              </article>
            </div>
          </section>

          <div class="grid gap-6 lg:grid-cols-3">
            <form class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900 lg:col-span-1" @submit.prevent="saveAccount">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                <span class="inline-flex items-center gap-2">
                  <Icon name="creditCard" size="sm" class="text-primary-600 dark:text-primary-400" />
                  {{ t('finance.distribution.payout') }}
                </span>
              </h2>
              <div v-if="payout" class="mt-4 border border-gray-200 bg-gray-50 p-3 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-300">{{ payout.real_name_mask }} · {{ payout.account_mask }}</div>
              <div class="mt-4 grid gap-3">
                <label class="grid gap-1 text-xs font-medium text-gray-500">
                  {{ t('finance.distribution.realName') }}
                  <input v-model="realName" class="input" :placeholder="t('finance.distribution.realName')" />
                </label>
                <label class="grid gap-1 text-xs font-medium text-gray-500">
                  {{ t('finance.distribution.alipay') }}
                  <input v-model="alipay" class="input" :placeholder="t('finance.distribution.alipay')" />
                </label>
              </div>
              <button class="btn btn-primary mt-4">{{ t('common.save') }}</button>
            </form>

            <form class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900" @submit.prevent="withdraw">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                <span class="inline-flex items-center gap-2">
                  <Icon name="arrowUp" size="sm" class="text-primary-600 dark:text-primary-400" />
                  {{ t('finance.distribution.withdraw') }}
                </span>
              </h2>
              <p class="mt-3 text-sm text-gray-500">{{ t('common.available') }}: {{ cny(dashboard.available_cny_minor) }}</p>
              <input v-model="withdrawAmount" class="input mt-4 w-full" inputmode="decimal" :disabled="previewDataEnabled" :placeholder="(dashboard.withdrawal_min_cny_minor / 100).toFixed(2)" />
              <p class="mt-2 text-xs leading-5 text-gray-500">{{ t('finance.distribution.withdrawRuleHint', { hours: dashboard.commission_freeze_hours, minimum: cny(dashboard.withdrawal_min_cny_minor), dailyLimit: dashboard.withdrawal_daily_limit }) }}</p>
              <button class="btn btn-primary mt-4" :disabled="previewDataEnabled">{{ t('finance.distribution.submitWithdrawal') }}</button>
            </form>

            <form class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900" @submit.prevent="convertBalance">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                <span class="inline-flex items-center gap-2">
                  <Icon name="swap" size="sm" class="text-primary-600 dark:text-primary-400" />
                  {{ t('finance.distribution.convertTitle') }}
                </span>
              </h2>
              <p class="mt-3 text-sm leading-6 text-gray-500">{{ t('finance.distribution.convertRate', { rate: purchaseMultiplier }) }}</p>
              <input v-model="convertAmount" class="input mt-4 w-full" inputmode="decimal" :disabled="previewDataEnabled" :placeholder="(dashboard.withdrawal_min_cny_minor / 100).toFixed(2)" />
              <p class="mt-2 text-xs leading-5 text-gray-500">{{ t('finance.distribution.convertHint') }}</p>
              <p v-if="convertPreview" class="mt-2 font-mono text-sm font-medium text-gray-900 dark:text-white">{{ t('finance.distribution.convertPreview', { amount: convertPreview }) }}</p>
              <button class="btn btn-secondary mt-4" :disabled="previewDataEnabled || converting">{{ converting ? t('common.saving') : t('finance.distribution.convertButton') }}</button>
            </form>
          </div>

          <section class="border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
            <div class="flex items-center justify-between gap-3 border-b border-gray-200 p-5 dark:border-dark-700">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.withdrawRecords') }}</h2>
              <Icon name="clipboard" size="sm" class="text-primary-600 dark:text-primary-400" />
            </div>
            <div class="overflow-x-auto">
              <table class="w-full min-w-[760px] text-sm">
                <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800">
                  <tr>
                    <th class="px-4 py-3 text-left">ID</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.distribution.withdraw') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.vouchers.fee') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.distribution.net') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('common.status') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('finance.distribution.createdAt') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in withdrawals" :key="item.id" class="border-t border-gray-100 dark:border-dark-700">
                    <td class="px-4 py-3">#{{ item.id }}</td>
                    <td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(item.amount_cny_minor) }}</td>
                    <td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(item.fee_cny_minor) }}</td>
                    <td class="px-4 py-3 text-right font-mono font-medium tabular-nums">{{ cny(item.amount_cny_minor - item.fee_cny_minor) }}</td>
                    <td class="px-4 py-3">
                      <span class="inline-flex items-center border px-2 py-0.5 text-xs font-medium" :class="statusClass(item.status)">{{ withdrawalStatusName(item.status) }}</span>
                      <p v-if="item.reject_reason" class="mt-1 text-xs text-rose-600">{{ item.reject_reason }}</p>
                    </td>
                    <td class="px-4 py-3 text-gray-500">{{ formatDateTime(item.submitted_at) }}</td>
                  </tr>
                  <tr v-if="withdrawals.length === 0">
                    <td colspan="6" class="px-4 py-8 text-center text-gray-500">{{ t('common.noData') }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import ReferralShareCard from '@/components/user/ReferralShareCard.vue'
import ComputeCompanyForecastCards from '@/components/charts/ComputeCompanyForecastCards.vue'
import ComputeCompanyTrendChart from '@/components/charts/ComputeCompanyTrendChart.vue'
import { convertToPlatformBalance, createWithdrawal, getDistributionAnalytics, getDistributionDashboard, getDistributionLedger, getDistributionTree, getPayoutAccount, listWithdrawals, savePayoutAccount, type Commission, type DistributionAnalytics, type DistributionDashboard, type PayoutAccount, type TeamNode, type Withdrawal } from '@/api/financial'
import userAPI from '@/api/user'
import type { UserAffiliateDetail } from '@/types'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const app = useAppStore()
const activeTab = ref('overview')
const dashboard = ref<DistributionDashboard>()
const analytics = ref<DistributionAnalytics>()
const analyticsLoading = ref(false)
const analyticsRange = ref<'7d' | '30d' | '90d'>('30d')
const team = ref<TeamNode[]>([])
const ledger = ref<Commission[]>([])
const withdrawals = ref<Withdrawal[]>([])
const payout = ref<PayoutAccount>()
const search = ref('')
const realName = ref('')
const alipay = ref('')
const withdrawAmount = ref('')
const convertAmount = ref('')
const converting = ref(false)
const transferringHistory = ref(false)
const inviteDetail = ref<UserAffiliateDetail>()
const { copyToClipboard } = useClipboard()
const PUBLIC_SITE_ORIGIN = 'https://3api.shop'
const previewDataEnabled = import.meta.env.DEV && typeof window !== 'undefined' && new URLSearchParams(window.location.search).get('preview') === 'compute-company'
const tabs = computed(() => [{ id: 'overview', label: t('finance.distribution.overview'), icon: 'chartBar' as const }, { id: 'team', label: t('finance.distribution.team'), icon: 'users' as const }, { id: 'ledger', label: t('finance.distribution.ledger'), icon: 'document' as const }, { id: 'withdraw', label: t('finance.distribution.withdraw'), icon: 'swap' as const }])
const purchaseMultiplier = computed(() => dashboard.value?.balance_recharge_multiplier || '1')
const convertPreview = computed(() => {
  const amount = Number(convertAmount.value)
  const multiplier = Number(purchaseMultiplier.value)
  if (!Number.isFinite(amount) || amount <= 0 || !Number.isFinite(multiplier) || multiplier <= 0) return ''
  return (amount * multiplier).toFixed(2)
})
const inviteLink = computed(() => inviteDetail.value ? `${PUBLIC_SITE_ORIGIN}/register?aff=${encodeURIComponent(inviteDetail.value.aff_code)}` : '')
// 推广计划没有档位：返佣规则只有三条，全部面向「商品 + 钱包」。
const rebateRules = computed(() => [
  { label: t('finance.distribution.rebateRuleSource'), hint: t('finance.distribution.rebateRuleSourceHint'), icon: 'grid' as const },
  { label: t('finance.distribution.rebateRuleProduct'), hint: t('finance.distribution.rebateRuleProductHint'), icon: 'badge' as const },
  { label: t('finance.distribution.rebateRuleWallet'), hint: t('finance.distribution.rebateRuleWalletHint', { hours: dashboard.value?.commission_freeze_hours ?? 0 }), icon: 'creditCard' as const },
])
const stats = computed(() => dashboard.value ? [
  { label: t('finance.distribution.availableCommission'), value: cny(dashboard.value.available_cny_minor), icon: 'creditCard' as const },
  { label: t('finance.distribution.lifetimeEarned'), value: cny(dashboard.value.lifetime_earned_cny_minor), icon: 'dollar' as const },
  { label: t('finance.distribution.teamVolume'), value: cny(dashboard.value.team_volume_cny_minor), icon: 'trendingUp' as const },
  { label: t('finance.distribution.companyMembers'), value: formatCount(dashboard.value.invitee_count), icon: 'users' as const },
] : [])
const currentPeriodCommission = computed(() => analytics.value?.summary.commission_cny_minor || 0)
const partnerStats = computed(() => dashboard.value ? [
  { label: t('finance.distribution.companyMembers'), value: formatCount(dashboard.value.invitee_count), icon: 'users' as const },
  { label: t('finance.distribution.teamVolume'), value: cny(dashboard.value.team_volume_cny_minor), icon: 'trendingUp' as const },
  { label: t('finance.distribution.availableCommission'), value: cny(dashboard.value.available_cny_minor), icon: 'creditCard' as const },
  { label: t('finance.distribution.currentPeriodCommission'), value: cny(currentPeriodCommission.value), icon: 'badge' as const },
] : [])
const earningsStats = computed(() => dashboard.value ? [
  { label: t('finance.distribution.currentPeriodCommission'), value: cny(currentPeriodCommission.value), icon: 'chartBar' as const },
  { label: t('finance.distribution.lifetimeEarned'), value: cny(dashboard.value.lifetime_earned_cny_minor), icon: 'dollar' as const },
  { label: t('finance.distribution.availableCommission'), value: cny(dashboard.value.available_cny_minor), icon: 'creditCard' as const },
  { label: t('finance.distribution.frozenCommission'), value: cny(dashboard.value.frozen_cny_minor), icon: 'clock' as const },
] : [])
const withdrawStats = computed(() => dashboard.value ? [
  { label: t('finance.distribution.availableCommission'), value: cny(dashboard.value.available_cny_minor), icon: 'creditCard' as const },
  { label: t('finance.distribution.minimumWithdrawal'), value: cny(dashboard.value.withdrawal_min_cny_minor), icon: 'arrowUp' as const },
  { label: t('finance.distribution.freezePeriod'), value: `${dashboard.value.commission_freeze_hours}h`, icon: 'clock' as const },
  { label: t('finance.distribution.dailyWithdrawalLimit'), value: formatCount(dashboard.value.withdrawal_daily_limit), icon: 'calendar' as const },
] : [])
const growthSteps = computed(() => [
  { label: t('finance.distribution.growthPathShare'), hint: t('finance.distribution.growthPathShareHint'), icon: 'link' as const },
  { label: t('finance.distribution.growthPathRecharge'), hint: t('finance.distribution.growthPathRechargeHint'), icon: 'creditCard' as const },
  { label: t('finance.distribution.growthPathEarn'), hint: t('finance.distribution.growthPathEarnHint'), icon: 'dollar' as const },
  { label: t('finance.distribution.growthPathWithdraw'), hint: t('finance.distribution.growthPathWithdrawHint'), icon: 'arrowUp' as const },
])
const analyticsRanges = ['7d', '30d', '90d'] as const
const shareCardRef = ref<InstanceType<typeof ReferralShareCard> | null>(null)
function cny(minor: number) { return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY' }).format(minor / 100) }
function formatCount(value: number) { return new Intl.NumberFormat().format(value) }
function formatCompactNumber(value: number) {
  if (Number.isInteger(value)) return String(value)
  return value.toFixed(1).replace(/\.0$/, '')
}
function percentFromBps(bps: number) { return `${formatCompactNumber(bps / 100)}%` }
function signedPercent(value: number) { return `${value >= 0 ? '+' : ''}${value.toFixed(1)}%` }
function commissionStatusName(status: string) {
  const labels: Record<string, string> = { FROZEN: t('finance.distribution.statusFrozen'), AVAILABLE: t('finance.distribution.statusAvailable'), WITHDRAWING: t('finance.distribution.statusWithdrawing'), REVERSED: t('finance.distribution.statusReversed') }
  return labels[status.toUpperCase()] || status
}
function withdrawalStatusName(status: string) {
  const labels: Record<string, string> = { PENDING: t('finance.distribution.statusPending'), APPROVED: t('finance.distribution.statusApproved'), PAID: t('finance.distribution.statusPaid'), REJECTED: t('finance.distribution.statusRejected'), CANCELLED: t('finance.distribution.statusCancelled') }
  return labels[status.toUpperCase()] || status
}
function statusClass(status: string) {
  const key = status.toUpperCase()
  if (key === 'AVAILABLE' || key === 'PAID' || key === 'APPROVED') return 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-950/30 dark:text-emerald-300'
  if (key === 'FROZEN' || key === 'PENDING' || key === 'WITHDRAWING') return 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-300'
  if (key === 'REVERSED' || key === 'REJECTED' || key === 'CANCELLED') return 'border-rose-200 bg-rose-50 text-rose-700 dark:border-rose-900/60 dark:bg-rose-950/30 dark:text-rose-300'
  return 'border-gray-200 bg-gray-50 text-gray-600 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300'
}
function formatDateTime(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return new Intl.DateTimeFormat(undefined, { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(date)
}
function openSharePoster() {
  activeTab.value = 'overview'
  requestAnimationFrame(() => shareCardRef.value?.openPreview())
}
type DistributionPreviewModule = typeof import('./distributionPreviewData')
let previewModule: DistributionPreviewModule | undefined
async function loadPreviewModule() {
  if (!import.meta.env.DEV) return undefined
  previewModule ||= await import('./distributionPreviewData')
  return previewModule
}
async function loadAnalytics(range = analyticsRange.value) {
  analyticsLoading.value = true
  try {
    const preview = previewDataEnabled ? await loadPreviewModule() : undefined
    analytics.value = preview ? preview.createPreviewAnalytics(Number(range.slice(0, -1))) : await getDistributionAnalytics(range)
  } catch (error) {
    const preview = previewDataEnabled ? await loadPreviewModule() : undefined
    if (preview) analytics.value = preview.createPreviewAnalytics(Number(range.slice(0, -1)))
    else app.showError(extractApiErrorMessage(error))
  } finally { analyticsLoading.value = false }
}
async function changeAnalyticsRange(range: typeof analyticsRange.value) { analyticsRange.value = range; await loadAnalytics(range) }
async function load() {
  const preview = previewDataEnabled ? await loadPreviewModule() : undefined
  try { dashboard.value = preview ? preview.usePreviewDashboard(await getDistributionDashboard()) : await getDistributionDashboard() } catch (error) {
    if (!previewDataEnabled) { app.showError(extractApiErrorMessage(error)); return }
    dashboard.value = preview?.usePreviewDashboard()
  }
  const [ledgerResult, withdrawalsResult, payoutResult, inviteResult] = await Promise.allSettled([getDistributionLedger(), listWithdrawals(), getPayoutAccount(), userAPI.getAffiliateDetail()])
  if (ledgerResult.status === 'fulfilled') ledger.value = preview ? preview.previewLedger : ledgerResult.value.items
  if (withdrawalsResult.status === 'fulfilled') withdrawals.value = withdrawalsResult.value.items
  if (payoutResult.status === 'fulfilled') payout.value = payoutResult.value
  if (inviteResult.status === 'fulfilled') inviteDetail.value = inviteResult.value
  if (previewDataEnabled && !inviteDetail.value) inviteDetail.value = { user_id: 0, aff_code: 'N5TGUUDY3QZU', aff_count: 64, aff_quota: 0, aff_frozen_quota: 0, aff_history_quota: 0, effective_rebate_rate_percent: 10, invitees: [] }
  if (!previewDataEnabled) await loadTeam()
  await loadAnalytics()
}
async function loadTeam() { if (previewDataEnabled) { const preview = await loadPreviewModule(); team.value = preview?.previewTeam || []; return } try { team.value = (await getDistributionTree(undefined, search.value)).items } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function saveAccount() { try { payout.value = await savePayoutAccount(alipay.value, realName.value); alipay.value = ''; realName.value = ''; app.showSuccess(t('common.saved')) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function withdraw() { if (previewDataEnabled) { app.showInfo(t('finance.distribution.previewDataReadOnly')); return } try { const minor = Math.round(Number(withdrawAmount.value) * 100); await createWithdrawal(minor); withdrawAmount.value = ''; dashboard.value = await getDistributionDashboard(); withdrawals.value = (await listWithdrawals()).items; app.showSuccess(t('common.success')) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
function historicalBalance(value: number) { return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD' }).format(value) }
async function copyText(value: string) { if (value) await copyToClipboard(value, t('finance.distribution.copied')) }
async function transferHistorical() { if (!inviteDetail.value || inviteDetail.value.aff_quota <= 0 || transferringHistory.value) return; transferringHistory.value = true; try { await userAPI.transferAffiliateQuota(); inviteDetail.value = await userAPI.getAffiliateDetail(); app.showSuccess(t('common.success')) } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { transferringHistory.value = false } }
function createIdempotencyKey() { return typeof crypto !== 'undefined' && 'randomUUID' in crypto ? crypto.randomUUID() : `distribution-convert-${Date.now()}-${Math.random().toString(36).slice(2)}` }
async function convertBalance() { if (previewDataEnabled) { app.showInfo(t('finance.distribution.previewDataReadOnly')); return } if (converting.value) return; converting.value = true; try { const minor = Math.round(Number(convertAmount.value) * 100); const result = await convertToPlatformBalance(minor, createIdempotencyKey()); convertAmount.value = ''; dashboard.value = await getDistributionDashboard(); app.showSuccess(t('finance.distribution.converted', { amount: result.usd_amount })) } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { converting.value = false } }
onMounted(load)
</script>
