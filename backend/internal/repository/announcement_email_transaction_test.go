package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAnnouncementCreateWithEmailBatchRollsBackTogether(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO announcements").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(42), now, now))
	mock.ExpectQuery("INSERT INTO announcement_email_batches").
		WillReturnError(errors.New("batch insert failed"))
	mock.ExpectRollback()

	repo := &announcementRepository{db: db}
	announcement := &service.Announcement{
		Title:      "New offer",
		Content:    "Offer details",
		Status:     service.AnnouncementStatusActive,
		NotifyMode: service.AnnouncementNotifyModePopup,
	}
	batch := &service.AnnouncementEmailBatch{
		CampaignID:  "d1374e45-3c42-4d82-b4cd-754fc0a0ecbf",
		Title:       announcement.Title,
		Content:     announcement.Content,
		Recipients:  []string{"user@example.com"},
		MaxAttempts: 5,
	}

	err = repo.CreateWithEmailBatch(context.Background(), announcement, batch)

	require.EqualError(t, err, "batch insert failed")
	require.NoError(t, mock.ExpectationsWereMet())
}
