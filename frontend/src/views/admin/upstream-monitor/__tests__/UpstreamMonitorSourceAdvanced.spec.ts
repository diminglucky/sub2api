import { mount } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";

import UpstreamMonitorSourceAdvanced from "../UpstreamMonitorSourceAdvanced.vue";
import type { UpstreamMonitorSourceConfig } from "@/api/admin/settings";

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
    auth_username: "",
    auth_header_name: "",
    auth_token: "",
    auth_configured: false,
    currency: "CNY",
    exchange_rate: 1,
    reference_multiplier: 0.08,
    upstream_group_options: [],
    last_sync_at: null,
    last_sync_status: "idle",
    last_sync_error: "",
    notes: "",
    ...partial,
  };
}

function mountAdvanced(overrides: Partial<Parameters<typeof mount>[1]> = {}) {
  const defaultProps = {
    source: source(),
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
    fetchModeOptions: [
      { value: "auto", labelKey: "auto" },
      { value: "json_path", labelKey: "json_path" },
    ],
  };

  return mount(UpstreamMonitorSourceAdvanced, {
    ...overrides,
    props: {
      ...defaultProps,
      ...(overrides.props || {}),
    },
    global: {
      stubs: {
        Icon: { template: "<span />" },
        Input: {
          props: ["modelValue", "label"],
          emits: ["update:modelValue"],
          template: "<label><span>{{ label }}</span><input :value=\"modelValue\" @input=\"$emit('update:modelValue', $event.target.value)\" /></label>",
        },
        TextArea: {
          props: ["modelValue", "label"],
          emits: ["update:modelValue"],
          template: "<label><span>{{ label }}</span><textarea :value=\"modelValue\" @input=\"$emit('update:modelValue', $event.target.value)\" /></label>",
        },
      },
    },
  });
}

describe("UpstreamMonitorSourceAdvanced", () => {
  it("emits the selected account id from the account picker", async () => {
    const wrapper = mountAdvanced();
    const accountButtons = wrapper.findAll(".choice-card");

    await accountButtons[1].trigger("click");

    expect(wrapper.emitted("toggle-account")?.[0]).toEqual([11]);
  });

  it("marks currently selected accounts in the account picker", () => {
    const wrapper = mountAdvanced();
    const accountButtons = wrapper.findAll(".choice-card");

    expect(accountButtons[0].classes()).toContain("choice-card-selected");
    expect(accountButtons[1].classes()).not.toContain("choice-card-selected");
  });

});
