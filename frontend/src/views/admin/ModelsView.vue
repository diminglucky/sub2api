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
            :placeholder="t('admin.models.searchPlaceholder')"
            class="input pl-10"
          />
        </div>

        <div class="flex flex-wrap gap-2">
          <router-link to="/admin/channels/pricing" class="btn btn-primary">
            <Icon name="plus" size="md" />
            <span>{{ t('admin.models.managePricing') }}</span>
          </router-link>
          <button
            @click="loadModels"
            :disabled="loading"
            class="btn btn-secondary"
            :title="t('common.refresh', 'Refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <div class="grid gap-5 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside class="space-y-5">
          <section class="space-y-2">
            <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400">{{ t('admin.models.filters.status') }}</h2>
            <div class="space-y-1">
              <button
                v-for="option in statusFilterOptions"
                :key="option.value"
                type="button"
                class="w-full rounded-md px-3 py-2 text-left text-sm transition-colors"
                :class="statusFilter === option.value ? 'bg-primary-600 text-white' : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-800'"
                @click="statusFilter = option.value"
              >
                {{ option.label }}
              </button>
            </div>
          </section>

          <section class="space-y-2">
            <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400">{{ t('admin.models.filters.platform') }}</h2>
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
          <div class="rounded-lg bg-gray-50 px-4 py-3 text-sm font-semibold text-gray-900 dark:bg-dark-800/70 dark:text-white">
            {{ t('admin.models.resultCount', { count: filteredRows.length }) }}
          </div>

          <div v-if="loading" class="py-20 text-center">
            <Icon name="refresh" size="lg" class="inline-block animate-spin text-gray-400" />
          </div>

          <div v-else-if="filteredRows.length === 0" class="card py-16 text-center">
            <Icon name="inbox" size="xl" class="mx-auto mb-3 h-12 w-12 text-gray-400" />
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.models.empty') }}</p>
          </div>

          <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            <article
              v-for="row in filteredRows"
              :key="`${row.channelId}-${row.platform}-${row.name}`"
              class="card flex min-h-[226px] flex-col border border-gray-200 p-5 dark:border-dark-700"
            >
              <div class="flex items-start gap-4">
                <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg border border-gray-200 text-gray-500 dark:border-dark-600 dark:text-gray-300">
                  <PlatformIcon :platform="row.platform as GroupPlatform" size="md" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.22em] text-gray-500 dark:text-gray-400">
                    {{ row.platform }}
                  </div>
                  <h2 class="mt-1 break-words text-lg font-semibold leading-snug text-gray-900 dark:text-white">
                    {{ row.name }}
                  </h2>
                </div>
                <span
                  class="rounded-full px-2 py-0.5 text-xs font-semibold"
                  :class="row.status === 'active' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'"
                >
                  {{ statusLabel(row.status) }}
                </span>
              </div>

              <div class="mt-5 border-t border-gray-200 pt-4 dark:border-dark-700">
                <div class="grid grid-cols-3 gap-3 text-sm">
                  <div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.models.billingType') }}</div>
                    <div class="mt-1 font-semibold text-gray-900 dark:text-white">{{ billingLabel(row.billingMode) }}</div>
                  </div>
                  <div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.models.input') }}</div>
                    <div class="mt-1 font-semibold text-gray-900 dark:text-white">{{ priceText(row.inputPrice, row) }}</div>
                  </div>
                  <div>
                    <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.models.output') }}</div>
                    <div class="mt-1 font-semibold text-gray-900 dark:text-white">{{ outputPriceText(row) }}</div>
                  </div>
                </div>
              </div>

              <div class="mt-4 flex flex-wrap gap-1.5">
                <span class="inline-flex rounded-md border border-gray-200 px-2 py-0.5 text-xs text-gray-700 dark:border-dark-600 dark:text-gray-300">
                  {{ row.channelName }}
                </span>
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
import adminAPI from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { BILLING_MODE_IMAGE, BILLING_MODE_PER_REQUEST, BILLING_MODE_TOKEN, CHANNEL_STATUS_ACTIVE, CHANNEL_STATUS_DISABLED, type BillingMode, type ChannelStatus } from '@/constants/channel'
import type { GroupPlatform } from '@/types'

interface AdminModelRow {
  name: string
  platform: string
  channelId: number
  channelName: string
  status: ChannelStatus
  billingMode: BillingMode
  inputPrice: number | null
  outputPrice: number | null
  imageOutputPrice: number | null
  perRequestPrice: number | null
}

const { t } = useI18n()
const appStore = useAppStore()

const rows = ref<AdminModelRow[]>([])
const loading = ref(false)
const searchQuery = ref('')
const statusFilter = ref<'all' | ChannelStatus>('all')
const platformFilter = ref('all')

const statusFilterOptions = computed(() => [
  { value: 'all' as const, label: t('admin.models.filters.allStatuses') },
  { value: CHANNEL_STATUS_ACTIVE, label: statusLabel(CHANNEL_STATUS_ACTIVE) },
  { value: CHANNEL_STATUS_DISABLED, label: statusLabel(CHANNEL_STATUS_DISABLED) },
])

const platformFilterOptions = computed(() => [
  { value: 'all', label: t('admin.models.filters.allPlatforms') },
  ...Array.from(new Set(rows.value.map((row) => row.platform)))
    .sort()
    .map((platform) => ({ value: platform, label: platform })),
])

const filteredRows = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return rows.value.filter((row) => {
    if (statusFilter.value !== 'all' && row.status !== statusFilter.value) return false
    if (platformFilter.value !== 'all' && row.platform !== platformFilter.value) return false
    if (!q) return true
    return row.name.toLowerCase().includes(q) ||
      row.platform.toLowerCase().includes(q) ||
      row.channelName.toLowerCase().includes(q)
  })
})

async function loadModels() {
  loading.value = true
  try {
    const response = await adminAPI.channels.list(1, 1000)
    const nextRows: AdminModelRow[] = []
    for (const channel of response.items || []) {
      for (const pricing of channel.model_pricing || []) {
        for (const model of pricing.models || []) {
          nextRows.push({
            name: model,
            platform: pricing.platform,
            channelId: channel.id,
            channelName: channel.name,
            status: channel.status,
            billingMode: pricing.billing_mode,
            inputPrice: pricing.input_price,
            outputPrice: pricing.output_price,
            imageOutputPrice: pricing.image_output_price,
            perRequestPrice: pricing.per_request_price,
          })
        }
      }
    }
    rows.value = nextRows.sort((a, b) =>
      a.platform.localeCompare(b.platform) ||
      a.name.localeCompare(b.name) ||
      a.channelName.localeCompare(b.channelName),
    )
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

function statusLabel(status: ChannelStatus) {
  if (status === CHANNEL_STATUS_ACTIVE) return t('admin.models.status.active')
  if (status === CHANNEL_STATUS_DISABLED) return t('admin.models.status.disabled')
  return status
}

function billingLabel(mode?: BillingMode) {
  if (mode === BILLING_MODE_PER_REQUEST) return t('availableChannels.pricing.billingModePerRequest')
  if (mode === BILLING_MODE_IMAGE) return t('availableChannels.pricing.billingModeImage')
  if (mode === BILLING_MODE_TOKEN) return t('availableChannels.pricing.billingModeToken')
  return t('availableChannels.noPricing')
}

function priceText(value: number | null | undefined, row: AdminModelRow) {
  if (value == null) return '-'
  const unit = row.billingMode === BILLING_MODE_TOKEN
    ? t('availableChannels.pricing.unitPerMillion')
    : t('availableChannels.pricing.unitPerRequest')
  const scale = row.billingMode === BILLING_MODE_TOKEN ? 1_000_000 : 1
  return `$${(value * scale).toFixed(2)} ${unit}`
}

function outputPriceText(row: AdminModelRow) {
  if (row.billingMode === BILLING_MODE_PER_REQUEST) return priceText(row.perRequestPrice, row)
  if (row.billingMode === BILLING_MODE_IMAGE) return priceText(row.imageOutputPrice, row)
  return priceText(row.outputPrice, row)
}

onMounted(loadModels)
</script>
