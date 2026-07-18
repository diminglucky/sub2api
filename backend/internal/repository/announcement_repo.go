package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/announcement"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

type announcementRepository struct {
	client *dbent.Client
	db     *sql.DB
}

func NewAnnouncementRepository(client *dbent.Client, db *sql.DB) service.AnnouncementRepository {
	return &announcementRepository{client: client, db: db}
}

func (r *announcementRepository) Create(ctx context.Context, a *service.Announcement) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Announcement.Create().
		SetTitle(a.Title).
		SetContent(a.Content).
		SetStatus(a.Status).
		SetNotifyMode(a.NotifyMode).
		SetTargeting(a.Targeting)

	if a.StartsAt != nil {
		builder.SetStartsAt(*a.StartsAt)
	}
	if a.EndsAt != nil {
		builder.SetEndsAt(*a.EndsAt)
	}
	if a.CreatedBy != nil {
		builder.SetCreatedBy(*a.CreatedBy)
	}
	if a.UpdatedBy != nil {
		builder.SetUpdatedBy(*a.UpdatedBy)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	applyAnnouncementEntityToService(a, created)
	return nil
}

func (r *announcementRepository) CreateWithEmailBatch(ctx context.Context, a *service.Announcement, batch *service.AnnouncementEmailBatch) (err error) {
	if batch == nil {
		return r.Create(ctx, a)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	targeting, err := json.Marshal(a.Targeting)
	if err != nil {
		return fmt.Errorf("encode announcement targeting: %w", err)
	}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO announcements (
			title, content, status, notify_mode, targeting, starts_at, ends_at,
			created_by, updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`, a.Title, a.Content, a.Status, a.NotifyMode, targeting, a.StartsAt, a.EndsAt, a.CreatedBy, a.UpdatedBy).
		Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return err
	}
	batch.AnnouncementID = a.ID
	if err = insertAnnouncementEmailBatch(ctx, tx, batch); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *announcementRepository) GetByID(ctx context.Context, id int64) (*service.Announcement, error) {
	m, err := r.client.Announcement.Query().
		Where(announcement.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrAnnouncementNotFound, nil)
	}
	return announcementEntityToService(m), nil
}

func (r *announcementRepository) Update(ctx context.Context, a *service.Announcement) error {
	client := clientFromContext(ctx, r.client)
	builder := client.Announcement.UpdateOneID(a.ID).
		SetTitle(a.Title).
		SetContent(a.Content).
		SetStatus(a.Status).
		SetNotifyMode(a.NotifyMode).
		SetTargeting(a.Targeting)

	if a.StartsAt != nil {
		builder.SetStartsAt(*a.StartsAt)
	} else {
		builder.ClearStartsAt()
	}
	if a.EndsAt != nil {
		builder.SetEndsAt(*a.EndsAt)
	} else {
		builder.ClearEndsAt()
	}
	if a.CreatedBy != nil {
		builder.SetCreatedBy(*a.CreatedBy)
	} else {
		builder.ClearCreatedBy()
	}
	if a.UpdatedBy != nil {
		builder.SetUpdatedBy(*a.UpdatedBy)
	} else {
		builder.ClearUpdatedBy()
	}

	updated, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, service.ErrAnnouncementNotFound, nil)
	}

	a.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *announcementRepository) UpdateWithEmailBatch(ctx context.Context, a *service.Announcement, batch *service.AnnouncementEmailBatch) (err error) {
	if batch == nil {
		return r.Update(ctx, a)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	targeting, err := json.Marshal(a.Targeting)
	if err != nil {
		return fmt.Errorf("encode announcement targeting: %w", err)
	}
	err = tx.QueryRowContext(ctx, `
		UPDATE announcements
		SET title = $2, content = $3, status = $4, notify_mode = $5, targeting = $6,
			starts_at = $7, ends_at = $8, created_by = $9, updated_by = $10, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`, a.ID, a.Title, a.Content, a.Status, a.NotifyMode, targeting,
		a.StartsAt, a.EndsAt, a.CreatedBy, a.UpdatedBy).Scan(&a.UpdatedAt)
	if err != nil {
		return translatePersistenceError(err, service.ErrAnnouncementNotFound, nil)
	}
	batch.AnnouncementID = a.ID
	if err = insertAnnouncementEmailBatch(ctx, tx, batch); err != nil {
		return err
	}
	return tx.Commit()
}

func insertAnnouncementEmailBatch(ctx context.Context, tx *sql.Tx, batch *service.AnnouncementEmailBatch) error {
	recipients, err := json.Marshal(batch.Recipients)
	if err != nil {
		return fmt.Errorf("encode announcement email recipients: %w", err)
	}
	status := service.AnnouncementEmailBatchStatusPending
	var completedAt any
	if len(batch.Recipients) == 0 {
		status = service.AnnouncementEmailBatchStatusCompleted
		completed := time.Now()
		completedAt = completed
		batch.CompletedAt = &completed
	}
	if batch.MaxAttempts <= 0 {
		batch.MaxAttempts = 5
	}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO announcement_email_batches (
			announcement_id, campaign_id, title, content, recipients, status,
			max_attempts, total_count, next_attempt_at, completed_at
		) VALUES ($1, $2::uuid, $3, $4, $5, $6, $7, $8, NOW(), $9)
		RETURNING id, status, next_attempt_at, created_at, updated_at
	`, batch.AnnouncementID, batch.CampaignID, batch.Title, batch.Content, recipients,
		status, batch.MaxAttempts, len(batch.Recipients), completedAt).
		Scan(&batch.ID, &batch.Status, &batch.NextAttemptAt, &batch.CreatedAt, &batch.UpdatedAt)
	if err != nil {
		return err
	}
	batch.TotalCount = len(batch.Recipients)
	return nil
}

func (r *announcementRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.Announcement.Delete().Where(announcement.IDEQ(id)).Exec(ctx)
	return err
}

func (r *announcementRepository) List(
	ctx context.Context,
	params pagination.PaginationParams,
	filters service.AnnouncementListFilters,
) ([]service.Announcement, *pagination.PaginationResult, error) {
	q := r.client.Announcement.Query()

	if filters.Status != "" {
		q = q.Where(announcement.StatusEQ(filters.Status))
	}
	if filters.Search != "" {
		q = q.Where(
			announcement.Or(
				announcement.TitleContainsFold(filters.Search),
				announcement.ContentContainsFold(filters.Search),
			),
		)
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	itemsQuery := q.
		Offset(params.Offset()).
		Limit(params.Limit())
	for _, order := range announcementListOrders(params) {
		itemsQuery = itemsQuery.Order(order)
	}

	items, err := itemsQuery.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	out := announcementEntitiesToService(items)
	return out, paginationResultFromTotal(int64(total), params), nil
}

func announcementListOrder(params pagination.PaginationParams) (string, string) {
	sortBy := strings.ToLower(strings.TrimSpace(params.SortBy))
	sortOrder := params.NormalizedSortOrder(pagination.SortOrderDesc)

	switch sortBy {
	case "title":
		return announcement.FieldTitle, sortOrder
	case "status":
		return announcement.FieldStatus, sortOrder
	case "notify_mode":
		return announcement.FieldNotifyMode, sortOrder
	case "starts_at":
		return announcement.FieldStartsAt, sortOrder
	case "ends_at":
		return announcement.FieldEndsAt, sortOrder
	case "id":
		return announcement.FieldID, sortOrder
	case "", "created_at":
		return announcement.FieldCreatedAt, sortOrder
	default:
		return announcement.FieldCreatedAt, pagination.SortOrderDesc
	}
}

func announcementListOrders(params pagination.PaginationParams) []func(*entsql.Selector) {
	field, sortOrder := announcementListOrder(params)

	if sortOrder == pagination.SortOrderAsc {
		if field == announcement.FieldID {
			return []func(*entsql.Selector){
				dbent.Asc(field),
			}
		}
		return []func(*entsql.Selector){
			dbent.Asc(field),
			dbent.Asc(announcement.FieldID),
		}
	}

	if field == announcement.FieldID {
		return []func(*entsql.Selector){
			dbent.Desc(field),
		}
	}
	return []func(*entsql.Selector){
		dbent.Desc(field),
		dbent.Desc(announcement.FieldID),
	}
}

func (r *announcementRepository) ListActive(ctx context.Context, now time.Time) ([]service.Announcement, error) {
	q := r.client.Announcement.Query().
		Where(
			announcement.StatusEQ(service.AnnouncementStatusActive),
			announcement.Or(announcement.StartsAtIsNil(), announcement.StartsAtLTE(now)),
			announcement.Or(announcement.EndsAtIsNil(), announcement.EndsAtGT(now)),
		).
		Order(dbent.Desc(announcement.FieldID)).
		Limit(200)

	items, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	return announcementEntitiesToService(items), nil
}

func applyAnnouncementEntityToService(dst *service.Announcement, src *dbent.Announcement) {
	if dst == nil || src == nil {
		return
	}
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}

func announcementEntityToService(m *dbent.Announcement) *service.Announcement {
	if m == nil {
		return nil
	}
	return &service.Announcement{
		ID:         m.ID,
		Title:      m.Title,
		Content:    m.Content,
		Status:     m.Status,
		NotifyMode: m.NotifyMode,
		Targeting:  m.Targeting,
		StartsAt:   m.StartsAt,
		EndsAt:     m.EndsAt,
		CreatedBy:  m.CreatedBy,
		UpdatedBy:  m.UpdatedBy,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func announcementEntitiesToService(models []*dbent.Announcement) []service.Announcement {
	out := make([]service.Announcement, 0, len(models))
	for i := range models {
		if s := announcementEntityToService(models[i]); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
