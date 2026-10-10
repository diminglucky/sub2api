package downstream

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// AdminManagementRepository is the main-site admin control-plane contract for
// provisioning subsites and their price overrides.
type AdminManagementRepository interface {
	ListSubsites(ctx context.Context) ([]Subsite, error)
	GetSubsiteByID(ctx context.Context, id int64) (*Subsite, error)
	CreateSubsite(ctx context.Context, input SubsiteUpsert) (*Subsite, error)
	UpdateSubsite(ctx context.Context, id int64, input SubsiteUpsert) (*Subsite, error)
	DisableSubsite(ctx context.Context, id int64) (*Subsite, error)
	ListSubsitePrices(ctx context.Context, subsiteID int64) ([]PriceOverride, error)
	CreateSubsitePrice(ctx context.Context, subsiteID int64, input PriceOverrideUpsert) (*PriceOverride, error)
	UpdateSubsitePrice(ctx context.Context, subsiteID, priceID int64, input PriceOverrideUpsert) (*PriceOverride, error)
	DeleteSubsitePrice(ctx context.Context, subsiteID, priceID int64) error
	ListSubsiteGroupAssignments(ctx context.Context, subsiteID int64) ([]SubsiteChannel, error)
	AddSubsiteGroup(ctx context.Context, subsiteID, groupID int64) error
	RemoveSubsiteGroup(ctx context.Context, subsiteID, groupID int64) error
}

// AdminHandler serves main-site admin provisioning for downstream subsites.
type AdminHandler struct {
	repo AdminManagementRepository
}

func NewAdminHandler(repo AdminManagementRepository) *AdminHandler {
	return &AdminHandler{repo: repo}
}

type subsiteRequest struct {
	Slug        string `json:"slug"`
	Domain      string `json:"domain"`
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	ThemeColor  string `json:"theme_color"`
	Status      string `json:"status"`
	AdminUserID *int64 `json:"admin_user_id"`
}

type priceOverrideRequest struct {
	Scope           string   `json:"scope"`
	GroupID         *int64   `json:"group_id"`
	Model           *string  `json:"model"`
	RateMultiplier  *float64 `json:"rate_multiplier"`
	InputPrice      *float64 `json:"input_price"`
	OutputPrice     *float64 `json:"output_price"`
	CacheWritePrice *float64 `json:"cache_write_price"`
	CacheReadPrice  *float64 `json:"cache_read_price"`
	PerRequestPrice *float64 `json:"per_request_price"`
	ImagePrice1K    *float64 `json:"image_price_1k"`
	ImagePrice2K    *float64 `json:"image_price_2k"`
	ImagePrice4K    *float64 `json:"image_price_4k"`
	Status          string   `json:"status"`
}

func (h *AdminHandler) ListSubsites(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	subsites, err := h.repo.ListSubsites(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to list subsites")
		return
	}
	response.Success(c, subsites)
}

func (h *AdminHandler) GetSubsite(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	subsite, err := h.repo.GetSubsiteByID(c.Request.Context(), id)
	if err != nil {
		writeAdminError(c, err, "failed to load subsite")
		return
	}
	response.Success(c, subsite)
}

func (h *AdminHandler) CreateSubsite(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	var req subsiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid subsite payload")
		return
	}
	subsite, err := h.repo.CreateSubsite(c.Request.Context(), req.toUpsert())
	if err != nil {
		writeAdminError(c, err, "failed to create subsite")
		return
	}
	response.Created(c, subsite)
}

func (h *AdminHandler) UpdateSubsite(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	var req subsiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid subsite payload")
		return
	}
	subsite, err := h.repo.UpdateSubsite(c.Request.Context(), id, req.toUpsert())
	if err != nil {
		writeAdminError(c, err, "failed to update subsite")
		return
	}
	response.Success(c, subsite)
}

func (h *AdminHandler) DisableSubsite(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	subsite, err := h.repo.DisableSubsite(c.Request.Context(), id)
	if err != nil {
		writeAdminError(c, err, "failed to disable subsite")
		return
	}
	response.Success(c, subsite)
}

func (h *AdminHandler) ListPrices(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	subsiteID, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	prices, err := h.repo.ListSubsitePrices(c.Request.Context(), subsiteID)
	if err != nil {
		writeAdminError(c, err, "failed to list subsite prices")
		return
	}
	response.Success(c, prices)
}

func (h *AdminHandler) CreatePrice(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	subsiteID, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	var req priceOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid price override payload")
		return
	}
	price, err := h.repo.CreateSubsitePrice(c.Request.Context(), subsiteID, req.toUpsert())
	if err != nil {
		writeAdminError(c, err, "failed to create price override")
		return
	}
	response.Created(c, price)
}

func (h *AdminHandler) UpdatePrice(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	subsiteID, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	priceID, ok := parsePositiveID(c, "price_id")
	if !ok {
		return
	}
	var req priceOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid price override payload")
		return
	}
	price, err := h.repo.UpdateSubsitePrice(c.Request.Context(), subsiteID, priceID, req.toUpsert())
	if err != nil {
		writeAdminError(c, err, "failed to update price override")
		return
	}
	response.Success(c, price)
}

func (h *AdminHandler) DeletePrice(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	subsiteID, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	priceID, ok := parsePositiveID(c, "price_id")
	if !ok {
		return
	}
	if err := h.repo.DeleteSubsitePrice(c.Request.Context(), subsiteID, priceID); err != nil {
		writeAdminError(c, err, "failed to delete price override")
		return
	}
	response.Success(c, gin.H{"deleted": true})
}

// ListSubsiteGroups returns every active group with its assignment state.
func (h *AdminHandler) ListSubsiteGroups(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	groups, err := h.repo.ListSubsiteGroupAssignments(c.Request.Context(), id)
	if err != nil {
		writeAdminError(c, err, "failed to list subsite groups")
		return
	}
	response.Success(c, groups)
}

// AssignSubsiteGroup opens a group to the sub-site.
func (h *AdminHandler) AssignSubsiteGroup(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	groupID, ok := parsePositiveID(c, "group_id")
	if !ok {
		return
	}
	if err := h.repo.AddSubsiteGroup(c.Request.Context(), id, groupID); err != nil {
		writeAdminError(c, err, "failed to assign group")
		return
	}
	response.Success(c, gin.H{"assigned": true})
}

// UnassignSubsiteGroup closes a group for the sub-site.
func (h *AdminHandler) UnassignSubsiteGroup(c *gin.Context) {
	if !h.requireRepo(c) {
		return
	}
	id, ok := parsePositiveID(c, "id")
	if !ok {
		return
	}
	groupID, ok := parsePositiveID(c, "group_id")
	if !ok {
		return
	}
	if err := h.repo.RemoveSubsiteGroup(c.Request.Context(), id, groupID); err != nil {
		writeAdminError(c, err, "failed to unassign group")
		return
	}
	response.Success(c, gin.H{"assigned": false})
}

func (h *AdminHandler) requireRepo(c *gin.Context) bool {
	if h != nil && h.repo != nil {
		return true
	}
	response.InternalError(c, "downstream management repository is not configured")
	return false
}

func (r subsiteRequest) toUpsert() SubsiteUpsert {
	return SubsiteUpsert{
		Slug:        r.Slug,
		Domain:      r.Domain,
		Name:        r.Name,
		LogoURL:     r.LogoURL,
		ThemeColor:  r.ThemeColor,
		Status:      r.Status,
		AdminUserID: r.AdminUserID,
	}
}

func (r priceOverrideRequest) toUpsert() PriceOverrideUpsert {
	return PriceOverrideUpsert{
		Scope:           r.Scope,
		GroupID:         r.GroupID,
		Model:           r.Model,
		RateMultiplier:  r.RateMultiplier,
		InputPrice:      r.InputPrice,
		OutputPrice:     r.OutputPrice,
		CacheWritePrice: r.CacheWritePrice,
		CacheReadPrice:  r.CacheReadPrice,
		PerRequestPrice: r.PerRequestPrice,
		ImagePrice1K:    r.ImagePrice1K,
		ImagePrice2K:    r.ImagePrice2K,
		ImagePrice4K:    r.ImagePrice4K,
		Status:          r.Status,
	}
}

func parsePositiveID(c *gin.Context, name string) (int64, bool) {
	raw := strings.TrimSpace(c.Param(name))
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return id, true
}

func writeAdminError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, ErrSubsiteNotFound):
		response.NotFound(c, "subsite not found")
	case strings.Contains(strings.ToLower(err.Error()), "invalid"),
		strings.Contains(strings.ToLower(err.Error()), "required"),
		strings.Contains(strings.ToLower(err.Error()), "unsupported"):
		response.BadRequest(c, err.Error())
	default:
		response.InternalError(c, fallback)
	}
}
