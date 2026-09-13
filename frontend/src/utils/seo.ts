import type { RouteLocationNormalizedLoaded } from 'vue-router'
import seoPages from '@/content/seo-pages.json'

const SITE_URL = 'https://3api.shop'
// 首页 meta 与 frontend/index.html 源文件保持一致（销售业务定位）。
// 运行时 updateRouteSeo 会覆盖 document.title，这里若写成旧的「AI API 中转站」文案，
// 会把预渲染好的正确标题盖掉（浏览器标签/分享卡片全错）。
const DEFAULT_TITLE = '3API - ChatGPT Plus/Pro 代充值与成品号独享账号'
const DEFAULT_DESCRIPTION = '3API 提供 ChatGPT Plus / Pro 官方直充、独享成品号与租号服务，支付宝支付，支付后提交 Session 即可到账，全程人工质保。'

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
  // 首页与 SEO 落地页的标题由本函数负责；其余路由（控制台、登录、法务、404…）的标题
  // 已由 router guard 的 resolveRouteDocumentTitle 统一设置（支持 titleKey 本地化）。
  // 这里若再用 route.meta.title 覆盖，会把「登录 - 3API」盖成英文静态 "Login"，
  // 也会让 tab / 分享卡片标题丢失本地化。
  const seoTitle = page?.title || (isHome ? DEFAULT_TITLE : '')
  if (seoTitle) {
    document.title = seoTitle
  }
  const title = seoTitle || document.title || '3API'
  const description = page?.description || (isHome ? DEFAULT_DESCRIPTION : '3API 用户与管理控制台')
  const canonicalPath = page?.path || (isHome ? '/' : route.path)
  const canonical = new URL(canonicalPath, SITE_URL).toString()

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
  // 与预渲染 index.html / SEO 快照保持一致：大图卡片。写成 summary 会把服务端
  // 已渲染好的 summary_large_image 降级，分享时缩略图变小。
  upsertMeta('meta[name="twitter:card"]', { name: 'twitter:card', content: 'summary_large_image' })
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

  if (isHome) {
    addStructuredData({
      '@context': 'https://schema.org',
      '@type': 'WebSite',
      name: '3API',
      url: SITE_URL,
    })
  }

  if (page) {
    addStructuredData({
      '@context': 'https://schema.org',
      '@type': 'BreadcrumbList',
      itemListElement: [
        { '@type': 'ListItem', position: 1, name: '3API', item: SITE_URL },
        { '@type': 'ListItem', position: 2, name: page.heading, item: canonical },
      ],
    })

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
