import { describe, expect, it } from 'vitest'
import { extractOAuthCallbackParams } from '../oauthCallback'

describe('extractOAuthCallbackParams', () => {
  it('extracts code and state from a full callback URL', () => {
    expect(extractOAuthCallbackParams('http://localhost:8085/callback?code=abc123&state=state-1')).toEqual({
      code: 'abc123',
      state: 'state-1'
    })
  })

  it('extracts code and state from a schemeless localhost URL', () => {
    expect(extractOAuthCallbackParams('localhost:8085/callback?state=state-2&code=xyz')).toEqual({
      code: 'xyz',
      state: 'state-2'
    })
  })

  it('extracts code and state from a raw query string', () => {
    expect(extractOAuthCallbackParams('?code=raw-code&state=raw-state')).toEqual({
      code: 'raw-code',
      state: 'raw-state'
    })
  })

  it('extracts code and state from an encoded nested redirect', () => {
    const input = 'localhost:8085/login?redirect=%2Fcallback%3Fstate%3Dnested-state%2526code%253Dnested-code'

    expect(extractOAuthCallbackParams(input)).toEqual({
      code: 'nested-code',
      state: 'nested-state'
    })
  })

  it('does not mistake generated auth URL code_challenge for an auth code', () => {
    const input = 'https://accounts.google.com/o/oauth2/v2/auth?code_challenge=challenge&state=auth-state'

    expect(extractOAuthCallbackParams(input)).toEqual({
      state: 'auth-state'
    })
  })
})
