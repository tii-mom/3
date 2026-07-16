import { describe, expect, it } from 'vitest'

import { isTotp2FARequired } from '@/api/auth'

describe('login response guards', () => {
  it('recognizes a two-factor authentication challenge', () => {
    expect(
      isTotp2FARequired({
        requires_2fa: true,
        temp_token: 'temporary-token'
      })
    ).toBe(true)
  })

  it('does not throw when an upstream proxy returns an HTML string', () => {
    expect(() => isTotp2FARequired('<!doctype html>' as never)).not.toThrow()
    expect(isTotp2FARequired('<!doctype html>' as never)).toBe(false)
  })
})
