import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";

import SourceMappingRowsEditor from "../SourceMappingRowsEditor.vue";
import type { UpstreamMonitorGroupMapping } from "@/api/admin/settings";

vi.mock("vue-i18n", () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params?.index ? `${key}:${params.index}` : key,
  }),
}));

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

function mountEditor(options: Partial<Parameters<typeof mount>[1]> = {}) {
  const rowMapping = mapping();
  return mount(SourceMappingRowsEditor, {
    props: {
      sourceId: "source_1",
      boundCount: 1,
      totalCount: 1,
      canAdd: true,
      rows: [
        {
          key: "source_1_mapping_1",
          elementId: "row_1",
          mapping: rowMapping,
          summary: "GPT -> GPT upstream",
          localGroupID: "1",
          isComplete: true,
          isNew: false,
        },
      ],
      localGroupOptions: [
        { value: "1", label: "GPT · 0.10x" },
        { value: "2", label: "Claude · 0.20x" },
      ],
      upstreamGroupOptions: [
        { value: "up:gpt", label: "GPT upstream · 0.08x" },
        { value: "up:claude", label: "Claude upstream · 0.18x" },
      ],
    },
    global: {
      stubs: {
        Icon: {
          props: ["name"],
          template: "<span />",
        },
      },
    },
    ...options,
  });
}

describe("SourceMappingRowsEditor", () => {
  it("emits add/remove/bind events from the visible controls", async () => {
    const wrapper = mountEditor();

    await wrapper.find("button.btn-secondary").trigger("click");
    await wrapper.find("button.text-gray-500").trigger("click");
    await wrapper.find("button.btn-primary").trigger("click");

    expect(wrapper.emitted("add")).toHaveLength(1);
    expect(wrapper.emitted("remove")?.[0]?.[0]).toMatchObject({ id: "mapping_1" });
    expect(wrapper.emitted("bind")?.[0]?.[0]).toMatchObject({ id: "mapping_1" });
  });

  it("emits selected local and upstream group values", async () => {
    const wrapper = mountEditor();
    const selects = wrapper.findAll("select");

    await selects[0].setValue("2");
    await selects[1].setValue("up:claude");

    expect(wrapper.emitted("update-local")?.[0]?.[0]).toMatchObject({
      mapping: { id: "mapping_1" },
      value: "2",
    });
    expect(wrapper.emitted("select-upstream")?.[0]?.[0]).toMatchObject({
      mapping: { id: "mapping_1" },
      value: "up:claude",
    });
  });

  it("disables adding when there are no local groups", async () => {
    const wrapper = mountEditor({
      props: {
        sourceId: "source_1",
        boundCount: 0,
        totalCount: 0,
        canAdd: false,
        rows: [],
        localGroupOptions: [],
        upstreamGroupOptions: [],
      },
    });

    const addButton = wrapper.find("button.btn-secondary");
    expect(addButton.attributes("disabled")).toBeDefined();
    await addButton.trigger("click");
    expect(wrapper.emitted("add")).toBeUndefined();
  });
});
