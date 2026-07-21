<template>
  <section class="relative overflow-hidden border border-gray-200 bg-[#11171d] p-5 text-white shadow-sm sm:p-6 dark:border-dark-700">
    <div class="pointer-events-none absolute inset-0 opacity-20" aria-hidden="true">
      <div class="absolute inset-y-0 left-1/3 w-px bg-primary-300"></div>
      <div class="absolute inset-x-0 top-1/2 h-px bg-primary-300"></div>
      <div class="absolute -right-16 -top-16 h-44 w-44 rounded-full border border-primary-300/40"></div>
    </div>
    <div class="relative flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
      <div class="max-w-xl">
        <div class="flex items-center gap-2 text-primary-300">
          <Icon name="users" size="sm" :stroke-width="1.8" />
          <span class="text-[11px] font-semibold uppercase tracking-[0.18em]">3API / COMPUTE COMPANY</span>
        </div>
        <h2 class="mt-3 text-xl font-semibold tracking-tight sm:text-2xl">{{ t('finance.distribution.shareCardTitle') }}</h2>
        <p class="mt-2 max-w-lg text-sm leading-6 text-gray-300">{{ t('finance.distribution.shareCardSubtitle') }}</p>
        <div v-if="inviteCode" class="mt-4 inline-flex items-center gap-2 border border-white/15 bg-white/5 px-3 py-2 font-mono text-sm text-gray-100">
          <span class="text-xs text-gray-400">{{ t('finance.distribution.inviteCode') }}</span>
          <span>{{ inviteCode }}</span>
        </div>
      </div>
      <div class="relative flex flex-col gap-2 sm:flex-row lg:flex-col lg:items-stretch">
        <button type="button" class="btn btn-primary inline-flex items-center justify-center gap-2 whitespace-nowrap" :disabled="!inviteLink" @click="openPreview">
          <Icon name="grid" size="sm" :stroke-width="1.8" />
          {{ t('finance.distribution.createShareCard') }}
        </button>
        <button type="button" class="btn btn-secondary inline-flex items-center justify-center gap-2 whitespace-nowrap !border-white/20 !bg-white/5 !text-white hover:!bg-white/10" :disabled="!inviteLink" @click="copyLink">
          <Icon name="link" size="sm" :stroke-width="1.8" />
          {{ t('finance.distribution.copyLink') }}
        </button>
      </div>
    </div>
  </section>

  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4 backdrop-blur-sm" role="dialog" aria-modal="true" :aria-label="t('finance.distribution.shareCardTitle')" @click.self="close">
      <div class="flex max-h-[92vh] w-full max-w-4xl flex-col overflow-y-auto border border-gray-200 bg-white shadow-2xl dark:border-dark-600 dark:bg-dark-900">
        <div class="flex items-center justify-between border-b border-gray-200 px-5 py-4 dark:border-dark-700">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-primary-600 dark:text-primary-400">3API / SHARE CARD</p>
            <h3 class="mt-1 text-lg font-semibold text-gray-950 dark:text-white">{{ t('finance.distribution.shareCardPreview') }}</h3>
          </div>
          <button type="button" class="btn btn-secondary btn-sm !px-2" :aria-label="t('common.close')" @click="close">
            <Icon name="x" size="sm" />
          </button>
        </div>
        <div class="grid gap-6 p-5 lg:grid-cols-[minmax(0,1fr)_260px] lg:p-6">
          <div class="flex min-h-[360px] items-center justify-center bg-gray-100 p-4 dark:bg-dark-800 sm:p-6">
            <div v-if="generating" class="text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
            <img v-else-if="posterDataUrl" :src="posterDataUrl" :alt="t('finance.distribution.shareCardAlt')" class="max-h-[62vh] w-auto max-w-full shadow-xl" />
          </div>
          <div class="flex flex-col justify-center gap-3">
            <p class="text-sm leading-6 text-gray-600 dark:text-gray-300">{{ t('finance.distribution.shareCardHint') }}</p>
            <button type="button" class="btn btn-primary inline-flex items-center justify-center gap-2" :disabled="!posterDataUrl || generating" @click="downloadPoster">
              <Icon name="download" size="sm" />
              {{ t('finance.distribution.downloadShareCard') }}
            </button>
            <button type="button" class="btn btn-secondary inline-flex items-center justify-center gap-2" :disabled="!posterDataUrl || generating" @click="copyPoster">
              <Icon name="copy" size="sm" />
              {{ t('finance.distribution.copyShareCard') }}
            </button>
            <button type="button" class="btn btn-secondary inline-flex items-center justify-center gap-2" @click="copyLink">
              <Icon name="link" size="sm" />
              {{ t('finance.distribution.copyLink') }}
            </button>
            <p class="text-xs leading-5 text-gray-500 dark:text-gray-400">{{ t('finance.distribution.shareCardPrivacy') }}</p>
          </div>
        </div>
      </div>
    </div>
  </Teleport>

  <canvas ref="posterCanvas" class="hidden" width="1080" height="1440"></canvas>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  inviteLink: string
  inviteCode: string
}>()

const { t, locale } = useI18n()
const app = useAppStore()
const { copyToClipboard } = useClipboard()
const open = ref(false)
const generating = ref(false)
const posterDataUrl = ref('')
const posterCanvas = ref<HTMLCanvasElement>()

function roundRect(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number) {
  ctx.beginPath()
  ctx.moveTo(x + radius, y)
  ctx.arcTo(x + width, y, x + width, y + height, radius)
  ctx.arcTo(x + width, y + height, x, y + height, radius)
  ctx.arcTo(x, y + height, x, y, radius)
  ctx.arcTo(x, y, x + width, y, radius)
  ctx.closePath()
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = reject
    image.src = src
  })
}

function wrapCopy() {
  if (locale.value.startsWith('zh')) return ['我开了一家神奇的', '算力公司，欢迎领导', '视察工作～']
  return ['I started a remarkable', 'compute company.', 'Leaders, welcome to inspect.']
}

async function renderPoster() {
  if (!props.inviteLink || !posterCanvas.value) return
  generating.value = true
  try {
    const canvas = posterCanvas.value
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    const qrDataUrl = await QRCode.toDataURL(props.inviteLink, { width: 460, margin: 2, errorCorrectionLevel: 'H', color: { dark: '#16212a', light: '#ffffff' } })
    const qrImage = await loadImage(qrDataUrl)
    const width = canvas.width
    const height = canvas.height

    ctx.fillStyle = '#10161d'
    ctx.fillRect(0, 0, width, height)
    ctx.strokeStyle = 'rgba(226, 107, 54, 0.12)'
    ctx.lineWidth = 1
    for (let x = 60; x < width; x += 120) {
      ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, height); ctx.stroke()
    }
    for (let y = 60; y < height; y += 120) {
      ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(width, y); ctx.stroke()
    }
    ctx.fillStyle = '#e26b36'
    ctx.fillRect(72, 74, 10, 10)
    ctx.font = '600 34px Arial, sans-serif'
    ctx.fillText('3API', 102, 84)
    ctx.fillStyle = 'rgba(255,255,255,0.52)'
    ctx.font = '500 18px Arial, sans-serif'
    ctx.fillText('COMPUTE COMPANY / INVITATION', 102, 116)

    ctx.fillStyle = '#ffffff'
    ctx.font = '600 56px Arial, sans-serif'
    wrapCopy().forEach((line, index) => ctx.fillText(line, 72, 300 + index * 76))
    ctx.fillStyle = 'rgba(255,255,255,0.62)'
    ctx.font = '400 22px Arial, sans-serif'
    ctx.fillText(t('finance.distribution.shareCardPosterHint'), 72, 560)

    ctx.fillStyle = '#e26b36'
    ctx.fillRect(72, 630, 120, 4)
    ctx.fillStyle = 'rgba(255,255,255,0.48)'
    ctx.font = '500 18px Arial, sans-serif'
    ctx.fillText(t('finance.distribution.inviteCode'), 72, 690)
    ctx.fillStyle = '#ffffff'
    ctx.font = '600 34px monospace'
    ctx.fillText(props.inviteCode || '-', 72, 735)

    ctx.fillStyle = '#ffffff'
    roundRect(ctx, 610, 770, 380, 430, 20)
    ctx.fill()
    ctx.drawImage(qrImage, 650, 815, 300, 300)
    ctx.fillStyle = '#16212a'
    ctx.font = '600 21px Arial, sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText(t('finance.distribution.scanToJoin'), 800, 1165)
    ctx.textAlign = 'left'
    ctx.fillStyle = 'rgba(255,255,255,0.52)'
    ctx.font = '400 18px Arial, sans-serif'
    ctx.fillText('3API / 3API.COM', 72, 1340)
    ctx.fillText('POWERED BY PRECISION INFRASTRUCTURE', 72, 1374)
    posterDataUrl.value = canvas.toDataURL('image/png')
  } finally {
    generating.value = false
  }
}

function openPreview() {
  if (!props.inviteLink) return
  open.value = true
}

function close() {
  open.value = false
}

async function copyLink() {
  if (props.inviteLink) await copyToClipboard(props.inviteLink, t('finance.distribution.copied'))
}

function downloadPoster() {
  if (!posterDataUrl.value) return
  const link = document.createElement('a')
  link.href = posterDataUrl.value
  link.download = '3api-compute-company-invite.png'
  link.click()
}

async function copyPoster() {
  if (!posterCanvas.value || !posterDataUrl.value || typeof ClipboardItem === 'undefined' || !navigator.clipboard?.write) {
    downloadPoster()
    return
  }
  posterCanvas.value.toBlob(async (blob) => {
    if (!blob) return
    try {
      await navigator.clipboard.write([new ClipboardItem({ 'image/png': blob })])
      app.showSuccess(t('finance.distribution.shareCardCopied'))
    } catch {
      downloadPoster()
    }
  }, 'image/png')
}

watch(open, async (value) => {
  if (value) {
    await nextTick()
    await renderPoster()
  }
})
watch(() => [props.inviteLink, props.inviteCode, locale.value], () => {
  if (open.value) void renderPoster()
})
</script>
