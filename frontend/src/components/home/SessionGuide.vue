<script setup lang="ts">
import { ref } from 'vue'
import { useClipboard } from '@/composables/useClipboard'

const { copyToClipboard } = useClipboard()
const copied = ref(false)

/**
 * theme: auto（默认）跟随全局明暗主题；dark 用于少数固定深色表面。
 * 销售侧页面已全面支持明暗双主题，因此默认 auto 即可正确渲染。
 */
const props = withDefaults(defineProps<{ theme?: 'auto' | 'dark' | 'light' }>(), { theme: 'auto' })

const SESSION_URL = 'https://chatgpt.com/api/auth/session'

async function copyUrl() {
  const ok = await copyToClipboard(SESSION_URL)
  copied.value = ok
  if (ok) {
    setTimeout(() => { copied.value = false }, 2000)
  }
}
</script>

<template>
  <div class="guide" :class="{ 'guide--dark': props.theme === 'dark' }">
    <ol class="guide__steps">
      <li>
        <span class="guide__index">1</span>
        <div>
          <p class="guide__title">在浏览器登录 ChatGPT</p>
          <p class="guide__desc">打开 chatgpt.com 并确认已登录你要充值的账号</p>
        </div>
      </li>
      <li>
        <span class="guide__index">2</span>
        <div>
          <p class="guide__title">新标签页打开这个地址</p>
          <div class="guide__url">
            <code>{{ SESSION_URL }}</code>
            <button type="button" class="guide__copy" @click="copyUrl">
              {{ copied ? '已复制' : '复制' }}
            </button>
          </div>
        </div>
      </li>
      <li>
        <span class="guide__index">3</span>
        <div>
          <p class="guide__title">全选复制页面内容</p>
          <p class="guide__desc">页面会显示一段文本，按 Ctrl/Cmd + A 全选后复制</p>
        </div>
      </li>
      <li>
        <span class="guide__index">4</span>
        <div>
          <p class="guide__title">粘贴到下方输入框</p>
          <p class="guide__desc">系统会自动识别出登录凭证，不用手动挑字段</p>
        </div>
      </li>
    </ol>

    <p class="guide__safety">
      我们只读取登录凭证用于本次充值，不索要密码。充值完成后，你可以在 ChatGPT 设置里一键登出所有设备来让凭证失效。
    </p>
  </div>
</template>

<style scoped>
/* 明暗令牌：默认跟随全局主题，--dark 可强制深色 */
.guide {
  --sg-title: #09090b;
  --sg-desc: #56565f;
  --sg-line: rgba(9, 9, 11, 0.12);
  --sg-soft: rgba(9, 9, 11, 0.03);
  --sg-accent-text: #b34b1f;
  --sg-accent-soft: rgba(216, 90, 40, 0.1);
  --sg-accent-line: rgba(216, 90, 40, 0.32);

  display: flex;
  flex-direction: column;
  gap: 20px;
}


.guide--dark {
  --sg-title: rgba(255, 255, 255, 0.92);
  --sg-desc: rgba(255, 255, 255, 0.5);
  --sg-line: rgba(255, 255, 255, 0.12);
  --sg-soft: rgba(255, 255, 255, 0.03);
  --sg-accent-text: #f0916a;
  --sg-accent-soft: rgba(224, 99, 47, 0.14);
  --sg-accent-line: rgba(224, 99, 47, 0.38);
}

.guide__steps {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.guide__steps li {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.guide__index {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  margin-top: 1px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  border: 1px solid var(--sg-accent-line);
  background: var(--sg-accent-soft);
  color: var(--sg-accent-text);
  font-size: 12px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.guide__title {
  font-size: 14px;
  font-weight: 500;
  color: var(--sg-title);
}

.guide__desc {
  margin-top: 4px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--sg-desc);
}

.guide__url {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid var(--sg-line);
  background: var(--sg-soft);
}

.guide__url code {
  flex: 1;
  overflow-x: auto;
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  color: var(--sg-desc);
  white-space: nowrap;
}

.guide__copy {
  flex-shrink: 0;
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid var(--sg-accent-line);
  background: transparent;
  color: var(--sg-accent-text);
  font-size: 12px;
  cursor: pointer;
  transition: background 160ms ease-out;
}

.guide__copy:hover {
  background: var(--sg-accent-soft);
}

.guide__safety {
  padding: 12px 14px;
  border-radius: 10px;
  background: var(--sg-soft);
  border: 1px solid var(--sg-line);
  font-size: 12px;
  line-height: 1.75;
  color: var(--sg-desc);
}
</style>

<style>
/* 深色覆盖放在**非 scoped** 块中：Vue 的 scoped-CSS 编译器会把
   `:global(.dark) X` 编译成只有 `.dark`，X 被丢弃，导致生产构建里
   深色规则整体失效。与 SettingsView 的处理方式保持一致。 */
.dark .guide {
  --sg-title: rgba(255, 255, 255, 0.92);
  --sg-desc: rgba(255, 255, 255, 0.5);
  --sg-line: rgba(255, 255, 255, 0.12);
  --sg-soft: rgba(255, 255, 255, 0.03);
  --sg-accent-text: #f0916a;
  --sg-accent-soft: rgba(224, 99, 47, 0.14);
  --sg-accent-line: rgba(224, 99, 47, 0.38);
}
</style>
