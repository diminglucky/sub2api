import { describe, expect, it } from 'vitest'
import { buildDocsExample, resolveDocsApiBaseUrl } from '../apiDocs'

describe('public API documentation examples', () => {
  it('uses the current site without configuration and preserves a configured gateway path', () => {
    expect(resolveDocsApiBaseUrl('', 'http://localhost:3000')).toBe('http://localhost:3000/v1')
    expect(resolveDocsApiBaseUrl('https://api.example.com/gateway/v1/', 'http://localhost:3000')).toBe('https://api.example.com/gateway/v1')
    expect(resolveDocsApiBaseUrl('https://api.example.com', 'http://localhost:3000')).toBe('https://api.example.com/v1')
    expect(resolveDocsApiBaseUrl('/gateway/v1', 'https://example.com')).toBe('https://example.com/gateway/v1')
    expect(resolveDocsApiBaseUrl('javascript:alert(1)', 'https://example.com')).toBe('https://example.com/v1')
  })

  it('uses protocol-specific authentication and required Messages parameters', () => {
    const curl = buildDocsExample('https://example.com/v1', 'messages', 'curl', false)
    expect(curl).toContain('x-api-key: $API_KEY')
    expect(curl).toContain('anthropic-version: 2023-06-01')
    expect(curl).toContain('"max_tokens": 1024')
    const python = buildDocsExample('https://example.com/v1', 'chat', 'python', false)
    expect(python).toContain('"Authorization": "Bearer " + os.environ["API_KEY"]')
    expect(python).not.toContain('os.environ[os.environ')
  })

  it('queries models with GET and no request body', () => {
    const js = buildDocsExample('https://example.com/v1', 'models', 'javascript', false)
    expect(js).toContain('method: "GET"')
    expect(js).not.toContain('body:')
    expect(js).not.toContain('YOUR_MODEL')
  })

  it('generates streaming examples for all supported languages', () => {
    expect(buildDocsExample('https://example.com/v1', 'responses', 'curl', true)).toContain('curl -N')
    expect(buildDocsExample('https://example.com/v1', 'responses', 'python', true)).toContain('for line in response:')
    expect(buildDocsExample('https://example.com/v1', 'chat', 'javascript', true)).toContain('for await (const chunk of response.body)')
  })
})
