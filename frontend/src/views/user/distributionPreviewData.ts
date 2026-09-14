import type { Commission, DistributionAnalytics, DistributionDashboard, TeamNode } from '@/api/financial'

export function createPreviewAnalytics(days: number): DistributionAnalytics {
  const today = new Date()
  const series = Array.from({ length: days }, (_, index) => {
    const date = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), today.getUTCDate() - days + index + 1))
    return {
      date: date.toISOString().slice(0, 10),
      spend_cny_minor: 22000 + index * 7200 + (index % 5) * 900,
      commission_cny_minor: 1800 + index * 520 + (index % 4) * 120,
    }
  })
  const recharge = series.reduce((total, point) => total + point.spend_cny_minor, 0)
  const commission = series.reduce((total, point) => total + point.commission_cny_minor, 0)
  return {
    as_of: new Date().toISOString(),
    range_days: days,
    series,
    summary: {
      spend_cny_minor: recharge,
      commission_cny_minor: commission,
      previous_spend_cny_minor: Math.round(recharge * 0.84),
      previous_commission_cny_minor: Math.round(commission * 0.88),
      spend_growth_percent: 19.1,
      commission_growth_percent: 13.6,
    },
    forecast: {
      method: 'weighted_recent_trend_preview',
      seven_days: { eligible: true, estimated_spend_cny_minor: 472000, estimated_commission_cny_minor: 42600, spend_growth_percent: 8.7, commission_growth_percent: 6.4 },
      thirty_days: { eligible: true, estimated_spend_cny_minor: 2240000, estimated_commission_cny_minor: 211000, spend_growth_percent: 11.2, commission_growth_percent: 9.5 },
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
  // 推广计划无档位：返佣比例跟着商品走，概览只汇总「人数 + 业绩 + 钱包」。
  invitee_count: 24,
  team_volume_cny_minor: 6350000,
  available_cny_minor: 164600,
  frozen_cny_minor: 29388,
  withdrawing_cny_minor: 12000,
  debt_cny_minor: 0,
  lifetime_earned_cny_minor: 229988,
}

export function usePreviewDashboard(base?: DistributionDashboard): DistributionDashboard {
  return {
    ...(base || previewDashboard),
    ...previewDashboard,
  }
}

// 单层推广计划：这里只展示直接邀请的好友（与后端 Tree 行为一致）。
// 只有人数与业绩，没有档位。
export const previewTeam: TeamNode[] = [
  { user_id: 101, parent_user_id: 0, email_masked: 'l***@example.net', username: '林川', direct_children: 8, team_volume_cny_minor: 6240000 },
  { user_id: 102, parent_user_id: 0, email_masked: 'm***@example.net', username: '周宁', direct_children: 6, team_volume_cny_minor: 2850000 },
  { user_id: 103, parent_user_id: 0, email_masked: 'q***@example.net', username: '秦越', direct_children: 4, team_volume_cny_minor: 1280000 },
  { user_id: 104, parent_user_id: 0, email_masked: 's***@example.net', username: '沈墨', direct_children: 3, team_volume_cny_minor: 640000 },
  { user_id: 105, parent_user_id: 0, email_masked: 'y***@example.net', username: '叶舟', direct_children: 1, team_volume_cny_minor: 210000 },
]

export const previewLedger: Commission[] = Array.from({ length: 8 }, (_, index) => ({
  id: 7000 + index,
  source_order_id: 93000 + index,
  source_user_id: 101 + (index % 5),
  depth: 1,
  // 历史字段：档位体系已下线，新返佣记录不再有档位，这里固定 0。
  tier: 0,
  rate_bps: [500, 800, 1000, 800, 1000][index % 5],
  base_cny_minor: 28000 + index * 6300,
  amount_cny_minor: 5600 + index * 820,
  team_volume_cny_minor: 920000 + index * 74000,
  status: index < 2 ? 'FROZEN' : 'AVAILABLE',
  frozen_until: new Date(Date.now() + 72 * 60 * 60 * 1000).toISOString(),
  created_at: new Date(Date.now() - index * 36 * 60 * 60 * 1000).toISOString(),
}))
