package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newAPIOnlyHostTestRouter(hosts []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(APIOnlyHost(hosts))
	r.GET("/login", func(c *gin.Context) {
		c.String(http.StatusOK, "login page")
	})
	r.POST("/api/v1/auth/register", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.POST("/api/event_logging/batch", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/v1/models", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"object": "list"})
	})
	r.GET("/v1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.POST("/chat/completions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.POST("/embeddings", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/images", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}

func TestAPIOnlyHostBlocksFrontendPages(t *testing.T) {
	r := newAPIOnlyHostTestRouter([]string{"api.dihappy.cfd"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.Host = "api.dihappy.cfd"
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Equal(t, "true", w.Header().Get("X-API-Only-Host"))
	require.Contains(t, w.Body.String(), "API_ONLY_HOST")
}

func TestAPIOnlyHostBlocksInternalAPIs(t *testing.T) {
	r := newAPIOnlyHostTestRouter([]string{"api.dihappy.cfd"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	req.Host = "api.dihappy.cfd"
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "API_ONLY_HOST")
}

func TestAPIOnlyHostBlocksNonGatewayAPIs(t *testing.T) {
	r := newAPIOnlyHostTestRouter([]string{"api.dihappy.cfd"})

	for _, path := range []string{"/api/event_logging/batch", "/health", "/v1", "/images"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, nil)
		if path == "/health" || path == "/v1" || path == "/images" {
			req = httptest.NewRequest(http.MethodGet, path, nil)
		}
		req.Host = "api.dihappy.cfd"
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code, path)
		require.Contains(t, w.Body.String(), "API_ONLY_HOST", path)
	}
}

func TestAPIOnlyHostAllowsTransportRoutes(t *testing.T) {
	r := newAPIOnlyHostTestRouter([]string{"api.dihappy.cfd"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Host = "api.dihappy.cfd"
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"object":"list"}`, w.Body.String())
}

func TestAPIOnlyHostAllowsRootTransportAliases(t *testing.T) {
	r := newAPIOnlyHostTestRouter([]string{"api.dihappy.cfd"})

	for _, path := range []string{"/chat/completions", "/embeddings"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Host = "api.dihappy.cfd"
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, path)
		require.JSONEq(t, `{"ok":true}`, w.Body.String(), path)
	}
}

func TestAPIOnlyHostAllowsOtherHosts(t *testing.T) {
	r := newAPIOnlyHostTestRouter([]string{"api.dihappy.cfd"})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.Host = "superai.dihappy.cfd"
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "login page", w.Body.String())
}

func TestAPIOnlyHostDisabledWhenNoHostsConfigured(t *testing.T) {
	r := newAPIOnlyHostTestRouter(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.Host = "api.dihappy.cfd"
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "login page", w.Body.String())
}
