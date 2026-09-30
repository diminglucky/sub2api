package middleware

import (
	"html"
	"net"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
)

var fallbackRegionCountryHeaders = []string{
	"CF-IPCountry",
	"CloudFront-Viewer-Country",
	"X-Vercel-IP-Country",
	"Fly-Client-IPCountry",
	"X-Country-Code",
	"X-Geo-Country",
}

var searchEngineCrawlerUserAgentTokens = []string{
	"googlebot",
	"google-inspectiontool",
	"googleother",
	"bingbot",
	"msnbot",
	"duckduckbot",
	"applebot",
	"yandexbot",
}

// RegionBlock blocks page navigation requests from configured countries/regions.
// It relies on a trusted proxy/CDN such as Cloudflare to populate country
// headers. API routes are always allowed.
// resolve 每次请求都会调用，因此后台保存的设置立即生效；实现必须是廉价读取
// （server 包传入的是原子快照读取）。
func RegionBlock(resolve func() config.RegionBlockConfig) gin.HandlerFunc {
	if resolve == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		cfg := resolve()
		if !cfg.Enabled {
			c.Next()
			return
		}
		if !isRegionBlockPageRequest(c) || !regionBlockHostMatches(c.Request, blockedHostsFor(cfg.Hosts)) {
			c.Next()
			return
		}

		countryCode := requestCountryCode(c.Request, headerNamesFor(cfg.HeaderNames))
		if countryCode == "" {
			c.Next()
			return
		}
		if !isBlockedCountry(cfg.BlockedCountries, countryCode) {
			c.Next()
			return
		}

		c.Header("Cache-Control", "no-store")
		c.Header("X-Region-Blocked", countryCode)
		if c.Request.Method == http.MethodHead {
			c.Status(http.StatusForbidden)
		} else {
			c.Data(http.StatusForbidden, "text/html; charset=utf-8", []byte(renderUnsupportedRegionPage(countryCode, strings.TrimSpace(cfg.SupportEmail))))
		}
		c.Abort()
	}
}

func headerNamesFor(configured []string) []string {
	if len(configured) == 0 {
		return fallbackRegionCountryHeaders
	}
	return configured
}

func blockedHostsFor(hosts []string) map[string]struct{} {
	blockedHosts := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		normalized := normalizeRequestHost(host)
		if normalized != "" {
			blockedHosts[normalized] = struct{}{}
		}
	}
	return blockedHosts
}

func isBlockedCountry(configured []string, countryCode string) bool {
	for _, country := range configured {
		if strings.ToUpper(strings.TrimSpace(country)) == countryCode {
			return true
		}
	}
	return false
}

func requestCountryCode(req *http.Request, headerNames []string) string {
	if req == nil {
		return ""
	}
	for _, headerName := range headerNames {
		raw := req.Header.Get(strings.TrimSpace(headerName))
		if raw == "" {
			continue
		}
		for _, part := range strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == ' '
		}) {
			code := strings.ToUpper(strings.TrimSpace(part))
			if len(code) == 2 {
				return code
			}
		}
	}
	return ""
}

func isRegionBlockPageRequest(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return false
	}
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		return false
	}
	if isRegionBlockExemptPath(c.Request.URL.Path) {
		return false
	}
	if isSearchEngineCrawler(c.Request) && isRegionBlockPublicCrawlerPath(c.Request.URL.Path) {
		return false
	}
	accept := strings.ToLower(c.GetHeader("Accept"))
	return accept == "" || strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*")
}

func isSearchEngineCrawler(req *http.Request) bool {
	if req == nil {
		return false
	}
	ua := strings.ToLower(req.UserAgent())
	for _, token := range searchEngineCrawlerUserAgentTokens {
		if strings.Contains(ua, token) {
			return true
		}
	}
	return false
}

func isRegionBlockPublicCrawlerPath(path string) bool {
	switch path {
	case "/", "/home", "/model-plaza", "/key-usage":
		return true
	}
	return false
}

func regionBlockHostMatches(req *http.Request, blockedHosts map[string]struct{}) bool {
	if len(blockedHosts) == 0 {
		return true
	}
	host := normalizeRequestHost(requestHost(req))
	if host == "" {
		return false
	}
	if _, ok := blockedHosts[host]; ok {
		return true
	}
	if _, ok := blockedHosts["*"]; ok {
		return true
	}
	return false
}

func requestHost(req *http.Request) string {
	if req == nil {
		return ""
	}
	return req.Host
}

func normalizeRequestHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return ""
	}
	if strings.Contains(host, "://") {
		if parts := strings.SplitN(host, "://", 2); len(parts) == 2 {
			host = parts[1]
		}
	}
	if strings.Contains(host, "/") {
		host = strings.SplitN(host, "/", 2)[0]
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.Trim(host, "[]")
}

// isRegionBlockExemptPath 汇总不该被地区拦截的路径：API 路由，以及搜索引擎/浏览器
// 需要直接抓取的爬虫文件与静态资源。拦截 robots.txt / sitemap.xml 会妨碍收录，
// 与"只拦网页访问"的初衷也相悖。
func isRegionBlockExemptPath(path string) bool {
	if isRegionBlockAPIRoute(path) {
		return true
	}
	switch path {
	case "/robots.txt", "/sitemap.xml", "/site.webmanifest", "/favicon.ico", "/logo.svg", "/BingSiteAuth.xml":
		return true
	}
	return strings.HasPrefix(path, "/assets/")
}

func isRegionBlockAPIRoute(path string) bool {
	return strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/v1/") ||
		strings.HasPrefix(path, "/v1beta/") ||
		strings.HasPrefix(path, "/antigravity/") ||
		strings.HasPrefix(path, "/backend-api/") ||
		strings.HasPrefix(path, "/responses") ||
		strings.HasPrefix(path, "/chat/completions") ||
		strings.HasPrefix(path, "/embeddings") ||
		strings.HasPrefix(path, "/images") ||
		strings.HasPrefix(path, "/health") ||
		strings.HasPrefix(path, "/setup/")
}

func renderUnsupportedRegionPage(countryCode, supportEmail string) string {
	escapedCountry := html.EscapeString(countryCode)
	escapedEmail := html.EscapeString(supportEmail)
	supportButton := ""
	if escapedEmail != "" {
		supportButton = `<a class="primary" href="mailto:` + escapedEmail + `">联系支持 · ` + escapedEmail + `</a>`
	}
	return `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>暂不支持你所在的地区</title>
  <style>
    :root{color-scheme:dark;background:#151515;color:#f5f5f5;font-family:Inter,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}
    *{box-sizing:border-box}
    body{margin:0;min-height:100vh;background:radial-gradient(circle at 50% 0%,#262626 0,#171717 42%,#111 100%);display:flex;align-items:center;justify-content:center;padding:32px}
    main{width:min(760px,100%);text-align:center}
    .icon{width:88px;height:88px;margin:0 auto 34px;border-radius:28px;background:#232323;border:1px solid #333;display:grid;place-items:center;color:#a3a3a3;font-size:44px;box-shadow:0 24px 80px rgba(0,0,0,.28)}
    h1{margin:0;color:#fff;font-size:clamp(38px,6vw,68px);line-height:1.05;font-weight:800;letter-spacing:0}
    p{margin:26px auto 0;max-width:680px;color:#a3a3a3;font-size:clamp(18px,2.4vw,24px);line-height:1.65;font-weight:600}
    .meta{margin-top:20px;color:#737373;font-size:14px}
    .actions{margin-top:42px;display:flex;gap:14px;justify-content:center;flex-wrap:wrap}
    a,.ghost{min-height:44px;border-radius:999px;padding:10px 20px;text-decoration:none;font-size:16px;font-weight:700;display:inline-flex;align-items:center;justify-content:center}
    .primary{background:#ff5c16;color:#fff;box-shadow:0 12px 34px rgba(255,92,22,.28)}
    .ghost{border:1px solid #3f3f46;color:#e5e5e5;background:#1b1b1b}
  </style>
</head>
<body>
  <main>
    <div class="icon" aria-hidden="true">⊘</div>
    <h1>暂不支持你所在的地区</h1>
    <p>很遗憾，本服务目前仅在部分地区开放。中国大陆、中国香港、中国澳门、中国台湾暂无法使用。</p>
    <div class="meta">检测地区：` + escapedCountry + `</div>
    <div class="actions"><span class="ghost">如果你认为这是误判，请联系支持</span>` + supportButton + `</div>
  </main>
</body>
</html>`
}
