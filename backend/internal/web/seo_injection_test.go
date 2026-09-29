//go:build embed

package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveRouteSeoDocPublicRoutesAreIndexable(t *testing.T) {
	for _, path := range []string{"/", "/home", "/home/", "/model-plaza", "/key-usage", "/legal/privacy"} {
		doc := resolveRouteSeoDoc(path, "SuperAI")
		assert.Equal(t, "index,follow", doc.Robots, "path %s must be indexable", path)
		assert.Contains(t, doc.Title, "SuperAI")
		assert.NotEmpty(t, doc.Description)
	}
}

func TestResolveRouteSeoDocPrivateRoutesAreNoIndex(t *testing.T) {
	for _, path := range []string{"/dashboard", "/admin/settings", "/keys", "/image-studio", "/profile"} {
		doc := resolveRouteSeoDoc(path, "SuperAI")
		assert.Equal(t, "noindex,nofollow", doc.Robots, "path %s must not be indexed", path)
	}
}

// 每个路径必须拿到自己的 canonical：所有子页面都指回首页会被当成重复内容。
func TestApplyRouteSeoUsesPerPathCanonical(t *testing.T) {
	base := []byte(`<html><head>` + seoBlockStart +
		`<title>old</title><link rel="canonical" href="https://example.com/" />` +
		seoBlockEnd + `</head><body></body></html>`)

	result := string(applyRouteSeo(base, "https://superai.dihappy.cfd", "/model-plaza", "SuperAI"))

	require.Contains(t, result, `<link rel="canonical" href="https://superai.dihappy.cfd/model-plaza" />`)
	require.Contains(t, result, `<meta property="og:url" content="https://superai.dihappy.cfd/model-plaza" />`)
	require.Contains(t, result, `<meta name="robots" content="index,follow" />`)
	require.Contains(t, result, "AI 模型广场与价格对比")
	require.NotContains(t, result, ">old<")
	// 标记只用于服务端定位，不能出现在响应里
	require.NotContains(t, result, seoBlockStart)
	require.NotContains(t, result, seoBlockEnd)
}

func TestApplyRouteSeoLeavesHTMLWithoutMarkersUntouched(t *testing.T) {
	base := []byte(`<html><head><title>keep</title></head></html>`)
	assert.Equal(t, string(base), string(applyRouteSeo(base, "https://example.com", "/home", "SuperAI")))
}

func TestApplyRouteSeoEscapesInjectedValues(t *testing.T) {
	base := []byte(`<head>` + seoBlockStart + `x` + seoBlockEnd + `</head>`)
	result := string(applyRouteSeo(base, "https://example.com", "/home", `Bad</title><script>alert(1)</script>`))
	assert.NotContains(t, result, "<script>alert(1)</script>")
	assert.Contains(t, result, "&lt;script&gt;")
}

func TestRequestOriginPrefersForwardedProto(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/home", nil)
	c.Request.Host = "superai.dihappy.cfd"
	c.Request.Header.Set("X-Forwarded-Proto", "https")

	assert.Equal(t, "https://superai.dihappy.cfd", requestOrigin(c))

	// 没有反代头时退回请求本身的 scheme
	c.Request.Header.Del("X-Forwarded-Proto")
	assert.Equal(t, "http://superai.dihappy.cfd", requestOrigin(c))
}

func TestBuildSeoBlockOmitsAbsoluteURLsWithoutOrigin(t *testing.T) {
	block := buildSeoBlock("", "/home", "SuperAI")
	assert.Contains(t, block, `<link rel="canonical" href="/home" />`)
	assert.Contains(t, block, `<meta property="og:url" content="/home" />`)
	assert.False(t, strings.Contains(block, "superai.dihappy.cfd"), "no host should be invented without one")
}
