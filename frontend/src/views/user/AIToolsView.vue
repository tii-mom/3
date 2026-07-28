<template>
  <div class="mx-auto max-w-7xl space-y-6 p-4 sm:p-6 lg:p-8">
    <section class="overflow-hidden rounded-3xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900 sm:p-8">
      <div class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="inline-flex rounded-full bg-primary-50 px-3 py-1 text-xs font-bold text-primary-700 dark:bg-primary-500/10 dark:text-primary-300">3API 工具箱</p>
          <h1 class="mt-4 text-2xl font-black tracking-tight text-gray-950 dark:text-white sm:text-4xl">常用工具，一键打开</h1>
          <p class="mt-3 max-w-2xl text-sm leading-6 text-gray-600 dark:text-gray-300">网络环境、API 管理、Codex 辅助工具都放在这里。需要哪个，直接打开。</p>
        </div>
        <button class="btn-secondary inline-flex items-center justify-center rounded-2xl px-4 py-2.5" :disabled="loading" @click="loadTools">
          <Icon name="refresh" size="sm" class="mr-2" :class="loading ? 'animate-spin' : ''" />
          刷新
        </button>
      </div>
    </section>

    <div v-if="loading" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <div v-for="i in 6" :key="i" class="h-52 animate-pulse rounded-3xl bg-gray-100 dark:bg-dark-800"></div>
    </div>

    <section v-else-if="tools.length === 0" class="rounded-3xl border border-dashed border-gray-300 bg-white p-10 text-center dark:border-dark-700 dark:bg-dark-900">
      <Icon name="inbox" size="xl" class="mx-auto mb-3 text-gray-400" />
      <h2 class="text-lg font-bold text-gray-950 dark:text-white">工具正在整理中</h2>
      <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">稍后再来看看，管理员会继续补充。</p>
    </section>

    <div v-else class="space-y-8">
      <section v-for="group in groupedTools" :key="group.category" class="space-y-4">
        <div class="flex items-center gap-3">
          <span class="flex h-10 w-10 items-center justify-center rounded-2xl bg-gray-950 text-white shadow-sm dark:bg-white dark:text-gray-950">
            <Icon :name="categoryIcon(group.category)" size="sm" />
          </span>
          <div>
            <h2 class="text-lg font-bold text-gray-950 dark:text-white">{{ group.category }}</h2>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ group.items.length }} 个工具</p>
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <article v-for="tool in group.items" :key="tool.id" class="group flex min-h-56 flex-col rounded-3xl border border-gray-200 bg-white p-5 shadow-sm transition hover:-translate-y-1 hover:shadow-xl dark:border-dark-700 dark:bg-dark-900">
            <div class="mb-4 flex items-start justify-between gap-3">
              <span class="rounded-full bg-gray-100 px-3 py-1 text-xs font-semibold text-gray-600 dark:bg-dark-800 dark:text-gray-300">{{ tool.category }}</span>
              <Icon name="externalLink" size="sm" class="text-gray-400 transition group-hover:text-primary-500" />
            </div>
            <h3 class="text-lg font-black text-gray-950 dark:text-white">{{ tool.name }}</h3>
            <p class="mt-3 line-clamp-3 flex-1 text-sm leading-6 text-gray-600 dark:text-gray-300">{{ tool.description || '点开就能用。' }}</p>
            <a class="btn-primary mt-5 inline-flex w-full items-center justify-center rounded-2xl py-3" :href="tool.url" target="_blank" rel="noopener noreferrer">
              立即打开
              <Icon name="externalLink" size="sm" class="ml-2" />
            </a>
          </article>
        </div>
      </section>
    </div>

    <p class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-xs leading-5 text-amber-700 dark:border-amber-500/20 dark:bg-amber-500/10 dark:text-amber-300">
      提示：第三方工具由对应服务方提供，购买或下载前请先确认价格和安全性。
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { aiToolsAPI, type AITool } from '@/api/aiTools'
import { useAppStore } from '@/stores'

const appStore = useAppStore()
const loading = ref(true)
const tools = ref<AITool[]>([])

const groupedTools = computed(() => {
  const groups = new Map<string, AITool[]>()
  for (const tool of tools.value) {
    const category = tool.category || 'AI 编程'
    if (!groups.has(category)) groups.set(category, [])
    groups.get(category)?.push(tool)
  }
  return Array.from(groups.entries()).map(([category, items]) => ({ category, items }))
})

function categoryIcon(category: string) {
  if (category.includes('网络')) return 'globe'
  if (category.includes('API') || category.includes('CCS')) return 'terminal'
  if (category.includes('模型') || category.includes('Codex')) return 'sparkles'
  return 'link'
}

async function loadTools() {
  loading.value = true
  try {
    const res = await aiToolsAPI.list()
    tools.value = res.data || []
  } catch (error: any) {
    appStore.showToast('error', error?.message || '工具箱加载失败', 3000)
    tools.value = []
  } finally {
    loading.value = false
  }
}

onMounted(loadTools)
</script>
