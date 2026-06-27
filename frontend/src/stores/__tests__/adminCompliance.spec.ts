import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAdminComplianceStore } from '../adminCompliance'

const { acceptCompliance } = vi.hoisted(() => ({
  acceptCompliance: vi.fn(),
}))

vi.mock('@/api/admin/compliance', () => ({
  default: {
    getStatus: vi.fn(),
    accept: acceptCompliance,
  },
}))

vi.mock('@/i18n', () => ({
  getLocale: () => 'zh',
}))

function status(required: boolean) {
  return {
    required,
    version: 'v-test',
    document_path_zh: 'docs/zh.md',
    document_path_en: 'docs/en.md',
    document_url_zh: 'https://example.test/zh',
    document_url_en: 'https://example.test/en',
    ack_phrase_zh: '确认',
    ack_phrase_en: 'confirm',
  }
}

describe('adminCompliance store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    acceptCompliance.mockReset()
  })

  it('dispatches an accepted event after acknowledgement is no longer required', async () => {
    const listener = vi.fn()
    const nextStatus = status(false)
    acceptCompliance.mockResolvedValue(nextStatus)
    window.addEventListener('admin-compliance-accepted', listener)

    try {
      await useAdminComplianceStore().accept('确认')
    } finally {
      window.removeEventListener('admin-compliance-accepted', listener)
    }

    expect(listener).toHaveBeenCalledTimes(1)
    expect(listener.mock.calls[0][0]).toBeInstanceOf(CustomEvent)
    expect((listener.mock.calls[0][0] as CustomEvent).detail).toEqual(nextStatus)
  })

  it('does not dispatch an accepted event when acknowledgement is still required', async () => {
    const listener = vi.fn()
    acceptCompliance.mockResolvedValue(status(true))
    window.addEventListener('admin-compliance-accepted', listener)

    try {
      await useAdminComplianceStore().accept('wrong phrase')
    } finally {
      window.removeEventListener('admin-compliance-accepted', listener)
    }

    expect(listener).not.toHaveBeenCalled()
  })
})
