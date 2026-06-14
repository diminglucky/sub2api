import { describe, expect, it } from 'vitest'
import { DEFAULT_PUBLIC_API_BASE_URL, resolvePublicApiEndpoint } from '../apiEndpoint'

describe('resolvePublicApiEndpoint', () => {
  it('defaults to the dedicated API domain', () => {
    expect(resolvePublicApiEndpoint('')).toBe(DEFAULT_PUBLIC_API_BASE_URL)
    expect(resolvePublicApiEndpoint(null)).toBe(DEFAULT_PUBLIC_API_BASE_URL)
  })

  it('normalizes the legacy web domain to the API domain', () => {
    expect(resolvePublicApiEndpoint('https://superai.dihappy.cfd/v1/')).toBe(DEFAULT_PUBLIC_API_BASE_URL)
  })

  it('adds the v1 path for the dedicated API root domain', () => {
    expect(resolvePublicApiEndpoint('https://api.dihappy.cfd')).toBe(DEFAULT_PUBLIC_API_BASE_URL)
    expect(resolvePublicApiEndpoint('https://api.dihappy.cfd/')).toBe(DEFAULT_PUBLIC_API_BASE_URL)
  })

  it('keeps custom paths on the web domain', () => {
    expect(resolvePublicApiEndpoint('https://superai.dihappy.cfd/custom-api/v1/')).toBe('https://superai.dihappy.cfd/custom-api/v1')
  })

  it('keeps custom API endpoints', () => {
    expect(resolvePublicApiEndpoint('https://custom.example.com/v1/')).toBe('https://custom.example.com/v1')
  })
})
