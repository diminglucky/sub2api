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
	r.Use(RegionBlock(func() config.RegionBlockConfig { return cfg }))
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

// 后台保存设置后中间件必须立刻按新配置放行/拦截，而不是沿用构造时的快照。
func TestRegionBlock_PicksUpUpdatedConfigWithoutRestart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.RegionBlockConfig{
		Enabled:          true,
		Hosts:            []string{"superai.dihappy.cfd"},
		BlockedCountries: []string{"CN"},
		HeaderNames:      []string{"CF-IPCountry"},
	}
	r := gin.New()
	r.Use(RegionBlock(func() config.RegionBlockConfig { return cfg }))
	r.GET("/home", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	probe := func() int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/home", nil)
		req.Host = "superai.dihappy.cfd"
		req.Header.Set("Accept", "text/html")
		req.Header.Set("CF-IPCountry", "CN")
		r.ServeHTTP(w, req)
		return w.Code
	}

	require.Equal(t, http.StatusForbidden, probe())

	cfg.BlockedCountries = []string{"US"}
	require.Equal(t, http.StatusOK, probe())
}

// 地区拦截只针对网页；搜索引擎需要抓取的 robots.txt / sitemap.xml 必须始终可达，
// 否则被屏蔽地区的爬虫连站点地图都拿不到。
func TestRegionBlock_AllowsCrawlerFilesFromBlockedCountry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RegionBlock(func() config.RegionBlockConfig {
		return config.RegionBlockConfig{
			Enabled:          true,
			Hosts:            []string{"superai.dihappy.cfd"},
			BlockedCountries: []string{"CN"},
			HeaderNames:      []string{"CF-IPCountry"},
		}
	}))
	r.GET("/robots.txt", func(c *gin.Context) { c.String(http.StatusOK, "User-agent: *") })
	r.GET("/sitemap.xml", func(c *gin.Context) { c.String(http.StatusOK, "<urlset/>") })
	r.GET("/BingSiteAuth.xml", func(c *gin.Context) { c.String(http.StatusOK, "<users/>") })
	r.GET("/home", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	probe := func(path, accept string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Host = "superai.dihappy.cfd"
		req.Header.Set("Accept", accept)
		req.Header.Set("CF-IPCountry", "CN")
		r.ServeHTTP(w, req)
		return w.Code
	}

	require.Equal(t, http.StatusOK, probe("/robots.txt", "*/*"))
	require.Equal(t, http.StatusOK, probe("/sitemap.xml", "*/*"))
	require.Equal(t, http.StatusOK, probe("/BingSiteAuth.xml", "*/*"))
	require.Equal(t, http.StatusForbidden, probe("/home", "text/html"))
}

// 搜索引擎出口节点可能被 CDN/GeoIP 判定为受限地区；公开 SEO 页面要允许爬虫抓取，
// 否则 sitemap 能读到，但条目页面仍会返回 403，收录会一直不稳定。
func TestRegionBlock_AllowsSearchEngineCrawlersOnPublicSEOPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RegionBlock(func() config.RegionBlockConfig {
		return config.RegionBlockConfig{
			Enabled:          true,
			Hosts:            []string{"superai.dihappy.cfd"},
			BlockedCountries: []string{"CN"},
			HeaderNames:      []string{"CF-IPCountry"},
		}
	}))
	for _, path := range []string{"/", "/home", "/model-plaza", "/key-usage"} {
		r.GET(path, func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	}
	r.GET("/dashboard", func(c *gin.Context) { c.String(http.StatusOK, "dashboard") })

	probe := func(path, userAgent string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Host = "superai.dihappy.cfd"
		req.Header.Set("Accept", "text/html")
		req.Header.Set("CF-IPCountry", "CN")
		if userAgent != "" {
			req.Header.Set("User-Agent", userAgent)
		}
		r.ServeHTTP(w, req)
		return w.Code
	}

	for _, path := range []string{"/", "/home", "/model-plaza", "/key-usage"} {
		require.Equal(t, http.StatusOK, probe(path, "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"), path)
	}
	require.Equal(t, http.StatusOK, probe("/home", "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)"))
	require.Equal(t, http.StatusForbidden, probe("/dashboard", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"))
	require.Equal(t, http.StatusForbidden, probe("/home", "Mozilla/5.0 AppleWebKit/537.36"))
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
