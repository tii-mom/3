import type { RouteLocationRaw, Router } from 'vue-router'

export const CHUNK_RELOAD_QUERY = '__chunk_reload'
export const CHUNK_RELOAD_STORAGE_KEY = 'chunk_reload_attempted'

const CHUNK_RELOAD_COOLDOWN_MS = 10_000
let inMemoryLastReload: number | null = null

function getSessionStorage(): Storage | null {
  if (typeof window === 'undefined') return null

  try {
    return window.sessionStorage
  } catch {
    return null
  }
}

function getCleanUrl(url: URL): string {
  return `${url.pathname}${url.search}${url.hash}`
}

export function isChunkLoadError(error: unknown): boolean {
  const candidate = error as { message?: unknown; name?: unknown } | null
  const message = typeof candidate?.message === 'string' ? candidate.message : ''
  const name = typeof candidate?.name === 'string' ? candidate.name : ''

  return (
    message.includes('Failed to fetch dynamically imported module') ||
    message.includes('Loading chunk') ||
    message.includes('Loading CSS chunk') ||
    name === 'ChunkLoadError'
  )
}

export function buildChunkReloadUrl(href: string, timestamp: number): string {
  const url = new URL(href, window.location.origin)
  url.searchParams.set(CHUNK_RELOAD_QUERY, String(timestamp))
  return getCleanUrl(url)
}

export function clearChunkReloadQuery(): void {
  if (typeof window === 'undefined') return

  const url = new URL(window.location.href)
  if (!url.searchParams.has(CHUNK_RELOAD_QUERY)) return

  url.searchParams.delete(CHUNK_RELOAD_QUERY)
  window.history.replaceState(window.history.state, document.title, getCleanUrl(url))
}

export function clearChunkReloadAttempt(): void {
  try {
    getSessionStorage()?.removeItem(CHUNK_RELOAD_STORAGE_KEY)
  } catch {
    // Ignore storage cleanup failures so successful navigation remains intact.
  }
  inMemoryLastReload = null
}

export function reloadForChunkError(
  router: Router,
  target: RouteLocationRaw,
  now = Date.now(),
  reload: (url: string) => void = (url) => window.location.replace(url)
): boolean {
  const storage = getSessionStorage()
  let lastReload: number | null = inMemoryLastReload

  try {
    const storedValue = storage?.getItem(CHUNK_RELOAD_STORAGE_KEY)
    if (storedValue) {
      const parsedValue = Number(storedValue)
      if (Number.isFinite(parsedValue)) lastReload = parsedValue
    }
  } catch {
    // Storage may be unavailable in privacy-restricted browser contexts.
  }

  if (lastReload !== null && now - lastReload <= CHUNK_RELOAD_COOLDOWN_MS) {
    return false
  }

  inMemoryLastReload = now
  try {
    storage?.setItem(CHUNK_RELOAD_STORAGE_KEY, String(now))
  } catch {
    // The in-memory timestamp still prevents a reload loop for this page.
  }

  const targetHref = router.resolve(target).href
  reload(buildChunkReloadUrl(targetHref, now))
  return true
}
