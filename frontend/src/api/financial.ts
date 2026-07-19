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
export interface DistributionTier { tier: number; threshold_cny_minor: number; rates_bps: [number, number, number, number, number] }
export interface DistributionDashboard { enabled: boolean; team_volume_cny_minor: number; current_tier: number; next_threshold_cny_minor: number; level_counts: Record<number, number>; available_cny_minor: number; frozen_cny_minor: number; withdrawing_cny_minor: number; debt_cny_minor: number; lifetime_earned_cny_minor: number; tiers: DistributionTier[] }
export interface TeamNode { user_id: number; parent_user_id: number; email_masked: string; username: string; direct_children: number; team_volume_cny_minor: number; current_tier: number }
export interface Commission { id: number; source_order_id: number; source_user_id: number; depth: number; tier: number; rate_bps: number; base_cny_minor: number; amount_cny_minor: number; team_volume_cny_minor: number; status: string; frozen_until: string; created_at: string }
export interface PayoutAccount { account_type: string; account_mask: string; real_name_mask: string }
export interface Withdrawal { id: number; amount_cny_minor: number; fee_cny_minor: number; fee_rate_bps: number; config_version: number; status: string; reject_reason?: string; payment_reference?: string; submitted_at: string }
export interface PartnerDashboard { available_cny_minor: number; frozen_cny_minor: number; withdrawing_cny_minor: number; lifetime_earned_cny_minor: number }
export interface TenantControl { tenant: { id: number; name: string; site_name: string; site_logo: string; primary_domain?: string; wholesale_balance_usd: string }; retail_multiplier: string; payment_provider: string; payment_configured: boolean; instance_config: string }

export async function createVoucher(amount: string, totpCode = ''): Promise<Voucher> { return (await apiClient.post<Voucher>('/user/vouchers', { amount, totp_code: totpCode })).data }
export async function listVouchers(page = 1): Promise<Paginated<Voucher>> { return (await apiClient.get<Paginated<Voucher>>('/user/vouchers', { params: { page } })).data }
export async function cancelVoucher(id: number): Promise<Voucher> { return (await apiClient.post<Voucher>(`/user/vouchers/${id}/cancel`)).data }
export async function getVoucherAvailability(): Promise<VoucherAvailability> { return (await apiClient.get<VoucherAvailability>('/user/vouchers/availability')).data }
export async function getDistributionDashboard(): Promise<DistributionDashboard> { return (await apiClient.get<DistributionDashboard>('/distribution/dashboard')).data }
export async function getDistributionTree(parentUserId?: number, search = '', page = 1): Promise<Paginated<TeamNode>> { return (await apiClient.get<Paginated<TeamNode>>('/distribution/tree', { params: { parent_user_id: parentUserId, search, page } })).data }
export async function getDistributionLedger(page = 1): Promise<Paginated<Commission>> { return (await apiClient.get<Paginated<Commission>>('/distribution/ledger', { params: { page } })).data }
export async function getPayoutAccount(): Promise<PayoutAccount> { return (await apiClient.get<PayoutAccount>('/distribution/payout-account')).data }
export async function savePayoutAccount(alipayAccount: string, realName: string): Promise<PayoutAccount> { return (await apiClient.put<PayoutAccount>('/distribution/payout-account', { alipay_account: alipayAccount, real_name: realName })).data }
export async function listWithdrawals(page = 1): Promise<Paginated<Withdrawal>> { return (await apiClient.get<Paginated<Withdrawal>>('/distribution/withdrawals', { params: { page } })).data }
export async function createWithdrawal(amountMinor: number): Promise<Withdrawal> { return (await apiClient.post<Withdrawal>('/distribution/withdrawals', { amount_cny_minor: amountMinor })).data }
export async function getPartnerDashboard(): Promise<PartnerDashboard> { return (await apiClient.get<PartnerDashboard>('/saas/partner/dashboard')).data }
export async function listPartnerWithdrawals(): Promise<Withdrawal[]> { return (await apiClient.get<Withdrawal[]>('/saas/partner/withdrawals')).data }
export async function createPartnerWithdrawal(amountMinor: number): Promise<Withdrawal> { return (await apiClient.post<Withdrawal>('/saas/partner/withdrawals', { amount_cny_minor: amountMinor })).data }
export async function getTenantControl(): Promise<TenantControl> { return (await apiClient.get<TenantControl>('/saas/tenant/config')).data }
export async function updateTenantControl(payload: { site_name: string; site_logo: string; retail_multiplier: string; payment_provider: string; payment_config: string; instance_config: string }): Promise<TenantControl> { return (await apiClient.put<TenantControl>('/saas/tenant/config', payload)).data }
export async function addTenantDomain(domain: string): Promise<{ verification_token: string }> { return (await apiClient.post('/saas/tenant/domains', { domain })).data }
