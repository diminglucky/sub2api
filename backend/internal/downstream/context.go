package downstream

import (
	"context"

	"github.com/gin-gonic/gin"
)

// ContextKeySubsite is the gin.Context key holding the resolved *Subsite. It is
// exported so tests and non-gin call sites can agree on the key, but callers
// should read the value through FromGin so the lookup stays type safe.
const ContextKeySubsite = "downstream.subsite"

// subsiteContextKey scopes the request-context value so no other package can
// spoof it by reusing an untyped string key.
type subsiteContextKey struct{}

// WithSubsite returns a child context carrying the resolved subsite so service
// and repository layers can scope their queries without importing gin.
func WithSubsite(ctx context.Context, subsite *Subsite) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, subsiteContextKey{}, subsite)
}

// FromContext reads the resolved subsite that WithSubsite stored.
func FromContext(ctx context.Context) (*Subsite, bool) {
	if ctx == nil {
		return nil, false
	}
	subsite, ok := ctx.Value(subsiteContextKey{}).(*Subsite)
	if !ok || subsite == nil {
		return nil, false
	}
	return subsite, true
}

// SetInGin stores the resolved subsite in the gin context and mirrors it into
// the request context. Handlers keep using FromGin; the gateway and services
// read FromContext to avoid a gin dependency.
func SetInGin(c *gin.Context, subsite *Subsite) {
	if c == nil || subsite == nil {
		return
	}
	c.Set(ContextKeySubsite, subsite)
	if c.Request != nil {
		c.Request = c.Request.WithContext(WithSubsite(c.Request.Context(), subsite))
	}
}

// FromGin reads the subsite resolved earlier in the request. It checks the gin
// context first, then the mirrored request context so it also works inside
// goroutines that only carry the request context.
func FromGin(c *gin.Context) (*Subsite, bool) {
	if c == nil {
		return nil, false
	}
	if value, ok := c.Get(ContextKeySubsite); ok {
		if subsite, ok := value.(*Subsite); ok && subsite != nil {
			return subsite, true
		}
	}
	if c.Request != nil {
		return FromContext(c.Request.Context())
	}
	return nil, false
}
