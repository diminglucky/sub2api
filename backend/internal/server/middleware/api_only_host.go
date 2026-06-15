package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIOnlyHost prevents configured transport-only hostnames from serving the
// web UI or administrative/auth APIs. It still allows AI-compatible gateway
// endpoints such as /v1/*, /v1beta/*, /responses and /chat/completions.
func APIOnlyHost(hosts []string) gin.HandlerFunc {
	apiOnlyHosts := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		normalized := normalizeRequestHost(host)
		if normalized != "" {
			apiOnlyHosts[normalized] = struct{}{}
		}
	}
	if len(apiOnlyHosts) == 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		if !apiOnlyHostMatches(c.Request, apiOnlyHosts) || isAPITransportRoute(c.Request.URL.Path) {
			c.Next()
			return
		}

		c.Header("Cache-Control", "no-store")
		c.Header("X-API-Only-Host", "true")
		if c.Request.Method == http.MethodHead {
			c.Status(http.StatusNotFound)
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "API_ONLY_HOST",
					"message": "This host only accepts API transport requests.",
				},
			})
		}
		c.Abort()
	}
}

func apiOnlyHostMatches(req *http.Request, apiOnlyHosts map[string]struct{}) bool {
	if len(apiOnlyHosts) == 0 {
		return false
	}
	host := normalizeRequestHost(requestHost(req))
	if host == "" {
		return false
	}
	if _, ok := apiOnlyHosts[host]; ok {
		return true
	}
	if _, ok := apiOnlyHosts["*"]; ok {
		return true
	}
	return false
}

func isAPITransportRoute(path string) bool {
	path = strings.TrimSpace(path)
	return hasAPIPathDirPrefix(path, "/v1") ||
		hasAPIPathDirPrefix(path, "/v1beta") ||
		hasAPIPathDirPrefix(path, "/antigravity") ||
		hasAPIPathDirPrefix(path, "/backend-api/codex") ||
		hasAPIPathPrefix(path, "/responses") ||
		hasAPIPathPrefix(path, "/chat/completions") ||
		hasAPIPathPrefix(path, "/embeddings") ||
		hasAPIPathDirPrefix(path, "/images")
}

func hasAPIPathPrefix(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func hasAPIPathDirPrefix(path, prefix string) bool {
	return strings.HasPrefix(path, prefix+"/")
}
