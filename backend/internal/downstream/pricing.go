package downstream

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
)

// Price is the subsite-facing model price shape. Nil fields mean "not set";
// callers should preserve that distinction so an explicit zero override can
// intentionally make a price free.
type Price struct {
	InputPrice              *float64
	InputPricePriority      *float64
	OutputPrice             *float64
	OutputPricePriority     *float64
	CacheWritePrice         *float64
	CacheWritePricePriority *float64
	CacheReadPrice          *float64
	CacheReadPricePriority  *float64
	PerRequestPrice         *float64
	ImagePrice1K            *float64
	ImagePrice2K            *float64
	ImagePrice4K            *float64
}

// Clone returns a copy with independently allocated numeric pointers.
func (p Price) Clone() Price {
	return Price{
		InputPrice:              cloneFloat64(p.InputPrice),
		InputPricePriority:      cloneFloat64(p.InputPricePriority),
		OutputPrice:             cloneFloat64(p.OutputPrice),
		OutputPricePriority:     cloneFloat64(p.OutputPricePriority),
		CacheWritePrice:         cloneFloat64(p.CacheWritePrice),
		CacheWritePricePriority: cloneFloat64(p.CacheWritePricePriority),
		CacheReadPrice:          cloneFloat64(p.CacheReadPrice),
		CacheReadPricePriority:  cloneFloat64(p.CacheReadPricePriority),
		PerRequestPrice:         cloneFloat64(p.PerRequestPrice),
		ImagePrice1K:            cloneFloat64(p.ImagePrice1K),
		ImagePrice2K:            cloneFloat64(p.ImagePrice2K),
		ImagePrice4K:            cloneFloat64(p.ImagePrice4K),
	}
}

// IsZero reports whether every price field is unset.
func (p Price) IsZero() bool {
	return p.InputPrice == nil &&
		p.InputPricePriority == nil &&
		p.OutputPrice == nil &&
		p.OutputPricePriority == nil &&
		p.CacheWritePrice == nil &&
		p.CacheWritePricePriority == nil &&
		p.CacheReadPrice == nil &&
		p.CacheReadPricePriority == nil &&
		p.PerRequestPrice == nil &&
		p.ImagePrice1K == nil &&
		p.ImagePrice2K == nil &&
		p.ImagePrice4K == nil
}

// Equal reports whether two prices carry the same set of values. Pointers are
// compared by value, so callers can detect a no-op override.
func (p Price) Equal(other Price) bool {
	return priceFieldEqual(p.InputPrice, other.InputPrice) &&
		priceFieldEqual(p.InputPricePriority, other.InputPricePriority) &&
		priceFieldEqual(p.OutputPrice, other.OutputPrice) &&
		priceFieldEqual(p.OutputPricePriority, other.OutputPricePriority) &&
		priceFieldEqual(p.CacheWritePrice, other.CacheWritePrice) &&
		priceFieldEqual(p.CacheWritePricePriority, other.CacheWritePricePriority) &&
		priceFieldEqual(p.CacheReadPrice, other.CacheReadPrice) &&
		priceFieldEqual(p.CacheReadPricePriority, other.CacheReadPricePriority) &&
		priceFieldEqual(p.PerRequestPrice, other.PerRequestPrice) &&
		priceFieldEqual(p.ImagePrice1K, other.ImagePrice1K) &&
		priceFieldEqual(p.ImagePrice2K, other.ImagePrice2K) &&
		priceFieldEqual(p.ImagePrice4K, other.ImagePrice4K)
}

func priceFieldEqual(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return math.Abs(*a-*b) < 1e-12
}

// ResolveSubsitePrice applies the subsite override precedence:
// model override, group override, then the supplied base price.
//
// The base price is expected to already include any main-site group/base
// pricing. Supplying a groupID is optional; when present, a group-scoped
// override is checked after the model scope misses.
func (r *Repository) ResolveSubsitePrice(
	ctx context.Context,
	subsiteID int64,
	model string,
	basePrice Price,
	groupID ...int64,
) (Price, error) {
	if subsiteID <= 0 {
		return basePrice.Clone(), nil
	}
	if r == nil || r.db == nil {
		return Price{}, errors.New("downstream: nil repository database")
	}

	model = strings.TrimSpace(model)
	if model != "" {
		override, err := r.lookupPriceOverride(ctx, subsiteID, PriceOverrideScopeModel, model, 0)
		if err == nil {
			return applyPriceOverride(basePrice, override), nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return Price{}, err
		}
	}

	if len(groupID) > 0 && groupID[0] > 0 {
		override, err := r.lookupPriceOverride(ctx, subsiteID, PriceOverrideScopeGroup, "", groupID[0])
		if err == nil {
			return applyPriceOverride(basePrice, override), nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return Price{}, err
		}
	}

	return basePrice.Clone(), nil
}

func (r *Repository) lookupPriceOverride(
	ctx context.Context,
	subsiteID int64,
	scope string,
	model string,
	groupID int64,
) (*PriceOverride, error) {
	var (
		query string
		args  []any
	)
	switch scope {
	case PriceOverrideScopeModel:
		query = `
			SELECT rate_multiplier, input_price, output_price,
			       cache_write_price, cache_read_price, per_request_price,
			       image_price_1k, image_price_2k, image_price_4k
			FROM subsite_prices
			WHERE subsite_id = $1
			  AND scope = 'model'
			  AND model = $2
			  AND status = 'active'
			ORDER BY id DESC
			LIMIT 1`
		args = []any{subsiteID, model}
	case PriceOverrideScopeGroup:
		query = `
			SELECT rate_multiplier, input_price, output_price,
			       cache_write_price, cache_read_price, per_request_price,
			       image_price_1k, image_price_2k, image_price_4k
			FROM subsite_prices
			WHERE subsite_id = $1
			  AND scope = 'group'
			  AND group_id = $2
			  AND status = 'active'
			ORDER BY id DESC
			LIMIT 1`
		args = []any{subsiteID, groupID}
	default:
		return nil, fmt.Errorf("downstream: unsupported price override scope %q", scope)
	}

	var (
		multiplier      float64
		inputPrice      sql.NullFloat64
		outputPrice     sql.NullFloat64
		cacheWritePrice sql.NullFloat64
		cacheReadPrice  sql.NullFloat64
		perRequestPrice sql.NullFloat64
		imagePrice1K    sql.NullFloat64
		imagePrice2K    sql.NullFloat64
		imagePrice4K    sql.NullFloat64
	)
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&multiplier,
		&inputPrice,
		&outputPrice,
		&cacheWritePrice,
		&cacheReadPrice,
		&perRequestPrice,
		&imagePrice1K,
		&imagePrice2K,
		&imagePrice4K,
	); err != nil {
		return nil, err
	}
	if math.IsNaN(multiplier) || math.IsInf(multiplier, 0) || multiplier < 0 {
		multiplier = 1
	}
	return &PriceOverride{
		Scope:           scope,
		GroupID:         nullableInt64FromOverride(scope, groupID),
		Model:           nullableModelFromOverride(scope, model),
		RateMultiplier:  multiplier,
		InputPrice:      nullableFloat64(inputPrice),
		OutputPrice:     nullableFloat64(outputPrice),
		CacheWritePrice: nullableFloat64(cacheWritePrice),
		CacheReadPrice:  nullableFloat64(cacheReadPrice),
		PerRequestPrice: nullableFloat64(perRequestPrice),
		ImagePrice1K:    nullableFloat64(imagePrice1K),
		ImagePrice2K:    nullableFloat64(imagePrice2K),
		ImagePrice4K:    nullableFloat64(imagePrice4K),
		Status:          SubsiteStatusActive,
	}, nil
}

func applyPriceOverride(base Price, override *PriceOverride) Price {
	out := base.Clone()
	if override == nil {
		return out
	}
	multiplier := override.RateMultiplier
	if math.IsNaN(multiplier) || math.IsInf(multiplier, 0) || multiplier < 0 {
		multiplier = 1
	}
	baseInputPrice := out.InputPrice
	baseInputPricePriority := out.InputPricePriority
	baseOutputPrice := out.OutputPrice
	baseOutputPricePriority := out.OutputPricePriority
	baseCacheWritePrice := out.CacheWritePrice
	baseCacheWritePricePriority := out.CacheWritePricePriority
	baseCacheReadPrice := out.CacheReadPrice
	baseCacheReadPricePriority := out.CacheReadPricePriority
	out.InputPrice = applyPriceField(out.InputPrice, override.InputPrice, multiplier)
	out.InputPricePriority = applyPriorityOverrideField(
		baseInputPrice,
		baseInputPricePriority,
		override.InputPrice,
		multiplier,
	)
	out.OutputPrice = applyPriceField(out.OutputPrice, override.OutputPrice, multiplier)
	out.OutputPricePriority = applyPriorityOverrideField(
		baseOutputPrice,
		baseOutputPricePriority,
		override.OutputPrice,
		multiplier,
	)
	out.CacheWritePrice = applyPriceField(out.CacheWritePrice, override.CacheWritePrice, multiplier)
	out.CacheWritePricePriority = applyPriorityOverrideField(
		baseCacheWritePrice,
		baseCacheWritePricePriority,
		override.CacheWritePrice,
		multiplier,
	)
	out.CacheReadPrice = applyPriceField(out.CacheReadPrice, override.CacheReadPrice, multiplier)
	out.CacheReadPricePriority = applyPriorityOverrideField(
		baseCacheReadPrice,
		baseCacheReadPricePriority,
		override.CacheReadPrice,
		multiplier,
	)
	out.PerRequestPrice = applyPriceField(out.PerRequestPrice, override.PerRequestPrice, multiplier)
	// 生图一口价是绝对价格（不随倍率缩放），直接覆盖。
	if override.ImagePrice1K != nil {
		out.ImagePrice1K = cloneFloat64(override.ImagePrice1K)
	}
	if override.ImagePrice2K != nil {
		out.ImagePrice2K = cloneFloat64(override.ImagePrice2K)
	}
	if override.ImagePrice4K != nil {
		out.ImagePrice4K = cloneFloat64(override.ImagePrice4K)
	}
	return out
}

func applyPriceField(base *float64, explicit *float64, multiplier float64) *float64 {
	if explicit != nil {
		return cloneFloat64(explicit)
	}
	if base == nil {
		return nil
	}
	value := *base * multiplier
	return &value
}

func applyPriorityOverrideField(
	base *float64,
	basePriority *float64,
	explicit *float64,
	multiplier float64,
) *float64 {
	if explicit == nil {
		return applyPriceField(basePriority, nil, multiplier)
	}
	if base == nil || basePriority == nil || *base == 0 || *basePriority == 0 {
		return nil
	}
	value := *explicit * (*basePriority / *base)
	return &value
}

func nullableFloat64(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	v := value.Float64
	return &v
}

func nullableInt64FromOverride(scope string, groupID int64) *int64 {
	if scope != PriceOverrideScopeGroup || groupID <= 0 {
		return nil
	}
	v := groupID
	return &v
}

func nullableModelFromOverride(scope, model string) *string {
	if scope != PriceOverrideScopeModel || strings.TrimSpace(model) == "" {
		return nil
	}
	v := strings.TrimSpace(model)
	return &v
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	v := *value
	return &v
}
