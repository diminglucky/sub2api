<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="relative w-full lg:max-w-xl">
          <Icon
            name="search"
            size="md"
            class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
          />
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('models.searchPlaceholder')"
            class="input pl-10"
          />
        </div>

        <button
          @click="loadModels"
          :disabled="loading"
          class="btn btn-secondary"
          :title="t('common.refresh', 'Refresh')"
        >
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div class="grid gap-5 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside class="space-y-5">
          <section class="space-y-2">
            <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400">{{ t('models.filters.billing') }}</h2>
            <div class="space-y-1">
              <button
                v-for="option in billingFilterOptions"
                :key="option.value"
                type="button"
                class="w-full rounded-md px-3 py-2 text-left text-sm transition-colors"
                :class="billingFilter === option.value ? 'bg-primary-600 text-white' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-800'"
                @click="billingFilter = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </section>

          <section class="space-y-2">
            <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400">{{ t('models.filters.platform') }}</h2>
            <div class="space-y-1">
              <button
                v-for="option in platformFilterOptions"
                :key="option.value"
                type="button"
                class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors"
                :class="platformFilter === option.value ? 'bg-primary-600 text-white' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-800'"
                @click="platformFilter = option.value"
              >
                <PlatformIcon
                  v-if="option.value !== 'all'"
                  :platform="option.value as GroupPlatform"
                  size="xs"
                />
                <span>{{ option.label }}</span>
              </button>
            </div>
          </section>
        </aside>

        <main class="min-w-0 space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-3 rounded-lg bg-gray-50 px-4 py-3 text-sm dark:bg-dark-800/70">
            <div class="font-semibold text-gray-900 dark:text-white">
              {{ t('models.resultCount', { count: filteredRows.length }) }}
            </div>
          </div>

          <div v-if="loading" class="py-20 text-center">
            <Icon name="refresh" size="lg" class="inline-block animate-spin text-gray-400" />
          </div>

          <div v-else-if="filteredRows.length === 0" class="card py-16 text-center">
            <Icon name="inbox" size="xl" class="mx-auto mb-3 h-12 w-12 text-gray-400" />
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('models.empty') }}</p>
          </div>

          <div v-else class="grid gap-4 xl:grid-cols-2">
            <article
              v-for="row in filteredRows"
              :key="`${row.platform}-${row.name}`"
              class="card flex min-h-[224px] flex-col border border-gray-200 p-6 transition-colors hover:border-primary-400 dark:border-dark-700 dark:hover:border-primary-500"
            >
              <div class="flex items-start gap-4">
                <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg border border-gray-200 text-gray-500 dark:border-dark-600 dark:text-gray-300">
                  <PlatformIcon :platform="row.platform as GroupPlatform" size="md" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="text-xs font-semibold uppercase tracking-[0.22em] text-gray-500 dark:text-gray-400">
                    {{ row.platform }}
                  </div>
                  <h2 class="mt-1 break-words text-lg font-semibold leading-snug text-gray-900 dark:text-white">
                    {{ row.name }}
                  </h2>
                </div>
              </div>

              <div class="mt-4 flex flex-wrap gap-2">
                <span
                  v-for="tag in modelTags(row)"
                  :key="`${row.platform}-${row.name}-${tag}`"
                  class="rounded-full bg-gray-100 px-2.5 py-1 text-xs font-semibold uppercase text-gray-600 dark:bg-dark-700 dark:text-gray-300"
                >
                  {{ tag }}
                </span>
              </div>

              <div class="mt-5 border-t border-gray-200 pt-4 dark:border-dark-700">
                <div class="grid grid-cols-[minmax(7rem,0.9fr)_minmax(8rem,1fr)_minmax(8rem,1fr)] gap-4 text-sm">
                  <div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('models.billingType') }}</div>
                    <div class="mt-1">
                      <span class="rounded-full border border-gray-300 px-2 py-0.5 text-xs font-semibold text-gray-700 dark:border-dark-500 dark:text-gray-200">
                        {{ billingLabel(row.model.pricing?.billing_mode) }}
                      </span>
                    </div>
                  </div>
                  <div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('models.input') }}</div>
                    <div class="mt-1 whitespace-nowrap font-semibold text-gray-900 dark:text-white">{{ inputPriceText(row) }}</div>
                  </div>
                  <div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('models.output') }}</div>
                    <div class="mt-1 whitespace-nowrap font-semibold text-gray-900 dark:text-white">{{ outputPriceText(row) }}</div>
                  </div>
                </div>
              </div>

            </article>
          </div>
        </main>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import userChannelsAPI, { type UserAvailableGroup, type UserSupportedModel } from '@/api/channels'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { BILLING_MODE_IMAGE, BILLING_MODE_PER_REQUEST, BILLING_MODE_TOKEN, type BillingMode } from '@/constants/channel'
import type { GroupPlatform } from '@/types'

interface ModelRow {
  name: string
  platform: string
  model: UserSupportedModel
  channels: string[]
  groups: UserAvailableGroup[]
}

const { t } = useI18n()
const appStore = useAppStore()

const rows = ref<ModelRow[]>([])
const loading = ref(false)
const searchQuery = ref('')
const billingFilter = ref<'all' | BillingMode>('all')
const platformFilter = ref('all')

const billingFilterOptions = computed(() => [
  { value: 'all' as const, label: t('models.filters.allBilling') },
  { value: BILLING_MODE_TOKEN, label: t('availableChannels.pricing.billingModeToken') },
  { value: BILLING_MODE_PER_REQUEST, label: t('availableChannels.pricing.billingModePerRequest') },
  { value: BILLING_MODE_IMAGE, label: t('availableChannels.pricing.billingModeImage') },
])

const platformFilterOptions = computed(() => [
  { value: 'all', label: t('models.filters.allPlatforms') },
  ...Array.from(new Set(rows.value.map((row) => row.platform)))
    .sort()
    .map((platform) => ({ value: platform, label: platform })),
])

const filteredRows = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return rows.value.filter((row) => {
    if (billingFilter.value !== 'all' && row.model.pricing?.billing_mode !== billingFilter.value) return false
    if (platformFilter.value !== 'all' && row.platform !== platformFilter.value) return false
    if (!q) return true
    return row.name.toLowerCase().includes(q) ||
      row.platform.toLowerCase().includes(q) ||
      row.channels.some((channel) => channel.toLowerCase().includes(q)) ||
      row.groups.some((group) => group.name.toLowerCase().includes(q))
  })
})

async function loadModels() {
  loading.value = true
  try {
    const channels = await userChannelsAPI.getAvailableModels()

    const byModel = new Map<string, ModelRow>()
    for (const channel of channels) {
      for (const section of channel.platforms) {
        for (const model of section.supported_models) {
          const platform = model.platform || section.platform
          const key = `${platform}\u0000${model.name}`
          const existing = byModel.get(key)
          if (existing) {
            if (!existing.channels.includes(channel.name)) existing.channels.push(channel.name)
            mergeGroups(existing.groups, section.groups)
            continue
          }

          byModel.set(key, {
            name: model.name,
            platform,
            model: { ...model, platform },
            channels: [channel.name],
            groups: [...section.groups],
          })
        }
      }
    }

    rows.value = Array.from(byModel.values()).sort((a, b) =>
      a.platform.localeCompare(b.platform) || a.name.localeCompare(b.name),
    )
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

function mergeGroups(target: UserAvailableGroup[], incoming: UserAvailableGroup[]) {
  const seen = new Set(target.map((group) => group.id))
  for (const group of incoming) {
    if (seen.has(group.id)) continue
    seen.add(group.id)
    target.push(group)
  }
}

function billingLabel(mode?: BillingMode) {
  if (mode === BILLING_MODE_PER_REQUEST) return t('availableChannels.pricing.billingModePerRequest')
  if (mode === BILLING_MODE_IMAGE) return t('availableChannels.pricing.billingModeImage')
  if (mode === BILLING_MODE_TOKEN) return t('availableChannels.pricing.billingModeToken')
  return t('availableChannels.noPricing')
}

function inputPriceText(row: ModelRow) {
  const pricing = row.model.pricing
  if (!pricing) return '-'
  if (pricing.billing_mode === BILLING_MODE_PER_REQUEST) return priceText(pricing.per_request_price, t('availableChannels.pricing.unitPerRequest'))
  if (pricing.billing_mode === BILLING_MODE_IMAGE) return priceText(pricing.image_output_price, t('availableChannels.pricing.unitPerRequest'))
  return priceText(pricing.input_price, t('availableChannels.pricing.unitPerMillion'))
}

function outputPriceText(row: ModelRow) {
  const pricing = row.model.pricing
  if (!pricing) return '-'
  if (pricing.billing_mode === BILLING_MODE_PER_REQUEST) return priceText(pricing.per_request_price, t('availableChannels.pricing.unitPerRequest'))
  if (pricing.billing_mode === BILLING_MODE_IMAGE) return priceText(pricing.image_output_price, t('availableChannels.pricing.unitPerRequest'))
  return priceText(pricing.output_price, t('availableChannels.pricing.unitPerMillion'))
}

function priceText(value: number | null | undefined, unit: string) {
  if (value == null) return '-'
  const scale = unit === t('availableChannels.pricing.unitPerMillion') ? 1_000_000 : 1
  return `$${(value * scale).toFixed(2)} ${unit}`
}

function modelTags(row: ModelRow) {
  const name = row.name.toLowerCase()
  const tags = ['text']
  if (name.includes('image') || name.includes('vision')) tags.push('vision')
  if (name.includes('reason') || name.includes('thinking') || name.includes('high')) tags.push('reasoning')
  if (name.includes('tool') || name.includes('codex')) tags.push('tools')
  return Array.from(new Set(tags))
}

onMounted(loadModels)
</script>
