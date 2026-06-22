import { computed, ref } from "vue";
import { describe, expect, it, vi } from "vitest";
import type {
  UpstreamMonitorConfig,
  UpstreamMonitorGroupMapping,
  UpstreamMonitorPreviewGroupOption,
  UpstreamMonitorSourceConfig,
  UpstreamMonitorUpstreamGroupOption,
} from "@/api/admin/settings";
import { useSourceMappingRows } from "../useSourceMappingRows";

function source(partial: Partial<UpstreamMonitorSourceConfig> = {}): UpstreamMonitorSourceConfig {
  return {
    id: "source_1",
    name: "NB",
    kind: "manual",
    enabled: true,
    auto_sync_enabled: true,
    account_ids: [],
    fetch_mode: "auto",
    base_url: "",
    pricing_url: "https://example.test/pricing",
    pricing_path_hint: "",
    auth_mode: "none",
    auth_header_name: "",
    auth_token: "",
    auth_configured: false,
    currency: "CNY",
    exchange_rate: 1,
    reference_multiplier: 0.08,
    upstream_group_options: [
      {
        key: "up:gpt",
        name: "GPT upstream",
        description: "",
        reference_multiplier: 0.08,
        raw_id: "gpt",
        path: "",
      },
    ],
    last_sync_at: null,
    last_sync_status: "idle",
    last_sync_error: "",
    notes: "",
    ...partial,
  };
}

function group(partial: Partial<UpstreamMonitorPreviewGroupOption> = {}): UpstreamMonitorPreviewGroupOption {
  return {
    group_id: 1,
    group_name: "GPT",
    platform: "OpenAI",
    multiplier: 0.1,
    is_exclusive: false,
    subscription_type: "standard",
    ...partial,
  };
}

function mapping(partial: Partial<UpstreamMonitorGroupMapping> = {}): UpstreamMonitorGroupMapping {
  return {
    id: "mapping_1",
    local_group_id: 1,
    local_group: "GPT",
    upstream_group_key: "up:gpt",
    upstream_group: "GPT upstream",
    model_family: "gpt",
    source_ids: ["source_1"],
    reference_multiplier: 0.08,
    notes: "",
    ...partial,
  };
}

function mountController(configPartial: Partial<UpstreamMonitorConfig> = {}) {
  const form = ref<UpstreamMonitorConfig>({
    enabled: true,
    auto_refresh_enabled: true,
    refresh_interval_minutes: 10,
    default_exchange_rate: 1,
    default_profit_rate_threshold: 0.15,
    warning_rate_threshold: 0.08,
    critical_rate_threshold: 0,
    notify_on_critical_only: true,
    sources: [source()],
    group_mappings: [],
    ...configPartial,
  });
  const errors: string[] = [];
  let idCounter = 0;
  const controller = useSourceMappingRows({
    form,
    localGroups: ref([group()]),
    previewGroupRows: computed(() => []),
    messages: () => ({
      noGroupOptions: "no groups",
      mappingRowAddFailed: "add failed",
      selectLocalGroup: "select local",
      selectUpstreamGroup: "select upstream",
      mappingDuplicate: "duplicate",
      mappingBindFailed: "bind failed",
      mappingRemoveFailed: "remove failed",
    }),
    createID: (prefix) => `${prefix}_${++idCounter}`,
    modelFamilyForGroup: () => "gpt",
    formatMultiplier: (value) => Number(value || 0).toFixed(2),
    formatPercent: (value) => `${(value * 100).toFixed(1)}%`,
    groupOptionLabel: (item) => `${item.group_name} ${item.multiplier}`,
    upstreamGroupOptionLabel: (item: UpstreamMonitorUpstreamGroupOption) => `${item.name} ${item.reference_multiplier}`,
    onError: (message) => errors.push(message),
  });
  return { controller, errors, form };
}

describe("useSourceMappingRows", () => {
  it("adds, edits, binds, serializes, and removes a mapping row", async () => {
    const { controller, form } = mountController();
    const currentSource = form.value.sources[0];

    const added = await controller.addMapping(currentSource);
    expect(added.ok).toBe(true);
    expect(form.value.group_mappings).toHaveLength(1);

    const row = form.value.group_mappings[0];
    controller.updateLocalGroup(currentSource, row, "1");
    controller.selectUpstreamGroup(currentSource, form.value.group_mappings[0], "up:gpt");

    const bindResult = controller.bindMapping(currentSource, form.value.group_mappings[0]);
    expect(bindResult.ok).toBe(true);
    expect(controller.sourceMappingRowViews(currentSource)[0]).toMatchObject({
      localGroupID: "1",
      isComplete: true,
      localMultiplierLabel: "0.10x",
      referenceMultiplierLabel: "0.08x",
      marginRateLabel: "20.0%",
    });

    expect(controller.serializeMappings({ dropIncompleteMappings: true })).toEqual([
      mapping({ id: "mapping_1" }),
    ]);

    const removeResult = controller.removeMapping(currentSource, form.value.group_mappings[0]);
    expect(removeResult.ok).toBe(true);
    expect(form.value.group_mappings).toEqual([]);
  });

  it("rejects duplicate local/upstream bindings on the same source", () => {
    const { controller, errors, form } = mountController({
      group_mappings: [
        mapping({ id: "existing" }),
        mapping({ id: "candidate" }),
      ],
    });

    const result = controller.bindMapping(form.value.sources[0], form.value.group_mappings[1]);

    expect(result).toEqual({ ok: false, reason: "duplicate" });
    expect(errors).toEqual(["duplicate"]);
  });

  it("shows a clear error when there is no local group to add", async () => {
    const form = ref<UpstreamMonitorConfig>({
      enabled: true,
      auto_refresh_enabled: true,
      refresh_interval_minutes: 10,
      default_exchange_rate: 1,
      default_profit_rate_threshold: 0.15,
      warning_rate_threshold: 0.08,
      critical_rate_threshold: 0,
      notify_on_critical_only: true,
      sources: [source()],
      group_mappings: [],
    });
    const onError = vi.fn();
    const controller = useSourceMappingRows({
      form,
      localGroups: ref([]),
      previewGroupRows: ref([]),
      messages: () => ({
        noGroupOptions: "no groups",
        mappingRowAddFailed: "add failed",
        selectLocalGroup: "select local",
        selectUpstreamGroup: "select upstream",
        mappingDuplicate: "duplicate",
        mappingBindFailed: "bind failed",
        mappingRemoveFailed: "remove failed",
      }),
      createID: (prefix) => `${prefix}_1`,
      modelFamilyForGroup: () => "mixed",
      formatMultiplier: String,
      formatPercent: String,
      groupOptionLabel: (item) => item.group_name,
      upstreamGroupOptionLabel: (item) => item.name,
      onError,
    });

    await expect(controller.addMapping(form.value.sources[0])).resolves.toEqual({ ok: false, reason: "no groups" });
    expect(onError).toHaveBeenCalledWith("no groups");
    expect(form.value.group_mappings).toEqual([]);
  });
});
