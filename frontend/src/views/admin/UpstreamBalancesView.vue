<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="relative w-full sm:w-72">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input v-model="search" class="input pl-10" type="search" :placeholder="t('admin.upstreamBalances.search')" />
          </div>
          <div class="w-full sm:w-44">
            <Select v-model="statusFilter" :options="statusOptions" />
          </div>
          <div class="flex flex-1 justify-end">
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="refreshingAll || loading"
              :title="t('admin.upstreamBalances.refreshAll')"
              @click="refreshBalances()"
            >
              <Icon name="refresh" size="md" :class="refreshingAll ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <div class="min-h-0 flex-1 overflow-auto pb-8">
          <div class="grid grid-cols-2 border-b border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900 lg:grid-cols-4">
            <div class="border-r border-gray-200 px-5 py-4 dark:border-dark-700">
              <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.upstreamBalances.totalBalance') }}</div>
              <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ formatCNY(overview.total_amount) }}</div>
            </div>
            <div class="border-r border-gray-200 px-5 py-4 dark:border-dark-700">
              <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.upstreamBalances.platforms') }}</div>
              <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ overview.platforms.length }}</div>
            </div>
            <div class="border-r border-gray-200 px-5 py-4 dark:border-dark-700">
              <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.upstreamBalances.lowBalance') }}</div>
              <div class="mt-1 text-2xl font-semibold" :class="lowBalanceCount > 0 ? 'text-red-600 dark:text-red-400' : 'text-gray-900 dark:text-white'">
                {{ lowBalanceCount }}
              </div>
            </div>
            <div class="px-5 py-4">
              <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.upstreamBalances.staleData') }}</div>
              <div class="mt-1 text-2xl font-semibold" :class="staleCount > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-gray-900 dark:text-white'">
                {{ staleCount }}
              </div>
            </div>
          </div>

          <section class="mt-7">
            <h2 class="mb-3 px-1 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.upstreamBalances.platformSummary') }}</h2>
            <DataTable :columns="platformColumns" :data="overview.platforms" :loading="loading">
              <template #cell-platform_name="{ row }">
                <span class="font-medium text-gray-900 dark:text-white">{{ row.platform_name }}</span>
              </template>
              <template #cell-amount="{ row }">
                <span class="font-mono font-semibold text-gray-900 dark:text-white">{{ formatCNY(row.amount) }}</span>
              </template>
              <template #cell-counts="{ row }">
                <span class="text-sm text-gray-700 dark:text-gray-300">{{ row.account_count }} / {{ row.funding_count }}</span>
              </template>
              <template #cell-status="{ row }">
                <div class="flex flex-wrap gap-1">
                  <span v-if="row.low_balance_count" class="badge badge-danger">{{ t('admin.upstreamBalances.low') }} {{ row.low_balance_count }}</span>
                  <span v-if="row.error_count" class="badge badge-danger">{{ t('admin.upstreamBalances.error') }} {{ row.error_count }}</span>
                  <span v-if="row.stale_count" class="badge badge-warning">{{ t('admin.upstreamBalances.stale') }} {{ row.stale_count }}</span>
                  <span v-if="!row.low_balance_count && !row.error_count && !row.stale_count" class="badge badge-success">{{ t('admin.upstreamBalances.normal') }}</span>
                </div>
              </template>
              <template #cell-updated_at="{ row }">
                <span class="text-xs text-gray-600 dark:text-gray-300">{{ formatTime(row.updated_at) }}</span>
              </template>
              <template #empty>
                <div class="py-8 text-sm text-gray-500">{{ t('admin.upstreamBalances.noConfiguredPlatforms') }}</div>
              </template>
            </DataTable>
          </section>

          <section class="mt-8">
            <h2 class="mb-3 px-1 text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.upstreamBalances.accountSources') }}</h2>
            <DataTable :columns="sourceColumns" :data="filteredSources" :loading="loading" row-key="account_id">
              <template #cell-account="{ row }">
                <div class="min-w-0">
                  <div class="font-medium text-gray-900 dark:text-white">{{ row.account_name }}</div>
                  <div class="text-xs text-gray-500">#{{ row.account_id }}</div>
                </div>
              </template>
              <template #cell-platform="{ row }">
                <div class="flex flex-col gap-1">
                  <span>{{ row.config.platform_name || '-' }}</span>
                  <span class="text-xs uppercase text-gray-500">{{ row.config.provider }}</span>
                </div>
              </template>
              <template #cell-protocol="{ row }">
                <span class="badge badge-gray">{{ row.protocol }} / {{ row.account_type }}</span>
              </template>
              <template #cell-base_url="{ row }">
                <span class="block max-w-56 truncate font-mono text-xs text-gray-600 dark:text-gray-300" :title="row.base_url">{{ row.base_url || '-' }}</span>
              </template>
              <template #cell-balance="{ row }">
                <span v-if="typeof row.snapshot.amount === 'number'" class="font-mono font-semibold text-gray-900 dark:text-white">
                  {{ formatCNY(row.snapshot.amount) }}
                </span>
                <span v-else class="text-gray-400">-</span>
              </template>
              <template #cell-threshold="{ row }">
                <span v-if="row.config.threshold > 0" class="font-mono text-sm">{{ formatCNY(row.config.threshold) }}</span>
                <span v-else class="text-gray-400">-</span>
              </template>
              <template #cell-status="{ row }">
                <span :class="['badge', statusClass(row)]" :title="row.snapshot.last_error || undefined">{{ statusLabel(row) }}</span>
              </template>
              <template #cell-updated_at="{ row }">
                <span class="text-xs text-gray-600 dark:text-gray-300">{{ formatTime(row.snapshot.last_success_at) }}</span>
              </template>
              <template #cell-actions="{ row }">
                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    class="flex h-8 w-8 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-primary-600 disabled:opacity-40 dark:hover:bg-dark-700"
                    :disabled="!row.config.enabled || refreshingIds.has(row.account_id)"
                    :title="t('admin.upstreamBalances.refreshOne')"
                    @click="refreshBalances(row.account_id)"
                  >
                    <Icon name="refresh" size="sm" :class="refreshingIds.has(row.account_id) ? 'animate-spin' : ''" />
                  </button>
                  <button type="button" class="flex h-8 w-8 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700" :title="t('admin.upstreamBalances.configure')" @click="openConfig(row)">
                    <Icon name="cog" size="sm" />
                  </button>
                </div>
              </template>
              <template #empty>
                <div class="py-8 text-sm text-gray-500">{{ t('admin.upstreamBalances.noSources') }}</div>
              </template>
            </DataTable>
          </section>
        </div>
      </template>
    </TablePageLayout>

    <BaseDialog :show="showConfig" :title="t('admin.upstreamBalances.configure')" width="normal" @close="closeConfig">
      <form id="upstream-balance-config-form" class="space-y-5" @submit.prevent="saveConfig">
        <label class="flex items-center justify-between gap-4">
          <span class="input-label mb-0">{{ t('admin.upstreamBalances.enabled') }}</span>
          <input v-model="form.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
        </label>

        <div>
          <label class="input-label">{{ t('admin.upstreamBalances.platformName') }}</label>
          <input v-model.trim="form.platform_name" class="input" type="text" :disabled="!form.enabled" required />
        </div>

        <div>
          <label class="input-label">{{ t('admin.upstreamBalances.provider') }}</label>
          <Select v-model="form.provider" :options="providerOptions" :disabled="!form.enabled" />
        </div>

        <template v-if="usesWalletAuth">
          <div class="border-t border-gray-200 pt-5 dark:border-dark-700">
            <div class="mb-4 flex items-center justify-between gap-3">
              <label class="input-label mb-0">{{ t('admin.upstreamBalances.walletAuth') }}</label>
              <span :class="['badge', form.auth_configured && !form.auth_cleared ? 'badge-success' : 'badge-gray']">
                {{ form.auth_configured && !form.auth_cleared ? t('admin.upstreamBalances.authConfigured') : t('admin.upstreamBalances.authPending') }}
              </span>
            </div>
            <div class="space-y-4">
              <div>
                <label class="input-label">{{ t('admin.upstreamBalances.authMode') }}</label>
                <Select v-model="form.auth_mode" :options="authModeOptions" :disabled="!form.enabled" />
              </div>
              <div v-if="form.auth_mode === 'login'">
                <label class="input-label">{{ t('admin.upstreamBalances.loginEmail') }}</label>
                <input v-model.trim="form.auth_username" class="input" type="email" autocomplete="off" :disabled="!form.enabled" required />
              </div>
              <div>
                <label class="input-label">{{ authTokenLabel }}</label>
                <input
                  v-model="form.auth_token"
                  class="input"
                  type="password"
                  autocomplete="new-password"
                  :placeholder="form.auth_configured && !form.auth_cleared ? t('admin.upstreamBalances.authSecretPreserved') : ''"
                  :disabled="!form.enabled"
                  :required="form.enabled && (!form.auth_configured || form.auth_cleared)"
                  @input="form.auth_cleared = false"
                />
              </div>
              <button
                v-if="form.auth_configured && !form.auth_cleared"
                type="button"
                class="inline-flex items-center gap-1.5 text-sm font-medium text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
                @click="clearWalletAuth"
              >
                <Icon name="x" size="xs" />
                <span>{{ t('admin.upstreamBalances.clearAuth') }}</span>
              </button>
            </div>
          </div>
        </template>

        <template v-if="form.provider === 'custom'">
          <div>
            <label class="input-label">{{ t('admin.upstreamBalances.endpoint') }}</label>
            <input v-model.trim="form.endpoint" class="input font-mono" type="text" :disabled="!form.enabled" required />
          </div>
          <div>
            <label class="input-label">{{ t('admin.upstreamBalances.jsonPath') }}</label>
            <input v-model.trim="form.json_path" class="input font-mono" type="text" :disabled="!form.enabled" required />
          </div>
          <div>
            <label class="input-label">{{ t('admin.upstreamBalances.divisor') }}</label>
            <input v-model.number="form.divisor" class="input" type="number" min="0.000001" step="any" :disabled="!form.enabled" />
          </div>
        </template>

        <div>
          <label class="input-label">{{ t('admin.upstreamBalances.fundingKey') }}</label>
          <input v-model.trim="form.funding_key" class="input" type="text" :disabled="!form.enabled" />
        </div>

        <div>
          <label class="input-label">{{ t('admin.upstreamBalances.threshold') }}</label>
          <input v-model.number="form.threshold" class="input" type="number" min="0" step="0.01" :disabled="!form.enabled" />
        </div>

        <label class="flex items-center justify-between gap-4">
          <span class="input-label mb-0">{{ t('admin.upstreamBalances.notifyEnabled') }}</span>
          <input v-model="form.notify_enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 disabled:opacity-40" :disabled="!form.enabled || form.threshold <= 0" />
        </label>
      </form>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="closeConfig">{{ t('common.cancel') }}</button>
        <button type="submit" form="upstream-balance-config-form" class="btn btn-primary" :disabled="saving">
          {{ saving ? t('admin.upstreamBalances.saving') : t('admin.upstreamBalances.save') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type {
  UpstreamBalanceConfig,
  UpstreamBalanceOverview,
  UpstreamBalanceProvider,
  UpstreamBalanceSource
} from '@/api/admin/upstreamBalances'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const emptyOverview = (): UpstreamBalanceOverview => ({ total_amount: 0, currency: 'CNY', platforms: [], sources: [] })
const overview = ref<UpstreamBalanceOverview>(emptyOverview())
const loading = ref(false)
const refreshingAll = ref(false)
const refreshingIds = reactive(new Set<number>())
const search = ref('')
const statusFilter = ref('')
const showConfig = ref(false)
const saving = ref(false)
const editingSource = ref<UpstreamBalanceSource | null>(null)

const form = reactive<UpstreamBalanceConfig>({
  enabled: false,
  provider: 'auto',
  platform_name: '',
  threshold: 0,
  notify_enabled: false,
  endpoint: '',
  json_path: '',
  divisor: 1,
  funding_key: '',
  auth_mode: '',
  auth_username: '',
  auth_token: '',
  auth_cleared: false,
  auth_configured: false
})

const platformColumns = computed<Column[]>(() => [
  { key: 'platform_name', label: t('admin.upstreamBalances.platformName') },
  { key: 'amount', label: t('admin.upstreamBalances.currentBalance') },
  { key: 'counts', label: `${t('admin.upstreamBalances.accountCount')} / ${t('admin.upstreamBalances.fundingCount')}` },
  { key: 'status', label: t('admin.upstreamBalances.status') },
  { key: 'updated_at', label: t('admin.upstreamBalances.updatedAt') }
])

const sourceColumns = computed<Column[]>(() => [
  { key: 'account', label: t('admin.upstreamBalances.account') },
  { key: 'platform', label: t('admin.upstreamBalances.platformName') },
  { key: 'protocol', label: t('admin.upstreamBalances.protocol') },
  { key: 'base_url', label: t('admin.upstreamBalances.baseUrl') },
  { key: 'balance', label: t('admin.upstreamBalances.currentBalance') },
  { key: 'threshold', label: t('admin.upstreamBalances.thresholdColumn') },
  { key: 'status', label: t('admin.upstreamBalances.status') },
  { key: 'updated_at', label: t('admin.upstreamBalances.updatedAt') },
  { key: 'actions', label: t('admin.upstreamBalances.actions') }
])

const providerOptions = computed(() => [
  { value: 'auto', label: t('admin.upstreamBalances.auto') },
  { value: 'deepseek', label: t('admin.upstreamBalances.deepseek') },
  { value: 'stepfun', label: t('admin.upstreamBalances.stepfun') },
  { value: 'siliconflow', label: t('admin.upstreamBalances.siliconflow') },
  { value: 'newapi', label: t('admin.upstreamBalances.newapi') },
  { value: 'sub2api', label: t('admin.upstreamBalances.sub2api') },
  { value: 'custom', label: t('admin.upstreamBalances.custom') }
])

const usesWalletAuth = computed(() => form.provider === 'newapi' || form.provider === 'sub2api')
const authModeOptions = computed(() => form.provider === 'newapi'
  ? [
      { value: 'login', label: t('admin.upstreamBalances.loginAuth') },
      { value: 'cookie', label: t('admin.upstreamBalances.cookieAuth') }
    ]
  : [
      { value: 'login', label: t('admin.upstreamBalances.loginAuth') },
      { value: 'bearer', label: t('admin.upstreamBalances.bearerAuth') }
    ])
const authTokenLabel = computed(() => {
  if (form.auth_mode === 'login') return t('admin.upstreamBalances.loginPassword')
  if (form.auth_mode === 'cookie') return t('admin.upstreamBalances.cookieCredential')
  return t('admin.upstreamBalances.accessToken')
})

const statusOptions = computed(() => [
  { value: '', label: t('admin.upstreamBalances.allStatuses') },
  { value: 'enabled', label: t('admin.upstreamBalances.enabledOnly') },
  { value: 'low', label: t('admin.upstreamBalances.lowOnly') },
  { value: 'error', label: t('admin.upstreamBalances.errorOnly') },
  { value: 'disabled', label: t('admin.upstreamBalances.disabledOnly') }
])

const lowBalanceCount = computed(() => overview.value.sources.filter(source => source.snapshot.alert_active).length)
const staleCount = computed(() => overview.value.sources.filter(source => source.config.enabled && source.snapshot.stale).length)

const filteredSources = computed(() => {
  const query = search.value.trim().toLowerCase()
  return overview.value.sources.filter((source) => {
    if (query && ![source.account_name, source.config.platform_name, source.base_url].some(value => value.toLowerCase().includes(query))) {
      return false
    }
    switch (statusFilter.value) {
      case 'enabled': return source.config.enabled
      case 'disabled': return !source.config.enabled
      case 'low': return source.snapshot.alert_active
      case 'error': return Boolean(source.snapshot.last_error)
      default: return true
    }
  })
})

function formatCNY(value: number): string {
  return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY', minimumFractionDigits: 2 }).format(value || 0)
}

function formatTime(value?: string): string {
  return value ? formatDateTime(value) : '-'
}

function statusKey(source: UpstreamBalanceSource): string {
  if (!source.config.enabled) return 'disabled'
  if (source.snapshot.last_error) return 'error'
  if (!source.snapshot.last_success_at) return 'waiting'
  if (source.snapshot.alert_active) return 'low'
  if (source.snapshot.stale) return 'stale'
  return 'normal'
}

function statusLabel(source: UpstreamBalanceSource): string {
  return t(`admin.upstreamBalances.${statusKey(source)}`)
}

function statusClass(source: UpstreamBalanceSource): string {
  switch (statusKey(source)) {
    case 'normal': return 'badge-success'
    case 'disabled': return 'badge-gray'
    case 'waiting': return 'badge-gray'
    case 'stale': return 'badge-warning'
    default: return 'badge-danger'
  }
}

async function loadOverview(): Promise<void> {
  loading.value = true
  try {
    overview.value = await adminAPI.upstreamBalances.getOverview()
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('admin.upstreamBalances.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function refreshBalances(accountId?: number): Promise<void> {
  if (accountId) refreshingIds.add(accountId)
  else refreshingAll.value = true
  try {
    const result = await adminAPI.upstreamBalances.refresh(accountId)
    appStore.showSuccess(t('admin.upstreamBalances.refreshResult', {
      succeeded: result.succeeded,
      failed: result.failed
    }))
    await loadOverview()
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('admin.upstreamBalances.refreshFailed')))
  } finally {
    if (accountId) refreshingIds.delete(accountId)
    else refreshingAll.value = false
  }
}

function openConfig(source: UpstreamBalanceSource): void {
  editingSource.value = source
  Object.assign(form, {
    enabled: source.config.enabled,
    provider: (source.config.provider || 'auto') as UpstreamBalanceProvider,
    platform_name: source.config.platform_name || '',
    threshold: source.config.threshold || 0,
    notify_enabled: source.config.notify_enabled,
    endpoint: source.config.endpoint || '',
    json_path: source.config.json_path || '',
    divisor: source.config.divisor || 1,
    funding_key: source.config.funding_key || '',
    auth_mode: source.config.auth_mode || '',
    auth_username: source.config.auth_username || '',
    auth_token: '',
    auth_cleared: false,
    auth_configured: Boolean(source.config.auth_configured)
  })
  showConfig.value = true
}

function closeConfig(): void {
  showConfig.value = false
  editingSource.value = null
}

async function saveConfig(): Promise<void> {
  if (!editingSource.value) return
  saving.value = true
  try {
    await adminAPI.upstreamBalances.configure(editingSource.value.account_id, { ...form })
    if (form.enabled) await adminAPI.upstreamBalances.refresh(editingSource.value.account_id)
    appStore.showSuccess(t('admin.upstreamBalances.saved'))
    closeConfig()
    await loadOverview()
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('admin.upstreamBalances.saveFailed')))
  } finally {
    saving.value = false
  }
}

function clearWalletAuth(): void {
  form.auth_token = ''
  form.auth_cleared = true
}

watch(() => form.provider, (provider) => {
  if (provider === 'newapi' && form.auth_mode !== 'login' && form.auth_mode !== 'cookie') {
    form.auth_mode = 'login'
  } else if (provider === 'sub2api' && form.auth_mode !== 'login' && form.auth_mode !== 'bearer') {
    form.auth_mode = 'login'
  }
})

onMounted(loadOverview)
</script>
