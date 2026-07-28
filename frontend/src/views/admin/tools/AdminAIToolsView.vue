<template>
  <AppLayout>
  <div class="space-y-6 p-4 sm:p-6 lg:p-8">
    <section class="rounded-3xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p class="text-sm font-semibold text-primary-600 dark:text-primary-300">AI 工具箱管理</p>
          <h1 class="mt-1 text-2xl font-bold text-gray-950 dark:text-white">添加工具，上架给用户</h1>
          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">填名称、介绍、链接，保存后用户就能打开使用。</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="router.push('/tools')">预览工具箱</button>
          <button type="button" class="btn-primary rounded-2xl px-4 py-2.5" @click="openDialog()">新增工具</button>
        </div>
      </div>
    </section>

    <section class="rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
      <div v-if="loading" class="p-8 text-center text-sm text-gray-500 dark:text-gray-400">正在加载工具...</div>
      <div v-else-if="tools.length === 0" class="p-10 text-center">
        <p class="text-base font-semibold text-gray-950 dark:text-white">还没有工具</p>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">点击“新增工具”添加第一条。</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-dark-700">
          <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-gray-400">
            <tr>
              <th class="px-5 py-3">工具</th>
              <th class="px-5 py-3">分类</th>
              <th class="px-5 py-3">状态</th>
              <th class="px-5 py-3">排序</th>
              <th class="px-5 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
            <tr v-for="tool in tools" :key="tool.id">
              <td class="px-5 py-4">
                <div class="font-semibold text-gray-950 dark:text-white">{{ tool.name }}</div>
                <div class="mt-1 max-w-xl truncate text-xs text-gray-500 dark:text-gray-400">{{ tool.description }}</div>
                <a :href="tool.url" target="_blank" rel="noopener noreferrer" class="mt-1 inline-flex text-xs font-medium text-primary-600 hover:text-primary-700 dark:text-primary-300">{{ tool.url }}</a>
              </td>
              <td class="px-5 py-4">{{ tool.category || 'AI 编程' }}</td>
              <td class="px-5 py-4">
                <span class="rounded-full px-2.5 py-1 text-xs font-semibold" :class="statusClass(tool.status)">{{ statusLabel(tool.status) }}</span>
              </td>
              <td class="px-5 py-4">{{ tool.sort_order }}</td>
              <td class="px-5 py-4 text-right">
                <button class="btn-secondary mr-2 rounded-xl px-3 py-1.5 text-xs" @click="openDialog(tool)">编辑</button>
                <button class="rounded-xl px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 dark:hover:bg-red-500/10" @click="removeTool(tool)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <div v-if="dialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/55 p-4 backdrop-blur-sm">
      <form class="tool-modal max-h-[92vh] w-full max-w-2xl overflow-y-auto rounded-3xl bg-white p-6 shadow-2xl dark:bg-dark-900" @submit.prevent="saveTool">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-sm font-semibold text-primary-600 dark:text-primary-300">工具信息</p>
            <h3 class="mt-1 text-xl font-bold text-gray-950 dark:text-white">{{ dialog.id ? '编辑工具' : '新增工具' }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">只填必要信息。状态为“上架”时，用户端可见。</p>
          </div>
          <button type="button" class="rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-800 dark:hover:text-white" aria-label="关闭" @click="dialog.open = false">✕</button>
        </div>

        <div class="mt-6 space-y-4">
          <label class="tool-field">
            <span>工具名称</span>
            <input v-model="form.name" class="input-field w-full" placeholder="例如：CCS API 管理器" required>
          </label>
          <label class="tool-field">
            <span>分类</span>
            <input v-model="form.category" class="input-field w-full" list="tool-category-options" placeholder="例如：API / CCS 工具" required>
            <datalist id="tool-category-options">
              <option value="网络环境"></option>
              <option value="API 工具"></option>
              <option value="多模型工具"></option>
              <option value="AI 编程"></option>
            </datalist>
          </label>
          <label class="tool-field">
            <span>简短介绍</span>
            <textarea v-model="form.description" class="input-field min-h-24 w-full" placeholder="一句话说明用途，例如：帮你管理 CCS API 配置。"></textarea>
          </label>
          <label class="tool-field">
            <span>直达链接</span>
            <input v-model="form.url" type="url" class="input-field w-full" placeholder="https://example.com/" required>
          </label>
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="tool-field">
              <span>状态</span>
              <select v-model="form.status" class="input-field w-full">
                <option value="draft">草稿</option>
                <option value="published">上架</option>
                <option value="archived">归档</option>
              </select>
            </label>
            <label class="tool-field">
              <span>排序</span>
              <input v-model.number="form.sort_order" type="number" class="input-field w-full" min="0">
            </label>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button type="button" class="btn-secondary rounded-2xl px-4 py-2.5" @click="dialog.open = false">取消</button>
          <button class="btn-primary rounded-2xl px-4 py-2.5" :disabled="saving">{{ saving ? '保存中...' : '保存' }}</button>
        </div>
      </form>
    </div>
  </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminAIToolsAPI, type AITool, type AIToolPayload } from '@/api/aiTools'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'

const router = useRouter()
const appStore = useAppStore()
const loading = ref(true)
const saving = ref(false)
const tools = ref<AITool[]>([])
const dialog = reactive({ open: false, id: 0 })
const form = reactive<AIToolPayload>({
  name: '',
  category: 'AI 编程',
  description: '',
  url: '',
  status: 'published',
  sort_order: 0,
})

function statusLabel(status: string) {
  return ({ draft: '草稿', published: '上架中', archived: '已归档' } as Record<string, string>)[status] || status
}

function statusClass(status: string) {
  if (status === 'published') return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
  if (status === 'draft') return 'bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-gray-300'
}

async function loadTools() {
  loading.value = true
  try {
    const res = await adminAIToolsAPI.list()
    tools.value = res.data || []
  } catch (error: any) {
    appStore.showToast('error', error?.message || '工具加载失败', 3000)
  } finally {
    loading.value = false
  }
}

function openDialog(tool?: AITool) {
  dialog.open = true
  dialog.id = tool?.id || 0
  Object.assign(form, tool ? {
    name: tool.name,
    category: tool.category,
    description: tool.description,
    url: tool.url,
    status: tool.status,
    sort_order: tool.sort_order,
  } : {
    name: '',
    category: 'AI 编程',
    description: '',
    url: '',
    status: 'published',
    sort_order: tools.value.length ? Math.max(...tools.value.map(item => item.sort_order)) + 10 : 10,
  })
}

async function saveTool() {
  saving.value = true
  try {
    const payload = { ...form }
    if (dialog.id) await adminAIToolsAPI.update(dialog.id, payload)
    else await adminAIToolsAPI.create(payload)
    dialog.open = false
    appStore.showToast('success', '工具已保存', 2500)
    await loadTools()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '保存工具失败', 3000)
  } finally {
    saving.value = false
  }
}

async function removeTool(tool: AITool) {
  if (!confirm(`确认删除工具「${tool.name}」？`)) return
  try {
    await adminAIToolsAPI.delete(tool.id)
    appStore.showToast('success', '工具已删除', 2500)
    await loadTools()
  } catch (error: any) {
    appStore.showToast('error', error?.message || '删除工具失败', 3000)
  }
}

onMounted(loadTools)
</script>

<style scoped>
.tool-modal {
  color: rgb(17 24 39);
}

:global(.dark) .tool-modal {
  color: rgb(243 244 246);
}

.tool-field {
  display: block;
}

.tool-field > span {
  margin-bottom: 0.35rem;
  display: block;
  font-size: 0.875rem;
  font-weight: 700;
  color: rgb(55 65 81);
}

:global(.dark) .tool-field > span {
  color: rgb(229 231 235);
}

.tool-modal :deep(.input-field) {
  min-height: 2.75rem;
  border-color: rgb(209 213 219);
  background-color: white;
  color: rgb(17 24 39);
}

:global(.dark) .tool-modal :deep(.input-field) {
  border-color: rgb(55 65 81);
  background-color: rgb(17 24 39);
  color: white;
}
</style>
