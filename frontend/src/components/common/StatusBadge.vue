<template>
  <div class="inline-flex min-h-6 items-center gap-1.5 rounded border border-gray-200 bg-gray-50 px-2 dark:border-dark-700 dark:bg-dark-800">
    <span
      :class="[
        'inline-block h-1.5 w-1.5 rounded-full',
        variantClass
      ]"
    ></span>
    <span class="text-xs font-medium text-gray-700 dark:text-gray-300">
      {{ label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
  label: string
}>()

const variantClass = computed(() => {
  switch (props.status) {
    case 'active':
    case 'success':
      return 'bg-emerald-500'
    case 'disabled':
    case 'inactive':
    case 'warning':
      return 'bg-amber-500'
    case 'error':
    case 'danger':
      return 'bg-red-500'
    default:
      return 'bg-gray-400'
  }
})
</script>
