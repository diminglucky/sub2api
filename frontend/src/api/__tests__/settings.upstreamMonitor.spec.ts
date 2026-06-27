import { describe, expect, it } from "vitest";

import { normalizeUpstreamMonitorConfig } from "@/api/admin/settings";

describe("admin settings upstream monitor helpers", () => {
  it("preserves newly entered upstream auth tokens while normalizing", () => {
    const config = normalizeUpstreamMonitorConfig({
      sources: [
        {
          id: "source_1",
          name: "Pool",
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
          auth_token: "user-login-token",
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

    expect(config.sources[0].auth_token).toBe("user-login-token");
  });
});
