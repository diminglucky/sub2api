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
const ImageUploadStub = {
  props: ['modelValue'],
  emits: ['update:modelValue'],
  template: '<div />',
}

function mountPlayground() {
  return mount(PlaygroundView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Icon: IconStub,
        ImageUpload: ImageUploadStub,
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
})
