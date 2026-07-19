<template>
  <AppLayout>
    <div class="admin-page-redesign space-y-5">
      <header class="flex flex-col gap-4 border-b border-gray-200 pb-5 lg:flex-row lg:items-end lg:justify-between dark:border-dark-700">
        <div><h1 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('finance.saas.title') }}</h1><p class="mt-1 text-sm text-gray-500 dark:text-dark-400">Core wholesale gateway · managed isolated instances</p></div>
        <div class="flex items-end gap-2"><label class="text-xs font-medium text-gray-500">TOTP<input v-model="security.totp" class="input mt-1 w-28" maxlength="6" inputmode="numeric" placeholder="000000" /></label><button class="btn btn-secondary btn-sm" @click="toggleSaaS">{{ enabled ? t('common.enabled') : t('common.disabled') }}</button></div>
      </header>

      <nav class="overflow-x-auto"><div class="inline-flex h-9 min-w-max border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-700 dark:bg-dark-800"><button v-for="tab in tabs" :key="tab" class="px-4 text-sm" :class="active === tab ? 'bg-white font-medium shadow-sm dark:bg-dark-700' : 'text-gray-500'" @click="active = tab">{{ t(`finance.saas.${tab}`) }}</button></div></nav>

      <template v-if="active === 'tenants'">
        <form class="grid gap-3 border border-gray-200 p-4 md:grid-cols-5 dark:border-dark-700" @submit.prevent="createTenantRecord"><input v-model="tenantForm.slug" class="input" placeholder="tenant-slug" /><input v-model="tenantForm.name" class="input" :placeholder="t('common.name')" /><input v-model.number="tenantForm.core_user_id" class="input" type="number" placeholder="Core user ID" /><input v-model="tenantForm.referral_code" class="input" placeholder="Referral code" /><button class="btn btn-primary">{{ t('common.create') }}</button></form>
        <section v-if="createdKey" class="border-l-4 border-emerald-500 bg-emerald-50 p-4 dark:bg-emerald-950/30"><p class="text-sm font-medium">Wholesale API Key</p><code class="mt-2 block break-all text-sm">{{ createdKey }}</code></section>
        <form class="grid gap-3 border border-gray-200 p-4 md:grid-cols-[160px_1fr_1fr_auto] dark:border-dark-700" @submit.prevent="fund"><select v-model.number="fundForm.tenant_id" class="input"><option :value="0">Tenant</option><option v-for="tenant in tenants" :key="tenant.id" :value="tenant.id">{{ tenant.name }}</option></select><input v-model="fundForm.amount" class="input" inputmode="decimal" placeholder="USD" /><input v-model="fundForm.reference" class="input" :placeholder="t('finance.admin.reference')" /><button class="btn btn-secondary">{{ t('finance.saas.fund') }}</button></form>
        <DataTableShell :empty="tenants.length === 0"><table class="w-full min-w-[900px] text-sm"><thead><tr><th>{{ t('common.name') }}</th><th>Slug</th><th>Domain</th><th class="text-right">Wholesale</th><th>{{ t('common.status') }}</th></tr></thead><tbody><tr v-for="tenant in tenants" :key="tenant.id"><td class="font-medium">{{ tenant.name }}</td><td class="font-mono">{{ tenant.slug }}</td><td>{{ tenant.primary_domain || '-' }}</td><td class="text-right">${{ tenant.wholesale_balance_usd }}</td><td>{{ tenant.status }}</td></tr></tbody></table></DataTableShell>
      </template>

      <template v-else-if="active === 'plans'">
        <form class="grid gap-3 border border-gray-200 p-4 md:grid-cols-[1fr_160px_180px_1fr_auto] dark:border-dark-700" @submit.prevent="createPlanRecord"><input v-model="planForm.name" class="input" :placeholder="t('common.name')" /><select v-model="planForm.billing_period" class="input"><option value="month">month</option><option value="year">year</option></select><input v-model.number="planForm.price_cny_minor" class="input" type="number" placeholder="CNY minor" /><input v-model="planForm.limits" class="input font-mono" placeholder="{}" /><button class="btn btn-primary">{{ t('common.create') }}</button></form>
        <DataTableShell :empty="plans.length === 0"><table class="w-full min-w-[700px] text-sm"><thead><tr><th>ID</th><th>{{ t('common.name') }}</th><th>Period</th><th class="text-right">Price</th><th>Limits</th></tr></thead><tbody><tr v-for="plan in plans" :key="plan.id"><td>#{{ plan.id }}</td><td class="font-medium">{{ plan.name }}</td><td>{{ plan.billing_period }}</td><td class="text-right">{{ cny(plan.price_cny_minor) }}</td><td class="max-w-[320px] truncate font-mono">{{ plan.limits }}</td></tr></tbody></table></DataTableShell>
      </template>

      <template v-else-if="active === 'subscriptions'">
        <form class="grid gap-3 border border-gray-200 p-4 md:grid-cols-5 dark:border-dark-700" @submit.prevent="recordSubscription"><select v-model.number="subscriptionForm.tenant_id" class="input"><option :value="0">Tenant</option><option v-for="tenant in tenants" :key="tenant.id" :value="tenant.id">{{ tenant.name }}</option></select><select v-model.number="subscriptionForm.plan_id" class="input"><option :value="0">Plan</option><option v-for="plan in plans" :key="plan.id" :value="plan.id">{{ plan.name }}</option></select><input v-model.number="subscriptionForm.paid_cny_minor" class="input" type="number" placeholder="Paid CNY minor" /><input v-model="subscriptionForm.reference" class="input" :placeholder="t('finance.admin.reference')" /><button class="btn btn-primary">{{ t('finance.saas.recordPayment') }}</button></form>
        <DataTableShell :empty="subscriptions.length === 0"><table class="w-full min-w-[820px] text-sm"><thead><tr><th>ID</th><th>Tenant</th><th>Plan</th><th class="text-right">Paid</th><th>{{ t('finance.admin.reference') }}</th><th>Expires</th></tr></thead><tbody><tr v-for="item in subscriptions" :key="item.id"><td>#{{ item.id }}</td><td>#{{ item.tenant_id }}</td><td>#{{ item.plan_id }}</td><td class="text-right">{{ cny(item.paid_cny_minor) }}</td><td>{{ item.payment_reference }}</td><td>{{ date(item.expires_at) }}</td></tr></tbody></table></DataTableShell>
      </template>

      <template v-else-if="active === 'domains'">
        <form class="grid gap-3 border border-gray-200 p-4 md:grid-cols-[200px_1fr_auto] dark:border-dark-700" @submit.prevent="addTenantDomain"><select v-model.number="domainForm.tenant_id" class="input"><option :value="0">Tenant</option><option v-for="tenant in tenants" :key="tenant.id" :value="tenant.id">{{ tenant.name }}</option></select><input v-model="domainForm.domain" class="input" placeholder="api.example.com" /><button class="btn btn-primary">{{ t('finance.saas.domain') }}</button></form>
        <DataTableShell :empty="domains.length === 0"><table class="w-full min-w-[920px] text-sm"><thead><tr><th>Tenant</th><th>Domain</th><th>TXT token</th><th>TLS</th><th>{{ t('common.status') }}</th><th>{{ t('common.actions') }}</th></tr></thead><tbody><tr v-for="item in domains" :key="item.id"><td>#{{ item.tenant_id }}</td><td class="font-medium">{{ item.domain }}</td><td class="max-w-[360px] truncate font-mono">{{ item.verification_token }}</td><td>{{ item.tls_status }}</td><td>{{ item.status }}</td><td><button class="font-medium text-primary-600" :disabled="item.status === 'verified'" @click="verify(item.id)">{{ t('finance.saas.verify') }}</button></td></tr></tbody></table></DataTableShell>
      </template>

      <template v-else-if="active === 'resources'">
        <form class="grid gap-3 border border-gray-200 p-4 md:grid-cols-6 dark:border-dark-700" @submit.prevent="assignResource"><select v-model.number="resourceForm.tenant_id" class="input"><option :value="0">Tenant</option><option v-for="tenant in tenants" :key="tenant.id" :value="tenant.id">{{ tenant.name }}</option></select><input v-model.number="resourceForm.group_id" class="input" type="number" placeholder="Group ID" /><select v-model="resourceForm.allocation_type" class="input"><option value="shared">shared</option><option value="dedicated">dedicated</option></select><input v-model.number="resourceForm.concurrency_limit" class="input" type="number" placeholder="Concurrency" /><input v-model="resourceForm.monthly_limit_usd" class="input" inputmode="decimal" placeholder="Monthly USD" /><button class="btn btn-primary">{{ t('common.save') }}</button></form>
        <DataTableShell :empty="resources.length === 0"><table class="w-full min-w-[760px] text-sm"><thead><tr><th>Tenant</th><th>Group</th><th>Type</th><th class="text-right">Concurrency</th><th class="text-right">Monthly USD</th></tr></thead><tbody><tr v-for="item in resources" :key="item.id"><td>#{{ item.tenant_id }}</td><td>#{{ item.group_id }}</td><td>{{ item.allocation_type }}</td><td class="text-right">{{ item.concurrency_limit }}</td><td class="text-right">${{ item.monthly_limit_usd }}</td></tr></tbody></table></DataTableShell>
      </template>

      <template v-else-if="active === 'jobs'">
        <DataTableShell :empty="jobs.length === 0"><table class="w-full min-w-[640px] text-sm"><thead><tr><th>ID</th><th>Tenant</th><th>Action</th><th>{{ t('common.status') }}</th><th>Created</th></tr></thead><tbody><tr v-for="job in jobs" :key="job.id"><td>#{{ job.id }}</td><td>#{{ job.tenant_id }}</td><td>{{ job.action }}</td><td>{{ job.status }}</td><td>{{ date(job.created_at) }}</td></tr></tbody></table></DataTableShell>
      </template>

      <template v-else>
        <section class="grid gap-3 border border-gray-200 p-4 sm:grid-cols-2 dark:border-dark-700"><label class="text-sm font-medium">{{ t('finance.admin.reference') }}<input v-model="security.reference" class="input mt-1.5 w-full" /></label><label class="text-sm font-medium">{{ t('finance.admin.reason') }}<input v-model="security.reason" class="input mt-1.5 w-full" /></label></section>
        <div v-if="payoutDetails" class="border-l-4 border-emerald-500 bg-emerald-50 px-4 py-3 text-sm text-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-200">{{ payoutDetails.real_name }} · {{ payoutDetails.account_type }} · {{ payoutDetails.account }}</div>
        <DataTableShell :empty="partnerWithdrawals.length === 0"><table class="w-full min-w-[860px] text-sm"><thead><tr><th>ID</th><th class="text-right">Amount</th><th class="text-right">{{ t('finance.vouchers.fee') }}</th><th>{{ t('common.status') }}</th><th>{{ t('common.actions') }}</th></tr></thead><tbody><tr v-for="item in partnerWithdrawals" :key="item.id"><td>#{{ item.id }}</td><td class="text-right">{{ cny(item.amount_cny_minor) }}</td><td class="text-right">{{ cny(item.fee_cny_minor) }}</td><td>{{ item.status }}</td><td><div class="flex gap-2"><button class="btn btn-secondary btn-sm" @click="showPartnerPayout(item.id)">{{ t('finance.admin.payoutDetails') }}</button><button v-if="item.status === 'SUBMITTED'" class="btn btn-secondary btn-sm" @click="transitionPartner(item.id, 'APPROVED')">{{ t('finance.admin.approve') }}</button><button v-if="item.status === 'APPROVED'" class="btn btn-primary btn-sm" @click="transitionPartner(item.id, 'PAID')">{{ t('finance.admin.paid') }}</button><button v-if="item.status === 'SUBMITTED' || item.status === 'APPROVED'" class="btn btn-danger btn-sm" @click="transitionPartner(item.id, 'REJECTED')">{{ t('finance.admin.reject') }}</button></div></td></tr></tbody></table></DataTableShell>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { defineComponent, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { addDomain, assignResourcePool, createPlan, createTenant, fundWholesale, getPartnerWithdrawalPayoutDetails, getSaaSConfig, listDomains, listPartnerWithdrawals, listPlans, listProvisioningJobs, listResourceAllocations, listSubscriptions, listTenants, recordPaidSubscription, transitionPartnerWithdrawal, updateSaaSConfig, verifyDomain, type PartnerWithdrawal, type ProvisioningJob, type ResourceAllocation, type SaaSDomain, type SaaSPlan, type SaaSSubscription, type Tenant } from '@/api/admin/saas'
import type { PayoutDetails } from '@/api/admin/finance'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const DataTableShell = defineComponent({ props: { empty: Boolean }, setup(props, { slots }) { return () => h('section', { class: 'overflow-x-auto border border-gray-200 dark:border-dark-700 [&_th]:bg-gray-50 [&_th]:px-4 [&_th]:py-3 [&_th]:text-left [&_th]:font-medium [&_th]:text-gray-500 dark:[&_th]:bg-dark-800 [&_td]:border-t [&_td]:border-gray-100 [&_td]:px-4 [&_td]:py-3 dark:[&_td]:border-dark-700' }, props.empty ? h('p', { class: 'py-10 text-center text-sm text-gray-500' }, 'No data') : slots.default?.()) } })
const { t } = useI18n()
const app = useAppStore()
const tabs = ['tenants', 'plans', 'subscriptions', 'domains', 'resources', 'jobs', 'withdrawals'] as const
const active = ref<(typeof tabs)[number]>('tenants')
const enabled = ref(false)
const tenants = ref<Tenant[]>([])
const plans = ref<SaaSPlan[]>([])
const subscriptions = ref<SaaSSubscription[]>([])
const domains = ref<SaaSDomain[]>([])
const resources = ref<ResourceAllocation[]>([])
const jobs = ref<ProvisioningJob[]>([])
const partnerWithdrawals = ref<PartnerWithdrawal[]>([])
const createdKey = ref('')
const payoutDetails = ref<PayoutDetails>()
const security = reactive({ totp: '', reference: '', reason: '' })
const tenantForm = reactive({ slug: '', name: '', core_user_id: 0, referral_code: '' })
const fundForm = reactive({ tenant_id: 0, amount: '', reference: '' })
const planForm = reactive({ name: '', billing_period: 'month', price_cny_minor: 0, limits: '{}' })
const subscriptionForm = reactive({ tenant_id: 0, plan_id: 0, paid_cny_minor: 0, reference: '' })
const domainForm = reactive({ tenant_id: 0, domain: '' })
const resourceForm = reactive({ tenant_id: 0, group_id: 0, allocation_type: 'shared', concurrency_limit: 0, monthly_limit_usd: '' })
const cny = (minor: number) => new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY' }).format(minor / 100)
const date = (value: string) => new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))

async function load() { try { const results = await Promise.all([getSaaSConfig(), listTenants(), listPlans(), listSubscriptions(), listDomains(), listResourceAllocations(), listProvisioningJobs(), listPartnerWithdrawals()]); enabled.value = results[0].enabled; tenants.value = results[1].items; plans.value = results[2]; subscriptions.value = results[3]; domains.value = results[4]; resources.value = results[5]; jobs.value = results[6]; partnerWithdrawals.value = results[7] } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function toggleSaaS() { if (!security.totp) return app.showError('TOTP required'); try { await updateSaaSConfig(!enabled.value, security.totp); enabled.value = !enabled.value } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function createTenantRecord() { if (!security.totp) return app.showError('TOTP required'); try { const result = await createTenant({ ...tenantForm, totp_code: security.totp }); createdKey.value = result.wholesale_api_key; Object.assign(tenantForm, { slug: '', name: '', core_user_id: 0, referral_code: '' }); await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function fund() { if (!security.totp) return app.showError('TOTP required'); try { await fundWholesale(fundForm.tenant_id, fundForm.amount, fundForm.reference, security.totp); fundForm.amount = ''; fundForm.reference = ''; await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function createPlanRecord() { try { await createPlan(planForm); Object.assign(planForm, { name: '', billing_period: 'month', price_cny_minor: 0, limits: '{}' }); await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function recordSubscription() { try { await recordPaidSubscription({ ...subscriptionForm, totp_code: security.totp }); subscriptionForm.reference = ''; await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function addTenantDomain() { try { await addDomain(domainForm.tenant_id, domainForm.domain); domainForm.domain = ''; await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function verify(id: number) { try { await verifyDomain(id); await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function assignResource() { try { await assignResourcePool(resourceForm.tenant_id, { group_id: resourceForm.group_id, allocation_type: resourceForm.allocation_type, concurrency_limit: resourceForm.concurrency_limit, monthly_limit_usd: resourceForm.monthly_limit_usd, totp_code: security.totp }); await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function transitionPartner(id: number, status: string) { try { await transitionPartnerWithdrawal(id, { status, reason: security.reason, payment_reference: security.reference, totp_code: security.totp }); payoutDetails.value = undefined; await load() } catch (error) { app.showError(extractApiErrorMessage(error)) } }
async function showPartnerPayout(id: number) { try { payoutDetails.value = await getPartnerWithdrawalPayoutDetails(id, security.totp) } catch (error) { app.showError(extractApiErrorMessage(error)) } }
onMounted(load)
</script>
