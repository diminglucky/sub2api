package downstream_test

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageLogRepositoryCreateCopiesDrawSubsiteFromContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := repository.NewUsageLogRepository(nil, db)
	createdAt := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	ctx := downstream.WithSubsite(context.Background(), &downstream.Subsite{
		ID:     7,
		Slug:   "draw",
		Domain: "draw.superai.sbs",
		Status: downstream.SubsiteStatusActive,
	})
	log := &service.UsageLog{
		UserID:    1,
		APIKeyID:  2,
		AccountID: 3,
		Model:     "gpt-5",
		CreatedAt: createdAt,
	}

	args := make([]driver.Value, 0, 63)
	for i := 0; i < 62; i++ {
		args = append(args, sqlmock.AnyArg())
	}
	args = append(args, int64(7))
	mock.ExpectQuery("INSERT INTO usage_logs").
		WithArgs(args...).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(99), createdAt))

	inserted, err := repo.Create(ctx, log)
	require.NoError(t, err)
	require.True(t, inserted)
	require.NotNil(t, log.SubsiteID)
	require.Equal(t, int64(7), *log.SubsiteID)
	require.NoError(t, mock.ExpectationsWereMet())
}
