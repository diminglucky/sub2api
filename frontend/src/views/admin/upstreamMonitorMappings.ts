import type {
  UpstreamMonitorGroupMapping,
  UpstreamMonitorSourceConfig,
} from "@/api/admin/settings";

export type MappingIDFactory = (prefix: string) => string;

export function uniqueStringIDs(values: readonly string[] = []): string[] {
  return Array.from(new Set(values.map((id) => id.trim()).filter(Boolean)));
}

export function getSourceID(source: Pick<UpstreamMonitorSourceConfig, "id">): string {
  return String(source.id || "").trim();
}

export function buildSourceScopedMappingID(baseID: string, sourceID: string): string {
  const cleanBase = String(baseID || "mapping").trim() || "mapping";
  const cleanSource = String(sourceID || "").trim();
  if (!cleanSource || cleanBase.endsWith(`__${cleanSource}`)) {
    return cleanBase;
  }
  return `${cleanBase}__${cleanSource}`;
}

export function getSourceMappingRowsBySourceID(
  mappings: readonly UpstreamMonitorGroupMapping[],
  sourceID: string,
): UpstreamMonitorGroupMapping[] {
  const sid = sourceID.trim();
  if (!sid) return [];
  return mappings.filter((mapping) => uniqueStringIDs(mapping.source_ids).includes(sid));
}

export function getSourceMappingRows(
  mappings: readonly UpstreamMonitorGroupMapping[],
  source: Pick<UpstreamMonitorSourceConfig, "id">,
): UpstreamMonitorGroupMapping[] {
  return getSourceMappingRowsBySourceID(mappings, getSourceID(source));
}

export function getSourceMappingRowElementID(
  sourceID: string,
  mapping: Pick<UpstreamMonitorGroupMapping, "id">,
): string {
  const sid = sourceID.trim() || "source";
  const mid = String(mapping.id || "mapping").trim() || "mapping";
  return `upstream-monitor-${sid}-${mid}`;
}

function lastSourceMappingIndex(
  mappings: readonly UpstreamMonitorGroupMapping[],
  sourceID: string,
): number {
  const sid = sourceID.trim();
  if (!sid) return -1;
  for (let index = mappings.length - 1; index >= 0; index -= 1) {
    if (uniqueStringIDs(mappings[index].source_ids).includes(sid)) {
      return index;
    }
  }
  return -1;
}

export function createSourceMappingRow(
  sourceID: string,
  createID: MappingIDFactory,
): UpstreamMonitorGroupMapping {
  return {
    id: createID("mapping"),
    local_group_id: 0,
    local_group: "",
    upstream_group_key: "",
    upstream_group: "",
    model_family: "mixed",
    source_ids: [sourceID],
    reference_multiplier: 0,
    notes: "",
  };
}

export function addSourceMappingRow(
  mappings: readonly UpstreamMonitorGroupMapping[],
  sourceID: string,
  createID: MappingIDFactory,
): { mappings: UpstreamMonitorGroupMapping[]; mapping: UpstreamMonitorGroupMapping } {
  const mapping = createSourceMappingRow(sourceID, createID);
  const insertAfter = lastSourceMappingIndex(mappings, sourceID);
  const nextMappings = [...mappings];
  if (insertAfter >= 0) {
    nextMappings.splice(insertAfter + 1, 0, mapping);
  } else {
    nextMappings.push(mapping);
  }
  return { mappings: nextMappings, mapping };
}

export function replaceSourceMappingRow(
  mappings: readonly UpstreamMonitorGroupMapping[],
  nextMapping: UpstreamMonitorGroupMapping,
): { mappings: UpstreamMonitorGroupMapping[]; replaced: boolean } {
  let replaced = false;
  const nextMappings = mappings.map((mapping) => {
    if (mapping.id !== nextMapping.id) {
      return mapping;
    }
    replaced = true;
    return nextMapping;
  });
  return { mappings: nextMappings, replaced };
}

export function replaceSourceMappingRowForSource(
  mappings: readonly UpstreamMonitorGroupMapping[],
  sourceID: string,
  nextMapping: UpstreamMonitorGroupMapping,
): { mappings: UpstreamMonitorGroupMapping[]; replaced: boolean } {
  const sid = sourceID.trim();
  if (!sid) {
    return { mappings: [...mappings], replaced: false };
  }

  let replaced = false;
  const nextMappings: UpstreamMonitorGroupMapping[] = [];
  for (const mapping of mappings) {
    if (mapping.id !== nextMapping.id) {
      nextMappings.push(mapping);
      continue;
    }

    const sourceIDs = uniqueStringIDs(mapping.source_ids);
    if (!sourceIDs.includes(sid)) {
      nextMappings.push(mapping);
      continue;
    }

    replaced = true;
    const otherSourceIDs = sourceIDs.filter((id) => id !== sid);
    if (otherSourceIDs.length > 0) {
      nextMappings.push({
        ...mapping,
        source_ids: otherSourceIDs,
      });
      nextMappings.push({
        ...nextMapping,
        id: buildSourceScopedMappingID(nextMapping.id, sid),
        source_ids: [sid],
      });
      continue;
    }

    nextMappings.push({
      ...nextMapping,
      source_ids: [sid],
    });
  }

  return { mappings: nextMappings, replaced };
}

export function removeSourceMappingRow(
  mappings: readonly UpstreamMonitorGroupMapping[],
  sourceID: string,
  mappingID: string,
): { mappings: UpstreamMonitorGroupMapping[]; removed: boolean } {
  const sid = sourceID.trim();
  let removed = false;
  const nextMappings = mappings
    .map((mapping) => {
      if (mapping.id !== mappingID) {
        return mapping;
      }
      const sourceIDs = uniqueStringIDs(mapping.source_ids);
      const nextSourceIDs = sourceIDs.filter((id) => id !== sid);
      removed = removed || nextSourceIDs.length !== sourceIDs.length;
      return {
        ...mapping,
        source_ids: nextSourceIDs,
      };
    })
    .filter((mapping) => mapping.source_ids.length > 0);
  return { mappings: nextMappings, removed };
}

export function removeSourceFromMappingRows(
  mappings: readonly UpstreamMonitorGroupMapping[],
  sourceID: string,
): UpstreamMonitorGroupMapping[] {
  const sid = sourceID.trim();
  if (!sid) return [...mappings];
  return mappings
    .map((mapping) => ({
      ...mapping,
      source_ids: uniqueStringIDs(mapping.source_ids).filter((id) => id !== sid),
    }))
    .filter((mapping) => mapping.source_ids.length > 0);
}

export function buildSourceMappingKey(
  sourceID: string,
  mapping: UpstreamMonitorGroupMapping,
  localGroupID: number,
): string {
  return [
    sourceID.trim(),
    localGroupID || mapping.local_group.trim().toLowerCase(),
    (mapping.upstream_group_key || mapping.upstream_group).trim().toLowerCase(),
    String(mapping.model_family || "").trim().toLowerCase(),
  ].join("|");
}

export function hasDuplicateSourceMappingRow(
  mappings: readonly UpstreamMonitorGroupMapping[],
  sourceID: string,
  mapping: UpstreamMonitorGroupMapping,
  options: {
    localGroupIDForMapping: (mapping: UpstreamMonitorGroupMapping) => number;
    isCompleteMapping: (mapping: UpstreamMonitorGroupMapping) => boolean;
  },
): boolean {
  const key = buildSourceMappingKey(
    sourceID,
    mapping,
    options.localGroupIDForMapping(mapping),
  );
  if (!key) return false;
  return getSourceMappingRowsBySourceID(mappings, sourceID).some((item) =>
    item.id !== mapping.id &&
    options.isCompleteMapping(item) &&
    buildSourceMappingKey(sourceID, item, options.localGroupIDForMapping(item)) === key,
  );
}
