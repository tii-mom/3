import axios from 'axios'
import type { ApiResponse } from '@/types'
import { getAPIBaseURL } from './url'

const CSRF_TOKEN_KEY = 'auth_csrf_token'
const LEGACY_REFRESH_TOKEN_KEY = 'refresh_token'
const TOKEN_EXPIRES_AT_KEY = 'token_expires_at'
const REFRESH_LOCK_NAME = 'sub2api-auth-refresh'

export interface BrowserSessionRefreshResponse {
  access_token: string
  refresh_token?: string
  csrf_token?: string
  expires_in: number
  token_type: string
}

let accessToken: string | null = null
let refreshPromise: Promise<BrowserSessionRefreshResponse> | null = null

export function getSessionAccessToken(): string | null {
  return accessToken
}

export function setSessionAccessToken(token: string | null): void {
  accessToken = token?.trim() || null
}

export function getSessionCSRFToken(): string | null {
  return sessionStorage.getItem(CSRF_TOKEN_KEY)
}

export function setSessionCSRFToken(token: string | null): void {
  if (token?.trim()) {
    sessionStorage.setItem(CSRF_TOKEN_KEY, token.trim())
    return
  }
  sessionStorage.removeItem(CSRF_TOKEN_KEY)
}

export function clearSessionAuth(): void {
  accessToken = null
  sessionStorage.removeItem(CSRF_TOKEN_KEY)
}

function unwrapAuthResponse<T>(response: unknown): T {
  const payload = response as ApiResponse<T> | T
  if (payload && typeof payload === 'object' && 'code' in payload) {
    const envelope = payload as ApiResponse<T>
    if (envelope.code !== 0 || !envelope.data) {
      throw Object.assign(new Error(envelope.message || 'Authentication request failed'), {
        code: envelope.code
      })
    }
    return envelope.data
  }
  return payload as T
}

export async function fetchBrowserCSRFToken(): Promise<string> {
  const response = await axios.get(`${getAPIBaseURL()}/auth/csrf`, {
    withCredentials: true,
    headers: { 'X-Auth-Transport': 'cookie' },
    timeout: 30000
  })
  const data = unwrapAuthResponse<{ csrf_token: string }>(response.data)
  const csrfToken = data.csrf_token?.trim()
  if (!csrfToken) {
    throw new Error('Browser session did not return a CSRF token')
  }
  setSessionCSRFToken(csrfToken)
  return csrfToken
}

async function performBrowserSessionRefresh(): Promise<BrowserSessionRefreshResponse> {
  const legacyRefreshToken =
    sessionStorage.getItem(LEGACY_REFRESH_TOKEN_KEY) ||
    localStorage.getItem(LEGACY_REFRESH_TOKEN_KEY)

  let csrfToken: string | null = null
  try {
    // Fetching the cookie-paired value on every refresh also repairs stale
    // per-tab CSRF state after another tab rotates the shared refresh cookie.
    csrfToken = await fetchBrowserCSRFToken()
  } catch (error) {
    if (!legacyRefreshToken) {
      throw error
    }
  }

  const response = await axios.post(
    `${getAPIBaseURL()}/auth/refresh`,
    legacyRefreshToken ? { refresh_token: legacyRefreshToken } : {},
    {
      withCredentials: true,
      headers: {
        'Content-Type': 'application/json',
        'X-Auth-Transport': 'cookie',
        ...(csrfToken ? { 'X-CSRF-Token': csrfToken } : {})
      },
      timeout: 30000
    }
  )
  const data = unwrapAuthResponse<BrowserSessionRefreshResponse>(response.data)

  setSessionAccessToken(data.access_token)
  setSessionCSRFToken(data.csrf_token || null)
  localStorage.removeItem('auth_token')
  localStorage.removeItem(LEGACY_REFRESH_TOKEN_KEY)
  localStorage.removeItem(TOKEN_EXPIRES_AT_KEY)
  sessionStorage.removeItem(LEGACY_REFRESH_TOKEN_KEY)
  sessionStorage.setItem(
    TOKEN_EXPIRES_AT_KEY,
    String(Date.now() + data.expires_in * 1000)
  )

  return data
}

async function withCrossTabRefreshLock<T>(callback: () => Promise<T>): Promise<T> {
  if (typeof navigator !== 'undefined' && navigator.locks) {
    return navigator.locks.request(REFRESH_LOCK_NAME, callback)
  }
  return callback()
}

export function refreshBrowserSession(): Promise<BrowserSessionRefreshResponse> {
  if (!refreshPromise) {
    refreshPromise = withCrossTabRefreshLock(performBrowserSessionRefresh).finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}
