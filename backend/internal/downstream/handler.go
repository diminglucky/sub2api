package downstream

import (
	"context"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// SummaryRepository is the read contract used by the internal downstream API.
// Implementations must scope every method by subsiteID.
type SummaryRepository interface {
	GetSubsiteUserSummary(ctx context.Context, subsiteID, userID int64) (*SubsiteUserSummary, error)
	GetSubsiteAdminSummary(ctx context.Context, subsiteID, userID int64) (*SubsiteAdminSummary, error)
}

// SiteConfig is the public configuration returned to the downstream frontend.
type SiteConfig struct {
	ID         int64  `json:"id"`
	Slug       string `json:"slug"`
	Domain     string `json:"domain"`
	Name       string `json:"name"`
	LogoURL    string `json:"logo_url"`
	ThemeColor string `json:"theme_color"`
	APIBaseURL string `json:"api_base_url"`
}

// Handler serves the internal downstream API.
type Handler struct {
	repo SummaryRepository
}

// NewHandler creates a downstream API handler.
func NewHandler(repo SummaryRepository) *Handler {
	return &Handler{repo: repo}
}

// Site returns the resolved subsite brand and OpenAI-compatible API base URL.
func (h *Handler) Site(c *gin.Context) {
	subsite, ok := FromGin(c)
	if !ok {
		response.NotFound(c, "downstream subsite context is required")
		return
	}
	response.Success(c, SiteConfig{
		ID:         subsite.ID,
		Slug:       subsite.Slug,
		Domain:     subsite.Domain,
		Name:       subsite.Name,
		LogoURL:    subsite.LogoURL,
		ThemeColor: subsite.ThemeColor,
		APIBaseURL: "https://" + subsite.Domain + "/v1",
	})
}

// MeSummary returns the authenticated user's shared account data plus only the
// usage and recharge rows attributed to the current subsite.
func (h *Handler) MeSummary(c *gin.Context) {
	subsite, ok := FromGin(c)
	if !ok {
		response.NotFound(c, "downstream subsite context is required")
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication is required")
		return
	}
	if h == nil || h.repo == nil {
		response.InternalError(c, "downstream repository is not configured")
		return
	}

	summary, err := h.repo.GetSubsiteUserSummary(c.Request.Context(), subsite.ID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSubsiteUserNotFound):
			response.NotFound(c, "downstream user not found")
		default:
			response.InternalError(c, "failed to load downstream user summary")
		}
		return
	}
	response.Success(c, summary)
}

// AdminSummary returns scoped dashboard aggregates after verifying that the
// authenticated user is an owner/admin of the current subsite.
func (h *Handler) AdminSummary(c *gin.Context) {
	subsite, ok := FromGin(c)
	if !ok {
		response.NotFound(c, "downstream subsite context is required")
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication is required")
		return
	}
	if h == nil || h.repo == nil {
		response.InternalError(c, "downstream repository is not configured")
		return
	}

	summary, err := h.repo.GetSubsiteAdminSummary(c.Request.Context(), subsite.ID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSubsiteAdminRequired):
			response.Forbidden(c, "downstream admin access required")
		case errors.Is(err, ErrSubsiteUserNotFound):
			response.NotFound(c, "downstream user not found")
		default:
			response.InternalError(c, "failed to load downstream admin summary")
		}
		return
	}
	response.Success(c, summary)
}

func userIDFromContext(c *gin.Context) (int64, bool) {
	if c == nil {
		return 0, false
	}
	value, ok := c.Get(ContextKeyUserID)
	if !ok {
		return 0, false
	}
	userID, ok := value.(int64)
	return userID, ok && userID > 0
}
