package downstream

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the internal downstream API.
//
// The site endpoint is safe before login because it only returns brand/config
// data for the Host-resolved subsite. User and admin summaries require JWT auth;
// userContextMiddleware copies the authenticated user ID into downstream's own
// context key so this package does not need to import the server middleware.
func RegisterRoutes(r *gin.Engine, repo SummaryRepository, jwtAuth gin.HandlerFunc, userContextMiddleware gin.HandlerFunc) {
	if r == nil {
		return
	}
	h := NewHandler(repo)
	group := r.Group("/api/internal/downstream/v1")
	group.GET("/site", h.Site)

	auth := group.Group("")
	if jwtAuth != nil {
		auth.Use(jwtAuth)
	}
	if userContextMiddleware != nil {
		auth.Use(userContextMiddleware)
	}
	auth.GET("/me/summary", h.MeSummary)
	auth.GET("/admin/summary", h.AdminSummary)

	// Sub-site-scoped price management. Authorized by sub-site membership role
	// (owner/admin), independent of the main-site admin role.
	if scopedRepo, ok := repo.(ScopedPriceRepository); ok && scopedRepo != nil {
		scoped := NewScopedAdminHandler(scopedRepo)
		admin := auth.Group("/admin")
		admin.GET("/channels", scoped.ListChannels)
		admin.GET("/prices", scoped.ListPrices)
		admin.POST("/prices", scoped.CreatePrice)
		admin.PUT("/prices/:id", scoped.UpdatePrice)
		admin.DELETE("/prices/:id", scoped.DeletePrice)
	}
}
