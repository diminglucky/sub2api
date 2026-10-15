<template>
  <div class="min-h-screen bg-[#f1f5f9] dark:bg-dark-950 lg:grid lg:grid-cols-2">
    <aside class="relative hidden overflow-hidden border-r border-slate-200 bg-[#f8fafc] lg:flex lg:flex-col lg:justify-between lg:p-12 dark:border-dark-800 dark:bg-dark-900">
      <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_18%_18%,rgba(20,184,166,0.20),transparent_36%),radial-gradient(circle_at_82%_78%,rgba(45,212,191,0.16),transparent_42%)]"></div>
      <div class="pointer-events-none absolute inset-0 bg-[linear-gradient(rgba(20,184,166,0.07)_1px,transparent_1px),linear-gradient(90deg,rgba(20,184,166,0.07)_1px,transparent_1px)] bg-[size:56px_56px]"></div>

      <div class="relative flex items-center gap-3">
        <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-11 w-11 rounded-xl object-contain" />
        <div>
          <p class="text-base font-semibold text-slate-900 dark:text-white">{{ siteName }}</p>
          <p class="text-xs text-slate-500 dark:text-dark-400">{{ t('auth.splitBrandSubtitle') }}</p>
        </div>
      </div>

      <div class="relative max-w-xl">
        <p class="text-sm font-semibold text-primary-600 dark:text-primary-400">{{ t('auth.splitEyebrow') }}</p>
        <h1 class="mt-4 text-5xl font-bold leading-[0.95] tracking-tight text-slate-900 sm:text-6xl dark:text-white">
          {{ t('auth.splitTitle') }}
        </h1>
        <p class="mt-6 max-w-lg text-base leading-8 text-slate-600 dark:text-dark-300">{{ t('auth.splitSubtitle') }}</p>
        <div class="mt-10 grid max-w-lg grid-cols-3 gap-3">
          <div v-for="feature in features" :key="feature.title" class="rounded-xl border border-slate-200 bg-white/80 p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900/70">
            <p class="text-sm font-semibold text-slate-900 dark:text-white">{{ t(feature.title) }}</p>
            <p class="mt-1 text-xs text-slate-500 dark:text-dark-400">{{ t(feature.desc) }}</p>
          </div>
        </div>
      </div>

      <div class="relative text-xs text-slate-400 dark:text-dark-500">&copy; {{ currentYear }} {{ siteName }}</div>
    </aside>

    <main class="flex min-h-screen items-center justify-center px-5 py-10 sm:px-8">
      <div :class="contentClass">
        <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-[0_24px_70px_-40px_rgba(15,23,42,0.35)] sm:p-8 dark:border-dark-800 dark:bg-dark-900">
          <slot />
        </div>
        <div class="mt-4 text-center text-sm text-slate-500 dark:text-dark-400">
          <slot name="footer" />
        </div>
        <div v-if="showCopyright" class="mt-6 text-center text-xs text-slate-400 dark:text-dark-500">
          &copy; {{ currentYear }} {{ siteName }}
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import { normalizeBrandName } from '@/utils/brand'

const props = withDefaults(defineProps<{
  showCopyright?: boolean
  compact?: boolean
}>(), {
  showCopyright: true,
  compact: false
})

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => normalizeBrandName(appStore.siteName))
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const currentYear = computed(() => new Date().getFullYear())
const contentClass = computed(() => (props.compact ? 'relative z-10 w-full max-w-[420px]' : 'relative z-10 w-full max-w-[500px]'))
const features = [
  { title: 'auth.splitFeatureMultiTitle', desc: 'auth.splitFeatureMultiDesc' },
  { title: 'auth.splitFeatureApiTitle', desc: 'auth.splitFeatureApiDesc' },
  { title: 'auth.splitFeatureQuotaTitle', desc: 'auth.splitFeatureQuotaDesc' }
]

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>
