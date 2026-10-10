import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import HomeView from '../HomeView.vue'

const site = {
  id: 1,
  slug: 'draw',
  domain: 'draw.superai.sbs',
  name: 'Draw',
  logo_url: '',
  theme_color: '#fb6415',
  api_base_url: 'https://draw.superai.sbs/v1'
}

describe('HomeView', () => {
  it('renders the drawing entry and hands off to the main app', () => {
    const wrapper = mount(HomeView, { props: { site } })

    expect(wrapper.text()).toContain('生成你想要的图片')
    expect(wrapper.text()).toContain('https://draw.superai.sbs/v1')

    const hrefs = wrapper.findAll('a').map((link) => link.attributes('href'))
    // 入口页只负责跳转，用户页面由主站前端提供，因此必须走整页 href。
    expect(hrefs).toContain('/image-studio')
    expect(hrefs).toContain('/dashboard')
    expect(hrefs).toContain('/login')
  })
})
