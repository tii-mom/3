import { afterEach, describe, expect, it, vi } from 'vitest'

function jsonResponse(body: unknown, init?: ResponseInit): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    status: 200,
    headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) },
  })
}

function sseResponse(body: string, init?: ResponseInit): Response {
  return new Response(
    new ReadableStream({
      start(controller) {
        controller.enqueue(new TextEncoder().encode(body))
        controller.close()
      },
    }),
    {
      ...init,
      status: 200,
      headers: { 'Content-Type': 'text/event-stream', ...(init?.headers || {}) },
    },
  )
}

describe('gatewayToolsAPI', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('lists unique model ids from /v1/models', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(jsonResponse({
      data: [
        { id: 'gpt-5.5' },
        { id: 'gpt-image-2' },
        { id: 'gpt-image-2' },
      ],
    }))
    const { listGatewayModels } = await import('../gatewayTools')

    await expect(listGatewayModels('sk-test')).resolves.toEqual(['gpt-5.5', 'gpt-image-2'])
    expect(String(fetchMock.mock.calls[0][0])).toMatch(/\/v1\/models$/)
    expect(fetchMock.mock.calls[0][1]).toEqual({
      headers: { Authorization: 'Bearer sk-test' },
    })
  })

  it('measures streaming chat first-token latency and concatenates deltas', async () => {
    vi.spyOn(performance, 'now')
      .mockReturnValueOnce(100)
      .mockReturnValueOnce(180)
      .mockReturnValueOnce(260)
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(sseResponse(
      'data: {"choices":[{"delta":{"content":"你"}}]}\n\n' +
      'data: {"choices":[{"delta":{"content":"好"}}]}\n\n' +
      'data: [DONE]\n\n',
      { headers: { 'X-Request-Id': 'req_1' } },
    ))
    const { testChatCompletion } = await import('../gatewayTools')

    const result = await testChatCompletion({
      apiKey: 'sk-test',
      model: 'gpt-5.5',
      messages: [{ role: 'user', content: 'hello' }],
    })

    expect(result.content).toBe('你好')
    expect(result.firstTokenMs).toBe(80)
    expect(result.durationMs).toBe(160)
    expect(result.requestId).toBe('req_1')
  })

  it('normalizes generated image base64 responses', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(jsonResponse({
      data: [
        { b64_json: 'aW1hZ2U=', revised_prompt: 'draw a cat' },
      ],
    }, { headers: { 'X-Request-Id': 'img_1' } }))
    const { generateImage } = await import('../gatewayTools')

    const result = await generateImage({
      apiKey: 'sk-test',
      prompt: 'draw',
      outputFormat: 'png',
    })

    expect(result.images).toEqual([
      {
        url: 'data:image/png;base64,aW1hZ2U=',
        mimeType: 'image/png',
        revisedPrompt: 'draw a cat',
      },
    ])
    expect(result.requestId).toBe('img_1')
  })
})
