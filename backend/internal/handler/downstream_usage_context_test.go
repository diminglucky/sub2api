package handler

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func drawUsageRecordSubsiteContext() context.Context {
	return downstream.WithSubsite(context.Background(), &downstream.Subsite{
		ID:     7,
		Slug:   "draw",
		Domain: "draw.superai.sbs",
		Status: downstream.SubsiteStatusActive,
	})
}

// TestUsageRecordContextCarriesDownstreamSubsite pins the P1 fix: the billing
// task context must keep the resolved subsite when it is rebuilt from the
// worker-pool base context, otherwise real draw.superai.sbs traffic records
// usage_logs.subsite_id = NULL.
func TestUsageRecordContextCarriesDownstreamSubsite(t *testing.T) {
	parent := drawUsageRecordSubsiteContext()

	got := usageRecordContext(parent, context.Background())
	subsite, ok := downstream.FromContext(got)
	require.True(t, ok, "rebuilt task context must inherit the resolved subsite")
	require.NotNil(t, subsite)
	require.Equal(t, int64(7), subsite.ID)

	subsiteID, ok := downstream.SubsiteIDFromContext(got)
	require.True(t, ok)
	require.Equal(t, int64(7), subsiteID)
}

// TestUsageRecordContextKeepsMainSiteNull proves the no-subsite path is
// unchanged: a plain request context still yields no subsite.
func TestUsageRecordContextKeepsMainSiteNull(t *testing.T) {
	got := usageRecordContext(context.Background(), context.Background())
	_, ok := downstream.SubsiteIDFromContext(got)
	require.False(t, ok, "main-site task context must not invent a subsite")
}

// TestWrapUsageRecordTaskContextPreservesSubsite exercises the exact seam the
// worker pool and the sync fallback both go through.
func TestWrapUsageRecordTaskContextPreservesSubsite(t *testing.T) {
	parent := drawUsageRecordSubsiteContext()

	var observed int64
	task := service.UsageRecordTask(func(ctx context.Context) {
		if subsite, ok := downstream.FromContext(ctx); ok && subsite != nil {
			observed = subsite.ID
		}
	})

	wrapped, abandon := wrapUsageRecordTaskContext(parent, task)
	defer abandon()
	// The worker pool calls the wrapped task with its own detached base
	// context; the subsite must survive that swap.
	wrapped(context.Background())

	require.Equal(t, int64(7), observed, "wrapped usage task must observe the subsite")
}
