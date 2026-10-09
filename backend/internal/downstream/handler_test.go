package downstream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type handlerRepoStub struct {
	userSummary       *SubsiteUserSummary
	adminSummary      *SubsiteAdminSummary
	userSummaryErr    error
	adminSummaryErr   error
	gotUserSubsiteID  int64
	gotAdminSubsiteID int64
	gotUserID         int64
}

func (s *handlerRepoStub) GetSubsiteUserSummary(_ context.Context, subsiteID, userID int64) (*SubsiteUserSummary, error) {
	s.gotUserSubsiteID = subsiteID
	s.gotUserID = userID
	return s.userSummary, s.userSummaryErr
}

func (s *handlerRepoStub) GetSubsiteAdminSummary(_ context.Context, subsiteID, userID int64) (*SubsiteAdminSummary, error) {
	s.gotAdminSubsiteID = subsiteID
	s.gotUserID = userID
	return s.adminSummary, s.adminSummaryErr
}

func TestSiteHandlerReturnsDrawConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &handlerRepoStub{}
	h := NewHandler(repo)
	router := gin.New()
	router.GET("/api/internal/downstream/v1/site", func(c *gin.Context) {
		SetInGin(c, &Subsite{
			ID:         7,
			Slug:       "draw",
			Domain:     "draw.superai.sbs",
			Name:       "Draw",
			LogoURL:    "https://cdn.example/logo.png",
			ThemeColor: "#0ea5e9",
			Status:     SubsiteStatusActive,
		})
		h.Site(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/internal/downstream/v1/site", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var got struct {
		Data SiteConfig `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, int64(7), got.Data.ID)
	require.Equal(t, "draw", got.Data.Slug)
	require.Equal(t, "draw.superai.sbs", got.Data.Domain)
	require.Equal(t, "Draw", got.Data.Name)
	require.Equal(t, "https://draw.superai.sbs/v1", got.Data.APIBaseURL)
}

func TestSiteHandlerRequiresSubsiteContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&handlerRepoStub{})
	router := gin.New()
	router.GET("/site", h.Site)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/site", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestMeSummaryUsesScopedUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &handlerRepoStub{userSummary: &SubsiteUserSummary{
		UserID:         42,
		SubsiteID:      7,
		Email:          "u@example.com",
		Balance:        12.5,
		Role:           SubsiteMemberRoleMember,
		Source:         SubsiteMemberSourceRegistration,
		UsageCount:     3,
		UsageCost:      1.25,
		RechargeCount:  2,
		RechargeAmount: 20,
	}}
	h := NewHandler(repo)
	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		SetInGin(c, &Subsite{ID: 7, Slug: "draw", Domain: "draw.superai.sbs", Name: "Draw", Status: SubsiteStatusActive})
		c.Set(ContextKeyUserID, int64(42))
		h.MeSummary(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	require.Equal(t, int64(7), repo.gotUserSubsiteID)
	require.Equal(t, int64(42), repo.gotUserID)
	var got struct {
		Data SubsiteUserSummary `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, int64(7), got.Data.SubsiteID)
	require.Equal(t, int64(42), got.Data.UserID)
	require.False(t, got.Data.Admin)
}

func TestAdminSummaryRequiresSubsiteAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &handlerRepoStub{adminSummaryErr: ErrSubsiteAdminRequired}
	h := NewHandler(repo)
	router := gin.New()
	router.GET("/admin", func(c *gin.Context) {
		SetInGin(c, &Subsite{ID: 7, Slug: "draw", Domain: "draw.superai.sbs", Name: "Draw", Status: SubsiteStatusActive})
		c.Set(ContextKeyUserID, int64(42))
		h.AdminSummary(c)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	require.Equal(t, int64(7), repo.gotAdminSubsiteID)
	require.Equal(t, int64(42), repo.gotUserID)
}
