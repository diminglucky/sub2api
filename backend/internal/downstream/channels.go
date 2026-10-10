package downstream

import (
	"context"
	"database/sql"
	"errors"
)

// SubsiteChannel is one main-site group (channel) offered to a sub-site,
// together with the sub-site's current price override if any.
//
// The main-site rate is informational only; a sub-site may never undercut it.
type SubsiteChannel struct {
	GroupID               int64    `json:"group_id"`
	Name                  string   `json:"name"`
	Platform              string   `json:"platform"`
	MainRateMultiplier    float64  `json:"main_rate_multiplier"`
	SubsiteRateMultiplier *float64 `json:"subsite_rate_multiplier"`
	// 主站生图一口价（按尺寸，USD/张）。
	MainImagePrice1K *float64 `json:"main_image_price_1k"`
	MainImagePrice2K *float64 `json:"main_image_price_2k"`
	MainImagePrice4K *float64 `json:"main_image_price_4k"`
	// 子站生图一口价（按尺寸），NULL 表示未覆盖。
	SubsiteImagePrice1K *float64 `json:"subsite_image_price_1k"`
	SubsiteImagePrice2K *float64 `json:"subsite_image_price_2k"`
	SubsiteImagePrice4K *float64 `json:"subsite_image_price_4k"`
	Enabled             bool     `json:"enabled"`
}

// ListSubsiteChannels returns every active main-site group and the sub-site's
// group-scoped price override for it, if one exists.
func (r *Repository) ListSubsiteChannels(ctx context.Context, subsiteID int64) ([]SubsiteChannel, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	const query = `
		SELECT g.id, g.name, g.platform, g.rate_multiplier,
		       g.image_price_1k, g.image_price_2k, g.image_price_4k,
		       p.rate_multiplier, p.image_price_1k, p.image_price_2k, p.image_price_4k
		FROM groups g
		LEFT JOIN subsite_prices p
		  ON p.subsite_id = $1
		 AND p.scope = 'group'
		 AND p.group_id = g.id
		 AND p.status = 'active'
		WHERE g.status = 'active' AND g.deleted_at IS NULL
		ORDER BY g.sort_order ASC, g.id ASC`

	rows, err := r.db.QueryContext(ctx, query, subsiteID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]SubsiteChannel, 0)
	for rows.Next() {
		var (
			item           SubsiteChannel
			mainImage1K    sql.NullFloat64
			mainImage2K    sql.NullFloat64
			mainImage4K    sql.NullFloat64
			subsiteMul     sql.NullFloat64
			subsiteImage1K sql.NullFloat64
			subsiteImage2K sql.NullFloat64
			subsiteImage4K sql.NullFloat64
		)
		if err := rows.Scan(
			&item.GroupID,
			&item.Name,
			&item.Platform,
			&item.MainRateMultiplier,
			&mainImage1K,
			&mainImage2K,
			&mainImage4K,
			&subsiteMul,
			&subsiteImage1K,
			&subsiteImage2K,
			&subsiteImage4K,
		); err != nil {
			return nil, err
		}
		item.MainImagePrice1K = nullableFloat64(mainImage1K)
		item.MainImagePrice2K = nullableFloat64(mainImage2K)
		item.MainImagePrice4K = nullableFloat64(mainImage4K)
		item.SubsiteImagePrice1K = nullableFloat64(subsiteImage1K)
		item.SubsiteImagePrice2K = nullableFloat64(subsiteImage2K)
		item.SubsiteImagePrice4K = nullableFloat64(subsiteImage4K)
		if subsiteMul.Valid {
			value := subsiteMul.Float64
			item.SubsiteRateMultiplier = &value
			item.Enabled = true
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// GroupRateMultiplier returns the main-site rate multiplier of one group. It is
// used to reject sub-site prices that would undercut the main site.
func (r *Repository) GroupRateMultiplier(ctx context.Context, groupID int64) (float64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("downstream: nil repository database")
	}
	var rate float64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT rate_multiplier FROM groups WHERE id = $1 AND status = 'active' AND deleted_at IS NULL`,
		groupID,
	).Scan(&rate)
	if err != nil {
		return 0, err
	}
	return rate, nil
}

// GroupImagePrices returns the main-site per-size image prices of one group.
func (r *Repository) GroupImagePrices(ctx context.Context, groupID int64) (ImagePriceTiers, error) {
	if r == nil || r.db == nil {
		return ImagePriceTiers{}, errors.New("downstream: nil repository database")
	}
	var (
		price1K sql.NullFloat64
		price2K sql.NullFloat64
		price4K sql.NullFloat64
	)
	err := r.db.QueryRowContext(
		ctx,
		`SELECT image_price_1k, image_price_2k, image_price_4k FROM groups WHERE id = $1 AND status = 'active' AND deleted_at IS NULL`,
		groupID,
	).Scan(&price1K, &price2K, &price4K)
	if err != nil {
		return ImagePriceTiers{}, err
	}
	return ImagePriceTiers{
		Price1K: nullableFloat64(price1K),
		Price2K: nullableFloat64(price2K),
		Price4K: nullableFloat64(price4K),
	}, nil
}

// ImagePriceTiers is the main-site per-size image price of a group.
type ImagePriceTiers struct {
	Price1K *float64
	Price2K *float64
	Price4K *float64
}
