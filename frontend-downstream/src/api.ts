export interface SiteConfig {
  id: number
  slug: string
  domain: string
  name: string
  logo_url: string
  theme_color: string
  api_base_url: string
}

interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

function apiOrigin(): string {
  const configured = import.meta.env.VITE_DOWNSTREAM_ORIGIN as string | undefined
  if (configured) {
    return configured.replace(/\/$/, '')
  }
  return ''
}

// 子站入口页只读取自身的品牌配置；用户数据全部由主站前端负责。
export async function getSite(): Promise<SiteConfig> {
  const response = await fetch(`${apiOrigin()}/api/internal/downstream/v1/site`, {
    headers: { Accept: 'application/json' },
    credentials: 'include'
  })
  const payload = (await response.json().catch(() => null)) as ApiEnvelope<SiteConfig> | null
  if (!response.ok) {
    throw new Error(payload?.message || `站点信息加载失败（${response.status}）`)
  }
  if (payload && typeof payload === 'object' && 'code' in payload) {
    if (payload.code !== 0) {
      throw new Error(payload.message || '站点信息加载失败')
    }
    return payload.data
  }
  return payload as unknown as SiteConfig
}
