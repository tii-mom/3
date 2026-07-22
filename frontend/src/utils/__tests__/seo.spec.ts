import { beforeEach, describe, expect, it } from 'vitest'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import { getSeoPage, updateRouteSeo } from '@/utils/seo'

function route(path: string, meta: Record<string, unknown>): RouteLocationNormalizedLoaded {
  return { path, fullPath: path, hash: '', query: {}, params: {}, matched: [], name: undefined, redirectedFrom: undefined, meta } as RouteLocationNormalizedLoaded
}

describe('SEO metadata', () => {
  beforeEach(() => {
    document.head.innerHTML = ''
  })

  it('provides complete content for every public SEO page', () => {
    const page = getSeoPage('openai-api')

    expect(page?.path).toBe('/openai-api')
    expect(page?.title).toContain('OpenAI API')
    expect(page?.faqs.length).toBeGreaterThanOrEqual(3)
  })

  it('sets canonical and index directives for public pages', () => {
    updateRouteSeo(route('/openai-api', { title: 'OpenAI API', seoKey: 'openai-api' }))

    expect(document.title).toContain('OpenAI API')
    expect(document.querySelector('link[rel="canonical"]')?.getAttribute('href')).toBe('https://3api.shop/openai-api')
    expect(document.querySelector('meta[name="robots"]')?.getAttribute('content')).toContain('index,follow')
    expect(document.querySelectorAll('script[data-seo-structured-data]')).toHaveLength(2)
  })

  it('keeps authenticated application routes out of search indexes', () => {
    updateRouteSeo(route('/dashboard', { title: 'Dashboard', requiresAuth: true }))

    expect(document.querySelector('meta[name="robots"]')?.getAttribute('content')).toBe('noindex,nofollow')
  })
})
