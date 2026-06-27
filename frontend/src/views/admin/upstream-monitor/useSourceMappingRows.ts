import { nextTick, ref, type ComputedRef, type Ref } from "vue";
import type {
  UpstreamMonitorConfig,
  UpstreamMonitorGroupMapping,
  UpstreamMonitorPreviewGroupOption,
  UpstreamMonitorPreviewGroupRow,
  UpstreamMonitorSourceConfig,
  UpstreamMonitorUpstreamGroupOption,
} from "@/api/admin/settings";
import {
  addSourceMappingRow,
  buildSourceScopedMappingID,
  getSourceID,
  getSourceMappingRowElementID,
  getSourceMappingRows,
  getSourceMappingRowsBySourceID,
  hasDuplicateSourceMappingRow,
  removeSourceFromMappingRows,
  removeSourceMappingRow,
  replaceSourceMappingRowForSource,
  uniqueStringIDs,
} from "../upstreamMonitorMappings";

export interface SourceMappingSelectOption {
  value: string;
  label: string;
}

export interface SourceMappingUpdatePayload {
  mapping: UpstreamMonitorGroupMapping;
  value: string;
}

export interface SourceMappingRowView {
  key: string;
  elementId: string;
  mapping: UpstreamMonitorGroupMapping;
  summary: string;
  localGroupID: string;
  isComplete: boolean;
  isNew: boolean;
}

export interface SourceMappingControllerMessages {
  noGroupOptions: string;
  mappingRowAddFailed: string;
  selectLocalGroup: string;
  selectUpstreamGroup: string;
  referenceMultiplierRequired: string;
  mappingDuplicate: string;
  mappingBindFailed: string;
  mappingRemoveFailed: string;
}

export interface SourceMappingControllerOptions {
  form: Ref<UpstreamMonitorConfig>;
  localGroups: Ref<UpstreamMonitorPreviewGroupOption[]> | ComputedRef<UpstreamMonitorPreviewGroupOption[]>;
  previewGroupRows: Ref<UpstreamMonitorPreviewGroupRow[]> | ComputedRef<UpstreamMonitorPreviewGroupRow[]>;
  messages: () => SourceMappingControllerMessages;
  createID: (prefix: string) => string;
  modelFamilyForGroup: (group: UpstreamMonitorPreviewGroupOption) => UpstreamMonitorGroupMapping["model_family"];
  formatMultiplier: (value: number) => string;
  formatPercent: (value: number) => string;
  groupOptionLabel: (group: UpstreamMonitorPreviewGroupOption) => string;
  upstreamGroupOptionLabel: (option: UpstreamMonitorUpstreamGroupOption) => string;
  onError: (message: string) => void;
}

export type SourceMappingControllerResult =
  | { ok: true; mapping?: UpstreamMonitorGroupMapping }
  | { ok: false; reason: string };

function sanitizeSourceID(source: UpstreamMonitorSourceConfig): string {
  return getSourceID(source);
}

function normalizeModelFamily(value: string): UpstreamMonitorGroupMapping["model_family"] {
  if (value === "gpt" || value === "claude" || value === "gemini" || value === "deepseek" || value === "mixed") {
    return value;
  }
  return "mixed";
}

function multiplierValueLabel(formatMultiplier: (value: number) => string, value: number): string {
  return Number(value || 0) > 0 ? `${formatMultiplier(value)}x` : "--";
}

function effectiveSourceCostMultiplier(source: UpstreamMonitorSourceConfig, referenceMultiplier: number): number {
  const multiplier = Number(referenceMultiplier || 0);
  if (multiplier <= 0) return 0;
  const currency = String(source.currency || "").trim().toUpperCase();
  if (!currency || currency === "CNY") {
    return multiplier;
  }
  const exchangeRate = Number(source.exchange_rate || 0);
  return multiplier * (exchangeRate > 0 ? exchangeRate : 1);
}

export function useSourceMappingRows(options: SourceMappingControllerOptions) {
  const lastAddedMappingID = ref("");

  function replaceGroupMappings(nextMappings: UpstreamMonitorGroupMapping[]) {
    options.form.value = {
      ...options.form.value,
      group_mappings: nextMappings,
    };
  }

  function ensureSourceID(source: UpstreamMonitorSourceConfig): string {
    const currentID = sanitizeSourceID(source);
    if (currentID) {
      return currentID;
    }
    const nextID = options.createID("source");
    source.id = nextID;
    return nextID;
  }

  function groupOptionByName(groupName: string): UpstreamMonitorPreviewGroupOption | undefined {
    const cleanName = groupName.trim();
    if (!cleanName) return undefined;
    return options.localGroups.value.find((group) => group.group_name.trim() === cleanName);
  }

  function groupOptionByID(groupID: number): UpstreamMonitorPreviewGroupOption | undefined {
    if (!Number.isFinite(groupID) || groupID <= 0) return undefined;
    return options.localGroups.value.find((group) => group.group_id === groupID);
  }

  function groupOptionForMapping(mapping: UpstreamMonitorGroupMapping): UpstreamMonitorPreviewGroupOption | undefined {
    if (mapping.local_group_id > 0) {
      const byID = groupOptionByID(mapping.local_group_id);
      if (byID) return byID;
    }
    return groupOptionByName(mapping.local_group);
  }

  function localGroupIDForMapping(mapping: UpstreamMonitorGroupMapping): number {
    if (mapping.local_group_id > 0) {
      return Number(mapping.local_group_id || 0);
    }
    return Number(groupOptionByName(mapping.local_group)?.group_id || 0);
  }

  function localMultiplierForMapping(mapping: UpstreamMonitorGroupMapping): number {
    const group = groupOptionForMapping(mapping);
    return Number(group?.multiplier || 0);
  }

  function upstreamGroupOptionsForSource(source: UpstreamMonitorSourceConfig): UpstreamMonitorUpstreamGroupOption[] {
    const sourceOptions = Array.isArray(source.upstream_group_options) ? [...source.upstream_group_options] : [];
    return sourceOptions.sort((left, right) => {
      const nameSort = left.name.localeCompare(right.name);
      return nameSort !== 0 ? nameSort : left.key.localeCompare(right.key);
    });
  }

  function findUpstreamGroupOption(
    source: UpstreamMonitorSourceConfig,
    key: string,
  ): UpstreamMonitorUpstreamGroupOption | undefined {
    const cleanKey = key.trim();
    if (!cleanKey) return undefined;
    return upstreamGroupOptionsForSource(source).find((option) => option.key === cleanKey);
  }

  function selectedUpstreamGroupOption(
    source: UpstreamMonitorSourceConfig,
    mapping: UpstreamMonitorGroupMapping,
  ): UpstreamMonitorUpstreamGroupOption | undefined {
    return findUpstreamGroupOption(source, mapping.upstream_group_key || "");
  }

  function selectedUpstreamGroupReferenceMultiplier(
    source: UpstreamMonitorSourceConfig,
    mapping: UpstreamMonitorGroupMapping,
  ): number {
    const option = selectedUpstreamGroupOption(source, mapping);
    return Number(option?.reference_multiplier || 0);
  }

  function mappingReferenceMultiplier(
    source: UpstreamMonitorSourceConfig,
    mapping: UpstreamMonitorGroupMapping,
    includeSourceFallback: boolean,
  ): number {
    const selectedMultiplier = selectedUpstreamGroupReferenceMultiplier(source, mapping);
    if (selectedMultiplier > 0) {
      return selectedMultiplier;
    }
    const directMultiplier = Number(mapping.reference_multiplier || 0);
    if (directMultiplier > 0) {
      return directMultiplier;
    }
    return includeSourceFallback ? Number(source.reference_multiplier || 0) : 0;
  }

  function isCompleteMapping(mapping: UpstreamMonitorGroupMapping): boolean {
    return Boolean(
      localGroupIDForMapping(mapping) > 0 &&
      mapping.local_group.trim() &&
      (mapping.upstream_group_key.trim() || mapping.upstream_group.trim()),
    );
  }

  function isCompleteMappingForSource(
    source: UpstreamMonitorSourceConfig,
    mapping: UpstreamMonitorGroupMapping,
  ): boolean {
    return isCompleteMapping(mapping) && mappingReferenceMultiplier(source, mapping, true) > 0;
  }

  function sourceMappingRows(source: UpstreamMonitorSourceConfig): UpstreamMonitorGroupMapping[] {
    return getSourceMappingRows(options.form.value.group_mappings, source);
  }

  function sourceMappingRowsBySourceID(sourceID: string): UpstreamMonitorGroupMapping[] {
    return getSourceMappingRowsBySourceID(options.form.value.group_mappings, sourceID);
  }

  function sourceMappedMappings(source: UpstreamMonitorSourceConfig): UpstreamMonitorGroupMapping[] {
    return sourceMappingRows(source).filter((mapping) => isCompleteMappingForSource(source, mapping));
  }

  function sourceGroupRowForMapping(
    source: UpstreamMonitorSourceConfig,
    mapping: UpstreamMonitorGroupMapping,
  ): UpstreamMonitorPreviewGroupRow | undefined {
    const sid = sanitizeSourceID(source);
    const groupID = localGroupIDForMapping(mapping);
    if (!sid) return undefined;
    const rows = options.previewGroupRows.value.filter((row) => uniqueStringIDs(row.source_ids).includes(sid));
    const byID = rows.find((row) => row.mapping_id === mapping.id);
    if (byID) return byID;

    const upstreamKey = String(mapping.upstream_group_key || "").trim();
    if (upstreamKey) {
      const byUpstreamKey = rows.find((row) =>
        row.upstream_group_key === upstreamKey &&
        (
          (groupID > 0 && row.local_group_id === groupID) ||
          (groupID <= 0 && row.local_group.trim() === mapping.local_group.trim())
        ),
      );
      if (byUpstreamKey) return byUpstreamKey;
    }

    return rows.find((row) =>
      (
        (groupID > 0 && row.local_group_id === groupID) ||
        (groupID <= 0 && row.local_group.trim() === mapping.local_group.trim())
      ),
    );
  }

  function estimatedMarginRate(localMultiplier: number, referenceMultiplier: number): number | null {
    if (localMultiplier <= 0 || referenceMultiplier <= 0) {
      return null;
    }
    return (localMultiplier - referenceMultiplier) / localMultiplier;
  }

  function effectiveLocalMultiplier(source: UpstreamMonitorSourceConfig, mapping: UpstreamMonitorGroupMapping): number {
    const direct = localMultiplierForMapping(mapping);
    if (direct > 0) return direct;
    return Number(sourceGroupRowForMapping(source, mapping)?.local_group_multiplier || 0);
  }

  function effectiveReferenceMultiplier(source: UpstreamMonitorSourceConfig, mapping: UpstreamMonitorGroupMapping): number {
    return effectiveSourceCostMultiplier(source, mappingReferenceMultiplier(source, mapping, true));
  }

  function isMappingLocalGroup(
    mapping: UpstreamMonitorGroupMapping,
    group: UpstreamMonitorPreviewGroupOption,
  ): boolean {
    return (
      (localGroupIDForMapping(mapping) > 0 && group.group_id > 0 && localGroupIDForMapping(mapping) === group.group_id) ||
      (localGroupIDForMapping(mapping) <= 0 && mapping.local_group.trim() === group.group_name.trim())
    );
  }

  function sourceHasLocalGroupMapping(
    source: UpstreamMonitorSourceConfig,
    group: UpstreamMonitorPreviewGroupOption,
  ): boolean {
    const sid = sanitizeSourceID(source);
    if (!sid) return false;
    return options.form.value.group_mappings.some((mapping) =>
      isMappingLocalGroup(mapping, group) && uniqueStringIDs(mapping.source_ids).includes(sid),
    );
  }

  function groupOptionsForSource(source: UpstreamMonitorSourceConfig): UpstreamMonitorPreviewGroupOption[] {
    const groups = [...options.localGroups.value];
    groups.sort((left, right) => {
      const leftSelected = sourceHasLocalGroupMapping(source, left);
      const rightSelected = sourceHasLocalGroupMapping(source, right);
      if (leftSelected !== rightSelected) {
        return leftSelected ? -1 : 1;
      }
      const leftFamily = options.modelFamilyForGroup(left);
      const rightFamily = options.modelFamilyForGroup(right);
      if (leftFamily !== rightFamily) {
        return leftFamily.localeCompare(rightFamily);
      }
      return left.group_name.localeCompare(right.group_name);
    });
    return groups;
  }

  function localGroupSelectOptionsForSource(source: UpstreamMonitorSourceConfig): SourceMappingSelectOption[] {
    return groupOptionsForSource(source).map((group) => ({
      value: String(group.group_id),
      label: options.groupOptionLabel(group),
    }));
  }

  function upstreamGroupSelectOptionsForSource(source: UpstreamMonitorSourceConfig): SourceMappingSelectOption[] {
    return upstreamGroupOptionsForSource(source).map((option) => ({
      value: option.key,
      label: options.upstreamGroupOptionLabel(option),
    }));
  }

  function sourceMappingRowElementID(source: UpstreamMonitorSourceConfig, mapping: UpstreamMonitorGroupMapping): string {
    return getSourceMappingRowElementID(sanitizeSourceID(source), mapping);
  }

  function sourceMappingSummary(source: UpstreamMonitorSourceConfig, mapping: UpstreamMonitorGroupMapping): string {
    const messages = options.messages();
    const local = mapping.local_group.trim() || messages.selectLocalGroup;
    const upstream = mapping.upstream_group.trim() || messages.selectUpstreamGroup;
    return `${local} -> ${upstream} · ${multiplierValueLabel(options.formatMultiplier, effectiveLocalMultiplier(source, mapping))} / ${multiplierValueLabel(options.formatMultiplier, effectiveReferenceMultiplier(source, mapping))}`;
  }

  function sourceMappingRowViews(source: UpstreamMonitorSourceConfig): SourceMappingRowView[] {
    return sourceMappingRows(source).map((mapping) => ({
      key: `${source.id}_${mapping.id}`,
      elementId: sourceMappingRowElementID(source, mapping),
      mapping,
      summary: sourceMappingSummary(source, mapping),
      localGroupID: String(localGroupIDForMapping(mapping) || ""),
      isComplete: isCompleteMapping(mapping),
      isNew: mapping.id === lastAddedMappingID.value,
    }));
  }

  function hasDuplicateSourceMapping(source: UpstreamMonitorSourceConfig, mapping: UpstreamMonitorGroupMapping): boolean {
    return hasDuplicateSourceMappingRow(options.form.value.group_mappings, sanitizeSourceID(source), mapping, {
      localGroupIDForMapping,
      isCompleteMapping,
    });
  }

  async function addMapping(source: UpstreamMonitorSourceConfig): Promise<SourceMappingControllerResult> {
    const sid = ensureSourceID(source);
    if (groupOptionsForSource(source).length === 0) {
      const reason = options.messages().noGroupOptions;
      options.onError(reason);
      return { ok: false, reason };
    }
    const beforeCount = sourceMappingRowsBySourceID(sid).length;
    const result = addSourceMappingRow(options.form.value.group_mappings, sid, options.createID);
    replaceGroupMappings(result.mappings);

    await nextTick();
    const afterCount = sourceMappingRowsBySourceID(sid).length;
    if (afterCount <= beforeCount) {
      const reason = options.messages().mappingRowAddFailed;
      options.onError(reason);
      return { ok: false, reason };
    }
    lastAddedMappingID.value = result.mapping.id;
    return { ok: true, mapping: result.mapping };
  }

  function removeMapping(
    source: UpstreamMonitorSourceConfig,
    mapping: UpstreamMonitorGroupMapping,
  ): SourceMappingControllerResult {
    const result = removeSourceMappingRow(options.form.value.group_mappings, sanitizeSourceID(source), mapping.id);
    replaceGroupMappings(result.mappings);
    if (!result.removed) {
      const reason = options.messages().mappingRemoveFailed;
      options.onError(reason);
      return { ok: false, reason };
    }
    return { ok: true };
  }

  function selectUpstreamGroup(
    source: UpstreamMonitorSourceConfig,
    mapping: UpstreamMonitorGroupMapping,
    key: string,
  ): SourceMappingControllerResult {
    const nextMapping = { ...mapping };
    const option = findUpstreamGroupOption(source, key);
    if (!option) {
      nextMapping.upstream_group_key = "";
      nextMapping.upstream_group = "";
      nextMapping.reference_multiplier = 0;
    } else {
      nextMapping.upstream_group_key = option.key;
      nextMapping.upstream_group = option.name;
      nextMapping.reference_multiplier = Number(option.reference_multiplier || 0);
    }
    const result = replaceSourceMappingRowForSource(options.form.value.group_mappings, sanitizeSourceID(source), nextMapping);
    if (result.replaced) {
      replaceGroupMappings(result.mappings);
      return { ok: true, mapping: nextMapping };
    }
    const reason = options.messages().mappingBindFailed;
    options.onError(reason);
    return { ok: false, reason };
  }

  function updateLocalGroup(
    source: UpstreamMonitorSourceConfig,
    mapping: UpstreamMonitorGroupMapping,
    value: string,
  ): SourceMappingControllerResult {
    const nextMapping = { ...mapping };
    const group = groupOptionByID(Number(value || 0));
    if (!group) {
      nextMapping.local_group_id = 0;
      nextMapping.local_group = "";
    } else {
      const oldLocalGroup = mapping.local_group.trim();
      nextMapping.local_group_id = Number(group.group_id || 0);
      nextMapping.local_group = group.group_name;
      nextMapping.model_family = options.modelFamilyForGroup(group);
      if (!mapping.upstream_group_key && mapping.upstream_group.trim() === oldLocalGroup) {
        nextMapping.upstream_group = "";
      }
    }
    const result = replaceSourceMappingRowForSource(options.form.value.group_mappings, sanitizeSourceID(source), nextMapping);
    if (result.replaced) {
      replaceGroupMappings(result.mappings);
      return { ok: true, mapping: nextMapping };
    }
    const reason = options.messages().mappingBindFailed;
    options.onError(reason);
    return { ok: false, reason };
  }

  function bindMapping(
    source: UpstreamMonitorSourceConfig,
    mapping: UpstreamMonitorGroupMapping,
  ): SourceMappingControllerResult {
    const sid = sanitizeSourceID(source);
    const group = groupOptionForMapping(mapping);
    if (!group) {
      const reason = options.messages().selectLocalGroup;
      options.onError(reason);
      return { ok: false, reason };
    }
    const upstreamOption = selectedUpstreamGroupOption(source, mapping);
    const isLegacyManualMapping = Boolean(
      !mapping.upstream_group_key.trim() &&
      mapping.upstream_group.trim() &&
      Number(mapping.reference_multiplier || 0) > 0,
    );
    if (!upstreamOption && !isLegacyManualMapping) {
      const reason = options.messages().referenceMultiplierRequired;
      options.onError(reason);
      return { ok: false, reason };
    }
    const fallbackUpstreamGroup = String(mapping.upstream_group || "").trim() || group.group_name;
    const referenceMultiplier = upstreamOption
      ? Number(upstreamOption.reference_multiplier || 0)
      : mappingReferenceMultiplier(source, mapping, true);
    const nextMapping: UpstreamMonitorGroupMapping = {
      ...mapping,
      local_group_id: Number(group.group_id || 0),
      local_group: group.group_name,
      model_family: options.modelFamilyForGroup(group),
      upstream_group_key: upstreamOption?.key || "",
      upstream_group: upstreamOption?.name || fallbackUpstreamGroup,
      reference_multiplier: Number.isFinite(referenceMultiplier) && referenceMultiplier > 0 ? referenceMultiplier : 0,
      source_ids: [sid],
    };
    if (mappingReferenceMultiplier(source, nextMapping, true) <= 0) {
      const reason = options.messages().referenceMultiplierRequired;
      options.onError(reason);
      return { ok: false, reason };
    }
    if (hasDuplicateSourceMapping(source, nextMapping)) {
      const reason = options.messages().mappingDuplicate;
      options.onError(reason);
      return { ok: false, reason };
    }
    const result = replaceSourceMappingRowForSource(options.form.value.group_mappings, sid, nextMapping);
    if (!result.replaced) {
      const reason = options.messages().mappingBindFailed;
      options.onError(reason);
      return { ok: false, reason };
    }
    replaceGroupMappings(result.mappings);
    return { ok: true, mapping: nextMapping };
  }

  function hydrateMappingReferenceMultipliers(config: UpstreamMonitorConfig): UpstreamMonitorConfig {
    const sourcesByID = new Map(config.sources.map((source) => [source.id, source]));
    for (const mapping of config.group_mappings) {
      if (!mapping.upstream_group_key) {
        continue;
      }
      for (const sid of uniqueStringIDs(mapping.source_ids)) {
        const source = sourcesByID.get(sid);
        if (!source) continue;
        const option = selectedUpstreamGroupOption(source, mapping);
        const multiplier = Number(option?.reference_multiplier || 0);
        if (multiplier <= 0) continue;
        mapping.reference_multiplier = multiplier;
        mapping.upstream_group = option?.name || mapping.upstream_group;
        break;
      }
    }
    return config;
  }

  function serializeMappings(optionsArg: { dropIncompleteMappings: boolean }): UpstreamMonitorGroupMapping[] {
    return options.form.value.group_mappings
      .flatMap((mapping) => {
        const sourceIds = uniqueStringIDs(mapping.source_ids);
        const baseID = (mapping.id || options.createID("mapping")).trim();
        return sourceIds.map((mappingSourceID) => {
          const source = options.form.value.sources.find((item) => sanitizeSourceID(item) === mappingSourceID);
          return {
            ...mapping,
            id: sourceIds.length > 1 ? buildSourceScopedMappingID(baseID, mappingSourceID) : baseID,
            local_group_id: localGroupIDForMapping(mapping),
            local_group: mapping.local_group.trim(),
            upstream_group_key: (mapping.upstream_group_key || "").trim(),
            upstream_group: mapping.upstream_group.trim(),
            source_ids: [mappingSourceID],
            reference_multiplier: source
              ? mappingReferenceMultiplier(source, mapping, true)
              : Number(mapping.reference_multiplier || 0),
            notes: mapping.notes.trim(),
          };
        });
      })
      .filter((mapping) => {
        if (!optionsArg.dropIncompleteMappings) return true;
        const mappingSourceID = uniqueStringIDs(mapping.source_ids)[0] || "";
        const source = options.form.value.sources.find((item) => sanitizeSourceID(item) === mappingSourceID);
        const isComplete = source
          ? isCompleteMappingForSource(source, mapping)
          : isCompleteMapping(mapping) && Number(mapping.reference_multiplier || 0) > 0;
        return mapping.id && isComplete && mapping.model_family && mapping.source_ids.length > 0;
      });
  }

  return {
    addMapping,
    bindMapping,
    effectiveLocalMultiplier,
    effectiveReferenceMultiplier,
    estimatedMarginRate,
    groupOptionForMapping,
    groupOptionsForSource,
    hydrateMappingReferenceMultipliers,
    isCompleteMapping,
    localGroupIDForMapping,
    localGroupSelectOptionsForSource,
    mappingReferenceMultiplier,
    removeMapping,
    removeSourceFromMappingRows,
    replaceGroupMappings,
    selectUpstreamGroup,
    serializeMappings,
    sourceGroupRowForMapping,
    sourceMappedMappings,
    sourceMappingRowElementID,
    sourceMappingRows,
    sourceMappingRowsBySourceID,
    sourceMappingRowViews,
    updateLocalGroup,
    upstreamGroupOptionsForSource,
    upstreamGroupSelectOptionsForSource,
    normalizeModelFamily,
  };
}
