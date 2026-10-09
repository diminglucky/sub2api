package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/gin-gonic/gin"
)

// DownstreamSlugHeader is the trusted, edge-provided header naming the resolved
// subsite slug. The edge layer must strip any client-supplied value before
// forwarding, so the backend only ever sees the authoritative slug.
const DownstreamSlugHeader = "X-Downstream-Slug"

// DownstreamSubsiteResolver resolves a registered subsite from its slug.
// *downstream.Repository satisfies it through GetSubsiteBySlug.
type DownstreamSubsiteResolver interface {
	GetSubsiteBySlug(ctx context.Context, slug string) (*downstream.Subsite, error)
}

// DownstreamSubsite resolves the subsite context for requests that arrive on a
// "<slug>.<baseDomain>" host, or that carry the trusted DownstreamSlugHeader.
// The resolved *downstream.Subsite is attached to the gin context so handlers
// and the gateway can read it with downstream.FromGin.
//
// Resolution is context only: it never rewrites paths, bodies or headers, so
// OpenAI-compatible /v1 requests keep the current gateway behavior. Requests
// that are not scoped to a subsite (apex domain, unrelated hosts, empty
// configuration) pass through untouched.
//
// The V1 pilot subsite is draw.superai.sbs.
func DownstreamSubsite(resolver DownstreamSubsiteResolver, baseDomain string) gin.HandlerFunc {
	base := normalizeDownstreamBaseDomain(baseDomain)
	if resolver == nil || base == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return func(c *gin.Context) {
		slug, scoped := downstreamSlugForRequest(c.Request, base)
		if !scoped {
			c.Next()
			return
		}

		subsite, err := resolver.GetSubsiteBySlug(c.Request.Context(), slug)
		if err != nil {
			if errors.Is(err, downstream.ErrSubsiteNotFound) {
				abortDownstreamSubsiteNotFound(c)
				return
			}
			abortDownstreamSubsiteResolutionFailed(c)
			return
		}
		// A registered-but-disabled subsite is not served, matching the edge
		// rule that only active subsites are forwarded.
		if subsite == nil || !subsite.IsActive() {
			abortDownstreamSubsiteNotFound(c)
			return
		}

		downstream.SetInGin(c, subsite)
		c.Next()
	}
}

// downstreamSlugForRequest picks the subsite slug from the request. The trusted
// edge header wins; otherwise the first label of a "<label>.<baseDomain>" host
// is used. Hosts outside the base domain (apex, localhost, unrelated domains)
// are not subsite-scoped.
func downstreamSlugForRequest(req *http.Request, baseDomain string) (string, bool) {
	if req == nil {
		return "", false
	}
	if headerSlug := downstream.NormalizeSlug(req.Header.Get(DownstreamSlugHeader)); headerSlug != "" {
		return headerSlug, true
	}

	host := normalizeRequestHost(requestHost(req))
	if host == "" || host == baseDomain {
		return "", false
	}
	suffix := "." + baseDomain
	if !strings.HasSuffix(host, suffix) {
		return "", false
	}

	// "first host label": draw.superai.sbs -> draw, a.b.superai.sbs -> a.
	label := strings.TrimSuffix(host, suffix)
	if idx := strings.Index(label, "."); idx >= 0 {
		label = label[:idx]
	}
	label = downstream.NormalizeSlug(label)
	if label == "" {
		return "", false
	}
	return label, true
}

// normalizeDownstreamBaseDomain reduces a configured base domain to the bare
// host form so it can be compared with normalizeRequestHost output.
func normalizeDownstreamBaseDomain(domain string) string {
	return strings.TrimPrefix(normalizeRequestHost(domain), ".")
}

func abortDownstreamSubsiteNotFound(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if c.Request != nil && c.Request.Method == http.MethodHead {
		c.Status(http.StatusNotFound)
		c.Abort()
		return
	}
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
		"error": gin.H{
			"code":    "SUBSITE_NOT_FOUND",
			"message": "Unknown downstream subsite.",
		},
	})
}

func abortDownstreamSubsiteResolutionFailed(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if c.Request != nil && c.Request.Method == http.MethodHead {
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"code":    "SUBSITE_RESOLUTION_FAILED",
			"message": "Failed to resolve downstream subsite.",
		},
	})
}
