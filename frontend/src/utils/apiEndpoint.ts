export const DEFAULT_PUBLIC_API_BASE_URL = 'https://api.dihappy.cfd/v1'

export function resolvePublicApiEndpoint(configured?: string | null): string {
  const trimmed = (configured || '').trim().replace(/\/+$/, '')
  if (!trimmed) return DEFAULT_PUBLIC_API_BASE_URL

  try {
    const parsed = new URL(trimmed)
    const host = parsed.hostname.toLowerCase()
    const path = parsed.pathname.replace(/\/+$/, '')
    if (host === 'superai.dihappy.cfd' && (path === '' || path === '/v1')) {
      return DEFAULT_PUBLIC_API_BASE_URL
    }
    if (host === 'api.dihappy.cfd' && (parsed.pathname === '' || parsed.pathname === '/')) {
      return DEFAULT_PUBLIC_API_BASE_URL
    }
  } catch {
    return trimmed
  }

  return trimmed
}
