export interface OAuthCallbackParams {
  code?: string
  state?: string
}

const NESTED_PARAM_HINTS = ['redirect', 'redirect_uri', 'callback', 'url', 'continue', 'next']

const safeDecode = (value: string) => {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

const decodeRepeatedly = (value: string) => {
  const decoded = [value]
  let current = value
  for (let i = 0; i < 3; i += 1) {
    const next = safeDecode(current)
    if (next === current) break
    decoded.push(next)
    current = next
  }
  return decoded
}

const normalizeUrlCandidate = (value: string) => {
  if (/^[a-z][a-z0-9+.-]*:\/\//i.test(value)) return value
  if (value.startsWith('/')) return `http://localhost${value}`
  if (/^[\w.-]+(?::\d+)?(?:[/?#]|$)/.test(value) && value.includes('?')) {
    return `http://${value}`
  }
  return null
}

const readFromSearchParams = (params: URLSearchParams): OAuthCallbackParams => {
  const code = params.get('code') || undefined
  const state = params.get('state') || undefined
  return {
    code: code ? safeDecode(code) : undefined,
    state: state ? safeDecode(state) : undefined
  }
}

const readByRegex = (value: string): OAuthCallbackParams => {
  const codeMatch = value.match(/(?:^|[?#&])code=([^&#\s]+)/)
  const stateMatch = value.match(/(?:^|[?#&])state=([^&#\s]+)/)
  return {
    code: codeMatch?.[1] ? safeDecode(codeMatch[1]) : undefined,
    state: stateMatch?.[1] ? safeDecode(stateMatch[1]) : undefined
  }
}

const enqueueNestedValues = (params: URLSearchParams, enqueue: (value: string) => void) => {
  for (const [key, value] of params.entries()) {
    const lowerKey = key.toLowerCase()
    const decodedValues = decodeRepeatedly(value)
    if (
      NESTED_PARAM_HINTS.includes(lowerKey) ||
      decodedValues.some((item) => item.includes('code=') || item.includes('state=') || item.includes('/callback'))
    ) {
      decodedValues.forEach(enqueue)
    }
  }
}

export function extractOAuthCallbackParams(input: string): OAuthCallbackParams {
  const initial = input.trim()
  if (!initial) return {}

  const queue = decodeRepeatedly(initial)
  const seen = new Set<string>()
  let fallbackState: string | undefined

  const enqueue = (value: string) => {
    const trimmed = value.trim()
    if (!trimmed || seen.has(trimmed) || queue.includes(trimmed)) return
    queue.push(trimmed)
  }

  while (queue.length > 0 && seen.size < 50) {
    const candidate = queue.shift()!.trim()
    if (!candidate || seen.has(candidate)) continue
    seen.add(candidate)

    const regexParams = readByRegex(candidate)
    if (regexParams.state) fallbackState = regexParams.state
    if (regexParams.code) {
      return { code: regexParams.code, state: regexParams.state || fallbackState }
    }

    const urlCandidate = normalizeUrlCandidate(candidate)
    if (urlCandidate) {
      try {
        const url = new URL(urlCandidate)
        const searchParams = readFromSearchParams(url.searchParams)
        if (searchParams.state) fallbackState = searchParams.state
        if (searchParams.code) {
          return { code: searchParams.code, state: searchParams.state || fallbackState }
        }
        enqueueNestedValues(url.searchParams, enqueue)

        const hash = url.hash.replace(/^#/, '')
        if (hash) {
          enqueue(hash)
          const hashParams = new URLSearchParams(hash.replace(/^\?/, ''))
          const parsedHash = readFromSearchParams(hashParams)
          if (parsedHash.state) fallbackState = parsedHash.state
          if (parsedHash.code) {
            return { code: parsedHash.code, state: parsedHash.state || fallbackState }
          }
          enqueueNestedValues(hashParams, enqueue)
        }
      } catch {
        // Continue with raw query parsing below.
      }
    }

    const query = candidate.includes('?') ? candidate.slice(candidate.indexOf('?') + 1) : candidate.replace(/^[?#]/, '')
    if (query.includes('=')) {
      const params = new URLSearchParams(query)
      const queryParams = readFromSearchParams(params)
      if (queryParams.state) fallbackState = queryParams.state
      if (queryParams.code) {
        return { code: queryParams.code, state: queryParams.state || fallbackState }
      }
      enqueueNestedValues(params, enqueue)
    }
  }

  return fallbackState ? { state: fallbackState } : {}
}
