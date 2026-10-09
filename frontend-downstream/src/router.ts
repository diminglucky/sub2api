import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { hasSession } from './api'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import UserDashboardView from './views/UserDashboardView.vue'
import AdminDashboardView from './views/AdminDashboardView.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', redirect: '/dashboard' },
  { path: '/login', name: 'login', component: LoginView },
  { path: '/register', name: 'register', component: RegisterView },
  { path: '/dashboard', name: 'dashboard', component: UserDashboardView, meta: { requiresAuth: true } },
  { path: '/admin', name: 'admin', component: AdminDashboardView, meta: { requiresAuth: true } }
]

export const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !hasSession()) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if ((to.name === 'login' || to.name === 'register') && hasSession()) {
    return { name: 'dashboard' }
  }
  return true
})
