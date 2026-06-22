import { describe, expect, it } from "vitest";

import type { UpstreamMonitorGroupMapping } from "@/api/admin/settings";
import {
  addSourceMappingRow,
  buildSourceScopedMappingID,
  getSourceMappingRowsBySourceID,
  hasDuplicateSourceMappingRow,
  removeSourceFromMappingRows,
  removeSourceMappingRow,
  replaceSourceMappingRow,
  replaceSourceMappingRowForSource,
  uniqueStringIDs,
} from "../upstreamMonitorMappings";

function mapping(partial: Partial<UpstreamMonitorGroupMapping>): UpstreamMonitorGroupMapping {
  return {
    id: "mapping_1",
    local_group_id: 0,
    local_group: "",
    upstream_group_key: "",
    upstream_group: "",
    model_family: "mixed",
    source_ids: [],
    reference_multiplier: 0,
    notes: "",
    ...partial,
  };
}

function isComplete(item: UpstreamMonitorGroupMapping): boolean {
  return Boolean(item.local_group_id > 0 && item.local_group && item.upstream_group_key);
}

describe("upstreamMonitorMappings", () => {
  it("deduplicates and trims source ids", () => {
    expect(uniqueStringIDs([" s1 ", "", "s1", "s2"])).toEqual(["s1", "s2"]);
  });

  it("adds a new binding row after existing rows for the same source", () => {
    let counter = 0;
    const rows = [
      mapping({ id: "a", source_ids: ["s1"] }),
      mapping({ id: "b", source_ids: ["s2"] }),
      mapping({ id: "c", source_ids: ["s1"] }),
    ];

    const result = addSourceMappingRow(rows, "s1", (prefix) => `${prefix}_${++counter}`);

    expect(result.mapping).toMatchObject({ id: "mapping_1", source_ids: ["s1"] });
    expect(result.mappings.map((item) => item.id)).toEqual(["a", "b", "c", "mapping_1"]);
  });

  it("replaces only the selected binding row", () => {
    const rows = [
      mapping({ id: "a", local_group: "A", source_ids: ["s1"] }),
      mapping({ id: "b", local_group: "B", source_ids: ["s1"] }),
    ];

    const result = replaceSourceMappingRow(rows, mapping({ id: "b", local_group: "Updated", source_ids: ["s1"] }));

    expect(result.replaced).toBe(true);
    expect(result.mappings).toEqual([
      mapping({ id: "a", local_group: "A", source_ids: ["s1"] }),
      mapping({ id: "b", local_group: "Updated", source_ids: ["s1"] }),
    ]);
  });

  it("updates a shared binding row without removing other sources", () => {
    const rows = [
      mapping({ id: "shared", local_group: "Old", source_ids: ["s1", "s2"] }),
      mapping({ id: "other", local_group: "Other", source_ids: ["s3"] }),
    ];

    const result = replaceSourceMappingRowForSource(
      rows,
      "s1",
      mapping({ id: "shared", local_group: "Updated", source_ids: ["s1"] }),
    );

    expect(result.replaced).toBe(true);
    expect(result.mappings).toEqual([
      mapping({ id: "shared", local_group: "Old", source_ids: ["s2"] }),
      mapping({ id: "shared__s1", local_group: "Updated", source_ids: ["s1"] }),
      mapping({ id: "other", local_group: "Other", source_ids: ["s3"] }),
    ]);
  });

  it("removes only the current source from a shared binding row", () => {
    const rows = [
      mapping({ id: "shared", local_group: "Pro", source_ids: ["s1", "s2"] }),
      mapping({ id: "solo", local_group: "Solo", source_ids: ["s1"] }),
    ];

    const result = removeSourceMappingRow(rows, "s1", "shared");

    expect(result.removed).toBe(true);
    expect(result.mappings).toEqual([
      mapping({ id: "shared", local_group: "Pro", source_ids: ["s2"] }),
      mapping({ id: "solo", local_group: "Solo", source_ids: ["s1"] }),
    ]);
  });

  it("deletes a row when the removed source was its only source", () => {
    const rows = [
      mapping({ id: "solo", local_group: "Solo", source_ids: ["s1"] }),
      mapping({ id: "other", local_group: "Other", source_ids: ["s2"] }),
    ];

    const result = removeSourceMappingRow(rows, "s1", "solo");

    expect(result.removed).toBe(true);
    expect(result.mappings).toEqual([
      mapping({ id: "other", local_group: "Other", source_ids: ["s2"] }),
    ]);
  });

  it("removes a deleted upstream source from every mapping row without affecting others", () => {
    const rows = [
      mapping({ id: "shared", source_ids: ["s1", "s2"] }),
      mapping({ id: "only", source_ids: ["s1"] }),
      mapping({ id: "other", source_ids: ["s3"] }),
    ];

    expect(removeSourceFromMappingRows(rows, "s1")).toEqual([
      mapping({ id: "shared", source_ids: ["s2"] }),
      mapping({ id: "other", source_ids: ["s3"] }),
    ]);
  });

  it("detects duplicate local/upstream bindings within the same source only", () => {
    const rows = [
      mapping({
        id: "existing",
        local_group_id: 10,
        local_group: "GPT",
        upstream_group_key: "up-1",
        upstream_group: "OpenAI",
        source_ids: ["s1"],
      }),
      mapping({
        id: "same-values-other-source",
        local_group_id: 10,
        local_group: "GPT",
        upstream_group_key: "up-1",
        upstream_group: "OpenAI",
        source_ids: ["s2"],
      }),
    ];
    const candidate = mapping({
      id: "candidate",
      local_group_id: 10,
      local_group: "GPT",
      upstream_group_key: "up-1",
      upstream_group: "OpenAI",
      source_ids: ["s1"],
    });

    expect(
      hasDuplicateSourceMappingRow(rows, "s1", candidate, {
        localGroupIDForMapping: (item) => item.local_group_id,
        isCompleteMapping: isComplete,
      }),
    ).toBe(true);
    expect(
      hasDuplicateSourceMappingRow(rows, "s3", candidate, {
        localGroupIDForMapping: (item) => item.local_group_id,
        isCompleteMapping: isComplete,
      }),
    ).toBe(false);
  });

  it("keeps scoped mapping ids stable when flattening shared mappings", () => {
    expect(buildSourceScopedMappingID("mapping_base", "source_a")).toBe("mapping_base__source_a");
    expect(buildSourceScopedMappingID("mapping_base__source_a", "source_a")).toBe("mapping_base__source_a");
  });

  it("returns only rows bound to the requested source", () => {
    const rows = [
      mapping({ id: "a", source_ids: ["s1", "s2"] }),
      mapping({ id: "b", source_ids: ["s2"] }),
      mapping({ id: "c", source_ids: ["s3"] }),
    ];

    expect(getSourceMappingRowsBySourceID(rows, "s2").map((item) => item.id)).toEqual(["a", "b"]);
  });
});
