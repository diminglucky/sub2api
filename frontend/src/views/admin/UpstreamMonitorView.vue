<template>
  <AppLayout>
    <div class="upstream-monitor-page mx-auto max-w-7xl space-y-6 pb-12">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
      </div>

      <form v-else novalidate class="upstream-monitor-form space-y-6" @submit.prevent="save">
        <section class="card">
          <div class="flex items-center justify-between gap-4 border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.upstreamMonitor.sources.title') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.upstreamMonitor.sources.description') }}
              </p>
            </div>
            <div class="flex items-center gap-3">
              <button type="button" class="btn btn-secondary btn-sm" :disabled="syncing" @click="syncSources">
                <Icon name="server" size="sm" :class="{ 'animate-pulse': syncing }" />
                <span>{{ syncing ? t('admin.upstreamMonitor.syncing') : t('admin.upstreamMonitor.sync') }}</span>
              </button>
              <button type="button" class="btn btn-primary btn-sm" @click="addSource">
                <Icon name="plus" size="sm" />
                <span>{{ t('admin.upstreamMonitor.sources.add') }}</span>
              </button>
            </div>
          </div>

          <div class="space-y-4 p-6">
            <div class="rounded-2xl border border-sky-200 bg-sky-50/70 px-4 py-3 text-sm text-sky-700 dark:border-sky-800/40 dark:bg-sky-900/10 dark:text-sky-300">
              {{ t('admin.upstreamMonitor.sources.syncHint') }}
            </div>

            <div class="grid grid-cols-1 gap-3 rounded-2xl border border-gray-200 bg-gray-50/70 p-4 dark:border-dark-700 dark:bg-dark-900/30 md:grid-cols-[minmax(0,1fr)_220px] md:items-center">
              <div>
                <div class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.upstreamMonitor.sources.refreshPolicy.title') }}
                </div>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.upstreamMonitor.sources.refreshPolicy.description', { minutes: form.refresh_interval_minutes }) }}
                </p>
              </div>
              <div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3">
                <Input
                  v-model="refreshIntervalInput"
                  type="number"
                  min="1"
                  step="1"
                  :label="t('admin.upstreamMonitor.sources.refreshPolicy.intervalLabel')"
                />
                <div class="pt-6">
                  <Toggle v-model="form.auto_refresh_enabled" />
                </div>
              </div>
            </div>

            <div
              v-if="form.sources.length === 0"
              class="rounded-xl border border-dashed border-gray-200 px-4 py-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
            >
              {{ t('admin.upstreamMonitor.sources.empty') }}
            </div>

            <UpstreamMonitorSourceCard
              v-for="(source, index) in form.sources"
              :key="source.id || index"
              :source="source"
              :source-id="sourceID(source)"
              :index="index"
              :expanded="isSourceExpanded(source)"
              :source-syncing="isSourceSyncing(source)"
              :can-sync="canSyncSource(source)"
              :mapped-groups-count="sourceMappedGroupsCount(source)"
              :mapping-rows-total-count="sourceMappingRowsForTemplate(source).length"
              :can-add-mapping="groupOptionsForSource(source).length > 0"
              :mapping-rows="sourceMappingRowViews(source)"
              :local-group-options="localGroupSelectOptionsForSource(source)"
              :upstream-group-options="upstreamGroupSelectOptionsForSource(source)"
              :upstream-group-option-chips="upstreamGroupOptionsPreview(source)"
              :hidden-upstream-group-option-count="hiddenUpstreamGroupOptionCount(source)"
              :summary-groups="visibleSourceSummaryGroups(source)"
              :hidden-summary-group-count="hiddenSourceSummaryGroupCount(source)"
              :worst-status="sourceWorstStatus(source)"
              :lowest-margin-label="sourceLowestMarginLabel(source)"
              :account-options="preview.account_options"
              :source-kind-options="sourceKindOptions"
              :auth-mode-options="authModeOptions"
              :fetch-mode-options="fetchModeOptions"
              @toggle-expanded="toggleSourceExpanded(source)"
              @remove="removeSource(index)"
              @update-source="updateSourceField(source, $event)"
              @apply-preset="applySourcePreset(source)"
              @sync="syncSource(source)"
              @toggle-account="toggleSourceAccount(source, $event)"
              @add-mapping="addSourceGroupMapping(source)"
              @remove-mapping="removeSourceGroupMapping(source, $event)"
              @bind-mapping="bindSourceGroupMapping(source, $event)"
              @update-local-mapping="updateMappingLocalGroup(source, $event)"
              @select-upstream-mapping="selectUpstreamGroupOption(source, $event)"
              @update-manual-upstream-mapping="updateManualUpstreamGroup(source, $event)"
              @update-reference-multiplier="updateMappingReferenceMultiplier(source, $event)"
            />
          </div>
        </section>

        <div class="flex items-center justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="reload">
            {{ t('common.reset') }}
          </button>
          <button type="submit" class="btn btn-primary" :disabled="saving">
            <span v-if="saving">{{ t('common.saving') }}</span>
            <span v-else>{{ t('common.save') }}</span>
          </button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import AppLayout from "@/components/layout/AppLayout.vue";
import Icon from "@/components/icons/Icon.vue";
import Input from "@/components/common/Input.vue";
import Toggle from "@/components/common/Toggle.vue";
import UpstreamMonitorSourceCard from "./upstream-monitor/UpstreamMonitorSourceCard.vue";
import {
  useSourceMappingRows,
  type SourceMappingRowView,
  type SourceMappingSelectOption,
  type SourceMappingUpdatePayload,
} from "./upstream-monitor/useSourceMappingRows";
import { adminAPI } from "@/api/admin";
import {
  normalizeUpstreamMonitorConfig,
  type UpstreamMonitorConfig,
  type UpstreamMonitorFetchMode,
  type UpstreamMonitorGroupMapping,
  type UpstreamMonitorPreviewGroupOption,
  type UpstreamMonitorPreviewGroupRow,
  type UpstreamMonitorPreviewSnapshot,
  type UpstreamMonitorSourceConfig,
  type UpstreamMonitorUpstreamGroupOption,
} from "@/api/admin/settings";
import type { AdminGroup } from "@/types";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";
import { getSourceID } from "./upstreamMonitorMappings";

interface SourceSummaryGroup {
  key: string;
  name: string;
  upstreamGroup: string;
  modelFamily: UpstreamMonitorGroupMapping["model_family"];
  localMultiplier: number;
  referenceMultiplier: number;
  marginRate: number | null;
  status: string;
}
const { t, locale } = useI18n();
const appStore = useAppStore();

const loading = ref(true);
const saving = ref(false);
const syncing = ref(false);
const syncingSourceIds = ref<string[]>([]);
const form = ref<UpstreamMonitorConfig>(normalizeUpstreamMonitorConfig());
const expandedSourceIDs = ref<string[]>([]);
const localGroupOptions = ref<UpstreamMonitorPreviewGroupOption[]>([]);
const preview = ref<UpstreamMonitorPreviewSnapshot>({
  generated_at: "",
  summary: {
    enabled: false,
    auto_refresh_enabled: false,
    source_count: 0,
    enabled_source_count: 0,
    mapped_group_count: 0,
    monitored_account_count: 0,
    unmapped_group_count: 0,
    healthy_count: 0,
    warning_count: 0,
    critical_count: 0,
    unknown_count: 0,
    average_margin_rate: 0,
    lowest_margin_rate: 0,
    highest_margin_rate: 0,
  },
  source_rows: [],
  group_rows: [],
  account_rows: [],
  account_options: [],
  group_options: [],
  unmapped_groups: [],
});

const availableLocalGroupOptions = computed(() => mergeGroupOptions(
  localGroupOptions.value,
  preview.value.group_options,
));

const mappingRows = useSourceMappingRows({
  form,
  localGroups: availableLocalGroupOptions,
  previewGroupRows: computed(() => preview.value.group_rows),
  messages: () => ({
    noGroupOptions: t("admin.upstreamMonitor.mapping.fields.noGroupOptions"),
    mappingRowAddFailed: t("admin.upstreamMonitor.sources.fields.mappingRowAddFailed"),
    selectLocalGroup: t("admin.upstreamMonitor.sources.fields.selectLocalGroupPlaceholder"),
    selectUpstreamGroup: t("admin.upstreamMonitor.sources.fields.selectUpstreamGroupPlaceholder"),
    referenceMultiplierRequired: t("admin.upstreamMonitor.sources.fields.referenceMultiplierRequired"),
    mappingDuplicate: t("admin.upstreamMonitor.sources.fields.mappingDuplicate"),
    mappingBindFailed: t("admin.upstreamMonitor.sources.fields.mappingBindFailed"),
    mappingRemoveFailed: t("admin.upstreamMonitor.sources.fields.mappingRemoveFailed"),
  }),
  createID: uid,
  modelFamilyForGroup,
  formatMultiplier: formatDecimal,
  formatPercent,
  groupOptionLabel,
  upstreamGroupOptionLabel,
  onError: appStore.showError,
});

const sourceKindOptions = [
  { value: "manual", labelKey: "admin.upstreamMonitor.sourceKinds.manual" },
  { value: "newapi", labelKey: "admin.upstreamMonitor.sourceKinds.newapi" },
  { value: "sub2api", labelKey: "admin.upstreamMonitor.sourceKinds.sub2api" },
  { value: "openai_compatible", labelKey: "admin.upstreamMonitor.sourceKinds.openaiCompatible" },
  { value: "custom", labelKey: "admin.upstreamMonitor.sourceKinds.custom" },
] as const;

const authModeOptions = [
  { value: "none", labelKey: "admin.upstreamMonitor.authModes.none" },
  { value: "bearer", labelKey: "admin.upstreamMonitor.authModes.bearer" },
  { value: "header", labelKey: "admin.upstreamMonitor.authModes.header" },
  { value: "cookie", labelKey: "admin.upstreamMonitor.authModes.cookie" },
] as const;

const fetchModeOptions: Array<{ value: UpstreamMonitorFetchMode; labelKey: string }> = [
  { value: "auto", labelKey: "admin.upstreamMonitor.fetchModes.auto" },
  { value: "json_path", labelKey: "admin.upstreamMonitor.fetchModes.jsonPath" },
  { value: "plain_text", labelKey: "admin.upstreamMonitor.fetchModes.plainText" },
];

const refreshIntervalInput = computed({
  get: () => form.value.refresh_interval_minutes,
  set: (value: string | number) => {
    form.value.refresh_interval_minutes = Number(value || 0);
  },
});

function formatPercent(value: number): string {
  return `${(Number(value || 0) * 100).toFixed(1)}%`;
}

function formatDecimal(value: number): string {
  return Number(value || 0).toFixed(2);
}

function uid(prefix: string): string {
  return `${prefix}_${Math.random().toString(36).slice(2, 10)}`;
}

function multiplierLabel(value: number): string {
  return Number(value || 0) > 0 ? `${formatDecimal(value)}x` : "--";
}

function statusRank(status: string): number {
  switch (status) {
    case "critical":
      return 0;
    case "warning":
      return 1;
    case "unknown":
      return 2;
    case "healthy":
      return 3;
    default:
      return 4;
  }
}

function createSource(): UpstreamMonitorSourceConfig {
  return {
    id: uid("source"),
    name: "",
    kind: "manual",
    enabled: true,
    auto_sync_enabled: true,
    account_ids: [],
    fetch_mode: "auto",
    base_url: "",
    pricing_url: "",
    pricing_path_hint: "",
    auth_mode: "none",
    auth_header_name: "",
    auth_token: "",
    auth_configured: false,
    currency: "CNY",
    exchange_rate: form.value.default_exchange_rate || 7.2,
    reference_multiplier: 0,
    upstream_group_options: [],
    last_sync_at: null,
    last_sync_status: "idle",
    last_sync_error: "",
    notes: "",
  };
}

function addSource() {
  const source = createSource();
  form.value.sources.push(source);
  expandedSourceIDs.value = [...expandedSourceIDs.value, source.id];
}

function removeSource(index: number) {
  const removed = form.value.sources[index];
  form.value.sources.splice(index, 1);
  if (!removed?.id) return;
  expandedSourceIDs.value = expandedSourceIDs.value.filter((id) => id !== removed.id);
  mappingRows.replaceGroupMappings(mappingRows.removeSourceFromMappingRows(form.value.group_mappings, removed.id));
}

type SourceFieldUpdatePayload = {
  field: keyof UpstreamMonitorSourceConfig;
  value: UpstreamMonitorSourceConfig[keyof UpstreamMonitorSourceConfig];
};

function updateSourceField(source: UpstreamMonitorSourceConfig, payload: SourceFieldUpdatePayload) {
  source[payload.field] = payload.value as never;
}

function uniqueNumberIDs(values: Array<number | string>): number[] {
  return Array.from(
    new Set(
      values
        .map((id) => Number(id))
        .filter((id) => Number.isFinite(id) && id > 0),
    ),
  );
}

function toggleSourceAccount(source: UpstreamMonitorSourceConfig, accountID: number) {
  const current = new Set(uniqueNumberIDs(source.account_ids));
  if (current.has(accountID)) {
    current.delete(accountID);
  } else {
    current.add(accountID);
  }
  source.account_ids = Array.from(current);
}

function modelFamilyForGroup(group: UpstreamMonitorPreviewGroupOption): UpstreamMonitorGroupMapping["model_family"] {
  const haystack = `${group.platform || ""} ${group.group_name || ""}`.toLowerCase();
  if (haystack.includes("openai") || haystack.includes("gpt") || haystack.includes("chatgpt")) {
    return "gpt";
  }
  if (haystack.includes("anthropic") || haystack.includes("claude")) {
    return "claude";
  }
  if (haystack.includes("gemini") || haystack.includes("google")) {
    return "gemini";
  }
  if (haystack.includes("deepseek")) {
    return "deepseek";
  }
  return "mixed";
}

function sourceID(source: UpstreamMonitorSourceConfig): string {
  return getSourceID(source);
}

function normalizeAdminGroupsForMonitor(groups: AdminGroup[]): UpstreamMonitorPreviewGroupOption[] {
  const options = groups
    .filter((group) => group.status === "active")
    .map((group) => ({
      group_id: Number(group.id || 0),
      group_name: String(group.name || ""),
      platform: String(group.platform || ""),
      multiplier: Number(group.rate_multiplier || 0),
      is_exclusive: Boolean(group.is_exclusive),
      subscription_type: String(group.subscription_type || ""),
    }))
    .filter((group) => group.group_id > 0 && group.group_name.trim());
  return mergeGroupOptions(options);
}

function mergeGroupOptions(...groups: UpstreamMonitorPreviewGroupOption[][]): UpstreamMonitorPreviewGroupOption[] {
  const byID = new Map<number, UpstreamMonitorPreviewGroupOption>();
  const byName = new Map<string, UpstreamMonitorPreviewGroupOption>();
  for (const list of groups) {
    for (const raw of Array.isArray(list) ? list : []) {
      const option: UpstreamMonitorPreviewGroupOption = {
        group_id: Number(raw.group_id || 0),
        group_name: String(raw.group_name || "").trim(),
        platform: String(raw.platform || ""),
        multiplier: Number(raw.multiplier || 0),
        is_exclusive: Boolean(raw.is_exclusive),
        subscription_type: String(raw.subscription_type || ""),
      };
      if (!option.group_name) continue;
      if (option.group_id > 0) {
        byID.set(option.group_id, option);
        continue;
      }
      byName.set(option.group_name.toLowerCase(), option);
    }
  }
  for (const option of byName.values()) {
    if (!Array.from(byID.values()).some((item) => item.group_name.trim().toLowerCase() === option.group_name.trim().toLowerCase())) {
      byID.set(-byID.size - 1, option);
    }
  }
  return Array.from(byID.values()).sort((left, right) => {
    const leftPlatform = left.platform.toLowerCase();
    const rightPlatform = right.platform.toLowerCase();
    if (leftPlatform !== rightPlatform) {
      return leftPlatform.localeCompare(rightPlatform);
    }
    const nameSort = left.group_name.localeCompare(right.group_name, locale.value === "zh" ? "zh-CN" : "en-US");
    return nameSort !== 0 ? nameSort : left.group_id - right.group_id;
  });
}

function isSourceSyncing(source: UpstreamMonitorSourceConfig): boolean {
  const sid = sourceID(source);
  return Boolean(sid) && syncingSourceIds.value.includes(sid);
}

function canSyncSource(source: UpstreamMonitorSourceConfig): boolean {
  return Boolean(sourceID(source) && source.pricing_url.trim()) && !saving.value && !syncing.value && !isSourceSyncing(source);
}

function setSourceSyncing(sourceID: string, value: boolean) {
  const current = new Set(syncingSourceIds.value);
  if (value) {
    current.add(sourceID);
  } else {
    current.delete(sourceID);
  }
  syncingSourceIds.value = Array.from(current);
}

function sourceMappedGroupsCount(source: UpstreamMonitorSourceConfig): number {
  return mappingRows.sourceMappedMappings(source).length;
}

function sourceMappedMappings(source: UpstreamMonitorSourceConfig): UpstreamMonitorGroupMapping[] {
  return mappingRows.sourceMappedMappings(source);
}

function sourceMappingRowsForTemplate(source: UpstreamMonitorSourceConfig): UpstreamMonitorGroupMapping[] {
  return mappingRows.sourceMappingRows(source);
}

function sourceMappingRowViews(source: UpstreamMonitorSourceConfig): SourceMappingRowView[] {
  return mappingRows.sourceMappingRowViews(source);
}

function upstreamGroupOptionsForSource(source: UpstreamMonitorSourceConfig): UpstreamMonitorUpstreamGroupOption[] {
  return mappingRows.upstreamGroupOptionsForSource(source);
}

function upstreamGroupSelectOptionsForSource(source: UpstreamMonitorSourceConfig): SourceMappingSelectOption[] {
  return mappingRows.upstreamGroupSelectOptionsForSource(source);
}

function upstreamGroupOptionsPreview(source: UpstreamMonitorSourceConfig): UpstreamMonitorUpstreamGroupOption[] {
  return upstreamGroupOptionsForSource(source).slice(0, 8);
}

function hiddenUpstreamGroupOptionCount(source: UpstreamMonitorSourceConfig): number {
  return Math.max(0, upstreamGroupOptionsForSource(source).length - upstreamGroupOptionsPreview(source).length);
}

function normalizeFormConfig(config?: Partial<UpstreamMonitorConfig> | null): UpstreamMonitorConfig {
  return mappingRows.hydrateMappingReferenceMultipliers(normalizeUpstreamMonitorConfig(config));
}

function setFormConfig(config?: Partial<UpstreamMonitorConfig> | null) {
  form.value = normalizeFormConfig(config);
}

function selectUpstreamGroupOption(
  source: UpstreamMonitorSourceConfig,
  payload: SourceMappingUpdatePayload,
) {
  mappingRows.selectUpstreamGroup(source, payload.mapping, payload.value);
}

function updateMappingLocalGroup(
  source: UpstreamMonitorSourceConfig,
  payload: SourceMappingUpdatePayload,
) {
  mappingRows.updateLocalGroup(source, payload.mapping, payload.value);
}

function updateManualUpstreamGroup(
  source: UpstreamMonitorSourceConfig,
  payload: SourceMappingUpdatePayload,
) {
  mappingRows.updateManualUpstreamGroup(source, payload.mapping, payload.value);
}

function updateMappingReferenceMultiplier(
  source: UpstreamMonitorSourceConfig,
  payload: SourceMappingUpdatePayload,
) {
  mappingRows.updateReferenceMultiplier(source, payload.mapping, payload.value);
}

function ensureSourceID(source: UpstreamMonitorSourceConfig): string {
  const currentID = sourceID(source);
  if (currentID) {
    return currentID;
  }
  const nextID = uid("source");
  source.id = nextID;
  return nextID;
}

function isSourceExpanded(source: UpstreamMonitorSourceConfig): boolean {
  const sid = sourceID(source);
  return Boolean(sid && expandedSourceIDs.value.includes(sid));
}

function toggleSourceExpanded(source: UpstreamMonitorSourceConfig) {
  const sid = ensureSourceID(source);
  const current = new Set(expandedSourceIDs.value);
  if (current.has(sid)) {
    current.delete(sid);
  } else {
    current.add(sid);
  }
  expandedSourceIDs.value = Array.from(current);
}

function normalizeSourceBaseURL(raw: string): string {
  const trimmed = String(raw || "").trim();
  if (!trimmed) return "";
  const withScheme = trimmed.includes("://") ? trimmed : `https://${trimmed}`;
  return withScheme.replace(/\/+$/, "");
}

function sourceBaseCandidate(source: UpstreamMonitorSourceConfig): string {
  const raw = String(source.base_url || source.pricing_url || "").trim();
  if (!raw) return "";
  const normalized = normalizeSourceBaseURL(raw);
  try {
    const parsed = new URL(normalized);
    return parsed.origin;
  } catch {
    return normalized;
  }
}

function defaultPricingURLForSource(source: UpstreamMonitorSourceConfig): string {
  const base = sourceBaseCandidate(source);
  if (!base) return "";
  if (source.kind === "sub2api") {
    return `${base}/api/v1/channels/available`;
  }
  if (source.kind === "newapi") {
    return `${base}/api/ratio_config`;
  }
  return "";
}

function applySourcePreset(source: UpstreamMonitorSourceConfig) {
  const pricingURL = defaultPricingURLForSource(source);
  if (!pricingURL) {
    appStore.showError(t("admin.upstreamMonitor.sources.applyPresetMissingBaseUrl"));
    return;
  }
  if (!source.base_url.trim()) {
    source.base_url = sourceBaseCandidate(source);
  }
  source.pricing_url = pricingURL;
  source.fetch_mode = "auto";
  source.pricing_path_hint = "";
  if (source.kind === "sub2api") {
    source.auth_mode = "bearer";
  }
  appStore.showSuccess(t("admin.upstreamMonitor.sources.applyPresetSuccess"));
}

async function addSourceGroupMapping(source: UpstreamMonitorSourceConfig) {
  const result = await mappingRows.addMapping(source);
  if (!result.ok || !result.mapping) {
    return;
  }
  appStore.showSuccess(t("admin.upstreamMonitor.sources.fields.mappingRowAdded"));
  document
    .getElementById(mappingRows.sourceMappingRowElementID(source, result.mapping))
    ?.scrollIntoView({ behavior: "smooth", block: "center" });
}

function bindSourceGroupMapping(source: UpstreamMonitorSourceConfig, mapping: UpstreamMonitorGroupMapping) {
  const result = mappingRows.bindMapping(source, mapping);
  if (result.ok) {
    appStore.showSuccess(t("admin.upstreamMonitor.sources.fields.mappingBound"));
  }
}

function removeSourceGroupMapping(source: UpstreamMonitorSourceConfig, mapping: UpstreamMonitorGroupMapping) {
  const result = mappingRows.removeMapping(source, mapping);
  if (result.ok) {
    appStore.showSuccess(t("admin.upstreamMonitor.sources.fields.mappingRemoved"));
  }
}

function effectiveReferenceMultiplier(source: UpstreamMonitorSourceConfig, mapping: UpstreamMonitorGroupMapping): number {
  return mappingRows.effectiveReferenceMultiplier(source, mapping);
}

function sourceGroupRowForMapping(source: UpstreamMonitorSourceConfig, mapping: UpstreamMonitorGroupMapping): UpstreamMonitorPreviewGroupRow | undefined {
  return mappingRows.sourceGroupRowForMapping(source, mapping);
}

function estimatedMarginRate(localMultiplier: number, referenceMultiplier: number): number | null {
  return mappingRows.estimatedMarginRate(localMultiplier, referenceMultiplier);
}

function marginStatus(localMultiplier: number, referenceMultiplier: number): string {
  const marginRate = estimatedMarginRate(localMultiplier, referenceMultiplier);
  if (marginRate === null) {
    return "unknown";
  }
  if (marginRate <= Number(form.value.critical_rate_threshold || 0)) {
    return "critical";
  }
  if (marginRate <= Number(form.value.warning_rate_threshold || 0)) {
    return "warning";
  }
  return "healthy";
}

function sourceSummaryGroups(source: UpstreamMonitorSourceConfig): SourceSummaryGroup[] {
  const sourceReferenceMultiplier = Number(source.reference_multiplier || 0);
  const groups: SourceSummaryGroup[] = [];
  for (const mapping of sourceMappedMappings(source)) {
    const name = mapping.local_group.trim();
    if (!name) continue;

    const row = sourceGroupRowForMapping(source, mapping);
    const option = mappingRows.groupOptionForMapping(mapping);
    const localMultiplier = Number(option?.multiplier || row?.local_group_multiplier || 0);
    const referenceMultiplier = Number(row?.reference_multiplier || effectiveReferenceMultiplier(source, mapping) || sourceReferenceMultiplier);
    const marginRate = estimatedMarginRate(localMultiplier, referenceMultiplier);
    const status = localMultiplier > 0 && referenceMultiplier > 0
      ? marginStatus(localMultiplier, referenceMultiplier)
      : (row?.status || "unknown");

    groups.push({
      key: mappingRows.sourceMappingRowElementID(source, mapping),
      name,
      upstreamGroup: mapping.upstream_group || name,
      modelFamily: mappingRows.normalizeModelFamily(String(row?.model_family || mapping.model_family || "mixed")),
      localMultiplier,
      referenceMultiplier,
      marginRate,
      status,
    });
  }

  return groups.sort((left, right) => {
    const statusDiff = statusRank(left.status) - statusRank(right.status);
    if (statusDiff !== 0) return statusDiff;
    const leftMargin = left.marginRate ?? Number.POSITIVE_INFINITY;
    const rightMargin = right.marginRate ?? Number.POSITIVE_INFINITY;
    if (leftMargin !== rightMargin) return leftMargin - rightMargin;
    return left.name.localeCompare(right.name, locale.value === "zh" ? "zh-CN" : "en-US");
  });
}

function visibleSourceSummaryGroups(source: UpstreamMonitorSourceConfig): SourceSummaryGroup[] {
  return sourceSummaryGroups(source).slice(0, 4);
}

function hiddenSourceSummaryGroupCount(source: UpstreamMonitorSourceConfig): number {
  return Math.max(0, sourceSummaryGroups(source).length - visibleSourceSummaryGroups(source).length);
}

function sourceWorstStatus(source: UpstreamMonitorSourceConfig): string {
  const groups = sourceSummaryGroups(source);
  return groups.length > 0 ? groups[0].status : "unknown";
}

function sourceLowestMarginLabel(source: UpstreamMonitorSourceConfig): string {
  const margins = sourceSummaryGroups(source)
    .map((group) => group.marginRate)
    .filter((value): value is number => value !== null);
  if (margins.length === 0) {
    return "--";
  }
  return formatPercent(Math.min(...margins));
}

function groupOptionsForSource(source: UpstreamMonitorSourceConfig): UpstreamMonitorPreviewGroupOption[] {
  return mappingRows.groupOptionsForSource(source);
}

function localGroupSelectOptionsForSource(source: UpstreamMonitorSourceConfig): SourceMappingSelectOption[] {
  return mappingRows.localGroupSelectOptionsForSource(source);
}

async function loadLocalGroupOptions() {
  const groups = await adminAPI.groups.getAll();
  localGroupOptions.value = normalizeAdminGroupsForMonitor(groups);
}

function sanitize(): UpstreamMonitorConfig {
  return sanitizeConfig({ dropIncompleteMappings: true });
}

function sanitizeForPreview(): UpstreamMonitorConfig {
  return sanitizeConfig({ dropIncompleteMappings: true });
}

function sanitizeConfig(options: { dropIncompleteMappings: boolean }): UpstreamMonitorConfig {
  const groupMappings = mappingRows.serializeMappings(options);

  return normalizeFormConfig({
    ...form.value,
    sources: form.value.sources.map((source) => ({
      ...source,
      id: (source.id || uid("source")).trim(),
      name: source.name.trim(),
      base_url: source.base_url.trim(),
      pricing_url: source.pricing_url.trim(),
      pricing_path_hint: source.pricing_path_hint.trim(),
      auth_header_name: source.auth_header_name.trim(),
      auth_token: (source.auth_token || "").trim(),
      account_ids: Array.isArray(source.account_ids) ? uniqueNumberIDs(source.account_ids) : [],
      last_sync_at: source.last_sync_at || null,
      last_sync_status: source.last_sync_status || "idle",
      last_sync_error: source.last_sync_error.trim(),
      notes: source.notes.trim(),
    })),
    group_mappings: groupMappings,
  });
}

function groupOptionLabel(group: UpstreamMonitorPreviewGroupOption): string {
  const tags = [group.platform || "--", formatDecimal(group.multiplier)];
  if (group.is_exclusive) {
    tags.push(t("admin.upstreamMonitor.preview.exclusive"));
  }
  if (group.subscription_type) {
    tags.push(group.subscription_type);
  }
  return `${group.group_name || `#${group.group_id}`} · ${tags.join(" · ")}`;
}

function upstreamGroupOptionLabel(option: UpstreamMonitorUpstreamGroupOption): string {
  const tags = [multiplierLabel(Number(option.reference_multiplier || 0))];
  if (option.description) {
    tags.push(option.description);
  }
  if (option.raw_id) {
    tags.push(`ID ${option.raw_id}`);
  } else if (option.path) {
    tags.push(t("admin.upstreamMonitor.sources.fields.pathFallback"));
    tags.push(option.path);
  } else if (option.key) {
    tags.push(option.key);
  }
  return `${option.name || option.key} · ${tags.join(" · ")}`;
}

async function refreshPreview() {
  try {
    preview.value = await adminAPI.settings.previewUpstreamMonitorConfig(sanitizeForPreview());
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t("admin.upstreamMonitor.preview.failed")));
  }
}

async function syncSources() {
  syncing.value = true;
  try {
    const updated = await adminAPI.settings.updateUpstreamMonitorConfig(sanitize());
    setFormConfig(updated);
    const result = await adminAPI.settings.refreshUpstreamMonitorConfig();
    setFormConfig(result.config);
    if (result.snapshot) {
      preview.value = result.snapshot;
    } else {
      await refreshPreview();
    }
    appStore.showSuccess(
      t("admin.upstreamMonitor.refreshed", {
        success: result.summary.success_count,
        failed: result.summary.failed_count,
        attempted: result.summary.attempted_count,
      }),
    );
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t("admin.upstreamMonitor.failedToRefresh")));
  } finally {
    syncing.value = false;
  }
}

async function syncSource(source: UpstreamMonitorSourceConfig) {
  const sid = sourceID(source);
  if (!sid || !source.pricing_url.trim()) {
    appStore.showError(t("admin.upstreamMonitor.sources.pullGroupsMissingUrl"));
    return;
  }

  setSourceSyncing(sid, true);
  try {
    const updated = await adminAPI.settings.updateUpstreamMonitorConfig(sanitize());
    setFormConfig(updated);
    const result = await adminAPI.settings.refreshUpstreamMonitorConfig(sid);
    setFormConfig(result.config);
    if (result.snapshot) {
      preview.value = result.snapshot;
    } else {
      await refreshPreview();
    }
    const refreshedSource = form.value.sources.find((item) => sourceID(item) === sid);
    appStore.showSuccess(
      t("admin.upstreamMonitor.sources.pullGroupsSuccess", {
        count: refreshedSource?.upstream_group_options.length || 0,
      }),
    );
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t("admin.upstreamMonitor.failedToRefresh")));
  } finally {
    setSourceSyncing(sid, false);
  }
}

async function reload() {
  loading.value = true;
  try {
    const [config] = await Promise.all([
      adminAPI.settings.getUpstreamMonitorConfig(),
      loadLocalGroupOptions(),
    ]);
    setFormConfig(config);
    expandedSourceIDs.value = expandedSourceIDs.value.filter((id) =>
      form.value.sources.some((source) => sourceID(source) === id),
    );
    loading.value = false;
    void refreshPreview();
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t("admin.upstreamMonitor.failedToLoad")));
    loading.value = false;
  }
}

async function save() {
  saving.value = true;
  try {
    const updated = await adminAPI.settings.updateUpstreamMonitorConfig(sanitize());
    setFormConfig(updated);
    expandedSourceIDs.value = expandedSourceIDs.value.filter((id) =>
      form.value.sources.some((source) => sourceID(source) === id),
    );
    appStore.showSuccess(t("admin.upstreamMonitor.saved"));
    void refreshPreview();
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t("admin.upstreamMonitor.failedToSave")));
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  reload();
});

</script>
