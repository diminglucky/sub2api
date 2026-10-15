<template>
  <section
    v-if="visible"
    id="pricing"
    class="scroll-mt-[72px] border-t border-[#eadfd8] bg-[#fffdfb] py-12 sm:py-16 dark:border-dark-800 dark:bg-dark-950"
  >
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div class="flex flex-col justify-between gap-6 lg:flex-row lg:items-end">
        <div class="max-w-2xl">
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#fb6415]">Pricing</p>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[#171311] sm:text-4xl dark:text-white">
            {{ t('homePricing.title') }}
          </h2>
          <p class="mt-4 text-base leading-7 text-[#766c66] dark:text-dark-300">
            {{ t('homePricing.description') }}
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <select
            v-model="selectedPlatform"
            :aria-label="t('modelPlaza.filters.platformLabel')"
            class="h-10 rounded-full border border-[#ded2ca] bg-white px-4 text-sm text-[#3f3631] outline-none transition focus:border-[#fb6415] dark:border-dark-700 dark:bg-dark-900 dark:text-dark-100"
          >
            <option value="all">{{ t('modelPlaza.filters.all') }} {{ t('modelPlaza.filters.platformLabel') }}</option>
            <option v-for="platform in platforms" :key="platform" :value="platform">{{ platformLabel(platform) }}</option>
          </select>
          <div class="relative">
            <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[#9a8e86]" />
            <input
              v-model="searchQuery"
              type="search"
              :aria-label="t('modelPlaza.filters.modelLabel')"
              list="home-pricing-models"
              :placeholder="t('modelPlaza.filters.searchPlaceholder')"
              class="h-10 w-52 rounded-full border border-[#ded2ca] bg-white pl-9 pr-4 text-sm text-[#3f3631] outline-none transition placeholder:text-[#9a8e86] focus:border-[#fb6415] dark:border-dark-700 dark:bg-dark-900 dark:text-dark-100"
            />
            <datalist id="home-pricing-models">
              <option v-for="name in modelNames" :key="name" :value="name" />
            </datalist>
          </div>
          <button
            type="button"
            class="inline-flex h-10 items-center gap-2 rounded-full border border-[#ded2ca] bg-white px-4 text-sm font-medium text-[#3f3631] transition hover:border-[#fb6415] hover:text-[#c34c0d] disabled:cursor-wait disabled:opacity-60 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-100"
            :disabled="loading"
            @click="load"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            {{ t('common.refresh') }}
          </button>
        </div>
      </div>

      <p v-if="error && response" role="status" class="mt-5 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-300">{{ t('homePricing.refreshFailed') }}</p>
      <div v-if="loading && !response" class="mt-8 flex min-h-48 items-center justify-center rounded-2xl border border-[#eadfd8] bg-white/70 dark:border-dark-800 dark:bg-dark-900/70">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-[#fb6415]/25 border-t-[#fb6415]"></div>
      </div>
      <div v-else-if="error && !response" role="alert" class="mt-8 rounded-2xl border border-red-200 bg-red-50 px-5 py-10 text-center text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300">
        {{ t('modelPlaza.loadFailed') }}
      </div>
      <div v-else-if="filteredGroups.length === 0" class="mt-8 rounded-2xl border border-dashed border-[#d9ccc4] bg-[#f8f3ef] px-5 py-12 text-center dark:border-dark-700 dark:bg-dark-900/50">
        <h3 class="text-base font-semibold text-[#3f3631] dark:text-dark-100">{{ hasFilters ? t('modelPlaza.noSearchResult') : t('homePricing.emptyTitle') }}</h3>
        <p v-if="!hasFilters" class="mx-auto mt-3 max-w-lg text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ canManage ? t('homePricing.emptyAdminDescription') : t('homePricing.emptyDescription') }}</p>
        <button v-if="hasFilters" type="button" class="mt-5 inline-flex min-h-10 items-center rounded-full border border-[#ded2ca] px-5 text-sm font-medium text-[#c34c0d] hover:border-[#fb6415] dark:border-dark-700 dark:text-orange-300" @click="clearFilters">{{ t('homePricing.clearFilters') }}</button>
        <router-link v-else-if="canManage" to="/admin/channels" class="mt-5 inline-flex min-h-10 items-center gap-2 rounded-full bg-[#fb6415] px-5 text-sm font-semibold text-black hover:bg-[#ff7b33]">{{ t('homePricing.configurePricing') }}<Icon name="arrowRight" size="sm" /></router-link>
      </div>
      <div v-else class="mt-8 space-y-4">
        <div class="flex items-center justify-between gap-3 text-xs text-[#8a7b72] dark:text-dark-400">
          <span>{{ t('homePricing.previewCount', { shown: shownCount, total: matchedCount }) }}</span>
          <span v-if="loadedAt">{{ t('homePricing.fetchedAt', { time: loadedAtLabel }) }}</span>
        </div>
        <PlazaGroupSection v-for="group in filteredGroups" :key="group.id" :group="group" collapsible :preview-model-count="8" />
      </div>
      <div class="mt-5 flex justify-end text-xs text-[#8a7b72] dark:text-dark-400">
        <router-link v-if="showEntry" to="/model-plaza" class="font-medium text-[#c34c0d] hover:text-[#fb6415]">{{ t('nav.modelPlaza') }} <span aria-hidden="true">→</span></router-link>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlazaGroupSection from '@/components/modelPlaza/PlazaGroupSection.vue'
import { getModelPlaza, type ModelPlazaResponse } from '@/api/modelPlaza'
import { platformLabel } from '@/utils/platformColors'

const props = defineProps<{ visible: boolean; showEntry: boolean; canManage?: boolean }>()
const { t, locale } = useI18n()
const canManage = computed(() => props.canManage === true)
const response = ref<ModelPlazaResponse | null>(null)
const loading = ref(false)
const error = ref(false)
const selectedPlatform = ref('all')
const searchQuery = ref('')
const hasFilters = computed(() => searchQuery.value.trim().length > 0 || selectedPlatform.value !== 'all')
const loadedAt = ref<Date | null>(null)
let requestController: AbortController | undefined

const platforms = computed(() => [...new Set((response.value?.groups ?? []).map(g => g.platform).filter(Boolean))].sort())
const modelNames = computed(() => [...new Set((response.value?.groups ?? [])
  .filter(group => selectedPlatform.value === 'all' || group.platform === selectedPlatform.value)
  .flatMap(group => group.models.map(model => model.name)))].sort())
const matchedGroups = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  return (response.value?.groups ?? [])
    .filter(group => selectedPlatform.value === 'all' || group.platform === selectedPlatform.value)
    .map(group => ({ ...group, models: group.models.filter(model => !query || model.name.toLowerCase().includes(query)) }))
    .filter(group => group.models.length > 0)
})
const matchedCount = computed(() => matchedGroups.value.reduce((count, group) => count + group.models.length, 0))
const filteredGroups = computed(() => {
  let remaining = 24
  return matchedGroups.value
    .map(group => {
      const models = group.models.slice(0, remaining)
      remaining -= models.length
      return { ...group, models }
    })
    .filter(group => group.models.length > 0)
})
const shownCount = computed(() => filteredGroups.value.reduce((count, group) => count + group.models.length, 0))
const loadedAtLabel = computed(() => loadedAt.value?.toLocaleTimeString(locale.value.startsWith('zh') ? 'zh-CN' : 'en-US', { hour: '2-digit', minute: '2-digit' }) ?? '')
function clearFilters() {
  searchQuery.value = ''
  selectedPlatform.value = 'all'
}
async function load() {
  if (loading.value) return
  loading.value = true
  error.value = false
  requestController = new AbortController()
  const controller = requestController
  try {
    const result = await getModelPlaza({ signal: controller.signal })
    if (controller.signal.aborted) return
    response.value = result
    loadedAt.value = new Date()
    if (selectedPlatform.value !== 'all' && !platforms.value.includes(selectedPlatform.value)) selectedPlatform.value = 'all'
  } catch {
    if (!controller.signal.aborted) error.value = true
  } finally {
    loading.value = false
  }
}
watch(() => props.visible, visible => { if (visible && !response.value) load() }, { immediate: true })
let refreshTimer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  refreshTimer = setInterval(() => {
    if (props.visible && !loading.value && document.visibilityState === 'visible') load()
  }, 5 * 60 * 1000)
})
onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
  requestController?.abort()
})
</script>
