export function shouldBootstrapAuthForRoute(
  path: string,
  meta: Record<string, unknown>,
  backendModeEnabled: boolean
): boolean {
  if (backendModeEnabled) {
    return true
  }
  if (meta.requiresAuth !== false) {
    return true
  }
  return path === '/login' || path === '/register' || path === '/setup'
}

type AuthBootstrapStorage = Pick<Storage, 'getItem'>

function hasStorageValue(storage: AuthBootstrapStorage | null | undefined, key: string): boolean {
  try {
    return Boolean(storage?.getItem(key))
  } catch {
    return false
  }
}

export function hasAuthBootstrapStorageHint(
  local: AuthBootstrapStorage | null | undefined,
  session: AuthBootstrapStorage | null | undefined
): boolean {
  return (
    hasStorageValue(local, 'auth_user') ||
    hasStorageValue(local, 'refresh_token') ||
    hasStorageValue(local, 'token_expires_at') ||
    hasStorageValue(session, 'refresh_token') ||
    hasStorageValue(session, 'token_expires_at')
  )
}

export function hasBrowserAuthBootstrapHint(): boolean {
  if (typeof window === 'undefined') {
    return false
  }
  return hasAuthBootstrapStorageHint(window.localStorage, window.sessionStorage)
}
