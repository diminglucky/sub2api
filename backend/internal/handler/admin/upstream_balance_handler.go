package admin

import (
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type UpstreamBalanceHandler struct {
	service *service.UpstreamBalanceService
}

func NewUpstreamBalanceHandler(balanceService *service.UpstreamBalanceService) *UpstreamBalanceHandler {
	return &UpstreamBalanceHandler{service: balanceService}
}

func (h *UpstreamBalanceHandler) Overview(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "upstream balance service is unavailable")
		return
	}
	overview, err := h.service.Overview(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, overview)
}

func (h *UpstreamBalanceHandler) Configure(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("account_id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "invalid account id")
		return
	}
	var cfg service.UpstreamBalanceConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "invalid balance configuration")
		return
	}
	if err := h.service.Configure(c.Request.Context(), accountID, cfg); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"account_id": accountID})
}

func (h *UpstreamBalanceHandler) Refresh(c *gin.Context) {
	var req struct {
		AccountID *int64 `json:"account_id"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid refresh request")
			return
		}
	}
	result, err := h.service.Refresh(c.Request.Context(), req.AccountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
