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
