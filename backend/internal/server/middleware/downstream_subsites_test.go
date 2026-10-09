package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// stubSubsiteResolver is a deterministic resolver for middleware tests. It only
// knows about the draw pilot subsite and reports every other slug as missing.
type stubSubsiteResolver struct {
	subsites map[string]*downstream.Subsite
	err      error
}

func (s *stubSubsiteResolver) GetSubsiteBySlug(_ context.Context, slug string) (*downstream.Subsite, error) {
	if s.err != nil {
		return nil, s.err
	}
	normalized := downstream.NormalizeSlug(slug)
	if subsite, ok := s.subsites[normalized]; ok {
		return subsite, nil
	}
	return nil, downstream.ErrSubsiteNotFound
}

func drawSubsite() *downstream.Subsite {
	return &downstream.Subsite{
		ID:         7,
		Slug:       "draw",
		Domain:     "draw.superai.sbs",
		Name:       "Draw",
		ThemeColor: "#0ea5e9",
		Status:     downstream.SubsiteStatusActive,
	}
}

func newDrawResolver() *stubSubsiteResolver {
	return &stubSubsiteResolver{
		subsites: map[string]*downstream.Subsite{"draw": drawSubsite()},
	}
}

// newDownstreamSubsiteTestRouter wires the middleware the same way the router
// will in Task 3: globally, ahead of the gateway and page handlers.
func newDownstreamSubsiteTestRouter(resolver *stubSubsiteResolver) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(DownstreamSubsite(resolver, "superai.sbs"))
	r.GET("/home", func(c *gin.Context) {
		subsite, ok := downstream.FromGin(c)
		if !ok {
			c.JSON(http.StatusOK, gin.H{"scoped": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"scoped": true, "slug": subsite.Slug, "id": subsite.ID})
	})
	r.GET("/v1/models", func(c *gin.Context) {
		// The OpenAI-compatible gateway payload must be untouched by resolution.
		_, _ = downstream.FromGin(c)
		c.JSON(http.StatusOK, gin.H{"object": "list", "data": []any{}})
	})
	return r
}

func serveDraw(r *gin.Engine, method, path, host string, headers map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	req.Host = host
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestDownstreamSubsiteMiddlewareResolvesDraw(t *testing.T) {
	r := newDownstreamSubsiteTestRouter(newDrawResolver())

	w := serveDraw(r, http.MethodGet, "/home", "draw.superai.sbs", nil)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"scoped":true,"slug":"draw","id":7}`, w.Body.String())
}

func TestDownstreamSubsiteMiddlewarePrefersTrustedSlugHeader(t *testing.T) {
	r := newDownstreamSubsiteTestRouter(newDrawResolver())

	// The edge sets X-Downstream-Slug; it must win over the host label.
	w := serveDraw(r, http.MethodGet, "/home", "superai.sbs", map[string]string{
		DownstreamSlugHeader: " DRAW ",
	})

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"scoped":true,"slug":"draw","id":7}`, w.Body.String())
}

func TestDownstreamSubsiteMiddlewareIgnoresApexHost(t *testing.T) {
	r := newDownstreamSubsiteTestRouter(newDrawResolver())

	w := serveDraw(r, http.MethodGet, "/home", "superai.sbs", nil)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"scoped":false}`, w.Body.String())
}

func TestDownstreamSubsiteMiddlewareReturns404ForUnknownSlug(t *testing.T) {
	r := newDownstreamSubsiteTestRouter(newDrawResolver())

	w := serveDraw(r, http.MethodGet, "/home", "missing.superai.sbs", nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "SUBSITE_NOT_FOUND")
}

func TestDownstreamSubsiteMiddlewareReturns404ForDisabledSubsite(t *testing.T) {
	disabled := drawSubsite()
	disabled.Status = downstream.SubsiteStatusDisabled
	resolver := &stubSubsiteResolver{subsites: map[string]*downstream.Subsite{"draw": disabled}}
	r := newDownstreamSubsiteTestRouter(resolver)

	w := serveDraw(r, http.MethodGet, "/home", "draw.superai.sbs", nil)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "SUBSITE_NOT_FOUND")
}

// TestDownstreamSubsiteMiddlewareKeepsV1GatewayBehavior pins that resolution
// attaches context for /v1/* without touching the OpenAI-compatible payload.
func TestDownstreamSubsiteMiddlewareKeepsV1GatewayBehavior(t *testing.T) {
	r := newDownstreamSubsiteTestRouter(newDrawResolver())

	w := serveDraw(r, http.MethodGet, "/v1/models", "draw.superai.sbs", nil)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"object":"list","data":[]}`, w.Body.String())
}

func TestDownstreamSubsiteMiddlewarePassthroughWhenDisabled(t *testing.T) {
	// A nil resolver or an empty base domain disables resolution entirely so
	// main-site behavior stays intact.
	cases := []struct {
		name     string
		resolver DownstreamSubsiteResolver
		base     string
	}{
		{name: "nil resolver", resolver: nil, base: "superai.sbs"},
		{name: "empty base domain", resolver: newDrawResolver(), base: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(DownstreamSubsite(tc.resolver, tc.base))
			r.GET("/home", func(c *gin.Context) {
				_, ok := downstream.FromGin(c)
				c.JSON(http.StatusOK, gin.H{"scoped": ok})
			})

			w := serveDraw(r, http.MethodGet, "/home", "draw.superai.sbs", nil)
			require.Equal(t, http.StatusOK, w.Code)
			require.JSONEq(t, `{"scoped":false}`, w.Body.String())
		})
	}
}

func TestDownstreamSubsiteMiddlewareSkipsUnrelatedHost(t *testing.T) {
	r := newDownstreamSubsiteTestRouter(newDrawResolver())

	for _, host := range []string{"localhost", "127.0.0.1", "other.example.com", "superai.sbs:8080"} {
		w := serveDraw(r, http.MethodGet, "/home", host, nil)
		require.Equal(t, http.StatusOK, w.Code, host)
		require.JSONEq(t, `{"scoped":false}`, w.Body.String(), host)
	}
}

func TestDownstreamSubsiteMiddlewareResolverFailureIs500(t *testing.T) {
	resolver := &stubSubsiteResolver{err: sql.ErrConnDone}
	r := newDownstreamSubsiteTestRouter(resolver)

	w := serveDraw(r, http.MethodGet, "/home", "draw.superai.sbs", nil)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "SUBSITE_RESOLUTION_FAILED")
}
