<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getMeSummary, getSite, hasSession, logout, type SiteConfig } from './api'

const router = useRouter()
const route = useRoute()
const site = ref<SiteConfig>({
  id: 0,
  slug: 'draw',
  domain: 'draw.superai.sbs',
  name: 'Draw',
  logo_url: '',
  theme_color: '#fb6415',
  api_base_url: 'https://draw.superai.sbs/v1'
})
const auth = ref(hasSession())
const isAdmin = ref(false)
const siteError = ref('')

const themeStyle = computed(() => ({
  '--site-color': site.value.theme_color || '#fb6415'
}))
const isAuthRoute = computed(() => route.name === 'login' || route.name === 'register')

onMounted(async () => {
  try {
    site.value = await getSite()
  } catch (error) {
    siteError.value = error instanceof Error ? error.message : '站点信息加载失败'
  }
  auth.value = hasSession()
  await refreshAccountState()
})

async function handleLogout() {
  await logout()
  auth.value = false
  isAdmin.value = false
  router.push('/login')
}

function handleSessionChange(event: Event) {
  auth.value = Boolean((event as CustomEvent<boolean>).detail)
  void refreshAccountState()
}

async function refreshAccountState() {
  if (!hasSession()) {
    isAdmin.value = false
    return
  }
  try {
    const summary = await getMeSummary()
    isAdmin.value = summary.is_admin
  } catch {
    isAdmin.value = false
  }
}

if (typeof window !== 'undefined') {
  window.addEventListener('draw-session-change', handleSessionChange)
}

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('draw-session-change', handleSessionChange)
  }
})
</script>

<template>
  <div v-if="isAuthRoute" class="auth-app-shell" :style="themeStyle">
    <div v-if="siteError" class="notice auth-notice">{{ siteError }}</div>
    <RouterView v-slot="{ Component }">
      <component :is="Component" :site="site" />
    </RouterView>
  </div>

  <div v-else class="app-shell" :style="themeStyle">
    <header class="site-header">
      <RouterLink class="brand" to="/">
        <img v-if="site.logo_url" :src="site.logo_url" alt="" width="38" height="38" />
        <span v-else class="brand-mark">D</span>
        <span class="brand-name">{{ site.name }}</span>
      </RouterLink>

      <nav class="site-nav" aria-label="主导航">
        <RouterLink v-if="!auth" class="nav-link" to="/">首页</RouterLink>
        <a v-if="!auth" class="nav-link" href="/#features">绘图能力</a>
        <RouterLink v-if="auth" class="nav-link" to="/studio">绘图工作台</RouterLink>
        <RouterLink v-if="auth" class="nav-link" to="/dashboard">用户中心</RouterLink>
        <RouterLink v-if="auth && isAdmin" class="nav-link" to="/admin">管理看板</RouterLink>
        <button v-if="auth" class="text-button" type="button" @click="handleLogout">退出登录</button>
        <RouterLink v-else class="nav-link nav-link--primary" to="/login">控制台</RouterLink>
      </nav>
    </header>

    <main class="page">
      <div v-if="siteError" class="notice">{{ siteError }}</div>
      <RouterView v-slot="{ Component }">
        <component :is="Component" :site="site" />
      </RouterView>
    </main>
  </div>
</template>
