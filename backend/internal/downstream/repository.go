package downstream

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrSubsiteNotFound is returned when no subsite matches the requested slug or
// domain. Callers should treat it as a client error (404) rather than a
// server failure.
var ErrSubsiteNotFound = errors.New("downstream: subsite not found")

// ErrSubsiteUserNotFound is returned when a subsite summary is requested for a
// main-site user that does not exist or has been deleted.
var ErrSubsiteUserNotFound = errors.New("downstream: subsite user not found")

// ErrSubsiteAdminRequired is returned when a user requests a subsite admin
// summary but is not an owner/admin of that subsite.
var ErrSubsiteAdminRequired = errors.New("downstream: subsite admin required")

// Repository provides read access to subsite-scoped data. Every method that
// returns subsite-owned rows must be filtered by subsite_id so a subsite can
// never observe another subsite's data.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a subsite repository backed by the given database.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const subsiteColumns = `id, slug, domain, name, logo_url, theme_color, status, admin_user_id, created_at, updated_at`

// GetSubsiteBySlug resolves a subsite from its slug (for example "draw" for
// draw.superai.sbs). The slug is normalized before lookup so host-derived
// values are matched case-insensitively.
func (r *Repository) GetSubsiteBySlug(ctx context.Context, slug string) (*Subsite, error) {
	normalized := NormalizeSlug(slug)
	if normalized == "" {
		return nil, ErrSubsiteNotFound
	}
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	row := r.db.QueryRowContext(ctx,
		`SELECT `+subsiteColumns+` FROM subsites WHERE slug = $1`,
		normalized,
	)
	return scanSubsite(row)
}

const subsiteUserSummarySQL = `
SELECT
    u.id,
    $1::bigint AS subsite_id,
    u.email,
    u.balance,
    COALESCE(sm.role, '') AS role,
    COALESCE(sm.source, '') AS source,
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
    ), 0) AS recharge_amount
FROM users u
LEFT JOIN subsite_members sm
  ON sm.subsite_id = $1 AND sm.user_id = u.id
WHERE u.id = $2 AND u.deleted_at IS NULL`

// GetSubsiteUserSummary returns the user's shared account data plus usage and
// recharge totals scoped to subsiteID. A user who exists on the main site but is
// not yet a member of the subsite still gets a valid summary with an empty role
// and zero subsite-specific totals.
func (r *Repository) GetSubsiteUserSummary(ctx context.Context, subsiteID, userID int64) (*SubsiteUserSummary, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if subsiteID <= 0 || userID <= 0 {
		return nil, ErrSubsiteUserNotFound
	}

	var summary SubsiteUserSummary
	if err := r.db.QueryRowContext(ctx, subsiteUserSummarySQL, subsiteID, userID).Scan(
		&summary.UserID,
		&summary.SubsiteID,
		&summary.Email,
		&summary.Balance,
		&summary.Role,
		&summary.Source,
		&summary.UsageCount,
		&summary.UsageCost,
		&summary.RechargeCount,
		&summary.RechargeAmount,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubsiteUserNotFound
		}
		return nil, fmt.Errorf("downstream: scan subsite user summary: %w", err)
	}
	summary.Admin = summary.IsAdmin()
	return &summary, nil
}

const subsiteAdminAccessSQL = `
SELECT 1
FROM subsites s
LEFT JOIN subsite_members sm
  ON sm.subsite_id = s.id AND sm.user_id = $2
WHERE s.id = $1
  AND s.status = 'active'
  AND (s.admin_user_id = $2 OR sm.role IN ('owner', 'admin'))`

const subsiteAdminSummarySQL = `
SELECT
    $1::bigint AS subsite_id,
    COALESCE((
        SELECT COUNT(*) FROM subsite_members sm
        WHERE sm.subsite_id = $1
    ), 0) AS member_count,
    COALESCE((
        SELECT COUNT(*) FROM payment_orders po
        WHERE po.subsite_id = $1
          AND po.status IN ('PAID', 'RECHARGING', 'COMPLETED')
    ), 0) AS recharge_count,
    COALESCE((
        SELECT SUM(po.amount) FROM payment_orders po
        WHERE po.subsite_id = $1
          AND po.status IN ('PAID', 'RECHARGING', 'COMPLETED')
    ), 0) AS recharge_amount,
    COALESCE((
        SELECT COUNT(*) FROM usage_logs ul
        WHERE ul.subsite_id = $1
    ), 0) AS usage_count,
    COALESCE((
        SELECT SUM(ul.actual_cost) FROM usage_logs ul
        WHERE ul.subsite_id = $1
    ), 0) AS usage_cost,
    COALESCE((
        SELECT SUM(sl.margin_amount) FROM settlement_ledger sl
        WHERE sl.subsite_id = $1 AND sl.status = 'pending'
    ), 0) AS settlement_pending`

// GetSubsiteAdminSummary verifies that userID is an owner/admin of subsiteID
// before returning scoped aggregates. Callers must not bypass this method for
// admin dashboard data.
func (r *Repository) GetSubsiteAdminSummary(ctx context.Context, subsiteID, userID int64) (*SubsiteAdminSummary, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if subsiteID <= 0 || userID <= 0 {
		return nil, ErrSubsiteAdminRequired
	}

	var allowed int
	if err := r.db.QueryRowContext(ctx, subsiteAdminAccessSQL, subsiteID, userID).Scan(&allowed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubsiteAdminRequired
		}
		return nil, fmt.Errorf("downstream: check subsite admin: %w", err)
	}

	var summary SubsiteAdminSummary
	if err := r.db.QueryRowContext(ctx, subsiteAdminSummarySQL, subsiteID).Scan(
		&summary.SubsiteID,
		&summary.MemberCount,
		&summary.RechargeCount,
		&summary.RechargeAmount,
		&summary.UsageCount,
		&summary.UsageCost,
		&summary.SettlementPending,
	); err != nil {
		return nil, fmt.Errorf("downstream: scan subsite admin summary: %w", err)
	}
	return &summary, nil
}

// NormalizeSlug lowercases and trims a slug so that slugs derived from
// hostnames ("DRAW.superai.sbs") and stored slugs ("draw") compare equal.
func NormalizeSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSubsite(row rowScanner) (*Subsite, error) {
	var (
		subsite Subsite
		adminID sql.NullInt64
	)
	if err := row.Scan(
		&subsite.ID,
		&subsite.Slug,
		&subsite.Domain,
		&subsite.Name,
		&subsite.LogoURL,
		&subsite.ThemeColor,
		&subsite.Status,
		&adminID,
		&subsite.CreatedAt,
		&subsite.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubsiteNotFound
		}
		return nil, fmt.Errorf("downstream: scan subsite: %w", err)
	}
	if adminID.Valid {
		value := adminID.Int64
		subsite.AdminUserID = &value
	}
	return &subsite, nil
}
