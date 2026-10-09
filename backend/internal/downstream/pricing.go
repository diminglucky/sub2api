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
	InputPrice      *float64
	OutputPrice     *float64
	CacheWritePrice *float64
	CacheReadPrice  *float64
	PerRequestPrice *float64
}

// Clone returns a copy with independently allocated numeric pointers.
func (p Price) Clone() Price {
	return Price{
		InputPrice:      cloneFloat64(p.InputPrice),
		OutputPrice:     cloneFloat64(p.OutputPrice),
		CacheWritePrice: cloneFloat64(p.CacheWritePrice),
		CacheReadPrice:  cloneFloat64(p.CacheReadPrice),
		PerRequestPrice: cloneFloat64(p.PerRequestPrice),
	}
}

// IsZero reports whether every price field is unset.
func (p Price) IsZero() bool {
	return p.InputPrice == nil &&
		p.OutputPrice == nil &&
		p.CacheWritePrice == nil &&
		p.CacheReadPrice == nil &&
		p.PerRequestPrice == nil
}

// Equal reports whether two prices carry the same set of values. Pointers are
// compared by value, so callers can detect a no-op override.
func (p Price) Equal(other Price) bool {
	return priceFieldEqual(p.InputPrice, other.InputPrice) &&
		priceFieldEqual(p.OutputPrice, other.OutputPrice) &&
		priceFieldEqual(p.CacheWritePrice, other.CacheWritePrice) &&
		priceFieldEqual(p.CacheReadPrice, other.CacheReadPrice) &&
		priceFieldEqual(p.PerRequestPrice, other.PerRequestPrice)
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
			       cache_write_price, cache_read_price, per_request_price
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
			       cache_write_price, cache_read_price, per_request_price
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
	)
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&multiplier,
		&inputPrice,
		&outputPrice,
		&cacheWritePrice,
		&cacheReadPrice,
		&perRequestPrice,
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
	out.InputPrice = applyPriceField(out.InputPrice, override.InputPrice, multiplier)
	out.OutputPrice = applyPriceField(out.OutputPrice, override.OutputPrice, multiplier)
	out.CacheWritePrice = applyPriceField(out.CacheWritePrice, override.CacheWritePrice, multiplier)
	out.CacheReadPrice = applyPriceField(out.CacheReadPrice, override.CacheReadPrice, multiplier)
	out.PerRequestPrice = applyPriceField(out.PerRequestPrice, override.PerRequestPrice, multiplier)
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
