<template>
  <div class="min-h-screen bg-gradient-to-br from-gray-50 via-primary-50/30 to-gray-100 dark:from-dark-950 dark:via-dark-900 dark:to-dark-950">
    <div class="pointer-events-none fixed inset-0 bg-[linear-gradient(rgba(20,184,166,0.04)_1px,transparent_1px),linear-gradient(90deg,rgba(20,184,166,0.04)_1px,transparent_1px)] bg-[size:64px_64px]" />

    <header class="relative z-10 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <RouterLink to="/home" class="flex items-center gap-3">
          <div class="h-10 w-10 overflow-hidden rounded-xl shadow-md">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="hidden text-sm font-semibold text-gray-900 dark:text-white sm:inline">{{ siteName }}</span>
        </RouterLink>
        <div class="flex items-center gap-3">
          <LocaleSwitcher />
          <RouterLink
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="inline-flex items-center rounded-full bg-gray-900 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </RouterLink>
        </div>
      </nav>
    </header>

    <main class="relative z-10 px-6 pb-16 pt-10">
      <div class="mx-auto max-w-6xl">
        <div class="mb-8 flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p class="text-sm font-semibold uppercase tracking-[0.24em] text-primary-600 dark:text-primary-300">
              {{ t('home.publicModels.eyebrow') }}
            </p>
            <h1 class="mt-3 text-4xl font-bold text-gray-900 dark:text-white md:text-5xl">
              {{ t('home.publicModels.title') }}
            </h1>
          </div>
          <div class="relative w-full lg:max-w-sm">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('models.searchPlaceholder')"
              class="input pl-10"
            />
          </div>
        </div>

        <div v-if="loading" class="rounded-xl border border-gray-200 bg-white/70 py-16 text-center dark:border-dark-700 dark:bg-dark-900/70">
          <Icon name="refresh" size="lg" class="inline-block animate-spin text-gray-400" />
        </div>

        <div v-else-if="filteredRows.length === 0" class="rounded-xl border border-gray-200 bg-white/70 py-16 text-center dark:border-dark-700 dark:bg-dark-900/70">
          <Icon name="inbox" size="xl" class="mx-auto mb-3 h-12 w-12 text-gray-400" />
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('models.empty') }}</p>
        </div>

        <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <button
            v-for="row in filteredRows"
            :key="modelKey(row)"
            type="button"
            class="group flex min-h-[190px] flex-col rounded-xl border border-gray-200 bg-white/75 p-5 text-left shadow-sm backdrop-blur-sm transition-all hover:-translate-y-0.5 hover:border-primary-300 hover:shadow-lg hover:shadow-primary-500/10 dark:border-dark-700 dark:bg-dark-900/75 dark:hover:border-primary-500/60"
            @click="selectModel(row)"
          >
            <div class="flex items-start gap-3">
              <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-lg border border-gray-200 text-gray-500 dark:border-dark-600 dark:text-gray-300">
                <PlatformIcon :platform="row.platform as GroupPlatform" size="md" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="text-xs font-semibold uppercase tracking-[0.2em] text-gray-500 dark:text-gray-400">{{ row.platform }}</div>
                <h2 class="mt-1 break-words text-lg font-semibold leading-snug text-gray-900 dark:text-white">{{ row.name }}</h2>
              </div>
            </div>

            <div class="mt-5 grid grid-cols-2 gap-3 text-sm">
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ primaryPriceLabel(row) }}</div>
                <div class="mt-1 whitespace-nowrap font-semibold text-gray-900 dark:text-white">{{ primaryPriceText(row) }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ secondaryPriceLabel(row) }}</div>
                <div class="mt-1 whitespace-nowrap font-semibold text-gray-900 dark:text-white">{{ secondaryPriceText(row) }}</div>
              </div>
            </div>

            <div class="mt-auto pt-5">
              <span class="inline-flex items-center rounded-full bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700 dark:bg-primary-500/10 dark:text-primary-300">
                {{ row.groups.length }} {{ t('home.publicModels.groups') }}
              </span>
            </div>
          </button>
        </div>
      </div>
    </main>

    <Teleport to="body">
      <div
        v-if="isDrawerOpen && selectedRow"
        class="fixed inset-0 z-50 flex justify-end bg-gray-950/45 backdrop-blur-[2px] dark:bg-black/60"
        @click.self="closeDrawer"
      >
        <aside
          class="flex h-full w-full max-w-[34rem] flex-col overflow-hidden border-l border-gray-200 bg-white shadow-2xl dark:border-dark-700 dark:bg-dark-900"
          role="dialog"
          aria-modal="true"
          :aria-label="selectedRow.name"
        >
          <header class="flex items-start justify-between gap-4 border-b border-gray-200 px-6 py-5 dark:border-dark-700">
            <div class="flex min-w-0 gap-3">
              <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-lg border border-gray-200 text-gray-500 dark:border-dark-600 dark:text-gray-300">
                <PlatformIcon :platform="selectedRow.platform as GroupPlatform" size="md" />
              </div>
              <div class="min-w-0">
                <div class="text-xs font-semibold uppercase tracking-[0.22em] text-primary-600 dark:text-primary-300">{{ selectedRow.platform }}</div>
                <h2 class="mt-1 break-words text-xl font-bold leading-tight text-gray-900 dark:text-white">{{ selectedRow.name }}</h2>
              </div>
            </div>
            <button
              type="button"
              class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-800 dark:hover:text-white"
              :aria-label="t('common.close')"
              @click="closeDrawer"
            >
              <Icon name="x" size="md" />
            </button>
          </header>

          <div class="flex-1 overflow-y-auto px-6 py-5">
            <div class="mb-5 flex flex-wrap gap-2">
              <span class="rounded-full border border-gray-300 px-3 py-1 text-xs font-semibold text-gray-700 dark:border-dark-500 dark:text-gray-200">
                {{ billingLabel(selectedRow.model.pricing?.billing_mode) }}
              </span>
              <span class="rounded-full bg-gray-100 px-3 py-1 text-xs font-semibold text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                {{ t('home.publicModels.publicOnly') }}
              </span>
            </div>

            <div class="mb-4">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('models.groupPrices') }}</h3>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('home.publicModels.priceHint') }}</p>
            </div>

            <div class="grid gap-3">
              <div
                v-for="item in selectedGroupPrices"
                :key="item.group.id"
                class="rounded-lg border border-gray-200 p-4 dark:border-dark-700"
              >
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div class="min-w-0">
                    <div class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ item.group.name }}</div>
                    <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ item.group.subscription_type === 'subscription' ? t('models.subscriptionGroup') : t('models.standardGroup') }}</div>
                  </div>
                  <span class="rounded-full bg-primary-50 px-2.5 py-1 text-xs font-semibold text-primary-700 dark:bg-primary-500/10 dark:text-primary-300">
                    {{ formatMultiplier(item.multiplier) }}
                  </span>
                </div>
                <div class="mt-4 grid grid-cols-2 gap-3 text-sm">
                  <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ item.secondaryLabel }}</div>
                    <div class="mt-1 whitespace-nowrap font-semibold text-gray-900 dark:text-white">{{ item.secondaryPrice }}</div>
                  </div>
                  <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ item.tertiaryLabel }}</div>
                    <div class="mt-1 whitespace-nowrap font-semibold text-gray-900 dark:text-white">{{ item.tertiaryPrice }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import userChannelsAPI, { type UserAvailableGroup, type UserSupportedModel } from '@/api/channels'
import { BILLING_MODE_IMAGE, BILLING_MODE_PER_REQUEST, BILLING_MODE_TOKEN, type BillingMode } from '@/constants/channel'
import type { GroupPlatform } from '@/types'
import { extractApiErrorMessage } from '@/utils/apiError'

interface ModelRow {
  name: string
  platform: string
  model: UserSupportedModel
  channels: string[]
  groups: UserAvailableGroup[]
}

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const rows = ref<ModelRow[]>([])
const loading = ref(false)
const searchQuery = ref('')
const selectedModelKey = ref('')
const isDrawerOpen = ref(false)

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'SuperAI')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')

const filteredRows = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return rows.value
  return rows.value.filter((row) =>
    row.name.toLowerCase().includes(q) ||
    row.platform.toLowerCase().includes(q) ||
    row.channels.some((channel) => channel.toLowerCase().includes(q)) ||
    row.groups.some((group) => group.name.toLowerCase().includes(q))
  )
})

const selectedRow = computed(() =>
  filteredRows.value.find((row) => modelKey(row) === selectedModelKey.value) || null
)

const selectedGroupPrices = computed(() => {
  const row = selectedRow.value
  if (!row?.model.pricing) return []
  return row.groups
    .map((group) => {
      const multiplier = Number.isFinite(group.rate_multiplier) ? group.rate_multiplier : 1
      return {
        group,
        multiplier,
        ...groupPriceColumns(row, multiplier)
      }
    })
    .sort((a, b) => a.multiplier - b.multiplier || a.group.name.localeCompare(b.group.name))
})

async function loadModels() {
  loading.value = true
  try {
    const channels = await userChannelsAPI.getPublicAvailableModels()
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
      a.platform.localeCompare(b.platform) || a.name.localeCompare(b.name)
    )
  } catch (err) {
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

function modelKey(row: ModelRow) {
  return `${row.platform}\u0000${row.name}`
}

function selectModel(row: ModelRow) {
  selectedModelKey.value = modelKey(row)
  isDrawerOpen.value = true
}

function closeDrawer() {
  isDrawerOpen.value = false
}

function groupPriceColumns(row: ModelRow, multiplier: number) {
  const pricing = row.model.pricing
  if (!pricing) {
    return {
      secondaryLabel: t('models.input'),
      secondaryPrice: '-',
      tertiaryLabel: t('models.output'),
      tertiaryPrice: '-'
    }
  }
  if (pricing.billing_mode === BILLING_MODE_PER_REQUEST) {
    return {
      secondaryLabel: t('models.requestPrice'),
      secondaryPrice: scaledPriceText(pricing.per_request_price, multiplier, t('availableChannels.pricing.unitPerRequest')),
      tertiaryLabel: t('models.billingType'),
      tertiaryPrice: billingLabel(pricing.billing_mode)
    }
  }
  if (pricing.billing_mode === BILLING_MODE_IMAGE) {
    return {
      secondaryLabel: t('models.imagePrice'),
      secondaryPrice: scaledPriceText(imageRequestPrice(pricing), multiplier, t('availableChannels.pricing.unitPerRequest')),
      tertiaryLabel: t('models.billingType'),
      tertiaryPrice: billingLabel(pricing.billing_mode)
    }
  }
  return {
    secondaryLabel: t('models.input'),
    secondaryPrice: scaledPriceText(pricing.input_price, multiplier, t('availableChannels.pricing.unitPerMillion')),
    tertiaryLabel: t('models.output'),
    tertiaryPrice: scaledPriceText(pricing.output_price, multiplier, t('availableChannels.pricing.unitPerMillion'))
  }
}

function billingLabel(mode?: BillingMode) {
  if (mode === BILLING_MODE_PER_REQUEST) return t('availableChannels.pricing.billingModePerRequest')
  if (mode === BILLING_MODE_IMAGE) return t('availableChannels.pricing.billingModeImage')
  if (mode === BILLING_MODE_TOKEN) return t('availableChannels.pricing.billingModeToken')
  return t('availableChannels.noPricing')
}

function primaryPriceLabel(row: ModelRow) {
  const mode = row.model.pricing?.billing_mode
  if (mode === BILLING_MODE_PER_REQUEST) return t('models.requestPrice')
  if (mode === BILLING_MODE_IMAGE) return t('models.imagePrice')
  return t('models.input')
}

function secondaryPriceLabel(row: ModelRow) {
  const mode = row.model.pricing?.billing_mode
  if (mode === BILLING_MODE_PER_REQUEST || mode === BILLING_MODE_IMAGE) return t('models.billingType')
  return t('models.output')
}

function primaryPriceText(row: ModelRow) {
  const pricing = row.model.pricing
  if (!pricing) return '-'
  if (pricing.billing_mode === BILLING_MODE_PER_REQUEST) return priceText(pricing.per_request_price, t('availableChannels.pricing.unitPerRequest'))
  if (pricing.billing_mode === BILLING_MODE_IMAGE) return priceText(imageRequestPrice(pricing), t('availableChannels.pricing.unitPerRequest'))
  return priceText(pricing.input_price, t('availableChannels.pricing.unitPerMillion'))
}

function secondaryPriceText(row: ModelRow) {
  const pricing = row.model.pricing
  if (!pricing) return '-'
  if (pricing.billing_mode === BILLING_MODE_PER_REQUEST || pricing.billing_mode === BILLING_MODE_IMAGE) {
    return billingLabel(pricing.billing_mode)
  }
  return priceText(pricing.output_price, t('availableChannels.pricing.unitPerMillion'))
}

function imageRequestPrice(pricing: { per_request_price?: number | null; image_output_price?: number | null }) {
  return pricing.per_request_price ?? pricing.image_output_price
}

function priceText(value: number | null | undefined, unit: string) {
  if (value == null) return '-'
  const scale = unit === t('availableChannels.pricing.unitPerMillion') ? 1_000_000 : 1
  return `¥${(value * scale).toFixed(2)} ${unit}`
}

function scaledPriceText(value: number | null | undefined, multiplier: number, unit: string) {
  if (value == null) return '-'
  const scale = unit === t('availableChannels.pricing.unitPerMillion') ? 1_000_000 : 1
  return `¥${(value * multiplier * scale).toFixed(2)} ${unit}`
}

function formatMultiplier(value: number) {
  return `x${Number(value.toPrecision(10))}`
}

watch(filteredRows, (items) => {
  if (items.length === 0 || !items.some((row) => modelKey(row) === selectedModelKey.value)) {
    selectedModelKey.value = ''
    isDrawerOpen.value = false
  }
})

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  loadModels()
})
</script>
