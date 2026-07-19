import { apiClient } from '../client'
import type { Paginated } from '../financial'
import type { PayoutDetails } from './finance'

export interface Tenant { id: number; slug: string; name: string; status: string; site_name: string; primary_domain?: string; core_user_id?: number; wholesale_balance_usd: string; created_at: string }
export interface TenantCreateResult { tenant: Tenant; wholesale_api_key: string }
export interface ProvisioningJob { id: number; tenant_id: number; action: string; status: string; created_at: string }
export interface SaaSPlan { id: number; name: string; billing_period: string; price_cny_minor: number; enabled: boolean; limits: string; created_at: string }
export interface SaaSSubscription { id: number; tenant_id: number; plan_id: number; status: string; paid_cny_minor: number; payment_reference: string; expires_at: string; created_at: string }
export interface SaaSDomain { id: number; tenant_id: number; domain: string; verification_token: string; verified_at?: string; tls_status: string; status: string }
export interface ResourceAllocation { id: number; tenant_id: number; group_id: number; allocation_type: string; concurrency_limit: number; monthly_limit_usd: string }
export interface PartnerWithdrawal { id: number; amount_cny_minor: number; fee_cny_minor: number; status: string; submitted_at: string }
export async function listTenants(page = 1): Promise<Paginated<Tenant>> { return (await apiClient.get<Paginated<Tenant>>('/admin/saas/tenants', { params: { page } })).data }
export async function createTenant(payload: { slug: string; name: string; site_name?: string; site_logo?: string; core_user_id: number; referral_code?: string; totp_code: string }): Promise<TenantCreateResult> { return (await apiClient.post<TenantCreateResult>('/admin/saas/tenants', payload)).data }
export async function fundWholesale(id: number, amountUSD: string, reference: string, totpCode: string): Promise<{ balance_usd: string }> { return (await apiClient.post<{ balance_usd: string }>(`/admin/saas/tenants/${id}/wholesale-funds`, { amount_usd: amountUSD, reference, totp_code: totpCode })).data }
export async function addDomain(id: number, domain: string): Promise<{ id: number; domain: string; verification_token: string }> { return (await apiClient.post(`/admin/saas/tenants/${id}/domains`, { domain })).data }
export async function verifyDomain(domainId: number): Promise<void> { await apiClient.post(`/admin/saas/domains/${domainId}/verify`) }
export async function listProvisioningJobs(tenantId?: number): Promise<ProvisioningJob[]> { return (await apiClient.get<ProvisioningJob[]>('/admin/saas/provisioning-jobs', { params: { tenant_id: tenantId } })).data }
export async function getSaaSConfig(): Promise<{ enabled: boolean }> { return (await apiClient.get('/admin/saas/config')).data }
export async function updateSaaSConfig(enabled: boolean, totpCode: string): Promise<void> { await apiClient.put('/admin/saas/config', { enabled, totp_code: totpCode }) }
export async function listPlans(): Promise<SaaSPlan[]> { return (await apiClient.get<SaaSPlan[]>('/admin/saas/plans')).data }
export async function createPlan(payload: { name: string; billing_period: string; price_cny_minor: number; limits: string }): Promise<SaaSPlan> { return (await apiClient.post<SaaSPlan>('/admin/saas/plans', payload)).data }
export async function listSubscriptions(tenantId?: number): Promise<SaaSSubscription[]> { return (await apiClient.get<SaaSSubscription[]>('/admin/saas/subscriptions', { params: { tenant_id: tenantId } })).data }
export async function recordPaidSubscription(payload: { tenant_id: number; plan_id: number; paid_cny_minor: number; reference: string; totp_code: string }): Promise<{ subscription_id: number }> { return (await apiClient.post('/admin/saas/subscriptions/paid', payload)).data }
export async function listDomains(tenantId?: number): Promise<SaaSDomain[]> { return (await apiClient.get<SaaSDomain[]>('/admin/saas/domains', { params: { tenant_id: tenantId } })).data }
export async function listResourceAllocations(tenantId?: number): Promise<ResourceAllocation[]> { return (await apiClient.get<ResourceAllocation[]>('/admin/saas/resource-pools', { params: { tenant_id: tenantId } })).data }
export async function assignResourcePool(tenantId: number, payload: { group_id: number; allocation_type: string; concurrency_limit: number; monthly_limit_usd: string; totp_code: string }): Promise<void> { await apiClient.put(`/admin/saas/tenants/${tenantId}/resource-pool`, payload) }
export async function listPartnerWithdrawals(status = ''): Promise<PartnerWithdrawal[]> { return (await apiClient.get<PartnerWithdrawal[]>('/admin/saas/partner-withdrawals', { params: { status } })).data }
export async function transitionPartnerWithdrawal(id: number, payload: { status: string; reason?: string; payment_reference?: string; proof_url?: string; totp_code: string }): Promise<PartnerWithdrawal> { return (await apiClient.post<PartnerWithdrawal>(`/admin/saas/partner-withdrawals/${id}/transition`, payload)).data }
export async function getPartnerWithdrawalPayoutDetails(id: number, totpCode: string): Promise<PayoutDetails> { return (await apiClient.post<PayoutDetails>(`/admin/saas/partner-withdrawals/${id}/payout-details`, { totp_code: totpCode })).data }
