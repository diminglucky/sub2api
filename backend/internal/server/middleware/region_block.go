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

// RegionBlock blocks page navigation requests from configured countries/regions.
// It relies on a trusted proxy/CDN such as Cloudflare to populate country
// headers. API routes are always allowed.
func RegionBlock(cfg config.RegionBlockConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	blockedCountries := make(map[string]struct{}, len(cfg.BlockedCountries))
	for _, country := range cfg.BlockedCountries {
		code := strings.ToUpper(strings.TrimSpace(country))
		if code != "" {
			blockedCountries[code] = struct{}{}
		}
	}
	if len(blockedCountries) == 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	headerNames := cfg.HeaderNames
	if len(headerNames) == 0 {
		headerNames = fallbackRegionCountryHeaders
	}
	blockedHosts := make(map[string]struct{}, len(cfg.Hosts))
	for _, host := range cfg.Hosts {
		normalized := normalizeRequestHost(host)
		if normalized != "" {
			blockedHosts[normalized] = struct{}{}
		}
	}
	supportEmail := strings.TrimSpace(cfg.SupportEmail)

	return func(c *gin.Context) {
		if !isRegionBlockPageRequest(c) || !regionBlockHostMatches(c.Request, blockedHosts) {
			c.Next()
			return
		}

		countryCode := requestCountryCode(c.Request, headerNames)
		if countryCode == "" {
			c.Next()
			return
		}
		if _, blocked := blockedCountries[countryCode]; !blocked {
			c.Next()
			return
		}

		c.Header("Cache-Control", "no-store")
		c.Header("X-Region-Blocked", countryCode)
		if c.Request.Method == http.MethodHead {
			c.Status(http.StatusForbidden)
		} else {
			c.Data(http.StatusForbidden, "text/html; charset=utf-8", []byte(renderUnsupportedRegionPage(countryCode, supportEmail)))
		}
		c.Abort()
	}
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
	if isRegionBlockAPIRoute(c.Request.URL.Path) {
		return false
	}
	accept := strings.ToLower(c.GetHeader("Accept"))
	return accept == "" || strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*")
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
