package service

import (
	"context"
	"errors"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	LotteryStatusDraft    = "draft"
	LotteryStatusActive   = "active"
	LotteryStatusDrawn    = "drawn"
	LotteryStatusArchived = "archived"

	LotteryPrizeBalance = "balance"
	LotteryPrizeCard    = "card"
)

var (
	ErrLotteryNotFound       = infraerrors.NotFound("LOTTERY_NOT_FOUND", "lottery event not found")
	ErrLotteryInvalidInput   = infraerrors.BadRequest("LOTTERY_INVALID_INPUT", "invalid lottery input")
	ErrLotteryNotJoinable    = infraerrors.Conflict("LOTTERY_NOT_JOINABLE", "lottery is not joinable")
	ErrLotteryAlreadyJoined  = infraerrors.Conflict("LOTTERY_ALREADY_JOINED", "already joined this lottery")
	ErrLotteryAlreadyDrawn   = infraerrors.Conflict("LOTTERY_ALREADY_DRAWN", "lottery already drawn")
	ErrLotteryNoParticipants = infraerrors.Conflict("LOTTERY_NO_PARTICIPANTS", "lottery has no participants")
	ErrLotteryNoActivePrizes = infraerrors.Conflict("LOTTERY_NO_PRIZES", "lottery has no active prizes")
)

type LotteryEvent struct {
	ID          int64
	Title       string
	Description string
	Status      string
	StartsAt    *time.Time
	DrawAt      time.Time
	DrawnAt     *time.Time
	CreatedBy   *int64
	UpdatedBy   *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Prizes      []LotteryPrize
	EntryCount  int64
	Winners     []LotteryWinner
	Joined      bool
	MyWinner    *LotteryWinner
	MyJoinedAt  *time.Time
}

type LotteryPrize struct {
	ID          int64
	EventID     int64
	Type        string
	Name        string
	Quantity    int
	Amount      *float64
	CardContent *string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type LotteryWinner struct {
	ID              int64
	EventID         int64
	PrizeID         int64
	UserID          int64
	UserEmail       string
	UserDisplayName string
	PrizeType       string
	PrizeName       string
	Amount          *float64
	CardContent     *string
	DeliveredAt     time.Time
	CreatedAt       time.Time
}

type LotteryRepository interface {
	Create(ctx context.Context, event *LotteryEvent) (*LotteryEvent, error)
	Update(ctx context.Context, event *LotteryEvent) (*LotteryEvent, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*LotteryEvent, error)
	ListAdmin(ctx context.Context, params pagination.PaginationParams, status, search string) ([]LotteryEvent, *pagination.PaginationResult, error)
	ListForUser(ctx context.Context, userID int64, now time.Time) ([]LotteryEvent, error)
	Join(ctx context.Context, eventID, userID int64, now time.Time) (*LotteryEvent, error)
	Draw(ctx context.Context, eventID int64, now time.Time) (*LotteryEvent, error)
	ListDue(ctx context.Context, now time.Time, limit int) ([]int64, error)
}

type LotteryService struct {
	repo LotteryRepository
}

func NewLotteryService(repo LotteryRepository) *LotteryService {
	return &LotteryService{repo: repo}
}

type CreateLotteryInput struct {
	Title       string
	Description string
	Status      string
	StartsAt    *time.Time
	DrawAt      time.Time
	Prizes      []LotteryPrizeInput
	ActorID     *int64
}

type UpdateLotteryInput struct {
	Title       *string
	Description *string
	Status      *string
	StartsAt    **time.Time
	DrawAt      *time.Time
	Prizes      *[]LotteryPrizeInput
	ActorID     *int64
}

type LotteryPrizeInput struct {
	Type        string
	Name        string
	Quantity    int
	Amount      *float64
	CardContent *string
	SortOrder   int
}

func (s *LotteryService) Create(ctx context.Context, input *CreateLotteryInput) (*LotteryEvent, error) {
	if input == nil {
		return nil, ErrLotteryInvalidInput
	}
	title := strings.TrimSpace(input.Title)
	if title == "" || len(title) > 200 || input.DrawAt.IsZero() {
		return nil, ErrLotteryInvalidInput
	}
	status := normalizeLotteryStatus(input.Status)
	if status == "" || status == LotteryStatusDrawn {
		return nil, ErrLotteryInvalidInput
	}
	if input.StartsAt != nil && !input.StartsAt.Before(input.DrawAt) {
		return nil, ErrLotteryInvalidInput
	}
	prizes, err := normalizeLotteryPrizeInputs(input.Prizes)
	if err != nil {
		return nil, err
	}
	event := &LotteryEvent{
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Status:      status,
		StartsAt:    input.StartsAt,
		DrawAt:      input.DrawAt,
		CreatedBy:   input.ActorID,
		UpdatedBy:   input.ActorID,
		Prizes:      prizes,
	}
	return s.repo.Create(ctx, event)
}

func (s *LotteryService) Update(ctx context.Context, id int64, input *UpdateLotteryInput) (*LotteryEvent, error) {
	if id <= 0 || input == nil {
		return nil, ErrLotteryInvalidInput
	}
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.Status == LotteryStatusDrawn && input.Prizes != nil {
		return nil, ErrLotteryAlreadyDrawn
	}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" || len(title) > 200 {
			return nil, ErrLotteryInvalidInput
		}
		current.Title = title
	}
	if input.Description != nil {
		current.Description = strings.TrimSpace(*input.Description)
	}
	if input.Status != nil {
		status := normalizeLotteryStatus(*input.Status)
		if status == "" {
			return nil, ErrLotteryInvalidInput
		}
		current.Status = status
	}
	if input.StartsAt != nil {
		current.StartsAt = *input.StartsAt
	}
	if input.DrawAt != nil {
		if input.DrawAt.IsZero() {
			return nil, ErrLotteryInvalidInput
		}
		current.DrawAt = *input.DrawAt
	}
	if current.StartsAt != nil && !current.StartsAt.Before(current.DrawAt) {
		return nil, ErrLotteryInvalidInput
	}
	if input.Prizes != nil {
		prizes, err := normalizeLotteryPrizeInputs(*input.Prizes)
		if err != nil {
			return nil, err
		}
		current.Prizes = prizes
	}
	current.UpdatedBy = input.ActorID
	return s.repo.Update(ctx, current)
}

func (s *LotteryService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrLotteryInvalidInput
	}
	return s.repo.Delete(ctx, id)
}

func (s *LotteryService) GetByID(ctx context.Context, id int64) (*LotteryEvent, error) {
	if id <= 0 {
		return nil, ErrLotteryInvalidInput
	}
	return s.repo.GetByID(ctx, id)
}

func (s *LotteryService) ListAdmin(ctx context.Context, params pagination.PaginationParams, status, search string) ([]LotteryEvent, *pagination.PaginationResult, error) {
	return s.repo.ListAdmin(ctx, params, strings.TrimSpace(status), strings.TrimSpace(search))
}

func (s *LotteryService) ListForUser(ctx context.Context, userID int64) ([]LotteryEvent, error) {
	if userID <= 0 {
		return nil, ErrLotteryInvalidInput
	}
	return s.repo.ListForUser(ctx, userID, time.Now())
}

func (s *LotteryService) Join(ctx context.Context, userID, eventID int64) (*LotteryEvent, error) {
	if userID <= 0 || eventID <= 0 {
		return nil, ErrLotteryInvalidInput
	}
	return s.repo.Join(ctx, eventID, userID, time.Now())
}

func (s *LotteryService) Draw(ctx context.Context, eventID int64) (*LotteryEvent, error) {
	if eventID <= 0 {
		return nil, ErrLotteryInvalidInput
	}
	return s.repo.Draw(ctx, eventID, time.Now())
}

func (s *LotteryService) DrawDue(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 20
	}
	now := time.Now()
	ids, err := s.repo.ListDue(ctx, now, limit)
	if err != nil {
		return 0, err
	}
	drawn := 0
	for _, id := range ids {
		if _, err := s.repo.Draw(ctx, id, now); err != nil {
			if errors.Is(err, ErrLotteryNoParticipants) || errors.Is(err, ErrLotteryNoActivePrizes) || errors.Is(err, ErrLotteryAlreadyDrawn) {
				continue
			}
			return drawn, err
		}
		drawn++
	}
	return drawn, nil
}

func normalizeLotteryStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "":
		return LotteryStatusDraft
	case LotteryStatusDraft, LotteryStatusActive, LotteryStatusDrawn, LotteryStatusArchived:
		return strings.TrimSpace(status)
	default:
		return ""
	}
}

func normalizeLotteryPrizeInputs(inputs []LotteryPrizeInput) ([]LotteryPrize, error) {
	if len(inputs) == 0 {
		return nil, ErrLotteryNoActivePrizes
	}
	prizes := make([]LotteryPrize, 0, len(inputs))
	for i, input := range inputs {
		name := strings.TrimSpace(input.Name)
		typ := strings.TrimSpace(input.Type)
		if name == "" || len(name) > 200 || input.Quantity <= 0 {
			return nil, ErrLotteryInvalidInput
		}
		if input.SortOrder == 0 {
			input.SortOrder = i + 1
		}
		prize := LotteryPrize{
			Type:      typ,
			Name:      name,
			Quantity:  input.Quantity,
			SortOrder: input.SortOrder,
		}
		switch typ {
		case LotteryPrizeBalance:
			if input.Amount == nil || *input.Amount <= 0 {
				return nil, ErrLotteryInvalidInput
			}
			prize.Amount = input.Amount
		case LotteryPrizeCard:
			if input.CardContent == nil || strings.TrimSpace(*input.CardContent) == "" {
				return nil, ErrLotteryInvalidInput
			}
			card := strings.TrimSpace(*input.CardContent)
			prize.CardContent = &card
		default:
			return nil, ErrLotteryInvalidInput
		}
		prizes = append(prizes, prize)
	}
	return prizes, nil
}
