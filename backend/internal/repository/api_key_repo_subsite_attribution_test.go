package repository

import (
	"context"
	"database/sql"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

// newAPIKeyRepoSubsiteSQLite builds an ent-backed sqlite repo and adds the
// migration-only subsite_id column so attribution SQL can execute. The ent
// schema intentionally omits subsite_id to keep the pilot change isolated.
func newAPIKeyRepoSubsiteSQLite(t *testing.T, name string) (*apiKeyRepository, *dbent.Client, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite", "file:"+name+"?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	_, err = db.Exec("ALTER TABLE api_keys ADD COLUMN subsite_id INTEGER")
	require.NoError(t, err)

	return &apiKeyRepository{client: client, sql: db}, client, db
}

func TestAPIKeyRepositoryCreateStampsSubsiteAtomically(t *testing.T) {
	repo, client, db := newAPIKeyRepoSubsiteSQLite(t, "api_key_repo_subsite_create")
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "subsite-create@test.com")

	subsiteID := int64(7)
	key := &service.APIKey{
		UserID:    user.ID,
		Key:       "sk-draw-subsite-create",
		Name:      "draw",
		Status:    service.StatusActive,
		SubsiteID: &subsiteID,
	}

	require.NoError(t, repo.Create(ctx, key))
	require.Greater(t, key.ID, int64(0))

	var stored sql.NullInt64
	require.NoError(t, db.QueryRowContext(ctx, "SELECT subsite_id FROM api_keys WHERE id = ?", key.ID).Scan(&stored))
	require.True(t, stored.Valid, "subsite_id must be persisted for a draw-created key")
	require.Equal(t, int64(7), stored.Int64)
}

func TestAPIKeyRepositoryCreateRollsBackWhenSubsiteAttributionFails(t *testing.T) {
	repo, client, db := newAPIKeyRepoSubsiteSQLite(t, "api_key_repo_subsite_rollback")
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "subsite-rollback@test.com")

	// Force the attribution UPDATE to fail so the atomicity guarantee is
	// exercised: the just-created key must not survive as an orphan.
	_, err := db.Exec(`CREATE TRIGGER block_subsite_update BEFORE UPDATE OF subsite_id ON api_keys BEGIN SELECT RAISE(ABORT, 'blocked by test'); END;`)
	require.NoError(t, err)

	subsiteID := int64(7)
	key := &service.APIKey{
		UserID:    user.ID,
		Key:       "sk-draw-subsite-rollback",
		Name:      "draw",
		Status:    service.StatusActive,
		SubsiteID: &subsiteID,
	}

	err = repo.Create(ctx, key)
	require.Error(t, err)

	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM api_keys WHERE key = ?", "sk-draw-subsite-rollback").Scan(&count))
	require.Equal(t, 0, count, "failed subsite attribution must roll back the created key")
}
