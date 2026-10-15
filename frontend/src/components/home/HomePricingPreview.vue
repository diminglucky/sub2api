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
        v-else-if="!lowestGroup"
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
      <div v-else data-testid="lowest-gpt-group" class="mt-8">
        <PlazaGroupSection :group="lowestGroup" collapsible :preview-model-count="8" />
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
import PlazaGroupSection from '@/components/modelPlaza/PlazaGroupSection.vue'
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

/**
 * 首页只展示一个分组的价格表：所有公开分组里包含 GPT 且倍率最低的那个。
 * 分组内的模型与价格沿用模型广场的表格，不再把每个分组都铺出来。
 */
const lowestGroup = computed(() => {
  const candidates = (response.value?.groups ?? []).filter(group => group.models.some(isGptModel))
  if (candidates.length === 0) return null
  return candidates.reduce((acc, group) => (effectiveRate(group) < effectiveRate(acc) ? group : acc))
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
