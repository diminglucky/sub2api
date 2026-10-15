import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import HomePricingPreview from '../HomePricingPreview.vue'

const getModelPlaza = vi.hoisted(() => vi.fn())
vi.mock('@/api/modelPlaza', () => ({ getModelPlaza }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key, locale: { value: 'en' } })
}))

const response = {
  description: '',
  groups: [
    { id: 1, name: 'OpenAI', platform: 'openai', models: [{ name: 'gpt-5', platform: 'openai' }] },
    { id: 2, name: 'Anthropic', platform: 'anthropic', models: [{ name: 'claude-sonnet', platform: 'anthropic' }] }
  ]
}

function mountPreview() {
  return mount(HomePricingPreview, {
    props: { visible: true, showEntry: true },
    global: {
      stubs: {
        Icon: true,
        RouterLink: true,
        PlazaGroupSection: {
          props: ['group'],
          template: '<div data-testid="group">{{ group.name }}: {{ group.models.map(model => model.name).join(", ") }}</div>'
        }
      }
    }
  })
}

describe('HomePricingPreview', () => {
  beforeEach(() => { getModelPlaza.mockReset() })
  afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks() })
  it('filters live pricing by platform and model, then refreshes it', async () => {
    getModelPlaza.mockResolvedValue(response)
    const wrapper = mountPreview()
    await flushPromises()
    expect(wrapper.findAll('[data-testid="group"]')).toHaveLength(2)

    await wrapper.get('select').setValue('openai')
    expect(wrapper.findAll('[data-testid="group"]')).toHaveLength(1)
    expect(wrapper.text()).toContain('gpt-5')

    await wrapper.get('input[type="search"]').setValue('missing')
    expect(wrapper.text()).toContain('modelPlaza.noSearchResult')

    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(getModelPlaza).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('shows a failed request and allows a retry', async () => {
    getModelPlaza.mockReset()
    getModelPlaza.mockRejectedValueOnce(new Error('unavailable')).mockResolvedValueOnce(response)
    const wrapper = mountPreview()
    await flushPromises()
    expect(wrapper.text()).toContain('modelPlaza.loadFailed')
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.findAll('[data-testid="group"]')).toHaveLength(2)
    wrapper.unmount()
  })

  it('shows added models after automatic refresh and stops refreshing after unmount', async () => {
    vi.useFakeTimers()
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
    getModelPlaza.mockResolvedValueOnce(response).mockResolvedValue({
      ...response,
      groups: [{ ...response.groups[0], models: [...response.groups[0].models, { name: 'new-model', platform: 'openai' }] }]
    })
    const wrapper = mountPreview()
    await flushPromises()
    expect(wrapper.text()).not.toContain('new-model')
    await vi.advanceTimersByTimeAsync(5 * 60 * 1000)
    await flushPromises()
    expect(wrapper.text()).toContain('new-model')
    wrapper.unmount()
    await vi.advanceTimersByTimeAsync(5 * 60 * 1000)
    expect(getModelPlaza).toHaveBeenCalledTimes(2)
  })

  it('preserves fetched prices when a refresh fails', async () => {
    getModelPlaza.mockResolvedValueOnce(response).mockRejectedValueOnce(new Error('network'))
    const wrapper = mountPreview()
    await flushPromises()
    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('gpt-5')
    expect(wrapper.get('[role="status"]').text()).toContain('homePricing.refreshFailed')
    wrapper.unmount()
  })

  it('can search models outside the initial 24-price preview and updates model suggestions per platform', async () => {
    getModelPlaza.mockResolvedValue({ ...response, groups: [{
      ...response.groups[0], models: Array.from({ length: 30 }, (_, index) => ({ name: `model-${index}`, platform: 'openai' }))
    }, response.groups[1]] })
    const wrapper = mountPreview()
    await flushPromises()
    expect(wrapper.findAll('datalist option')).toHaveLength(31)
    expect(wrapper.text()).not.toContain('model-29')
    await wrapper.get('input').setValue('model-29')
    expect(wrapper.get('[data-testid="group"]').text()).toContain('model-29')
    await wrapper.get('select').setValue('anthropic')
    expect(wrapper.findAll('datalist option')).toHaveLength(1)
    expect(wrapper.text()).toContain('modelPlaza.noSearchResult')
    wrapper.unmount()
  })
})
