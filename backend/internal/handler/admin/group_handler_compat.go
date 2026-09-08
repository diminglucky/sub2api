package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func NewGroupHandlerWithConfig(adminService service.AdminService, dashboardService *service.DashboardService, groupCapacityService *service.GroupCapacityService, _ *config.Config) *GroupHandler {
	return NewGroupHandler(adminService, dashboardService, groupCapacityService)
}
