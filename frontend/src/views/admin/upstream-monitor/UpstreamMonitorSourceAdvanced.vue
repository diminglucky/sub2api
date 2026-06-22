<template>
  <details class="source-advanced-panel">
    <summary class="source-advanced-summary">
      <div>
        <span class="source-advanced-title">{{ t('admin.upstreamMonitor.sources.advanced.title') }}</span>
        <span class="source-advanced-description">{{ t('admin.upstreamMonitor.sources.advanced.description') }}</span>
      </div>
      <Icon name="chevronDown" size="sm" class="source-advanced-chevron" />
    </summary>

    <div class="source-advanced-content">
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div>
          <label class="input-label mb-1.5 block">{{ t('admin.upstreamMonitor.sources.fields.currency') }}</label>
          <select v-model="source.currency" class="input w-full">
            <option value="CNY">CNY</option>
            <option value="USD">USD</option>
          </select>
        </div>
        <Input v-model="source.exchange_rate" type="number" min="0" step="any" :label="t('admin.upstreamMonitor.sources.fields.exchangeRate')" />
      </div>

      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div>
          <label class="input-label mb-1.5 block">{{ t('admin.upstreamMonitor.sources.fields.fetchMode') }}</label>
          <select v-model="source.fetch_mode" class="input w-full">
            <option v-for="option in fetchModeOptions" :key="option.value" :value="option.value">
              {{ t(option.labelKey) }}
            </option>
          </select>
        </div>
        <Input
          v-model="source.pricing_path_hint"
          :label="t('admin.upstreamMonitor.sources.fields.pricingHint')"
          :hint="t('admin.upstreamMonitor.sources.fields.pricingHintHint')"
        />
        <div>
          <label class="input-label mb-1.5 block">{{ t('admin.upstreamMonitor.sources.fields.authMode') }}</label>
          <select v-model="source.auth_mode" class="input w-full">
            <option v-for="option in authModeOptions" :key="option.value" :value="option.value">
              {{ t(option.labelKey) }}
            </option>
          </select>
        </div>
        <Input
          v-if="source.auth_mode === 'header'"
          v-model="source.auth_header_name"
          :label="t('admin.upstreamMonitor.sources.fields.authHeaderName')"
          :hint="t('admin.upstreamMonitor.sources.fields.authHeaderNameHint')"
        />
        <div v-else class="rounded-xl border border-dashed border-gray-200 bg-gray-50/70 p-4 dark:border-dark-600 dark:bg-dark-800/60">
          <div class="text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.upstreamMonitor.sources.fields.status') }}
          </div>
          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
            {{ source.auth_configured ? t('admin.upstreamMonitor.sources.statusConfigured') : t('admin.upstreamMonitor.sources.statusPending') }}
          </p>
        </div>
      </div>

      <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        <Input
          v-model="source.auth_token"
          :label="t('admin.upstreamMonitor.sources.fields.authToken')"
          :placeholder="source.auth_configured ? t('admin.upstreamMonitor.sources.tokenMasked') : ''"
          :hint="source.auth_configured ? t('admin.upstreamMonitor.sources.tokenConfigured') : t('admin.upstreamMonitor.sources.tokenHint')"
        />
        <div class="rounded-xl border border-dashed border-gray-200 bg-gray-50/70 p-4 dark:border-dark-600 dark:bg-dark-800/60">
          <div class="text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.upstreamMonitor.sources.fields.status') }}
          </div>
          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
            {{ sourceSyncDetail(t, locale, source) }}
          </p>
        </div>
        <div v-if="source.last_sync_error" class="rounded-xl border border-rose-200 bg-rose-50/70 p-4 text-sm text-rose-700 dark:border-rose-900/40 dark:bg-rose-950/20 dark:text-rose-300">
          {{ source.last_sync_error }}
        </div>
      </div>

      <div class="mt-4">
        <div class="mb-2 flex items-center justify-between gap-3">
          <label class="input-label">{{ t('admin.upstreamMonitor.sources.fields.accountIds') }}</label>
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('admin.upstreamMonitor.sources.fields.selectedCount', { count: source.account_ids.length }) }}
          </span>
        </div>
        <div
          v-if="accountOptions.length === 0"
          class="rounded-xl border border-dashed border-gray-200 px-4 py-6 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
        >
          {{ t('admin.upstreamMonitor.sources.fields.noAccountOptions') }}
        </div>
        <div v-else class="choice-panel">
          <button
            v-for="account in accountOptions"
            :key="account.account_id"
            type="button"
            class="choice-card"
            :title="accountOptionLabel(t, account)"
            :class="{ 'choice-card-selected': isAccountSelected(account.account_id) }"
            @click="emit('toggle-account', account.account_id)"
          >
            <span class="choice-check" :class="{ 'choice-check-selected': isAccountSelected(account.account_id) }">
              <Icon v-if="isAccountSelected(account.account_id)" name="check" size="xs" />
            </span>
            <span class="min-w-0 flex-1">
              <span class="block truncate text-sm font-medium text-gray-900 dark:text-white">
                {{ account.account_name || `#${account.account_id}` }}
              </span>
              <span class="mt-1 block text-xs text-gray-500 dark:text-gray-400">
                {{ account.platform || "--" }} · {{ account.type || "--" }} · {{ formatSourceDecimal(account.rate_multiplier) }}x
              </span>
              <span class="mt-2 flex flex-wrap gap-1.5">
                <span
                  v-for="groupName in account.group_names.slice(0, 3)"
                  :key="`${account.account_id}_${groupName}`"
                  class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] text-gray-600 dark:bg-dark-700 dark:text-gray-300"
                >
                  {{ groupName }}
                </span>
                <span
                  v-if="account.group_names.length > 3"
                  class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] text-gray-600 dark:bg-dark-700 dark:text-gray-300"
                >
                  +{{ account.group_names.length - 3 }}
                </span>
              </span>
            </span>
          </button>
        </div>
        <p class="input-hint mt-1.5">{{ t('admin.upstreamMonitor.sources.fields.accountIdsHint') }}</p>
      </div>

      <div class="mt-4">
        <TextArea
          v-model="source.notes"
          :label="t('admin.upstreamMonitor.sources.fields.notes')"
          :rows="2"
        />
      </div>
    </div>
  </details>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
import Input from "@/components/common/Input.vue";
import TextArea from "@/components/common/TextArea.vue";
import type {
  UpstreamMonitorPreviewAccountInfo,
  UpstreamMonitorSourceConfig,
} from "@/api/admin/settings";
import {
  accountOptionLabel,
  formatSourceDecimal,
  sourceSyncDetail,
  uniqueSourceNumberIDs,
} from "./sourceCardFormat";
import type { SourceFetchModeOptionConfig, SourceSelectOptionConfig } from "./sourceCardTypes";

const props = defineProps<{
  source: UpstreamMonitorSourceConfig;
  accountOptions: UpstreamMonitorPreviewAccountInfo[];
  authModeOptions: ReadonlyArray<SourceSelectOptionConfig>;
  fetchModeOptions: ReadonlyArray<SourceFetchModeOptionConfig>;
}>();

const emit = defineEmits<{
  (event: "toggle-account", accountID: number): void;
}>();

const { t, locale } = useI18n();

function isAccountSelected(accountID: number): boolean {
  return uniqueSourceNumberIDs(props.source.account_ids).includes(accountID);
}
</script>

<style scoped>
.source-advanced-panel {
  @apply mt-4 overflow-hidden rounded-2xl border border-gray-200 bg-gray-50/70 dark:border-dark-700 dark:bg-dark-900/30;
}

.source-advanced-summary {
  @apply flex cursor-pointer list-none items-center justify-between gap-4 px-4 py-3 transition-colors hover:bg-gray-100/70 dark:hover:bg-dark-800/70;
}

.source-advanced-summary::-webkit-details-marker {
  display: none;
}

.source-advanced-title {
  @apply block text-sm font-semibold text-gray-900 dark:text-white;
}

.source-advanced-description {
  @apply mt-0.5 block text-xs text-gray-500 dark:text-gray-400;
}

.source-advanced-chevron {
  @apply shrink-0 text-gray-500 transition-transform dark:text-gray-400;
}

.source-advanced-panel[open] .source-advanced-chevron {
  @apply rotate-180;
}

.source-advanced-content {
  @apply border-t border-gray-200 px-4 py-4 dark:border-dark-700;
}

.choice-panel {
  @apply grid max-h-72 grid-cols-1 gap-2 overflow-y-auto rounded-xl border border-gray-200 bg-gray-50/60 p-2 dark:border-dark-600 dark:bg-dark-900/30;
}

.choice-card {
  @apply flex min-h-[4.5rem] w-full items-start gap-3 rounded-xl border border-gray-200 bg-white px-3 py-3 text-left transition-colors hover:border-primary-200 hover:bg-primary-50/40 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-primary-500/40 dark:hover:bg-primary-950/20;
}

.choice-card-selected {
  @apply border-primary-300 bg-primary-50/70 shadow-sm dark:border-primary-500/60 dark:bg-primary-950/30;
}

.choice-check {
  @apply mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border border-gray-300 text-transparent dark:border-dark-500;
}

.choice-check-selected {
  @apply border-primary-500 bg-primary-500 text-white;
}
</style>
