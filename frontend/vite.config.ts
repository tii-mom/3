import { defineConfig, loadEnv, Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import checker from 'vite-plugin-checker'
import { resolve } from 'path'
import { readFileSync, writeFileSync } from 'node:fs'
import seoPages from './src/content/seo-pages.json'

const siteUrl = 'https://3api.shop'

type SeoPage = (typeof seoPages)[keyof typeof seoPages]

const publicSeoNav = [
  { path: '/api-relay', label: 'API 中转站' },
  { path: '/claude-code-api', label: 'Claude Code' },
  { path: '/openai-api', label: 'OpenAI API' },
  { path: '/codex-api', label: 'Codex API' },
  { path: '/cursor-api-key', label: 'Cursor' },
  { path: '/token-guide', label: 'Token 指南' },
]

const publicSeoRelatedLinks = [
  { path: '/claude-code-api', label: 'Claude Code 国内使用', text: 'API Key、Base URL 与模型配置要点' },
  { path: '/codex-api', label: 'Codex API 配置', text: 'AI 编程客户端接入与用量追踪' },
  { path: '/cursor-api-key', label: 'Cursor API Key', text: 'Cursor 接入 Claude 和 OpenAI 兼容模型' },
  { path: '/api-base-url', label: 'Base URL 怎么填', text: '接口地址、密钥和模型名称的区别' },
  { path: '/openai-api-domestic-payment', label: 'OpenAI API 国内接入', text: '充值、订阅和 API 调用的差异' },
  { path: '/token-guide', label: 'Token 计费指南', text: '输入、输出和缓存 Token 成本控制' },
]

function replaceTag(html: string, pattern: RegExp, replacement: string): string {
  return pattern.test(html) ? html.replace(pattern, replacement) : html.replace('</head>', `  ${replacement}\n</head>`)
}

function renderSeoBody(page: SeoPage): string {
  const highlights = page.highlights.map((item) => `<article><h2>${escapeHtml(item.title)}</h2><p>${escapeHtml(item.text)}</p></article>`).join('')
  const sections = page.sections.map((section) => `<section><h2>${escapeHtml(section.title)}</h2><p>${escapeHtml(section.body)}</p></section>`).join('')
  const faqs = page.faqs.map((faq) => `<details><summary>${escapeHtml(faq.question)}</summary><p>${escapeHtml(faq.answer)}</p></details>`).join('')
  const nav = publicSeoNav.map((item) => `<a href="${item.path}">${escapeHtml(item.label)}</a>`).join('')
  const related = publicSeoRelatedLinks
    .filter((item) => item.path !== page.path)
    .slice(0, 4)
    .map((item) => `<a href="${item.path}"><strong>${escapeHtml(item.label)}</strong><span>${escapeHtml(item.text)}</span></a>`)
    .join('')
  return `<div id="seo-snapshot"><header><a href="/">3API</a><nav>${nav}</nav></header><main><p class="eyebrow">${escapeHtml(page.eyebrow)}</p><h1>${escapeHtml(page.heading)}</h1><p class="summary">${escapeHtml(page.summary)}</p><a class="cta" href="/register">${escapeHtml(page.primaryCta)}</a><div class="highlights">${highlights}</div><div class="content">${sections}</div><section class="faq"><h2>常见问题</h2>${faqs}</section><section class="related"><h2>相关接入指南</h2>${related}</section></main><footer>3API · 独立第三方 AI API 接入平台</footer></div>`
}

const snapshotStyle = `<style id="seo-snapshot-style">#seo-snapshot{font-family:system-ui,-apple-system,"PingFang SC",sans-serif;max-width:1120px;margin:auto;padding:24px;color:#171717}#seo-snapshot header{display:flex;justify-content:space-between;gap:24px;padding:12px 0;border-bottom:1px solid #ddd}#seo-snapshot nav{display:flex;gap:18px;flex-wrap:wrap}#seo-snapshot a{color:#b94300}#seo-snapshot main{padding:72px 0}#seo-snapshot .eyebrow{font-size:12px;color:#c84d0b}#seo-snapshot h1{font-size:clamp(40px,7vw,72px);line-height:1.1;max-width:900px}#seo-snapshot .summary{font-size:19px;line-height:1.7;max-width:780px}#seo-snapshot .cta{display:inline-block;margin:20px 0 50px;padding:13px 20px;background:#e85d11;color:white;text-decoration:none}.highlights{display:grid;grid-template-columns:repeat(3,1fr);gap:1px;background:#ccc}.highlights article{padding:24px;background:#fff}.content{margin-top:70px}.content section,.faq details{padding:24px 0;border-top:1px solid #ddd}.content p,.faq p,.related span{line-height:1.7;color:#555}.related{margin-top:70px;padding-top:24px;border-top:1px solid #ddd}.related a{display:block;margin:14px 0}.related strong{display:block;color:#171717}.related span{display:block}@media(max-width:700px){#seo-snapshot header{display:block}#seo-snapshot nav{margin-top:16px}.highlights{grid-template-columns:1fr}}</style>`

function renderStructuredData(page: SeoPage | null): string {
  const organization = {
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: '3API',
    url: siteUrl,
    description: '3API 为国内开发者提供 OpenAI、Claude、Gemini 等模型的统一 API 接入，适配 Claude Code、Codex、Cursor、Agent 和应用开发场景。',
  }
  const website = {
    '@context': 'https://schema.org',
    '@type': 'WebSite',
    name: '3API',
    url: siteUrl,
  }
  const data: object[] = page ? [
    organization,
    {
      '@context': 'https://schema.org',
      '@type': 'BreadcrumbList',
      itemListElement: [
        { '@type': 'ListItem', position: 1, name: '3API', item: siteUrl },
        { '@type': 'ListItem', position: 2, name: page.heading, item: new URL(page.path, siteUrl).toString() },
      ],
    },
    {
      '@context': 'https://schema.org',
      '@type': 'FAQPage',
      mainEntity: page.faqs.map((faq) => ({
        '@type': 'Question',
        name: faq.question,
        acceptedAnswer: { '@type': 'Answer', text: faq.answer },
      })),
    },
  ] : [organization, website]

  return data.map((item) => `<script type="application/ld+json" data-seo-structured-data="true">${JSON.stringify(item).replace(/</g, '\\u003c')}</script>`).join('')
}

function withSeoMetadata(html: string, page: SeoPage | null): string {
  const title = page?.title || '3API - AI API 中转站与多模型统一接入'
  const description = page?.description || '3API 为国内开发者提供 OpenAI、Claude、Gemini 等模型的统一 API 接入，适配 Claude Code、Codex、Cursor、Agent 和应用开发场景。'
  const path = page?.path || '/'
  const canonical = new URL(path, siteUrl).toString()
  let output = replaceTag(html, /<title>[^<]*<\/title>/i, `<title>${escapeHtml(title)}</title>`)
  output = replaceTag(output, /<meta\s+name=["']description["'][^>]*>/i, `<meta name="description" content="${escapeHtml(description)}" />`)
  output = replaceTag(output, /<link\s+rel=["']canonical["'][^>]*>/i, `<link rel="canonical" href="${canonical}" />`)
  output = replaceTag(output, /<meta\s+property=["']og:title["'][^>]*>/i, `<meta property="og:title" content="${escapeHtml(title)}" />`)
  output = replaceTag(output, /<meta\s+property=["']og:description["'][^>]*>/i, `<meta property="og:description" content="${escapeHtml(description)}" />`)
  output = replaceTag(output, /<meta\s+property=["']og:url["'][^>]*>/i, `<meta property="og:url" content="${canonical}" />`)
  output = output.replace('</head>', `${renderStructuredData(page)}</head>`)
  return output
}

function withSeoSnapshot(html: string, page: SeoPage): string {
  let output = withSeoMetadata(html, page)
  output = output.replace('</head>', `${snapshotStyle}</head>`)
  return output.replace('<div id="app"></div>', `<div id="app">${renderSeoBody(page)}</div>`)
}

function generateSeoFiles(): Plugin {
  return {
    name: 'generate-seo-files',
    apply: 'build',
    closeBundle() {
      const outDir = resolve(__dirname, '../backend/internal/web/dist')
      const indexPath = resolve(outDir, 'index.html')
      const baseHtml = readFileSync(indexPath, 'utf8')
      // Keep the public homepage as a clean Vue shell. Cloudflare Pages serves
      // index.html directly, so embedding a visible SEO snapshot here causes a
      // flash before hydration for real users. Dedicated SEO landing pages
      // still receive static snapshots below.
      writeFileSync(indexPath, withSeoMetadata(baseHtml, null))
      Object.values(seoPages).forEach((page) => {
        writeFileSync(resolve(outDir, `${page.path.slice(1)}.html`), withSeoSnapshot(baseHtml, page))
      })
      const urls = ['/', ...Object.values(seoPages).map((page) => page.path)]
      const today = new Date().toISOString().slice(0, 10)
      const sitemap = `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls.map((path) => `  <url><loc>${new URL(path, siteUrl)}</loc><lastmod>${today}</lastmod></url>`).join('\n')}\n</urlset>\n`
      writeFileSync(resolve(outDir, 'sitemap.xml'), sitemap)
      writeFileSync(resolve(outDir, 'robots.txt'), `User-agent: *\nAllow: /\nDisallow: /admin/\nDisallow: /console\nDisallow: /dashboard\nDisallow: /keys\nDisallow: /usage\nDisallow: /profile\nDisallow: /login\nDisallow: /register\nDisallow: /purchase\nDisallow: /shop\nDisallow: /orders\nDisallow: /payment\nSitemap: ${siteUrl}/sitemap.xml\n`)
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
