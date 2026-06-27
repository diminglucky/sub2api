import { enableAutoUnmount, flushPromises, mount } from "@vue/test-utils";
import { ref } from "vue";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import UpstreamMonitorView from "../UpstreamMonitorView.vue";

enableAutoUnmount(afterEach);

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

async function mountView(extraStubs: Record<string, unknown> = {}) {
  const wrapper = mount(UpstreamMonitorView, {
    global: {
      stubs: {
        AppLayout: { template: "<div><slot /></div>" },
        Icon: true,
        Input: true,
        Toggle: true,
        UpstreamMonitorSourceCard: true,
        ...extraStubs,
      },
    },
  });
  await flushPromises();
  return wrapper;
}

describe("UpstreamMonitorView sources page", () => {
  beforeEach(() => {
    getUpstreamMonitorConfig.mockReset();
    previewUpstreamMonitorConfig.mockReset();
    refreshUpstreamMonitorConfig.mockReset();
    updateUpstreamMonitorConfig.mockReset();
    getGroups.mockReset();
    showError.mockReset();
    showSuccess.mockReset();

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

  it("reloads monitor data after admin compliance is accepted", async () => {
    getGroups
      .mockRejectedValueOnce(new Error("ADMIN_COMPLIANCE_ACK_REQUIRED"))
      .mockResolvedValueOnce([]);

    await mountView();

    expect(getUpstreamMonitorConfig).toHaveBeenCalledTimes(1);
    expect(getGroups).toHaveBeenCalledTimes(1);

    window.dispatchEvent(new CustomEvent("admin-compliance-accepted"));
    await flushPromises();

    expect(getUpstreamMonitorConfig).toHaveBeenCalledTimes(2);
    expect(getGroups).toHaveBeenCalledTimes(2);
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
          pricing_url: "https://pool.example.com/api/v1/groups/available",
          pricing_path_hint: "",
          auth_mode: "bearer",
          auth_username: "",
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

  it("auto-fills the standard endpoint internally when choosing a supported source kind", async () => {
    getUpstreamMonitorConfig.mockResolvedValue(
      monitorConfig({
        sources: [
          {
            id: "source_1",
            name: "Draft Pool",
            kind: "manual",
            enabled: true,
            auto_sync_enabled: true,
            account_ids: [],
            fetch_mode: "auto",
            base_url: "https://pool.example.com",
            pricing_url: "",
            pricing_path_hint: "",
            auth_mode: "none",
            auth_username: "",
            auth_header_name: "",
            auth_token: "",
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
      }),
    );

    const wrapper = await mountView({
      UpstreamMonitorSourceCard: {
        props: [
          "source",
          "sourceId",
          "index",
          "expanded",
          "sourceSyncing",
          "canSync",
          "mappedGroupsCount",
          "mappingRowsTotalCount",
          "canAddMapping",
          "mappingRows",
          "localGroupOptions",
          "upstreamGroupOptions",
          "upstreamGroupOptionChips",
          "hiddenUpstreamGroupOptionCount",
          "summaryGroups",
          "hiddenSummaryGroupCount",
          "worstStatus",
          "lowestMarginLabel",
          "accountOptions",
          "sourceKindOptions",
          "authModeOptions",
          "fetchModeOptions",
        ],
        emits: ["update-source"],
        template: `
          <div
            data-test="source-props"
            :data-auth-mode="source.auth_mode"
            :data-pricing-url="source.pricing_url"
            :data-fetch-mode="source.fetch_mode"
          >
            <button
              type="button"
              data-test="change-kind"
              @click="$emit('update-source', { field: 'kind', value: 'sub2api' })"
            >
              change kind
            </button>
          </div>
        `,
      },
    });
    await flushPromises();

    await wrapper.get('[data-test="change-kind"]').trigger("click");
    await flushPromises();

    const sourceProps = wrapper.get('[data-test="source-props"]');
    expect(sourceProps.attributes("data-pricing-url")).toBe(
      "https://pool.example.com/api/v1/groups/available",
    );
    expect(sourceProps.attributes("data-auth-mode")).toBe("login");
    expect(sourceProps.attributes("data-fetch-mode")).toBe("auto");
  });

  it("uses site URL instead of pricing URL to enable sync for standard sources", async () => {
    const draftConfig = monitorConfig({
      sources: [
        {
          id: "source_1",
          name: "NewAPI Pool",
          kind: "newapi",
          enabled: true,
          auto_sync_enabled: true,
          account_ids: [],
          fetch_mode: "auto",
          base_url: "https://relay.example.com",
          pricing_url: "",
          pricing_path_hint: "",
          auth_mode: "login",
          auth_username: "monitor@example.com",
          auth_header_name: "",
          auth_token: "password",
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

    const wrapper = await mountView({
      UpstreamMonitorSourceCard: {
        props: ["source", "canSync"],
        emits: ["sync"],
        template: `
          <button
            type="button"
            data-test="sync-source"
            :data-can-sync="String(canSync)"
            @click="$emit('sync')"
          >
            sync
          </button>
        `,
      },
    });
    await flushPromises();

    const syncButton = wrapper.get('[data-test="sync-source"]');
    expect(syncButton.attributes("data-can-sync")).toBe("true");

    await syncButton.trigger("click");
    await flushPromises();

    expect(showError).not.toHaveBeenCalledWith("admin.upstreamMonitor.sources.pullGroupsMissingUrl");
    expect(refreshUpstreamMonitorConfig).toHaveBeenCalledWith("source_1");
  });

  it("shows the source sync error instead of a success toast when pulling groups fails", async () => {
    const failedConfig = monitorConfig({
      sources: [
        {
          id: "source_1",
          name: "NB",
          kind: "newapi",
          enabled: true,
          auto_sync_enabled: true,
          account_ids: [],
          fetch_mode: "auto",
          base_url: "https://pool.gptstore.club",
          pricing_url: "",
          pricing_path_hint: "",
          auth_mode: "cookie",
          auth_username: "",
          auth_header_name: "",
          auth_token: "",
          auth_configured: true,
          currency: "CNY",
          exchange_rate: 1,
          reference_multiplier: 0,
          upstream_group_options: [],
          last_sync_at: "2026-06-27T09:24:02Z",
          last_sync_status: "error",
          last_sync_error: "newapi cookie auth requires JSON credential",
          notes: "",
        },
      ],
    });
    getUpstreamMonitorConfig.mockResolvedValue(failedConfig);
    updateUpstreamMonitorConfig.mockResolvedValue(failedConfig);
    refreshUpstreamMonitorConfig.mockResolvedValue({
      config: failedConfig,
      summary: {
        attempted_count: 1,
        success_count: 0,
        failed_count: 1,
        skipped_count: 0,
      },
    });

    const wrapper = await mountView({
      UpstreamMonitorSourceCard: {
        props: ["source", "canSync"],
        emits: ["sync"],
        template: `
          <button
            type="button"
            data-test="sync-source"
            :data-can-sync="String(canSync)"
            @click="$emit('sync')"
          >
            sync
          </button>
        `,
      },
    });
    await flushPromises();

    await wrapper.get('[data-test="sync-source"]').trigger("click");
    await flushPromises();

    expect(showError).toHaveBeenCalledWith("newapi cookie auth requires JSON credential");
    expect(showSuccess).not.toHaveBeenCalledWith(
      expect.stringContaining("admin.upstreamMonitor.sources.pullGroupsSuccess"),
    );
  });

  it("replaces legacy NewAPI pricing endpoints internally when the site URL changes", async () => {
    getUpstreamMonitorConfig.mockResolvedValue(
      monitorConfig({
        sources: [
          {
            id: "source_1",
            name: "Draft Pool",
            kind: "newapi",
            enabled: true,
            auto_sync_enabled: true,
            account_ids: [],
            fetch_mode: "auto",
            base_url: "https://relay.example.com",
            pricing_url: "https://relay.example.com/api/ratio_config",
            pricing_path_hint: "",
            auth_mode: "none",
            auth_username: "",
            auth_header_name: "",
            auth_token: "",
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
      }),
    );

    const wrapper = await mountView({
      UpstreamMonitorSourceCard: {
        props: ["source"],
        emits: ["update-source"],
        template: `
          <div data-test="source-props" :data-pricing-url="source.pricing_url">
            <button
              type="button"
              data-test="change-base-url"
              @click="$emit('update-source', { field: 'base_url', value: 'https://relay.example.com' })"
            >
              change
            </button>
          </div>
        `,
      },
    });
    await flushPromises();

    await wrapper.get('[data-test="change-base-url"]').trigger("click");
    await flushPromises();

    expect(wrapper.get('[data-test="source-props"]').attributes("data-pricing-url")).toBe(
      "https://relay.example.com/api/user/self/groups",
    );
  });
});
