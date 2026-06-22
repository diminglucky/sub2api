import { flushPromises, mount } from "@vue/test-utils";
import { ref } from "vue";
import { beforeEach, describe, expect, it, vi } from "vitest";

import UpstreamMonitorView from "../UpstreamMonitorView.vue";

const {
  getUpstreamMonitorConfig,
  previewUpstreamMonitorConfig,
  refreshUpstreamMonitorConfig,
  updateUpstreamMonitorConfig,
  getGroups,
  showError,
  showSuccess,
} = vi.hoisted(() => ({
  getUpstreamMonitorConfig: vi.fn(),
  previewUpstreamMonitorConfig: vi.fn(),
  refreshUpstreamMonitorConfig: vi.fn(),
  updateUpstreamMonitorConfig: vi.fn(),
  getGroups: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}));

vi.mock("@/api/admin", () => ({
  adminAPI: {
    settings: {
      getUpstreamMonitorConfig,
      previewUpstreamMonitorConfig,
      refreshUpstreamMonitorConfig,
      updateUpstreamMonitorConfig,
    },
    groups: {
      getAll: getGroups,
    },
  },
}));

vi.mock("@/stores/app", () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}));

vi.mock("vue-i18n", async () => {
  const actual = await vi.importActual<typeof import("vue-i18n")>("vue-i18n");
  return {
    ...actual,
    useI18n: () => ({
      locale: ref("zh"),
      t: (key: string, params?: Record<string, unknown>) =>
        params ? `${key}:${JSON.stringify(params)}` : key,
    }),
  };
});

function previewSnapshot() {
  return {
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
  };
}

function monitorConfig(overrides: Record<string, unknown> = {}) {
  return {
    enabled: false,
    auto_refresh_enabled: true,
    refresh_interval_minutes: 10,
    default_exchange_rate: 1,
    default_profit_rate_threshold: 0.15,
    warning_rate_threshold: 0.08,
    critical_rate_threshold: 0,
    notify_on_critical_only: true,
    sources: [],
    group_mappings: [],
    ...overrides,
  };
}

async function mountView() {
  const wrapper = mount(UpstreamMonitorView, {
    global: {
      stubs: {
        AppLayout: { template: "<div><slot /></div>" },
        Icon: true,
        Input: true,
        Toggle: true,
        UpstreamMonitorSourceCard: true,
      },
    },
  });
  await flushPromises();
  return wrapper;
}

describe("UpstreamMonitorView sources page", () => {
  beforeEach(() => {
    getUpstreamMonitorConfig.mockResolvedValue(monitorConfig());
    previewUpstreamMonitorConfig.mockResolvedValue(previewSnapshot());
    refreshUpstreamMonitorConfig.mockResolvedValue({
      config: monitorConfig(),
      summary: {
        attempted_count: 0,
        success_count: 0,
        failed_count: 0,
        skipped_count: 0,
      },
    });
    updateUpstreamMonitorConfig.mockResolvedValue(monitorConfig());
    getGroups.mockResolvedValue([]);
    showError.mockReset();
    showSuccess.mockReset();
  });

  it("renders upstream source management directly without overview tabs", async () => {
    const wrapper = await mountView();

    expect(wrapper.text()).toContain("admin.upstreamMonitor.sources.title");
    expect(wrapper.find('[data-monitor-tab="overview"]').exists()).toBe(false);
    expect(wrapper.find('[data-monitor-tab="sources"]').exists()).toBe(false);
    expect(wrapper.find("#overview-panel").exists()).toBe(false);
    expect(wrapper.find("#sources-panel").exists()).toBe(false);
    expect(wrapper.find(".upstream-monitor-page").attributes("data-active-tab")).toBeUndefined();
  });

  it("recalculates source summary margin from visible local and upstream multipliers", async () => {
    getGroups.mockResolvedValue([
      {
        id: 1,
        name: "0.2倍率",
        platform: "OpenAI",
        rate_multiplier: 0.2,
        status: "active",
        is_exclusive: false,
        subscription_type: "standard",
      },
    ]);
    getUpstreamMonitorConfig.mockResolvedValue(monitorConfig({
      sources: [
        {
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
          upstream_group_options: [],
          last_sync_at: null,
          last_sync_status: "idle",
          last_sync_error: "",
          notes: "",
        },
      ],
      group_mappings: [
        {
          id: "mapping_1",
          local_group_id: 1,
          local_group: "0.2倍率",
          upstream_group_key: "name:codex-cheap",
          upstream_group: "codex-cheap",
          model_family: "gpt",
          source_ids: ["source_1"],
          reference_multiplier: 0.08,
          notes: "",
        },
      ],
    }));
    previewUpstreamMonitorConfig.mockResolvedValue({
      ...previewSnapshot(),
      group_rows: [
        {
          mapping_id: "mapping_1",
          local_group: "0.2倍率",
          upstream_group_key: "name:codex-cheap",
          upstream_group: "codex-cheap",
          local_group_id: 1,
          local_group_platform: "OpenAI",
          local_group_multiplier: 0.2,
          model_family: "gpt",
          source_ids: ["source_1"],
          source_names: ["NB"],
          source_count: 1,
          enabled_source_count: 1,
          reference_multiplier: 0.08,
          mapping_multiplier: 0.08,
          estimated_margin_rate: 0.6,
          status: "healthy",
          issues: [],
          notes: "",
        },
      ],
    });

    const wrapper = await mountView();
    await flushPromises();

    const sourceCard = wrapper.findComponent({ name: "UpstreamMonitorSourceCard" });
    expect(sourceCard.props("summaryGroups")).toMatchObject([
      {
        localMultiplier: 0.2,
        referenceMultiplier: 0.08,
        marginRate: 0.6,
      },
    ]);
    expect(sourceCard.props("lowestMarginLabel")).toBe("60.0%");
  });

  it("saves the current form before refreshing all upstream sources", async () => {
    const draftConfig = monitorConfig({
      sources: [
        {
          id: "source_1",
          name: "Draft Pool",
          kind: "sub2api",
          enabled: true,
          auto_sync_enabled: true,
          account_ids: [],
          fetch_mode: "auto",
          base_url: "https://pool.example.com",
          pricing_url: "https://pool.example.com/api/v1/channels/available",
          pricing_path_hint: "",
          auth_mode: "bearer",
          auth_header_name: "",
          auth_token: "fresh-token",
          auth_configured: false,
          currency: "CNY",
          exchange_rate: 1,
          reference_multiplier: 0,
          upstream_group_options: [],
          last_sync_at: null,
          last_sync_status: "idle",
          last_sync_error: "",
          notes: "",
        },
      ],
    });
    getUpstreamMonitorConfig.mockResolvedValue(draftConfig);
    updateUpstreamMonitorConfig.mockResolvedValue(draftConfig);
    refreshUpstreamMonitorConfig.mockResolvedValue({
      config: draftConfig,
      summary: {
        attempted_count: 1,
        success_count: 1,
        failed_count: 0,
        skipped_count: 0,
      },
    });

    const wrapper = await mountView();
    await wrapper.find("button.btn-secondary").trigger("click");
    await flushPromises();

    expect(updateUpstreamMonitorConfig.mock.invocationCallOrder[0]).toBeLessThan(
      refreshUpstreamMonitorConfig.mock.invocationCallOrder[0],
    );
    expect(updateUpstreamMonitorConfig.mock.calls.at(-1)?.[0].sources[0].auth_token).toBe("fresh-token");
    expect(refreshUpstreamMonitorConfig).toHaveBeenCalledWith();
  });
});
