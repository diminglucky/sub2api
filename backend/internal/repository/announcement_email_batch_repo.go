package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type announcementEmailBatchRepository struct {
	db *sql.DB
}

func NewAnnouncementEmailBatchRepository(db *sql.DB) service.AnnouncementEmailBatchRepository {
	return &announcementEmailBatchRepository{db: db}
}

func (r *announcementEmailBatchRepository) ClaimDue(ctx context.Context, staleAfter time.Duration) (*service.AnnouncementEmailBatch, error) {
	if staleAfter <= 0 {
		staleAfter = 10 * time.Minute
	}
	row := r.db.QueryRowContext(ctx, `
		WITH next AS (
			SELECT id
			FROM announcement_email_batches
			WHERE attempt_count < max_attempts
			  AND (
				(status IN ('pending', 'retrying') AND next_attempt_at <= NOW())
				OR (status = 'processing' AND locked_at < NOW() - ($1 * interval '1 second'))
			  )
			ORDER BY next_attempt_at ASC, id ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE announcement_email_batches AS batches
		SET status = 'processing',
			attempt_count = attempt_count + 1,
			locked_at = NOW(),
			updated_at = NOW()
		FROM next
		WHERE batches.id = next.id
		RETURNING batches.id, batches.announcement_id, batches.campaign_id::text,
			batches.title, batches.content, batches.recipients, batches.status,
			batches.attempt_count, batches.max_attempts, batches.total_count,
			batches.processed_count, batches.failed_count, COALESCE(batches.last_error, ''),
			batches.next_attempt_at, batches.locked_at, batches.completed_at,
			batches.created_at, batches.updated_at
	`, int64(staleAfter/time.Second))

	batch, err := scanAnnouncementEmailBatch(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return batch, err
}

func (r *announcementEmailBatchRepository) MarkCompleted(ctx context.Context, id int64, processedCount int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE announcement_email_batches
		SET status = 'completed', processed_count = $2, failed_count = 0,
			last_error = NULL, locked_at = NULL, completed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'processing'
	`, id, processedCount)
	return err
}

func (r *announcementEmailBatchRepository) RefreshLock(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE announcement_email_batches
		SET locked_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'processing'
	`, id)
	return err
}

func (r *announcementEmailBatchRepository) MarkRetry(
	ctx context.Context,
	id int64,
	processedCount, failedCount int,
	lastError string,
	nextAttemptAt time.Time,
) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE announcement_email_batches
		SET status = CASE WHEN attempt_count >= max_attempts THEN 'failed' ELSE 'retrying' END,
			processed_count = $2,
			failed_count = $3,
			last_error = $4,
			next_attempt_at = $5,
			locked_at = NULL,
			completed_at = CASE WHEN attempt_count >= max_attempts THEN NOW() ELSE NULL END,
			updated_at = NOW()
		WHERE id = $1 AND status = 'processing'
	`, id, processedCount, failedCount, lastError, nextAttemptAt)
	return err
}

func (r *announcementEmailBatchRepository) ListByAnnouncement(ctx context.Context, announcementID int64) ([]service.AnnouncementEmailBatch, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, announcement_id, campaign_id::text, title, content, recipients, status,
			attempt_count, max_attempts, total_count, processed_count, failed_count,
			COALESCE(last_error, ''), next_attempt_at, locked_at, completed_at, created_at, updated_at
		FROM announcement_email_batches
		WHERE announcement_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 100
	`, announcementID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.AnnouncementEmailBatch, 0)
	for rows.Next() {
		batch, err := scanAnnouncementEmailBatch(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *batch)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type announcementEmailBatchRowScanner interface {
	Scan(dest ...any) error
}

func scanAnnouncementEmailBatch(row announcementEmailBatchRowScanner) (*service.AnnouncementEmailBatch, error) {
	var (
		batch         service.AnnouncementEmailBatch
		recipientsRaw []byte
		lockedAt      sql.NullTime
		completedAt   sql.NullTime
	)
	if err := row.Scan(
		&batch.ID, &batch.AnnouncementID, &batch.CampaignID,
		&batch.Title, &batch.Content, &recipientsRaw, &batch.Status,
		&batch.AttemptCount, &batch.MaxAttempts, &batch.TotalCount,
		&batch.ProcessedCount, &batch.FailedCount, &batch.LastError,
		&batch.NextAttemptAt, &lockedAt, &completedAt,
		&batch.CreatedAt, &batch.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(recipientsRaw, &batch.Recipients); err != nil {
		return nil, fmt.Errorf("decode announcement email recipients: %w", err)
	}
	if lockedAt.Valid {
		batch.LockedAt = &lockedAt.Time
	}
	if completedAt.Valid {
		batch.CompletedAt = &completedAt.Time
	}
	return &batch, nil
}
