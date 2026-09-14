import { apiClient } from '../client'
import type { Paginated, Voucher, Withdrawal } from '../financial'

export interface AdminCommission {
  // source: 'shop' = 商城订单返点，'distribution' = 历史充值返佣；两表 id 会重号，key 用 `${source}-${id}`。
  source?: 'shop' | 'distribution'
  id: number; source_order_id: number; source_user_id: number; beneficiary_user_id: number; depth: number; tier: number; rate_bps: number; base_cny_minor: number; amount_cny_minor: number; team_volume_cny_minor: number; status: string; frozen_until: string; created_at: string
}
export interface RechargeEvent { id: number; source_order_id: number; user_id: number; base_cny_minor: number; credited_usd: string; first_recharge_bonus_usd: string; config_version: number; status: string; reversal_reason?: string; reversed_at?: string; created_at: string }
export interface DistributionReversal { id: number; recharge_event_id: number; source_order_id: number; user_id: number; reversal_type: string; base_cny_minor: number; principal_usd: string; bonus_usd: string; legacy_rebate_usd: string; commission_cny_minor: number; reason: string; operator_user_id: number; created_at: string }
export interface DistributionRelation { ancestor_user_id: number; descendant_user_id: number; depth: number; created_at: string }
export interface DistributionConversionAudit { id: number; user_id: number; amount_cny_minor: number; usd_amount: string; cny_to_usd_rate?: string; rate_source?: string; usd_to_cny_rate: string; config_version: number; idempotency_key: string; created_at: string }
// 推广成员：只有真正邀请过人的用户才会出现。没有档位，只有人数、业绩与钱包余额。
export interface AdminDistributionMember { user_id: number; email: string; username: string; invitee_count: number; team_volume_cny_minor: number; available_cny_minor: number; frozen_cny_minor: number; lifetime_earned_cny_minor: number }
export interface PayoutDetails { withdrawal_id: number; user_id: number; account_type: string; account: string; real_name: string }
// 推广计划没有档位：返佣比例由每个商品上架时各自设定，配置里只保留钱包与提现规则。
export interface DistributionConfig {
  enabled: boolean; stack_with_legacy: boolean; balance_recharge_multiplier: string; usd_to_cny_rate: string; commission_freeze_hours: number; withdrawal_min_cny_minor: number; withdrawal_daily_limit: number; withdrawal_fee_bps: number; first_recharge_bonus_bps: number; first_recharge_bonus_cap_usd: string; current_config_version: number
}
export interface DistributionPolicyInput { commission_freeze_hours: number; withdrawal_min_cny_minor: number; withdrawal_daily_limit: number; withdrawal_fee_bps: number; first_recharge_bonus_bps: number; first_recharge_bonus_cap_usd: string }

export async function listAdminVouchers(status = '', page = 1): Promise<Paginated<Voucher>> { return (await apiClient.get<Paginated<Voucher>>('/admin/vouchers', { params: { status, page } })).data }
export async function setVoucherRiskLock(id: number, locked: boolean, reason: string): Promise<Voucher> { return (await apiClient.post<Voucher>(`/admin/vouchers/${id}/risk-lock`, { locked, reason })).data }
export async function listAdminWithdrawals(status = '', page = 1): Promise<Paginated<Withdrawal>> { return (await apiClient.get<Paginated<Withdrawal>>('/admin/distribution/withdrawals', { params: { status, page } })).data }
export async function transitionWithdrawal(id: number, payload: { status: string; reason?: string; payment_reference?: string; proof_url?: string }): Promise<Withdrawal> { return (await apiClient.post<Withdrawal>(`/admin/distribution/withdrawals/${id}/transition`, payload)).data }
export async function getVoucherConfig(): Promise<{ enabled: boolean }> { return (await apiClient.get('/admin/vouchers/config')).data }
export async function updateVoucherConfig(enabled: boolean): Promise<void> { await apiClient.put('/admin/vouchers/config', { enabled }) }
export async function getDistributionConfig(): Promise<DistributionConfig> { return (await apiClient.get('/admin/distribution/config')).data }
export async function updateDistributionConfig(enabled: boolean): Promise<void> { await apiClient.put('/admin/distribution/config', { enabled, stack_with_legacy: false }) }
export async function getFinancialRuntimeConfig(): Promise<{ credit_bucket_enforce_enabled: boolean }> { return (await apiClient.get('/admin/distribution/financial-runtime')).data }
export async function updateFinancialRuntimeConfig(enabled: boolean): Promise<void> { await apiClient.put('/admin/distribution/financial-runtime', { credit_bucket_enforce_enabled: enabled }) }
export async function createDistributionPolicyVersion(payload: DistributionPolicyInput): Promise<{ config_version: number }> { return (await apiClient.post('/admin/distribution/config/versions', payload)).data }
export async function listAdminCommissions(page = 1): Promise<Paginated<AdminCommission>> { return (await apiClient.get<Paginated<AdminCommission>>('/admin/distribution/commissions', { params: { page } })).data }
export async function listRechargeEvents(page = 1): Promise<Paginated<RechargeEvent>> { return (await apiClient.get<Paginated<RechargeEvent>>('/admin/distribution/recharge-events', { params: { page } })).data }
export async function reverseRechargeEvent(id: number, payload: { reversal_type: string; reason: string }): Promise<DistributionReversal> { return (await apiClient.post<DistributionReversal>(`/admin/distribution/recharge-events/${id}/reverse`, payload)).data }
export async function listDistributionRelations(page = 1): Promise<Paginated<DistributionRelation>> { return (await apiClient.get<Paginated<DistributionRelation>>('/admin/distribution/relations', { params: { page } })).data }
export async function listDistributionConversions(page = 1): Promise<Paginated<DistributionConversionAudit>> { return (await apiClient.get<Paginated<DistributionConversionAudit>>('/admin/distribution/conversions', { params: { page } })).data }
export async function listDistributionMembers(search = '', page = 1): Promise<Paginated<AdminDistributionMember>> { return (await apiClient.get<Paginated<AdminDistributionMember>>('/admin/distribution/members', { params: { search, page } })).data }
export async function getWithdrawalPayoutDetails(id: number): Promise<PayoutDetails> { return (await apiClient.post<PayoutDetails>(`/admin/distribution/withdrawals/${id}/payout-details`)).data }
export async function getDistributionExchangeRate(): Promise<{ usd_to_cny_rate: string }> { return (await apiClient.get<{ usd_to_cny_rate: string }>('/admin/distribution/exchange-rate')).data }
export async function updateDistributionExchangeRate(rate: string): Promise<{ usd_to_cny_rate: string }> { return (await apiClient.put<{ usd_to_cny_rate: string }>('/admin/distribution/exchange-rate', { usd_to_cny_rate: rate })).data }
