import { buildGatewayUrl } from './client'

export interface GatewayModel {
  id: string
  object?: string
  created?: number
  owned_by?: string
}

export interface GatewayModelsResponse {
  object?: string
  data?: GatewayModel[]
}

export interface ChatTestMessage {
  role: 'system' | 'user' | 'assistant'
  content: string
}

export interface ChatTestRequest {
  apiKey: string
  model: string
  messages: ChatTestMessage[]
  temperature?: number
  maxTokens?: number
}

export interface ChatTestResult {
  content: string
  firstTokenMs: number | null
  durationMs: number
  model: string
  requestId: string
}

export interface ImageGenerationRequest {
  apiKey: string
  model?: string
  prompt: string
  size?: string
  quality?: string
  outputFormat?: string
  n?: number
}

export interface GeneratedImage {
  url: string
  mimeType: string
  revisedPrompt?: string
}

export interface ImageGenerationResult {
  images: GeneratedImage[]
  durationMs: number
  requestId: string
  raw: unknown
}

function authHeaders(apiKey: string, extra?: HeadersInit): HeadersInit {
  return {
    Authorization: `Bearer ${apiKey}`,
    ...extra,
  }
}

async function parseGatewayError(response: Response): Promise<Error> {
  const requestId = response.headers.get('X-Request-Id') || ''
  const fallbackMessage = response.statusText || `HTTP ${response.status}`
  try {
    const text = await response.text()
    let body: any = null
    try {
      body = text ? JSON.parse(text) : null
    } catch {
      // Keep the plain-text body below.
    }
    const message = body?.error?.message || body?.message || text.trim().slice(0, 240) || fallbackMessage
    const error = new Error(message)
    ;(error as any).status = response.status
    ;(error as any).code = body?.error?.code || body?.error?.type || response.status
    ;(error as any).requestId = requestId
    return error
  } catch {
    const error = new Error(fallbackMessage)
    ;(error as any).status = response.status
    ;(error as any).code = response.status
    ;(error as any).requestId = requestId
    return error
  }
}

export async function listGatewayModels(apiKey: string): Promise<string[]> {
  const response = await fetch(buildGatewayUrl('/v1/models'), {
    headers: authHeaders(apiKey),
  })
  if (!response.ok) throw await parseGatewayError(response)

  const body = (await response.json()) as GatewayModelsResponse
  const seen = new Set<string>()
  return (body.data || [])
    .map(model => String(model?.id || '').trim())
    .filter((model) => {
      if (!model || seen.has(model)) return false
      seen.add(model)
      return true
    })
}

function extractTextDelta(event: any): string {
  if (!event || typeof event !== 'object') return ''

  const chatChoice = event.choices?.[0]
  const chatDelta = chatChoice?.delta?.content
  if (typeof chatDelta === 'string') return chatDelta

  const chatMessage = chatChoice?.message?.content
  if (typeof chatMessage === 'string') return chatMessage

  if (event.type === 'response.output_text.delta' && typeof event.delta === 'string') {
    return event.delta
  }
  if (event.type === 'content_block_delta' && typeof event.delta?.text === 'string') {
    return event.delta.text
  }

  return ''
}

function parseSSEDataBlocks(chunk: string, buffer: { value: string }): string[] {
  buffer.value += chunk
  const normalized = buffer.value.replace(/\r\n/g, '\n')
  const blocks = normalized.split('\n\n')
  buffer.value = blocks.pop() || ''
  return blocks
    .map(block => block
      .split('\n')
      .filter(line => line.startsWith('data:'))
      .map(line => line.slice(5).trimStart())
      .join('\n')
      .trim())
    .filter(Boolean)
}

async function readStreamingChatResponse(response: Response, model: string, startedAt: number): Promise<ChatTestResult> {
  if (!response.body) {
    throw new Error('当前浏览器不支持读取流式响应')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  const buffer = { value: '' }
  let content = ''
  let firstTokenMs: number | null = null

  try {
    while (true) {
      const { value, done } = await reader.read()
      if (done) break
      const blocks = parseSSEDataBlocks(decoder.decode(value, { stream: true }), buffer)
      for (const block of blocks) {
        if (block === '[DONE]') continue
        try {
          const event = JSON.parse(block)
          const delta = extractTextDelta(event)
          if (delta) {
            if (firstTokenMs === null) {
              firstTokenMs = Math.round(performance.now() - startedAt)
            }
            content += delta
          }
        } catch {
          // Ignore malformed keepalive/event fragments. The final empty result
          // still makes the test failure visible to the user without exposing raw internals.
        }
      }
    }
  } finally {
    reader.releaseLock()
  }

  const tail = parseSSEDataBlocks(decoder.decode(), buffer)
  for (const block of tail) {
    if (block === '[DONE]') continue
    try {
      const delta = extractTextDelta(JSON.parse(block))
      if (delta) {
        if (firstTokenMs === null) firstTokenMs = Math.round(performance.now() - startedAt)
        content += delta
      }
    } catch {
      // ignore
    }
  }

  return {
    content,
    firstTokenMs,
    durationMs: Math.round(performance.now() - startedAt),
    model,
    requestId: response.headers.get('X-Request-Id') || '',
  }
}

export async function testChatCompletion(request: ChatTestRequest): Promise<ChatTestResult> {
  const startedAt = performance.now()
  const response = await fetch(buildGatewayUrl('/v1/chat/completions'), {
    method: 'POST',
    headers: authHeaders(request.apiKey, {
      'Content-Type': 'application/json',
    }),
    body: JSON.stringify({
      model: request.model,
      messages: request.messages,
      temperature: request.temperature ?? 0.7,
      max_tokens: request.maxTokens ?? 512,
      stream: true,
    }),
  })
  if (!response.ok) throw await parseGatewayError(response)

  const contentType = response.headers.get('Content-Type') || ''
  if (contentType.includes('text/event-stream')) {
    return readStreamingChatResponse(response, request.model, startedAt)
  }

  const body = await response.json()
  return {
    content: extractTextDelta(body),
    firstTokenMs: null,
    durationMs: Math.round(performance.now() - startedAt),
    model: request.model,
    requestId: response.headers.get('X-Request-Id') || '',
  }
}

function imageMimeTypeFromFormat(format?: string, fallback = 'image/png'): string {
  const normalized = String(format || '').toLowerCase()
  if (!normalized) return fallback
  if (normalized === 'jpeg' || normalized === 'jpg') return 'image/jpeg'
  if (normalized === 'webp') return 'image/webp'
  return 'image/png'
}

function normalizeImageItem(item: any, fallbackMimeType: string): GeneratedImage | null {
  const b64 = item?.b64_json || item?.base64 || item?.image_base64 || item?.result
  if (typeof b64 === 'string' && b64.trim()) {
    const mimeType = imageMimeTypeFromFormat(item?.output_format, fallbackMimeType)
    return {
      url: b64.startsWith('data:') ? b64 : `data:${mimeType};base64,${b64}`,
      mimeType,
      revisedPrompt: item?.revised_prompt,
    }
  }

  if (typeof item?.url === 'string' && item.url.trim()) {
    return {
      url: item.url,
      mimeType: fallbackMimeType,
      revisedPrompt: item?.revised_prompt,
    }
  }

  return null
}

export async function generateImage(request: ImageGenerationRequest): Promise<ImageGenerationResult> {
  const startedAt = performance.now()
  const outputFormat = request.outputFormat || 'png'
  const response = await fetch(buildGatewayUrl('/v1/images/generations'), {
    method: 'POST',
    headers: authHeaders(request.apiKey, {
      'Content-Type': 'application/json',
    }),
    body: JSON.stringify({
      model: request.model || 'gpt-image-2',
      prompt: request.prompt,
      size: request.size || '1024x1024',
      quality: request.quality || 'auto',
      output_format: outputFormat,
      response_format: 'b64_json',
      n: request.n || 1,
    }),
  })
  if (!response.ok) throw await parseGatewayError(response)

  const body = await response.json()
  const fallbackMimeType = imageMimeTypeFromFormat(outputFormat)
  const candidates = [
    ...(Array.isArray(body?.data) ? body.data : []),
    ...(Array.isArray(body?.output) ? body.output : []),
  ]
  const images = candidates
    .map(item => normalizeImageItem(item, fallbackMimeType))
    .filter((item): item is GeneratedImage => Boolean(item))

  return {
    images,
    durationMs: Math.round(performance.now() - startedAt),
    requestId: response.headers.get('X-Request-Id') || '',
    raw: body,
  }
}

export const gatewayToolsAPI = {
  listGatewayModels,
  testChatCompletion,
  generateImage,
}
