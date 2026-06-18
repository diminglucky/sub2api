import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import UserCreateModal from '../UserCreateModal.vue'
import UserEditModal from '../UserEditModal.vue'
import type { AdminGroup, AdminUser } from '@/types'

const apiMocks = vi.hoisted(() => ({
  createUser: vi.fn(),
  updateUser: vi.fn(),
  getAllGroups: vi.fn(),
  updateUserAttributeValues: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      create: apiMocks.createUser,
      update: apiMocks.updateUser,
    },
    groups: {
      getAll: apiMocks.getAllGroups,
    },
    userAttributes: {
      updateUserAttributeValues: apiMocks.updateUserAttributeValues,
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
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

vi.mock('@/components/common/BaseDialog.vue', () => ({
  default: {
    name: 'BaseDialog',
    props: ['show', 'title', 'width'],
    template: '<div v-if="show"><slot /><slot name="footer" /></div>',
  },
}))

vi.mock('@/components/icons/Icon.vue', () => ({
  default: {
    name: 'Icon',
    template: '<span />',
  },
}))

vi.mock('@/components/user/UserAttributeForm.vue', () => ({
  default: {
    name: 'UserAttributeForm',
    props: ['modelValue'],
    emits: ['update:modelValue'],
    template: '<div />',
  },
}))

function makeExclusiveGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 9,
    name: 'VIP 专属',
    description: '',
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
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    allow_messages_dispatch: false,
    require_oauth_only: false,
    require_privacy_set: false,
    rpm_limit: 0,
    created_at: '2026-06-18T00:00:00Z',
    updated_at: '2026-06-18T00:00:00Z',
    model_routing: {},
    model_routing_enabled: false,
    mcp_xml_inject: false,
    default_mapped_model: '',
    messages_dispatch_model_config: {} as any,
    models_list_config: {} as any,
    supported_model_scopes: [],
    account_groups: [],
    account_count: 0,
    active_account_count: 0,
    rate_limited_account_count: 0,
    sort_order: 0,
    ...overrides,
  }
}

function makeUser(overrides: Partial<AdminUser> = {}): AdminUser {
  return {
    id: 7,
    username: 'vip-user',
    email: 'vip@example.com',
    role: 'user',
    balance: 0,
    concurrency: 1,
    status: 'active',
    allowed_groups: [9],
    balance_notify_enabled: false,
    balance_notify_threshold: null,
    balance_notify_extra_emails: [],
    created_at: '2026-06-18T00:00:00Z',
    updated_at: '2026-06-18T00:00:00Z',
    notes: '',
    rpm_limit: 0,
    ...overrides,
  }
}

describe('User create/edit modals exclusive group assignment', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    apiMocks.getAllGroups.mockResolvedValue([makeExclusiveGroup()])
    apiMocks.createUser.mockResolvedValue({})
    apiMocks.updateUser.mockResolvedValue({})
    apiMocks.updateUserAttributeValues.mockResolvedValue({})
  })

  it('submits selected exclusive groups when creating a user', async () => {
    const wrapper = mount(UserCreateModal, {
      props: { show: false },
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('new@example.com')
    await inputs[1].setValue('password123')

    const checkbox = wrapper.find('input[type="checkbox"]')
    expect(checkbox.exists()).toBe(true)
    await checkbox.setValue(true)

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(apiMocks.createUser).toHaveBeenCalledWith(
      expect.objectContaining({
        email: 'new@example.com',
        password: 'password123',
        allowed_groups: [9],
      })
    )
  })

  it('submits edited exclusive groups when updating a user', async () => {
    const wrapper = mount(UserEditModal, {
      props: {
        show: false,
        user: makeUser(),
      },
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    const checkbox = wrapper.find('input[type="checkbox"]')
    expect(checkbox.exists()).toBe(true)
    await checkbox.setValue(false)

    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(apiMocks.updateUser).toHaveBeenCalledWith(
      7,
      expect.objectContaining({
        allowed_groups: [],
      })
    )
  })
})
