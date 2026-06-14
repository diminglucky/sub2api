package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newRegionBlockTestRouter(cfg config.RegionBlockConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RegionBlock(cfg))
	r.GET("/home", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestRegionBlock_BlockedCountryReturnsHTMLPage(t *testing.T) {
	r := newRegionBlockTestRouter(config.RegionBlockConfig{
		Enabled:          true,
		Hosts:            []string{"superai.dihappy.cfd"},
		BlockedCountries: []string{"CN", "HK", "MO", "TW"},
		HeaderNames:      []string{"CF-IPCountry"},
		SupportEmail:     "support@example.com",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	req.Host = "superai.dihappy.cfd"
	req.Header.Set("Accept", "text/html")
	req.Header.Set("CF-IPCountry", "CN")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Header().Get("Content-Type"), "text/html")
	require.Equal(t, "CN", w.Header().Get("X-Region-Blocked"))
	require.Contains(t, w.Body.String(), "暂不支持你所在的地区")
	require.Contains(t, w.Body.String(), "中国大陆、中国香港、中国澳门、中国台湾暂无法使用")
	require.Contains(t, w.Body.String(), "support@example.com")
}

func TestRegionBlock_AllowsAPIRouteOnBlockedHostAndCountry(t *testing.T) {
	r := newRegionBlockTestRouter(config.RegionBlockConfig{
		Enabled:          true,
		Hosts:            []string{"superai.dihappy.cfd"},
		BlockedCountries: []string{"CN"},
		HeaderNames:      []string{"CF-IPCountry"},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Host = "superai.dihappy.cfd"
	req.Header.Set("CF-IPCountry", "cn")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"ok":true}`, w.Body.String())
}

func TestRegionBlock_AllowsUnblockedAndMissingCountries(t *testing.T) {
	r := newRegionBlockTestRouter(config.RegionBlockConfig{
		Enabled:          true,
		Hosts:            []string{"superai.dihappy.cfd"},
		BlockedCountries: []string{"CN"},
		HeaderNames:      []string{"CF-IPCountry"},
	})

	for _, country := range []string{"US", ""} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/home", nil)
		req.Host = "superai.dihappy.cfd"
		if country != "" {
			req.Header.Set("CF-IPCountry", country)
		}
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "ok", w.Body.String())
	}
}

func TestRegionBlock_AllowsOtherHosts(t *testing.T) {
	r := newRegionBlockTestRouter(config.RegionBlockConfig{
		Enabled:          true,
		Hosts:            []string{"superai.dihappy.cfd"},
		BlockedCountries: []string{"CN"},
		HeaderNames:      []string{"CF-IPCountry"},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	req.Host = "api.dihappy.cfd"
	req.Header.Set("Accept", "text/html")
	req.Header.Set("CF-IPCountry", "CN")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", w.Body.String())
}

func TestRegionBlock_DisabledAllowsBlockedCountry(t *testing.T) {
	r := newRegionBlockTestRouter(config.RegionBlockConfig{
		Enabled:          false,
		Hosts:            []string{"superai.dihappy.cfd"},
		BlockedCountries: []string{"CN"},
		HeaderNames:      []string{"CF-IPCountry"},
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	req.Host = "superai.dihappy.cfd"
	req.Header.Set("CF-IPCountry", "CN")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", w.Body.String())
}
