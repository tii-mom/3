import type { Commission, DistributionAnalytics, DistributionDashboard, TeamNode } from '@/api/financial'

export function createPreviewAnalytics(days: number): DistributionAnalytics {
  const today = new Date()
  const series = Array.from({ length: days }, (_, index) => {
    const date = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), today.getUTCDate() - days + index + 1))
    return {
      date: date.toISOString().slice(0, 10),
      recharge_cny_minor: 22000 + index * 7200 + (index % 5) * 900,
      commission_cny_minor: 1800 + index * 520 + (index % 4) * 120,
    }
  })
  const recharge = series.reduce((total, point) => total + point.recharge_cny_minor, 0)
  const commission = series.reduce((total, point) => total + point.commission_cny_minor, 0)
  return {
    as_of: new Date().toISOString(),
    range_days: days,
    series,
    summary: {
      recharge_cny_minor: recharge,
      commission_cny_minor: commission,
      previous_recharge_cny_minor: Math.round(recharge * 0.84),
      previous_commission_cny_minor: Math.round(commission * 0.88),
      recharge_growth_percent: 19.1,
      commission_growth_percent: 13.6,
    },
    forecast: {
      method: 'weighted_recent_trend_preview',
      seven_days: { eligible: true, estimated_recharge_cny_minor: 472000, estimated_commission_cny_minor: 42600, recharge_growth_percent: 8.7, commission_growth_percent: 6.4 },
      thirty_days: { eligible: true, estimated_recharge_cny_minor: 2240000, estimated_commission_cny_minor: 211000, recharge_growth_percent: 11.2, commission_growth_percent: 9.5 },
    },
  }
}

export const previewDashboard: DistributionDashboard = {
  enabled: true,
  balance_recharge_multiplier: '0.14',
  usd_to_cny_rate: '7.14',
  commission_freeze_hours: 168,
  withdrawal_min_cny_minor: 2000,
  withdrawal_daily_limit: 1,
  team_volume_cny_minor: 1545200,
  current_tier: 3,
  auto_tier: 3,
  next_threshold_cny_minor: 0,
  level_counts: { 1: 24, 2: 18, 3: 11, 4: 7, 5: 4 },
  levels: [
    { depth: 1, member_count: 24, recharge_cny_minor: 680000, commission_cny_minor: 136000, available_cny_minor: 118000, frozen_cny_minor: 18000 },
    { depth: 2, member_count: 18, recharge_cny_minor: 420000, commission_cny_minor: 36400, available_cny_minor: 29400, frozen_cny_minor: 7000 },
    { depth: 3, member_count: 11, recharge_cny_minor: 257800, commission_cny_minor: 15468, available_cny_minor: 12000, frozen_cny_minor: 3468 },
    { depth: 4, member_count: 7, recharge_cny_minor: 118600, commission_cny_minor: 4744, available_cny_minor: 4000, frozen_cny_minor: 744 },
    { depth: 5, member_count: 4, recharge_cny_minor: 68800, commission_cny_minor: 1376, available_cny_minor: 1200, frozen_cny_minor: 176 },
  ],
  available_cny_minor: 164600,
  frozen_cny_minor: 29388,
  withdrawing_cny_minor: 12000,
  debt_cny_minor: 0,
  lifetime_earned_cny_minor: 229988,
  tiers: [
    { tier: 0, threshold_cny_minor: 0, rates_bps: [1000, 0, 0, 0, 0] },
    { tier: 1, threshold_cny_minor: 100000, rates_bps: [1000, 400, 300, 200, 100] },
    { tier: 2, threshold_cny_minor: 1000000, rates_bps: [1500, 600, 400, 300, 200] },
    { tier: 3, threshold_cny_minor: 10000000, rates_bps: [2000, 800, 600, 400, 200] },
  ],
}

export function usePreviewDashboard(base?: DistributionDashboard): DistributionDashboard {
  return {
    ...(base || previewDashboard),
    ...previewDashboard,
    tiers: base?.tiers?.length ? base.tiers : previewDashboard.tiers,
    levels: previewDashboard.levels,
  }
}

export const previewTeam: TeamNode[] = [
  { user_id: 101, parent_user_id: 0, email_masked: 'l***@example.net', username: '林川', direct_children: 8, team_volume_cny_minor: 326000, current_tier: 2, auto_tier: 2, effective_tier: 2 },
  { user_id: 102, parent_user_id: 0, email_masked: 'm***@example.net', username: '周宁', direct_children: 6, team_volume_cny_minor: 214800, current_tier: 2, auto_tier: 2, effective_tier: 2 },
  { user_id: 103, parent_user_id: 0, email_masked: 'q***@example.net', username: '秦越', direct_children: 4, team_volume_cny_minor: 168200, current_tier: 1, auto_tier: 1, effective_tier: 1 },
  { user_id: 104, parent_user_id: 0, email_masked: 's***@example.net', username: '沈墨', direct_children: 3, team_volume_cny_minor: 92800, current_tier: 1, auto_tier: 1, effective_tier: 1 },
  { user_id: 105, parent_user_id: 0, email_masked: 'y***@example.net', username: '叶舟', direct_children: 1, team_volume_cny_minor: 41600, current_tier: 0, auto_tier: 0, effective_tier: 0 },
]

export const previewLedger: Commission[] = Array.from({ length: 8 }, (_, index) => ({
  id: 7000 + index,
  source_order_id: 93000 + index,
  source_user_id: 101 + (index % 5),
  depth: (index % 5) + 1,
  tier: Math.min(3, 1 + Math.floor(index / 3)),
  rate_bps: [2000, 800, 600, 400, 200][index % 5],
  base_cny_minor: 28000 + index * 6300,
  amount_cny_minor: 5600 + index * 820,
  team_volume_cny_minor: 920000 + index * 74000,
  status: index < 2 ? 'FROZEN' : 'AVAILABLE',
  frozen_until: new Date(Date.now() + 72 * 60 * 60 * 1000).toISOString(),
  created_at: new Date(Date.now() - index * 36 * 60 * 60 * 1000).toISOString(),
}))
