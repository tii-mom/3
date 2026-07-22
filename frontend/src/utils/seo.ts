import type { RouteLocationNormalizedLoaded } from 'vue-router'
import seoPages from '@/content/seo-pages.json'

const SITE_URL = 'https://3api.shop'
const DEFAULT_TITLE = '3API - AI API 中转站与多模型统一接入'
const DEFAULT_DESCRIPTION = '3API 为开发者提供 OpenAI、Claude、Gemini 等模型的统一 API 接入、智能路由、Token 用量统计和按量计费服务。'

type SeoPage = (typeof seoPages)[keyof typeof seoPages]

function upsertMeta(selector: string, attributes: Record<string, string>) {
  let element = document.head.querySelector<HTMLMetaElement>(selector)
  if (!element) {
    element = document.createElement('meta')
    document.head.appendChild(element)
  }
  Object.entries(attributes).forEach(([name, value]) => element?.setAttribute(name, value))
}
function upsertLink(rel: string, href: string) {
  let element = document.head.querySelector<HTMLLinkElement>(`link[rel="${rel}"]`)
  if (!element) {
    element = document.createElement('link')
    element.rel = rel
    document.head.appendChild(element)
  }
  element.href = href
}

function removeStructuredData() {
  document.head.querySelectorAll('script[data-seo-structured-data]').forEach((node) => node.remove())
}

function addStructuredData(data: object) {
  const script = document.createElement('script')
  script.type = 'application/ld+json'
  script.dataset.seoStructuredData = 'true'
  script.textContent = JSON.stringify(data)
  document.head.appendChild(script)
}

export function getSeoPage(key: string | undefined): SeoPage | undefined {
  return key ? (seoPages as Record<string, SeoPage>)[key] : undefined
}

export function updateRouteSeo(route: RouteLocationNormalizedLoaded) {
  const page = getSeoPage(route.meta.seoKey)
  const isHome = route.path === '/'
  const shouldIndex = isHome || Boolean(page)
  const title = page?.title || (isHome ? DEFAULT_TITLE : String(route.meta.title || '3API'))
  const description = page?.description || (isHome ? DEFAULT_DESCRIPTION : '3API 用户与管理控制台')
  const canonicalPath = page?.path || (isHome ? '/' : route.path)
  const canonical = new URL(canonicalPath, SITE_URL).toString()

  document.title = title
  upsertMeta('meta[name="description"]', { name: 'description', content: description })
  upsertMeta('meta[name="robots"]', {
    name: 'robots',
    content: shouldIndex ? 'index,follow,max-image-preview:large' : 'noindex,nofollow',
  })
  upsertMeta('meta[property="og:title"]', { property: 'og:title', content: title })
  upsertMeta('meta[property="og:description"]', { property: 'og:description', content: description })
  upsertMeta('meta[property="og:type"]', { property: 'og:type', content: 'website' })
  upsertMeta('meta[property="og:url"]', { property: 'og:url', content: canonical })
  upsertMeta('meta[property="og:site_name"]', { property: 'og:site_name', content: '3API' })
  upsertMeta('meta[name="twitter:card"]', { name: 'twitter:card', content: 'summary' })
  upsertLink('canonical', canonical)

  removeStructuredData()
  if (!shouldIndex) return

  addStructuredData({
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: '3API',
    url: SITE_URL,
    description: DEFAULT_DESCRIPTION,
  })

  if (page) {
    addStructuredData({
      '@context': 'https://schema.org',
      '@type': 'FAQPage',
      mainEntity: page.faqs.map((faq) => ({
        '@type': 'Question',
        name: faq.question,
        acceptedAnswer: { '@type': 'Answer', text: faq.answer },
      })),
    })
  }
}
