package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type LotteryHandler struct {
	lotteryService *service.LotteryService
}

func NewLotteryHandler(lotteryService *service.LotteryService) *LotteryHandler {
	return &LotteryHandler{lotteryService: lotteryService}
}

type lotteryPrizeRequest struct {
	Type        string   `json:"type" binding:"required,oneof=balance card"`
	Name        string   `json:"name" binding:"required"`
	Quantity    int      `json:"quantity" binding:"required,min=1"`
	Amount      *float64 `json:"amount"`
	CardContent *string  `json:"card_content"`
	SortOrder   int      `json:"sort_order"`
}

type createLotteryRequest struct {
	Title       string                `json:"title" binding:"required"`
	Description string                `json:"description"`
	Status      string                `json:"status" binding:"omitempty,oneof=draft active archived"`
	StartsAt    *int64                `json:"starts_at"`
	DrawAt      int64                 `json:"draw_at" binding:"required"`
	Prizes      []lotteryPrizeRequest `json:"prizes" binding:"required"`
}

type updateLotteryRequest struct {
	Title       *string                `json:"title"`
	Description *string                `json:"description"`
	Status      *string                `json:"status" binding:"omitempty,oneof=draft active drawn archived"`
	StartsAt    *int64                 `json:"starts_at"`
	DrawAt      *int64                 `json:"draw_at"`
	Prizes      *[]lotteryPrizeRequest `json:"prizes"`
}

func (h *LotteryHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
	status := strings.TrimSpace(c.Query("status"))
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 200 {
		search = search[:200]
	}
	items, paginationResult, err := h.lotteryService.ListAdmin(c.Request.Context(), params, status, search)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, adminLotteryEventsToResponse(items), paginationResult.Total, page, pageSize)
}

func (h *LotteryHandler) GetByID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || eventID <= 0 {
		response.BadRequest(c, "Invalid lottery ID")
		return
	}
	event, err := h.lotteryService.GetByID(c.Request.Context(), eventID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, adminLotteryEventToResponse(event))
}

func (h *LotteryHandler) Create(c *gin.Context) {
	var req createLotteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	input := &service.CreateLotteryInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		DrawAt:      time.Unix(req.DrawAt, 0),
		Prizes:      lotteryPrizeInputsFromRequest(req.Prizes),
		ActorID:     &subject.UserID,
	}
	if req.StartsAt != nil && *req.StartsAt > 0 {
		startsAt := time.Unix(*req.StartsAt, 0)
		input.StartsAt = &startsAt
	}
	event, err := h.lotteryService.Create(c.Request.Context(), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, adminLotteryEventToResponse(event))
}

func (h *LotteryHandler) Update(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || eventID <= 0 {
		response.BadRequest(c, "Invalid lottery ID")
		return
	}
	var req updateLotteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	input := &service.UpdateLotteryInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		ActorID:     &subject.UserID,
	}
	if req.StartsAt != nil {
		if *req.StartsAt == 0 {
			var cleared *time.Time
			input.StartsAt = &cleared
		} else {
			startsAt := time.Unix(*req.StartsAt, 0)
			ptr := &startsAt
			input.StartsAt = &ptr
		}
	}
	if req.DrawAt != nil && *req.DrawAt > 0 {
		drawAt := time.Unix(*req.DrawAt, 0)
		input.DrawAt = &drawAt
	}
	if req.Prizes != nil {
		prizes := lotteryPrizeInputsFromRequest(*req.Prizes)
		input.Prizes = &prizes
	}
	event, err := h.lotteryService.Update(c.Request.Context(), eventID, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, adminLotteryEventToResponse(event))
}

func (h *LotteryHandler) Delete(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || eventID <= 0 {
		response.BadRequest(c, "Invalid lottery ID")
		return
	}
	if err := h.lotteryService.Delete(c.Request.Context(), eventID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Lottery deleted successfully"})
}

func (h *LotteryHandler) Draw(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || eventID <= 0 {
		response.BadRequest(c, "Invalid lottery ID")
		return
	}
	event, err := h.lotteryService.Draw(c.Request.Context(), eventID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, adminLotteryEventToResponse(event))
}

func lotteryPrizeInputsFromRequest(items []lotteryPrizeRequest) []service.LotteryPrizeInput {
	out := make([]service.LotteryPrizeInput, 0, len(items))
	for _, item := range items {
		out = append(out, service.LotteryPrizeInput{
			Type:        item.Type,
			Name:        item.Name,
			Quantity:    item.Quantity,
			Amount:      item.Amount,
			CardContent: item.CardContent,
			SortOrder:   item.SortOrder,
		})
	}
	return out
}

type adminLotteryEventResponse struct {
	ID          int64                        `json:"id"`
	Title       string                       `json:"title"`
	Description string                       `json:"description"`
	Status      string                       `json:"status"`
	StartsAt    *int64                       `json:"starts_at,omitempty"`
	DrawAt      int64                        `json:"draw_at"`
	DrawnAt     *int64                       `json:"drawn_at,omitempty"`
	CreatedAt   int64                        `json:"created_at"`
	UpdatedAt   int64                        `json:"updated_at"`
	Prizes      []adminLotteryPrizeResponse  `json:"prizes"`
	EntryCount  int64                        `json:"entry_count"`
	Winners     []adminLotteryWinnerResponse `json:"winners"`
}

type adminLotteryPrizeResponse struct {
	ID          int64    `json:"id"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Quantity    int      `json:"quantity"`
	Amount      *float64 `json:"amount,omitempty"`
	CardContent *string  `json:"card_content,omitempty"`
	SortOrder   int      `json:"sort_order"`
}

type adminLotteryWinnerResponse struct {
	ID              int64    `json:"id"`
	EventID         int64    `json:"event_id"`
	PrizeID         int64    `json:"prize_id"`
	UserID          int64    `json:"user_id"`
	UserEmail       string   `json:"user_email"`
	UserDisplayName string   `json:"user_display_name"`
	PrizeType       string   `json:"prize_type"`
	PrizeName       string   `json:"prize_name"`
	Amount          *float64 `json:"amount,omitempty"`
	CardContent     *string  `json:"card_content,omitempty"`
	DeliveredAt     int64    `json:"delivered_at"`
}

func adminLotteryEventsToResponse(items []service.LotteryEvent) []adminLotteryEventResponse {
	out := make([]adminLotteryEventResponse, 0, len(items))
	for i := range items {
		out = append(out, adminLotteryEventToResponse(&items[i]))
	}
	return out
}

func adminLotteryEventToResponse(event *service.LotteryEvent) adminLotteryEventResponse {
	resp := adminLotteryEventResponse{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Status:      event.Status,
		DrawAt:      event.DrawAt.Unix(),
		CreatedAt:   event.CreatedAt.Unix(),
		UpdatedAt:   event.UpdatedAt.Unix(),
		EntryCount:  event.EntryCount,
		Prizes:      adminLotteryPrizesToResponse(event.Prizes),
		Winners:     adminLotteryWinnersToResponse(event.Winners),
	}
	if event.StartsAt != nil {
		v := event.StartsAt.Unix()
		resp.StartsAt = &v
	}
	if event.DrawnAt != nil {
		v := event.DrawnAt.Unix()
		resp.DrawnAt = &v
	}
	return resp
}

func adminLotteryPrizesToResponse(prizes []service.LotteryPrize) []adminLotteryPrizeResponse {
	out := make([]adminLotteryPrizeResponse, 0, len(prizes))
	for _, prize := range prizes {
		out = append(out, adminLotteryPrizeResponse{
			ID:          prize.ID,
			Type:        prize.Type,
			Name:        prize.Name,
			Quantity:    prize.Quantity,
			Amount:      prize.Amount,
			CardContent: prize.CardContent,
			SortOrder:   prize.SortOrder,
		})
	}
	return out
}

func adminLotteryWinnersToResponse(winners []service.LotteryWinner) []adminLotteryWinnerResponse {
	out := make([]adminLotteryWinnerResponse, 0, len(winners))
	for _, winner := range winners {
		out = append(out, adminLotteryWinnerResponse{
			ID:              winner.ID,
			EventID:         winner.EventID,
			PrizeID:         winner.PrizeID,
			UserID:          winner.UserID,
			UserEmail:       winner.UserEmail,
			UserDisplayName: winner.UserDisplayName,
			PrizeType:       winner.PrizeType,
			PrizeName:       winner.PrizeName,
			Amount:          winner.Amount,
			CardContent:     winner.CardContent,
			DeliveredAt:     winner.DeliveredAt.Unix(),
		})
	}
	return out
}
