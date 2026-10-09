package downstream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type adminManagementRepoStub struct {
	subsites       []Subsite
	prices         []PriceOverride
	createdSubsite *SubsiteUpsert
	updatedSubsite *SubsiteUpsert
	createdPrice   *PriceOverrideUpsert
	updatedPrice   *PriceOverrideUpsert
}

func (s *adminManagementRepoStub) ListSubsites(context.Context) ([]Subsite, error) {
	return s.subsites, nil
}

func (s *adminManagementRepoStub) GetSubsiteByID(_ context.Context, id int64) (*Subsite, error) {
	return &Subsite{ID: id, Slug: "draw", Domain: "draw.superai.sbs", Name: "Draw", Status: SubsiteStatusActive}, nil
}

func (s *adminManagementRepoStub) CreateSubsite(_ context.Context, input SubsiteUpsert) (*Subsite, error) {
	s.createdSubsite = &input
	return &Subsite{
		ID: 7, Slug: input.Slug, Domain: input.Domain, Name: input.Name,
		LogoURL: input.LogoURL, ThemeColor: input.ThemeColor, Status: input.Status,
		AdminUserID: input.AdminUserID,
	}, nil
}

func (s *adminManagementRepoStub) UpdateSubsite(_ context.Context, id int64, input SubsiteUpsert) (*Subsite, error) {
	s.updatedSubsite = &input
	return &Subsite{
		ID: id, Slug: input.Slug, Domain: input.Domain, Name: input.Name,
		LogoURL: input.LogoURL, ThemeColor: input.ThemeColor, Status: input.Status,
		AdminUserID: input.AdminUserID,
	}, nil
}

func (s *adminManagementRepoStub) DisableSubsite(_ context.Context, id int64) (*Subsite, error) {
	return &Subsite{ID: id, Slug: "draw", Domain: "draw.superai.sbs", Name: "Draw", Status: SubsiteStatusDisabled}, nil
}

func (s *adminManagementRepoStub) ListSubsitePrices(context.Context, int64) ([]PriceOverride, error) {
	return s.prices, nil
}

func (s *adminManagementRepoStub) CreateSubsitePrice(_ context.Context, _ int64, input PriceOverrideUpsert) (*PriceOverride, error) {
	s.createdPrice = &input
	return &PriceOverride{ID: 9, SubsiteID: 7, Scope: input.Scope, GroupID: input.GroupID, Model: input.Model, Status: input.Status}, nil
}

func (s *adminManagementRepoStub) UpdateSubsitePrice(_ context.Context, _, priceID int64, input PriceOverrideUpsert) (*PriceOverride, error) {
	s.updatedPrice = &input
	return &PriceOverride{ID: priceID, SubsiteID: 7, Scope: input.Scope, GroupID: input.GroupID, Model: input.Model, Status: input.Status}, nil
}

func (s *adminManagementRepoStub) DeleteSubsitePrice(context.Context, int64, int64) error {
	return nil
}

func TestDownstreamAdminRoutesRequireAdminAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	adminGroup := router.Group("/api/v1/admin")
	adminGroup.Use(func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "Authorization required"})
			return
		}
		c.Next()
	})
	RegisterAdminRoutes(adminGroup, &adminManagementRepoStub{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/downstream/subsites", nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/downstream/subsites", nil)
	req.Header.Set("Authorization", "Bearer admin")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestDownstreamAdminCreatesAndUpdatesSubsiteAndPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &adminManagementRepoStub{}
	router := gin.New()
	adminGroup := router.Group("/api/v1/admin")
	adminGroup.Use(func(c *gin.Context) { c.Next() })
	RegisterAdminRoutes(adminGroup, repo)

	body := `{"slug":"draw","domain":"draw.superai.sbs","name":"Draw","logo_url":"https://cdn.example/logo.png","theme_color":"#0ea5e9","status":"active","admin_user_id":42}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/downstream/subsites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, repo.createdSubsite)
	require.Equal(t, "draw", repo.createdSubsite.Slug)
	require.Equal(t, int64(42), *repo.createdSubsite.AdminUserID)

	body = `{"slug":"draw","domain":"draw.superai.sbs","name":"Draw Studio","logo_url":"","theme_color":"#111827","status":"disabled","admin_user_id":43}`
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/downstream/subsites/7", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, repo.updatedSubsite)
	require.Equal(t, "Draw Studio", repo.updatedSubsite.Name)
	require.Equal(t, SubsiteStatusDisabled, repo.updatedSubsite.Status)

	priceBody := `{"scope":"model","model":"gpt-image-2","per_request_price":0.08,"status":"active"}`
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/downstream/subsites/7/prices", strings.NewReader(priceBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, repo.createdPrice)
	require.Equal(t, PriceOverrideScopeModel, repo.createdPrice.Scope)
	require.NotNil(t, repo.createdPrice.Model)
	require.Equal(t, "gpt-image-2", *repo.createdPrice.Model)
	require.NotNil(t, repo.createdPrice.PerRequestPrice)
	require.InDelta(t, 0.08, *repo.createdPrice.PerRequestPrice, 1e-12)

	priceBody = `{"scope":"group","group_id":5,"rate_multiplier":1.5,"status":"active"}`
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/downstream/subsites/7/prices/9", strings.NewReader(priceBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, repo.updatedPrice)
	require.Equal(t, PriceOverrideScopeGroup, repo.updatedPrice.Scope)
	require.NotNil(t, repo.updatedPrice.GroupID)
	require.Equal(t, int64(5), *repo.updatedPrice.GroupID)
	require.NotNil(t, repo.updatedPrice.RateMultiplier)
	require.InDelta(t, 1.5, *repo.updatedPrice.RateMultiplier, 1e-12)

}
