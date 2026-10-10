import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import HomeView from './views/HomeView.vue'

// 子站只提供入口页。其余路由（/login、/dashboard、/keys、/image-studio …）
// 由后端直接返回主站前端，因此用户页面与主站完全一致，不做重复实现。
const routes: RouteRecordRaw[] = [{ path: '/', name: 'home', component: HomeView }]

export const router = createRouter({
  history: createWebHistory(),
  routes
})
