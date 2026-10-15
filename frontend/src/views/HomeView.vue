<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="hasHomeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Compact Home Page -->
  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="flex min-h-screen flex-col bg-gray-50 text-gray-900 dark:bg-dark-950 dark:text-white"
  >
    <header class="border-b border-gray-200 px-4 py-4 sm:px-6 dark:border-dark-800">
      <nav class="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3 sm:gap-4">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <img
            :src="siteLogo || '/logo.svg'"
            alt="Logo"
            class="h-9 w-9 shrink-0 rounded-lg object-contain"
          />
          <span class="min-w-0 truncate text-base font-semibold">{{ siteName }}</span>
        </div>
        <div class="flex max-w-full shrink-0 flex-wrap items-center justify-end gap-2">
          <router-link to="/docs" class="inline-flex h-10 items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-800">
            <Icon name="book" size="sm" />{{ t('apiDocs.title') }}
          </router-link>
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <router-link
            v-if="showModelPlazaEntry"
            to="/model-plaza"
            class="flex h-10 shrink-0 items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('nav.modelPlaza')"
          >
            <Icon name="grid" size="md" />
            <span class="hidden sm:inline">{{ t('nav.modelPlaza') }}</span>
          </router-link>
          <button
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg text-gray-500 hover:bg-gray-100 dark:text-dark-400 dark:hover:bg-dark-800"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex min-h-10 shrink-0 items-center justify-center rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-200"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex min-w-0 flex-1 items-center justify-center px-4 py-16 sm:px-6">
      <div class="min-w-0 max-w-2xl text-center">
        <img
          :src="siteLogo || '/logo.svg'"
          alt="Logo"
          class="mx-auto mb-6 h-20 w-20 rounded-2xl object-contain"
        />
        <h1 class="[overflow-wrap:anywhere] text-3xl font-bold md:text-4xl">{{ siteName }}</h1>
        <p class="mt-4 whitespace-pre-wrap [overflow-wrap:anywhere] text-base text-gray-600 dark:text-dark-300">{{ siteSubtitle }}</p>
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="mt-8 inline-flex min-h-10 items-center justify-center rounded-lg bg-primary-600 px-5 py-2.5 text-sm font-medium text-white hover:bg-primary-700"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.login') }}
        </router-link>
      </div>
    </main>

    <footer class="min-w-0 border-t border-gray-200 px-4 py-5 text-center text-sm text-gray-500 [overflow-wrap:anywhere] sm:px-6 dark:border-dark-800 dark:text-dark-400">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <!-- Default Home Page -->
  <div v-else class="min-h-screen bg-[#fffdfb] text-[#171311] dark:bg-dark-950 dark:text-white">
    <header class="sticky inset-x-0 top-0 z-30 border-b border-[#eadfd8] bg-[#fffdfb]/90 backdrop-blur-xl dark:border-dark-800 dark:bg-dark-950/90">
      <nav class="mx-auto flex h-[72px] max-w-7xl items-center gap-4 px-4 sm:px-6 lg:px-8">
        <router-link to="/" class="flex min-w-0 items-center gap-2.5" aria-label="Home">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-8 w-8 rounded-lg object-contain" />
          <span class="truncate text-lg font-semibold tracking-tight">{{ siteName }}</span>
        </router-link>
        <div class="ml-auto hidden items-center gap-7 text-sm text-[#766c66] lg:flex dark:text-dark-300">
          <a href="#overview" class="transition-colors hover:text-[#171311] dark:hover:text-white">{{ siteName }}</a>
          <a href="#features" class="transition-colors hover:text-[#171311] dark:hover:text-white">{{ t('home.features.unifiedGateway') }}</a>
          <a href="#providers" class="transition-colors hover:text-[#171311] dark:hover:text-white">{{ t('home.providers.title') }}</a>
          <a v-if="showModelPlazaEntry" href="#pricing" class="transition-colors hover:text-[#171311] dark:hover:text-white">{{ t('homePricing.title') }}</a>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="transition-colors hover:text-[#171311] dark:hover:text-white">{{ t('home.docs') }}</a>
          <router-link v-if="showModelPlazaEntry" to="/model-plaza" class="transition-colors hover:text-[#171311] dark:hover:text-white">{{ t('nav.modelPlaza') }}</router-link>
        </div>
        <div class="ml-auto flex items-center gap-1.5 sm:gap-2 lg:ml-0">
          <router-link to="/docs" class="inline-flex h-9 shrink-0 items-center gap-1.5 rounded-full px-2 text-xs font-medium text-[#766c66] transition hover:bg-[#f4ebe5] hover:text-[#171311] sm:px-3 sm:text-sm dark:text-dark-300 dark:hover:bg-dark-800 dark:hover:text-white">
            <Icon name="book" size="sm" class="hidden sm:block" />{{ t('apiDocs.title') }}
          </router-link>
          <LocaleSwitcher />
          <button class="flex h-9 w-9 items-center justify-center rounded-full text-[#766c66] transition hover:bg-[#f4ebe5] hover:text-[#171311] dark:text-dark-300 dark:hover:bg-dark-800 dark:hover:text-white" :title="isDark ? t('home.switchToLight') : t('home.switchToDark')" @click="toggleTheme">
            <Icon v-if="isDark" name="sun" size="sm" />
            <Icon v-else name="moon" size="sm" />
          </button>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="hidden h-9 items-center justify-center rounded-full bg-[#fb6415] px-4 text-sm font-medium text-black transition hover:bg-[#ff7b33] sm:inline-flex">
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="flex h-9 w-9 items-center justify-center rounded-full bg-[#fb6415] text-sm font-semibold text-black sm:hidden" :aria-label="isAuthenticated ? t('home.dashboard') : t('home.login')">
            <span v-if="isAuthenticated">{{ userInitial }}</span><Icon v-else name="arrowRight" size="sm" />
          </router-link>
        </div>
      </nav>
    </header>

    <main>
      <section id="overview" class="relative overflow-hidden border-b border-[#eadfd8] dark:border-dark-800">
        <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_8%_25%,rgba(251,100,21,0.12),transparent_30%),radial-gradient(circle_at_88%_20%,rgba(100,128,255,0.10),transparent_28%)]"></div>
        <div class="relative mx-auto grid max-w-7xl items-center gap-14 px-4 pb-20 pt-16 sm:px-6 lg:grid-cols-[minmax(0,1fr)_480px] lg:px-8 lg:pb-24 lg:pt-24">
          <div class="max-w-3xl">
            <p class="mb-5 inline-flex rounded-full border border-[#eadfd8] bg-white/70 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-[#8a7b72] dark:border-dark-700 dark:bg-dark-900/70 dark:text-dark-300">AI API Gateway</p>
            <h1 class="max-w-3xl text-6xl font-bold leading-[0.94] tracking-[-0.04em] text-[#171311] sm:text-7xl lg:text-8xl dark:text-white">
              <span class="block">{{ siteName }}</span>
              <span class="mt-2 block bg-gradient-to-r from-[#171311] via-[#926f5d] to-[#fb6415] bg-clip-text text-transparent dark:from-white dark:via-[#f7c8ad] dark:to-[#fb6415]">API GATEWAY.</span>
            </h1>
            <p class="mt-7 max-w-2xl text-base leading-8 text-[#766c66] sm:text-lg dark:text-dark-300">{{ siteSubtitle }}</p>
            <div class="mt-8 flex flex-col gap-3 sm:flex-row">
              <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="inline-flex h-11 items-center justify-center gap-2 rounded-full bg-[#fb6415] px-6 text-sm font-semibold text-black transition hover:bg-[#ff7b33]">
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}<Icon name="arrowRight" size="sm" />
              </router-link>
              <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="inline-flex h-11 items-center justify-center rounded-full border border-[#ded2ca] bg-white/70 px-6 text-sm font-medium text-[#3f3631] transition hover:bg-white dark:border-dark-700 dark:bg-dark-900/70 dark:text-dark-100 dark:hover:bg-dark-800">{{ t('home.viewDocs') }}</a>
              <a v-else-if="showModelPlazaEntry" href="#pricing" class="inline-flex h-11 items-center justify-center rounded-full border border-[#ded2ca] bg-white/70 px-6 text-sm font-medium text-[#3f3631] transition hover:bg-white dark:border-dark-700 dark:bg-dark-900/70 dark:text-dark-100 dark:hover:bg-dark-800">{{ t('homePricing.viewPricing') }}</a>
            </div>
          </div>
          <div class="flex justify-center lg:justify-end">
            <div class="terminal-container w-full max-w-[480px]">
              <div class="terminal-window w-full">
                <div class="terminal-header"><div class="terminal-buttons"><span class="btn-close"></span><span class="btn-minimize"></span><span class="btn-maximize"></span></div><span class="terminal-title">gateway request</span></div>
                <div class="terminal-body"><div class="code-line line-1"><span class="code-prompt">$</span><span class="code-cmd">curl</span><span class="code-flag">-X POST</span><span class="code-url">/v1/messages</span></div><div class="code-line line-2"><span class="code-comment"># Routing to upstream...</span></div><div class="code-line line-3"><span class="code-success">200 OK</span><span class="code-response">{ "content": "Hello!" }</span></div><div class="code-line line-4"><span class="code-prompt">$</span><span class="cursor"></span></div></div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="mx-auto max-w-7xl px-4 py-14 sm:px-6 lg:px-8">
        <div class="grid gap-3 sm:grid-cols-3">
          <div class="flex items-center gap-3 rounded-xl border border-[#eadfd8] bg-white/70 px-4 py-3 dark:border-dark-800 dark:bg-dark-900/70"><Icon name="swap" size="sm" class="text-[#fb6415]" /><span class="text-sm font-medium text-[#4e443e] dark:text-dark-200">{{ t('home.tags.subscriptionToApi') }}</span></div>
          <div class="flex items-center gap-3 rounded-xl border border-[#eadfd8] bg-white/70 px-4 py-3 dark:border-dark-800 dark:bg-dark-900/70"><Icon name="shield" size="sm" class="text-[#fb6415]" /><span class="text-sm font-medium text-[#4e443e] dark:text-dark-200">{{ t('home.tags.stickySession') }}</span></div>
          <div class="flex items-center gap-3 rounded-xl border border-[#eadfd8] bg-white/70 px-4 py-3 dark:border-dark-800 dark:bg-dark-900/70"><Icon name="chart" size="sm" class="text-[#fb6415]" /><span class="text-sm font-medium text-[#4e443e] dark:text-dark-200">{{ t('home.tags.realtimeBilling') }}</span></div>
        </div>
      </section>

      <HomePricingPreview :visible="showModelPlazaEntry" :show-entry="showModelPlazaEntry" :can-manage="isAdmin" />

      <section id="features" class="border-y border-[#eadfd8] bg-[#f8f3ef] py-20 dark:border-dark-800 dark:bg-dark-900/50">
        <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div class="mb-10 max-w-2xl"><p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#fb6415]">Core capabilities</p><h2 class="mt-3 text-3xl font-semibold tracking-tight text-[#171311] sm:text-4xl dark:text-white">{{ t('home.providers.title') }}</h2><p class="mt-4 text-base leading-7 text-[#766c66] dark:text-dark-300">{{ t('home.providers.description') }}</p></div>
          <div class="grid gap-4 md:grid-cols-3">
            <article class="rounded-2xl border border-[#eadfd8] bg-[#fffdfb] p-6 dark:border-dark-700 dark:bg-dark-900"><div class="mb-5 flex h-10 w-10 items-center justify-center rounded-xl bg-[#fff0e8] text-[#fb6415] dark:bg-[#4a2414]"><Icon name="server" size="md" /></div><h3 class="text-lg font-semibold">{{ t('home.features.unifiedGateway') }}</h3><p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('home.features.unifiedGatewayDesc') }}</p></article>
            <article class="rounded-2xl border border-[#eadfd8] bg-[#fffdfb] p-6 dark:border-dark-700 dark:bg-dark-900"><div class="mb-5 flex h-10 w-10 items-center justify-center rounded-xl bg-[#fff0e8] text-[#fb6415] dark:bg-[#4a2414]"><Icon name="users" size="md" /></div><h3 class="text-lg font-semibold">{{ t('home.features.multiAccount') }}</h3><p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('home.features.multiAccountDesc') }}</p></article>
            <article class="rounded-2xl border border-[#eadfd8] bg-[#fffdfb] p-6 dark:border-dark-700 dark:bg-dark-900"><div class="mb-5 flex h-10 w-10 items-center justify-center rounded-xl bg-[#fff0e8] text-[#fb6415] dark:bg-[#4a2414]"><Icon name="chart" size="md" /></div><h3 class="text-lg font-semibold">{{ t('home.features.balanceQuota') }}</h3><p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('home.features.balanceQuotaDesc') }}</p></article>
          </div>
        </div>
      </section>

      <section id="providers" class="mx-auto max-w-7xl px-4 py-20 sm:px-6 lg:px-8">
        <div class="grid gap-10 lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] lg:items-center"><div><p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#fb6415]">Supported models</p><h2 class="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">{{ t('home.providers.title') }}</h2><p class="mt-4 text-base leading-7 text-[#766c66] dark:text-dark-300">{{ t('home.providers.description') }}</p></div><div class="flex flex-wrap gap-3">
          <div v-for="provider in [{ key: 'claude', letter: 'C' }, { key: 'gpt', letter: 'G' }, { key: 'gemini', letter: 'G' }, { key: 'antigravity', letter: 'A' }]" :key="provider.key" class="flex items-center gap-3 rounded-xl border border-[#eadfd8] bg-white px-4 py-3 dark:border-dark-800 dark:bg-dark-900"><span class="flex h-8 w-8 items-center justify-center rounded-lg bg-[#fff0e8] text-sm font-bold text-[#fb6415] dark:bg-[#4a2414]">{{ provider.letter }}</span><span class="text-sm font-medium">{{ provider.key === 'gpt' ? 'GPT' : t(`home.providers.${provider.key}`) }}</span><span class="rounded-full bg-[#fff0e8] px-2 py-0.5 text-[10px] font-semibold text-[#c34c0d] dark:bg-[#4a2414] dark:text-[#ffb28d]">{{ t('home.providers.supported') }}</span>
          </div><div class="flex items-center gap-3 rounded-xl border border-dashed border-[#d9ccc4] px-4 py-3 text-[#8a7b72] dark:border-dark-700 dark:text-dark-400"><span class="flex h-8 w-8 items-center justify-center rounded-lg bg-[#f4ebe5] text-sm font-bold dark:bg-dark-800">+</span><span class="text-sm">{{ t('home.providers.more') }}</span></div>
        </div></div>
      </section>

      <section class="mx-auto max-w-7xl px-4 pb-20 sm:px-6 lg:px-8"><div class="rounded-[1.75rem] bg-[#1f292d] px-6 py-12 text-center text-[#fffaf0] sm:px-10"><p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#fb6415]">{{ t('home.getStarted') }}</p><h2 class="mt-4 text-3xl font-semibold sm:text-4xl">{{ siteName }} API Gateway</h2><p class="mx-auto mt-4 max-w-2xl text-sm leading-7 text-[#d8e0e2]">{{ siteSubtitle }}</p><router-link :to="isAuthenticated ? dashboardPath : '/login'" class="mt-8 inline-flex h-11 items-center justify-center gap-2 rounded-full bg-[#fb6415] px-6 text-sm font-semibold text-[#172126] transition hover:bg-[#ff7b33]">{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}<Icon name="arrowRight" size="sm" /></router-link></div></section>
    </main>

    <footer class="border-t border-[#eadfd8] px-4 py-8 dark:border-dark-800"><div class="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 text-center text-sm text-[#8a7b72] sm:flex-row sm:text-left dark:text-dark-400"><p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p><a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="transition hover:text-[#fb6415]">{{ t('home.docs') }}</a></div></footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import HomePricingPreview from '@/components/home/HomePricingPreview.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'SuperAI')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const compactHomeEnabled = computed(() => appStore.cachedPublicSettings?.compact_home_enabled === true)
const modelPlazaEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.modelPlaza))

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const modelPlazaRequiresAuth = computed(
  () => appStore.cachedPublicSettings?.model_plaza_require_auth === true,
)
const showModelPlazaEntry = computed(
  () => modelPlazaEnabled.value && (isAuthenticated.value || !modelPlazaRequiresAuth.value),
)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
/* Terminal Container */
.terminal-container {
  position: relative;
  display: inline-block;
}

/* Terminal Window */
.terminal-window {
  width: 420px;
  background: linear-gradient(145deg, #1e293b 0%, #0f172a 100%);
  border-radius: 14px;
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.4),
    0 0 0 1px rgba(255, 255, 255, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  overflow: hidden;
  transform: perspective(1000px) rotateX(2deg) rotateY(-2deg);
  transition: transform 0.3s ease;
}

.terminal-window:hover {
  transform: perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(-4px);
}

/* Terminal Header */
.terminal-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: rgba(30, 41, 59, 0.8);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.btn-close {
  background: #ef4444;
}
.btn-minimize {
  background: #eab308;
}
.btn-maximize {
  background: #22c55e;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 12px;
  font-family: ui-monospace, monospace;
  color: #64748b;
  margin-right: 52px;
}

/* Terminal Body */
.terminal-body {
  padding: 20px 24px;
  font-family: ui-monospace, 'Fira Code', monospace;
  font-size: 14px;
  line-height: 2;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  opacity: 0;
  animation: line-appear 0.5s ease forwards;
}

.line-1 {
  animation-delay: 0.3s;
}
.line-2 {
  animation-delay: 1s;
}
.line-3 {
  animation-delay: 1.8s;
}
.line-4 {
  animation-delay: 2.5s;
}

@keyframes line-appear {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.code-prompt {
  color: #22c55e;
  font-weight: bold;
}
.code-cmd {
  color: #38bdf8;
}
.code-flag {
  color: #a78bfa;
}
.code-url {
  color: #14b8a6;
}
.code-comment {
  color: #64748b;
  font-style: italic;
}
.code-success {
  color: #22c55e;
  background: rgba(34, 197, 94, 0.15);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
}
.code-response {
  color: #fbbf24;
}

/* Blinking Cursor */
.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: #22c55e;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}

/* Dark mode adjustments */
:deep(.dark) .terminal-window {
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.6),
    0 0 0 1px rgba(20, 184, 166, 0.2),
    0 0 40px rgba(20, 184, 166, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}
</style>
