import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post } = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn()
}))

vi.mock('axios', () => ({
  default: { get, post }
}))

describe('browser auth session', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
    get.mockReset()
    post.mockReset()
  })

  it('deduplicates refreshes and keeps access credentials out of localStorage', async () => {
    get.mockResolvedValue({ data: { code: 0, data: { csrf_token: 'csrf-before' } } })
    post.mockResolvedValue({
      data: {
        code: 0,
        data: {
          access_token: 'access-after',
          csrf_token: 'csrf-after',
          expires_in: 3600,
          token_type: 'Bearer'
        }
      }
    })
    const session = await import('@/api/authSession')

    const [first, second] = await Promise.all([
      session.refreshBrowserSession(),
      session.refreshBrowserSession()
    ])

    expect(first).toEqual(second)
    expect(get).toHaveBeenCalledTimes(1)
    expect(post).toHaveBeenCalledTimes(1)
    expect(session.getSessionAccessToken()).toBe('access-after')
    expect(sessionStorage.getItem('auth_csrf_token')).toBe('csrf-after')
    expect(localStorage.getItem('auth_token')).toBeNull()
    expect(localStorage.getItem('refresh_token')).toBeNull()
  })

  it('uses a legacy refresh token once when no browser cookie exists', async () => {
    sessionStorage.setItem('refresh_token', 'legacy-refresh')
    get.mockRejectedValue(new Error('No cookie'))
    post.mockResolvedValue({
      data: {
        code: 0,
        data: {
          access_token: 'migrated-access',
          csrf_token: 'migrated-csrf',
          expires_in: 1800,
          token_type: 'Bearer'
        }
      }
    })
    const session = await import('@/api/authSession')

    await session.refreshBrowserSession()

    expect(post).toHaveBeenCalledWith(
      '/api/v1/auth/refresh',
      { refresh_token: 'legacy-refresh' },
      expect.objectContaining({ withCredentials: true })
    )
    expect(sessionStorage.getItem('refresh_token')).toBeNull()
    expect(session.getSessionAccessToken()).toBe('migrated-access')
  })
})
