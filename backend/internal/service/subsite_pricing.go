package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// DownstreamPricingRepository resolves sub-site price overrides. It is
// satisfied by *downstream.Repository.
//
// The main-site price passed as basePrice is the wholesale cost; the returned
// price is what the sub-site charges its users. NULL subsite rows (main site)
// are never resolved here.
type DownstreamPricingRepository interface {
	ResolveSubsitePrice(ctx context.Context, subsiteID int64, model string, basePrice downstream.Price, groupID ...int64) (downstream.Price, error)
	ResolveSubsiteOverride(ctx context.Context, subsiteID int64, model string, groupID ...int64) (*downstream.PriceOverride, error)
}

// DownstreamUsageSettlementRecorder records one usage-based settlement ledger
// entry. It is satisfied by *downstream.Repository.
type DownstreamUsageSettlementRecorder interface {
	RecordUsageSettlement(ctx context.Context, entry downstream.UsageSettlement) error
}

// SetDownstreamPricing wires the optional sub-site price resolver and usage
// settlement recorder. When unset (default, and for every main-site request)
// billing behaves exactly as before.
func (s *OpenAIGatewayService) SetDownstreamPricing(
	pricing DownstreamPricingRepository,
	settlement DownstreamUsageSettlementRecorder,
) {
	if s == nil {
		return
	}
	s.downstreamPricing = pricing
	s.downstreamSettlement = settlement
}

// subsitePricingResolution is the resolved sub-site price for one request.
type subsitePricingResolution struct {
	SubsiteID int64
	Price     downstream.Price
	Applied   bool
}

// resolveSubsitePricing resolves the sub-site price override for the current
// request. It returns ok=false for main-site traffic, when no resolver is
// wired, or when the resolved price is identical to the main-site price.
func (s *OpenAIGatewayService) resolveSubsitePricing(
	ctx context.Context,
	model string,
	apiKey *APIKey,
	base *ResolvedPricing,
) (subsitePricingResolution, bool) {
	if s == nil || s.downstreamPricing == nil {
		return subsitePricingResolution{}, false
	}
	subsiteID, ok := downstream.SubsiteIDFromContext(ctx)
	if !ok || subsiteID <= 0 {
		return subsitePricingResolution{}, false
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return subsitePricingResolution{}, false
	}
	basePrice := downstreamPriceFromResolved(base)
	if basePrice.IsZero() {
		return subsitePricingResolution{}, false
	}
	var groupID []int64
	if apiKey != nil && apiKey.GroupID != nil && *apiKey.GroupID > 0 {
		groupID = []int64{*apiKey.GroupID}
	}
	price, err := s.downstreamPricing.ResolveSubsitePrice(ctx, subsiteID, model, basePrice, groupID...)
	if err != nil {
		return subsitePricingResolution{}, false
	}
	if price.Equal(basePrice) {
		return subsitePricingResolution{SubsiteID: subsiteID, Price: price}, false
	}
	return subsitePricingResolution{SubsiteID: subsiteID, Price: price, Applied: true}, true
}

// resolveSubsitePricingForModel resolves the main-site base pricing for a model
// and then applies any sub-site override. It returns ok=false for main-site
// traffic or when no override changes the price.
func (s *OpenAIGatewayService) resolveSubsitePricingForModel(
	ctx context.Context,
	model string,
	apiKey *APIKey,
) (subsitePricingResolution, bool) {
	if s == nil || s.downstreamPricing == nil || s.resolver == nil ||
		apiKey == nil || apiKey.Group == nil {
		return subsitePricingResolution{}, false
	}
	if _, ok := downstream.SubsiteIDFromContext(ctx); !ok {
		return subsitePricingResolution{}, false
	}
	gid := apiKey.Group.ID
	base := s.resolver.Resolve(ctx, PricingInput{Model: model, GroupID: &gid, Group: apiKey.Group})
	return s.resolveSubsitePricing(ctx, model, apiKey, base)
}

// resolveDownstreamMediaPricing resolves a sub-site per-request override for
// image/video requests. Media pricing normally comes from the group media price
// fields (ImagePrice1K/2K/4K, VideoPrice*, per-model video prices, or the code
// defaults), which are not part of the resolved token base price. Gating media
// overrides on the token base being non-zero therefore drops the override
// entirely whenever the media model has no token price — the user is charged the
// main-site media price and no settlement row is written.
//
// The base price here is the main-site per-unit media price for the same
// dimensions/duration, so a rate multiplier scales the wholesale media price
// while an explicit per-request price wins outright. It returns ok=true only
// when the per-request component actually changed, so a token-only override on
// a media request stays with the token resolution path.
func (s *OpenAIGatewayService) resolveDownstreamMediaPricing(
	ctx context.Context,
	model string,
	apiKey *APIKey,
	result *OpenAIForwardResult,
) (subsitePricingResolution, bool) {
	if s == nil || s.downstreamPricing == nil || apiKey == nil || apiKey.Group == nil {
		return subsitePricingResolution{}, false
	}
	if result == nil || (result.ImageCount <= 0 && result.VideoCount <= 0) {
		return subsitePricingResolution{}, false
	}
	subsiteID, ok := downstream.SubsiteIDFromContext(ctx)
	if !ok || subsiteID <= 0 {
		return subsitePricingResolution{}, false
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return subsitePricingResolution{}, false
	}
	// When the model is channel/group-priced by token the media cost path is not
	// used at all; keep the token resolution so a token override is not lost.
	if resolved := s.resolveOpenAIChannelPricing(ctx, model, apiKey); resolved != nil && resolved.Mode == BillingModeToken {
		return subsitePricingResolution{}, false
	}
	base := s.downstreamMediaBasePrice(ctx, model, apiKey, result)
	var groupID []int64
	if apiKey.GroupID != nil && *apiKey.GroupID > 0 {
		groupID = []int64{*apiKey.GroupID}
	}
	price, err := s.downstreamPricing.ResolveSubsitePrice(ctx, subsiteID, model, base, groupID...)
	if err != nil {
		return subsitePricingResolution{}, false
	}
	changed := downstreamPriceFieldChanged(price.PerRequestPrice, base.PerRequestPrice) ||
		downstreamPriceFieldChanged(price.ImagePrice1K, base.ImagePrice1K) ||
		downstreamPriceFieldChanged(price.ImagePrice2K, base.ImagePrice2K) ||
		downstreamPriceFieldChanged(price.ImagePrice4K, base.ImagePrice4K)
	if !changed {
		return subsitePricingResolution{}, false
	}
	return subsitePricingResolution{SubsiteID: subsiteID, Price: price, Applied: true}, true
}

// downstreamMediaBasePrice snapshots the main-site per-unit media price for one
// image or one video (video unit = per-second price × duration). It runs the
// existing main-site media cost path with multiplier 1 and a single unit, so
// group media prices, per-model video prices and code defaults all resolve the
// same way they do during wholesale settlement.
func (s *OpenAIGatewayService) downstreamMediaBasePrice(
	ctx context.Context,
	model string,
	apiKey *APIKey,
	result *OpenAIForwardResult,
) downstream.Price {
	if s == nil || s.billingService == nil || apiKey == nil || apiKey.Group == nil || result == nil {
		return downstream.Price{}
	}
	switch {
	case result.VideoCount > 0:
		single := *result
		single.VideoCount = 1
		if cost := s.calculateOpenAIVideoCost(ctx, model, apiKey, &single, 1, nil); cost != nil {
			unit := cost.ActualCost
			return downstream.Price{PerRequestPrice: &unit}
		}
	case result.ImageCount > 0:
		single := *result
		single.ImageCount = 1
		if cost := s.calculateOpenAIImageCost(ctx, model, apiKey, &single, 1, nil); cost != nil {
			unit := cost.ActualCost
			return downstream.Price{PerRequestPrice: &unit}
		}
	}
	return downstream.Price{}
}

// downstreamPriceFieldChanged reports whether an override produced a different
// value than the supplied base. A nil override field is not a change; a nil base
// with a non-nil override is.
func downstreamPriceFieldChanged(value, base *float64) bool {
	if value == nil {
		return false
	}
	if base == nil {
		return true
	}
	return math.Abs(*value-*base) > 1e-12
}

// recordDownstreamUsageSettlement snapshots the main-site wholesale cost for a
// billed sub-site request and writes one usage-based ledger row. It recomputes
// the cost without the sub-site override so revenue (sub-site price) and cost
// (main-site price) share the exact same usage, model and multiplier context.
func (s *OpenAIGatewayService) recordDownstreamUsageSettlement(
	ctx context.Context,
	requestID string,
	userID int64,
	apiKey *APIKey,
	result *OpenAIForwardResult,
	billingModels []string,
	multiplier, imageMultiplier, videoMultiplier, webSearchMultiplier float64,
	tokens UsageTokens,
	serviceTier string,
	longContextBillingGate *bool,
	pricingAt time.Time,
	revenue *CostBreakdown,
	subsitePricing subsitePricingResolution,
) {
	if s == nil || s.downstreamSettlement == nil || revenue == nil {
		return
	}
	subsiteID := subsitePricing.SubsiteID
	requestID = strings.TrimSpace(requestID)
	if subsiteID <= 0 || requestID == "" {
		return
	}
	wholesaleCost, err := s.calculateOpenAIRecordUsageCostWithSubsite(
		ctx,
		result,
		apiKey,
		billingModels,
		multiplier,
		imageMultiplier,
		videoMultiplier,
		webSearchMultiplier,
		tokens,
		serviceTier,
		longContextBillingGate,
		pricingAt,
		nil,
	)
	if err != nil || wholesaleCost == nil {
		return
	}
	entry := downstream.UsageSettlement{
		SubsiteID:      subsiteID,
		UserID:         userID,
		UsageRequestID: requestID,
		BillingMode:    strings.TrimSpace(revenue.BillingMode),
		Currency:       "USD",
		RevenueAmount:  revenue.ActualCost,
		CostAmount:     wholesaleCost.ActualCost,
	}
	// Best-effort: a ledger write must never fail the billing path. The ledger
	// insert is idempotent on (subsite_id, usage_request_id), so a retry driven
	// by a later duplicate callback is safe.
	if err := s.downstreamSettlement.RecordUsageSettlement(ctx, entry); err != nil {
		logger.LegacyPrintf(
			"service.openai_gateway",
			"downstream settlement record failed: subsite_id=%d request_id=%s err=%v",
			subsiteID, requestID, err,
		)
	}
}

// downstreamPriceFromResolved snapshots the main-site unit prices that a
// sub-site override is applied on top of.
func downstreamPriceFromResolved(resolved *ResolvedPricing) downstream.Price {
	if resolved == nil {
		return downstream.Price{}
	}
	var price downstream.Price
	if resolved.BasePricing != nil {
		input := resolved.BasePricing.InputPricePerToken
		inputPriority := resolved.BasePricing.InputPricePerTokenPriority
		output := resolved.BasePricing.OutputPricePerToken
		outputPriority := resolved.BasePricing.OutputPricePerTokenPriority
		cacheWrite := resolved.BasePricing.CacheCreationPricePerToken
		cacheWritePriority := resolved.BasePricing.CacheCreationPricePerTokenPriority
		cacheRead := resolved.BasePricing.CacheReadPricePerToken
		cacheReadPriority := resolved.BasePricing.CacheReadPricePerTokenPriority
		price.InputPrice = &input
		if inputPriority > 0 {
			price.InputPricePriority = &inputPriority
		}
		price.OutputPrice = &output
		if outputPriority > 0 {
			price.OutputPricePriority = &outputPriority
		}
		if resolved.BasePricing.CacheCreationPriceExplicit || cacheWrite > 0 {
			price.CacheWritePrice = &cacheWrite
			if cacheWritePriority > 0 {
				price.CacheWritePricePriority = &cacheWritePriority
			}
		}
		price.CacheReadPrice = &cacheRead
		if cacheReadPriority > 0 {
			price.CacheReadPricePriority = &cacheReadPriority
		}
	}
	if resolved.DefaultPerRequestPrice > 0 {
		perRequest := resolved.DefaultPerRequestPrice
		price.PerRequestPrice = &perRequest
	}
	if price.IsZero() && len(resolved.Intervals) > 0 {
		if iv := FindMatchingInterval(resolved.Intervals, 0); iv != nil {
			price.InputPrice = iv.InputPrice
			price.OutputPrice = iv.OutputPrice
			price.CacheWritePrice = iv.CacheWritePrice
			price.CacheReadPrice = iv.CacheReadPrice
			price.PerRequestPrice = iv.PerRequestPrice
		}
	}
	return price
}

// applyDownstreamPriceToResolved returns a copy of resolved with the sub-site
// absolute prices applied. The main-site (wholesale) pricing is never mutated.
func applyDownstreamPriceToResolved(resolved *ResolvedPricing, price downstream.Price) *ResolvedPricing {
	if resolved == nil || price.IsZero() {
		return resolved
	}
	out := *resolved
	if resolved.BasePricing != nil {
		cloned := *resolved.BasePricing
		if price.InputPrice != nil {
			cloned.InputPricePerToken = *price.InputPrice
		}
		if price.InputPricePriority != nil {
			cloned.InputPricePerTokenPriority = *price.InputPricePriority
		}
		if price.OutputPrice != nil {
			cloned.OutputPricePerToken = *price.OutputPrice
		}
		if price.OutputPricePriority != nil {
			cloned.OutputPricePerTokenPriority = *price.OutputPricePriority
		}
		if price.CacheWritePrice != nil {
			cloned.CacheCreationPricePerToken = *price.CacheWritePrice
			cloned.CacheCreation5mPrice = *price.CacheWritePrice
			cloned.CacheCreation1hPrice = *price.CacheWritePrice
			cloned.CacheCreationPriceExplicit = true
		}
		if price.CacheWritePricePriority != nil {
			cloned.CacheCreationPricePerTokenPriority = *price.CacheWritePricePriority
		}
		if price.CacheReadPrice != nil {
			cloned.CacheReadPricePerToken = *price.CacheReadPrice
		}
		if price.CacheReadPricePriority != nil {
			cloned.CacheReadPricePerTokenPriority = *price.CacheReadPricePriority
		}
		out.BasePricing = &cloned
	}
	if price.PerRequestPrice != nil {
		out.DefaultPerRequestPrice = *price.PerRequestPrice
	}
	if len(resolved.Intervals) > 0 {
		intervals := make([]PricingInterval, len(resolved.Intervals))
		copy(intervals, resolved.Intervals)
		for i := range intervals {
			if price.InputPrice != nil {
				value := *price.InputPrice
				intervals[i].InputPrice = &value
				intervals[i].InputMultiplier = nil
			}
			if price.OutputPrice != nil {
				value := *price.OutputPrice
				intervals[i].OutputPrice = &value
				intervals[i].OutputMultiplier = nil
			}
			if price.CacheWritePrice != nil {
				value := *price.CacheWritePrice
				intervals[i].CacheWritePrice = &value
				intervals[i].CacheWriteMultiplier = nil
			}
			if price.CacheReadPrice != nil {
				value := *price.CacheReadPrice
				intervals[i].CacheReadPrice = &value
				intervals[i].CacheReadMultiplier = nil
			}
			if price.PerRequestPrice != nil {
				value := *price.PerRequestPrice
				intervals[i].PerRequestPrice = &value
			}
		}
		out.Intervals = intervals
	}
	return &out
}
