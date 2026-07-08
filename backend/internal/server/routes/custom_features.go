package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerCustomAdminRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	registerCustomLotteryRoutes(admin, h)
}

func registerCustomLotteryRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	lotteries := admin.Group("/lotteries")
	{
		lotteries.GET("", h.Admin.Lottery.List)
		lotteries.GET("/:id", h.Admin.Lottery.GetByID)
		lotteries.POST("", h.Admin.Lottery.Create)
		lotteries.PUT("/:id", h.Admin.Lottery.Update)
		lotteries.DELETE("/:id", h.Admin.Lottery.Delete)
		lotteries.POST("/:id/draw", h.Admin.Lottery.Draw)
	}
}

func registerCustomAdminSettingsRoutes(adminSettings *gin.RouterGroup, h *handler.Handlers) {
	adminSettings.GET("/upstream-monitor", h.Admin.Setting.GetUpstreamMonitorConfig)
	adminSettings.PUT("/upstream-monitor", h.Admin.Setting.UpdateUpstreamMonitorConfig)
	adminSettings.POST("/upstream-monitor/preview", h.Admin.Setting.PreviewUpstreamMonitorConfig)
	adminSettings.POST("/upstream-monitor/refresh", h.Admin.Setting.RefreshUpstreamMonitorConfig)
}

func registerCustomUserRoutes(authenticated *gin.RouterGroup, h *handler.Handlers) {
	lotteries := authenticated.Group("/lotteries")
	{
		lotteries.GET("", h.Lottery.List)
		lotteries.POST("/:id/join", h.Lottery.Join)
	}
}

func registerCustomPublicRoutes(public *gin.RouterGroup, h *handler.Handlers) {
	public.GET("/models/available", h.AvailableChannel.ListPublicModels)
}
