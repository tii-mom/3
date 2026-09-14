import { apiClient } from './client'

export interface Paginated<T> { items: T[]; total: number; page: number; page_size: number; pages: number }
export interface Voucher { id: number; issuer_user_id: number; redeemer_user_id?: number; code_last4: string; face_value: string; fee_amount: string; fee_rate_bps: number; status: string; expires_at: string; created_at: string; code?: string }
export interface VoucherAvailability {
  enabled: boolean
  credit_buckets_enforced: boolean
  transferable_credit: string
  non_transferable_credit: string
  debt: string
  fee_bps: number
  minimum_usd: string
  maximum_usd: string
  daily_maximum_usd: string
  daily_used_usd: string
  daily_remaining_usd: string
  daily_count: number
  daily_used_count: number
  daily_remaining_count: number
  expiry_days: number
  step_up_minimum_usd: string
  maximum_face_value_usd: string
}
// 推广计划没有档位：返佣比例来自每个商品上架时的设置（shop_products.commission_bps），
// 下单时快照进订单。`team_volume_cny_minor` 保留为「直属邀请成员在商城的实付总额」。
export interface DistributionDashboard { enabled: boolean; balance_recharge_multiplier: string; usd_to_cny_rate: string; commission_freeze_hours: number; withdrawal_min_cny_minor: number; withdrawal_daily_limit: number; invitee_count: number; team_volume_cny_minor: number; available_cny_minor: number; frozen_cny_minor: number; withdrawing_cny_minor: number; debt_cny_minor: number; lifetime_earned_cny_minor: number }
export interface TeamNode { user_id: number; parent_user_id: number; email_masked: string; username: string; direct_children: number; team_volume_cny_minor: number }
// source 区分佣金来源：'shop' = 商城订单返点（当前唯一在写入的），'distribution' = 历史充值返佣。
// 两张表自增 id 会重号，列表 key 必须用 `${source}-${id}`。
export interface Commission { id: number; source?: 'shop' | 'distribution'; source_order_id: number; source_user_id: number; depth: number; tier: number; rate_bps: number; base_cny_minor: number; amount_cny_minor: number; team_volume_cny_minor: number; status: string; frozen_until: string; created_at: string }
export interface PayoutAccount { account_type: string; account_mask: string; real_name_mask: string }
export interface Withdrawal { id: number; amount_cny_minor: number; fee_cny_minor: number; fee_rate_bps: number; config_version: number; status: string; reject_reason?: string; payment_reference?: string; submitted_at: string }
export interface DistributionConversion { id: number; amount_cny_minor: number; usd_amount: string; cny_to_usd_rate?: string; rate_source?: string; usd_to_cny_rate: string; config_version: number; created_at: string }
export interface DistributionAnalyticsPoint { date: string; spend_cny_minor: number; commission_cny_minor: number }
export interface DistributionAnalyticsSummary { spend_cny_minor: number; commission_cny_minor: number; previous_spend_cny_minor: number; previous_commission_cny_minor: number; spend_growth_percent: number; commission_growth_percent: number }
export interface DistributionForecastHorizon { eligible: boolean; reason?: 'insufficient_history' | 'insufficient_activity'; estimated_spend_cny_minor: number; estimated_commission_cny_minor: number; spend_growth_percent: number; commission_growth_percent: number }
export interface DistributionAnalytics { as_of: string; range_days: number; series: DistributionAnalyticsPoint[]; summary: DistributionAnalyticsSummary; forecast: { method: string; seven_days: DistributionForecastHorizon; thirty_days: DistributionForecastHorizon } }

export async function createVoucher(amount: string, totpCode: string, idempotencyKey: string): Promise<Voucher> {
  return (await apiClient.post<Voucher>('/user/vouchers', { amount, totp_code: totpCode }, {
    headers: { 'Idempotency-Key': idempotencyKey },
  })).data
}
export async function listVouchers(page = 1): Promise<Paginated<Voucher>> { return (await apiClient.get<Paginated<Voucher>>('/user/vouchers', { params: { page } })).data }
export async function cancelVoucher(id: number): Promise<Voucher> { return (await apiClient.post<Voucher>(`/user/vouchers/${id}/cancel`)).data }
export async function getVoucherAvailability(): Promise<VoucherAvailability> { return (await apiClient.get<VoucherAvailability>('/user/vouchers/availability')).data }
export async function getDistributionDashboard(): Promise<DistributionDashboard> { return (await apiClient.get<DistributionDashboard>('/distribution/dashboard')).data }
export async function getDistributionAnalytics(range: '7d' | '30d' | '90d' = '30d'): Promise<DistributionAnalytics> { return (await apiClient.get<DistributionAnalytics>('/distribution/analytics', { params: { range } })).data }
export async function getDistributionTree(parentUserId?: number, search = '', page = 1): Promise<Paginated<TeamNode>> { return (await apiClient.get<Paginated<TeamNode>>('/distribution/tree', { params: { parent_user_id: parentUserId, search, page } })).data }
export async function getDistributionLedger(page = 1): Promise<Paginated<Commission>> { return (await apiClient.get<Paginated<Commission>>('/distribution/ledger', { params: { page } })).data }
export async function getPayoutAccount(): Promise<PayoutAccount> { return (await apiClient.get<PayoutAccount>('/distribution/payout-account')).data }
export async function savePayoutAccount(alipayAccount: string, realName: string): Promise<PayoutAccount> { return (await apiClient.put<PayoutAccount>('/distribution/payout-account', { alipay_account: alipayAccount, real_name: realName })).data }
export async function listWithdrawals(page = 1): Promise<Paginated<Withdrawal>> { return (await apiClient.get<Paginated<Withdrawal>>('/distribution/withdrawals', { params: { page } })).data }
export async function createWithdrawal(amountMinor: number): Promise<Withdrawal> { return (await apiClient.post<Withdrawal>('/distribution/withdrawals', { amount_cny_minor: amountMinor })).data }
export async function convertToPlatformBalance(amountMinor: number, idempotencyKey: string): Promise<DistributionConversion> { return (await apiClient.post<DistributionConversion>('/distribution/convert', { amount_cny_minor: amountMinor, idempotency_key: idempotencyKey })).data }
