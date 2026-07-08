import { describe, expect, it } from 'vitest'

import {
  adminCustomFeatureRoutesAfterAnnouncements,
  adminCustomFeatureRoutesAfterSettings,
  publicCustomFeatureRoutes,
  userCustomFeatureRoutesAfterAvailableChannels,
  userCustomFeatureRoutesAfterKeys,
  userCustomFeatureRoutesAfterRedeem,
} from '../customFeatureRoutes'
import { routes } from '../index'

function expectPathOrder(paths: string[], orderedPaths: string[]) {
  const indexes = orderedPaths.map((path) => paths.indexOf(path))
  expect(indexes.every((index) => index >= 0)).toBe(true)
  expect([...indexes].sort((a, b) => a - b)).toEqual(indexes)
}

describe('custom feature routes', () => {
  it('keeps public custom routes registered', () => {
    expect(publicCustomFeatureRoutes.map((route) => route.path)).toEqual(['/public-models'])
  })

  it('keeps user custom routes registered at stable insertion points', () => {
    expect(userCustomFeatureRoutesAfterKeys.map((route) => route.path)).toEqual(['/playground'])
    expect(userCustomFeatureRoutesAfterRedeem.map((route) => route.path)).toEqual(['/lottery'])
    expect(userCustomFeatureRoutesAfterAvailableChannels.map((route) => route.path)).toEqual([
      '/manual',
      '/manual/:platform',
      '/models',
    ])
  })

  it('keeps admin custom routes registered at stable insertion points', () => {
    expect(adminCustomFeatureRoutesAfterAnnouncements.map((route) => route.path)).toEqual(['/admin/lottery'])
    expect(adminCustomFeatureRoutesAfterSettings.map((route) => route.path)).toEqual(['/admin/upstream-monitor'])
  })

  it('mounts custom routes in the main router at stable positions', () => {
    const paths = routes.map((route) => route.path)

    expectPathOrder(paths, ['/key-usage', '/public-models', '/legal/:documentId'])
    expectPathOrder(paths, ['/keys', '/playground', '/usage'])
    expectPathOrder(paths, ['/redeem', '/lottery', '/affiliate'])
    expectPathOrder(paths, ['/available-channels', '/manual', '/manual/:platform', '/models', '/profile'])
    expectPathOrder(paths, ['/admin/announcements', '/admin/lottery', '/admin/proxies'])
    expectPathOrder(paths, ['/admin/settings', '/admin/upstream-monitor', '/admin/backups'])
  })
})
