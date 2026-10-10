//go:build embed

package web

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	htmlpkg "html"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

const (
	// NonceHTMLPlaceholder is the placeholder for nonce in HTML script tags
	NonceHTMLPlaceholder = "__CSP_NONCE_VALUE__"
	// seoBlockStart/seoBlockEnd 包住 index.html 里的 SEO 标签，按请求路径整体替换。
	seoBlockStart = "<!--seo:start-->"
	seoBlockEnd   = "<!--seo:end-->"
)

const defaultSeoKeywords = "AI API 中转站,API 中转站,GPT API,OpenAI API,Claude API,DeepSeek API,AI API 聚合平台,OpenAI 兼容接口"

type routeSeoDoc struct {
	Title       string
	Description string
	Keywords    string
	Robots      string
}

//go:embed all:dist
var frontendFS embed.FS

//go:embed all:dist-downstream
var downstreamFrontendFS embed.FS

// PublicSettingsProvider is an interface to fetch public settings
type PublicSettingsProvider interface {
	GetPublicSettingsForInjection(ctx context.Context) (any, error)
}

// FrontendServer serves the embedded frontend with settings injection
type FrontendServer struct {
	distFS               fs.FS
	fileServer           http.Handler
	downstreamDistFS     fs.FS
	downstreamFileServer http.Handler
	baseHTML             []byte
	cache                *HTMLCache
	settings             PublicSettingsProvider
	overrideDir          string // local file override directory
}

// NewFrontendServer creates a new frontend server with settings injection
func NewFrontendServer(settingsProvider PublicSettingsProvider) (*FrontendServer, error) {
	distFS, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		return nil, err
	}

	// Read base HTML once
	file, err := distFS.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	baseHTML, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	cache := NewHTMLCache()
	cache.SetBaseHTML(baseHTML)

	downstreamDistFS, downstreamFileServer := loadDownstreamFrontendFS()

	return &FrontendServer{
		distFS:               distFS,
		fileServer:           http.FileServer(http.FS(distFS)),
		downstreamDistFS:     downstreamDistFS,
		downstreamFileServer: downstreamFileServer,
		baseHTML:             baseHTML,
		cache:                cache,
		settings:             settingsProvider,
		overrideDir:          filepath.Join("data", "public"),
	}, nil
}

func loadDownstreamFrontendFS() (fs.FS, http.Handler) {
	distFS, err := fs.Sub(downstreamFrontendFS, "dist-downstream")
	if err != nil {
		return nil, nil
	}
	if _, err := fs.Stat(distFS, "index.html"); err != nil {
		return nil, nil
	}
	return distFS, http.FileServer(http.FS(distFS))
}

// InvalidateCache invalidates the HTML cache (call when settings change)
func (s *FrontendServer) InvalidateCache() {
	if s != nil && s.cache != nil {
		s.cache.Invalidate()
	}
}

// Middleware returns the Gin middleware handler
func (s *FrontendServer) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API routes
		if shouldBypassEmbeddedFrontend(path) {
			c.Next()
			return
		}

		if _, ok := downstream.FromGin(c); ok {
			s.serveDownstreamEntry(c)
			return
		}

		s.serveMainFrontend(c)
	}
}

// serveMainFrontend serves the main application shell and its static assets.
func (s *FrontendServer) serveMainFrontend(c *gin.Context) {
	cleanPath := strings.TrimPrefix(c.Request.URL.Path, "/")
	if cleanPath == "" {
		cleanPath = "index.html"
	}

	// For index.html or SPA routes, serve with injected settings
	if cleanPath == "index.html" || !s.fileExists(cleanPath) {
		s.serveIndexHTML(c)
		return
	}

	// Try local override first
	if s.tryServeOverride(c, cleanPath) {
		return
	}

	// Serve static files normally (hashed assets get long-lived cache headers)
	applyStaticAssetCacheHeaders(c.Writer.Header(), cleanPath)
	s.fileServer.ServeHTTP(c.Writer, c.Request)
	c.Abort()
}

// serveDownstreamEntry serves the branded drawing landing for a resolved
// subsite. Only the entry page and its own static assets come from the
// downstream build; every other route falls back to the main frontend so the
// signed-in user pages stay identical to the main site.
func (s *FrontendServer) serveDownstreamEntry(c *gin.Context) {
	cleanPath := strings.TrimPrefix(c.Request.URL.Path, "/")
	if cleanPath == "" {
		cleanPath = "index.html"
	}

	if cleanPath != "index.html" && !s.downstreamFileExists(cleanPath) {
		s.serveMainFrontendForSubsite(c)
		return
	}

	if s.downstreamDistFS == nil || s.downstreamFileServer == nil {
		s.serveMainFrontendForSubsite(c)
		return
	}

	if cleanPath == "index.html" {
		serveIndexHTML(c, s.downstreamDistFS)
		return
	}

	applyStaticAssetCacheHeaders(c.Writer.Header(), cleanPath)
	s.downstreamFileServer.ServeHTTP(c.Writer, c.Request)
	c.Abort()
}

func (s *FrontendServer) downstreamFileExists(path string) bool {
	return fileExists(s.downstreamDistFS, path)
}

// serveMainFrontendForSubsite serves the main application shell on a subsite
// host, injecting the subsite brand so the reused user pages render as Draw.
// Static assets are shared with the main build; only the HTML shell differs.
func (s *FrontendServer) serveMainFrontendForSubsite(c *gin.Context) {
	cleanPath := strings.TrimPrefix(c.Request.URL.Path, "/")
	if cleanPath == "" {
		cleanPath = "index.html"
	}

	if cleanPath == "index.html" || !s.fileExists(cleanPath) {
		name, logo := "", ""
		if subsite, ok := downstream.FromGin(c); ok && subsite != nil {
			name = subsite.Name
			logo = subsite.LogoURL
		}
		s.serveIndexHTMLForSubsite(c, name, logo)
		return
	}

	if s.tryServeOverride(c, cleanPath) {
		return
	}

	applyStaticAssetCacheHeaders(c.Writer.Header(), cleanPath)
	s.fileServer.ServeHTTP(c.Writer, c.Request)
	c.Abort()
}

func (s *FrontendServer) fileExists(path string) bool {
	return fileExists(s.distFS, path)
}

func fileExists(fsys fs.FS, path string) bool {
	if fsys == nil {
		return false
	}
	file, err := fsys.Open(path)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

// tryServeOverride checks if a local override file exists and serves it.
// Files in overrideDir take precedence over embedded files.
func (s *FrontendServer) tryServeOverride(c *gin.Context, cleanPath string) bool {
	if s.overrideDir == "" {
		return false
	}
	filePath := filepath.Join(s.overrideDir, filepath.Clean("/"+cleanPath))
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return false
	}
	c.File(filePath)
	c.Abort()
	return true
}

func (s *FrontendServer) serveIndexHTML(c *gin.Context) {
	// Get nonce from context (generated by SecurityHeaders middleware)
	nonce := middleware.GetNonceFromContext(c)

	// Check cache first
	cached := s.cache.Get()
	if cached != nil {
		// Check If-None-Match for 304 response
		if match := c.GetHeader("If-None-Match"); match == cached.ETag {
			c.Status(http.StatusNotModified)
			c.Abort()
			return
		}

		// Replace nonce placeholder with actual nonce before serving
		content := replaceNoncePlaceholder(cached.Content, nonce)
		content = applyRouteSeo(content, requestOrigin(c), c.Request.URL.Path, cached.SiteName)

		c.Header("ETag", cached.ETag)
		c.Header("Cache-Control", "no-cache") // Must revalidate
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
		c.Abort()
		return
	}

	// Cache miss - fetch settings and render
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	settings, err := s.settings.GetPublicSettingsForInjection(ctx)
	if err != nil {
		// Fallback: serve without injection
		c.Data(http.StatusOK, "text/html; charset=utf-8", s.baseHTML)
		c.Abort()
		return
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		// Fallback: serve without injection
		c.Data(http.StatusOK, "text/html; charset=utf-8", s.baseHTML)
		c.Abort()
		return
	}

	rendered := s.injectSettings(settingsJSON)
	s.cache.Set(rendered, settingsJSON)

	// Replace nonce placeholder with actual nonce before serving
	content := replaceNoncePlaceholder(rendered, nonce)
	content = applyRouteSeo(content, requestOrigin(c), c.Request.URL.Path, siteNameFromSettingsJSON(settingsJSON))

	cached = s.cache.Get()
	if cached != nil {
		c.Header("ETag", cached.ETag)
	}
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	c.Abort()
}

// serveIndexHTMLForSubsite renders the main app shell with the subsite brand
// injected. The shared HTML cache is bypassed because the injected brand
// differs per host, and settings are cheap to read through the settings cache.
func (s *FrontendServer) serveIndexHTMLForSubsite(c *gin.Context, name, logo string) {
	if s.settings == nil {
		s.serveIndexHTML(c)
		return
	}

	nonce := middleware.GetNonceFromContext(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	settings, err := s.settings.GetPublicSettingsForInjection(ctx)
	if err != nil {
		c.Data(http.StatusOK, "text/html; charset=utf-8", s.baseHTML)
		c.Abort()
		return
	}
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		c.Data(http.StatusOK, "text/html; charset=utf-8", s.baseHTML)
		c.Abort()
		return
	}
	settingsJSON = applySubsiteBranding(settingsJSON, name, logo)
	rendered := s.injectSettings(settingsJSON)
	content := replaceNoncePlaceholder(rendered, nonce)
	content = applyRouteSeo(content, requestOrigin(c), c.Request.URL.Path, siteNameFromSettingsJSON(settingsJSON))
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	c.Abort()
}

// applySubsiteBranding overrides the public site name/logo with the subsite's
// own brand so the reused main app renders as the downstream brand.
func applySubsiteBranding(settingsJSON []byte, name, logo string) []byte {
	name = strings.TrimSpace(name)
	logo = strings.TrimSpace(logo)
	if name == "" && logo == "" {
		return settingsJSON
	}
	var cfg map[string]any
	if err := json.Unmarshal(settingsJSON, &cfg); err != nil {
		return settingsJSON
	}
	if name != "" {
		cfg["site_name"] = name
	}
	if logo != "" {
		cfg["site_logo"] = logo
	}
	out, err := json.Marshal(cfg)
	if err != nil {
		return settingsJSON
	}
	return out
}

func (s *FrontendServer) injectSettings(settingsJSON []byte) []byte {
	// Create the script tag to inject with nonce placeholder
	// The placeholder will be replaced with actual nonce at request time
	script := []byte(`<script nonce="` + NonceHTMLPlaceholder + `">window.__APP_CONFIG__=` + string(settingsJSON) + `;</script>`)

	// Inject before </head>
	headClose := []byte("</head>")
	result := bytes.Replace(s.baseHTML, headClose, append(script, headClose...), 1)

	// Apply custom branding before the browser paints the static defaults.
	result = injectSiteTitle(result, settingsJSON)
	result = injectSiteFavicon(result, settingsJSON)

	return result
}

// injectSiteFavicon replaces the static favicon with a configured, browser-safe image URL.
func injectSiteFavicon(html, settingsJSON []byte) []byte {
	var cfg struct {
		SiteLogo string `json:"site_logo"`
	}
	if err := json.Unmarshal(settingsJSON, &cfg); err != nil {
		return html
	}

	logoURL := safeImageURL(cfg.SiteLogo)
	if logoURL == "" {
		return html
	}

	linkStart := bytes.Index(html, []byte(`<link rel="icon"`))
	if linkStart == -1 {
		return html
	}
	linkEndOffset := bytes.IndexByte(html[linkStart:], '>')
	if linkEndOffset == -1 {
		return html
	}
	linkEnd := linkStart + linkEndOffset + 1
	replacement := []byte(`<link rel="icon" href="` + htmlpkg.EscapeString(logoURL) + `" />`)

	var buf bytes.Buffer
	buf.Write(html[:linkStart])
	buf.Write(replacement)
	buf.Write(html[linkEnd:])
	return buf.Bytes()
}

func safeImageURL(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "/") && !strings.HasPrefix(trimmed, "//") {
		return trimmed
	}
	if strings.HasPrefix(strings.ToLower(trimmed), "data:image/") {
		return trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return ""
	}
	return trimmed
}

// injectSiteTitle replaces the static <title> in HTML with the configured site name.
// This ensures the browser tab shows the correct title before JS executes.
func injectSiteTitle(html, settingsJSON []byte) []byte {
	var cfg struct {
		SiteName string `json:"site_name"`
	}
	if err := json.Unmarshal(settingsJSON, &cfg); err != nil || cfg.SiteName == "" {
		return html
	}

	// Find and replace the existing <title>...</title>
	titleStart := bytes.Index(html, []byte("<title>"))
	titleEnd := bytes.Index(html, []byte("</title>"))
	if titleStart == -1 || titleEnd == -1 || titleEnd <= titleStart {
		return html
	}

	newTitle := []byte("<title>" + htmlpkg.EscapeString(cfg.SiteName) + " - AI API Gateway</title>")
	var buf bytes.Buffer
	buf.Write(html[:titleStart])
	buf.Write(newTitle)
	buf.Write(html[titleEnd+len("</title>"):])
	return buf.Bytes()
}

// replaceNoncePlaceholder replaces the nonce placeholder with actual nonce value
func replaceNoncePlaceholder(html []byte, nonce string) []byte {
	return bytes.ReplaceAll(html, []byte(NonceHTMLPlaceholder), []byte(nonce))
}

func normalizeSeoPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	if len(trimmed) > 1 {
		trimmed = strings.TrimRight(trimmed, "/")
	}
	if trimmed == "" {
		return "/"
	}
	return trimmed
}

// resolveRouteSeoDoc 决定每个路径服务端直接输出的 title/description/robots。
// 只有公开页面允许收录；后台与登录后页面一律 noindex，避免被索引到空壳页。
func resolveRouteSeoDoc(path, siteName string) routeSeoDoc {
	name := strings.TrimSpace(siteName)
	if name == "" {
		name = "SuperAI"
	}
	normalized := normalizeSeoPath(path)
	switch {
	case normalized == "/" || normalized == "/home":
		return routeSeoDoc{
			Title:       name + " - GPT、OpenAI、Claude、DeepSeek API 中转站",
			Description: "兼容 OpenAI 接口的 AI API 聚合平台，支持 GPT、Claude、Gemini、DeepSeek 等模型，提供模型价格对比与灵活计费。",
			Keywords:    defaultSeoKeywords,
			Robots:      "index,follow",
		}
	case normalized == "/model-plaza":
		return routeSeoDoc{
			Title:       name + " - AI 模型广场与价格对比",
			Description: "对比 GPT、Claude、Gemini、DeepSeek 等模型的实时价格、上下文窗口与计费方式，OpenAI 兼容接口可直接接入。",
			Keywords:    "AI API 价格,GPT API 价格,Claude API 价格,DeepSeek API 价格,AI 模型价格对比",
			Robots:      "index,follow",
		}
	case normalized == "/key-usage":
		return routeSeoDoc{
			Title:       name + " - API 密钥用量查询",
			Description: "按密钥、模型与时间范围查询 API 调用量与额度消耗明细。",
			Keywords:    "API 密钥用量,API 额度查询,Token 用量",
			Robots:      "index,follow",
		}
	case strings.HasPrefix(normalized, "/legal/"):
		return routeSeoDoc{
			Title:       name + " - 条款与政策",
			Description: name + " 的服务条款、隐私政策与使用规范。",
			Keywords:    defaultSeoKeywords,
			Robots:      "index,follow",
		}
	}
	return routeSeoDoc{
		Title:       name + " - AI API Gateway",
		Description: "兼容 OpenAI 接口的 AI API 聚合平台。",
		Keywords:    defaultSeoKeywords,
		Robots:      "noindex,nofollow",
	}
}

// requestOrigin 依据可信反代头推导外部访问源，canonical/og:url 必须用真实域名。
func requestOrigin(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		return ""
	}
	scheme := ""
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		scheme = strings.ToLower(strings.TrimSpace(strings.Split(forwarded, ",")[0]))
	}
	if scheme != "http" && scheme != "https" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + strings.TrimRight(host, "/")
}

// applyRouteSeo 用当前请求对应的 SEO 标签整体替换 index.html 里的标记区块。
func applyRouteSeo(html []byte, origin, path, siteName string) []byte {
	start := bytes.Index(html, []byte(seoBlockStart))
	end := bytes.Index(html, []byte(seoBlockEnd))
	if start == -1 || end == -1 || end <= start {
		return html
	}
	block := buildSeoBlock(origin, path, siteName)
	var buf bytes.Buffer
	buf.Grow(len(html) + len(block))
	buf.Write(html[:start])
	buf.WriteString(block)
	buf.Write(html[end+len(seoBlockEnd):])
	return buf.Bytes()
}

func buildSeoBlock(origin, path, siteName string) string {
	doc := resolveRouteSeoDoc(path, siteName)
	canonicalPath := normalizeSeoPath(path)
	canonical := canonicalPath
	imageURL := "/logo.svg"
	if origin != "" {
		canonical = origin + canonicalPath
		imageURL = origin + "/logo.svg"
	}

	var b strings.Builder
	esc := htmlpkg.EscapeString
	writeTag := func(line string) {
		b.WriteString("    " + line + "\n")
	}
	writeMeta := func(attr, key, content string) {
		writeTag(`<meta ` + attr + `="` + key + `" content="` + esc(content) + `" />`)
	}

	writeTag("<title>" + esc(doc.Title) + "</title>")
	writeMeta("name", "description", doc.Description)
	writeMeta("name", "keywords", doc.Keywords)
	writeMeta("name", "robots", doc.Robots)
	writeMeta("name", "msvalidate.01", "B55F2646907C6E19E99891EF692EBBF8")
	writeMeta("name", "application-name", strings.TrimSpace(siteName))
	writeMeta("name", "theme-color", "#14b8a6")
	writeTag(`<link rel="canonical" href="` + esc(canonical) + `" />`)
	writeMeta("property", "og:type", "website")
	writeMeta("property", "og:site_name", strings.TrimSpace(siteName))
	writeMeta("property", "og:title", doc.Title)
	writeMeta("property", "og:description", doc.Description)
	writeMeta("property", "og:url", canonical)
	writeMeta("property", "og:image", imageURL)
	writeMeta("name", "twitter:card", "summary")
	writeMeta("name", "twitter:title", doc.Title)
	writeMeta("name", "twitter:description", doc.Description)
	b.WriteString(buildStructuredData(origin, canonical, siteName))
	return b.String()
}

func buildStructuredData(origin, canonical, siteName string) string {
	name := strings.TrimSpace(siteName)
	if name == "" {
		name = "SuperAI"
	}
	provider := map[string]any{"@type": "Organization", "name": name}
	if origin != "" {
		provider["url"] = origin + "/"
		provider["logo"] = origin + "/logo.svg"
	}
	payload := map[string]any{
		"@context": "https://schema.org",
		"@graph": []map[string]any{
			provider,
			{"@type": "WebSite", "name": name, "url": canonical},
			{
				"@type":       "Service",
				"name":        name + " AI API 中转与聚合服务",
				"serviceType": "OpenAI-compatible API gateway",
				"provider":    provider,
				"description": "兼容 OpenAI 接口，可接入 GPT、Claude、Gemini、DeepSeek 等模型。",
			},
		},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return `    <script type="application/ld+json">` + string(encoded) + "</script>\n"
}

// ServeEmbeddedFrontend returns a middleware for serving embedded frontend
// This is the legacy function for backward compatibility when no settings provider is available
func ServeEmbeddedFrontend() gin.HandlerFunc {
	distFS, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		panic("failed to get dist subdirectory: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(distFS))
	overrideDir := filepath.Join("data", "public")

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if shouldBypassEmbeddedFrontend(path) {
			c.Next()
			return
		}

		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		if file, err := distFS.Open(cleanPath); err == nil {
			_ = file.Close()
			// Try local override first
			if tryServeOverrideFile(c, overrideDir, cleanPath) {
				return
			}
			applyStaticAssetCacheHeaders(c.Writer.Header(), cleanPath)
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		serveIndexHTML(c, distFS)
	}
}

// tryServeOverrideFile is a standalone version of tryServeOverride for legacy usage.
func tryServeOverrideFile(c *gin.Context, overrideDir, cleanPath string) bool {
	if overrideDir == "" {
		return false
	}
	filePath := filepath.Join(overrideDir, filepath.Clean("/"+cleanPath))
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return false
	}
	c.File(filePath)
	c.Abort()
	return true
}

func shouldBypassEmbeddedFrontend(path string) bool {
	trimmed := strings.TrimSpace(path)
	return strings.HasPrefix(trimmed, "/api/") ||
		strings.HasPrefix(trimmed, "/v1/") ||
		strings.HasPrefix(trimmed, "/v1beta/") ||
		strings.HasPrefix(trimmed, "/backend-api/") ||
		strings.HasPrefix(trimmed, "/antigravity/") ||
		strings.HasPrefix(trimmed, "/setup/") ||
		trimmed == "/health" ||
		trimmed == "/models" ||
		trimmed == "/responses" ||
		strings.HasPrefix(trimmed, "/responses/") ||
		strings.HasPrefix(trimmed, "/chat/completions") ||
		strings.HasPrefix(trimmed, "/embeddings") ||
		trimmed == "/alpha/search" ||
		strings.HasPrefix(trimmed, "/images/") ||
		strings.HasPrefix(trimmed, "/videos/")
}

func serveIndexHTML(c *gin.Context, fsys fs.FS) {
	file, err := fsys.Open("index.html")
	if err != nil {
		c.String(http.StatusNotFound, "Frontend not found")
		c.Abort()
		return
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read index.html")
		c.Abort()
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	c.Abort()
}

func HasEmbeddedFrontend() bool {
	_, err := frontendFS.ReadFile("dist/index.html")
	return err == nil
}
