package service

import (
	"context"
	"database/sql"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

// newAuthServiceSubsiteSQLite builds an AuthService wired to an sqlite ent
// client plus the migration-only subsite_members table. The table lives outside
// the ent schema on purpose so the pilot change stays additive.
func newAuthServiceSubsiteSQLite(t *testing.T, name string) (*AuthService, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite", "file:"+name+"?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	_, err = db.Exec(`
		CREATE TABLE subsite_members (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			subsite_id INTEGER NOT NULL,
			user_id    INTEGER NOT NULL,
			role       TEXT    NOT NULL DEFAULT 'member',
			source     TEXT    NOT NULL DEFAULT 'registration',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (subsite_id, user_id)
		)
	`)
	require.NoError(t, err)

	return &AuthService{entClient: client}, db
}

func drawRegistrationSubsiteContext() context.Context {
	return downstream.WithSubsite(context.Background(), &downstream.Subsite{
		ID:     7,
		Slug:   "draw",
		Domain: "draw.superai.sbs",
		Status: downstream.SubsiteStatusActive,
	})
}

func countSubsiteMembers(t *testing.T, db *sql.DB, subsiteID, userID int64) int {
	t.Helper()
	var count int
	require.NoError(t, db.QueryRow(
		"SELECT COUNT(*) FROM subsite_members WHERE subsite_id = ? AND user_id = ?",
		subsiteID, userID,
	).Scan(&count))
	return count
}

func TestAuthServiceAttributeDownstreamRegistrationWritesMember(t *testing.T) {
	svc, db := newAuthServiceSubsiteSQLite(t, "auth_subsite_attribute")

	require.NoError(t, svc.attributeDownstreamRegistration(drawRegistrationSubsiteContext(), 42))
	require.Equal(t, 1, countSubsiteMembers(t, db, 7, 42))
}

func TestAuthServiceAttributeDownstreamRegistrationSkipsMainSite(t *testing.T) {
	svc, db := newAuthServiceSubsiteSQLite(t, "auth_subsite_main_site")

	require.NoError(t, svc.attributeDownstreamRegistration(context.Background(), 42))
	require.Equal(t, 0, countSubsiteMembers(t, db, 7, 42))
}

// TestAuthServiceDownstreamMembershipSharesRegistrationTransaction proves the
// no-invitation registration path is atomic: a rolled-back transaction leaves
// neither the user nor its subsite membership behind.
func TestAuthServiceDownstreamMembershipSharesRegistrationTransaction(t *testing.T) {
	svc, db := newAuthServiceSubsiteSQLite(t, "auth_subsite_tx")
	ctx := drawRegistrationSubsiteContext()

	tx, err := svc.entClient.Tx(ctx)
	require.NoError(t, err)
	txCtx := dbent.NewTxContext(ctx, tx)

	require.NoError(t, svc.attributeDownstreamRegistration(txCtx, 42))
	// Read through the transaction client: the shared-cache sqlite connection
	// would block a separate connection while this write tx is open.
	var inTx int
	row, err := tx.Client().QueryContext(txCtx, "SELECT COUNT(*) FROM subsite_members WHERE subsite_id = ? AND user_id = ?", 7, 42)
	require.NoError(t, err)
	require.True(t, row.Next())
	require.NoError(t, row.Scan(&inTx))
	require.NoError(t, row.Close())
	require.Equal(t, 1, inTx, "membership must be visible inside the transaction")

	require.NoError(t, tx.Rollback())
	require.Equal(t, 0, countSubsiteMembers(t, db, 7, 42), "rolled-back registration must not keep membership")
}

func TestAuthServiceRemoveDownstreamRegistrationAttribution(t *testing.T) {
	svc, db := newAuthServiceSubsiteSQLite(t, "auth_subsite_remove")

	require.NoError(t, svc.attributeDownstreamRegistration(drawRegistrationSubsiteContext(), 42))
	require.Equal(t, 1, countSubsiteMembers(t, db, 7, 42))

	require.NoError(t, svc.removeDownstreamRegistrationAttribution(drawRegistrationSubsiteContext(), 42))
	require.Equal(t, 0, countSubsiteMembers(t, db, 7, 42))
}

func TestAuthServiceRemoveDownstreamRegistrationAttributionSkipsMainSite(t *testing.T) {
	svc, db := newAuthServiceSubsiteSQLite(t, "auth_subsite_remove_main")

	require.NoError(t, svc.attributeDownstreamRegistration(drawRegistrationSubsiteContext(), 42))
	require.Equal(t, 1, countSubsiteMembers(t, db, 7, 42))

	require.NoError(t, svc.removeDownstreamRegistrationAttribution(context.Background(), 42))
	require.Equal(t, 1, countSubsiteMembers(t, db, 7, 42), "main-site rollback must not touch subsite memberships")
}
