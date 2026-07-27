<template>
  <AppLayout>
    <main class="mx-auto flex w-full max-w-7xl flex-col gap-4">
      <section class="overflow-hidden rounded-[28px] border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <header class="flex flex-col gap-5 border-b border-gray-100 p-5 dark:border-dark-700 sm:p-6 lg:flex-row lg:items-center lg:justify-between">
          <div class="min-w-0">
            <div class="inline-flex items-center gap-2 rounded-full bg-primary-50 px-3 py-1 text-sm font-medium text-primary-700 dark:bg-primary-900/20 dark:text-primary-200">
              <Icon name="beaker" size="sm" />
              {{ t('apiTools.workbench.badge') }}
            </div>
            <h1 class="mt-4 text-2xl font-semibold tracking-tight text-gray-950 dark:text-white sm:text-3xl">
              {{ t('apiTools.workbench.title') }}
            </h1>
            <p class="mt-3 max-w-2xl text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ t('apiTools.workbench.subtitle') }}
            </p>
          </div>

          <div class="inline-grid w-full grid-cols-2 rounded-2xl border border-gray-200 bg-gray-50 p-1 text-sm font-medium dark:border-dark-700 dark:bg-dark-900 sm:w-auto">
            <button
              type="button"
              class="inline-flex items-center justify-center gap-2 rounded-xl px-4 py-2.5 transition"
              :class="mode === 'chat'
                ? 'bg-white text-gray-950 shadow-sm dark:bg-dark-700 dark:text-white'
                : 'text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
              @click="setMode('chat')"
            >
              <Icon name="chat" size="sm" />
              {{ t('apiTools.workbench.chatMode') }}
            </button>
            <button
              type="button"
              class="inline-flex items-center justify-center gap-2 rounded-xl px-4 py-2.5 transition"
              :class="mode === 'image'
                ? 'bg-white text-gray-950 shadow-sm dark:bg-dark-700 dark:text-white'
                : 'text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
              @click="setMode('image')"
            >
              <Icon name="sparkles" size="sm" />
              {{ t('apiTools.workbench.imageMode') }}
            </button>
          </div>
        </header>

        <div class="grid min-h-[calc(100dvh-17rem)] lg:grid-cols-[320px_minmax(0,1fr)]">
          <aside class="border-b border-gray-100 bg-white dark:border-dark-700 dark:bg-dark-800 lg:border-b-0 lg:border-r">
            <div class="space-y-5 p-5 sm:p-6">
              <div class="grid gap-3 rounded-2xl bg-gray-50 p-4 text-sm dark:bg-dark-900/70">
                <div class="flex items-center justify-between gap-3">
                  <span class="text-gray-500 dark:text-gray-400">{{ t('apiTools.workbench.currentMode') }}</span>
                  <span class="font-semibold text-gray-900 dark:text-white">{{ currentModeLabel }}</span>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <span class="text-gray-500 dark:text-gray-400">{{ t('apiTools.common.activeKeys') }}</span>
                  <span class="font-semibold text-gray-900 dark:text-white">{{ activeApiKeys.length }}</span>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <span class="text-gray-500 dark:text-gray-400">{{ t('apiTools.common.currentModel') }}</span>
                  <span class="max-w-[160px] truncate font-semibold text-gray-900 dark:text-white" :title="form.model || '-'">
                    {{ form.model || '-' }}
                  </span>
                </div>
              </div>

              <div class="space-y-4">
                <div>
                  <label class="input-label">{{ t('apiTools.common.apiKey') }}</label>
                  <select v-model.number="form.apiKeyId" class="input" :disabled="loadingKeys || activeApiKeys.length === 0">
                    <option :value="0">
                      {{ loadingKeys ? t('apiTools.common.loadingKeys') : t('apiTools.common.selectApiKey') }}
                    </option>
                    <option v-for="key in activeApiKeys" :key="key.id" :value="key.id">
                      {{ apiKeyLabel(key) }}
                    </option>
                  </select>
                  <p v-if="!loadingKeys && activeApiKeys.length === 0" class="input-hint text-amber-600 dark:text-amber-400">
                    {{ noKeyHint }}
                    <router-link to="/keys" class="font-medium text-primary-600 hover:text-primary-700 dark:text-primary-300">
                      {{ t('apiTools.common.createApiKey') }}
                    </router-link>
                  </p>
                </div>

                <div>
                  <label class="input-label">{{ t('apiTools.common.model') }}</label>
                  <select v-model="form.model" class="input" :disabled="loadingModels || modelOptions.length === 0">
                    <option value="">{{ loadingModels ? t('apiTools.common.loadingModels') : t('apiTools.common.selectModel') }}</option>
                    <option v-for="model in modelOptions" :key="model" :value="model">{{ model }}</option>
                  </select>
                  <p v-if="modelLoadError" class="input-hint text-amber-600 dark:text-amber-400">{{ modelLoadError }}</p>
                </div>

                <button type="button" class="btn btn-secondary w-full justify-center" :disabled="loadingKeys || loadingModels" @click="refreshAll">
                  <Icon name="refresh" size="sm" class="mr-1.5" :class="loadingKeys || loadingModels ? 'animate-spin' : ''" />
                  {{ t('common.refresh') }}
                </button>
              </div>

            </div>
          </aside>

          <section class="flex min-h-[calc(100dvh-17rem)] flex-col bg-gray-50/70 dark:bg-dark-900/50">
            <div ref="timelineRef" class="flex-1 overflow-y-auto p-5 sm:p-6">
              <div v-if="currentMessages.length === 0" class="flex min-h-[420px] flex-col items-center justify-center text-center">
                <div class="mb-5 flex h-16 w-16 items-center justify-center rounded-3xl border border-gray-200 bg-white text-primary-600 shadow-sm dark:border-dark-700 dark:bg-dark-800 dark:text-primary-300">
                  <Icon :name="mode === 'chat' ? 'chat' : 'sparkles'" size="xl" />
                </div>
                <h2 class="text-xl font-semibold text-gray-950 dark:text-white">{{ emptyTitle }}</h2>
                <p class="mt-3 max-w-xl text-sm leading-6 text-gray-500 dark:text-gray-400">
                  {{ emptyHint }}
                </p>
                <div class="mt-5 flex flex-wrap justify-center gap-2">
                  <button
                    v-for="prompt in starterPrompts"
                    :key="prompt"
                    type="button"
                    class="rounded-full border border-gray-200 bg-white px-4 py-2 text-sm text-gray-700 transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-primary-900 dark:hover:bg-primary-950/30"
                    @click="useStarterPrompt(prompt)"
                  >
                    {{ prompt }}
                  </button>
                </div>
              </div>

              <div v-else class="mx-auto flex max-w-3xl flex-col gap-4">
                <article
                  v-for="message in currentMessages"
                  :key="message.id"
                  class="max-w-[92%] rounded-3xl px-4 py-3 shadow-sm sm:px-5"
                  :class="message.role === 'user'
                    ? 'ml-auto bg-primary-600 text-white'
                    : message.error
                      ? 'mr-auto border border-red-200 bg-red-50 text-red-800 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-100'
                      : 'mr-auto border border-gray-200 bg-white text-gray-800 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-100'"
                >
                  <div
                    class="mb-2 flex items-center justify-between gap-3 text-xs"
                    :class="message.role === 'user' ? 'text-primary-50/80' : 'text-gray-500 dark:text-gray-400'"
                  >
                    <span>{{ message.role === 'user' ? t('apiTools.workbench.you') : t('apiTools.workbench.assistant') }}</span>
                    <span>{{ formatTime(message.createdAt) }}</span>
                  </div>
                  <p v-if="message.content" class="whitespace-pre-wrap break-words text-sm leading-7">
                    {{ message.content }}
                  </p>

                  <div v-if="message.images?.length" class="mt-4 grid gap-3 md:grid-cols-2">
                    <article
                      v-for="(image, index) in message.images"
                      :key="`${image.safeUrl}-${index}`"
                      class="overflow-hidden rounded-2xl border border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-900/50"
                    >
                      <div class="flex aspect-square items-center justify-center bg-white dark:bg-dark-950">
                        <img :src="image.safeUrl" :alt="t('apiTools.workbench.generatedAlt')" class="h-full w-full object-contain" />
                      </div>
                      <div class="space-y-3 p-3">
                        <p v-if="image.revisedPrompt" class="text-xs leading-5 text-gray-500 dark:text-gray-400">
                          {{ image.revisedPrompt }}
                        </p>
                        <button type="button" class="btn btn-secondary btn-sm w-full justify-center" @click="downloadImage(image, index)">
                          <Icon name="download" size="sm" class="mr-1.5" />
                          {{ t('apiTools.workbench.download') }}
                        </button>
                      </div>
                    </article>
                  </div>

                  <p v-if="message.meta" class="mt-3 text-xs" :class="message.role === 'user' ? 'text-primary-50/80' : 'text-gray-500 dark:text-gray-400'">
                    {{ message.meta }}
                  </p>
                </article>
              </div>
            </div>

            <form class="border-t border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800 sm:p-5" @submit.prevent="sendCurrent">
              <details v-if="mode === 'image'" class="mb-4 rounded-2xl border border-gray-200 bg-gray-50 p-4 text-sm dark:border-dark-700 dark:bg-dark-900/60">
                <summary class="cursor-pointer select-none font-medium text-gray-800 dark:text-gray-100">
                  {{ t('apiTools.workbench.advancedOptions') }}
                </summary>
                <div class="mt-4 grid gap-4 sm:grid-cols-3">
                  <div>
                    <label class="input-label">{{ t('apiTools.workbench.size') }}</label>
                    <select v-model="form.size" class="input">
                      <option value="1024x1024">1024×1024</option>
                      <option value="1536x1024">1536×1024</option>
                      <option value="1024x1536">1024×1536</option>
                      <option value="2048x2048">2048×2048</option>
                    </select>
                  </div>
                  <div>
                    <label class="input-label">{{ t('apiTools.workbench.quality') }}</label>
                    <select v-model="form.quality" class="input">
                      <option value="auto">Auto</option>
                      <option value="low">Low</option>
                      <option value="medium">Medium</option>
                      <option value="high">High</option>
                    </select>
                  </div>
                  <div>
                    <label class="input-label">{{ t('apiTools.workbench.outputFormat') }}</label>
                    <select v-model="form.outputFormat" class="input">
                      <option value="png">PNG</option>
                      <option value="jpeg">JPEG</option>
                      <option value="webp">WebP</option>
                    </select>
                  </div>
                </div>
              </details>

              <div class="mx-auto flex max-w-3xl flex-col gap-3 rounded-3xl border border-gray-200 bg-gray-50 p-3 shadow-sm dark:border-dark-700 dark:bg-dark-900/70">
                <textarea
                  ref="composerRef"
                  v-model="composerText"
                  rows="3"
                  class="min-h-[86px] resize-none rounded-2xl border-0 bg-transparent px-2 py-2 text-sm leading-6 text-gray-900 outline-none placeholder:text-gray-400 focus:ring-0 dark:text-white dark:placeholder:text-dark-400"
                  :placeholder="composerPlaceholder"
                  @keydown.enter.exact.prevent="sendCurrent"
                />
                <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('apiTools.workbench.enterHint') }}</p>
                  <button
                    type="submit"
                    class="btn btn-primary min-h-10 justify-center px-5"
                    :disabled="sendDisabled"
                  >
                    <Icon :name="working ? 'refresh' : mode === 'chat' ? 'chat' : 'sparkles'" size="sm" class="mr-2" :class="working ? 'animate-spin' : ''" />
                    {{ working ? busyLabel : sendLabel }}
                  </button>
                </div>
              </div>
            </form>
          </section>
        </div>
      </section>
    </main>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI } from '@/api'
import { generateImage, listGatewayModels, testChatCompletion, type GeneratedImage } from '@/api/gatewayTools'
import { useAppStore } from '@/stores/app'
import { sanitizeUrl } from '@/utils/url'
import type { ApiKey } from '@/types'

type WorkbenchMode = 'chat' | 'image'
type WorkbenchRole = 'user' | 'assistant'

type SafeGeneratedImage = GeneratedImage & {
  safeUrl: string
}

interface WorkbenchMessage {
  id: string
  role: WorkbenchRole
  content: string
  createdAt: number
  meta?: string
  images?: SafeGeneratedImage[]
  error?: boolean
}

const CHAT_CAPABLE_PLATFORMS = new Set(['openai', 'anthropic', 'gemini', 'grok'])

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const mode = ref<WorkbenchMode>(resolveMode(route.query.mode))
const apiKeys = ref<ApiKey[]>([])
const loadingKeys = ref(false)
const loadingModels = ref(false)
const working = ref(false)
const modelOptions = ref<string[]>([])
const modelLoadError = ref('')
const timelineRef = ref<HTMLElement | null>(null)
const composerRef = ref<HTMLTextAreaElement | null>(null)
let modelRequestSeq = 0

const form = reactive({
  apiKeyId: 0,
  model: '',
  size: '1024x1024',
  quality: 'auto',
  outputFormat: 'png',
})

const chatDraft = ref('')
const imageDraft = ref('一张高级科技感的 3API 品牌海报，深色背景，发光线条，简洁留白。')
const chatMessages = ref<WorkbenchMessage[]>([])
const imageMessages = ref<WorkbenchMessage[]>([])

const chatApiKeys = computed(() =>
  apiKeys.value.filter(key =>
    key.status === 'active' &&
    Boolean(key.key) &&
    (!key.group?.platform || CHAT_CAPABLE_PLATFORMS.has(key.group.platform)),
  ),
)

const imageApiKeys = computed(() =>
  apiKeys.value.filter(key =>
    key.status === 'active' &&
    Boolean(key.key) &&
    (!key.group || (key.group.platform === 'openai' && key.group.allow_image_generation === true)),
  ),
)

const activeApiKeys = computed(() => mode.value === 'image' ? imageApiKeys.value : chatApiKeys.value)
const currentMessages = computed(() => mode.value === 'image' ? imageMessages.value : chatMessages.value)
const selectedApiKey = computed(() => activeApiKeys.value.find(key => key.id === Number(form.apiKeyId)) || null)
const currentModeLabel = computed(() => mode.value === 'image' ? t('apiTools.workbench.imageMode') : t('apiTools.workbench.chatMode'))
const noKeyHint = computed(() => mode.value === 'image' ? t('apiTools.workbench.noImageKey') : t('apiTools.common.noApiKeys'))
const emptyTitle = computed(() => mode.value === 'image' ? t('apiTools.workbench.imageEmptyTitle') : t('apiTools.workbench.chatEmptyTitle'))
const emptyHint = computed(() => mode.value === 'image' ? t('apiTools.workbench.imageEmptyHint') : t('apiTools.workbench.chatEmptyHint'))
const composerPlaceholder = computed(() => mode.value === 'image' ? t('apiTools.workbench.imagePlaceholder') : t('apiTools.workbench.chatPlaceholder'))
const sendLabel = computed(() => mode.value === 'image' ? t('apiTools.workbench.generateImage') : t('apiTools.workbench.sendChat'))
const busyLabel = computed(() => mode.value === 'image' ? t('apiTools.workbench.generating') : t('apiTools.workbench.sending'))

const starterPrompts = computed(() =>
  mode.value === 'image'
    ? [t('apiTools.workbench.imageStarter1'), t('apiTools.workbench.imageStarter2')]
    : [t('apiTools.workbench.chatStarter1'), t('apiTools.workbench.chatStarter2')],
)

const composerText = computed({
  get() {
    return mode.value === 'image' ? imageDraft.value : chatDraft.value
  },
  set(value: string) {
    if (mode.value === 'image') {
      imageDraft.value = value
    } else {
      chatDraft.value = value
    }
  },
})

const sendDisabled = computed(() =>
  working.value ||
  !selectedApiKey.value ||
  !form.model ||
  !composerText.value.trim(),
)

function resolveMode(value: unknown): WorkbenchMode {
  return value === 'image' ? 'image' : 'chat'
}

function setMode(nextMode: WorkbenchMode) {
  if (mode.value === nextMode) return
  mode.value = nextMode
  void router.replace({
    path: '/api-workbench',
    query: nextMode === 'image' ? { mode: 'image' } : {},
  })
}

function messageId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`
}

function apiKeyLabel(key: ApiKey): string {
  const groupName = key.group?.name || key.group?.platform || t('apiTools.workbench.defaultGroup')
  return `${key.name || `API Key #${key.id}`} · ${groupName}`
}

function formatMs(value: number): string {
  if (value < 1000) return `${Math.max(0, Math.round(value))} ms`
  return `${(value / 1000).toFixed(2)} s`
}

function formatTime(value: number): string {
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

function getErrorMessage(error: unknown, fallback: string): string {
  const err = error as { message?: string; requestId?: string }
  const suffix = err?.requestId ? ` · Request ID: ${err.requestId}` : ''
  return `${err?.message || fallback}${suffix}`
}

function chatModelsFrom(models: string[]): string[] {
  return models.filter(model => !model.toLowerCase().startsWith('gpt-image-'))
}

function imageModelsFrom(models: string[]): string[] {
  const imageModels = models.filter(model => model.toLowerCase().startsWith('gpt-image-'))
  if (!imageModels.includes('gpt-image-2')) {
    imageModels.unshift('gpt-image-2')
  }
  return Array.from(new Set(imageModels))
}

function ensureSelectedKey() {
  if (selectedApiKey.value) return
  form.apiKeyId = activeApiKeys.value[0]?.id || 0
}

function safeImage(item: GeneratedImage): SafeGeneratedImage | null {
  const safeUrl = sanitizeUrl(item.url, { allowDataUrl: true })
  if (!safeUrl) return null
  return { ...item, safeUrl }
}

async function scrollToBottom() {
  await nextTick()
  const el = timelineRef.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

async function loadApiKeys() {
  loadingKeys.value = true
  try {
    const response = await keysAPI.list(1, 100, { status: 'active', sort_by: 'created_at', sort_order: 'desc' })
    apiKeys.value = response.items || []
    ensureSelectedKey()
  } catch (error) {
    appStore.showError(getErrorMessage(error, t('apiTools.common.loadKeysFailed')))
  } finally {
    loadingKeys.value = false
  }
}

async function loadModels() {
  ensureSelectedKey()
  const key = selectedApiKey.value
  const requestId = ++modelRequestSeq
  modelLoadError.value = ''
  modelOptions.value = mode.value === 'image' ? ['gpt-image-2'] : []
  form.model = mode.value === 'image' ? 'gpt-image-2' : ''
  if (!key) return

  loadingModels.value = true
  try {
    const models = await listGatewayModels(key.key)
    if (requestId !== modelRequestSeq) return
    modelOptions.value = mode.value === 'image' ? imageModelsFrom(models) : chatModelsFrom(models)
    form.model = modelOptions.value[0] || ''
    if (!form.model) {
      modelLoadError.value = t('apiTools.workbench.noModels')
    }
  } catch (error) {
    if (requestId !== modelRequestSeq) return
    modelLoadError.value = getErrorMessage(error, t('apiTools.common.loadModelsFailed'))
  } finally {
    if (requestId === modelRequestSeq) {
      loadingModels.value = false
    }
  }
}

async function refreshAll() {
  await loadApiKeys()
  await loadModels()
}

function useStarterPrompt(prompt: string) {
  composerText.value = prompt
  void nextTick(() => composerRef.value?.focus())
}

function pushMessage(message: Omit<WorkbenchMessage, 'id' | 'createdAt'>) {
  const target = mode.value === 'image' ? imageMessages : chatMessages
  target.value.push({
    ...message,
    id: messageId(),
    createdAt: Date.now(),
  })
  void scrollToBottom()
}

async function sendCurrent() {
  if (sendDisabled.value) return
  if (mode.value === 'image') {
    await submitImage()
  } else {
    await submitChat()
  }
}

async function submitChat() {
  const key = selectedApiKey.value
  const text = composerText.value.trim()
  if (!key || !form.model || !text) return

  pushMessage({ role: 'user', content: text })
  composerText.value = ''
  working.value = true
  try {
    const result = await testChatCompletion({
      apiKey: key.key,
      model: form.model,
      messages: [{ role: 'user', content: text }],
    })
    const firstToken = result.firstTokenMs !== null
      ? t('apiTools.workbench.firstTokenMeta', { time: formatMs(result.firstTokenMs) })
      : t('apiTools.workbench.firstTokenUnavailable')
    pushMessage({
      role: 'assistant',
      content: result.content || t('apiTools.workbench.chatEmptyReply'),
      meta: `${firstToken} · ${t('apiTools.workbench.durationMeta', { time: formatMs(result.durationMs) })}${result.requestId ? ` · ${result.requestId}` : ''}`,
    })
  } catch (error) {
    pushMessage({
      role: 'assistant',
      content: getErrorMessage(error, t('apiTools.workbench.chatFailed')),
      error: true,
    })
  } finally {
    working.value = false
  }
}

async function submitImage() {
  const key = selectedApiKey.value
  const prompt = composerText.value.trim()
  if (!key || !form.model || !prompt) return

  pushMessage({ role: 'user', content: prompt })
  composerText.value = ''
  working.value = true
  try {
    const response = await generateImage({
      apiKey: key.key,
      model: form.model,
      prompt,
      size: form.size,
      quality: form.quality,
      outputFormat: form.outputFormat,
      n: 1,
    })
    const images = response.images
      .map(safeImage)
      .filter((item): item is SafeGeneratedImage => Boolean(item))

    pushMessage({
      role: 'assistant',
      content: images.length
        ? t('apiTools.workbench.imageDone', { count: images.length })
        : t('apiTools.workbench.noImageReturned'),
      images,
      meta: `${t('apiTools.workbench.durationMeta', { time: formatMs(response.durationMs) })}${response.requestId ? ` · ${response.requestId}` : ''}`,
      error: images.length === 0,
    })
  } catch (error) {
    pushMessage({
      role: 'assistant',
      content: getErrorMessage(error, t('apiTools.workbench.imageFailed')),
      error: true,
    })
  } finally {
    working.value = false
  }
}

function downloadImage(image: SafeGeneratedImage, index: number) {
  const link = document.createElement('a')
  link.href = image.safeUrl
  link.download = `3api-image-${Date.now()}-${index + 1}.${form.outputFormat === 'jpeg' ? 'jpg' : form.outputFormat}`
  link.rel = 'noopener'
  document.body.appendChild(link)
  link.click()
  link.remove()
}

watch(() => route.query.mode, (value) => {
  const nextMode = resolveMode(value)
  if (mode.value !== nextMode) {
    mode.value = nextMode
  }
})

watch(mode, async () => {
  ensureSelectedKey()
  await loadModels()
  await scrollToBottom()
})

watch(() => form.apiKeyId, () => {
  void loadModels()
})

watch(() => currentMessages.value.length, () => {
  void scrollToBottom()
})

onMounted(async () => {
  await refreshAll()
})
</script>
