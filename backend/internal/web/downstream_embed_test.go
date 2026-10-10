//go:build embed

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestFrontendMiddlewareServesDownstreamForResolvedSubsite(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mainFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("main-index")},
	}
	downstreamFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("draw-index")},
		"robots.txt": &fstest.MapFile{Data: []byte("User-agent: *\nAllow: /\n")},
	}
	cache := NewHTMLCache()
	cache.SetBaseHTML([]byte("main-index"))
	cache.Set([]byte("main-index"), []byte("{}"))
	server := &FrontendServer{
		distFS:               mainFS,
		fileServer:           http.FileServer(http.FS(mainFS)),
		downstreamDistFS:     downstreamFS,
		downstreamFileServer: http.FileServer(http.FS(downstreamFS)),
		baseHTML:             []byte("main-index"),
		cache:                cache,
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		downstream.SetInGin(c, &downstream.Subsite{
			ID:     7,
			Slug:   "draw",
			Domain: "draw.superai.sbs",
			Name:   "Draw",
			Status: downstream.SubsiteStatusActive,
		})
		c.Next()
	})
	router.Use(server.Middleware())
	router.GET("/v1/models", func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "missing api key"}})
	})

	t.Run("root serves downstream index", func(t *testing.T) {
		w := performDownstreamFrontendRequest(router, "/")
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "draw-index", w.Body.String())
	})

	t.Run("app routes reuse the main frontend", func(t *testing.T) {
		w := performDownstreamFrontendRequest(router, "/dashboard")
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "main-index")
	})

	t.Run("robots serves downstream static file", func(t *testing.T) {
		w := performDownstreamFrontendRequest(router, "/robots.txt")
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "User-agent: *")
	})

	t.Run("v1 stays on gateway path", func(t *testing.T) {
		w := performDownstreamFrontendRequest(router, "/v1/models")
		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.JSONEq(t, `{"error":{"message":"missing api key"}}`, w.Body.String())
	})
}

func performDownstreamFrontendRequest(router *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = "draw.superai.sbs"
	router.ServeHTTP(w, req)
	return w
}
