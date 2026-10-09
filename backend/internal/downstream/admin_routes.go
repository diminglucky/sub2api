package downstream

import "github.com/gin-gonic/gin"

// RegisterAdminRoutes mounts main-site admin provisioning under the existing
// /api/v1/admin group. Callers must attach admin auth to adminGroup.
func RegisterAdminRoutes(adminGroup *gin.RouterGroup, repo AdminManagementRepository) {
	if adminGroup == nil || repo == nil {
		return
	}
	h := NewAdminHandler(repo)
	subsites := adminGroup.Group("/downstream/subsites")
	{
		subsites.GET("", h.ListSubsites)
		subsites.POST("", h.CreateSubsite)
		subsites.GET("/:id", h.GetSubsite)
		subsites.PUT("/:id", h.UpdateSubsite)
		subsites.POST("/:id/disable", h.DisableSubsite)
		subsites.GET("/:id/prices", h.ListPrices)
		subsites.POST("/:id/prices", h.CreatePrice)
		subsites.PUT("/:id/prices/:price_id", h.UpdatePrice)
		subsites.DELETE("/:id/prices/:price_id", h.DeletePrice)
	}
}
