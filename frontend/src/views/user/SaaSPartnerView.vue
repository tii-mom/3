<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <header class="border-b border-gray-200 pb-5 dark:border-dark-700">
        <h1 class="text-xl font-semibold text-gray-950 dark:text-white">
          {{ t('finance.partner.title') }}
        </h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
          {{ t('finance.partner.subtitle') }}
        </p>
      </header>

      <section class="rounded-lg border border-amber-200 bg-amber-50 p-5 text-amber-950 dark:border-amber-800/60 dark:bg-amber-950/30 dark:text-amber-100">
        <h2 class="text-base font-semibold">{{ t('finance.partner.pausedTitle') }}</h2>
        <p class="mt-2 text-sm leading-6">{{ t('finance.partner.pausedDesc') }}</p>
      </section>

      <section class="grid gap-px overflow-hidden rounded-lg border border-gray-200 bg-gray-200 md:grid-cols-3 dark:border-dark-700 dark:bg-dark-700">
        <div v-for="item in valueProps" :key="item" class="bg-white p-5 dark:bg-dark-900">
          <p class="text-sm font-medium text-gray-900 dark:text-white">{{ item }}</p>
        </div>
      </section>

      <section v-if="contactInfo" class="rounded-lg border border-gray-200 p-5 text-sm text-gray-600 dark:border-dark-700 dark:text-dark-300">
        <p>{{ t('finance.partner.contactHint') }}</p>
        <p class="mt-2 font-medium text-gray-950 dark:text-white">{{ contactInfo }}</p>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const valueProps = computed(() => [
  t('finance.partner.value1'),
  t('finance.partner.value2'),
  t('finance.partner.value3'),
])
const contactInfo = computed(() => appStore.cachedPublicSettings?.contact_info || '')
</script>
