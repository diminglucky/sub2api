<template>
  <article
    class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800"
    :class="{ 'source-card-editing': expanded }"
    :data-source-id="sourceId"
  >
    <div class="source-card-header mb-4 flex items-start justify-between gap-4">
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <span class="truncate text-base font-semibold text-gray-900 dark:text-white">
            {{ source.name || t('admin.upstreamMonitor.sources.untitled', { index: index + 1 }) }}
          </span>
          <span
            class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
            :class="source.enabled ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'"
          >
            {{ source.enabled ? t('common.enabled') : t('common.disabled') }}
          </span>
          <span class="status-pill" :class="sourceSyncStatusClass(source.last_sync_status)">
            {{ sourceSyncStatusLabel(t, source.last_sync_status) }}
          </span>
        </div>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ source.pricing_url || source.base_url || t('admin.upstreamMonitor.sources.noUrl') }}
        </p>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ sourceSyncDetail(t, locale, source) }}
        </p>
      </div>
      <div class="flex items-center gap-3">
        <button
          v-if="expanded"
          type="button"
          class="source-edit-collapse-label inline-flex h-9 items-center gap-1.5 rounded-xl border border-gray-200 px-3 text-sm font-medium text-gray-600 transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-600 dark:text-gray-300 dark:hover:border-primary-500/40 dark:hover:bg-primary-950/30 dark:hover:text-primary-200"
          data-source-card-action="collapse"
          @click="emit('toggle-expanded')"
        >
          <Icon name="chevronUp" size="sm" />
          <span>{{ t('common.collapse') }}</span>
        </button>
        <button
          v-else
          type="button"
          class="inline-flex h-9 items-center gap-1.5 rounded-xl border border-gray-200 px-3 text-sm font-medium text-gray-600 transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 dark:border-dark-600 dark:text-gray-300 dark:hover:border-primary-500/40 dark:hover:bg-primary-950/30 dark:hover:text-primary-200"
          data-source-card-action="edit"
          @click="emit('toggle-expanded')"
        >
          <Icon name="edit" size="sm" />
          <span>{{ t('common.edit') }}</span>
        </button>
        <Toggle v-model="source.enabled" />
        <button
          type="button"
          class="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-gray-200 text-gray-500 transition hover:border-red-200 hover:text-red-500 dark:border-dark-600 dark:text-gray-400 dark:hover:border-red-800/40 dark:hover:text-red-300"
          data-source-card-action="remove"
          @click="emit('remove')"
        >
          <Icon name="trash" size="sm" />
        </button>
      </div>
    </div>

    <UpstreamMonitorSourceSummary
      :mapped-groups-count="mappedGroupsCount"
      :account-count="source.account_ids.length"
      :reference-multiplier="Number(source.reference_multiplier || 0)"
      :summary-groups="summaryGroups"
      :hidden-summary-group-count="hiddenSummaryGroupCount"
      :worst-status="worstStatus"
      :lowest-margin-label="lowestMarginLabel"
    />

    <div class="source-edit-detail-panel">
      <section class="source-edit-section">
        <div class="source-section-header">
          <div>
            <h3>{{ t('admin.upstreamMonitor.sources.basic.title') }}</h3>
            <p>{{ t('admin.upstreamMonitor.sources.basic.description') }}</p>
          </div>
          <span class="status-pill" :class="sourceSyncStatusClass(source.last_sync_status)">
            {{ sourceSyncStatusLabel(t, source.last_sync_status) }}
          </span>
        </div>

        <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
          <Input v-model="source.name" :label="t('admin.upstreamMonitor.sources.fields.name')" />
          <div>
            <label class="input-label mb-1.5 block">{{ t('admin.upstreamMonitor.sources.fields.kind') }}</label>
            <select v-model="source.kind" class="input w-full">
              <option v-for="option in sourceKindOptions" :key="option.value" :value="option.value">
                {{ t(option.labelKey) }}
              </option>
            </select>
          </div>
          <Input
            v-model="source.base_url"
            :label="t('admin.upstreamMonitor.sources.fields.baseUrl')"
            :placeholder="t('admin.upstreamMonitor.sources.fields.baseUrlPlaceholder')"
            :hint="t('admin.upstreamMonitor.sources.fields.baseUrlHint')"
          />
          <div class="space-y-2">
            <Input v-model="source.pricing_url" :label="t('admin.upstreamMonitor.sources.fields.pricingUrl')" />
            <div class="flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                data-source-card-action="apply-preset"
                @click="emit('apply-preset')"
              >
                <Icon name="sparkles" size="sm" />
                <span>{{ t('admin.upstreamMonitor.sources.applyPreset') }}</span>
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="!canSync"
                data-source-card-action="sync"
                @click="emit('sync')"
              >
                <Icon name="refresh" size="sm" :class="{ 'animate-spin': sourceSyncing }" />
                <span>
                  {{
                    sourceSyncing
                      ? t('admin.upstreamMonitor.sources.pullingGroups')
                      : t('admin.upstreamMonitor.sources.pullGroups')
                  }}
                </span>
              </button>
              <span class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.upstreamMonitor.sources.pullGroupsHint') }}
              </span>
            </div>
            <div
              v-if="upstreamGroupOptions.length > 0"
              class="rounded-xl border border-emerald-200 bg-emerald-50/70 p-3 text-xs text-emerald-900 dark:border-emerald-900/40 dark:bg-emerald-950/20 dark:text-emerald-100"
            >
              <div class="mb-2 font-medium">
                {{ t('admin.upstreamMonitor.sources.pulledGroupsTitle', { count: upstreamGroupOptions.length }) }}
              </div>
              <div class="flex flex-wrap gap-1.5">
                <span
                  v-for="option in upstreamGroupOptionChips"
                  :key="`${sourceId}_preview_${option.key}`"
                  class="max-w-full rounded-full bg-white/80 px-2 py-1 text-emerald-800 shadow-sm dark:bg-emerald-950/60 dark:text-emerald-100"
                  :title="upstreamGroupOptionLabel(t, option)"
                >
                  {{ upstreamGroupOptionChipLabel(option) }}
                </span>
                <span
                  v-if="hiddenUpstreamGroupOptionCount > 0"
                  class="rounded-full bg-white/80 px-2 py-1 text-emerald-800 shadow-sm dark:bg-emerald-950/60 dark:text-emerald-100"
                >
                  {{ t('admin.upstreamMonitor.sources.morePulledGroups', { count: hiddenUpstreamGroupOptionCount }) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-xl border border-dashed border-gray-200 bg-gray-50/70 p-4 dark:border-dark-600 dark:bg-dark-800/60">
            <div class="flex items-center justify-between gap-4">
              <div>
                <div class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ t('admin.upstreamMonitor.sources.fields.autoSync') }}
                </div>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.upstreamMonitor.sources.fields.autoSyncHint') }}
                </p>
              </div>
              <Toggle v-model="source.auto_sync_enabled" />
            </div>
          </div>
          <Input
            v-model="source.reference_multiplier"
            type="number"
            min="0"
            step="any"
            :label="t('admin.upstreamMonitor.sources.fields.referenceMultiplier')"
            :hint="t('admin.upstreamMonitor.sources.fields.referenceMultiplierHint')"
          />
        </div>

        <SourceMappingRowsEditor
          :source-id="sourceId"
          :bound-count="mappedGroupsCount"
          :total-count="mappingRowsTotalCount"
          :can-add="canAddMapping"
          :rows="mappingRows"
          :local-group-options="localGroupOptions"
          :upstream-group-options="upstreamGroupOptions"
          @add="emit('add-mapping')"
          @remove="emit('remove-mapping', $event)"
          @bind="emit('bind-mapping', $event)"
          @update-local="handleUpdateLocalMapping"
          @select-upstream="handleSelectUpstreamMapping"
        />
      </section>

      <UpstreamMonitorSourceAdvanced
        :source="source"
        :account-options="accountOptions"
        :auth-mode-options="authModeOptions"
        :fetch-mode-options="fetchModeOptions"
        @toggle-account="emit('toggle-account', $event)"
      />
    </div>
  </article>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
import Input from "@/components/common/Input.vue";
import Toggle from "@/components/common/Toggle.vue";
import type {
  UpstreamMonitorGroupMapping,
  UpstreamMonitorPreviewAccountInfo,
  UpstreamMonitorSourceConfig,
  UpstreamMonitorUpstreamGroupOption,
} from "@/api/admin/settings";
import SourceMappingRowsEditor from "./SourceMappingRowsEditor.vue";
import UpstreamMonitorSourceAdvanced from "./UpstreamMonitorSourceAdvanced.vue";
import UpstreamMonitorSourceSummary from "./UpstreamMonitorSourceSummary.vue";
import {
  sourceSyncDetail,
  sourceSyncStatusClass,
  sourceSyncStatusLabel,
  upstreamGroupOptionChipLabel,
  upstreamGroupOptionLabel,
} from "./sourceCardFormat";
import type {
  SourceFetchModeOptionConfig,
  SourceSelectOptionConfig,
  SourceSummaryGroupView,
} from "./sourceCardTypes";
import type {
  SourceMappingRowView,
  SourceMappingSelectOption,
  SourceMappingUpdatePayload,
} from "./useSourceMappingRows";

defineProps<{
  source: UpstreamMonitorSourceConfig;
  sourceId: string;
  index: number;
  expanded: boolean;
  sourceSyncing: boolean;
  canSync: boolean;
  mappedGroupsCount: number;
  mappingRowsTotalCount: number;
  canAddMapping: boolean;
  mappingRows: SourceMappingRowView[];
  localGroupOptions: SourceMappingSelectOption[];
  upstreamGroupOptions: SourceMappingSelectOption[];
  upstreamGroupOptionChips: UpstreamMonitorUpstreamGroupOption[];
  hiddenUpstreamGroupOptionCount: number;
  summaryGroups: SourceSummaryGroupView[];
  hiddenSummaryGroupCount: number;
  worstStatus: string;
  lowestMarginLabel: string;
  accountOptions: UpstreamMonitorPreviewAccountInfo[];
  sourceKindOptions: ReadonlyArray<SourceSelectOptionConfig>;
  authModeOptions: ReadonlyArray<SourceSelectOptionConfig>;
  fetchModeOptions: ReadonlyArray<SourceFetchModeOptionConfig>;
}>();

const emit = defineEmits<{
  (event: "toggle-expanded"): void;
  (event: "remove"): void;
  (event: "sync"): void;
  (event: "apply-preset"): void;
  (event: "toggle-account", accountID: number): void;
  (event: "add-mapping"): void;
  (event: "remove-mapping", mapping: UpstreamMonitorGroupMapping): void;
  (event: "bind-mapping", mapping: UpstreamMonitorGroupMapping): void;
  (event: "update-local-mapping", payload: SourceMappingUpdatePayload): void;
  (event: "select-upstream-mapping", payload: SourceMappingUpdatePayload): void;
}>();

const { t, locale } = useI18n();

function handleUpdateLocalMapping(payload: SourceMappingUpdatePayload) {
  emit("update-local-mapping", payload);
}

function handleSelectUpstreamMapping(payload: SourceMappingUpdatePayload) {
  emit("select-upstream-mapping", payload);
}
</script>

<style scoped>
.status-pill {
  @apply inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium;
}

.status-healthy {
  @apply bg-emerald-100 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300;
}

.status-warning {
  @apply bg-amber-100 text-amber-700 dark:bg-amber-900/20 dark:text-amber-300;
}

.status-critical {
  @apply bg-rose-100 text-rose-700 dark:bg-rose-900/20 dark:text-rose-300;
}

.status-unknown {
  @apply bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300;
}

.source-edit-collapse-label,
.source-edit-detail-panel {
  display: none;
}

.source-card-editing .source-edit-collapse-label {
  display: inline-flex;
}

.source-card-editing .source-summary-panel {
  display: none;
}

.source-card-editing .source-edit-detail-panel {
  display: block;
}

.source-edit-section {
  @apply rounded-2xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800;
}

.source-section-header {
  @apply flex flex-col gap-3 border-b border-gray-100 pb-4 dark:border-dark-700 sm:flex-row sm:items-start sm:justify-between;
}

.source-section-header h3 {
  @apply text-base font-semibold text-gray-900 dark:text-white;
}

.source-section-header p {
  @apply mt-1 text-sm text-gray-500 dark:text-gray-400;
}
</style>
