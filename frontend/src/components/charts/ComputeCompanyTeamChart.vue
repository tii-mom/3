<template>
  <section class="border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
    <div class="mb-5 flex items-start justify-between gap-3">
      <div>
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('finance.distribution.teamStructure') }}</h2>
        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('finance.distribution.teamStructureHint') }}</p>
      </div>
      <Icon name="users" size="sm" class="text-primary-600 dark:text-primary-400" />
    </div>
    <div class="space-y-3">
      <button v-for="segment in segments" :key="segment.key" type="button" class="group flex w-full items-center gap-3 text-left" @click="emit('select', segment.key)">
        <span class="w-20 shrink-0 text-xs text-gray-500 transition-colors group-hover:text-gray-900 dark:text-dark-400 dark:group-hover:text-white">{{ segment.label }}</span>
        <span class="relative h-8 min-w-0 flex-1 overflow-hidden bg-gray-100 dark:bg-dark-800">
          <span class="absolute inset-y-0 left-0 bg-primary-500/85 transition-all duration-500 group-hover:bg-primary-600 dark:bg-primary-500/70 dark:group-hover:bg-primary-400" :style="{ width: `${segment.percent}%` }"></span>
          <span class="relative z-10 flex h-full items-center justify-end px-2 font-mono text-xs font-medium tabular-nums text-gray-800 dark:text-white">{{ segment.count }}</span>
        </span>
      </button>
    </div>
    <div class="mt-5 flex items-center justify-between border-t border-gray-100 pt-4 text-xs dark:border-dark-700">
      <span class="text-gray-500 dark:text-dark-400">{{ t('finance.distribution.memberCount') }}</span>
      <span class="font-mono font-semibold tabular-nums text-gray-900 dark:text-white">{{ total }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

interface TeamSegment { key: string; label: string; count: number; percent: number }
defineProps<{ segments: TeamSegment[]; total: number }>()
const emit = defineEmits<{ select: [key: string] }>()
const { t } = useI18n()
</script>
