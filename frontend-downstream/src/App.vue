<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getMeSummary, getSite, hasSession, logout, type SiteConfig } from './api'

const router = useRouter()
const site = ref<SiteConfig>({
  id: 0,
  slug: 'draw',
  domain: 'draw.superai.sbs',
  name: 'Draw',
  logo_url: '',
  theme_color: '#0f766e',
  api_base_url: 'https://draw.superai.sbs/v1'
})
const auth = ref(hasSession())
const isAdmin = ref(false)
const siteError = ref('')

const themeStyle = computed(() => ({
  '--site-color': site.value.theme_color || '#0f766e'
}))

onMounted(async () => {
  try {
    site.value = await getSite()
  } catch (error) {
    siteError.value = error instanceof Error ? error.message : 'Failed to load site'
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
  <div class="app-shell" :style="themeStyle">
    <header class="site-header">
      <RouterLink class="brand" to="/">
        <img v-if="site.logo_url" :src="site.logo_url" alt="" width="38" height="38" />
        <span v-else class="brand-mark">D</span>
        <span class="brand-name">{{ site.name }}</span>
      </RouterLink>

      <nav class="site-nav" aria-label="Primary">
        <RouterLink v-if="auth" class="nav-link" to="/dashboard">Dashboard</RouterLink>
        <RouterLink v-if="auth && isAdmin" class="nav-link" to="/admin">Admin</RouterLink>
        <button v-if="auth" class="text-button" type="button" @click="handleLogout">Sign out</button>
        <RouterLink v-else class="nav-link" to="/login">Sign in</RouterLink>
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
