import { describe, expect, it } from 'vitest'
import {
  buildDocsExample,
  docsEndpointOrder,
  resolveDocsApiBaseUrl,
  resolveDocsRequestUrl,
  resolveDocsRootUrl
} from '../apiDocs'

describe('public API documentation examples', () => {
  it('uses the current site without configuration and preserves a configured gateway path', () => {
    expect(resolveDocsApiBaseUrl('', 'http://localhost:3000')).toBe('https://api.superai.sbs/v1')
    expect(resolveDocsApiBaseUrl('http://127.0.0.1:3000/v1', 'http://localhost:3000')).toBe('https://api.superai.sbs/v1')
    expect(resolveDocsApiBaseUrl('https://api.example.com/gateway/v1/', 'http://localhost:3000')).toBe('https://api.example.com/gateway/v1')
    expect(resolveDocsApiBaseUrl('https://api.example.com', 'http://localhost:3000')).toBe('https://api.example.com/v1')
    expect(resolveDocsApiBaseUrl('/gateway/v1', 'https://example.com')).toBe('https://example.com/gateway/v1')
    expect(resolveDocsApiBaseUrl('javascript:alert(1)', 'https://example.com')).toBe('https://api.superai.sbs/v1')
    expect(resolveDocsApiBaseUrl('', 'https://superai.sbs')).toBe('https://api.superai.sbs/v1')
    expect(resolveDocsApiBaseUrl('', 'https://www.superai.sbs')).toBe('https://api.superai.sbs/v1')
    expect(resolveDocsApiBaseUrl('https://superai.sbs/v1', 'https://superai.sbs')).toBe('https://api.superai.sbs/v1')
    expect(resolveDocsApiBaseUrl('https://api.superai.sbs/v1', 'https://superai.sbs')).toBe('https://api.superai.sbs/v1')
  })

  it('exposes the full SuperAI-compatible endpoint set', () => {
    expect(docsEndpointOrder).toEqual([
      'models',
      'chat',
      'responses',
      'messages',
      'embeddings',
      'imagesGenerations',
      'imagesEdits',
      'gemini'
    ])
  })

  it('uses protocol-specific authentication headers and URLs', () => {
    const messages = buildDocsExample('https://api.superai.sbs/v1', 'messages', 'curl', false)
    expect(messages).toContain('x-api-key: $API_KEY')
    expect(messages).toContain('anthropic-version: 2023-06-01')
    expect(messages).toContain('"max_tokens": 1024')

    const gemini = buildDocsExample('https://api.superai.sbs/v1', 'gemini', 'curl', false)
    expect(gemini).toContain('x-goog-api-key: $API_KEY')
    expect(gemini).toContain('https://api.superai.sbs/v1beta/models/gemini-2.5-flash:generateContent')
    expect(gemini).not.toContain('/v1/v1beta')
    expect(gemini).toContain('"contents"')
  })

  it('queries models with GET and no request body', () => {
    const js = buildDocsExample('https://api.superai.sbs/v1', 'models', 'javascript', false)
    expect(js).toContain('method: "GET"')
    expect(js).not.toContain('body:')
    expect(js).not.toContain('gpt-5.6-sol')
  })

  it('generates image generation and edit examples', () => {
    const generation = buildDocsExample('https://api.superai.sbs/v1', 'imagesGenerations', 'curl', false)
    expect(generation).toContain('"model": "gpt-image-2"')
    expect(generation).toContain('"response_format": "b64_json"')

    const editCurl = buildDocsExample('https://api.superai.sbs/v1', 'imagesEdits', 'curl', false)
    expect(editCurl).toContain('-F "image=@input.png"')
    expect(editCurl).toContain('-F "prompt=')

    const editPython = buildDocsExample('https://api.superai.sbs/v1', 'imagesEdits', 'python', false)
    expect(editPython).toContain('requests.post')
    expect(editPython).toContain('files = {"image": open("input.png", "rb")}')

    const editJavaScript = buildDocsExample('https://api.superai.sbs/v1', 'imagesEdits', 'javascript', false)
    expect(editJavaScript).toContain('new FormData()')
    expect(editJavaScript).toContain('form.append("image"')
  })

  it('generates streaming examples for every stream-capable protocol', () => {
    for (const endpoint of ['chat', 'responses', 'messages', 'gemini'] as const) {
      expect(buildDocsExample('https://api.superai.sbs/v1', endpoint, 'curl', true)).toContain('curl -N')
      expect(buildDocsExample('https://api.superai.sbs/v1', endpoint, 'python', true)).toContain('for line in response:')
      expect(buildDocsExample('https://api.superai.sbs/v1', endpoint, 'javascript', true)).toContain('for await (const chunk of response.body)')
    }
  })

  it('resolves request URLs without duplicating the /v1 segment', () => {
    expect(resolveDocsRootUrl('https://api.superai.sbs/v1')).toBe('https://api.superai.sbs')
    expect(resolveDocsRequestUrl('https://api.superai.sbs/v1', 'chat')).toBe('https://api.superai.sbs/v1/chat/completions')
    expect(resolveDocsRequestUrl('https://api.superai.sbs/v1', 'gemini')).toBe('https://api.superai.sbs/v1beta/models/gemini-2.5-flash:generateContent')
  })
})
