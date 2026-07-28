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
                  {{ t('finance.distribution.heroBadge', { rate: maxRatePercent }) }}
                </span>
                <span class="inline-flex items-center gap-1.5 border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 dark:border-white/15 dark:bg-white/5 dark:text-gray-100">
                  <Icon name="badge" size="xs" />
                  {{ t('finance.distribution.currentBenefit') }} · {{ tierDisplay(effectiveTier, tierThreshold(effectiveTier)) }} · {{ currentRatePercent }}
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
                    <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.tiers') }}</h2>
                    <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('finance.distribution.tierStatus', { tier: effectiveTier }) }}</p>
                  </div>
                  <span v-if="dashboard.tier_override !== undefined" class="inline-flex items-center gap-1.5 border border-amber-200 bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700 dark:border-amber-800/60 dark:bg-amber-950/30 dark:text-amber-200">
                    <Icon name="badge" size="xs" />
                    {{ t('finance.distribution.manualTier') }} · {{ tierDisplay(dashboard.tier_override, tierThreshold(dashboard.tier_override)) }}
                  </span>
                </div>
                <div class="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                  <article v-for="tier in tierCards" :key="tier.tier" class="relative overflow-hidden border p-4 transition-colors" :class="tier.tier === effectiveTier ? 'border-primary-400 bg-primary-50 dark:border-primary-700 dark:bg-primary-950/20' : 'border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900'">
                    <div class="flex items-start justify-between gap-3">
                      <div>
                        <p class="font-mono text-lg font-semibold text-gray-950 dark:text-white">{{ tierDisplay(tier.tier, tier.threshold_cny_minor) }}</p>
                        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ tier.range }}</p>
                      </div>
                      <span v-if="tier.tier === effectiveTier" class="inline-flex items-center border border-primary-200 bg-white px-2 py-0.5 text-xs font-medium text-primary-700 dark:border-primary-800 dark:bg-dark-800 dark:text-primary-300">{{ t('finance.distribution.current') }}</span>
                    </div>
                    <p class="mt-5 font-mono text-2xl font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">{{ tier.totalRatePercent }}</p>
                    <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.distribution.tierRateLabel') }}</p>
                  </article>
                </div>
              </section>
            </div>

            <aside id="compute-company-share-panel" class="space-y-6">
              <ComputeCompanyShareCard ref="shareCardRef" :invite-link="inviteLink" :invite-code="inviteDetail?.aff_code || ''" />
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

          <details class="group border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
            <summary class="flex cursor-pointer list-none items-center justify-between gap-3 text-sm font-semibold text-gray-900 marker:hidden dark:text-white">
              <span class="inline-flex items-center gap-2">
                <Icon name="document" size="sm" class="text-primary-600 dark:text-primary-400" />
                {{ t('finance.distribution.sourceDetails') }}
              </span>
              <Icon name="chevronDown" size="sm" class="transition-transform group-open:rotate-180" />
            </summary>
            <p class="mt-3 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ t('finance.distribution.sourceDetailsHint') }}</p>
            <div class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
              <article v-for="summary in companySummaries" :key="`source-card-${summary.depth}`" class="min-w-0 border border-gray-200 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-800">
                <span class="flex h-9 w-9 items-center justify-center border border-primary-200 bg-white text-primary-700 dark:border-primary-900/60 dark:bg-dark-900 dark:text-primary-300">
                  <Icon :name="companyUnitIcons[summary.depth - 1]" size="sm" />
                </span>
                <h3 class="mt-3 text-sm font-semibold leading-5 text-gray-900 dark:text-white">{{ companyUnitName(summary.depth) }}</h3>
                <dl class="mt-3 space-y-2 text-xs">
                  <div class="flex items-center justify-between gap-2">
                    <dt class="text-gray-500">{{ t('finance.distribution.memberCount') }}</dt>
                    <dd class="font-mono font-medium tabular-nums text-gray-900 dark:text-white">{{ formatCount(summary.member_count) }}</dd>
                  </div>
                  <div class="flex items-center justify-between gap-2">
                    <dt class="text-gray-500">{{ t('finance.distribution.recharge') }}</dt>
                    <dd class="font-mono font-medium tabular-nums text-gray-900 dark:text-white">{{ cny(summary.recharge_cny_minor) }}</dd>
                  </div>
                  <div class="flex items-center justify-between gap-2">
                    <dt class="text-gray-500">{{ t('finance.distribution.commission') }}</dt>
                    <dd class="font-mono font-medium tabular-nums text-emerald-600 dark:text-emerald-400">{{ cny(summary.commission_cny_minor) }}</dd>
                  </div>
                </dl>
              </article>
            </div>
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
          <div class="grid gap-6 xl:grid-cols-[360px_minmax(0,1fr)]">
            <ComputeCompanyTeamChart :segments="teamSegments" :total="teamChartTotal" @select="selectTeamSegment" />
            <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.companyMembersTitle') }}</h2>
                  <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.distribution.partnerHint') }}</p>
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <input v-model="search" class="input min-w-[220px] flex-1 sm:flex-none" :placeholder="t('common.search')" @keyup.enter="() => loadTeam(teamParent)" />
                  <button v-if="teamParent" class="btn btn-secondary btn-sm" @click="loadTeam()">{{ t('finance.distribution.backToTeam') }}</button>
                  <button class="btn btn-secondary btn-sm" @click="loadTeam(teamParent)">{{ t('common.search') }}</button>
                </div>
              </div>
              <div class="mt-4 overflow-x-auto border border-gray-200 dark:border-dark-700">
                <table class="w-full min-w-[820px] text-sm">
                  <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800">
                    <tr>
                      <th class="px-4 py-3 text-left">{{ t('finance.distribution.member') }}</th>
                      <th class="px-4 py-3 text-right">{{ t('finance.distribution.directChildren') }}</th>
                      <th class="px-4 py-3 text-right">{{ t('finance.distribution.teamVolume') }}</th>
                      <th class="px-4 py-3 text-right">{{ t('finance.distribution.autoTier') }}</th>
                      <th class="px-4 py-3 text-right">{{ t('finance.distribution.effectiveTier') }}</th>
                      <th class="px-4 py-3 text-left">{{ t('common.actions') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="node in team" :key="node.user_id" class="border-t border-gray-100 dark:border-dark-700">
                      <td class="px-4 py-3">
                        <span class="block font-medium text-gray-900 dark:text-white">{{ node.username || node.email_masked }}</span>
                        <span class="text-xs text-gray-500">{{ node.email_masked }}</span>
                      </td>
                      <td class="px-4 py-3 text-right font-mono tabular-nums">{{ node.direct_children }}</td>
                      <td class="px-4 py-3 text-right font-mono tabular-nums">{{ cny(node.team_volume_cny_minor) }}</td>
                      <td class="px-4 py-3 text-right">{{ tierDisplay(node.auto_tier, tierThreshold(node.auto_tier)) }}</td>
                      <td class="px-4 py-3 text-right">
                        {{ tierDisplay(node.effective_tier, tierThreshold(node.effective_tier)) }}
                        <span v-if="node.tier_override !== undefined" class="ml-1 text-xs text-amber-600">{{ t('finance.distribution.manual') }}</span>
                      </td>
                      <td class="px-4 py-3">
                        <button v-if="node.direct_children" class="font-medium text-gray-700 underline decoration-gray-300 underline-offset-4 dark:text-gray-200" @click="expandNode(node)">{{ t('finance.distribution.viewTeam') }}</button>
                        <span v-else class="text-gray-400">-</span>
                      </td>
                    </tr>
                    <tr v-if="team.length === 0">
                      <td colspan="6" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>
          </div>
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
                <p class="text-xs text-gray-500">{{ t('finance.distribution.recharge') }} {{ t('finance.distribution.periodComparison') }}</p>
                <p class="mt-1 font-mono text-sm font-semibold tabular-nums" :class="analytics.summary.recharge_growth_percent >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">{{ signedPercent(analytics.summary.recharge_growth_percent) }}</p>
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
              <table class="w-full min-w-[860px] text-sm">
                <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800">
                  <tr>
                    <th class="px-4 py-3 text-left">{{ t('finance.distribution.order') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('finance.distribution.companyUnit') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.distribution.rate') }}</th>
                    <th class="px-4 py-3 text-right">{{ t('finance.distribution.commission') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('common.status') }}</th>
                    <th class="px-4 py-3 text-left">{{ t('finance.distribution.createdAt') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in ledger" :key="item.id" class="border-t border-gray-100 dark:border-dark-700">
                    <td class="px-4 py-3">#{{ item.source_order_id }}</td>
                    <td class="px-4 py-3">{{ companyUnitName(item.depth) }}</td>
                    <td class="px-4 py-3 text-right font-mono tabular-nums">{{ percentFromBps(item.rate_bps) }}</td>
                    <td class="px-4 py-3 text-right font-mono font-medium tabular-nums text-emerald-600 dark:text-emerald-400">{{ cny(item.amount_cny_minor) }}</td>
                    <td class="px-4 py-3">
                      <span class="inline-flex items-center border px-2 py-0.5 text-xs font-medium" :class="statusClass(item.status)">{{ commissionStatusName(item.status) }}</span>
                    </td>
                    <td class="px-4 py-3 text-gray-500">{{ formatDateTime(item.created_at) }}</td>
                  </tr>
                  <tr v-if="ledger.length === 0">
                    <td colspan="6" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td>
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
import ComputeCompanyShareCard from '@/components/user/ComputeCompanyShareCard.vue'
import ComputeCompanyForecastCards from '@/components/charts/ComputeCompanyForecastCards.vue'
import ComputeCompanyTeamChart from '@/components/charts/ComputeCompanyTeamChart.vue'
import ComputeCompanyTrendChart from '@/components/charts/ComputeCompanyTrendChart.vue'
import { convertToPlatformBalance, createWithdrawal, getDistributionAnalytics, getDistributionDashboard, getDistributionLedger, getDistributionTree, getPayoutAccount, listWithdrawals, savePayoutAccount, type Commission, type DistributionAnalytics, type DistributionDashboard, type PayoutAccount, type TeamNode, type Withdrawal } from '@/api/financial'
import userAPI from '@/api/user'
import type { UserAffiliateDetail } from '@/types'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { COMPUTE_COMPANY_UNIT_KEYS } from '@/constants/distribution'

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
const teamParent = ref<number>()
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
const companyUnitKeys = COMPUTE_COMPANY_UNIT_KEYS
const companyUnitIcons = ['server', 'link', 'terminal', 'globe', 'shield'] as const
const companyUnits = computed(() => companyUnitKeys.map(key => t(`finance.distribution.companyUnits.${key}`)))
const tabs = computed(() => [{ id: 'overview', label: t('finance.distribution.overview'), icon: 'chartBar' as const }, { id: 'team', label: t('finance.distribution.team'), icon: 'users' as const }, { id: 'ledger', label: t('finance.distribution.ledger'), icon: 'document' as const }, { id: 'withdraw', label: t('finance.distribution.withdraw'), icon: 'swap' as const }])
const companySummaries = computed(() => dashboard.value?.levels?.length ? dashboard.value.levels : Array.from({ length: companyUnitKeys.length }, (_, index) => ({ depth: index + 1, member_count: dashboard.value?.level_counts[index + 1] || 0, recharge_cny_minor: 0, commission_cny_minor: 0, available_cny_minor: 0, frozen_cny_minor: 0 })))
const effectiveTier = computed(() => dashboard.value?.current_tier || 0)
const purchaseMultiplier = computed(() => dashboard.value?.balance_recharge_multiplier || '1')
const convertPreview = computed(() => {
  const amount = Number(convertAmount.value)
  const multiplier = Number(purchaseMultiplier.value)
  if (!Number.isFinite(amount) || amount <= 0 || !Number.isFinite(multiplier) || multiplier <= 0) return ''
  return (amount * multiplier).toFixed(2)
})
const inviteLink = computed(() => inviteDetail.value ? `${PUBLIC_SITE_ORIGIN}/register?aff=${encodeURIComponent(inviteDetail.value.aff_code)}` : '')
const companyMemberCount = computed(() => companySummaries.value.reduce((total, summary) => total + summary.member_count, 0))
const tierCards = computed(() => {
  const tiers = [...(dashboard.value?.tiers || [])].sort((left, right) => left.tier - right.tier)
  return tiers.map((tier, index) => {
    const totalRateBps = tier.rates_bps.reduce((total, rate) => total + rate, 0)
    return {
      ...tier,
      range: tierRange(tier.threshold_cny_minor, tiers[index + 1]?.threshold_cny_minor),
      totalRateBps,
      totalRatePercent: percentFromBps(totalRateBps),
    }
  })
})
const maxRatePercent = computed(() => percentFromBps(Math.max(0, ...tierCards.value.map(tier => tier.totalRateBps))))
const currentRatePercent = computed(() => tierCards.value.find(tier => tier.tier === effectiveTier.value)?.totalRatePercent || '0%')
const stats = computed(() => dashboard.value ? [
  { label: t('finance.distribution.availableCommission'), value: cny(dashboard.value.available_cny_minor), icon: 'creditCard' as const },
  { label: t('finance.distribution.lifetimeEarned'), value: cny(dashboard.value.lifetime_earned_cny_minor), icon: 'dollar' as const },
  { label: t('finance.distribution.teamVolume'), value: cny(dashboard.value.team_volume_cny_minor), icon: 'trendingUp' as const },
  { label: t('finance.distribution.companyMembers'), value: formatCount(companyMemberCount.value), icon: 'users' as const },
] : [])
const directPartnerCount = computed(() => Number(dashboard.value?.level_counts[1] || companySummaries.value[0]?.member_count || 0))
const currentPeriodCommission = computed(() => analytics.value?.summary.commission_cny_minor || 0)
const partnerStats = computed(() => dashboard.value ? [
  { label: t('finance.distribution.companyMembers'), value: formatCount(companyMemberCount.value), icon: 'users' as const },
  { label: t('finance.distribution.directChildren'), value: formatCount(directPartnerCount.value), icon: 'userPlus' as const },
  { label: t('finance.distribution.teamVolume'), value: cny(dashboard.value.team_volume_cny_minor), icon: 'trendingUp' as const },
  { label: t('finance.distribution.effectiveTier'), value: `${tierDisplay(effectiveTier.value, tierThreshold(effectiveTier.value))} · ${currentRatePercent.value}`, icon: 'badge' as const },
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
const teamSegments = computed(() => {
  const labels = [
    t('finance.distribution.teamSegmentDirect'),
    t('finance.distribution.teamSegmentExpanded'),
    t('finance.distribution.teamSegmentCollaboration'),
    t('finance.distribution.teamSegmentEcosystem'),
    t('finance.distribution.teamSegmentSupport'),
  ]
  const counts = labels.map((_, index) => Number(dashboard.value?.level_counts[index + 1] || companySummaries.value[index]?.member_count || 0))
  const max = Math.max(...counts, 1)
  return labels.map((label, index) => ({ key: `segment-${index + 1}`, label, count: counts[index], percent: (counts[index] / max) * 100 }))
})
const teamChartTotal = computed(() => teamSegments.value.reduce((total, segment) => total + segment.count, 0))
const analyticsRanges = ['7d', '30d', '90d'] as const
const shareCardRef = ref<InstanceType<typeof ComputeCompanyShareCard> | null>(null)
function cny(minor: number) { return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY' }).format(minor / 100) }
function formatCount(value: number) { return new Intl.NumberFormat().format(value) }
function formatCompactNumber(value: number) {
  if (Number.isInteger(value)) return String(value)
  return value.toFixed(1).replace(/\.0$/, '')
}
function compactAmount(minor: number) {
  const amount = minor / 100
  if (amount >= 1_000_000) return `${formatCompactNumber(amount / 1_000_000)}M`
  if (amount >= 1_000) return `${formatCompactNumber(amount / 1_000)}K`
  return `${amount}`
}
function tierRange(currentMinor: number, nextMinor?: number) { return nextMinor && nextMinor > currentMinor ? `${compactAmount(currentMinor)}–<${compactAmount(nextMinor)}` : `${compactAmount(currentMinor)} ${t('finance.distribution.andAbove')}` }
function percentFromBps(bps: number) { return `${formatCompactNumber(bps / 100)}%` }
function tierThreshold(tier: number) { return dashboard.value?.tiers.find(candidate => candidate.tier === tier)?.threshold_cny_minor || 0 }
function tierDisplay(tier: number, _thresholdMinor: number) { return `T${tier}` }
function companyUnitName(depth: number) { return companyUnits.value[depth - 1] || `${t('finance.distribution.companyUnit')} ${depth}` }
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
async function selectTeamSegment() { activeTab.value = 'team'; await loadTeam() }
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
async function loadTeam(parent?: number) { if (previewDataEnabled) { const preview = await loadPreviewModule(); teamParent.value = parent; team.value = parent ? [] : (preview?.previewTeam || []); return } try { teamParent.value = parent; team.value = (await getDistributionTree(parent, search.value)).items } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function expandNode(node: TeamNode) { if (node.direct_children) await loadTeam(node.user_id) }
async function saveAccount() { try { payout.value = await savePayoutAccount(alipay.value, realName.value); alipay.value = ''; realName.value = ''; app.showSuccess(t('common.saved')) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function withdraw() { if (previewDataEnabled) { app.showInfo(t('finance.distribution.previewDataReadOnly')); return } try { const minor = Math.round(Number(withdrawAmount.value) * 100); await createWithdrawal(minor); withdrawAmount.value = ''; dashboard.value = await getDistributionDashboard(); withdrawals.value = (await listWithdrawals()).items; app.showSuccess(t('common.success')) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
function historicalBalance(value: number) { return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD' }).format(value) }
async function copyText(value: string) { if (value) await copyToClipboard(value, t('finance.distribution.copied')) }
async function transferHistorical() { if (!inviteDetail.value || inviteDetail.value.aff_quota <= 0 || transferringHistory.value) return; transferringHistory.value = true; try { await userAPI.transferAffiliateQuota(); inviteDetail.value = await userAPI.getAffiliateDetail(); app.showSuccess(t('common.success')) } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { transferringHistory.value = false } }
function createIdempotencyKey() { return typeof crypto !== 'undefined' && 'randomUUID' in crypto ? crypto.randomUUID() : `distribution-convert-${Date.now()}-${Math.random().toString(36).slice(2)}` }
async function convertBalance() { if (previewDataEnabled) { app.showInfo(t('finance.distribution.previewDataReadOnly')); return } if (converting.value) return; converting.value = true; try { const minor = Math.round(Number(convertAmount.value) * 100); const result = await convertToPlatformBalance(minor, createIdempotencyKey()); convertAmount.value = ''; dashboard.value = await getDistributionDashboard(); app.showSuccess(t('finance.distribution.converted', { amount: result.usd_amount })) } catch (error) { app.showError(extractApiErrorMessage(error)) } finally { converting.value = false } }
onMounted(load)
</script>
