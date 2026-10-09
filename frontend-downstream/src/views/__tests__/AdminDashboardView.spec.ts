import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AdminDashboardView from '../AdminDashboardView.vue'

const { getAdminSummary } = vi.hoisted(() => ({
  getAdminSummary: vi.fn()
}))

vi.mock('../../api', () => ({
  getAdminSummary
}))

describe('AdminDashboardView', () => {
  beforeEach(() => {
    getAdminSummary.mockReset()
  })

  it('renders only the draw subsite summary returned by the scoped API', async () => {
    getAdminSummary.mockResolvedValue({
      subsite_id: 7,
      member_count: 12,
      recharge_count: 4,
      recharge_amount: 80,
      usage_count: 99,
      usage_cost: 21.5,
      settlement_pending: 3.25
    })

    const wrapper = mount(AdminDashboardView, {
      props: {
        site: {
          id: 7,
          slug: 'draw',
          domain: 'draw.superai.sbs',
          name: 'Draw',
          logo_url: '',
          theme_color: '#0f766e',
          api_base_url: 'https://draw.superai.sbs/v1'
        }
      }
    })

    await flushPromises()

    expect(getAdminSummary).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('Draw')
    expect(wrapper.text()).toContain('12')
    expect(wrapper.text()).toContain('99')
    expect(wrapper.text()).toContain('21.50')
    expect(wrapper.text()).toContain('3.25')
  })
})
