# Custom Feature Modules

This project keeps several SuperAI-specific features outside the upstream
`Wei-Shaw/sub2api` baseline. When syncing upstream, review this checklist before
accepting deletes or large refactors.

## Frontend Entry Points

- `frontend/src/router/customFeatureRoutes.ts`
  - Public models page: `/public-models`
  - User pages: `/playground`, `/lottery`, `/manual`, `/models`
  - Admin pages: `/admin/lottery`, `/admin/upstream-monitor`
- `frontend/src/components/layout/customFeatureNav.ts`
  - User sidebar entries for Playground, models, lottery, manual
  - Admin sidebar entries for lottery and upstream monitor
- `frontend/src/router/index.ts`
  - Should only contain spread insertion points for custom routes.
- `frontend/src/components/layout/AppSidebar.vue`
  - Should only call custom nav builders, not hard-code custom entries.

## Backend Route Entry Points

- `backend/internal/server/routes/custom_features.go`
  - Public API: `/api/v1/public/models/available`
  - User lottery API: `/api/v1/lotteries`
  - Admin lottery API: `/api/v1/admin/lotteries`
  - Admin upstream monitor settings API: `/api/v1/admin/settings/upstream-monitor`
- `backend/internal/server/routes/admin.go`
  - Should call `registerCustomAdminRoutes` and `registerCustomAdminSettingsRoutes`.
- `backend/internal/server/routes/user.go`
  - Should call `registerCustomUserRoutes`.
- `backend/internal/server/routes/public.go`
  - Should call `registerCustomPublicRoutes`.

## Backend Dependency Entry Points

- `backend/internal/service/wire.go`
  - SuperAI-only services should be registered through `CustomFeatureProviderSet`.
- `backend/internal/repository/wire.go`
  - SuperAI-only repositories should be registered through `CustomFeatureProviderSet`.
- `backend/internal/handler/wire.go`
  - SuperAI-only handlers should be registered through `CustomFeatureProviderSet`.

## Feature-Owned Files

### Lottery

- `backend/internal/handler/lottery_handler.go`
- `backend/internal/handler/admin/lottery_handler.go`
- `backend/internal/repository/lottery_repo.go`
- `backend/internal/service/lottery.go`
- `backend/internal/service/lottery_draw_runner.go`
- `backend/migrations/149_add_lottery.sql`
- `frontend/src/api/lottery.ts`
- `frontend/src/api/admin/lottery.ts`
- `frontend/src/views/user/LotteryView.vue`
- `frontend/src/views/admin/LotteryView.vue`

### Playground

- `frontend/src/views/user/PlaygroundView.vue`
- `frontend/src/views/user/__tests__/PlaygroundView.spec.ts`
- `frontend/src/utils/apiEndpoint.ts`
- `frontend/src/utils/__tests__/apiEndpoint.spec.ts`

### Public Models And Manual

- `frontend/src/views/public/PublicModelsView.vue`
- `frontend/src/views/user/ModelsView.vue`
- `frontend/src/views/user/ManualView.vue`
- `backend/internal/handler/available_channel_handler.go`

### Upstream Monitor

- `backend/internal/service/upstream_monitor_config.go`
- `backend/internal/service/upstream_monitor_runner.go`
- `backend/internal/handler/admin/setting_handler.go`
- `backend/internal/handler/admin/setting_handler_upstream_monitor_test.go`
- `frontend/src/api/admin/settings.ts`
- `frontend/src/views/admin/UpstreamMonitorView.vue`
- `frontend/src/views/admin/upstream-monitor/*`
- `frontend/src/views/admin/upstreamMonitorMappings.ts`

### Domain And Region Restrictions

- `backend/internal/server/middleware/api_only_host.go`
- `backend/internal/server/middleware/region_block.go`
- `deploy/.env.example`
- `deploy/config.example.yaml`

### Deployment

- `.github/workflows/superai-image.yml`
- `deploy/docker-compose.ghcr.yml`
- `deploy/docker-compose.neon.yml`
- `deploy/docker-compose.superai.yml`

## Upstream Sync Rules

1. Do not accept upstream deletes for files in this checklist without manually
   porting the feature into the new upstream structure.
2. If upstream splits a large file, move custom logic into the new split file
   instead of restoring the old monolith.
3. If upstream splits i18n locale files, migrate custom translation keys into
   the new domain locale modules.
4. Keep endpoint paths stable unless a migration plan updates the frontend,
   backend, and deployment docs together.
5. After each upstream sync, run:
   - `go test ./internal/server/...`
   - `pnpm --dir frontend build`
