<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { getSeoPage } from '@/utils/seo'

const route = useRoute()
const page = computed(() => getSeoPage(route.meta.seoKey))
</script>

<template>
  <div v-if="page" class="seo-page">
    <header class="seo-header">
      <router-link to="/" class="seo-brand" aria-label="3API 首页">
        <span class="seo-brand-mark">3</span>
        <span>3API</span>
      </router-link>
      <nav class="seo-nav" aria-label="主要导航">
        <router-link to="/api-relay">API 中转站</router-link>
        <router-link to="/openai-api">OpenAI API</router-link>
        <router-link to="/codex-api">Codex</router-link>
        <router-link to="/token-guide">Token 指南</router-link>
        <router-link to="/compute-company">渠道合作</router-link>
      </nav>
      <router-link to="/login" class="seo-login">登录</router-link>
    </header>

    <main>
      <section class="seo-hero">
        <div class="seo-hero-copy">
          <p class="seo-eyebrow">{{ page.eyebrow }}</p>
          <h1>{{ page.heading }}</h1>
          <p class="seo-summary">{{ page.summary }}</p>
          <div class="seo-actions">
            <router-link to="/register" class="seo-primary">{{ page.primaryCta }}</router-link>
            <router-link :to="page.secondaryPath" class="seo-secondary">{{ page.secondaryCta }}</router-link>
          </div>
        </div>
        <div class="seo-console" aria-label="3API API 路由示意">
          <div class="seo-console-head"><span>3API ROUTING</span><strong>AVAILABLE</strong></div>
          <div class="seo-console-row"><span>OpenAI</span><code>ready</code></div>
          <div class="seo-console-row"><span>Claude</span><code>ready</code></div>
          <div class="seo-console-row"><span>Gemini</span><code>ready</code></div>
          <div class="seo-console-foot"><span>计量</span><strong>Token 明细</strong></div>
        </div>
      </section>

      <section class="seo-highlights" aria-label="核心能力">
        <article v-for="item in page.highlights" :key="item.title">
          <h2>{{ item.title }}</h2>
          <p>{{ item.text }}</p>
        </article>
      </section>

      <section class="seo-content">
        <div class="seo-content-heading">
          <p class="seo-eyebrow">3API KNOWLEDGE BASE</p>
          <h2>做决定前，你需要了解这些</h2>
        </div>
        <article v-for="(section, index) in page.sections" :key="section.title" class="seo-section">
          <span>{{ String(index + 1).padStart(2, '0') }}</span>
          <div>
            <h3>{{ section.title }}</h3>
            <p>{{ section.body }}</p>
          </div>
        </article>
      </section>

      <section class="seo-faq">
        <p class="seo-eyebrow">FAQ</p>
        <h2>常见问题</h2>
        <details v-for="faq in page.faqs" :key="faq.question">
          <summary>{{ faq.question }}</summary>
          <p>{{ faq.answer }}</p>
        </details>
      </section>

      <section class="seo-final">
        <div>
          <p class="seo-eyebrow">START BUILDING</p>
          <h2>用一个密钥开始调用</h2>
          <p>注册后查看实时可用模型、分组价格和完整接入参数。</p>
        </div>
        <router-link to="/register" class="seo-primary">注册 3API</router-link>
      </section>
    </main>

    <footer class="seo-footer">
      <span>© 2026 3API</span>
      <span>独立第三方 AI API 接入平台</span>
      <router-link to="/chatgpt-plus-vs-api">Plus 与 API 的区别</router-link>
    </footer>
  </div>
</template>

<style scoped>
.seo-page { min-height: 100vh; background: #f7f7f5; color: #161616; font-family: Outfit, "PingFang SC", sans-serif; }
.seo-header { height: 68px; max-width: 1180px; margin: 0 auto; padding: 0 24px; display: flex; align-items: center; gap: 36px; border-bottom: 1px solid #d8d8d2; }
.seo-brand { display: flex; align-items: center; gap: 10px; color: #111; font-size: 20px; font-weight: 800; text-decoration: none; }
.seo-brand-mark { width: 31px; height: 31px; display: grid; place-items: center; background: #ff6a1a; color: white; border-radius: 6px; }
.seo-nav { display: flex; gap: 24px; margin-left: auto; }
.seo-nav a, .seo-login { color: #4d4d49; font-size: 14px; font-weight: 600; text-decoration: none; }
.seo-nav a:hover, .seo-login:hover { color: #e65300; }
.seo-login { padding: 8px 14px; border: 1px solid #bebeb8; border-radius: 6px; }
.seo-hero { max-width: 1180px; margin: 0 auto; padding: 88px 24px 72px; display: grid; grid-template-columns: minmax(0, 1.35fr) minmax(310px, .65fr); gap: 72px; align-items: center; }
.seo-eyebrow { margin: 0 0 18px; color: #d94f00; font: 700 12px/1.2 "JetBrains Mono", monospace; }
.seo-hero h1 { max-width: 760px; margin: 0; font-size: clamp(42px, 6vw, 72px); line-height: 1.08; font-weight: 800; letter-spacing: 0; }
.seo-summary { max-width: 720px; margin: 28px 0 0; color: #565650; font-size: 19px; line-height: 1.8; }
.seo-actions { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 34px; }
.seo-primary, .seo-secondary { min-height: 46px; display: inline-flex; align-items: center; justify-content: center; padding: 0 20px; border-radius: 6px; font-weight: 700; text-decoration: none; }
.seo-primary { background: #ff6a1a; color: white; }
.seo-primary:hover { background: #dc5208; }
.seo-secondary { border: 1px solid #bdbdb6; color: #272724; }
.seo-console { border: 1px solid #252525; background: #151515; color: #e9e9e4; box-shadow: 12px 12px 0 #ff6a1a; }
.seo-console-head, .seo-console-row, .seo-console-foot { display: flex; justify-content: space-between; padding: 15px 17px; border-bottom: 1px solid #343434; font: 12px/1.4 "JetBrains Mono", monospace; }
.seo-console-head strong { color: #69d996; font-weight: 600; }
.seo-console-row code { color: #ff9d68; }
.seo-console-foot { border-bottom: 0; color: #a5a59f; }
.seo-console-foot strong { color: white; }
.seo-highlights { max-width: 1180px; margin: 0 auto; padding: 0 24px 88px; display: grid; grid-template-columns: repeat(3, 1fr); }
.seo-highlights article { min-height: 168px; padding: 28px; border: 1px solid #d4d4ce; border-right: 0; background: #fff; }
.seo-highlights article:last-child { border-right: 1px solid #d4d4ce; }
.seo-highlights h2 { margin: 0 0 15px; font-size: 19px; }
.seo-highlights p, .seo-section p, .seo-faq p, .seo-final p { color: #60605a; line-height: 1.75; }
.seo-content { padding: 88px max(24px, calc((100vw - 1132px) / 2)); background: #191919; color: #f8f8f4; }
.seo-content-heading { display: grid; grid-template-columns: 1fr 2fr; margin-bottom: 46px; }
.seo-content-heading h2 { margin: 0; max-width: 650px; font-size: 40px; line-height: 1.2; }
.seo-section { display: grid; grid-template-columns: 1fr 2fr; gap: 30px; padding: 34px 0; border-top: 1px solid #383834; }
.seo-section > span { color: #ff7b35; font: 500 14px/1.4 "JetBrains Mono", monospace; }
.seo-section h3 { margin: 0 0 10px; font-size: 22px; }
.seo-section p { margin: 0; color: #bcbcb5; }
.seo-faq { max-width: 860px; margin: 0 auto; padding: 96px 24px; }
.seo-faq h2 { margin: 0 0 34px; font-size: 40px; }
.seo-faq details { border-top: 1px solid #cdcdc7; padding: 20px 0; }
.seo-faq details:last-child { border-bottom: 1px solid #cdcdc7; }
.seo-faq summary { cursor: pointer; font-size: 18px; font-weight: 700; }
.seo-faq p { padding-right: 40px; }
.seo-final { max-width: 1132px; margin: 0 auto 80px; padding: 46px; display: flex; align-items: center; justify-content: space-between; gap: 32px; background: #e7ff64; border: 1px solid #151515; }
.seo-final h2 { margin: 0; font-size: 36px; }
.seo-final p:not(.seo-eyebrow) { margin-bottom: 0; color: #49493f; }
.seo-footer { max-width: 1180px; margin: 0 auto; padding: 28px 24px 44px; display: flex; gap: 24px; border-top: 1px solid #d8d8d2; color: #686862; font-size: 13px; }
.seo-footer a { margin-left: auto; color: #333; }
@media (max-width: 820px) {
  .seo-header { height: auto; padding-top: 16px; padding-bottom: 16px; flex-wrap: wrap; }
  .seo-nav { order: 3; width: 100%; overflow-x: auto; gap: 18px; }
  .seo-nav { flex-wrap: wrap; overflow-x: visible; row-gap: 10px; }
  .seo-login { margin-left: auto; }
  .seo-hero { padding-top: 58px; grid-template-columns: 1fr; gap: 44px; }
  .seo-hero h1 { font-size: 44px; }
  .seo-highlights { grid-template-columns: 1fr; }
  .seo-highlights article { border-right: 1px solid #d4d4ce; border-bottom: 0; }
  .seo-highlights article:last-child { border-bottom: 1px solid #d4d4ce; }
  .seo-content-heading, .seo-section { grid-template-columns: 1fr; }
  .seo-final { margin: 0 24px 60px; padding: 34px 26px; align-items: flex-start; flex-direction: column; }
  .seo-footer { flex-wrap: wrap; }
  .seo-footer a { width: 100%; margin-left: 0; }
}
</style>
