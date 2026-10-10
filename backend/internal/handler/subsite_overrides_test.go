package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubsitePriceOverridesGroupAssigned(t *testing.T) {
	// No restriction when the assigned set is nil.
	require.True(t, (&SubsitePriceOverrides{}).GroupAssigned(7))
	require.True(t, (*SubsitePriceOverrides)(nil).GroupAssigned(7))

	restricted := &SubsitePriceOverrides{AssignedGroups: map[int64]struct{}{3: {}}}
	require.True(t, restricted.GroupAssigned(3))
	require.False(t, restricted.GroupAssigned(4))
}

func TestScaleUserSupportedModelPricingScalesIntervals(t *testing.T) {
	input := 1.0
	output := 2.0
	intervalInput := 0.5
	pricing := &userSupportedModelPricing{
		InputPrice:  &input,
		OutputPrice: &output,
		Intervals: []userPricingIntervalDTO{
			{MinTokens: 1000, InputPrice: &intervalInput},
		},
	}

	scaleUserSupportedModelPricing(pricing, 1.25)

	require.InDelta(t, 1.25, *pricing.InputPrice, 1e-12)
	require.InDelta(t, 2.5, *pricing.OutputPrice, 1e-12)
	require.InDelta(t, 0.625, *pricing.Intervals[0].InputPrice, 1e-12)
}
