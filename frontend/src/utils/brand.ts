export const DEFAULT_BRAND_NAME = 'SuperAI'

/**
 * Normalizes the configurable site name. Legacy deployments stored "Sub2API"
 * as the site name, but this project is branded SuperAI.
 */
export function normalizeBrandName(name?: string | null): string {
  const trimmed = typeof name === 'string' ? name.trim() : ''
  if (!trimmed) return DEFAULT_BRAND_NAME
  if (['sub2api', 'sub2 api', 'sub2-api'].includes(trimmed.toLowerCase())) return DEFAULT_BRAND_NAME
  return trimmed
}
