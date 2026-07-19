<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <section class="border-b border-gray-200 pb-5 dark:border-dark-700">
        <h1 class="text-xl font-semibold text-gray-950 dark:text-white">{{ t('finance.vouchers.title') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('finance.vouchers.subtitle') }}</p>
      </section>

      <section class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div class="card p-5">
          <form class="grid gap-4 sm:grid-cols-[1fr_1fr_auto] sm:items-end" @submit.prevent="create">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('finance.vouchers.amount') }}
              <input v-model="amount" class="input mt-2 w-full" inputmode="decimal" placeholder="100.00" />
            </label>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('finance.vouchers.totp') }}
              <input v-model="totpCode" class="input mt-2 w-full" inputmode="numeric" maxlength="6" placeholder="000000" />
            </label>
            <button class="btn btn-primary h-10" :disabled="creating">{{ creating ? t('common.processing') : t('common.create') }}</button>
          </form>
          <p class="mt-3 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.vouchers.feeHint') }}</p>
        </div>
        <div class="border-l-4 border-amber-400 bg-amber-50 p-4 text-sm text-amber-900 dark:bg-amber-950/30 dark:text-amber-200">
          {{ t('finance.vouchers.securityNote') }}
        </div>
      </section>

      <section v-if="issuedCode" class="border border-emerald-300 bg-emerald-50 p-5 dark:border-emerald-800 dark:bg-emerald-950/30">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <p class="text-sm font-medium text-emerald-800 dark:text-emerald-200">{{ t('finance.vouchers.createdCode') }}</p>
            <code class="mt-2 block break-all text-base font-semibold text-emerald-950 dark:text-emerald-100">{{ issuedCode }}</code>
          </div>
          <button class="btn btn-secondary btn-sm" @click="copyIssued">{{ t('common.copy') }}</button>
        </div>
      </section>

      <section>
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.vouchers.history') }}</h2>
          <button class="btn btn-secondary btn-sm" @click="load">{{ t('common.refresh') }}</button>
        </div>
        <div class="overflow-x-auto border border-gray-200 dark:border-dark-700">
          <table class="w-full min-w-[760px] text-left text-sm">
            <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800 dark:text-dark-400"><tr><th class="px-4 py-3">ID</th><th class="px-4 py-3">{{ t('finance.vouchers.code') }}</th><th class="px-4 py-3">{{ t('finance.vouchers.amount') }}</th><th class="px-4 py-3">{{ t('finance.vouchers.fee') }}</th><th class="px-4 py-3">{{ t('common.status') }}</th><th class="px-4 py-3">{{ t('common.actions') }}</th></tr></thead>
            <tbody><tr v-for="item in items" :key="item.id" class="border-t border-gray-100 dark:border-dark-700"><td class="px-4 py-3">{{ item.id }}</td><td class="px-4 py-3 font-mono">•••• {{ item.code_last4 }}</td><td class="px-4 py-3">${{ item.face_value }}</td><td class="px-4 py-3">${{ item.fee_amount }}</td><td class="px-4 py-3">{{ item.status }}</td><td class="px-4 py-3"><button v-if="item.status === 'ISSUED'" class="text-sm font-medium text-red-600" @click="cancel(item.id)">{{ t('common.cancel') }}</button></td></tr><tr v-if="!loading && items.length === 0"><td colspan="6" class="px-4 py-10 text-center text-gray-500">{{ t('common.noData') }}</td></tr></tbody>
          </table>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { cancelVoucher, createVoucher, listVouchers, type Voucher } from '@/api/financial'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n(); const app = useAppStore()
const amount = ref(''); const totpCode = ref(''); const creating = ref(false); const loading = ref(false); const items = ref<Voucher[]>([]); const issuedCode = ref('')
async function load() { loading.value = true; try { items.value = (await listVouchers()).items } catch (e) { app.showError(extractApiErrorMessage(e)) } finally { loading.value = false } }
async function create() { creating.value = true; try { const item = await createVoucher(amount.value, totpCode.value); issuedCode.value = item.code || ''; amount.value = ''; totpCode.value = ''; await load() } catch (e) { app.showError(extractApiErrorMessage(e)) } finally { creating.value = false } }
async function cancel(id: number) { try { await cancelVoucher(id); await load() } catch (e) { app.showError(extractApiErrorMessage(e)) } }
async function copyIssued() { if (issuedCode.value) { await navigator.clipboard.writeText(issuedCode.value); app.showSuccess(t('common.copied')) } }
onMounted(load)
</script>
