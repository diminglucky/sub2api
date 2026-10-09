// Package downstream implements the downstream subsite foundation for Sub2API.
//
// A subsite is a branded downstream site (for example draw.superai.sbs) that
// shares the main site's users, balance, payment orders, API keys, models and
// gateway. The main site remains the single control plane. Every
// subsite-scoped row carries a subsite_id; a NULL subsite_id means main-site
// data.
//
// This package only defines the schema-facing types and read helpers. Business
// rules, pricing resolution and settlement generation live in later tasks.
package downstream

import "time"

// Subsite status values.
const (
	SubsiteStatusActive   = "active"
	SubsiteStatusDisabled = "disabled"
)

// SubsiteMember role values. An admin can manage the subsite's own prices and
// view its scoped dashboard; a member is a regular user of the subsite.
const (
	SubsiteMemberRoleOwner  = "owner"
	SubsiteMemberRoleAdmin  = "admin"
	SubsiteMemberRoleMember = "member"
)

// SubsiteMember source values record how the membership was created.
const (
	SubsiteMemberSourceRegistration = "registration"
	SubsiteMemberSourceManual       = "manual"
)

// PriceOverride scope values. An override targets either a whole group or a
// single model name. Model scope wins over group scope during resolution.
const (
	PriceOverrideScopeGroup = "group"
	PriceOverrideScopeModel = "model"
)

// Settlement status values. V1 only records payable ledger entries; it never
// pays out automatically.
const (
	SettlementStatusPending = "pending"
	SettlementStatusSettled = "settled"
	SettlementStatusVoid    = "void"
)

// Subsite is a downstream tenant serving a branded site on
// "<slug>.superai.sbs". The V1 pilot tenant is draw.superai.sbs.
type Subsite struct {
	ID          int64
	Slug        string
	Domain      string
	Name        string
	LogoURL     string
	ThemeColor  string
	Status      string
	AdminUserID *int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsActive reports whether the subsite may serve traffic.
func (s *Subsite) IsActive() bool {
	return s != nil && s.Status == SubsiteStatusActive
}

// SubsiteMember links a main-site user to a subsite. Users may belong to more
// than one subsite; only rows for the current subsite are aggregated there.
type SubsiteMember struct {
	ID        int64
	SubsiteID int64
	UserID    int64
	Role      string
	Source    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsAdmin reports whether the member may manage the subsite.
func (m *SubsiteMember) IsAdmin() bool {
	if m == nil {
		return false
	}
	return m.Role == SubsiteMemberRoleOwner || m.Role == SubsiteMemberRoleAdmin
}

// PriceOverride is a subsite-scoped price override for a group or a model.
// RateMultiplier is applied on top of the resolved main-site price; explicit
// price fields, when set, take precedence over the multiplier.
type PriceOverride struct {
	ID              int64
	SubsiteID       int64
	Scope           string
	GroupID         *int64
	Model           *string
	RateMultiplier  float64
	InputPrice      *float64
	OutputPrice     *float64
	CacheWritePrice *float64
	CacheReadPrice  *float64
	PerRequestPrice *float64
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// SettlementEntry records the payable difference for one paid recharge order.
// GrossAmount is what the main site collected; CostAmount is the platform cost
// attributed to the order; MarginAmount is what the subsite is owed. V1 only
// writes ledger rows and leaves payout to a manual process.
type SettlementEntry struct {
	ID           int64
	SubsiteID    int64
	OrderID      *int64
	UserID       *int64
	Currency     string
	GrossAmount  float64
	CostAmount   float64
	MarginAmount float64
	Status       string
	Notes        string
	SettledAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SubsiteUserSummary is the subsite-scoped view of one main-site user. Balance
// and account identity are shared with the main site; usage and recharge totals
// are limited to rows whose subsite_id matches the current subsite.
type SubsiteUserSummary struct {
	UserID         int64   `json:"user_id"`
	SubsiteID      int64   `json:"subsite_id"`
	Email          string  `json:"email"`
	Balance        float64 `json:"balance"`
	Role           string  `json:"role"`
	Source         string  `json:"source"`
	Admin          bool    `json:"is_admin"`
	UsageCount     int64   `json:"usage_count"`
	UsageCost      float64 `json:"usage_cost"`
	RechargeCount  int64   `json:"recharge_count"`
	RechargeAmount float64 `json:"recharge_amount"`
}

// IsAdmin reports whether this member can view the subsite admin dashboard.
func (s *SubsiteUserSummary) IsAdmin() bool {
	if s == nil {
		return false
	}
	return s.Role == SubsiteMemberRoleOwner || s.Role == SubsiteMemberRoleAdmin
}

// SubsiteAdminSummary is the aggregate dashboard view for one subsite. All
// totals are scoped to the current subsite; NULL subsite_id rows are main-site
// data and are never included.
type SubsiteAdminSummary struct {
	SubsiteID         int64   `json:"subsite_id"`
	MemberCount       int64   `json:"member_count"`
	RechargeCount     int64   `json:"recharge_count"`
	RechargeAmount    float64 `json:"recharge_amount"`
	UsageCount        int64   `json:"usage_count"`
	UsageCost         float64 `json:"usage_cost"`
	SettlementPending float64 `json:"settlement_pending"`
}
