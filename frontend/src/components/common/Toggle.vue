<template>
  <button
    type="button"
    @click="toggle"
    class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border border-transparent transition-colors duration-200 ease-out focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-500/50 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:focus-visible:ring-offset-dark-900"
    :class="[modelValue ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600']"
    role="switch"
    :aria-checked="modelValue"
  >
    <span
      class="pointer-events-none inline-block h-[22px] w-[22px] -translate-y-px transform rounded-full border border-black/10 bg-white shadow-sm transition duration-200 ease-out"
      :class="[modelValue ? 'translate-x-[19px]' : '-translate-x-px']"
    />
  </button>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>
