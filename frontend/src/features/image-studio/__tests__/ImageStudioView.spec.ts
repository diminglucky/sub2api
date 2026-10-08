import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ImageStudioView from '../ImageStudioView.vue'

const { listKeys, getPublicSettings, listGallery, saveGallery, copyToClipboard } = vi.hoisted(() => ({
  listKeys: vi.fn(),
  getPublicSettings: vi.fn(),
  listGallery: vi.fn(),
  saveGallery: vi.fn(),
  copyToClipboard: vi.fn(),
}))

vi.mock('@/api', () => ({
  keysAPI: {
    list: listKeys,
  },
  authAPI: {
    getPublicSettings,
  },
  imageStudioGalleryAPI: {
    list: listGallery,
    save: saveGallery,
  },
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../ImageStudioView.vue')
const componentSource = readFileSync(componentPath, 'utf8')

function mountImageStudio() {
  return mount(ImageStudioView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: {
          props: ['show', 'title', 'closeOnClickOutside'],
          emits: ['close'],
          template: '<div v-if="show"><button v-if="closeOnClickOutside" data-testid="dialog-backdrop" @click="$emit(\'close\')">backdrop</button><h3>{{ title }}</h3><slot /><slot name="footer" /></div>',
        },
        Icon: { template: '<span />' },
      },
    },
  })
}

function resetCommonMocks() {
  listKeys.mockReset()
  getPublicSettings.mockReset()
  listGallery.mockReset()
  saveGallery.mockReset()
  copyToClipboard.mockReset()
  listGallery.mockResolvedValue([])
  copyToClipboard.mockResolvedValue(true)
  saveGallery.mockResolvedValue({
    id: 'gallery-1',
    url: 'https://cdn.example.com/gallery-1.png',
    prompt: '',
    model: '',
    size: '',
    format: '',
    created_at: 0,
    expires_at: 0,
  })
}

describe('ImageStudioView API key selector', () => {
  it('shows only the API key name', () => {
    expect(componentSource).toContain('label: key.name')
    expect(componentSource).not.toContain('maskApiKey')
  })
})

describe('ImageStudioView image model loading', () => {
  beforeEach(() => {
    resetCommonMocks()
    vi.stubGlobal('fetch', vi.fn())
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { hostname: 'localhost' },
    })
  })

  it('automatically loads and deduplicates image models from the batch endpoint', async () => {
    listKeys.mockResolvedValue({
      items: [
        {
          id: 1,
          name: 'dd',
          key: 'sk-local',
          status: 'active',
          group_id: 2,
        },
      ],
    })
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://api.dihappy.cfd/v1' })
    const fetchMock = vi.mocked(fetch)
    fetchMock.mockImplementation(async (input) => {
      const url = String(input)
      if (url.endsWith('/images/batches/models')) {
        return {
          ok: true,
          json: async () => ({
            data: [
              { id: 'gpt-image-1' },
              { id: ' gpt-image-2-5-flare ' },
              { id: 'GPT-IMAGE-2-5-FLARE' },
              { id: 'gpt-image-2-5-flare' },
            ],
          }),
        } as Response
      }
      if (url.endsWith('/models')) {
        throw new Error('generic models endpoint should not be requested when batch models exist')
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    const wrapper = mountImageStudio()
    await flushPromises()
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledWith('/v1/images/batches/models', {
      headers: { Authorization: 'Bearer sk-local' },
    })
    expect(fetchMock).not.toHaveBeenCalledWith('/v1/models', expect.anything())

    const apiKeySelect = wrapper.get('[data-testid="api-key-select"]')
    expect(apiKeySelect.text()).toContain('dd')
    expect(apiKeySelect.classes()).toContain('studio-select--down')

    const modelSelect = wrapper.get('[data-testid="image-model-select"]')
    expect(modelSelect.text()).toContain('gpt-image-1')
    expect(modelSelect.classes()).toContain('studio-select--down')
    expect(modelSelect.text()).not.toContain('nano-banana-2')
    expect(modelSelect.text()).not.toContain('gpt-5')

    await apiKeySelect.get('button').trigger('click')
    expect(apiKeySelect.text()).not.toContain('选择密钥')
    expect(apiKeySelect.text()).toContain('dd')

    await modelSelect.get('button').trigger('click')
    expect(modelSelect.text()).toContain('gpt-image-1')
    expect(modelSelect.text()).toContain('gpt-image-2-5-flare')
    expect(modelSelect.text().match(/gpt-image-2-5-flare/g)).toHaveLength(1)
    expect(wrapper.find('input[placeholder="输入图片模型"]').exists()).toBe(false)
  })

  it('only exposes image-capable API keys', async () => {
    listKeys.mockResolvedValue({
      items: [
        {
          id: 1,
          name: 'text-only',
          key: 'sk-text',
          status: 'active',
          group_id: 1,
          group: { platform: 'anthropic', allow_image_generation: false },
        },
        {
          id: 2,
          name: 'image-key',
          key: 'sk-image',
          status: 'active',
          group_id: 2,
          group: { platform: 'openai', allow_image_generation: true },
        },
        {
          id: 3,
          name: 'ungrouped',
          key: 'sk-ungrouped',
          status: 'active',
          group_id: null,
        },
        {
          id: 4,
          name: 'legacy-group',
          key: 'sk-legacy',
          status: 'active',
          group_id: 4,
          group: { platform: 'openai' },
        },
      ],
    })
    getPublicSettings.mockResolvedValue({ api_base_url: '' })

    const wrapper = mountImageStudio()
    await flushPromises()

    const apiKeySelect = wrapper.get('[data-testid="api-key-select"]')
    expect(apiKeySelect.text()).toContain('image-key')
    expect(apiKeySelect.text()).not.toContain('text-only')

    await apiKeySelect.get('button').trigger('click')
    expect(apiKeySelect.text()).toContain('image-key')
    expect(apiKeySelect.text()).toContain('ungrouped')
    expect(apiKeySelect.text()).toContain('legacy-group')
    expect(apiKeySelect.text()).not.toContain('text-only')
  })

  it('renders keys before public settings finish loading', async () => {
    let resolveSettings: (value: any) => void = () => {}
    listKeys.mockResolvedValue({
      items: [
        {
          id: 1,
          name: 'dd',
          key: 'sk-local',
          status: 'active',
          group_id: 2,
        },
      ],
    })
    getPublicSettings.mockReturnValue(new Promise((resolve) => {
      resolveSettings = resolve
    }))

    const wrapper = mountImageStudio()
    await flushPromises()

    expect(wrapper.get('[data-testid="api-key-select"]').text()).toContain('dd')

    resolveSettings({ api_base_url: 'https://api.dihappy.cfd/v1' })
    await flushPromises()
  })

  it('keeps loaded keys visible when public settings fail', async () => {
    listKeys.mockResolvedValue({
      items: [
        {
          id: 1,
          name: 'dd',
          key: 'sk-local',
          status: 'active',
          group_id: 2,
        },
      ],
    })
    getPublicSettings.mockRejectedValue({
      status: 0,
      code: 'ERR_NETWORK',
      message: 'Network error. Please check your connection.',
    })

    const wrapper = mountImageStudio()
    await flushPromises()
    await flushPromises()

    expect(wrapper.get('[data-testid="api-key-select"]').text()).toContain('dd')
    expect(wrapper.text()).toContain('无法连接到 API 服务，请检查网络或 API 地址。')
  })

  it('shows insufficient balance errors in Chinese', async () => {
    listKeys.mockResolvedValue({
      items: [
        {
          id: 1,
          name: 'dd',
          key: 'sk-local',
          status: 'active',
          group_id: 2,
        },
      ],
    })
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://api.dihappy.cfd/v1' })
    vi.mocked(fetch).mockResolvedValue({
      ok: false,
      status: 402,
      statusText: 'Payment Required',
      json: async () => ({
        error: {
          code: 'INSUFFICIENT_BALANCE',
          message: 'Insufficient account balance',
        },
      }),
    } as Response)

    const wrapper = mountImageStudio()
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('账户余额不足')
    expect(wrapper.text()).not.toContain('Insufficient account balance')
  })

  it('does not show a redundant error when no image models are available', async () => {
    listKeys.mockResolvedValue({
      items: [
        {
          id: 1,
          name: 'dd',
          key: 'sk-local',
          status: 'active',
          group_id: 2,
        },
      ],
    })
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://api.dihappy.cfd/v1' })
    vi.mocked(fetch).mockResolvedValue({
      ok: true,
      json: async () => ({ data: [] }),
    } as Response)

    const wrapper = mountImageStudio()
    await flushPromises()
    await flushPromises()

    expect(wrapper.get('[data-testid="image-model-select"]').text()).toContain('未找到图片模型')
    expect(wrapper.text()).not.toContain('该 API Key 暂无可用图片模型')
  })
})

describe('ImageStudioView image size dialog', () => {
  beforeEach(() => {
    resetCommonMocks()
    listKeys.mockResolvedValue({ items: [] })
    getPublicSettings.mockResolvedValue({ api_base_url: '' })
    vi.stubGlobal('fetch', vi.fn())
  })

  it('opens the size dialog and applies custom dimensions', async () => {
    const wrapper = mountImageStudio()
    await flushPromises()

    const trigger = wrapper.get('[data-testid="image-size-trigger"]')
    expect(trigger.text()).toContain('1024x1024')

    await trigger.trigger('click')
    expect(wrapper.text()).toContain('设置图像尺寸')
    expect(wrapper.text()).toContain('自动')
    expect(wrapper.text()).toContain('按比例')
    expect(wrapper.text()).toContain('自定义宽高')

    await wrapper.get('[data-testid="size-mode-custom"]').trigger('click')
    await wrapper.get('[data-testid="custom-width"]').setValue('1280')
    await wrapper.get('[data-testid="custom-height"]').setValue('1024')
    await wrapper.get('[data-testid="confirm-image-size"]').trigger('click')

    expect(trigger.text()).toContain('1280x1024')
  })
})

describe('ImageStudioView gallery', () => {
  beforeEach(() => {
    resetCommonMocks()
    vi.stubGlobal('fetch', vi.fn())
    listKeys.mockResolvedValue({ items: [] })
    getPublicSettings.mockResolvedValue({ api_base_url: '' })
  })

  it('loads saved gallery images on entry', async () => {
    listGallery.mockResolvedValue([
      {
        id: 'gallery-1',
        url: 'https://cdn.example.com/gallery-1.png',
        prompt: 'a cat',
        model: 'gpt-image-1',
        size: '1024x1024',
        format: 'png',
        created_at: 1,
        expires_at: 9999999999,
      },
    ])

    const wrapper = mountImageStudio()
    await flushPromises()

    expect(wrapper.text()).toContain('历史图片 1')
    const image = wrapper.get('img[src="https://cdn.example.com/gallery-1.png"]')
    expect(image.exists()).toBe(true)
    expect(image.attributes('loading')).toBe('lazy')
  })

  it('opens the image detail dialog and copies the original prompt', async () => {
    listGallery.mockResolvedValue([
      {
        id: 'gallery-1',
        url: 'https://cdn.example.com/gallery-1.png',
        prompt: 'a cat wearing a spacesuit',
        model: 'gpt-image-1',
        size: '1024x1024',
        format: 'png',
        created_at: 1,
        expires_at: 9999999999,
      },
    ])

    const wrapper = mountImageStudio()
    await flushPromises()

    const imageButton = wrapper.findAll('button').find((button) =>
      button.find('img[src="https://cdn.example.com/gallery-1.png"]').exists(),
    )
    expect(imageButton).toBeDefined()
    await imageButton!.trigger('click')

    expect(wrapper.text()).toContain('提示词')
    expect(wrapper.text()).toContain('a cat wearing a spacesuit')

    const copyButton = wrapper.findAll('button').find((button) => button.text().includes('复制'))
    expect(copyButton).toBeDefined()
    await copyButton!.trigger('click')

    expect(copyToClipboard).toHaveBeenCalledWith('a cat wearing a spacesuit', expect.any(String))
  })

  it('closes the image detail dialog from the backdrop', async () => {
    listGallery.mockResolvedValue([
      {
        id: 'gallery-1',
        url: 'https://cdn.example.com/gallery-1.png',
        prompt: 'a cat',
        model: 'gpt-image-1',
        size: '1024x1024',
        format: 'png',
        created_at: 1,
        expires_at: 9999999999,
      },
    ])

    const wrapper = mountImageStudio()
    await flushPromises()

    const imageButton = wrapper.findAll('button').find((button) =>
      button.find('img[src="https://cdn.example.com/gallery-1.png"]').exists(),
    )
    await imageButton!.trigger('click')
    expect(wrapper.text()).toContain('a cat')

    await wrapper.get('[data-testid="dialog-backdrop"]').trigger('click')
    expect(wrapper.text()).not.toContain('a cat')
  })

  it('saves a generated image to the seven-day gallery', async () => {
    listKeys.mockResolvedValue({
      items: [{ id: 1, name: 'dd', key: 'sk-local', status: 'active', group_id: 2 }],
    })
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://api.dihappy.cfd/v1' })
    vi.mocked(fetch).mockImplementation(async (input) => {
      const url = String(input)
      if (url.endsWith('/images/batches/models')) {
        return { ok: true, json: async () => ({ data: [] }) } as Response
      }
      if (url.endsWith('/models')) {
        return { ok: true, json: async () => ({ data: [{ id: 'gpt-image-1' }] }) } as Response
      }
      if (url.endsWith('/images/generations')) {
        return { ok: true, json: async () => ({ data: [{ b64_json: 'aGVsbG8=' }] }) } as Response
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    const wrapper = mountImageStudio()
    await flushPromises()
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(saveGallery).toHaveBeenCalledWith(expect.objectContaining({
      image_data_url: expect.stringContaining('data:image/png;base64,'),
      model: 'gpt-image-1',
      size: '1024x1024',
      format: 'png',
    }))
  })

  it('does not show the freshly generated image twice when the gallery lists it', async () => {
    listKeys.mockResolvedValue({
      items: [{ id: 1, name: 'dd', key: 'sk-local', status: 'active', group_id: 2 }],
    })
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://api.dihappy.cfd/v1' })
    // 首次加载图库为空，保存之后图库里就出现了这条记录
    listGallery.mockResolvedValueOnce([])
    listGallery.mockResolvedValue([
      {
        id: 'gallery-1',
        url: 'https://cdn.example.com/gallery-1.png',
        prompt: 'a cat',
        model: 'gpt-image-1',
        size: '1024x1024',
        format: 'png',
        created_at: 1,
        expires_at: 2,
      },
    ])
    vi.mocked(fetch).mockImplementation(async (input) => {
      const url = String(input)
      if (url.endsWith('/images/batches/models')) {
        return { ok: true, json: async () => ({ data: [] }) } as Response
      }
      if (url.endsWith('/models')) {
        return { ok: true, json: async () => ({ data: [{ id: 'gpt-image-1' }] }) } as Response
      }
      if (url.endsWith('/images/generations')) {
        return { ok: true, json: async () => ({ data: [{ b64_json: 'aGVsbG8=' }] }) } as Response
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    const wrapper = mountImageStudio()
    await flushPromises()
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('生成结果 1')
    // 同一张图不能既作为生成结果显示、又以「历史图片」再显示一遍
    expect(wrapper.text()).not.toContain('历史图片')
  })

  it('shows and persists a pending image while generation is in flight', async () => {
    listKeys.mockResolvedValue({
      items: [{ id: 1, name: 'dd', key: 'sk-local', status: 'active', group_id: 2 }],
    })
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://api.dihappy.cfd/v1' })
    let resolveGeneration!: (response: Response) => void
    vi.mocked(fetch).mockImplementation(async (input) => {
      const url = String(input)
      if (url.endsWith('/images/batches/models')) {
        return { ok: true, json: async () => ({ data: [] }) } as Response
      }
      if (url.endsWith('/models')) {
        return { ok: true, json: async () => ({ data: [{ id: 'gpt-image-1' }] }) } as Response
      }
      if (url.endsWith('/images/generations')) {
        return await new Promise<Response>((resolve) => { resolveGeneration = resolve })
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    const wrapper = mountImageStudio()
    await flushPromises()
    await flushPromises()
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('正在生成')
    expect(wrapper.text()).not.toContain('可以切换到其他页面')
    expect(localStorage.getItem('sub2api:image-studio:pending')).toContain('"pending":true')

    resolveGeneration({
      ok: true,
      json: async () => ({ data: [{ b64_json: 'aGVsbG8=' }] }),
    } as Response)
    await flushPromises()

    expect(localStorage.getItem('sub2api:image-studio:pending')).toBeNull()
    expect(wrapper.text()).toContain('生成结果 1')
  })

  it('restores pending image placeholders from browser storage on entry', async () => {
    localStorage.setItem('sub2api:image-studio:pending', JSON.stringify([
      { id: 55, title: '正在生成', src: '', meta: 'PNG · 1024x1024 · gpt-image-1', pending: true },
    ]))

    const wrapper = mountImageStudio()
    await flushPromises()

    expect(wrapper.text()).toContain('正在生成')
  })
})

describe('ImageStudioView composer selects', () => {
  beforeEach(() => {
    resetCommonMocks()
    listKeys.mockResolvedValue({ items: [] })
    getPublicSettings.mockResolvedValue({ api_base_url: '' })
    vi.stubGlobal('fetch', vi.fn())
  })

  it('uses the soft dropdown for image quality', async () => {
    const wrapper = mountImageStudio()
    await flushPromises()

    const qualitySelect = wrapper.get('[data-testid="quality-select"]')
    expect(qualitySelect.text()).toContain('auto')
    expect(qualitySelect.classes()).not.toContain('studio-select--down')

    await qualitySelect.get('button').trigger('click')
    expect(qualitySelect.text()).toContain('low')
    expect(qualitySelect.text()).toContain('medium')
    expect(qualitySelect.text()).toContain('high')
  })

  it('uses a soft stepper for the image count', async () => {
    const wrapper = mountImageStudio()
    await flushPromises()

    const countControl = wrapper.get('[data-testid="image-count-control"]')
    const input = countControl.get('input')
    expect((input.element as HTMLInputElement).value).toBe('1')

    await countControl.get('[data-testid="count-increment"]').trigger('click')
    expect((input.element as HTMLInputElement).value).toBe('2')

    await countControl.get('[data-testid="count-decrement"]').trigger('click')
    expect((input.element as HTMLInputElement).value).toBe('1')
  })
})
