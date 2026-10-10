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
	Enabled               bool     `json:"enabled"`
}

// ListSubsiteChannels returns every active main-site group and the sub-site's
// group-scoped price override for it, if one exists.
func (r *Repository) ListSubsiteChannels(ctx context.Context, subsiteID int64) ([]SubsiteChannel, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	const query = `
		SELECT g.id, g.name, g.platform, g.rate_multiplier, p.rate_multiplier
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
			item       SubsiteChannel
			subsiteMul sql.NullFloat64
		)
		if err := rows.Scan(&item.GroupID, &item.Name, &item.Platform, &item.MainRateMultiplier, &subsiteMul); err != nil {
			return nil, err
		}
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
