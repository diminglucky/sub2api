package service

import (
	"context"
	"time"
)

const (
	AnnouncementEmailBatchStatusPending    = "pending"
	AnnouncementEmailBatchStatusProcessing = "processing"
	AnnouncementEmailBatchStatusRetrying   = "retrying"
	AnnouncementEmailBatchStatusCompleted  = "completed"
	AnnouncementEmailBatchStatusFailed     = "failed"
)

type AnnouncementEmailBatch struct {
	ID             int64      `json:"id"`
	AnnouncementID int64      `json:"announcement_id"`
	CampaignID     string     `json:"campaign_id"`
	Title          string     `json:"title"`
	Content        string     `json:"-"`
	Recipients     []string   `json:"-"`
	Status         string     `json:"status"`
	AttemptCount   int        `json:"attempt_count"`
	MaxAttempts    int        `json:"max_attempts"`
	TotalCount     int        `json:"total_count"`
	ProcessedCount int        `json:"processed_count"`
	FailedCount    int        `json:"failed_count"`
	LastError      string     `json:"last_error,omitempty"`
	NextAttemptAt  time.Time  `json:"next_attempt_at"`
	LockedAt       *time.Time `json:"locked_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type AnnouncementEmailBatchRepository interface {
	ClaimDue(ctx context.Context, staleAfter time.Duration) (*AnnouncementEmailBatch, error)
	RefreshLock(ctx context.Context, id int64) error
	MarkCompleted(ctx context.Context, id int64, processedCount int) error
	MarkRetry(ctx context.Context, id int64, processedCount, failedCount int, lastError string, nextAttemptAt time.Time) error
	ListByAnnouncement(ctx context.Context, announcementID int64) ([]AnnouncementEmailBatch, error)
}
