# Downstream Subsite V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Add `draw.superai.sbs` as a low-coupling downstream frontend with shared users, balance, recharge, usage, API keys, custom prices, and settlement records.

**Architecture:** Main Sub2API stays the control plane. A new `backend/internal/downstream` module owns subsite metadata, membership, price overrides, settlement records, and internal APIs. A separate `frontend-downstream` app serves `draw.superai.sbs`. OpenAI-compatible traffic remains `https://draw.superai.sbs/v1/...`.

**Tech Stack:** Go, Gin, SQL migrations, Redis cache, Vue 3, Vite, Vitest, Cloudflare wildcard routing.

## Global Constraints

- Do not change OpenAI-compatible `/v1/...` request or response formats.
- V1 pilot domain is `draw.superai.sbs`.
- Main users, balance, payment orders, API keys, models, upstream accounts, and gateway remain single-instance.
- Subsite can only read and write data scoped by `subsite_id`.
- `subsite_id = NULL` means main-site data.
- V1 settlement records ledger entries; it does not pay out automatically.

---

### Task 1: Add Subsite Schema And Migration

**Files:**

- Create: `backend/migrations/242_downstream_subsites.sql`
- Create: `backend/internal/downstream/models.go`
- Create: `backend/internal/downstream/repository.go`
- Test: `backend/internal/downstream/repository_test.go`

**Interfaces:**

- Produces: `Subsite`, `PriceOverride`, `SettlementEntry`
- Produces: `NewRepository(db *sql.DB) *Repository`
- Produces: `Repository.GetSubsiteBySlug(ctx, slug)`

**Steps:**

- [ ] Write failing repository test for `draw.superai.sbs`.
- [ ] Run `go test ./internal/downstream -run TestRepositoryGetSubsiteBySlug -v` and expect failure.
- [ ] Add tables `subsites`, `subsite_members`, `subsite_prices`, `settlement_ledger`.
- [ ] Add nullable `subsite_id` columns to `payment_orders`, `usage_logs`, `api_keys`.
- [ ] Run the repository test and expect PASS.
- [ ] Commit with `feat: add downstream subsite schema`.

### Task 2: Resolve Subsite Context

**Files:**

- Create: `backend/internal/downstream/context.go`
- Create: `backend/internal/server/middleware/downstream_subsites.go`
- Test: `backend/internal/server/middleware/downstream_subsites_test.go`

**Interfaces:**

- Consumes: `Repository.GetSubsiteBySlug`
- Produces: `downstream.FromGin(c) (*Subsite, bool)`

**Steps:**

- [ ] Write failing middleware test for `Host: draw.superai.sbs`.
- [ ] Run `go test ./internal/server/middleware -run TestDownstreamSubsiteMiddlewareResolvesDraw -v` and expect failure.
- [ ] Resolve the subsite slug from Host only. Do not trust client-supplied `X-Downstream-Slug` or equivalent headers in V1.
- [ ] Return 404 for unknown slugs; do not resolve `/v1/*` differently from current gateway behavior.
- [ ] Run the middleware test and expect PASS.
- [ ] Commit with `feat: resolve downstream subsite context`.

### Task 3: Add Internal Downstream API

**Files:**

- Create: `backend/internal/downstream/handler.go`
- Create: `backend/internal/downstream/routes.go`
- Modify: `backend/internal/server/router.go`
- Test: `backend/internal/downstream/handler_test.go`

**Interfaces:**

- Produces: `GET /api/internal/downstream/v1/site`
- Produces: `GET /api/internal/downstream/v1/me/summary`
- Produces: `GET /api/internal/downstream/v1/admin/summary`

**Steps:**

- [ ] Write failing handler test that returns `draw.superai.sbs` config.
- [ ] Run `go test ./internal/downstream -run TestSiteHandlerReturnsDrawConfig -v` and expect failure.
- [ ] Implement handlers using only `downstream.FromGin` and repository methods scoped by `subsiteID`.
- [ ] Register routes from a separate `routes.go` file.
- [ ] Run handler tests and expect PASS.
- [ ] Commit with `feat: add downstream internal api`.

### Task 4: Attribute Users, Orders, Keys, And Usage

**Files:**

- Modify: `backend/internal/service/auth_service.go`
- Modify: `backend/internal/service/payment_order.go`
- Modify: `backend/internal/service/api_key_service.go`
- Modify: `backend/internal/service/usage_service.go`
- Test: `backend/internal/downstream/attribution_test.go`

**Interfaces:**

- Consumes: resolved subsite from request context.
- Produces: `subsite_members`, `payment_orders.subsite_id`, `api_keys.subsite_id`, `usage_logs.subsite_id`.

**Steps:**

- [ ] Write failing tests for registration, order creation, API key creation, and usage logging on `draw.superai.sbs`.
- [ ] Run `go test ./internal/downstream ./internal/service -run Subsite -v` and expect failure.
- [ ] Add a request context value for `subsite_id`; existing requests keep `NULL`.
- [ ] Registration creates `subsite_members`; orders, keys, and usage copy the request subsite ID.
- [ ] Run focused tests and expect PASS.
- [ ] Commit with `feat: attribute downstream users and usage`.

### Task 5: Price Overrides And Settlement Ledger

**Files:**

- Create: `backend/internal/downstream/pricing.go`
- Create: `backend/internal/downstream/settlement.go`
- Modify: `backend/internal/service/payment_order.go`
- Test: `backend/internal/downstream/pricing_test.go`
- Test: `backend/internal/downstream/settlement_test.go`

**Interfaces:**

- Produces: `ResolveSubsitePrice(ctx, subsiteID, model, basePrice)`
- Produces: `RecordUsageSettlement(ctx, usage, cost, revenue)`

**Steps:**

- [ ] Write failing tests for price override precedence and settlement rows.
- [ ] Run `go test ./internal/downstream -run 'Price|Settlement' -v` and expect failure.
- [ ] Resolve price as `subsite_prices`, then group price, then base price.
- [ ] Treat the main-site price as wholesale cost; sub-site price defaults to the same value.
- [ ] During actual API billing, snapshot main-site cost, sub-site revenue, and sub-site margin in `settlement_ledger`.
- [ ] Run focused tests and expect PASS.
- [ ] Commit with `feat: add downstream pricing and settlement`.

### Task 6: Build Downstream Frontend For draw.superai.sbs

**Files:**

- Create: `frontend-downstream/package.json`
- Create: `frontend-downstream/src/main.ts`
- Create: `frontend-downstream/src/router.ts`
- Create: `frontend-downstream/src/api.ts`
- Create: `frontend-downstream/src/views/LoginView.vue`
- Create: `frontend-downstream/src/views/RegisterView.vue`
- Create: `frontend-downstream/src/views/UserDashboardView.vue`
- Create: `frontend-downstream/src/views/AdminDashboardView.vue`
- Test: `frontend-downstream/src/views/__tests__/AdminDashboardView.spec.ts`

**Interfaces:**

- Consumes: `/api/internal/downstream/v1/site`
- Consumes: `/api/internal/downstream/v1/me/summary`
- Consumes: `/api/internal/downstream/v1/admin/summary`

**Steps:**

- [ ] Write failing admin dashboard test that asserts only draw data is rendered.
- [ ] Run `pnpm --dir frontend-downstream exec vitest run src/views/__tests__/AdminDashboardView.spec.ts` and expect failure.
- [ ] Implement same-origin API client with base `https://draw.superai.sbs`.
- [ ] Keep OpenAI SDK examples on `https://draw.superai.sbs/v1`.
- [ ] Run frontend tests and build.
- [ ] Commit with `feat: add draw downstream frontend`.

### Task 7: Edge Routing And Pilot Deployment

**Files:**

- Create: `deploy/downstream/README.md`
- Create: `deploy/downstream/cloudflared.example.yml`
- Modify: `deploy/Dockerfile`
- Test: `deploy/downstream/README.md`

**Interfaces:**

- Produces: `draw.superai.sbs` route to downstream frontend.
- Preserves: `/v1/*` route to main gateway.

**Steps:**

- [ ] Document exact curl checks for `/`, `/v1/models`, and `/robots.txt`.
- [ ] Run `docker compose -f deploy/docker-compose.local.yml config` and expect valid config.
- [ ] Route frontend assets to downstream frontend without rewriting paths.
- [ ] Route `/v1/*` and OpenAI-compatible paths to main backend unchanged.
- [ ] Verify `curl -sS https://draw.superai.sbs/v1/models` returns OpenAI-compatible JSON auth/error, not HTML 404.
- [ ] Commit with `chore: wire draw downstream routing`.

## Self-Review

- Spec coverage: subsite schema, membership, pricing, settlement, auth, internal API, frontend, routing, and `/v1` preservation are covered.
- Placeholder scan: no `TODO` or `TBD` markers remain.
- Type consistency: `subsite_id`, `Subsite`, `PriceOverride`, and `SettlementEntry` are used consistently.
