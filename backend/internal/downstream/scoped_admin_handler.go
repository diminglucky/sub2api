package downstream

import (
	"context"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// ScopedPriceRepository is the sub-site-scoped control-plane contract. It is
// satisfied by *Repository. Authorization is by sub-site membership role, not
// by main-site admin role, so a sub-site owner manages only its own prices.
type ScopedPriceRepository interface {
	GetSubsiteUserSummary(ctx context.Context, subsiteID, userID int64) (*SubsiteUserSummary, error)
	ListSubsitePrices(ctx context.Context, subsiteID int64) ([]PriceOverride, error)
	CreateSubsitePrice(ctx context.Context, subsiteID int64, input PriceOverrideUpsert) (*PriceOverride, error)
	UpdateSubsitePrice(ctx context.Context, subsiteID, priceID int64, input PriceOverrideUpsert) (*PriceOverride, error)
	DeleteSubsitePrice(ctx context.Context, subsiteID, priceID int64) error
}

// ScopedAdminHandler serves price management for the current sub-site only.
type ScopedAdminHandler struct {
	repo ScopedPriceRepository
}

// NewScopedAdminHandler creates a sub-site-scoped admin handler.
func NewScopedAdminHandler(repo ScopedPriceRepository) *ScopedAdminHandler {
	return &ScopedAdminHandler{repo: repo}
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
