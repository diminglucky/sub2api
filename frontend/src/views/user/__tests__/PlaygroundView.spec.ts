import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PlaygroundView from '../PlaygroundView.vue'

const { listKeys, getPublicSettings } = vi.hoisted(() => ({
  listKeys: vi.fn(),
  getPublicSettings: vi.fn(),
}))

vi.mock('@/api', () => ({
  keysAPI: {
    list: listKeys,
  },
  authAPI: {
    getPublicSettings,
  },
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('marked', () => ({
  marked: {
    setOptions: vi.fn(),
    parse: vi.fn((value: string) => value),
  },
}))

vi.mock('dompurify', () => ({
  default: {
    sanitize: vi.fn((value: string) => value),
  },
}))

const AppLayoutStub = { template: '<div><slot /></div>' }
const IconStub = { template: '<span />' }

function mountPlayground() {
  return mount(PlaygroundView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Icon: IconStub,
      },
    },
  })
}

describe('PlaygroundView model loading', () => {
  beforeEach(() => {
    listKeys.mockReset()
    getPublicSettings.mockReset()
    vi.stubGlobal('fetch', vi.fn())
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { hostname: 'localhost' },
    })
  })

  it('separates chat and image models from local /v1 on localhost', async () => {
    listKeys.mockResolvedValue({
      items: [
        {
          id: 1,
          name: '测试',
          key: 'sk-local',
          status: 'active',
          group_id: 2,
        },
      ],
    })
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://api.dihappy.cfd/v1' })
    const fetchMock = vi.mocked(fetch)
    fetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({
        object: 'list',
        data: [
          { id: 'gpt-5.5' },
          { id: 'gpt-5.4-mini' },
          { id: 'gpt-4o-audio-preview' },
          { id: 'gpt-4o-realtime-preview' },
          { id: 'codex-auto-review' },
          { id: 'gpt-image-1' },
          'gpt-image-2',
        ],
      }),
      headers: new Headers({ 'content-type': 'application/json' }),
    } as Response)

    const wrapper = mountPlayground()
    await flushPromises()
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledWith('/v1/models', {
      headers: { Authorization: 'Bearer sk-local' },
    })
    expect(wrapper.text()).toContain('gpt-image-2')
    expect(wrapper.text()).toContain('gpt-image-1')
    expect(wrapper.text()).not.toContain('gpt-5.5')
    expect(wrapper.text()).not.toContain('gpt-4o-audio-preview')
    expect(wrapper.text()).not.toContain('gpt-4o-realtime-preview')
    expect(wrapper.text()).not.toContain('codex-auto-review')

    const chatModeButton = wrapper.findAll('button').find((button) => button.text() === 'playground.chatMode')
    expect(chatModeButton).toBeTruthy()
    await chatModeButton!.trigger('click')

    expect(wrapper.text()).toContain('gpt-5.5')
    expect(wrapper.text()).not.toContain('gpt-image-1')
    expect(wrapper.text()).not.toContain('gpt-4o-audio-preview')
    expect(wrapper.text()).not.toContain('gpt-4o-realtime-preview')
    expect(wrapper.text()).not.toContain('playground.noChatModelsAvailable')
    expect(wrapper.text()).not.toContain('playground.noModelsAvailable')
  })

  it('sends reference images to /images/edits as multipart form data', async () => {
    listKeys.mockResolvedValue({
      items: [
        {
          id: 1,
          name: '测试',
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
      if (url.endsWith('/models')) {
        return {
          ok: true,
          json: async () => ({
            object: 'list',
            data: [{ id: 'gpt-image-1' }],
          }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response
      }
      if (url.endsWith('/images/edits')) {
        return {
          ok: true,
          json: async () => ({ data: [{ b64_json: 'aGVsbG8=' }] }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    const wrapper = mountPlayground()
    await flushPromises()
    await flushPromises()

    const fileInput = wrapper.get('input[type="file"]')
    const file = new File(['png'], 'reference.png', { type: 'image/png' })
    Object.defineProperty(fileInput.element, 'files', {
      configurable: true,
      value: [file],
    })
    await fileInput.trigger('change')
    for (let attempt = 0; attempt < 20; attempt += 1) {
      if (wrapper.find('.image-reference-chip').exists()) break
      await new Promise((resolve) => setTimeout(resolve, 5))
    }

    expect(wrapper.find('.image-reference-chip').exists()).toBe(true)
    await wrapper.get('form.image-composer').trigger('submit')
    await flushPromises()

    const editCall = fetchMock.mock.calls.find(([url]) => String(url).endsWith('/images/edits'))
    expect(editCall).toBeTruthy()

    const requestInit = editCall?.[1] as RequestInit
    expect(requestInit.body).toBeInstanceOf(FormData)
    expect((requestInit.body as FormData).getAll('image[]')).toHaveLength(1)
    expect((requestInit.headers as Record<string, string>)['Content-Type']).toBeUndefined()
  })

  it('sends Gemini image models through the native generateContent endpoint', async () => {
    listKeys.mockResolvedValue({
      items: [
        {
          id: 1,
          name: 'Gemini',
          key: 'sk-local',
          status: 'active',
          group_id: 2,
          group: { platform: 'gemini' },
        },
      ],
    })
    getPublicSettings.mockResolvedValue({ api_base_url: 'https://api.dihappy.cfd/v1' })
    const fetchMock = vi.mocked(fetch)
    fetchMock.mockImplementation(async (input) => {
      const url = String(input)
      if (url.endsWith('/models')) {
        return {
          ok: true,
          json: async () => ({
            object: 'list',
            data: [{ id: 'gemini-2.5-flash-image' }],
          }),
          headers: new Headers({ 'content-type': 'application/json' }),
        } as Response
      }
      if (url.includes('/v1beta/models/gemini-2.5-flash-image:streamGenerateContent')) {
        return {
          ok: true,
          text: async () => 'data: {"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/png","data":"aGVsbG8="}}]}}]}\n\n',
          headers: new Headers({ 'content-type': 'text/event-stream' }),
        } as Response
      }
      throw new Error(`Unexpected request: ${url}`)
    })

    const wrapper = mountPlayground()
    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('gemini-2.5-flash-image')
    await wrapper.get('form.image-composer').trigger('submit')
    await flushPromises()

    const generationCall = fetchMock.mock.calls.find(([url]) =>
      String(url).includes('/v1beta/models/gemini-2.5-flash-image:streamGenerateContent'),
    )
    expect(generationCall).toBeTruthy()

    const requestInit = generationCall?.[1] as RequestInit
    expect(requestInit.headers).toMatchObject({
      Authorization: 'Bearer sk-local',
      'Content-Type': 'application/json',
    })
    const body = JSON.parse(String(requestInit.body))
    expect(body.generationConfig.responseModalities).toEqual(['TEXT', 'IMAGE'])
    expect(body.generationConfig.imageConfig.aspectRatio).toBe('1:1')
    expect(wrapper.find('.studio-card').exists()).toBe(true)
  })
})
