import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import ImageStudioView from '../ImageStudioView.vue'

const { listKeys, getPublicSettings, listGallery, saveGallery } = vi.hoisted(() => ({
  listKeys: vi.fn(),
  getPublicSettings: vi.fn(),
  listGallery: vi.fn(),
  saveGallery: vi.fn(),
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

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../ImageStudioView.vue')
const componentSource = readFileSync(componentPath, 'utf8')

function mountImageStudio() {
  return mount(ImageStudioView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        BaseDialog: {
          props: ['show', 'title'],
          template: '<div v-if="show"><h3>{{ title }}</h3><slot /><slot name="footer" /></div>',
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
  listGallery.mockResolvedValue([])
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

  it('automatically loads image models for the selected API key', async () => {
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
          json: async () => ({ data: [{ id: 'nano-banana-2' }] }),
        } as Response
      }
      if (url.endsWith('/models')) {
        return {
          ok: true,
          json: async () => ({ data: [{ id: 'gpt-image-1' }, { id: 'gpt-5' }] }),
        } as Response
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    const wrapper = mountImageStudio()
    await flushPromises()
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledWith('/v1/images/batches/models', {
      headers: { Authorization: 'Bearer sk-local' },
    })
    expect(fetchMock).toHaveBeenCalledWith('/v1/models', {
      headers: { Authorization: 'Bearer sk-local' },
    })

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
    expect(modelSelect.text()).toContain('nano-banana-2')
    expect(modelSelect.text()).toContain('gpt-image-1')
    expect(modelSelect.text()).not.toContain('gpt-5')
    expect(wrapper.find('input[placeholder="输入图片模型"]').exists()).toBe(false)
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
    expect(wrapper.get('img[src="https://cdn.example.com/gallery-1.png"]').exists()).toBe(true)
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
