package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type announcementEmailBatchRepoStub struct {
	completedID    int64
	processedCount int
	retryID        int64
	failedCount    int
	lastError      string
}

func (*announcementEmailBatchRepoStub) ClaimDue(context.Context, time.Duration) (*AnnouncementEmailBatch, error) {
	return nil, nil
}

func (*announcementEmailBatchRepoStub) RefreshLock(context.Context, int64) error { return nil }

func (s *announcementEmailBatchRepoStub) MarkCompleted(_ context.Context, id int64, processedCount int) error {
	s.completedID = id
	s.processedCount = processedCount
	return nil
}

func (s *announcementEmailBatchRepoStub) MarkRetry(_ context.Context, id int64, processedCount, failedCount int, lastError string, _ time.Time) error {
	s.retryID = id
	s.processedCount = processedCount
	s.failedCount = failedCount
	s.lastError = lastError
	return nil
}

func (*announcementEmailBatchRepoStub) ListByAnnouncement(context.Context, int64) ([]AnnouncementEmailBatch, error) {
	return nil, nil
}

func TestPersistentAnnouncementBatchRecordsRetryAfterRecipientFailure(t *testing.T) {
	settings := newNotificationEmailMemorySettingRepo()
	settings.values[SettingKeySMTPHost] = "smtp.example.com"
	settings.values[SettingKeySMTPFrom] = "sender@example.com"
	emailSvc := NewEmailService(settings, nil)
	emailSvc.sendWithConfig = func(_ *SMTPConfig, to, _, _ string) error {
		if to == "failed@example.com" {
			return errors.New("smtp rejected recipient")
		}
		return nil
	}
	NewNotificationEmailService(settings, emailSvc)
	repo := &announcementEmailBatchRepoStub{}
	queue := &EmailQueueService{emailService: emailSvc, announcementBatchRepo: repo}

	queue.processPersistentAnnouncementBatch(&AnnouncementEmailBatch{
		ID:             9,
		AnnouncementID: 42,
		CampaignID:     "campaign-1",
		Title:          "New offer",
		Content:        "Offer details",
		Recipients:     []string{"ok@example.com", "failed@example.com"},
		AttemptCount:   1,
	})

	require.Equal(t, int64(9), repo.retryID)
	require.Equal(t, 1, repo.processedCount)
	require.Equal(t, 1, repo.failedCount)
	require.Contains(t, repo.lastError, "smtp rejected recipient")
	require.Zero(t, repo.completedID)
}

func TestPersistentAnnouncementBatchMarksCompleted(t *testing.T) {
	settings := newNotificationEmailMemorySettingRepo()
	settings.values[SettingKeySMTPHost] = "smtp.example.com"
	settings.values[SettingKeySMTPFrom] = "sender@example.com"
	emailSvc := NewEmailService(settings, nil)
	emailSvc.sendWithConfig = func(_ *SMTPConfig, _, _, _ string) error { return nil }
	NewNotificationEmailService(settings, emailSvc)
	repo := &announcementEmailBatchRepoStub{}
	queue := &EmailQueueService{emailService: emailSvc, announcementBatchRepo: repo}

	queue.processPersistentAnnouncementBatch(&AnnouncementEmailBatch{
		ID:             10,
		AnnouncementID: 43,
		CampaignID:     "campaign-2",
		Title:          "New offer",
		Content:        "Offer details",
		Recipients:     []string{"first@example.com", "second@example.com"},
	})

	require.Equal(t, int64(10), repo.completedID)
	require.Equal(t, 2, repo.processedCount)
	require.Zero(t, repo.retryID)
}
