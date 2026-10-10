package server

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/server/routes"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/web"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const frameSrcRefreshTimeout = 5 * time.Second

// SetupRouter 配置路由器中间件和路由
func SetupRouter(
	r *gin.Engine,
	handlers *handler.Handlers,
	jwtAuth middleware2.JWTAuthMiddleware,
	optionalJWTAuth middleware2.OptionalJWTAuthMiddleware,
	adminAuth middleware2.AdminAuthMiddleware,
	apiKeyAuth middleware2.APIKeyAuthMiddleware,
	auditLog middleware2.AuditLogMiddleware,
	stepUpAuth middleware2.StepUpAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	compositeResolver *service.CompositeRouteResolver,
	downstreamRepo *downstream.Repository,
	cfg *config.Config,
	redisClient *redis.Client,
) *gin.Engine {
	middleware2.SetIngressRejectRecorder(opsService)
	// 缓存 iframe 页面的 origin 列表，用于动态注入 CSP frame-src
	var cachedFrameOrigins atomic.Pointer[[]string]
	emptyOrigins := []string{}
	cachedFrameOrigins.Store(&emptyOrigins)

	refreshFrameOrigins := func() {
		ctx, cancel := context.WithTimeout(context.Background(), frameSrcRefreshTimeout)
		defer cancel()
		origins, err := settingService.GetFrameSrcOrigins(ctx)
		if err != nil {
			// 获取失败时保留已有缓存，避免 frame-src 被意外清空
			return
		}
		cachedFrameOrigins.Store(&origins)
	}
	refreshFrameOrigins() // 启动时初始化

	// 应用中间件
	r.Use(middleware2.RequestLogger())
	// 将客户端 IP + UA 注入 request context，供 token 签发/会话绑定/审计日志统一读取。
	// 解析模式按请求快照：兼容开关开启时信任原始转发头，关闭时使用 server.trusted_proxies。
	r.Use(middleware2.SessionBindingContext(cfg))
	r.Use(middleware2.Logger())
	r.Use(middleware2.APIOnlyHost(cfg.Server.APIOnlyHosts))
	r.Use(middleware2.CORS(cfg.CORS))
	r.Use(middleware2.SecurityHeaders(cfg.Security.CSP, func() []string {
		if p := cachedFrameOrigins.Load(); p != nil {
			return *p
		}
		return nil
	}))
	r.Use(middleware2.RegionBlock(cfg.RegionBlockSettings))
	r.Use(middleware2.ServerTiming(cfg.Server.EnableServerTiming))
	r.Use(middleware2.DownstreamSubsite(downstreamRepo, downstream.SubsiteBaseDomain))

	// Serve embedded frontend with settings injection if available
	if web.HasEmbeddedFrontend() {
		frontendServer, err := web.NewFrontendServer(settingService) //nolint:staticcheck // SA4023: the !embed stub always errors; embed builds can return nil
		if err != nil {                                              //nolint:staticcheck // SA4023: see above
			log.Printf("Warning: Failed to create frontend server with settings injection: %v, using legacy mode", err)
			r.Use(web.ServeEmbeddedFrontend())
			settingService.SetOnUpdateCallback(refreshFrameOrigins)
		} else {
			// Register combined callback: invalidate HTML cache + refresh frame origins
			settingService.SetOnUpdateCallback(func() {
				frontendServer.InvalidateCache()
				refreshFrameOrigins()
			})
			r.Use(frontendServer.Middleware())
		}
	} else {
		settingService.SetOnUpdateCallback(refreshFrameOrigins)
	}

	// 注册路由
	registerRoutes(r, handlers, jwtAuth, optionalJWTAuth, adminAuth, apiKeyAuth, auditLog, stepUpAuth, apiKeyService, subscriptionService, opsService, settingService, compositeResolver, downstreamRepo, cfg, redisClient)

	return r
}

// subsiteModelCatalog exposes the main-site pricing catalog (group, multiplier,
// model, billing mode) to the sub-site price management endpoint.
func subsiteModelCatalog(h *handler.Handlers) downstream.ModelCatalogProvider {
	return func(ctx context.Context) ([]downstream.ModelCatalogEntry, error) {
		if h == nil || h.ModelPlaza == nil {
			return nil, nil
		}
		groups, err := h.ModelPlaza.ListGroups(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]downstream.ModelCatalogEntry, 0)
		for _, group := range groups {
			for _, model := range group.Models {
				billingMode := ""
				if model.Pricing != nil {
					billingMode = string(model.Pricing.BillingMode)
				}
				out = append(out, downstream.ModelCatalogEntry{
					GroupID:        group.ID,
					GroupName:      group.Name,
					MainMultiplier: group.RateMultiplier,
					Model:          model.Name,
					Platform:       model.Platform,
					BillingMode:    billingMode,
				})
			}
		}
		return out, nil
	}
}

// registerRoutes 注册所有 HTTP 路由
func registerRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	jwtAuth middleware2.JWTAuthMiddleware,
	optionalJWTAuth middleware2.OptionalJWTAuthMiddleware,
	adminAuth middleware2.AdminAuthMiddleware,
	apiKeyAuth middleware2.APIKeyAuthMiddleware,
	auditLog middleware2.AuditLogMiddleware,
	stepUpAuth middleware2.StepUpAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	compositeResolver *service.CompositeRouteResolver,
	downstreamRepo *downstream.Repository,
	cfg *config.Config,
	redisClient *redis.Client,
) {
	// 通用路由（健康检查、状态等）
	routes.RegisterCommonRoutes(r)
	// Sub-site users must see the sub-site price, not the main-site price.
	handler.SetSubsitePriceOverrideProvider(func(ctx context.Context, subsiteID int64) (*handler.SubsitePriceOverrides, error) {
		prices, err := downstreamRepo.ListSubsitePrices(ctx, subsiteID)
		if err != nil {
			return nil, err
		}
		overrides := &handler.SubsitePriceOverrides{
			GroupRates:     map[int64]float64{},
			ModelRates:     map[string]float64{},
			AssignedGroups: map[int64]struct{}{},
		}
		if groups, err := downstreamRepo.ListSubsiteGroupAssignments(ctx, subsiteID); err == nil {
			for _, group := range groups {
				if group.Assigned {
					overrides.AssignedGroups[group.GroupID] = struct{}{}
				}
			}
		}
		for _, price := range prices {
			switch price.Scope {
			case downstream.PriceOverrideScopeGroup:
				if price.GroupID != nil {
					overrides.GroupRates[*price.GroupID] = price.RateMultiplier
				}
			case downstream.PriceOverrideScopeModel:
				if price.Model != nil {
					overrides.ModelRates[*price.Model] = price.RateMultiplier
				}
			}
		}
		return overrides, nil
	})
	// Sub-site API keys may only use groups open to that sub-site.
	middleware2.SetSubsiteGroupGuard(func(ctx context.Context, subsiteID, groupID int64) bool {
		groups, err := downstreamRepo.ListSubsiteGroupAssignments(ctx, subsiteID)
		if err != nil {
			return false
		}
		for _, group := range groups {
			if group.GroupID == groupID {
				return group.Assigned
			}
		}
		return false
	})
	downstream.RegisterRoutes(r, downstreamRepo, gin.HandlerFunc(jwtAuth), downstreamUserContext(), subsiteModelCatalog(h))

	// API v1
	v1 := r.Group("/api/v1")

	// 面板 API 限流器：认证接口按用户 ID、公开接口按安全客户端 IP，
	// 防止高频刷管理面接口打爆数据库（阈值可在系统设置中调整）。
	panelRateLimiter := middleware2.NewPanelRateLimiter(redisClient, settingService)

	// 注册各模块路由
	routes.RegisterAuthRoutes(v1, h, jwtAuth, auditLog, redisClient, settingService, panelRateLimiter)
	routes.RegisterUserRoutes(v1, h, jwtAuth, auditLog, settingService, panelRateLimiter)
	routes.RegisterModelPlazaRoutes(v1, h, optionalJWTAuth, settingService, panelRateLimiter)
	// SuperAI 自定义功能：公开的可用模型接口
	routes.RegisterPublicRoutes(v1, h)
	routes.RegisterAdminRoutes(v1, h, adminAuth, auditLog, stepUpAuth, settingService, panelRateLimiter, downstreamRepo)
	routes.RegisterGatewayRoutes(r, h, apiKeyAuth, apiKeyService, subscriptionService, opsService, settingService, compositeResolver, cfg)
	routes.RegisterPaymentRoutes(v1, h.Payment, h.PaymentWebhook, h.Admin.Payment, jwtAuth, adminAuth, auditLog, settingService, panelRateLimiter, redisClient)

	handler.RegisterPageRoutes(v1, cfg.Pricing.DataDir, gin.HandlerFunc(jwtAuth), gin.HandlerFunc(adminAuth), settingService)
}

func downstreamUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := middleware2.GetAuthSubjectFromContext(c)
		if !ok || subject.UserID <= 0 {
			middleware2.AbortWithError(c, 401, "UNAUTHORIZED", "authentication is required")
			return
		}
		c.Set(downstream.ContextKeyUserID, subject.UserID)
		c.Next()
	}
}
