package downstream

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

// TestRepositoryGetSubsiteBySlug covers the pilot subsite draw.superai.sbs.
// The slug is normalized (trim + lowercase) before the lookup so host-derived
// slugs such as "DRAW " resolve to the same row.
func TestRepositoryGetSubsiteBySlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	createdAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	mock.ExpectQuery(`(?s)FROM subsites\s+WHERE slug = \$1`).
		WithArgs("draw").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "slug", "domain", "name", "logo_url", "theme_color",
			"status", "admin_user_id", "created_at", "updated_at",
		}).AddRow(
			int64(7), "draw", "draw.superai.sbs", "Draw", "", "#0ea5e9",
			SubsiteStatusActive, int64(42), createdAt, updatedAt,
		))

	repo := NewRepository(db)
	subsite, err := repo.GetSubsiteBySlug(context.Background(), "  DRAW  ")
	require.NoError(t, err)
	require.NotNil(t, subsite)
	require.Equal(t, int64(7), subsite.ID)
	require.Equal(t, "draw", subsite.Slug)
	require.Equal(t, "draw.superai.sbs", subsite.Domain)
	require.Equal(t, "Draw", subsite.Name)
	require.True(t, subsite.IsActive())
	require.NotNil(t, subsite.AdminUserID)
	require.Equal(t, int64(42), *subsite.AdminUserID)
	require.Equal(t, createdAt, subsite.CreatedAt)
	require.Equal(t, updatedAt, subsite.UpdatedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRepositoryGetSubsiteBySlugAllowsNullAdmin ensures a subsite without a
// bound administrator is still returned; main-site data keeps subsite_id NULL.
func TestRepositoryGetSubsiteBySlugAllowsNullAdmin(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	createdAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`(?s)FROM subsites\s+WHERE slug = \$1`).
		WithArgs("draw").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "slug", "domain", "name", "logo_url", "theme_color",
			"status", "admin_user_id", "created_at", "updated_at",
		}).AddRow(
			int64(7), "draw", "draw.superai.sbs", "Draw", "", "#0ea5e9",
			SubsiteStatusDisabled, nil, createdAt, createdAt,
		))

	repo := NewRepository(db)
	subsite, err := repo.GetSubsiteBySlug(context.Background(), "draw")
	require.NoError(t, err)
	require.Nil(t, subsite.AdminUserID)
	require.False(t, subsite.IsActive())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryGetSubsiteBySlugReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`(?s)FROM subsites\s+WHERE slug = \$1`).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	repo := NewRepository(db)
	subsite, err := repo.GetSubsiteBySlug(context.Background(), "missing")
	require.ErrorIs(t, err, ErrSubsiteNotFound)
	require.Nil(t, subsite)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryGetSubsiteBySlugRejectsEmptySlug(t *testing.T) {
	repo := NewRepository(nil)
	subsite, err := repo.GetSubsiteBySlug(context.Background(), "   ")
	require.ErrorIs(t, err, ErrSubsiteNotFound)
	require.Nil(t, subsite)
}

func TestRepositoryGetSubsiteBySlugRejectsNilReceiver(t *testing.T) {
	var repo *Repository
	_, err := repo.GetSubsiteBySlug(context.Background(), "draw")
	require.Error(t, err)
}

// TestDownstreamSubsiteMigration pins the schema contract Task 1 must provide:
// the four subsite tables plus nullable subsite_id columns on the shared tables.
func TestDownstreamSubsiteMigration(t *testing.T) {
	content, err := migrations.FS.ReadFile("242_downstream_subsites.sql")
	require.NoError(t, err)
	sqlText := string(content)

	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS subsites",
		"CREATE TABLE IF NOT EXISTS subsite_members",
		"CREATE TABLE IF NOT EXISTS subsite_prices",
		"CREATE TABLE IF NOT EXISTS settlement_ledger",
		"ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS subsite_id BIGINT",
		"ALTER TABLE usage_logs     ADD COLUMN IF NOT EXISTS subsite_id BIGINT",
		"ALTER TABLE api_keys       ADD COLUMN IF NOT EXISTS subsite_id BIGINT",
		"REFERENCES subsites(id) ON DELETE CASCADE",
	} {
		require.Contains(t, sqlText, required)
	}

	// V1 must not pay out automatically: no payout/transfer statements.
	require.NotContains(t, sqlText, "TRANSFER")
	require.NotContains(t, sqlText, "payout")

	// This file runs inside a transaction, so it must not use CONCURRENTLY.
	// usage_logs indexes belong in a separate _notx.sql migration.
	require.NotContains(t, strings.ToUpper(stripSQLLineComments(sqlText)), "CONCURRENTLY")
	require.NotContains(t, sqlText, "idx_usage_logs_subsite_id")
}

// stripSQLLineComments removes `--` comment lines so structural assertions do
// not match explanatory text.
func stripSQLLineComments(sqlText string) string {
	var builder strings.Builder
	for _, line := range strings.Split(sqlText, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	return builder.String()
}
