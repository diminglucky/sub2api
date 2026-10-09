//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/stretchr/testify/require"
)

// applyDownstreamPriceToResolved rewrites the main-site unit prices so the
// user-facing charge follows the sub-site selling price.
func TestApplyDownstreamPriceToResolvedOverridesTokenPrices(t *testing.T) {
	base := &ResolvedPricing{
		Mode:   BillingModeToken,
		Source: PricingSourceChannel,
		BasePricing: &ModelPricing{
			InputPricePerToken:  3e-6,
			OutputPricePerToken: 15e-6,
		},
	}
	subInput := 6e-6
	subOutput := 30e-6
	applied := applyDownstreamPriceToResolved(base, downstream.Price{
		InputPrice:  &subInput,
		OutputPrice: &subOutput,
	})

	require.NotSame(t, base, applied)
	require.InDelta(t, 6e-6, applied.BasePricing.InputPricePerToken, 1e-12)
	require.InDelta(t, 30e-6, applied.BasePricing.OutputPricePerToken, 1e-12)
	// The wholesale base must not be mutated.
	require.InDelta(t, 3e-6, base.BasePricing.InputPricePerToken, 1e-12)
	require.InDelta(t, 15e-6, base.BasePricing.OutputPricePerToken, 1e-12)
}

func TestApplyDownstreamPriceToResolvedOverridesIntervals(t *testing.T) {
	intervalInput := 1e-6
	base := &ResolvedPricing{
		Mode: BillingModeToken,
		BasePricing: &ModelPricing{
			InputPricePerToken: 3e-6,
		},
		Intervals: []PricingInterval{
			{InputPrice: &intervalInput},
		},
	}
	subInput := 4e-6
	applied := applyDownstreamPriceToResolved(base, downstream.Price{InputPrice: &subInput})

	require.NotNil(t, applied.Intervals[0].InputPrice)
	require.InDelta(t, 4e-6, *applied.Intervals[0].InputPrice, 1e-12)
	require.Nil(t, applied.Intervals[0].InputMultiplier)
	// Original interval untouched.
	require.InDelta(t, 1e-6, *base.Intervals[0].InputPrice, 1e-12)
}

func TestApplyDownstreamPriceToResolvedNoOpWhenPriceUnset(t *testing.T) {
	base := &ResolvedPricing{Mode: BillingModeToken}
	require.Same(t, base, applyDownstreamPriceToResolved(base, downstream.Price{}))
}

// The sub-site price is the user-facing charge (revenue); the main-site price
// is the wholesale cost used for the settlement margin.
func TestSubsitePriceChargesRevenueWhileMainPriceIsWholesaleCost(t *testing.T) {
	bs := newTestBillingService()
	tokens := UsageTokens{InputTokens: 1000, OutputTokens: 500}

	base := &ResolvedPricing{
		Mode:   BillingModeToken,
		Source: PricingSourceChannel,
		BasePricing: &ModelPricing{
			InputPricePerToken:  3e-6,
			OutputPricePerToken: 15e-6,
		},
	}
	subInput := 6e-6
	subOutput := 30e-6
	subsiteResolved := applyDownstreamPriceToResolved(base, downstream.Price{
		InputPrice:  &subInput,
		OutputPrice: &subOutput,
	})

	wholesale, err := bs.CalculateCostUnified(CostInput{
		Ctx: context.Background(), Model: "claude-sonnet-4",
		Tokens: tokens, RateMultiplier: 1, Resolved: base,
		Resolver: NewModelPricingResolver(nil, bs),
	})
	require.NoError(t, err)
	revenue, err := bs.CalculateCostUnified(CostInput{
		Ctx: context.Background(), Model: "claude-sonnet-4",
		Tokens: tokens, RateMultiplier: 1, Resolved: subsiteResolved,
		Resolver: NewModelPricingResolver(nil, bs),
	})
	require.NoError(t, err)

	require.InDelta(t, 1000*3e-6+500*15e-6, wholesale.ActualCost, 1e-12)
	require.InDelta(t, 1000*6e-6+500*30e-6, revenue.ActualCost, 1e-12)
	require.Greater(t, revenue.ActualCost, wholesale.ActualCost)
	require.InDelta(t, revenue.ActualCost-wholesale.ActualCost,
		downstream.UsageSettlement{
			RevenueAmount: revenue.ActualCost,
			CostAmount:    wholesale.ActualCost,
		}.MarginAmount(), 1e-12)
}

func TestDownstreamPriceFromResolvedSnapshotsBasePricing(t *testing.T) {
	resolved := &ResolvedPricing{
		Mode: BillingModeToken,
		BasePricing: &ModelPricing{
			InputPricePerToken:  3e-6,
			OutputPricePerToken: 15e-6,
		},
	}
	price := downstreamPriceFromResolved(resolved)
	require.NotNil(t, price.InputPrice)
	require.InDelta(t, 3e-6, *price.InputPrice, 1e-12)
	require.InDelta(t, 15e-6, *price.OutputPrice, 1e-12)
	require.False(t, price.IsZero())
}
