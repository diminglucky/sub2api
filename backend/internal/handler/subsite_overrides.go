package handler

import "context"

// SubsitePriceOverrides carries a sub-site's effective pricing overrides so
// user-facing endpoints (group list, model plaza) can show the sub-site price
// instead of the main-site price.
type SubsitePriceOverrides struct {
	// GroupRates maps group_id -> sub-site group multiplier.
	GroupRates map[int64]float64
	// ModelRates maps model -> sub-site model multiplier.
	ModelRates map[string]float64
}

// SubsitePriceOverrideProvider loads the overrides for one sub-site.
type SubsitePriceOverrideProvider func(ctx context.Context, subsiteID int64) (*SubsitePriceOverrides, error)

var subsitePriceOverrideProvider SubsitePriceOverrideProvider

// SetSubsitePriceOverrideProvider wires the sub-site override provider.
func SetSubsitePriceOverrideProvider(provider SubsitePriceOverrideProvider) {
	subsitePriceOverrideProvider = provider
}

// LoadSubsitePriceOverrides returns the overrides for a sub-site, or nil when
// no provider is wired or no override exists.
func LoadSubsitePriceOverrides(ctx context.Context, subsiteID int64) *SubsitePriceOverrides {
	if subsitePriceOverrideProvider == nil || subsiteID <= 0 {
		return nil
	}
	overrides, err := subsitePriceOverrideProvider(ctx, subsiteID)
	if err != nil {
		return nil
	}
	return overrides
}
