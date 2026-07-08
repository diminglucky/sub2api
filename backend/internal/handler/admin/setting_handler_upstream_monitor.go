package admin

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// SetUpstreamMonitorService attaches the focused upstream monitor service without
// changing the constructor signature used by existing unit tests.
func (h *SettingHandler) SetUpstreamMonitorService(upstreamMonitorService *service.UpstreamMonitorService) {
	h.upstreamMonitorService = upstreamMonitorService
}

func (h *SettingHandler) upstreamMonitor() *service.UpstreamMonitorService {
	if h == nil {
		return nil
	}
	return h.upstreamMonitorService
}

func (h *SettingHandler) requireUpstreamMonitor(c *gin.Context) *service.UpstreamMonitorService {
	svc := h.upstreamMonitor()
	if svc == nil {
		response.InternalError(c, "upstream monitor service not configured")
		return nil
	}
	return svc
}

// GetUpstreamMonitorConfig 获取上游监测配置
// GET /api/v1/admin/settings/upstream-monitor
func (h *SettingHandler) GetUpstreamMonitorConfig(c *gin.Context) {
	svc := h.requireUpstreamMonitor(c)
	if svc == nil {
		return
	}
	cfg, err := svc.GetUpstreamMonitorConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, cfg)
}

// UpdateUpstreamMonitorConfig 更新上游监测配置
// PUT /api/v1/admin/settings/upstream-monitor
func (h *SettingHandler) UpdateUpstreamMonitorConfig(c *gin.Context) {
	var req service.UpstreamMonitorConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	svc := h.requireUpstreamMonitor(c)
	if svc == nil {
		return
	}
	if err := svc.SaveUpstreamMonitorConfig(c.Request.Context(), &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := svc.GetUpstreamMonitorConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, updated)
}

// PreviewUpstreamMonitorConfig 预览上游监测快照
// POST /api/v1/admin/settings/upstream-monitor/preview
func (h *SettingHandler) PreviewUpstreamMonitorConfig(c *gin.Context) {
	var req service.UpstreamMonitorConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	svc := h.requireUpstreamMonitor(c)
	if svc == nil {
		return
	}
	snapshot, err := svc.PreviewUpstreamMonitorConfig(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, snapshot)
}

// RefreshUpstreamMonitorConfig 立即刷新上游监测来源
// POST /api/v1/admin/settings/upstream-monitor/refresh
func (h *SettingHandler) RefreshUpstreamMonitorConfig(c *gin.Context) {
	var req struct {
		SourceID string `json:"source_id"`
	}
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request: "+err.Error())
			return
		}
	}

	var result *service.UpstreamMonitorRefreshResult
	var err error
	sourceID := strings.TrimSpace(req.SourceID)
	svc := h.requireUpstreamMonitor(c)
	if svc == nil {
		return
	}
	if sourceID != "" {
		result, err = svc.RefreshStoredUpstreamMonitorSource(c.Request.Context(), sourceID)
	} else {
		result, err = svc.RefreshStoredUpstreamMonitorConfig(c.Request.Context())
	}
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}
