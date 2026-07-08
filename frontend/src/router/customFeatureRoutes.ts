import type { RouteRecordRaw } from 'vue-router'

export const publicCustomFeatureRoutes: RouteRecordRaw[] = [
  {
    path: '/public-models',
    name: 'PublicModels',
    component: () => import('@/views/public/PublicModelsView.vue'),
    meta: {
      requiresAuth: false,
      title: 'Models',
      titleKey: 'models.title'
    }
  }
]

export const userCustomFeatureRoutesAfterKeys: RouteRecordRaw[] = [
  {
    path: '/playground',
    name: 'Playground',
    component: () => import('@/views/user/PlaygroundView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: false,
      title: 'Playground',
      titleKey: 'playground.title',
      descriptionKey: 'playground.description'
    }
  }
]

export const userCustomFeatureRoutesAfterRedeem: RouteRecordRaw[] = [
  {
    path: '/lottery',
    name: 'Lottery',
    component: () => import('@/views/user/LotteryView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: false,
      title: 'Lottery',
      titleKey: 'lottery.title',
      descriptionKey: 'lottery.description'
    }
  }
]

export const userCustomFeatureRoutesAfterAvailableChannels: RouteRecordRaw[] = [
  {
    path: '/manual',
    name: 'UserManual',
    component: () => import('@/views/user/ManualView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: false,
      title: 'User Manual',
      titleKey: 'manual.title',
      descriptionKey: 'manual.description'
    }
  },
  {
    path: '/manual/:platform',
    name: 'UserManualPlatform',
    component: () => import('@/views/user/ManualView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: false,
      title: 'User Manual',
      titleKey: 'manual.title',
      descriptionKey: 'manual.description'
    }
  },
  {
    path: '/models',
    name: 'UserModels',
    component: () => import('@/views/user/ModelsView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: false,
      title: 'Models',
      titleKey: 'models.title',
      descriptionKey: 'models.description'
    }
  }
]

export const adminCustomFeatureRoutesAfterAnnouncements: RouteRecordRaw[] = [
  {
    path: '/admin/lottery',
    name: 'AdminLottery',
    component: () => import('@/views/admin/LotteryView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Lottery Management',
      titleKey: 'admin.lottery.title',
      descriptionKey: 'admin.lottery.description'
    }
  }
]

export const adminCustomFeatureRoutesAfterSettings: RouteRecordRaw[] = [
  {
    path: '/admin/upstream-monitor',
    name: 'AdminUpstreamMonitor',
    component: () => import('@/views/admin/UpstreamMonitorView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Upstream Monitor',
      titleKey: 'admin.upstreamMonitor.title',
      descriptionKey: 'admin.upstreamMonitor.description'
    }
  }
]
