package downstream

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type attributionExecRecorder struct {
	mu    sync.Mutex
	query string
	args  []any
}

func (r *attributionExecRecorder) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.query = query
	r.args = append([]any(nil), args...)
	return sqlmockResult{}, nil
}

type sqlmockResult struct{}

func (sqlmockResult) LastInsertId() (int64, error) { return 0, nil }
func (sqlmockResult) RowsAffected() (int64, error) { return 1, nil }

func drawAttributionContext() context.Context {
	return WithSubsite(context.Background(), &Subsite{
		ID:     7,
		Slug:   "draw",
		Domain: "draw.superai.sbs",
		Status: SubsiteStatusActive,
	})
}

func TestSubsiteAttributionRegistrationLinksDrawMember(t *testing.T) {
	ctx := drawAttributionContext()
	subsiteID, ok := SubsiteIDFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, int64(7), subsiteID)

	exec := &attributionExecRecorder{}
	err := RecordSubsiteMember(ctx, exec, subsiteID, 42, SubsiteMemberSourceRegistration)
	require.NoError(t, err)
	require.Contains(t, strings.ToUpper(exec.query), "INSERT INTO SUBSITE_MEMBERS")
	require.Equal(t, []any{int64(7), int64(42), SubsiteMemberRoleMember, SubsiteMemberSourceRegistration}, exec.args)
}

func TestSubsiteAttributionOrderCopiesDrawSubsite(t *testing.T) {
	ctx := drawAttributionContext()
	subsiteID, ok := SubsiteIDFromContext(ctx)
	require.True(t, ok)

	exec := &attributionExecRecorder{}
	err := AttributePaymentOrder(ctx, exec, subsiteID, 295)
	require.NoError(t, err)
	require.Contains(t, strings.ToUpper(exec.query), "UPDATE PAYMENT_ORDERS SET SUBSITE_ID")
	require.Equal(t, []any{int64(7), int64(295)}, exec.args)
}

func TestSubsiteAttributionAPIKeyCopiesDrawSubsite(t *testing.T) {
	ctx := drawAttributionContext()
	subsiteID, ok := SubsiteIDFromContext(ctx)
	require.True(t, ok)

	exec := &attributionExecRecorder{}
	err := AttributeAPIKey(ctx, exec, subsiteID, 88)
	require.NoError(t, err)
	require.Contains(t, strings.ToUpper(exec.query), "UPDATE API_KEYS SET SUBSITE_ID")
	require.Equal(t, []any{int64(7), int64(88)}, exec.args)
}

func TestSubsiteAttributionUsageLogKeepsMainSiteNullWithoutSubsiteContext(t *testing.T) {
	_, ok := SubsiteIDFromContext(context.Background())
	require.False(t, ok)
}
