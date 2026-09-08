package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func NewGroupHandlerWithConfig(adminService service.AdminService, dashboardService *service.DashboardService, groupCapacityService *service.GroupCapacityService, _ *config.Config) *GroupHandler {
	return NewGroupHandler(adminService, dashboardService, groupCapacityService)
}

func (h *GroupHandler) GetGroupModelAllowlistCandidates(c *gin.Context) {
	response.Success(c, []string{})
}
