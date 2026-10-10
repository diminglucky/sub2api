package downstream

import (
	"context"
	"errors"
	"time"
)

// SubsiteMemberSummary is one member of a sub-site with the same scoped totals
// as SubsiteUserSummary, used by the sub-site user-management list.
type SubsiteMemberSummary struct {
	UserID         int64     `json:"user_id"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	Source         string    `json:"source"`
	Balance        float64   `json:"balance"`
	UsageCount     int64     `json:"usage_count"`
	UsageCost      float64   `json:"usage_cost"`
	RechargeCount  int64     `json:"recharge_count"`
	RechargeAmount float64   `json:"recharge_amount"`
	JoinedAt       time.Time `json:"joined_at"`
}

const subsiteMembersSQL = `
SELECT
    u.id,
    u.email,
    COALESCE(sm.role, '') AS role,
    COALESCE(sm.source, '') AS source,
    u.balance,
    COALESCE((
        SELECT COUNT(*) FROM usage_logs ul
        WHERE ul.subsite_id = $1 AND ul.user_id = u.id
    ), 0) AS usage_count,
    COALESCE((
        SELECT SUM(ul.actual_cost) FROM usage_logs ul
        WHERE ul.subsite_id = $1 AND ul.user_id = u.id
    ), 0) AS usage_cost,
    COALESCE((
        SELECT COUNT(*) FROM payment_orders po
        WHERE po.subsite_id = $1
          AND po.user_id = u.id
          AND po.status IN ('PAID', 'RECHARGING', 'COMPLETED')
    ), 0) AS recharge_count,
    COALESCE((
        SELECT SUM(po.amount) FROM payment_orders po
        WHERE po.subsite_id = $1
          AND po.user_id = u.id
          AND po.status IN ('PAID', 'RECHARGING', 'COMPLETED')
    ), 0) AS recharge_amount,
    sm.created_at
FROM subsite_members sm
JOIN users u ON u.id = sm.user_id
WHERE sm.subsite_id = $1
  AND u.deleted_at IS NULL
ORDER BY sm.created_at DESC, u.id DESC
LIMIT $2 OFFSET $3`

// ListSubsiteMembers returns a page of members belonging to subsiteID with
// their sub-site-scoped usage and recharge totals.
func (r *Repository) ListSubsiteMembers(ctx context.Context, subsiteID int64, limit, offset int) ([]SubsiteMemberSummary, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if subsiteID <= 0 {
		return nil, ErrSubsiteNotFound
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, subsiteMembersSQL, subsiteID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]SubsiteMemberSummary, 0)
	for rows.Next() {
		var item SubsiteMemberSummary
		if err := rows.Scan(
			&item.UserID,
			&item.Email,
			&item.Role,
			&item.Source,
			&item.Balance,
			&item.UsageCount,
			&item.UsageCost,
			&item.RechargeCount,
			&item.RechargeAmount,
			&item.JoinedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
