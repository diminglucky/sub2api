package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type lotteryRepository struct {
	db *sql.DB
}

func NewLotteryRepository(db *sql.DB) service.LotteryRepository {
	return &lotteryRepository{db: db}
}

func (r *lotteryRepository) Create(ctx context.Context, event *service.LotteryEvent) (*service.LotteryEvent, error) {
	err := r.runInTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO lottery_events (title, description, status, starts_at, draw_at, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at
		`, event.Title, event.Description, event.Status, event.StartsAt, event.DrawAt, event.CreatedBy, event.UpdatedBy).
			Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt); err != nil {
			return fmt.Errorf("insert lottery event: %w", err)
		}
		return replaceLotteryPrizes(ctx, tx, event.ID, event.Prizes)
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, event.ID)
}

func (r *lotteryRepository) Update(ctx context.Context, event *service.LotteryEvent) (*service.LotteryEvent, error) {
	err := r.runInTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			UPDATE lottery_events
			SET title = $2, description = $3, status = $4, starts_at = $5, draw_at = $6, updated_by = $7, updated_at = NOW()
			WHERE id = $1
		`, event.ID, event.Title, event.Description, event.Status, event.StartsAt, event.DrawAt, event.UpdatedBy)
		if err != nil {
			return fmt.Errorf("update lottery event: %w", err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return service.ErrLotteryNotFound
		}
		if event.Prizes != nil {
			if _, err := tx.ExecContext(ctx, `DELETE FROM lottery_prizes WHERE event_id = $1`, event.ID); err != nil {
				return fmt.Errorf("delete lottery prizes: %w", err)
			}
			if err := replaceLotteryPrizes(ctx, tx, event.ID, event.Prizes); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, event.ID)
}

func (r *lotteryRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM lottery_events WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete lottery event: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return service.ErrLotteryNotFound
	}
	return nil
}

func (r *lotteryRepository) GetByID(ctx context.Context, id int64) (*service.LotteryEvent, error) {
	event, err := scanLotteryEvent(r.db.QueryRowContext(ctx, `
		SELECT id, title, description, status, starts_at, draw_at, drawn_at, created_by, updated_by, created_at, updated_at
		FROM lottery_events WHERE id = $1
	`, id))
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(ctx, event, 0, false); err != nil {
		return nil, err
	}
	return event, nil
}

func (r *lotteryRepository) ListAdmin(ctx context.Context, params pagination.PaginationParams, status, search string) ([]service.LotteryEvent, *pagination.PaginationResult, error) {
	params = normalizeLotteryPagination(params)
	where := []string{"1 = 1"}
	args := []any{}
	if status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", len(args), len(args)))
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM lottery_events WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("count lottery events: %w", err)
	}

	orderBy := lotteryOrderBy(params)
	queryArgs := append(args, params.PageSize, params.Offset())
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, status, starts_at, draw_at, drawn_at, created_by, updated_by, created_at, updated_at
		FROM lottery_events
		WHERE `+whereSQL+`
		ORDER BY `+orderBy+`
		LIMIT $`+fmt.Sprint(len(args)+1)+` OFFSET $`+fmt.Sprint(len(args)+2), queryArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("list lottery events: %w", err)
	}
	defer rows.Close()

	items, err := scanLotteryEvents(rows)
	if err != nil {
		return nil, nil, err
	}
	for i := range items {
		if err := r.hydrate(ctx, &items[i], 0, false); err != nil {
			return nil, nil, err
		}
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *lotteryRepository) ListForUser(ctx context.Context, userID int64, now time.Time) ([]service.LotteryEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, status, starts_at, draw_at, drawn_at, created_by, updated_by, created_at, updated_at
		FROM lottery_events
		WHERE status IN ('active', 'drawn')
		  AND (starts_at IS NULL OR starts_at <= $1)
		ORDER BY CASE status WHEN 'active' THEN 0 ELSE 1 END, draw_at ASC, id DESC
	`, now)
	if err != nil {
		return nil, fmt.Errorf("list user lottery events: %w", err)
	}
	defer rows.Close()
	items, err := scanLotteryEvents(rows)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if err := r.hydrate(ctx, &items[i], userID, true); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (r *lotteryRepository) Join(ctx context.Context, eventID, userID int64, now time.Time) (*service.LotteryEvent, error) {
	err := r.runInTx(ctx, func(tx *sql.Tx) error {
		var status string
		var startsAt *time.Time
		var drawAt time.Time
		err := tx.QueryRowContext(ctx, `
			SELECT status, starts_at, draw_at
			FROM lottery_events
			WHERE id = $1
			FOR UPDATE
		`, eventID).Scan(&status, &startsAt, &drawAt)
		if err == sql.ErrNoRows {
			return service.ErrLotteryNotFound
		}
		if err != nil {
			return fmt.Errorf("load lottery event for join: %w", err)
		}
		if status != service.LotteryStatusActive || (startsAt != nil && now.Before(*startsAt)) || !now.Before(drawAt) {
			return service.ErrLotteryNotJoinable
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO lottery_entries (event_id, user_id, joined_at)
			VALUES ($1, $2, $3)
		`, eventID, userID, now); err != nil {
			if isUniqueViolation(err) {
				return service.ErrLotteryAlreadyJoined
			}
			return fmt.Errorf("join lottery: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	event, err := r.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if err := r.hydrateUserState(ctx, event, userID); err != nil {
		return nil, err
	}
	return event, nil
}

func (r *lotteryRepository) Draw(ctx context.Context, eventID int64, now time.Time) (*service.LotteryEvent, error) {
	err := r.runInTx(ctx, func(tx *sql.Tx) error {
		var status string
		res, err := tx.ExecContext(ctx, `
			UPDATE lottery_events
			SET status = 'drawn', drawn_at = $2, updated_at = NOW()
			WHERE id = $1 AND status = 'active' AND drawn_at IS NULL
		`, eventID, now)
		if err != nil {
			return fmt.Errorf("mark lottery drawn: %w", err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			err := tx.QueryRowContext(ctx, `SELECT status FROM lottery_events WHERE id = $1`, eventID).Scan(&status)
			if err == sql.ErrNoRows {
				return service.ErrLotteryNotFound
			}
			if err != nil {
				return fmt.Errorf("load lottery status: %w", err)
			}
			return service.ErrLotteryAlreadyDrawn
		}

		prizes, err := listLotteryPrizesTx(ctx, tx, eventID)
		if err != nil {
			return err
		}
		if len(prizes) == 0 {
			return service.ErrLotteryNoActivePrizes
		}
		userIDs, err := listLotteryParticipantIDsTx(ctx, tx, eventID)
		if err != nil {
			return err
		}
		if len(userIDs) == 0 {
			return service.ErrLotteryNoParticipants
		}
		idx := 0
		for _, prize := range prizes {
			for i := 0; i < prize.Quantity && idx < len(userIDs); i++ {
				userID := userIDs[idx]
				idx++
				if prize.Type == service.LotteryPrizeBalance && prize.Amount != nil {
					if _, err := tx.ExecContext(ctx, `
						UPDATE users
						SET balance = balance + $2, total_recharged = total_recharged + $2, updated_at = NOW()
						WHERE id = $1 AND deleted_at IS NULL
					`, userID, *prize.Amount); err != nil {
						return fmt.Errorf("deliver lottery balance: %w", err)
					}
				}
				if _, err := tx.ExecContext(ctx, `
					INSERT INTO lottery_winners (event_id, prize_id, user_id, prize_type, prize_name, amount, card_content, delivered_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
					ON CONFLICT (event_id, user_id) DO NOTHING
				`, eventID, prize.ID, userID, prize.Type, prize.Name, prize.Amount, prize.CardContent, now); err != nil {
					return fmt.Errorf("insert lottery winner: %w", err)
				}
			}
			if idx >= len(userIDs) {
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, eventID)
}

func (r *lotteryRepository) ListDue(ctx context.Context, now time.Time, limit int) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id
		FROM lottery_events
		WHERE status = 'active' AND drawn_at IS NULL AND draw_at <= $1
		ORDER BY draw_at ASC, id ASC
		LIMIT $2
	`, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list due lotteries: %w", err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *lotteryRepository) runInTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func replaceLotteryPrizes(ctx context.Context, tx *sql.Tx, eventID int64, prizes []service.LotteryPrize) error {
	for _, prize := range prizes {
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO lottery_prizes (event_id, type, name, quantity, amount, card_content, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at
		`, eventID, prize.Type, prize.Name, prize.Quantity, prize.Amount, prize.CardContent, prize.SortOrder).
			Scan(&prize.ID, &prize.CreatedAt, &prize.UpdatedAt); err != nil {
			return fmt.Errorf("insert lottery prize: %w", err)
		}
	}
	return nil
}

func (r *lotteryRepository) hydrate(ctx context.Context, event *service.LotteryEvent, userID int64, maskWinnerCards bool) error {
	prizes, err := r.listPrizes(ctx, event.ID)
	if err != nil {
		return err
	}
	event.Prizes = prizes
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM lottery_entries WHERE event_id = $1`, event.ID).Scan(&event.EntryCount); err != nil {
		return fmt.Errorf("count lottery entries: %w", err)
	}
	winners, err := r.listWinners(ctx, event.ID, maskWinnerCards)
	if err != nil {
		return err
	}
	event.Winners = winners
	if userID > 0 {
		return r.hydrateUserState(ctx, event, userID)
	}
	return nil
}

func (r *lotteryRepository) hydrateUserState(ctx context.Context, event *service.LotteryEvent, userID int64) error {
	var joinedAt time.Time
	err := r.db.QueryRowContext(ctx, `SELECT joined_at FROM lottery_entries WHERE event_id = $1 AND user_id = $2`, event.ID, userID).Scan(&joinedAt)
	if err == nil {
		event.Joined = true
		event.MyJoinedAt = &joinedAt
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("load lottery entry state: %w", err)
	}
	winners, err := r.listWinnersByUser(ctx, event.ID, userID)
	if err != nil {
		return err
	}
	if len(winners) > 0 {
		event.MyWinner = &winners[0]
	}
	return nil
}

func (r *lotteryRepository) listPrizes(ctx context.Context, eventID int64) ([]service.LotteryPrize, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_id, type, name, quantity, amount, card_content, sort_order, created_at, updated_at
		FROM lottery_prizes
		WHERE event_id = $1
		ORDER BY sort_order ASC, id ASC
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("list lottery prizes: %w", err)
	}
	defer rows.Close()
	return scanLotteryPrizes(rows)
}

func (r *lotteryRepository) listWinners(ctx context.Context, eventID int64, maskCards bool) ([]service.LotteryWinner, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT w.id, w.event_id, w.prize_id, w.user_id, COALESCE(u.email, ''),
		       COALESCE(NULLIF(BTRIM(uav.value), ''), NULLIF(split_part(COALESCE(u.email, ''), '@', 1), ''), CONCAT('用户', w.user_id::text)),
		       w.prize_type, w.prize_name, w.amount,
		       CASE WHEN $2 THEN NULL ELSE w.card_content END, w.delivered_at, w.created_at
		FROM lottery_winners w
		LEFT JOIN users u ON u.id = w.user_id
		LEFT JOIN user_attribute_definitions uad ON uad.key = 'username' AND uad.deleted_at IS NULL
		LEFT JOIN user_attribute_values uav ON uav.attribute_id = uad.id AND uav.user_id = w.user_id
		WHERE w.event_id = $1
		ORDER BY w.id ASC
	`, eventID, maskCards)
	if err != nil {
		return nil, fmt.Errorf("list lottery winners: %w", err)
	}
	defer rows.Close()
	return scanLotteryWinners(rows)
}

func (r *lotteryRepository) listWinnersByUser(ctx context.Context, eventID, userID int64) ([]service.LotteryWinner, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT w.id, w.event_id, w.prize_id, w.user_id, COALESCE(u.email, ''),
		       COALESCE(NULLIF(BTRIM(uav.value), ''), NULLIF(split_part(COALESCE(u.email, ''), '@', 1), ''), CONCAT('用户', w.user_id::text)),
		       w.prize_type, w.prize_name, w.amount,
		       w.card_content, w.delivered_at, w.created_at
		FROM lottery_winners w
		LEFT JOIN users u ON u.id = w.user_id
		LEFT JOIN user_attribute_definitions uad ON uad.key = 'username' AND uad.deleted_at IS NULL
		LEFT JOIN user_attribute_values uav ON uav.attribute_id = uad.id AND uav.user_id = w.user_id
		WHERE w.event_id = $1 AND w.user_id = $2
		ORDER BY w.id ASC
	`, eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("list user lottery winners: %w", err)
	}
	defer rows.Close()
	return scanLotteryWinners(rows)
}

func listLotteryPrizesTx(ctx context.Context, tx *sql.Tx, eventID int64) ([]service.LotteryPrize, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, event_id, type, name, quantity, amount, card_content, sort_order, created_at, updated_at
		FROM lottery_prizes
		WHERE event_id = $1
		ORDER BY sort_order ASC, id ASC
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("list lottery prizes: %w", err)
	}
	defer rows.Close()
	return scanLotteryPrizes(rows)
}

func listLotteryParticipantIDsTx(ctx context.Context, tx *sql.Tx, eventID int64) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT user_id
		FROM lottery_entries
		WHERE event_id = $1
		ORDER BY random()
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("list lottery participants: %w", err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func scanLotteryEvent(row interface{ Scan(dest ...any) error }) (*service.LotteryEvent, error) {
	event := &service.LotteryEvent{}
	if err := row.Scan(&event.ID, &event.Title, &event.Description, &event.Status, &event.StartsAt, &event.DrawAt, &event.DrawnAt, &event.CreatedBy, &event.UpdatedBy, &event.CreatedAt, &event.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrLotteryNotFound
		}
		return nil, fmt.Errorf("scan lottery event: %w", err)
	}
	return event, nil
}

func scanLotteryEvents(rows *sql.Rows) ([]service.LotteryEvent, error) {
	var items []service.LotteryEvent
	for rows.Next() {
		event, err := scanLotteryEvent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *event)
	}
	return items, rows.Err()
}

func scanLotteryPrizes(rows *sql.Rows) ([]service.LotteryPrize, error) {
	var prizes []service.LotteryPrize
	for rows.Next() {
		var prize service.LotteryPrize
		if err := rows.Scan(&prize.ID, &prize.EventID, &prize.Type, &prize.Name, &prize.Quantity, &prize.Amount, &prize.CardContent, &prize.SortOrder, &prize.CreatedAt, &prize.UpdatedAt); err != nil {
			return nil, err
		}
		prizes = append(prizes, prize)
	}
	return prizes, rows.Err()
}

func scanLotteryWinners(rows *sql.Rows) ([]service.LotteryWinner, error) {
	var winners []service.LotteryWinner
	for rows.Next() {
		var winner service.LotteryWinner
		if err := rows.Scan(&winner.ID, &winner.EventID, &winner.PrizeID, &winner.UserID, &winner.UserEmail, &winner.UserDisplayName, &winner.PrizeType, &winner.PrizeName, &winner.Amount, &winner.CardContent, &winner.DeliveredAt, &winner.CreatedAt); err != nil {
			return nil, err
		}
		winners = append(winners, winner)
	}
	return winners, rows.Err()
}

func normalizeLotteryPagination(params pagination.PaginationParams) pagination.PaginationParams {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	if params.SortOrder != pagination.SortOrderAsc {
		params.SortOrder = pagination.SortOrderDesc
	}
	return params
}

func lotteryOrderBy(params pagination.PaginationParams) string {
	order := "DESC"
	if params.SortOrder == pagination.SortOrderAsc {
		order = "ASC"
	}
	switch params.SortBy {
	case "title":
		return "title " + order + ", id DESC"
	case "status":
		return "status " + order + ", id DESC"
	case "draw_at":
		return "draw_at " + order + ", id DESC"
	case "created_at":
		return "created_at " + order + ", id DESC"
	default:
		return "created_at DESC, id DESC"
	}
}
