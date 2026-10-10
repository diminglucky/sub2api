export interface SiteConfig {
  id: number
  slug: string
  domain: string
  name: string
  logo_url: string
  theme_color: string
  api_base_url: string
}

export interface UserSummary {
  user_id: number
  subsite_id: number
  email: string
  balance: number
  role: string
  source: string
  is_admin: boolean
  usage_count: number
  usage_cost: number
  recharge_count: number
  recharge_amount: number
}

export interface AdminSummary {
  subsite_id: number
  member_count: number
  recharge_count: number
  recharge_amount: number
  usage_count: number
  usage_cost: number
  settlement_pending: number
}

export interface AuthUser {
  id: number
  email: string
  username?: string
}

export interface AuthResponse {
  access_token: string
  refresh_token?: string
  expires_in?: number
  token_type: string
  user: AuthUser
}

export interface ApiKeyInfo {
  id: number
  key: string
  name: string
  status: string
  group?: {
    allow_image_generation?: boolean
  } | null
}

export interface ApiKeyPage {
  items: ApiKeyInfo[]
  total: number
  page: number
  page_size: number
  pages: number
}

interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

const TOKEN_KEY = 'draw_downstream_access_token'
const USER_KEY = 'draw_downstream_user'

function apiOrigin(): string {
  const configured = import.meta.env.VITE_DOWNSTREAM_ORIGIN as string | undefined
  if (configured) {
    return configured.replace(/\/$/, '')
  }
  return ''
}

function authHeaders(): HeadersInit {
  const token = getAccessToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Content-Type', 'application/json')
  for (const [key, value] of Object.entries(authHeaders())) {
    headers.set(key, value)
  }

  const response = await fetch(`${apiOrigin()}${path}`, {
    ...init,
    headers,
    credentials: 'include'
  })

  let payload: ApiEnvelope<T> | null = null
  try {
    payload = (await response.json()) as ApiEnvelope<T>
  } catch {
    payload = null
  }

  if (!response.ok) {
    const message = payload?.message || `Request failed with status ${response.status}`
    throw new Error(message)
  }
  if (payload && typeof payload === 'object' && 'code' in payload) {
    if (payload.code !== 0) {
      throw new Error(payload.message || 'Request failed')
    }
    return payload.data
  }
  return payload as T
}

export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function hasSession(): boolean {
  return !!getAccessToken()
}

export function clearSession(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

function saveSession(response: AuthResponse): void {
  localStorage.setItem(TOKEN_KEY, response.access_token)
  localStorage.setItem(USER_KEY, JSON.stringify(response.user))
}

export async function getSite(): Promise<SiteConfig> {
  return request<SiteConfig>('/api/internal/downstream/v1/site')
}

export async function getMeSummary(): Promise<UserSummary> {
  return request<UserSummary>('/api/internal/downstream/v1/me/summary')
}

export async function getAdminSummary(): Promise<AdminSummary> {
  return request<AdminSummary>('/api/internal/downstream/v1/admin/summary')
}

// List the signed-in user's own API keys. The workbench needs a key to call
// the image endpoints, which always authenticate with an API key.
export async function listApiKeys(pageSize = 100): Promise<ApiKeyPage> {
  return request<ApiKeyPage>(`/api/v1/keys?page=1&page_size=${pageSize}`)
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const response = await request<AuthResponse>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password })
  })
  saveSession(response)
  return response
}

export async function register(email: string, password: string): Promise<AuthResponse> {
  const response = await request<AuthResponse>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify({ email, password })
  })
  saveSession(response)
  return response
}

export async function logout(): Promise<void> {
  try {
    await request('/api/v1/auth/logout', { method: 'POST' })
  } finally {
    clearSession()
  }
}
