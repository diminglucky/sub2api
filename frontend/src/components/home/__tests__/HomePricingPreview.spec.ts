import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import HomePricingPreview from '../HomePricingPreview.vue'

const getModelPlaza = vi.hoisted(() => vi.fn())
vi.mock('@/api/modelPlaza', () => ({ getModelPlaza }))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key })
}))

function group(id: number, name: string, platform: string, rate: number, models: string[]) {
  return {
    id,
    name,
    description: '',
    platform,
    subscription_type: 'standard',
    rate_multiplier: rate,
    peak_rate_enabled: false,
    peak_start: '',
    peak_end: '',
    peak_rate_multiplier: 1,
    is_exclusive: false,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    video_rate_independent: false,
    video_rate_multiplier: 1,
    long_context_pricing_enabled: false,
    models: models.map(model => ({ name: model, platform, pricing: null, official_pricing: null }))
  }
}

const response = {
  description: '',
  groups: [
    group(1, 'OpenAI 标准', 'openai', 0.2, ['gpt-5', 'gpt-5-mini']),
    group(2, 'Anthropic', 'anthropic', 0.05, ['claude-sonnet']),
    group(3, 'OpenAI 特价', 'openai', 0.08, ['gpt-5.1-codex'])
  ]
}

function mountPreview() {
  return mount(HomePricingPreview, {
    props: { visible: true, showEntry: true },
    global: { stubs: { Icon: true, RouterLink: true } }
  })
}

describe('HomePricingPreview', () => {
  beforeEach(() => { getModelPlaza.mockReset() })
  afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks() })

  it('shows only the lowest GPT rate, ignoring cheaper non-GPT groups', async () => {
    getModelPlaza.mockResolvedValue(response)
    const wrapper = mountPreview()
    await flushPromises()

    const card = wrapper.get('[data-testid="lowest-gpt-rate"]')
    expect(card.text()).toContain('0.08x')
    expect(card.text()).toContain('OpenAI 特价')
    expect(card.text()).toContain('gpt-5.1-codex')
    // Anthropic 分组倍率更低（0.05x），但不是 GPT，不应被选中
    expect(card.text()).not.toContain('claude-sonnet')
    expect(card.text()).not.toContain('0.05x')
    wrapper.unmount()
  })

  it('shows a failed request instead of a rate', async () => {
    getModelPlaza.mockRejectedValue(new Error('unavailable'))
    const wrapper = mountPreview()
    await flushPromises()

    expect(wrapper.text()).toContain('modelPlaza.loadFailed')
    expect(wrapper.find('[data-testid="lowest-gpt-rate"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('shows the empty state when no GPT pricing is published', async () => {
    getModelPlaza.mockResolvedValue({ description: '', groups: [group(2, 'Anthropic', 'anthropic', 0.05, ['claude-sonnet'])] })
    const wrapper = mountPreview()
    await flushPromises()

    expect(wrapper.text()).toContain('homePricing.gptEmptyTitle')
    expect(wrapper.find('[data-testid="lowest-gpt-rate"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('refreshes the rate on the timer and stops after unmount', async () => {
    vi.useFakeTimers()
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
    getModelPlaza
      .mockResolvedValueOnce(response)
      .mockResolvedValue({ description: '', groups: [group(4, 'OpenAI 闪购', 'openai', 0.03, ['gpt-5'])] })
    const wrapper = mountPreview()
    await flushPromises()
    expect(wrapper.get('[data-testid="lowest-gpt-rate"]').text()).toContain('0.08x')

    await vi.advanceTimersByTimeAsync(5 * 60 * 1000)
    await flushPromises()
    expect(wrapper.get('[data-testid="lowest-gpt-rate"]').text()).toContain('0.03x')

    wrapper.unmount()
    await vi.advanceTimersByTimeAsync(5 * 60 * 1000)
    expect(getModelPlaza).toHaveBeenCalledTimes(2)
  })

  it('keeps the last rate when a refresh fails', async () => {
    vi.useFakeTimers()
    vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('visible')
    getModelPlaza.mockResolvedValueOnce(response).mockRejectedValue(new Error('network'))
    const wrapper = mountPreview()
    await flushPromises()

    await vi.advanceTimersByTimeAsync(5 * 60 * 1000)
    await flushPromises()

    expect(wrapper.get('[data-testid="lowest-gpt-rate"]').text()).toContain('0.08x')
    expect(wrapper.get('[role="status"]').text()).toContain('homePricing.refreshFailed')
    wrapper.unmount()
  })
})
