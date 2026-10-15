<template>
  <div class="min-h-screen bg-[#fffdfb] text-[#171311] dark:bg-dark-950 dark:text-white">
    <header class="sticky top-0 z-30 border-b border-[#eadfd8] bg-[#fffdfb]/95 backdrop-blur-xl dark:border-dark-800 dark:bg-dark-950/95">
      <nav class="mx-auto flex h-[72px] max-w-7xl items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
        <router-link to="/home" class="flex min-w-0 items-center gap-2.5">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-8 w-8 shrink-0 rounded-lg object-contain" />
          <span class="truncate font-semibold">{{ siteName }}</span>
          <span class="hidden border-l border-[#ded2ca] pl-3 text-sm text-[#8a7b72] sm:inline dark:border-dark-700">{{ t('apiDocs.title') }}</span>
        </router-link>
        <div class="flex shrink-0 items-center gap-2">
          <router-link to="/home" class="hidden text-sm text-[#766c66] hover:text-[#fb6415] sm:mr-3 sm:inline dark:text-dark-300">{{ t('apiDocs.home') }}</router-link>
          <LocaleSwitcher />
          <button class="flex h-9 w-9 items-center justify-center rounded-full text-[#766c66] hover:bg-[#f4ebe5] dark:text-dark-300 dark:hover:bg-dark-800" :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')" @click="toggleTheme">
            <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
          </button>
          <router-link :to="authStore.isAuthenticated ? '/keys' : { path: '/login', query: { redirect: '/keys' } }" class="inline-flex h-9 items-center rounded-full bg-[#fb6415] px-4 text-sm font-medium text-black hover:bg-[#ff7b33]">{{ authStore.isAuthenticated ? t('apiDocs.createKey') : t('home.login') }}</router-link>
        </div>
      </nav>
    </header>

    <div class="mx-auto grid max-w-7xl gap-10 px-4 py-10 sm:px-6 lg:grid-cols-[200px_minmax(0,1fr)] lg:px-8 lg:py-14">
      <aside class="hidden lg:block">
        <nav :aria-label="t('apiDocs.title')" class="sticky top-28 space-y-1 text-sm">
          <p class="mb-4 px-3 text-[10px] font-semibold uppercase tracking-[0.2em] text-[#a69489]">Developers</p>
          <a v-for="section in sections" :key="section.id" :href="`#${section.id}`" class="flex rounded-lg px-3 py-2.5 text-[#766c66] hover:bg-[#fff0e8] hover:text-[#c34c0d] dark:text-dark-300 dark:hover:bg-dark-800">{{ t(section.label) }}</a>
          <div class="!mt-8 border-t border-[#eadfd8] px-3 pt-5 dark:border-dark-800">
            <router-link to="/keys" class="inline-flex items-center gap-2 text-[#c34c0d] dark:text-[#ffb28d]">{{ t('apiDocs.createKey') }} <Icon name="arrowRight" size="sm" /></router-link>
          </div>
        </nav>
      </aside>

      <main class="min-w-0 space-y-14">
        <section id="quickstart" class="scroll-mt-28">
          <p class="inline-flex rounded-full border border-[#eadfd8] bg-[#fff5ef] px-3 py-1 text-xs text-[#bd551e] dark:border-dark-700 dark:bg-dark-900 dark:text-[#ffb28d]">{{ t('apiDocs.public') }}</p>
          <h1 class="mt-6 text-4xl font-semibold leading-tight tracking-tight sm:text-5xl">{{ t('apiDocs.startTitle') }}</h1>
          <p class="mt-5 max-w-2xl text-base leading-8 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.startDescription') }}</p>
          <div class="mt-8 grid gap-4 md:grid-cols-3">
            <article v-for="step in [1, 2, 3]" :key="step" class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <span class="flex h-8 w-8 items-center justify-center rounded-full bg-[#fff0e8] font-mono text-xs font-semibold text-[#c34c0d] dark:bg-[#4a2414] dark:text-[#ffb28d]">0{{ step }}</span>
              <h2 class="mt-4 text-sm font-semibold">{{ t(`apiDocs.step${step}`) }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(`apiDocs.step${step}Description`) }}</p>
            </article>
          </div>
        </section>

        <section id="connection" class="scroll-mt-28">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.connection') }}</h2>
          <div class="mt-5 overflow-hidden rounded-2xl border border-[#eadfd8] bg-white dark:border-dark-800 dark:bg-dark-900">
            <div class="flex flex-wrap items-center justify-between gap-3 border-b border-[#eadfd8] px-5 py-3 dark:border-dark-800">
              <span class="text-xs font-semibold uppercase tracking-widest text-[#8a7b72]">Base URL</span>
              <button type="button" data-testid="copy-base" class="inline-flex items-center gap-1.5 text-xs text-[#766c66] hover:text-[#fb6415] dark:text-dark-300" @click="copyToClipboard(baseUrl)"><Icon name="copy" size="xs" />{{ t('common.copy') }}</button>
            </div>
            <code data-testid="docs-base-url" class="block break-all px-5 py-5 text-sm font-semibold text-[#c34c0d] dark:text-[#ffb28d]">{{ baseUrl }}</code>
          </div>
          <p class="mt-3 text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.baseHint') }}</p>
          <div class="mt-5 flex items-start gap-3 rounded-xl bg-[#f8f3ef] p-4 text-sm leading-7 text-[#766c66] dark:bg-dark-900 dark:text-dark-300"><Icon name="key" size="sm" class="mt-1 shrink-0 text-[#fb6415]" /><p>{{ t('apiDocs.keyHint') }}</p></div>
          <p class="mt-4 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.modelHint') }}</p>
        </section>

        <section id="examples" class="scroll-mt-28">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.examples') }}</h2>
          <p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.requestHint') }}</p>
          <div class="mt-5 flex flex-wrap gap-2" role="tablist" :aria-label="t('apiDocs.examples')">
            <button v-for="(spec, key) in docsEndpoints" :id="`endpoint-tab-${key}`" :key="key" type="button" role="tab" :aria-selected="endpoint === key" aria-controls="request-panel" :data-endpoint="key" :class="endpoint === key ? 'border-[#fb6415] bg-[#fff0e8] text-[#bd551e] dark:bg-[#4a2414] dark:text-[#ffb28d]' : 'border-[#eadfd8] bg-white text-[#766c66] hover:border-[#fb6415] dark:border-dark-700 dark:bg-dark-900 dark:text-dark-300'" class="rounded-full border px-4 py-2 text-sm font-medium transition" @click="endpoint = key">{{ spec.label }}</button>
          </div>
          <div id="request-panel" role="tabpanel" :aria-labelledby="`endpoint-tab-${endpoint}`" class="mt-4 overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
            <div class="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 px-5 py-4">
              <div class="flex min-w-0 items-center gap-3 text-sm"><span class="rounded bg-emerald-400/10 px-2 py-1 font-mono text-[11px] font-semibold text-emerald-300">{{ currentEndpoint.method }}</span><code class="break-all">/v1{{ currentEndpoint.path }}</code></div>
              <button type="button" data-testid="copy-example" class="inline-flex items-center gap-1.5 text-xs text-[#bdcccf] hover:text-white" @click="copyToClipboard(example)"><Icon :name="copied ? 'check' : 'copy'" size="xs" />{{ t(copied ? 'common.copied' : 'common.copy') }}</button>
            </div>
            <div class="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 px-5 py-3">
              <div class="flex gap-4" :aria-label="t('apiDocs.examples')">
                <button v-for="item in languages" :key="item.key" type="button" :data-language="item.key" :aria-pressed="language === item.key" class="py-1 text-xs transition" :class="language === item.key ? 'font-semibold text-[#ffb28d]' : 'text-[#9bafb4] hover:text-white'" @click="language = item.key">{{ item.label }}</button>
              </div>
              <label v-if="endpoint !== 'models'" class="flex items-center gap-2 text-xs text-[#bdcccf]"><input v-model="stream" type="checkbox" class="accent-[#fb6415]" />{{ t('apiDocs.streaming') }}</label>
            </div>
            <pre class="overflow-x-auto px-5 py-6 text-[12px] leading-6 sm:text-[13px]"><code data-testid="docs-example">{{ example }}</code></pre>
          </div>
          <p class="mt-3 text-xs leading-6 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.environment') }}</p>
          <p class="mt-2 text-xs leading-6 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.protocolHint') }}</p>
          <p v-if="stream && endpoint !== 'models'" class="mt-2 text-xs leading-6 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.streamingHint') }}</p>
          <h3 class="mt-8 text-base font-semibold">{{ t('apiDocs.parameters') }}</h3>
          <p v-if="endpoint === 'models'" class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.modelsHint') }}</p>
          <div v-else class="mt-4 overflow-hidden rounded-xl border border-[#eadfd8] dark:border-dark-800">
            <table class="w-full text-left text-sm">
              <thead class="bg-[#f8f3ef] text-xs text-[#8a7b72] dark:bg-dark-900"><tr><th class="px-4 py-3 font-medium">{{ t('apiDocs.field') }}</th><th class="px-4 py-3 font-medium">{{ t('apiDocs.meaning') }}</th></tr></thead>
              <tbody><tr v-for="field in fields" :key="field.name" class="border-t border-[#eadfd8] dark:border-dark-800"><td class="px-4 py-3 align-top font-mono text-xs">{{ field.name }}</td><td class="px-4 py-3 leading-6 text-[#766c66] dark:text-dark-300">{{ t(field.description) }}</td></tr></tbody>
            </table>
          </div>
        </section>

        <section id="errors" class="scroll-mt-28">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.errors') }}</h2>
          <div class="mt-5 divide-y divide-[#eadfd8] rounded-2xl border border-[#eadfd8] bg-white px-5 dark:divide-dark-800 dark:border-dark-800 dark:bg-dark-900">
            <div v-for="status in [400, 401, 403, 429, 503]" :key="status" class="flex items-start gap-5 py-4"><code class="rounded-md bg-[#f8f3ef] px-2 py-1 text-xs text-[#a64d22] dark:bg-dark-800 dark:text-[#ffb28d]">{{ status }}</code><p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(`apiDocs.error${status}`) }}</p></div>
          </div>
          <p class="mt-4 text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.troubleshoot') }}</p>
        </section>

        <section class="rounded-2xl bg-[#f8f3ef] p-7 sm:p-9 dark:bg-dark-900">
          <h2 class="text-xl font-semibold">{{ t('apiDocs.nextTitle') }}</h2>
          <p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.nextDescription') }}</p>
          <router-link to="/keys" class="mt-5 inline-flex h-10 items-center gap-2 rounded-full bg-[#fb6415] px-5 text-sm font-medium text-black hover:bg-[#ff7b33]">{{ t('apiDocs.createKey') }}<Icon name="arrowRight" size="sm" /></router-link>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { sanitizeUrl } from '@/utils/url'
import { buildDocsExample, docsEndpoints, resolveDocsApiBaseUrl, type DocsEndpoint, type DocsLanguage } from '@/utils/apiDocs'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copied, copyToClipboard } = useClipboard()
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const baseUrl = computed(() => resolveDocsApiBaseUrl(appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl, window.location.origin))
const endpoint = ref<DocsEndpoint>('chat')
const language = ref<DocsLanguage>('curl')
const stream = ref(false)
const isDark = ref(document.documentElement.classList.contains('dark'))
const currentEndpoint = computed(() => docsEndpoints[endpoint.value])
const example = computed(() => buildDocsExample(baseUrl.value, endpoint.value, language.value, stream.value))
const sections = [
  { id: 'quickstart', label: 'apiDocs.quickstart' },
  { id: 'connection', label: 'apiDocs.connection' },
  { id: 'examples', label: 'apiDocs.examples' },
  { id: 'errors', label: 'apiDocs.errors' }
]
const languages: { key: DocsLanguage; label: string }[] = [
  { key: 'curl', label: 'cURL' },
  { key: 'python', label: 'Python' },
  { key: 'javascript', label: 'Node.js' }
]
const fields = computed(() => [
  { name: 'model', description: 'apiDocs.modelField' },
  { name: endpoint.value === 'responses' ? 'input' : 'messages', description: endpoint.value === 'responses' ? 'apiDocs.inputField' : 'apiDocs.messagesField' },
  ...(endpoint.value === 'messages' ? [{ name: 'max_tokens', description: 'apiDocs.maxTokensField' }] : []),
  { name: 'stream', description: 'apiDocs.streamField' }
])
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>
