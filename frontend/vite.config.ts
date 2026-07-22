import { defineConfig, loadEnv, Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import checker from 'vite-plugin-checker'
import { resolve } from 'path'
import { readFileSync, writeFileSync } from 'node:fs'
import seoPages from './src/content/seo-pages.json'

const siteUrl = 'https://3api.shop'

type SeoPage = (typeof seoPages)[keyof typeof seoPages]

function replaceTag(html: string, pattern: RegExp, replacement: string): string {
  return pattern.test(html) ? html.replace(pattern, replacement) : html.replace('</head>', `  ${replacement}\n</head>`)
}

function renderSeoBody(page: SeoPage): string {
  const highlights = page.highlights.map((item) => `<article><h2>${escapeHtml(item.title)}</h2><p>${escapeHtml(item.text)}</p></article>`).join('')
  const sections = page.sections.map((section) => `<section><h2>${escapeHtml(section.title)}</h2><p>${escapeHtml(section.body)}</p></section>`).join('')
  const faqs = page.faqs.map((faq) => `<details><summary>${escapeHtml(faq.question)}</summary><p>${escapeHtml(faq.answer)}</p></details>`).join('')
  return `<div id="seo-snapshot"><header><a href="/">3API</a><nav><a href="/api-relay">API 中转站</a><a href="/openai-api">OpenAI API</a><a href="/codex-api">Codex API</a><a href="/token-guide">Token 指南</a><a href="/compute-company">渠道合作</a></nav></header><main><p class="eyebrow">${escapeHtml(page.eyebrow)}</p><h1>${escapeHtml(page.heading)}</h1><p class="summary">${escapeHtml(page.summary)}</p><a class="cta" href="/register">${escapeHtml(page.primaryCta)}</a><div class="highlights">${highlights}</div><div class="content">${sections}</div><section class="faq"><h2>常见问题</h2>${faqs}</section></main><footer>3API · 独立第三方 AI API 接入平台</footer></div>`
}

function renderHomeBody(): string {
  return `<div id="seo-snapshot"><header><strong>3API</strong><nav><a href="/api-relay">API 中转站</a><a href="/openai-api">OpenAI API</a><a href="/codex-api">Codex API</a><a href="/token-guide">Token 指南</a><a href="/compute-company">渠道合作</a></nav></header><main><p class="eyebrow">MULTI-MODEL API GATEWAY</p><h1>3API AI API 中转站</h1><p class="summary">一个 API Key，统一接入后台已开通的 OpenAI、Claude、Gemini 等主流 AI 模型。支持智能路由、Token 用量统计与按量计费。</p><a class="cta" href="/register">立即开始</a><div class="highlights"><article><h2>统一模型接口</h2><p>减少多套 SDK、密钥和账单的维护成本。</p></article><article><h2>Token 计费透明</h2><p>按模型查看输入、输出与缓存用量。</p></article><article><h2>开发者工作流</h2><p>接入 Codex、Agent、IDE 与已有应用。</p></article></div></main><footer>3API · 独立第三方 AI API 接入平台</footer></div>`
}

const snapshotStyle = `<style id="seo-snapshot-style">#seo-snapshot{font-family:system-ui,-apple-system,"PingFang SC",sans-serif;max-width:1120px;margin:auto;padding:24px;color:#171717}#seo-snapshot header{display:flex;justify-content:space-between;gap:24px;padding:12px 0;border-bottom:1px solid #ddd}#seo-snapshot nav{display:flex;gap:18px;flex-wrap:wrap}#seo-snapshot a{color:#b94300}#seo-snapshot main{padding:72px 0}#seo-snapshot .eyebrow{font-size:12px;color:#c84d0b}#seo-snapshot h1{font-size:clamp(40px,7vw,72px);line-height:1.1;max-width:900px}#seo-snapshot .summary{font-size:19px;line-height:1.7;max-width:780px}#seo-snapshot .cta{display:inline-block;margin:20px 0 50px;padding:13px 20px;background:#e85d11;color:white;text-decoration:none}.highlights{display:grid;grid-template-columns:repeat(3,1fr);gap:1px;background:#ccc}.highlights article{padding:24px;background:#fff}.content{margin-top:70px}.content section,.faq details{padding:24px 0;border-top:1px solid #ddd}.content p,.faq p{line-height:1.7;color:#555}@media(max-width:700px){#seo-snapshot header{display:block}#seo-snapshot nav{margin-top:16px}.highlights{grid-template-columns:1fr}}</style>`

function withSeo(html: string, page: SeoPage | null): string {
  const title = page?.title || '3API - AI API 中转站与多模型统一接入'
  const description = page?.description || '3API 为开发者提供 OpenAI、Claude、Gemini 等模型的统一 API 接入、智能路由、Token 用量统计和按量计费服务。'
  const path = page?.path || '/'
  const canonical = new URL(path, siteUrl).toString()
  let output = replaceTag(html, /<title>[^<]*<\/title>/i, `<title>${escapeHtml(title)}</title>`)
  output = replaceTag(output, /<meta\s+name=["']description["'][^>]*>/i, `<meta name="description" content="${escapeHtml(description)}" />`)
  output = replaceTag(output, /<link\s+rel=["']canonical["'][^>]*>/i, `<link rel="canonical" href="${canonical}" />`)
  output = replaceTag(output, /<meta\s+property=["']og:title["'][^>]*>/i, `<meta property="og:title" content="${escapeHtml(title)}" />`)
  output = replaceTag(output, /<meta\s+property=["']og:description["'][^>]*>/i, `<meta property="og:description" content="${escapeHtml(description)}" />`)
  output = replaceTag(output, /<meta\s+property=["']og:url["'][^>]*>/i, `<meta property="og:url" content="${canonical}" />`)
  output = output.replace('</head>', `${snapshotStyle}</head>`)
  return output.replace('<div id="app"></div>', `<div id="app">${page ? renderSeoBody(page) : renderHomeBody()}</div>`)
}

function generateSeoFiles(): Plugin {
  return {
    name: 'generate-seo-files',
    apply: 'build',
    closeBundle() {
      const outDir = resolve(__dirname, '../backend/internal/web/dist')
      const indexPath = resolve(outDir, 'index.html')
      const baseHtml = readFileSync(indexPath, 'utf8')
      writeFileSync(indexPath, withSeo(baseHtml, null))
      Object.values(seoPages).forEach((page) => {
        writeFileSync(resolve(outDir, `${page.path.slice(1)}.html`), withSeo(baseHtml, page))
      })
      const urls = ['/', ...Object.values(seoPages).map((page) => page.path)]
      const today = new Date().toISOString().slice(0, 10)
      const sitemap = `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls.map((path) => `  <url><loc>${new URL(path, siteUrl)}</loc><lastmod>${today}</lastmod></url>`).join('\n')}\n</urlset>\n`
      writeFileSync(resolve(outDir, 'sitemap.xml'), sitemap)
      writeFileSync(resolve(outDir, 'robots.txt'), `User-agent: *\nAllow: /\nDisallow: /admin/\nDisallow: /dashboard\nDisallow: /keys\nDisallow: /usage\nDisallow: /profile\nDisallow: /login\nDisallow: /register\nSitemap: ${siteUrl}/sitemap.xml\n`)
    },
  }
}

function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, (character) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  })[character] || character)
}

function isSafeImageUrl(value: string): boolean {
  const trimmed = value.trim()
  if ((trimmed.startsWith('/') && !trimmed.startsWith('//')) || /^data:image\//i.test(trimmed)) {
    return true
  }
  try {
    const parsed = new URL(trimmed)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

function injectBranding(html: string, config: { site_name?: string; site_logo?: string }): string {
  let brandedHtml = html
  const siteName = config.site_name?.trim()
  if (siteName) {
    brandedHtml = brandedHtml.replace(
      /<title>[^<]*<\/title>/i,
      `<title>${escapeHtml(siteName)} - AI API Gateway</title>`,
    )
  }

  const siteLogo = config.site_logo?.trim()
  if (siteLogo && isSafeImageUrl(siteLogo)) {
    brandedHtml = brandedHtml.replace(
      /<link\s+rel=["']icon["'][^>]*>/i,
      `<link rel="icon" href="${escapeHtml(siteLogo)}" />`,
    )
  }
  return brandedHtml
}

/**
 * Vite 插件：开发模式下注入公开配置到 index.html
 * 与生产模式的后端注入行为保持一致，消除闪烁
 */
function injectPublicSettings(backendUrl: string): Plugin {
  return {
    name: 'inject-public-settings',
    apply: 'serve',
    transformIndexHtml: {
      order: 'pre',
      async handler(html) {
        try {
          const response = await fetch(`${backendUrl}/api/v1/settings/public`, {
            signal: AbortSignal.timeout(2000)
          })
          if (response.ok) {
            const data = await response.json()
            if (data.code === 0 && data.data) {
              const script = `<script>window.__APP_CONFIG__=${JSON.stringify(data.data)};</script>`
              return injectBranding(html, data.data).replace('</head>', `${script}\n</head>`)
            }
          }
        } catch (e) {
          console.warn('[vite] 无法获取公开配置，将回退到 API 调用:', (e as Error).message)
        }
        return html
      }
    }
  }
}

export default defineConfig(({ mode }) => {
  // 加载环境变量
  const env = loadEnv(mode, process.cwd(), '')
  const backendUrl = env.VITE_DEV_PROXY_TARGET || 'http://localhost:8080'
  const devPort = Number(env.VITE_DEV_PORT || 3000)

  return {
    plugins: [
      vue(),
      checker({
        vueTsc: true
      }),
      injectPublicSettings(backendUrl),
      generateSeoFiles()
    ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      // 使用 vue-i18n 运行时版本，避免 CSP unsafe-eval 问题
      'vue-i18n': 'vue-i18n/dist/vue-i18n.runtime.esm-bundler.js'
    }
  },
  define: {
    // 启用 vue-i18n JIT 编译，在 CSP 环境下处理消息插值
    // JIT 编译器生成 AST 对象而非 JS 代码，无需 unsafe-eval
    __INTLIFY_JIT_COMPILATION__: true
  },
  build: {
    outDir: '../backend/internal/web/dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        /**
         * 手动分包配置
         * 分离第三方库并按功能合并应用代码，避免循环依赖
         */
        manualChunks(id: string) {
          if (id.includes('node_modules')) {
            // Vue 核心库
            if (
              id.includes('/vue/') ||
              id.includes('/vue-router/') ||
              id.includes('/pinia/') ||
              id.includes('/@vue/')
            ) {
              return 'vendor-vue'
            }

            // UI 工具库（较大，单独分离）
            if (id.includes('/@vueuse/')) {
              return 'vendor-ui'
            }

            // 图表库
            if (id.includes('/chart.js/') || id.includes('/vue-chartjs/')) {
              return 'vendor-chart'
            }

            // 国际化
            if (id.includes('/vue-i18n/') || id.includes('/@intlify/')) {
              return 'vendor-i18n'
            }

            // Stripe 仅在支付流程中按需加载，避免进入首页公共依赖。
            if (id.includes('/@stripe/stripe-js/')) {
              return 'vendor-stripe'
            }

            // 其他小型第三方库合并
            return 'vendor-misc'
          }

          // 应用代码：按入口点自动分包，不手动干预
          // 这样可以避免循环依赖，同时保持合理的 chunk 数量
        }
      }
    }
  },
    server: {
      host: '0.0.0.0',
      port: devPort,
      proxy: {
        // Match the API root and nested API paths without capturing public
        // SEO routes such as /api-relay.
        '^/api(?:/|$)': {
          target: backendUrl,
          changeOrigin: true
        },
        '/v1': {
          target: backendUrl,
          changeOrigin: true
        },
        '/setup': {
          target: backendUrl,
          changeOrigin: true
        }
      }
    }
  }
})
