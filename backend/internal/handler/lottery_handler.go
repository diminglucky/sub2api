package handler

import (
	"strconv"

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

func (h *LotteryHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	items, err := h.lotteryService.ListForUser(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, lotteryEventsToResponse(items, false))
}

func (h *LotteryHandler) Join(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || eventID <= 0 {
		response.BadRequest(c, "Invalid lottery ID")
		return
	}
	event, err := h.lotteryService.Join(c.Request.Context(), subject.UserID, eventID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, lotteryEventToResponse(event, false))
}

type lotteryEventResponse struct {
	ID          int64                   `json:"id"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Status      string                  `json:"status"`
	StartsAt    *int64                  `json:"starts_at,omitempty"`
	DrawAt      int64                   `json:"draw_at"`
	DrawnAt     *int64                  `json:"drawn_at,omitempty"`
	CreatedAt   int64                   `json:"created_at"`
	UpdatedAt   int64                   `json:"updated_at"`
	Prizes      []lotteryPrizeResponse  `json:"prizes"`
	EntryCount  int64                   `json:"entry_count"`
	Winners     []lotteryWinnerResponse `json:"winners,omitempty"`
	Joined      bool                    `json:"joined"`
	MyJoinedAt  *int64                  `json:"my_joined_at,omitempty"`
	MyWinner    *lotteryWinnerResponse  `json:"my_winner,omitempty"`
}

type lotteryPrizeResponse struct {
	ID          int64    `json:"id"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Quantity    int      `json:"quantity"`
	Amount      *float64 `json:"amount,omitempty"`
	CardContent *string  `json:"card_content,omitempty"`
	SortOrder   int      `json:"sort_order"`
}

type lotteryWinnerResponse struct {
	ID              int64    `json:"id"`
	EventID         int64    `json:"event_id"`
	PrizeID         int64    `json:"prize_id"`
	UserID          int64    `json:"user_id"`
	UserEmail       string   `json:"user_email,omitempty"`
	UserDisplayName string   `json:"user_display_name,omitempty"`
	PrizeType       string   `json:"prize_type"`
	PrizeName       string   `json:"prize_name"`
	Amount          *float64 `json:"amount,omitempty"`
	CardContent     *string  `json:"card_content,omitempty"`
	DeliveredAt     int64    `json:"delivered_at"`
}

func lotteryEventsToResponse(items []service.LotteryEvent, adminView bool) []lotteryEventResponse {
	out := make([]lotteryEventResponse, 0, len(items))
	for i := range items {
		out = append(out, lotteryEventToResponse(&items[i], adminView))
	}
	return out
}

func lotteryEventToResponse(event *service.LotteryEvent, adminView bool) lotteryEventResponse {
	resp := lotteryEventResponse{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Status:      event.Status,
		DrawAt:      event.DrawAt.Unix(),
		CreatedAt:   event.CreatedAt.Unix(),
		UpdatedAt:   event.UpdatedAt.Unix(),
		EntryCount:  event.EntryCount,
		Joined:      event.Joined,
		Prizes:      lotteryPrizesToResponse(event.Prizes, adminView),
	}
	if adminView || len(event.Winners) > 0 {
		resp.Winners = lotteryWinnersToResponse(event.Winners)
	}
	if event.StartsAt != nil {
		v := event.StartsAt.Unix()
		resp.StartsAt = &v
	}
	if event.DrawnAt != nil {
		v := event.DrawnAt.Unix()
		resp.DrawnAt = &v
	}
	if event.MyJoinedAt != nil {
		v := event.MyJoinedAt.Unix()
		resp.MyJoinedAt = &v
	}
	if event.MyWinner != nil {
		w := lotteryWinnerToResponse(*event.MyWinner)
		resp.MyWinner = &w
	}
	return resp
}

func lotteryPrizesToResponse(prizes []service.LotteryPrize, includeCard bool) []lotteryPrizeResponse {
	out := make([]lotteryPrizeResponse, 0, len(prizes))
	for _, prize := range prizes {
		item := lotteryPrizeResponse{
			ID:        prize.ID,
			Type:      prize.Type,
			Name:      prize.Name,
			Quantity:  prize.Quantity,
			Amount:    prize.Amount,
			SortOrder: prize.SortOrder,
		}
		if includeCard {
			item.CardContent = prize.CardContent
		}
		out = append(out, item)
	}
	return out
}

func lotteryWinnersToResponse(winners []service.LotteryWinner) []lotteryWinnerResponse {
	out := make([]lotteryWinnerResponse, 0, len(winners))
	for _, winner := range winners {
		out = append(out, lotteryWinnerToResponse(winner))
	}
	return out
}

func lotteryWinnerToResponse(winner service.LotteryWinner) lotteryWinnerResponse {
	return lotteryWinnerResponse{
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
	}
}
