import { createRouter, createWebHistory, type RouteLocationNormalized, type RouteRecordRaw } from 'vue-router'
import { getMeSummary, hasSession } from './api'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import UserDashboardView from './views/UserDashboardView.vue'
import AdminDashboardView from './views/AdminDashboardView.vue'
import HomeView from './views/HomeView.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: HomeView },
  { path: '/login', name: 'login', component: LoginView },
  { path: '/register', name: 'register', component: RegisterView },
  { path: '/dashboard', name: 'dashboard', component: UserDashboardView, meta: { requiresAuth: true } },
  { path: '/admin', name: 'admin', component: AdminDashboardView, meta: { requiresAuth: true } }
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})

export async function downstreamRouteGuard(to: RouteLocationNormalized) {
  if (to.meta.requiresAuth && !hasSession()) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'admin' && hasSession()) {
    try {
      const summary = await getMeSummary()
      if (!summary.is_admin) {
        return { name: 'dashboard' }
      }
    } catch {
      return { name: 'dashboard' }
    }
  }
  if ((to.name === 'login' || to.name === 'register') && hasSession()) {
    return { name: 'dashboard' }
  }
  return true
}

router.beforeEach(downstreamRouteGuard)
