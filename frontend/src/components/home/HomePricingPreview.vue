<template>
  <section
    v-if="visible"
    id="pricing"
    class="scroll-mt-[72px] border-t border-[#e2e8f0] bg-[#f8fafc] py-12 sm:py-16 dark:border-dark-800 dark:bg-dark-950"
  >
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <div class="max-w-2xl">
        <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#14b8a6]">Pricing</p>
        <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[#0f172a] sm:text-4xl dark:text-white">
          {{ t('homePricing.title') }}
        </h2>
        <p class="mt-4 text-base leading-7 text-[#475569] dark:text-dark-300">
          {{ t('homePricing.description') }}
        </p>
      </div>

      <p
        v-if="error && response"
        role="status"
        class="mt-5 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-300"
      >
        {{ t('homePricing.refreshFailed') }}
      </p>

      <div
        v-if="loading && !response"
        class="mt-8 flex min-h-48 items-center justify-center rounded-2xl border border-[#e2e8f0] bg-white/70 dark:border-dark-800 dark:bg-dark-900/70"
      >
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-[#14b8a6]/25 border-t-[#14b8a6]"></div>
      </div>
      <div
        v-else-if="error && !response"
        role="alert"
        class="mt-8 rounded-2xl border border-red-200 bg-red-50 px-5 py-10 text-center text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300"
      >
        {{ t('modelPlaza.loadFailed') }}
      </div>
      <div
        v-else-if="!lowestGpt"
        class="mt-8 rounded-2xl border border-dashed border-[#cbd5e1] bg-[#f0fdfa] px-5 py-12 text-center dark:border-dark-700 dark:bg-dark-900/50"
      >
        <h3 class="text-base font-semibold text-[#1e293b] dark:text-dark-100">{{ t('homePricing.gptEmptyTitle') }}</h3>
        <p class="mx-auto mt-3 max-w-lg text-sm leading-7 text-[#64748b] dark:text-dark-400">
          {{ canManage ? t('homePricing.emptyAdminDescription') : t('homePricing.emptyDescription') }}
        </p>
        <router-link
          v-if="canManage"
          to="/admin/channels"
          class="mt-5 inline-flex min-h-10 items-center gap-2 rounded-full bg-[#14b8a6] px-5 text-sm font-semibold text-black hover:bg-[#2dd4bf]"
        >
          {{ t('homePricing.configurePricing') }}
          <Icon name="arrowRight" size="sm" />
        </router-link>
      </div>
      <div
        v-else
        data-testid="lowest-gpt-rate"
        class="mt-8 flex flex-col gap-6 rounded-2xl border border-[#e2e8f0] bg-white/80 p-6 sm:flex-row sm:items-center sm:justify-between sm:p-8 dark:border-dark-800 dark:bg-dark-900/70"
      >
        <div class="min-w-0">
          <p class="text-xs font-semibold uppercase tracking-[0.18em] text-[#14b8a6]">
            {{ t('homePricing.lowestRateLabel') }}
          </p>
          <p class="mt-3 truncate text-lg font-semibold text-[#0f172a] dark:text-white">{{ lowestGpt.groupName }}</p>
          <p class="mt-1 text-sm text-[#64748b] dark:text-dark-400">{{ lowestGpt.modelsLabel }}</p>
          <p v-if="lowestGpt.peakNote" class="mt-2 text-xs text-[#94a3b8] dark:text-dark-400">{{ lowestGpt.peakNote }}</p>
        </div>
        <p class="shrink-0 text-4xl font-semibold tracking-tight text-[#0f766e] sm:text-5xl dark:text-[#2dd4bf]">
          {{ lowestGpt.rateLabel }}
        </p>
      </div>

      <div class="mt-5 flex justify-end text-xs text-[#64748b] dark:text-dark-400">
        <router-link v-if="showEntry" to="/model-plaza" class="font-medium text-[#0f766e] hover:text-[#14b8a6]">
          {{ t('nav.modelPlaza') }} <span aria-hidden="true">→</span>
        </router-link>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getModelPlaza, type ModelPlazaGroup, type ModelPlazaResponse, type PlazaModel } from '@/api/modelPlaza'

const props = defineProps<{ visible: boolean; showEntry: boolean; canManage?: boolean }>()
const { t } = useI18n()
const canManage = computed(() => props.canManage === true)
const response = ref<ModelPlazaResponse | null>(null)
const loading = ref(false)
const error = ref(false)
let requestController: AbortController | undefined

function isGptModel(model: PlazaModel) {
  return /gpt/i.test(model.name ?? '')
}

/** 倍率挂在分组上；专属倍率优先于分组倍率。 */
function effectiveRate(group: ModelPlazaGroup) {
  return group.user_rate_multiplier ?? group.rate_multiplier
}

function formatRate(rate: number) {
  return `${Number(rate.toFixed(4))}x`
}

/**
 * 首页只展示一个数字：所有公开分组里 GPT 的最低倍率。
 * 取“包含 GPT 模型且倍率最低”的那个分组，附带该分组的 GPT 模型清单。
 */
const lowestGpt = computed(() => {
  const candidates = (response.value?.groups ?? [])
    .map(group => ({ group, models: group.models.filter(isGptModel) }))
    .filter(entry => entry.models.length > 0)
  if (candidates.length === 0) return null

  const best = candidates.reduce((acc, entry) => (effectiveRate(entry.group) < effectiveRate(acc.group) ? entry : acc))
  const rate = effectiveRate(best.group)
  const shownModels = best.models.slice(0, 3).map(model => model.name).join(' · ')
  return {
    rateLabel: formatRate(rate),
    groupName: best.group.name,
    modelsLabel: best.models.length > 3 ? `${shownModels} …` : shownModels,
    peakNote: best.group.peak_rate_enabled
      ? t('homePricing.peakNote', {
          multiplier: formatRate(best.group.peak_rate_multiplier),
          start: best.group.peak_start,
          end: best.group.peak_end
        })
      : ''
  }
})

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
