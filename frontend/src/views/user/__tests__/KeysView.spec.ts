import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import KeysView from '../KeysView.vue'
import type { Group } from '@/types'

const {
  listApiKeys,
  getDashboardApiKeysUsage,
  getAvailableGroups,
  getUserGroupRates,
  getPublicSettings,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  listApiKeys: vi.fn(),
  getDashboardApiKeysUsage: vi.fn(),
  getAvailableGroups: vi.fn(),
  getUserGroupRates: vi.fn(),
  getPublicSettings: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api', () => ({
  keysAPI: {
    list: listApiKeys,
    update: vi.fn(),
    create: vi.fn(),
    delete: vi.fn(),
    toggleStatus: vi.fn(),
  },
  usageAPI: {
    getDashboardApiKeysUsage,
  },
  authAPI: {
    getPublicSettings,
  },
  userGroupsAPI: {
    getAvailable: getAvailableGroups,
    getUserGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('@/stores/onboarding', () => ({
  useOnboardingStore: () => ({
    isCurrentStep: vi.fn(() => false),
    nextStep: vi.fn(),
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true),
  }),
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

const AppLayoutStub = { template: '<div><slot /></div>' }
const TablePageLayoutStub = {
  template: '<div><slot name="actions" /><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>',
}
const DataTableStub = {
  props: ['data'],
  template: '<div><slot name="empty" /></div>',
}
const SelectStub = {
  props: ['options', 'modelValue'],
  emits: ['update:modelValue'],
  template: '<select><option v-for="opt in options" :key="String(opt.value)" :value="opt.value">{{ opt.label }}</option></select>',
}
const GroupBadgeStub = {
  props: ['name'],
  template: '<span data-test="group-badge">{{ name }}</span>',
}

function group(overrides: Partial<Group>): Group {
  return {
    id: 1,
    name: 'VIP',
    description: null,
    platform: 'openai',
    rate_multiplier: 0.8,
    is_exclusive: true,
    status: 'active',
    subscription_type: 'standard',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    allow_image_generation: false,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    ...overrides,
  }
}

function mountKeysView() {
  return mount(KeysView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        TablePageLayout: TablePageLayoutStub,
        DataTable: DataTableStub,
        Pagination: true,
        BaseDialog: true,
        ConfirmDialog: true,
        EmptyState: true,
        Select: SelectStub,
        SearchInput: true,
        Icon: true,
        UseKeyModal: true,
        EndpointPopover: true,
        GroupBadge: GroupBadgeStub,
        GroupOptionItem: true,
        Teleport: true,
      },
    },
  })
}

describe('user KeysView exclusive groups panel', () => {
  beforeEach(() => {
    listApiKeys.mockReset()
    getDashboardApiKeysUsage.mockReset()
    getAvailableGroups.mockReset()
    getUserGroupRates.mockReset()
    getPublicSettings.mockReset()
    showError.mockReset()
    showSuccess.mockReset()

    listApiKeys.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20, pages: 0 })
    getDashboardApiKeysUsage.mockResolvedValue({ stats: {} })
    getUserGroupRates.mockResolvedValue({ 9: 0.6 })
    getPublicSettings.mockResolvedValue({})
  })

  it('shows exclusive groups granted to the current user', async () => {
    getAvailableGroups.mockResolvedValue([
      group({ id: 9, name: 'VIP 专属', is_exclusive: true }),
      group({ id: 2, name: '公开分组', is_exclusive: false }),
    ])

    const wrapper = mountKeysView()
    await flushPromises()

    expect(wrapper.find('[data-test="exclusive-groups-panel"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="exclusive-groups-panel"]').text()).toContain('VIP 专属')
    expect(wrapper.find('[data-test="exclusive-groups-panel"]').text()).not.toContain('公开分组')
  })

  it('hides the panel when the user has no exclusive groups', async () => {
    getAvailableGroups.mockResolvedValue([
      group({ id: 2, name: '公开分组', is_exclusive: false }),
    ])

    const wrapper = mountKeysView()
    await flushPromises()

    expect(wrapper.find('[data-test="exclusive-groups-panel"]').exists()).toBe(false)
  })
})
