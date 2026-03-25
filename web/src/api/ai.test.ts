import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { streamChat } from './ai'

// Helper to build a ReadableStream from an array of SSE event strings.
function makeSSEStream(events: string[]): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder()
  return new ReadableStream({
    start(controller) {
      for (const event of events) {
        controller.enqueue(encoder.encode(event))
      }
      controller.close()
    },
  })
}

describe('streamChat', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    // Reset localStorage token
    localStorage.removeItem('access_token')
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
    vi.restoreAllMocks()
  })

  it('calls fetch with correct URL, method, and Content-Type', async () => {
    const stream = makeSSEStream(['data: [DONE]\n\n'])
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      body: stream,
    })

    await streamChat('Hello', [], vi.fn(), vi.fn(), vi.fn())

    expect(globalThis.fetch).toHaveBeenCalledWith(
      '/api/v1/ai/chat',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json',
        }),
        body: expect.stringContaining('"message":"Hello"'),
      })
    )
  })

  it('calls onDelta with delta text from SSE stream', async () => {
    const stream = makeSSEStream([
      'data: {"delta":"Hello"}\n\n',
      'data: {"delta":" World"}\n\n',
      'data: [DONE]\n\n',
    ])
    globalThis.fetch = vi.fn().mockResolvedValue({ ok: true, body: stream })

    const deltas: string[] = []
    const onDelta = vi.fn((d: string) => deltas.push(d))
    const onDone = vi.fn()

    await streamChat('Hi', [], onDelta, onDone, vi.fn())

    expect(onDelta).toHaveBeenCalledWith('Hello')
    expect(onDelta).toHaveBeenCalledWith(' World')
    expect(onDone).toHaveBeenCalledTimes(1)
  })

  it('calls onDone when [DONE] is received', async () => {
    const stream = makeSSEStream(['data: [DONE]\n\n'])
    globalThis.fetch = vi.fn().mockResolvedValue({ ok: true, body: stream })

    const onDone = vi.fn()
    await streamChat('test', [], vi.fn(), onDone, vi.fn())

    expect(onDone).toHaveBeenCalledTimes(1)
  })

  it('calls onError when fetch returns non-ok status', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 403,
      body: null,
    })

    const onError = vi.fn()
    await streamChat('test', [], vi.fn(), vi.fn(), onError)

    expect(onError).toHaveBeenCalledWith(expect.stringContaining('403'))
  })

  it('calls onError when SSE stream contains error payload', async () => {
    const stream = makeSSEStream(['data: {"error":"model unavailable"}\n\n'])
    globalThis.fetch = vi.fn().mockResolvedValue({ ok: true, body: stream })

    const onError = vi.fn()
    await streamChat('test', [], vi.fn(), vi.fn(), onError)

    expect(onError).toHaveBeenCalledWith('model unavailable')
  })

  it('attaches Authorization header when token is in localStorage', async () => {
    localStorage.setItem('access_token', 'test-jwt-token')
    const stream = makeSSEStream(['data: [DONE]\n\n'])
    globalThis.fetch = vi.fn().mockResolvedValue({ ok: true, body: stream })

    await streamChat('Hi', [], vi.fn(), vi.fn(), vi.fn())

    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer test-jwt-token',
        }),
      })
    )
  })

  it('sends history in request body', async () => {
    const stream = makeSSEStream(['data: [DONE]\n\n'])
    globalThis.fetch = vi.fn().mockResolvedValue({ ok: true, body: stream })

    const history = [{ role: 'user' as const, content: 'previous message' }]
    await streamChat('follow up', history, vi.fn(), vi.fn(), vi.fn())

    const callArgs = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][1]
    const body = JSON.parse(callArgs.body as string)
    expect(body.history).toEqual(history)
  })
})
