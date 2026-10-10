package downstream

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// priceFloorEpsilon absorbs float noise when comparing multipliers.
const priceFloorEpsilon = 1e-9

// validateFloor rejects a sub-site price that would undercut the main site.
// Group overrides must be >= the group's main multiplier; model overrides must
// be >= 1 because the base price already includes the main-site pricing.
func (h *ScopedAdminHandler) validateFloor(c *gin.Context, req priceOverrideRequest) bool {
	rate := 1.0
	if req.RateMultiplier != nil {
		rate = *req.RateMultiplier
	}
	switch req.Scope {
	case PriceOverrideScopeGroup:
		if req.GroupID == nil || *req.GroupID <= 0 {
			response.BadRequest(c, "group_id is required for a group price override")
			return false
		}
		// 同刻度替换：子站倍率与主站分组倍率同一刻度，下限即主站倍率。
		mainRate, err := h.repo.GroupRateMultiplier(c.Request.Context(), *req.GroupID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				response.BadRequest(c, "unknown group")
				return false
			}
			response.InternalError(c, "failed to load main-site group price")
			return false
		}
		if rate+priceFloorEpsilon < mainRate {
			response.BadRequest(c, "子站价格不能低于主站价格")
			return false
		}
		mainImages, err := h.repo.GroupImagePrices(c.Request.Context(), *req.GroupID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				response.BadRequest(c, "unknown group")
				return false
			}
			response.InternalError(c, "failed to load main-site image price")
			return false
		}
		if !imagePriceAtLeast(req.ImagePrice1K, mainImages.Price1K) ||
			!imagePriceAtLeast(req.ImagePrice2K, mainImages.Price2K) ||
			!imagePriceAtLeast(req.ImagePrice4K, mainImages.Price4K) {
			response.BadRequest(c, "子站生图价格不能低于主站价格")
			return false
		}
	case PriceOverrideScopeModel:
		model := ""
		if req.Model != nil {
			model = strings.TrimSpace(*req.Model)
		}
		if model == "" {
			response.BadRequest(c, "model is required for a model price override")
			return false
		}
		// 同刻度替换：模型级下限是主站上该模型的倍率。同一模型可能出现在多个
		// 分组，取其中最高的分组倍率，保证任何时候都不低于主站。
		mainRate := 1.0
		matched := false
		if h.catalog != nil {
			if catalog, err := h.catalog(c.Request.Context()); err == nil {
				for _, entry := range catalog {
					if entry.Model != model {
						continue
					}
					if !matched || entry.MainMultiplier > mainRate {
						mainRate = entry.MainMultiplier
					}
					matched = true
				}
			}
		}
		// 目录里查不到该模型时无法确定主站倍率，放行（由主站配置兜底）。
		if matched && rate+priceFloorEpsilon < mainRate {
			response.BadRequest(c, "子站倍率不能低于主站倍率")
			return false
		}
	}
	return true
}

// imagePriceAtLeast reports whether the sub-site image price is unset or not
// below the main-site price.
func imagePriceAtLeast(subsite, main *float64) bool {
	if subsite == nil || main == nil {
		return true
	}
	return *subsite+priceFloorEpsilon >= *main
}

// ScopedPriceRepository is the sub-site-scoped control-plane contract. It is
// satisfied by *Repository. Authorization is by sub-site membership role, not
// by main-site admin role, so a sub-site owner manages only its own prices.
type ScopedPriceRepository interface {
	GetSubsiteUserSummary(ctx context.Context, subsiteID, userID int64) (*SubsiteUserSummary, error)
	GetSubsiteByID(ctx context.Context, id int64) (*Subsite, error)
	UpdateSubsite(ctx context.Context, id int64, input SubsiteUpsert) (*Subsite, error)
	ListSubsiteMembers(ctx context.Context, subsiteID int64, limit, offset int) ([]SubsiteMemberSummary, error)
	ListSubsiteModelPrices(ctx context.Context, subsiteID int64, catalog []ModelCatalogEntry) ([]SubsiteModelPrice, error)
	ListSubsiteChannels(ctx context.Context, subsiteID int64) ([]SubsiteChannel, error)
	ListSubsitePrices(ctx context.Context, subsiteID int64) ([]PriceOverride, error)
	CreateSubsitePrice(ctx context.Context, subsiteID int64, input PriceOverrideUpsert) (*PriceOverride, error)
	UpdateSubsitePrice(ctx context.Context, subsiteID, priceID int64, input PriceOverrideUpsert) (*PriceOverride, error)
	DeleteSubsitePrice(ctx context.Context, subsiteID, priceID int64) error
	GroupRateMultiplier(ctx context.Context, groupID int64) (float64, error)
	GroupImagePrices(ctx context.Context, groupID int64) (ImagePriceTiers, error)
}

// SubsiteSettings is the display/branding settings a sub-site admin may edit.
// Slug, domain and status stay under main-site control.
type SubsiteSettings struct {
	Slug       string `json:"slug"`
	Domain     string `json:"domain"`
	Name       string `json:"name"`
	LogoURL    string `json:"logo_url"`
	ThemeColor string `json:"theme_color"`
	Status     string `json:"status"`
}

// ScopedAdminHandler serves price management for the current sub-site only.
type ScopedAdminHandler struct {
	repo    ScopedPriceRepository
	catalog ModelCatalogProvider
}

// NewScopedAdminHandler creates a sub-site-scoped admin handler.
func NewScopedAdminHandler(repo ScopedPriceRepository, catalog ModelCatalogProvider) *ScopedAdminHandler {
	return &ScopedAdminHandler{repo: repo, catalog: catalog}
}

// ListPrices returns the price overrides of the current sub-site.
func (h *ScopedAdminHandler) ListPrices(c *gin.Context) {
	subsiteID, ok := h.authorize(c)
	if !ok {
		return
	}
	prices, err := h.repo.ListSubsitePrices(c.Request.Context(), subsiteID)
	if err != nil {
		response.InternalError(c, "failed to list subsite prices")
		return
	}
	response.Success(c, prices)
}

// ListChannels returns the main-site channels available to the current sub-site
// together with the sub-site's own override, if any.
func (h *ScopedAdminHandler) ListChannels(c *gin.Context) {
	subsiteID, ok := h.authorize(c)
	if !ok {
		return
	}
	channels, err := h.repo.ListSubsiteChannels(c.Request.Context(), subsiteID)
	if err != nil {
		response.InternalError(c, "failed to list subsite channels")
		return
	}
	response.Success(c, channels)
}

// ListModels returns the "group -> model" price list for the current sub-site:
// the main-site multiplier of every model, plus the sub-site's own override.
func (h *ScopedAdminHandler) ListModels(c *gin.Context) {
	subsiteID, ok := h.authorize(c)
	if !ok {
		return
	}
	if h.catalog == nil {
		response.Success(c, []SubsiteModelPrice{})
		return
	}
	catalog, err := h.catalog(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to load model catalog")
		return
	}
	models, err := h.repo.ListSubsiteModelPrices(c.Request.Context(), subsiteID, catalog)
	if err != nil {
		response.InternalError(c, "failed to load subsite model prices")
		return
	}
	response.Success(c, models)
}

// ListUsers returns the members of the current sub-site with scoped totals.
func (h *ScopedAdminHandler) ListUsers(c *gin.Context) {
	subsiteID, ok := h.authorize(c)
	if !ok {
		return
	}
	page, pageSize := parsePageQuery(c, 50, 200)
	users, err := h.repo.ListSubsiteMembers(c.Request.Context(), subsiteID, pageSize, (page-1)*pageSize)
	if err != nil {
		response.InternalError(c, "failed to list subsite members")
		return
	}
	response.Success(c, gin.H{"items": users, "page": page, "page_size": pageSize})
}

// parsePageQuery reads page/page_size with sane bounds.
func parsePageQuery(c *gin.Context, defaultSize, maxSize int) (int, int) {
	page := 1
	size := defaultSize
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			size = parsed
		}
	}
	if size > maxSize {
		size = maxSize
	}
	return page, size
}

// GetSettings returns the editable branding settings of the current sub-site.
func (h *ScopedAdminHandler) GetSettings(c *gin.Context) {
	subsiteID, ok := h.authorize(c)
	if !ok {
		return
	}
	subsite, err := h.repo.GetSubsiteByID(c.Request.Context(), subsiteID)
	if err != nil {
		writeAdminError(c, err, "failed to load subsite settings")
		return
	}
	response.Success(c, SubsiteSettings{
		Slug:       subsite.Slug,
		Domain:     subsite.Domain,
		Name:       subsite.Name,
		LogoURL:    subsite.LogoURL,
		ThemeColor: subsite.ThemeColor,
		Status:     subsite.Status,
	})
}

// UpdateSettings updates the branding fields a sub-site admin owns. Slug,
// domain, status and owner are preserved from the stored record so a sub-site
// admin can never change its own routing or lifecycle.
func (h *ScopedAdminHandler) UpdateSettings(c *gin.Context) {
	subsiteID, ok := h.authorize(c)
	if !ok {
		return
	}
	var req struct {
		Name       string `json:"name"`
		LogoURL    string `json:"logo_url"`
		ThemeColor string `json:"theme_color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid subsite settings payload")
		return
	}
	current, err := h.repo.GetSubsiteByID(c.Request.Context(), subsiteID)
	if err != nil {
		writeAdminError(c, err, "failed to load subsite settings")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = current.Name
	}
	updated, err := h.repo.UpdateSubsite(c.Request.Context(), subsiteID, SubsiteUpsert{
		Slug:        current.Slug,
		Domain:      current.Domain,
		Name:        name,
		LogoURL:     strings.TrimSpace(req.LogoURL),
		ThemeColor:  strings.TrimSpace(req.ThemeColor),
		Status:      current.Status,
		AdminUserID: current.AdminUserID,
	})
	if err != nil {
		writeAdminError(c, err, "failed to update subsite settings")
		return
	}
	response.Success(c, SubsiteSettings{
		Slug:       updated.Slug,
		Domain:     updated.Domain,
		Name:       updated.Name,
		LogoURL:    updated.LogoURL,
		ThemeColor: updated.ThemeColor,
		Status:     updated.Status,
	})
}

// CreatePrice adds a price override to the current sub-site.
func (h *ScopedAdminHandler) CreatePrice(c *gin.Context) {
	subsiteID, ok := h.authorize(c)
	if !ok {
		return
	}
	var req priceOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid price override payload")
		return
	}
	if !h.validateFloor(c, req) {
		return
	}
	price, err := h.repo.CreateSubsitePrice(c.Request.Context(), subsiteID, req.toUpsert())
	if err != nil {
		writeAdminError(c, err, "failed to create price override")
		return
	}
	response.Created(c, price)
}

// UpdatePrice updates one price override of the current sub-site.
func (h *ScopedAdminHandler) UpdatePrice(c *gin.Context) {
	subsiteID, ok := h.authorize(c)
	if !ok {
		return
	}
	priceID, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	var req priceOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid price override payload")
		return
	}
	if !h.validateFloor(c, req) {
		return
	}
	price, err := h.repo.UpdateSubsitePrice(c.Request.Context(), subsiteID, priceID, req.toUpsert())
	if err != nil {
		writeAdminError(c, err, "failed to update price override")
		return
	}
	response.Success(c, price)
}

// DeletePrice removes one price override of the current sub-site.
func (h *ScopedAdminHandler) DeletePrice(c *gin.Context) {
	subsiteID, ok := h.authorize(c)
	if !ok {
		return
	}
	priceID, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	if err := h.repo.DeleteSubsitePrice(c.Request.Context(), subsiteID, priceID); err != nil {
		writeAdminError(c, err, "failed to delete price override")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

// authorize resolves the current sub-site and requires the caller to be an
// owner/admin of that same sub-site.
func (h *ScopedAdminHandler) authorize(c *gin.Context) (int64, bool) {
	if h == nil || h.repo == nil {
		response.InternalError(c, "downstream management repository is not configured")
		return 0, false
	}
	subsite, ok := FromGin(c)
	if !ok || subsite == nil {
		response.NotFound(c, "downstream subsite context is required")
		return 0, false
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication is required")
		return 0, false
	}
	summary, err := h.repo.GetSubsiteUserSummary(c.Request.Context(), subsite.ID, userID)
	if err != nil {
		if errors.Is(err, ErrSubsiteUserNotFound) {
			response.NotFound(c, "downstream user not found")
			return 0, false
		}
		response.InternalError(c, "failed to load downstream user summary")
		return 0, false
	}
	if !summary.IsAdmin() {
		response.Forbidden(c, "downstream admin access required")
		return 0, false
	}
	return subsite.ID, true
}
