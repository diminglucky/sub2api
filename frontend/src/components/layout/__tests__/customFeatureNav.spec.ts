import { describe, expect, it } from 'vitest'

import {
  buildAdminCustomNavAfterAnnouncements,
  buildAdminCustomNavAfterSettings,
  buildUserCustomNavAfterKeys,
  buildUserCustomNavAfterUsage,
  buildUserCustomNavBeforeProfile,
} from '../customFeatureNav'

const t = (key: string) => key
const icon = {}

describe('custom feature sidebar navigation', () => {
  it('keeps user custom nav entries in stable groups', () => {
    expect(buildUserCustomNavAfterKeys(t, { PlaygroundIcon: icon }).map((item) => item.path)).toEqual(['/playground'])
    expect(buildUserCustomNavAfterUsage(t, { ChannelIcon: icon, GiftIcon: icon }).map((item) => item.path)).toEqual([
      '/models',
      '/lottery',
    ])
    expect(buildUserCustomNavBeforeProfile(t, { ManualIcon: icon }).map((item) => item.path)).toEqual(['/manual'])
  })

  it('keeps admin custom nav entries in stable groups', () => {
    expect(buildAdminCustomNavAfterAnnouncements(t, { GiftIcon: icon }).map((item) => item.path)).toEqual([
      '/admin/lottery',
    ])
    expect(buildAdminCustomNavAfterSettings(t, { ServerIcon: icon }).map((item) => item.path)).toEqual([
      '/admin/upstream-monitor',
    ])
  })
})
