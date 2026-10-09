import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { RouteLocationNormalized } from 'vue-router'

const { getMeSummary, hasSession } = vi.hoisted(() => ({
  getMeSummary: vi.fn(),
  hasSession: vi.fn()
}))

vi.mock('../api', () => ({
  getMeSummary,
  hasSession
}))

import { downstreamRouteGuard } from '../router'

function route(name: string, requiresAuth = true): RouteLocationNormalized {
  return {
    name,
    fullPath: `/${name}`,
    meta: { requiresAuth },
    path: `/${name}`
  } as RouteLocationNormalized
}

describe('downstreamRouteGuard', () => {
  beforeEach(() => {
    getMeSummary.mockReset()
    hasSession.mockReset()
  })

  it('redirects unauthenticated admin requests to login', async () => {
    hasSession.mockReturnValue(false)

    await expect(downstreamRouteGuard(route('admin'))).resolves.toEqual({
      name: 'login',
      query: { redirect: '/admin' }
    })
  })

  it('blocks non-admin users from the admin route', async () => {
    hasSession.mockReturnValue(true)
    getMeSummary.mockResolvedValue({ is_admin: false })

    await expect(downstreamRouteGuard(route('admin'))).resolves.toEqual({ name: 'dashboard' })
  })

  it('allows subsite admins through the admin route', async () => {
    hasSession.mockReturnValue(true)
    getMeSummary.mockResolvedValue({ is_admin: true })

    await expect(downstreamRouteGuard(route('admin'))).resolves.toBe(true)
  })
})
