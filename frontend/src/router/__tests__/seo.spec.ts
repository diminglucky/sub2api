import { describe, expect, it, vi } from 'vitest'

import { resolveRouteSeo } from '../seo'

vi.mock('@/i18n', () => ({
  i18n: {
    global: {
      t: (key: string) => key === 'home.seoDescription'
        ? 'SuperAI 兼容 OpenAI、GPT、Claude、DeepSeek 的 API 中转站'
        : key,
    },
  },
}))

describe('resolveRouteSeo', () => {
  it('adds indexed search metadata to public routes', () => {
    const seo = resolveRouteSeo({
      path: '/home',
      meta: {
        requiresAuth: false,
        descriptionKey: 'home.seoDescription',
        seoKeywords: ['SuperAI', 'GPT', 'OpenAI', 'Claude', 'DeepSeek'],
      },
    }, 'SuperAI - AI API Gateway', 'https://superai.dihappy.cfd')

    expect(seo.description).toContain('SuperAI')
    expect(seo.description).toContain('OpenAI')
    expect(seo.keywords).toContain('GPT')
    expect(seo.canonical).toBe('https://superai.dihappy.cfd/home')
    expect(seo.robots).toBe('index,follow')
  })

  it('keeps authenticated routes out of search indexes', () => {
    const seo = resolveRouteSeo({
      path: '/admin/settings',
      meta: {
        requiresAuth: true,
        requiresAdmin: true,
      },
    }, 'System Settings - SuperAI', 'https://superai.dihappy.cfd')

    expect(seo.robots).toBe('noindex,nofollow')
  })
})
