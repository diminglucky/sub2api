<template>
  <div class="source-summary-panel">
    <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
      <div class="source-summary-metric">
        <div class="metric-label">{{ t('admin.upstreamMonitor.sources.summary.mappedGroups') }}</div>
        <div class="metric-value">{{ mappedGroupsCount }}</div>
      </div>
      <div class="source-summary-metric">
        <div class="metric-label">{{ t('admin.upstreamMonitor.sources.summary.accounts') }}</div>
        <div class="metric-value">{{ accountCount }}</div>
      </div>
      <div class="source-summary-metric">
        <div class="metric-label">{{ t('admin.upstreamMonitor.sources.summary.upstreamMultiplier') }}</div>
        <div class="metric-value">{{ sourceMultiplierLabel(referenceMultiplier) }}</div>
      </div>
      <div class="source-summary-metric">
        <div class="metric-label">{{ t('admin.upstreamMonitor.sources.summary.lowestMargin') }}</div>
        <div class="metric-value" :class="sourceSummaryStatusTextClass(worstStatus)">
          {{ lowestMarginLabel }}
        </div>
      </div>
    </div>

    <div
      v-if="summaryGroups.length === 0"
      class="mt-4 rounded-xl border border-dashed border-gray-200 px-4 py-5 text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
    >
      {{ t('admin.upstreamMonitor.sources.summary.noMappedGroups') }}
    </div>
    <div v-else class="mt-4 space-y-2">
      <div
        v-for="group in summaryGroups"
        :key="group.key"
        class="source-summary-group-row"
      >
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <span class="truncate text-sm font-semibold text-gray-900 dark:text-white">
              {{ group.name }}
            </span>
            <span class="status-pill" :class="sourceStatusClass(group.status)">
              {{ sourceStatusLabel(t, group.status) }}
            </span>
            <span class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] text-gray-600 dark:bg-dark-700 dark:text-gray-300">
              {{ t(`admin.upstreamMonitor.modelFamilies.${group.modelFamily}`) }}
            </span>
          </div>
          <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t('admin.upstreamMonitor.sources.summary.upstreamGroup') }} {{ group.upstreamGroup || group.name }}
          </div>
        </div>
        <div class="source-summary-values">
          <span>{{ t('admin.upstreamMonitor.sources.summary.localMultiplier') }} {{ sourceMultiplierLabel(group.localMultiplier) }}</span>
          <span>{{ t('admin.upstreamMonitor.sources.summary.upstreamMultiplier') }} {{ sourceMultiplierLabel(group.referenceMultiplier) }}</span>
          <span>{{ t('admin.upstreamMonitor.sources.summary.spreadMultiplier') }} {{ sourceMultiplierSpreadLabel(group.localMultiplier, group.referenceMultiplier) }}</span>
          <span :class="sourceSummaryStatusTextClass(group.status)">
            {{ t('admin.upstreamMonitor.sources.summary.marginRate') }} {{ sourceNullablePercentLabel(group.marginRate) }}
          </span>
        </div>
      </div>
      <div
        v-if="hiddenSummaryGroupCount > 0"
        class="rounded-xl border border-dashed border-gray-200 px-4 py-2 text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
      >
        {{ t('admin.upstreamMonitor.sources.summary.moreGroups', { count: hiddenSummaryGroupCount }) }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import {
  sourceMultiplierLabel,
  sourceMultiplierSpreadLabel,
  sourceNullablePercentLabel,
  sourceStatusClass,
  sourceStatusLabel,
  sourceSummaryStatusTextClass,
} from "./sourceCardFormat";
import type { SourceSummaryGroupView } from "./sourceCardTypes";

defineProps<{
  mappedGroupsCount: number;
  accountCount: number;
  referenceMultiplier: number;
  summaryGroups: SourceSummaryGroupView[];
  hiddenSummaryGroupCount: number;
  worstStatus: string;
  lowestMarginLabel: string;
}>();

const { t } = useI18n();
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

.metric-label {
  @apply text-xs text-gray-500 dark:text-gray-400;
}

.metric-value {
  @apply mt-1 text-base font-semibold text-gray-900 dark:text-white;
}

.source-summary-panel {
  @apply rounded-2xl border border-gray-200 bg-gray-50/70 p-4 dark:border-dark-600 dark:bg-dark-900/30;
}

.source-summary-metric {
  @apply rounded-xl border border-gray-200 bg-white px-3 py-3 dark:border-dark-600 dark:bg-dark-800/80;
}

.source-summary-group-row {
  @apply flex flex-col gap-3 rounded-xl border border-gray-200 bg-white px-4 py-3 dark:border-dark-600 dark:bg-dark-800/80 lg:flex-row lg:items-center lg:justify-between;
}

.source-summary-values {
  @apply flex shrink-0 flex-wrap gap-x-4 gap-y-1 text-xs font-medium text-gray-500 dark:text-gray-400;
}
</style>
