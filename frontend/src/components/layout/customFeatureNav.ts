export interface SidebarNavItem {
  path: string
  label: string
  icon: unknown
  iconSvg?: string
  hideInSimpleMode?: boolean
  children?: SidebarNavItem[]
  expandOnly?: boolean
  featureFlag?: () => boolean | undefined
}

type TranslateFn = (key: string) => string

interface UserCustomNavIcons {
  PlaygroundIcon: unknown
  ChannelIcon: unknown
  GiftIcon: unknown
  ManualIcon: unknown
}

interface AdminCustomNavIcons {
  GiftIcon: unknown
  ServerIcon: unknown
}

export function buildUserCustomNavAfterKeys(t: TranslateFn, icons: Pick<UserCustomNavIcons, 'PlaygroundIcon'>): SidebarNavItem[] {
  return [
    { path: '/playground', label: t('nav.playground'), icon: icons.PlaygroundIcon, hideInSimpleMode: true },
  ]
}

export function buildUserCustomNavAfterUsage(t: TranslateFn, icons: Pick<UserCustomNavIcons, 'ChannelIcon' | 'GiftIcon'>): SidebarNavItem[] {
  return [
    { path: '/models', label: t('nav.models'), icon: icons.ChannelIcon, hideInSimpleMode: true },
    { path: '/lottery', label: t('nav.lottery'), icon: icons.GiftIcon, hideInSimpleMode: true },
  ]
}

export function buildUserCustomNavBeforeProfile(t: TranslateFn, icons: Pick<UserCustomNavIcons, 'ManualIcon'>): SidebarNavItem[] {
  return [
    { path: '/manual', label: t('nav.manual'), icon: icons.ManualIcon, hideInSimpleMode: true },
  ]
}

export function buildAdminCustomNavAfterAnnouncements(t: TranslateFn, icons: Pick<AdminCustomNavIcons, 'GiftIcon'>): SidebarNavItem[] {
  return [
    { path: '/admin/lottery', label: t('nav.lotteryManagement'), icon: icons.GiftIcon },
  ]
}

export function buildAdminCustomNavAfterSettings(t: TranslateFn, icons: Pick<AdminCustomNavIcons, 'ServerIcon'>): SidebarNavItem[] {
  return [
    { path: '/admin/upstream-monitor', label: t('nav.upstreamMonitor'), icon: icons.ServerIcon },
  ]
}
