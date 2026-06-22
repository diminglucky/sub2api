import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";

import UpstreamMonitorSourceCard from "../UpstreamMonitorSourceCard.vue";
import type {
  UpstreamMonitorGroupMapping,
  UpstreamMonitorSourceConfig,
} from "@/api/admin/settings";

vi.mock("vue-i18n", () => ({
  useI18n: () => ({
    locale: "zh",
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key,
  }),
}));

function source(partial: Partial<UpstreamMonitorSourceConfig> = {}): UpstreamMonitorSourceConfig {
  return {
    id: "source_1",
    name: "NBility",
    kind: "newapi",
    enabled: true,
    auto_sync_enabled: true,
    account_ids: [10],
    fetch_mode: "auto",
    base_url: "https://api.example.com",
    pricing_url: "https://api.example.com/pricing",
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
        reference_multiplier: 0.08,
        description: "",
        raw_id: "",
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

function mountCard(overrides: Partial<Parameters<typeof mount>[1]> = {}) {
  const rowMapping = mapping();
  const defaultProps = {
    source: source(),
    sourceId: "source_1",
    index: 0,
    expanded: true,
    sourceSyncing: false,
    canSync: true,
    mappedGroupsCount: 1,
    mappingRowsTotalCount: 1,
    canAddMapping: true,
    mappingRows: [
      {
        key: "source_1_mapping_1",
        elementId: "mapping-row-1",
        mapping: rowMapping,
        summary: "GPT -> GPT upstream",
        localGroupID: "1",
        isComplete: true,
        isNew: false,
        localMultiplierLabel: "0.10x",
        referenceMultiplierLabel: "0.08x",
        marginRateLabel: "20.0%",
      },
    ],
    localGroupOptions: [
      { value: "1", label: "GPT · 0.10x" },
    ],
    upstreamGroupOptions: [
      { value: "up:gpt", label: "GPT upstream · 0.08x" },
    ],
    upstreamGroupOptionChips: [
      {
        key: "up:gpt",
        name: "GPT upstream",
        reference_multiplier: 0.08,
        description: "",
        raw_id: "",
        path: "",
      },
    ],
    hiddenUpstreamGroupOptionCount: 0,
    summaryGroups: [
      {
        key: "summary_1",
        name: "GPT",
        upstreamGroup: "GPT upstream",
        modelFamily: "gpt" as const,
        localMultiplier: 0.1,
        referenceMultiplier: 0.08,
        marginRate: 0.2,
        status: "healthy",
      },
    ],
    hiddenSummaryGroupCount: 0,
    worstStatus: "healthy",
    lowestMarginLabel: "20.0%",
    accountOptions: [
      {
        account_id: 10,
        account_name: "OpenAI key",
        platform: "openai",
        type: "api_key",
        rate_multiplier: 1,
        group_names: ["GPT"],
      },
      {
        account_id: 11,
        account_name: "Claude key",
        platform: "anthropic",
        type: "api_key",
        rate_multiplier: 1,
        group_names: ["Claude"],
      },
    ],
    sourceKindOptions: [
      { value: "manual", labelKey: "manual" },
      { value: "newapi", labelKey: "newapi" },
    ],
    authModeOptions: [
      { value: "none", labelKey: "none" },
      { value: "bearer", labelKey: "bearer" },
    ],
    fetchModeOptions: [
      { value: "auto" as const, labelKey: "auto" },
      { value: "json_path" as const, labelKey: "json_path" },
    ],
  };

  return mount(UpstreamMonitorSourceCard, {
    global: {
      stubs: {
        Icon: { template: "<span />" },
        Input: {
          props: ["modelValue", "label"],
          emits: ["update:modelValue"],
          template: "<label><span>{{ label }}</span><input :value=\"modelValue\" @input=\"$emit('update:modelValue', $event.target.value)\" /></label>",
        },
        Toggle: {
          props: ["modelValue"],
          emits: ["update:modelValue"],
          template: "<button type=\"button\" data-test=\"toggle\" @click=\"$emit('update:modelValue', !modelValue)\">toggle</button>",
        },
        TextArea: {
          props: ["modelValue", "label"],
          emits: ["update:modelValue"],
          template: "<label><span>{{ label }}</span><textarea :value=\"modelValue\" @input=\"$emit('update:modelValue', $event.target.value)\" /></label>",
        },
        SourceMappingRowsEditor: {
          props: ["rows"],
          emits: [
            "add",
            "remove",
            "bind",
            "update-local",
            "select-upstream",
            "update-manual-upstream",
            "update-reference-multiplier",
          ],
          template: `
            <div data-test="mapping-editor">
              <button type="button" data-test="mapping-add" @click="$emit('add')">add</button>
              <button type="button" data-test="mapping-remove" @click="$emit('remove', rows[0].mapping)">remove</button>
              <button type="button" data-test="mapping-bind" @click="$emit('bind', rows[0].mapping)">bind</button>
              <button type="button" data-test="mapping-update-local" @click="$emit('update-local', { mapping: rows[0].mapping, value: '2' })">local</button>
              <button type="button" data-test="mapping-select-upstream" @click="$emit('select-upstream', { mapping: rows[0].mapping, value: 'up:claude' })">upstream</button>
              <button type="button" data-test="mapping-update-manual-upstream" @click="$emit('update-manual-upstream', { mapping: rows[0].mapping, value: 'manual-gpt' })">manual</button>
              <button type="button" data-test="mapping-update-reference" @click="$emit('update-reference-multiplier', { mapping: rows[0].mapping, value: '0.12' })">reference</button>
            </div>
          `,
        },
      },
    },
    ...overrides,
    props: {
      ...defaultProps,
      ...(overrides.props || {}),
    },
  });
}

describe("UpstreamMonitorSourceCard", () => {
  it("emits explicit card actions without parent state coupling", async () => {
    const wrapper = mountCard();

    await wrapper.find('[data-source-card-action="collapse"]').trigger("click");
    await wrapper.find('[data-source-card-action="remove"]').trigger("click");
    await wrapper.find('[data-source-card-action="apply-preset"]').trigger("click");
    await wrapper.find('[data-source-card-action="sync"]').trigger("click");

    expect(wrapper.emitted("toggle-expanded")).toHaveLength(1);
    expect(wrapper.emitted("remove")).toHaveLength(1);
    expect(wrapper.emitted("apply-preset")).toHaveLength(1);
    expect(wrapper.emitted("sync")).toHaveLength(1);
  });

  it("keeps the sync action disabled when the parent marks it unavailable", async () => {
    const wrapper = mountCard({
      props: {
        canSync: false,
      },
    });

    const syncButton = wrapper.find('[data-source-card-action="sync"]');
    expect(syncButton.attributes("disabled")).toBeDefined();
  });

  it("bubbles mapping editor events with the selected mapping payload", async () => {
    const wrapper = mountCard();

    await wrapper.find('[data-test="mapping-add"]').trigger("click");
    await wrapper.find('[data-test="mapping-remove"]').trigger("click");
    await wrapper.find('[data-test="mapping-bind"]').trigger("click");
    await wrapper.find('[data-test="mapping-update-local"]').trigger("click");
    await wrapper.find('[data-test="mapping-select-upstream"]').trigger("click");
    await wrapper.find('[data-test="mapping-update-manual-upstream"]').trigger("click");
    await wrapper.find('[data-test="mapping-update-reference"]').trigger("click");

    expect(wrapper.emitted("add-mapping")).toHaveLength(1);
    expect(wrapper.emitted("remove-mapping")?.[0]?.[0]).toMatchObject({ id: "mapping_1" });
    expect(wrapper.emitted("bind-mapping")?.[0]?.[0]).toMatchObject({ id: "mapping_1" });
    expect(wrapper.emitted("update-local-mapping")?.[0]?.[0]).toMatchObject({
      mapping: { id: "mapping_1" },
      value: "2",
    });
    expect(wrapper.emitted("select-upstream-mapping")?.[0]?.[0]).toMatchObject({
      mapping: { id: "mapping_1" },
      value: "up:claude",
    });
    expect(wrapper.emitted("update-manual-upstream-mapping")?.[0]?.[0]).toMatchObject({
      mapping: { id: "mapping_1" },
      value: "manual-gpt",
    });
    expect(wrapper.emitted("update-reference-multiplier")?.[0]?.[0]).toMatchObject({
      mapping: { id: "mapping_1" },
      value: "0.12",
    });
  });

  it("emits the account id when an account choice is clicked", async () => {
    const wrapper = mountCard();
    const accountButtons = wrapper.findAll(".choice-card");

    await accountButtons[1].trigger("click");

    expect(wrapper.emitted("toggle-account")?.[0]).toEqual([11]);
  });
});
