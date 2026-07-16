const DEFAULT_API_BASE_URL = '/api/v1'
const API_BASE_URL = normalizeAPIBaseURL(import.meta.env.VITE_API_BASE_URL)

function normalizePath(path: string): string {
  return path.startsWith('/') ? path : `/${path}`
}

function normalizeAPIBaseURL(value: unknown): string {
  const raw = String(value || DEFAULT_API_BASE_URL).trim() || DEFAULT_API_BASE_URL
  const withoutTrailingSlash = raw.replace(/\/+$/, '')
  if (/^[a-z][a-z\d+.-]*:\/\//i.test(withoutTrailingSlash) || withoutTrailingSlash.startsWith('//')) {
    return withoutTrailingSlash
  }
  return normalizePath(withoutTrailingSlash)
}

export function getAPIBaseURL(): string {
  if (typeof window !== 'undefined') {
    const host = window.location?.hostname || ''
    // 当运行在 Cloudflare Pages 静态域名下时，将 API 请求分发至腾讯云 API 接口域
    if (
      host === '3api.shop' ||
      host === 'api.3api.shop' ||
      host === 'sub2api-cj7.pages.dev' ||
      host.endsWith('.pages.dev')
    ) {
      return 'https://api.3api.shop/api/v1'
    }
  }
  return API_BASE_URL
}

/**
 * Resolve the axios baseURL for admin/user JSON APIs.
 *
 * Site setting `api_base_url` is often the public gateway origin for client tools
 * (e.g. `https://api.3api.shop` for Claude Code / OpenAI clients) and may omit
 * `/api/v1`. The SPA client must always call the versioned admin/user API path.
 * Without this, requests hit the SPA shell HTML and crash settings pages.
 */
export function resolveApiClientBaseURL(configured?: string | null): string {
  const fallback = getAPIBaseURL()
  const raw = String(configured || '').trim().replace(/\/+$/, '')
  if (!raw) {
    return fallback
  }

  // Relative base
  if (raw.startsWith('/')) {
    if (raw === DEFAULT_API_BASE_URL || raw.startsWith(`${DEFAULT_API_BASE_URL}/`)) {
      return DEFAULT_API_BASE_URL
    }
    return fallback
  }

  // Absolute / protocol-relative
  if (/^[a-z][a-z\d+.-]*:\/\//i.test(raw) || raw.startsWith('//')) {
    try {
      const absolute = raw.startsWith('//') ? `https:${raw}` : raw
      const url = new URL(absolute)
      const path = url.pathname.replace(/\/+$/, '') || '/'
      if (path === DEFAULT_API_BASE_URL || path.startsWith(`${DEFAULT_API_BASE_URL}/`)) {
        return `${url.origin}${DEFAULT_API_BASE_URL}`
      }
      if (path === '/' ) {
        return `${url.origin}${DEFAULT_API_BASE_URL}`
      }
      // Custom prefix mount, e.g. https://example.com/sub2api
      return `${url.origin}${path}${DEFAULT_API_BASE_URL}`
    } catch {
      return fallback
    }
  }

  return fallback
}

export function buildApiUrl(path: string): string {
  const base = getAPIBaseURL().replace(/\/+$/, '')
  let suffix = normalizePath(path)
  if (suffix === DEFAULT_API_BASE_URL) {
    suffix = ''
  } else if (suffix.startsWith(`${DEFAULT_API_BASE_URL}/`)) {
    suffix = suffix.slice(DEFAULT_API_BASE_URL.length)
  }
  return `${base}${suffix}`
}

export function buildGatewayUrl(path: string): string {
  const suffix = normalizePath(path)
  try {
    const origin =
      typeof window === 'undefined'
        ? new URL(getAPIBaseURL()).origin
        : new URL(getAPIBaseURL(), window.location.origin).origin
    return `${origin}${suffix}`
  } catch {
    return suffix
  }
}
