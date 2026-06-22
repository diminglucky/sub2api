<template>
  <div class="mt-4">
    <div class="mb-3 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <label class="input-label">{{ t('admin.upstreamMonitor.sources.fields.localGroupMappings') }}</label>
        <p class="input-hint mt-1.5">{{ t('admin.upstreamMonitor.sources.fields.localGroupMappingsHint') }}</p>
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <span class="text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.upstreamMonitor.sources.fields.mappingRowsCount', { bound: boundCount, total: totalCount }) }}
        </span>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="!canAdd" @click="emit('add')">
          <Icon name="plus" size="sm" />
          <span>{{ t('admin.upstreamMonitor.sources.fields.addMapping') }}</span>
        </button>
      </div>
    </div>

    <div
      v-if="localGroupOptions.length === 0"
      class="rounded-xl border border-dashed border-gray-200 px-4 py-6 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
    >
      {{ t('admin.upstreamMonitor.mapping.fields.noGroupOptions') }}
    </div>

    <div v-else class="space-y-3">
      <div
        v-if="rows.length === 0"
        class="rounded-xl border border-dashed border-gray-200 px-4 py-6 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
      >
        {{ t('admin.upstreamMonitor.sources.fields.noMappingRows') }}
      </div>

      <div
        v-for="(row, mappingIndex) in rows"
        :id="row.elementId"
        :key="row.key"
        class="mapping-bind-panel"
        :class="{ 'mapping-bind-panel-new': row.isNew }"
      >
        <div class="mb-4 flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('admin.upstreamMonitor.sources.fields.mappingRowTitle', { index: mappingIndex + 1 }) }}
            </div>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ row.summary }}
            </p>
          </div>
          <button
            type="button"
            class="inline-flex h-8 items-center gap-1.5 rounded-lg border border-gray-200 px-2.5 text-xs font-medium text-gray-500 transition hover:border-red-200 hover:text-red-500 dark:border-dark-600 dark:text-gray-400 dark:hover:border-red-800/40 dark:hover:text-red-300"
            @click="emit('remove', row.mapping)"
          >
            <Icon name="trash" size="xs" />
            <span>{{ t('common.delete') }}</span>
          </button>
        </div>

        <div class="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] lg:items-end">
          <div>
            <label class="input-label">{{ t('admin.upstreamMonitor.sources.fields.selectLocalGroup') }}</label>
            <select
              class="input mt-1.5 w-full"
              :value="row.localGroupID"
              @change="emit('update-local', { mapping: row.mapping, value: ($event.target as HTMLSelectElement).value })"
            >
              <option value="">{{ t('admin.upstreamMonitor.sources.fields.selectLocalGroupPlaceholder') }}</option>
              <option
                v-for="group in localGroupOptions"
                :key="`${sourceId}_local_${row.mapping.id}_${group.value}`"
                :value="group.value"
              >
                {{ group.label }}
              </option>
            </select>
            <p class="input-hint mt-1.5">{{ t('admin.upstreamMonitor.sources.fields.localGroupSelectHint') }}</p>
          </div>

          <div>
            <label class="input-label">{{ t('admin.upstreamMonitor.sources.fields.selectUpstreamGroup') }}</label>
            <select
              class="input mt-1.5 w-full"
              :value="row.mapping.upstream_group_key"
              :disabled="upstreamGroupOptions.length === 0"
              @change="emit('select-upstream', { mapping: row.mapping, value: ($event.target as HTMLSelectElement).value })"
            >
              <option value="">
                {{
                  upstreamGroupOptions.length > 0
                    ? t('admin.upstreamMonitor.sources.fields.selectUpstreamGroupPlaceholder')
                    : t('admin.upstreamMonitor.sources.fields.pullUpstreamGroupsFirst')
                }}
              </option>
              <option
                v-for="option in upstreamGroupOptions"
                :key="`${sourceId}_upstream_${row.mapping.id}_${option.value}`"
                :value="option.value"
              >
                {{ option.label }}
              </option>
            </select>
            <p class="input-hint mt-1.5">
              {{
                upstreamGroupOptions.length > 0
                  ? t('admin.upstreamMonitor.sources.fields.upstreamGroupSelectHint')
                  : t('admin.upstreamMonitor.sources.fields.upstreamGroupEmptyHint')
              }}
            </p>
          </div>

          <button type="button" class="btn btn-primary" @click="emit('bind', row.mapping)">
            {{ row.isComplete ? t('admin.upstreamMonitor.sources.fields.updateMapping') : t('admin.upstreamMonitor.sources.fields.bindMapping') }}
          </button>
        </div>

        <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.upstreamMonitor.sources.fields.manualUpstreamGroup') }}</label>
            <input
              class="input mt-1.5 w-full"
              type="text"
              :value="row.mapping.upstream_group"
              :placeholder="t('admin.upstreamMonitor.sources.fields.upstreamGroup')"
              @input="emit('update-manual-upstream', { mapping: row.mapping, value: ($event.target as HTMLInputElement).value })"
            />
            <p class="input-hint mt-1.5">{{ t('admin.upstreamMonitor.sources.fields.upstreamGroupHint') }}</p>
          </div>

          <div>
            <label class="input-label">{{ t('admin.upstreamMonitor.sources.fields.mappingReferenceMultiplier') }}</label>
            <input
              class="input mt-1.5 w-full"
              type="number"
              min="0"
              step="any"
              :value="row.mapping.reference_multiplier || ''"
              @input="emit('update-reference-multiplier', { mapping: row.mapping, value: ($event.target as HTMLInputElement).value })"
            />
            <p class="input-hint mt-1.5">
              {{
                row.mapping.upstream_group_key
                  ? t('admin.upstreamMonitor.sources.fields.mappingReferenceMultiplierAutoHint')
                  : t('admin.upstreamMonitor.sources.fields.mappingReferenceMultiplierHint')
              }}
            </p>
          </div>
        </div>

        <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-3">
          <div class="rounded-xl border border-gray-200 bg-white px-3 py-3 dark:border-dark-600 dark:bg-dark-800/80">
            <div class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.upstreamMonitor.sources.mappingDetails.localMultiplier') }}
            </div>
            <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
              {{ row.localMultiplierLabel }}
            </div>
          </div>
          <div class="rounded-xl border border-gray-200 bg-white px-3 py-3 dark:border-dark-600 dark:bg-dark-800/80">
            <div class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.upstreamMonitor.sources.mappingDetails.effectiveUpstreamMultiplier') }}
            </div>
            <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
              {{ row.referenceMultiplierLabel }}
            </div>
          </div>
          <div class="rounded-xl border border-gray-200 bg-white px-3 py-3 dark:border-dark-600 dark:bg-dark-800/80">
            <div class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.upstreamMonitor.sources.summary.marginRate') }}
            </div>
            <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
              {{ row.marginRateLabel }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import Icon from "@/components/icons/Icon.vue";
import type { UpstreamMonitorGroupMapping } from "@/api/admin/settings";
import type { SourceMappingUpdatePayload } from "./useSourceMappingRows";

interface SelectOptionView {
  value: string;
  label: string;
}

interface MappingRowView {
  key: string;
  elementId: string;
  mapping: UpstreamMonitorGroupMapping;
  summary: string;
  localGroupID: string;
  isComplete: boolean;
  isNew: boolean;
  localMultiplierLabel: string;
  referenceMultiplierLabel: string;
  marginRateLabel: string;
}

defineProps<{
  sourceId: string;
  boundCount: number;
  totalCount: number;
  canAdd: boolean;
  rows: MappingRowView[];
  localGroupOptions: SelectOptionView[];
  upstreamGroupOptions: SelectOptionView[];
}>();

const emit = defineEmits<{
  (event: "add"): void;
  (event: "remove", mapping: UpstreamMonitorGroupMapping): void;
  (event: "bind", mapping: UpstreamMonitorGroupMapping): void;
  (event: "update-local", payload: SourceMappingUpdatePayload): void;
  (event: "select-upstream", payload: SourceMappingUpdatePayload): void;
  (event: "update-manual-upstream", payload: SourceMappingUpdatePayload): void;
  (event: "update-reference-multiplier", payload: SourceMappingUpdatePayload): void;
}>();

const { t } = useI18n();
</script>

<style scoped>
.mapping-bind-panel {
  @apply rounded-xl border border-gray-200 bg-gray-50/60 p-4 dark:border-dark-600 dark:bg-dark-900/30;
}

.mapping-bind-panel-new {
  @apply border-primary-300 bg-primary-50/70 shadow-sm ring-2 ring-primary-200/70 dark:border-primary-500/60 dark:bg-primary-950/30 dark:ring-primary-500/20;
}
</style>
