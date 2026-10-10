import type { SiteConfig } from './api'

// 绘图工作台直接调用站点同源的 OpenAI 兼容网关（/v1/...），不经过内部 API。
// 这样可以复用主站现有的图片模型、计费和存储能力，也保持子站前端低耦合。

export type DrawFormat = 'png' | 'jpeg' | 'webp'

export interface DrawRequest {
  model: string
  prompt: string
  size: string
  quality: string
  format: DrawFormat
  count: number
}

export interface DrawResult {
  src: string
  format: DrawFormat
  model: string
  size: string
  prompt: string
  createdAt: number
}

export const DRAW_SIZES = [
  { value: '1024x1024', label: '1:1 方形' },
  { value: '1536x1024', label: '3:2 横向' },
  { value: '1024x1536', label: '2:3 竖向' },
  { value: '1792x1024', label: '16:9 宽屏' },
  { value: '1024x1792', label: '9:16 竖屏' }
]

export const DRAW_QUALITIES = [
  { value: 'auto', label: '自动' },
  { value: 'low', label: '快速' },
  { value: 'medium', label: '标准' },
  { value: 'high', label: '精细' }
]

export const DRAW_FORMATS = [
  { value: 'png', label: 'PNG' },
  { value: 'jpeg', label: 'JPEG' },
  { value: 'webp', label: 'WEBP' }
]

export function apiBasePath(site?: SiteConfig): string {
  // 站内调用用相对路径，避免把生产域名写死进开发环境。
  if (site?.api_base_url) {
    try {
      const url = new URL(site.api_base_url)
      if (url.host === window.location.host) {
        return url.pathname.replace(/\/$/, '')
      }
    } catch {
      // 配置异常时回退到默认路径。
    }
  }
  return '/v1'
}

export function isImageModel(name: string): boolean {
  const value = name.trim().toLowerCase()
  return (
    value.startsWith('gpt-image-') ||
    value.includes('dall-e') ||
    value.includes('image') ||
    value.includes('imagen') ||
    value.includes('flux') ||
    value.includes('sdxl') ||
    value.includes('nano-banana')
  )
}

export function errorMessageFrom(error: unknown, fallback = '生成失败，请稍后重试'): string {
  const message =
    error instanceof Error
      ? error.message
      : typeof error === 'object' && error !== null && 'message' in error
        ? String((error as { message?: unknown }).message || '')
        : ''
  const normalized = message.trim().toLowerCase()
  if (!normalized) return fallback
  if (normalized === 'failed to fetch' || normalized.includes('networkerror')) {
    return '无法连接到接口服务，请检查网络后重试。'
  }
  if (normalized.includes('insufficient') || normalized.includes('余额不足')) {
    return '账户余额不足，请先充值后再生成。'
  }
  if (normalized.includes('concurrency') || normalized.includes('并发')) {
    return '当前并发已满，请稍后重试。'
  }
  if (normalized.includes('not assigned to any group')) {
    return '当前 API Key 还没有分配分组，请联系管理员分配后再生成。'
  }
  if (normalized.includes('permission_error') || normalized.includes('forbidden')) {
    return '当前 API Key 没有生成图片的权限，请更换密钥。'
  }
  if (normalized.includes('model') && (normalized.includes('not found') || normalized.includes('unsupported'))) {
    return '当前模型不可用，请换一个模型。'
  }
  return message || fallback
}

export async function listImageModels(apiKey: string, site?: SiteConfig): Promise<string[]> {
  const response = await fetch(`${apiBasePath(site)}/models`, {
    headers: { Authorization: `Bearer ${apiKey}` }
  })
  const payload = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(payload?.error?.message || payload?.message || `加载模型失败（${response.status}）`)
  }
  const names: string[] = Array.isArray(payload?.data)
    ? payload.data
        .map((item: unknown) => {
          if (typeof item === 'string') return item
          const record = item as { id?: unknown; model?: unknown; name?: unknown }
          return record?.id || record?.model || record?.name || ''
        })
        .map((name: unknown) => String(name).trim())
        .filter(Boolean)
    : []
  const unique = Array.from(new Set(names)).filter(isImageModel)
  return unique.sort((a, b) => a.localeCompare(b, undefined, { numeric: true, sensitivity: 'base' }))
}

export async function generateImage(
  apiKey: string,
  request: DrawRequest,
  site?: SiteConfig
): Promise<DrawResult[]> {
  const response = await fetch(`${apiBasePath(site)}/images/generations`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${apiKey}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      model: request.model,
      prompt: request.prompt,
      size: request.size,
      quality: request.quality,
      output_format: request.format,
      n: Math.min(Math.max(Number(request.count) || 1, 1), 4),
      response_format: 'b64_json'
    })
  })
  const payload = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(payload?.error?.message || payload?.message || `生成失败（${response.status}）`)
  }
  return extractResults(payload, request)
}

export async function editImage(
  apiKey: string,
  request: DrawRequest,
  reference: File,
  site?: SiteConfig
): Promise<DrawResult[]> {
  const form = new FormData()
  form.append('model', request.model)
  form.append('prompt', request.prompt)
  form.append('size', request.size)
  form.append('quality', request.quality)
  form.append('output_format', request.format)
  form.append('n', String(Math.min(Math.max(Number(request.count) || 1, 1), 4)))
  form.append('response_format', 'b64_json')
  form.append('image[]', reference, reference.name || 'reference.png')

  const response = await fetch(`${apiBasePath(site)}/images/edits`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${apiKey}` },
    body: form
  })
  const payload = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(payload?.error?.message || payload?.message || `生成失败（${response.status}）`)
  }
  return extractResults(payload, request)
}

function extractResults(payload: any, request: DrawRequest): DrawResult[] {
  const items: any[] = []
  const collect = (value: any) => {
    if (!value) return
    if (Array.isArray(value)) {
      value.forEach(collect)
      return
    }
    if (Array.isArray(value.data)) return collect(value.data)
    if (Array.isArray(value.candidates)) {
      value.candidates.forEach((candidate: any) => {
        ;(candidate?.content?.parts || []).forEach((part: any) => {
          if (part?.inlineData?.data) {
            items.push({ b64_json: part.inlineData.data, mime_type: part.inlineData.mimeType })
          }
        })
      })
      return
    }
    if (Array.isArray(value.output)) return collect(value.output)
    if (value.b64_json || value.url || value.result) items.push(value)
  }
  collect(payload)

  const now = Date.now()
  return items
    .map((item: any): DrawResult | null => {
      const mime = String(item.mime_type || item.mimeType || `image/${request.format}`)
      const b64 = String(item.b64_json || item.data || '')
      const src = b64 ? `data:${mime};base64,${b64}` : String(item.url || item.result || '')
      if (!src) return null
      return {
        src,
        format: request.format,
        model: request.model,
        size: request.size,
        prompt: request.prompt,
        createdAt: now
      }
    })
    .filter((item): item is DrawResult => item !== null)
}

const GALLERY_KEY = 'draw_workbench_gallery'
const GALLERY_LIMIT = 8

export function loadGallery(): DrawResult[] {
  try {
    const raw = localStorage.getItem(GALLERY_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed.filter(
      (item): item is DrawResult =>
        !!item && typeof item === 'object' && typeof (item as DrawResult).src === 'string'
    )
  } catch {
    return []
  }
}

// 图片是 base64 内联数据，体积较大。这里保留最近的结果，并在超限时逐步丢弃最旧记录。
export function saveGallery(results: DrawResult[]): void {
  const trimmed = results.slice(0, GALLERY_LIMIT)
  for (let size = trimmed.length; size >= 0; size -= 1) {
    try {
      localStorage.setItem(GALLERY_KEY, JSON.stringify(trimmed.slice(0, size)))
      return
    } catch {
      // 存储超限，继续缩减后重试。
    }
  }
  try {
    localStorage.removeItem(GALLERY_KEY)
  } catch {
    // 忽略：存储不可用时不再持久化。
  }
}
