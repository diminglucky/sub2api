package downstream

import (
	"context"
	"database/sql"
	"errors"
)

// ModelCatalogEntry is one (group, model) pair from the main-site pricing
// catalog, provided by the service layer so this package stays decoupled.
type ModelCatalogEntry struct {
	GroupID        int64
	GroupName      string
	MainMultiplier float64
	Model          string
	Platform       string
	BillingMode    string
	// MainPerRequestPrice is the main-site effective flat price per request
	// (channel per-request price x group multiplier) for per-request/image
	// models. Nil for token models.
	MainPerRequestPrice *float64
}

// ModelCatalogProvider lists the main-site pricing catalog.
type ModelCatalogProvider func(ctx context.Context) ([]ModelCatalogEntry, error)

// SubsiteModelPrice is one model offered by the sub-site: the main-site
// multiplier, and the sub-site's per-model multiplier override if set.
type SubsiteModelPrice struct {
	GroupID           int64    `json:"group_id"`
	GroupName         string   `json:"group_name"`
	MainMultiplier    float64  `json:"main_multiplier"`
	Model             string   `json:"model"`
	Platform          string   `json:"platform"`
	BillingMode       string   `json:"billing_mode"`
	SubsiteMultiplier *float64 `json:"subsite_multiplier"`
	// 主站一次价（按次/生图模型），只读。
	MainPerRequestPrice *float64 `json:"main_per_request_price"`
	// 子站一次价覆盖（按次/生图模型）；nil 表示未覆盖。
	SubsitePerRequestPrice *float64 `json:"subsite_per_request_price"`
}

// ListSubsiteModelPrices joins the main-site catalog with the sub-site's
// per-model multiplier overrides, scoped to the groups open to the sub-site.
func (r *Repository) ListSubsiteModelPrices(
	ctx context.Context,
	subsiteID int64,
	catalog []ModelCatalogEntry,
) ([]SubsiteModelPrice, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if subsiteID <= 0 {
		return nil, ErrSubsiteNotFound
	}

	// Groups open to the sub-site.
	rows, err := r.db.QueryContext(ctx,
		`SELECT group_id FROM subsite_groups WHERE subsite_id = $1`,
		subsiteID,
	)
	if err != nil {
		return nil, err
	}
	assigned := make(map[int64]struct{})
	for rows.Next() {
		var groupID int64
		if err := rows.Scan(&groupID); err != nil {
			_ = rows.Close()
			return nil, err
		}
		assigned[groupID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	_ = rows.Close()

	// Per-model multiplier overrides.
	type overrideRow struct {
		multiplier float64
		perRequest *float64
	}
	overrides := make(map[string]overrideRow)
	priceRows, err := r.db.QueryContext(ctx, `
		SELECT model, rate_multiplier, per_request_price FROM subsite_prices
		WHERE subsite_id = $1 AND scope = 'model' AND status = 'active'`,
		subsiteID,
	)
	if err != nil {
		return nil, err
	}
	for priceRows.Next() {
		var (
			model      sql.NullString
			multiplier float64
			perRequest sql.NullFloat64
		)
		if err := priceRows.Scan(&model, &multiplier, &perRequest); err != nil {
			_ = priceRows.Close()
			return nil, err
		}
		if model.Valid {
			row := overrideRow{multiplier: multiplier}
			if perRequest.Valid {
				value := perRequest.Float64
				row.perRequest = &value
			}
			overrides[model.String] = row
		}
	}
	if err := priceRows.Err(); err != nil {
		_ = priceRows.Close()
		return nil, err
	}
	_ = priceRows.Close()

	out := make([]SubsiteModelPrice, 0)
	seen := make(map[string]struct{})
	for _, entry := range catalog {
		if _, ok := assigned[entry.GroupID]; !ok {
			continue
		}
		key := entry.Model
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		item := SubsiteModelPrice{
			GroupID:             entry.GroupID,
			GroupName:           entry.GroupName,
			MainMultiplier:      entry.MainMultiplier,
			Model:               entry.Model,
			Platform:            entry.Platform,
			BillingMode:         entry.BillingMode,
			MainPerRequestPrice: entry.MainPerRequestPrice,
		}
		if override, ok := overrides[entry.Model]; ok {
			value := override.multiplier
			item.SubsiteMultiplier = &value
			item.SubsitePerRequestPrice = override.perRequest
		}
		out = append(out, item)
	}
	return out, nil
}
