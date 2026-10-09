package downstream

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
)

// UsageSettlement is one usage-based settlement ledger entry.
//
// Business rule (2026-10): the main-site model price is the sub-site wholesale
// cost; a sub-site defaults to the same price and may charge more. Settlement
// is derived from actual API usage:
//
//	RevenueAmount = what the user was charged at the sub-site price
//	CostAmount    = the main-site wholesale cost for the same usage
//	MarginAmount  = RevenueAmount - CostAmount (what the sub-site is owed)
//
// Amounts are in the usage-billing currency (USD), never in a recharge gateway
// currency such as CNY.
type UsageSettlement struct {
	SubsiteID      int64
	UserID         int64
	UsageRequestID string
	BillingMode    string
	Currency       string
	RevenueAmount  float64
	CostAmount     float64
}

// MarginAmount is the amount owed to the sub-site.
func (u UsageSettlement) MarginAmount() float64 {
	return u.RevenueAmount - u.CostAmount
}

// RecordUsageSettlement records a ledger row for one billed usage request using
// the repository DB. Main-site usage (SubsiteID <= 0) is a no-op.
func (r *Repository) RecordUsageSettlement(ctx context.Context, entry UsageSettlement) error {
	if entry.SubsiteID <= 0 {
		return nil
	}
	if r == nil || r.db == nil {
		return errors.New("downstream: nil repository database")
	}
	return RecordUsageSettlement(ctx, r.db, entry)
}

// RecordUsageSettlement records a payable ledger entry for one billed API usage
// request. Main-site usage (SubsiteID <= 0) is a no-op so the main-site billing
// path never gains a ledger side effect.
//
// Idempotency: the (subsite_id, usage_request_id) unique index makes retries
// safe. On conflict the existing row is loaded and its identity/amounts are
// verified; a mismatch is returned as an error instead of being silently
// ignored, so a retry can never leave a wrong historical row in place.
func RecordUsageSettlement(ctx context.Context, exec SQLExecutor, entry UsageSettlement) error {
	if entry.SubsiteID <= 0 {
		return nil
	}
	if exec == nil {
		return errors.New("downstream: nil settlement executor")
	}
	requestID := strings.TrimSpace(entry.UsageRequestID)
	if requestID == "" {
		return errors.New("downstream: invalid usage settlement request id")
	}
	if !isFiniteAmount(entry.RevenueAmount) || !isFiniteAmount(entry.CostAmount) {
		return errors.New("downstream: invalid usage settlement amount")
	}
	currency := normalizeSettlementCurrency(entry.Currency)
	billingMode := strings.TrimSpace(entry.BillingMode)
	margin := entry.MarginAmount()

	executor, hasReturning := exec.(settlementRowQuerier)
	if hasReturning {
		var id int64
		err := executor.QueryRowContext(ctx, `
			INSERT INTO settlement_ledger (
				subsite_id, user_id, usage_request_id, billing_mode, currency,
				gross_amount, cost_amount, margin_amount, status, notes
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (subsite_id, usage_request_id) WHERE usage_request_id IS NOT NULL DO NOTHING
			RETURNING id
		`,
			entry.SubsiteID,
			nullableInt64(entry.UserID),
			requestID,
			nullableString(billingMode),
			currency,
			entry.RevenueAmount,
			entry.CostAmount,
			margin,
			SettlementStatusPending,
			"",
		).Scan(&id)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		return verifyExistingUsageSettlement(ctx, executor, entry, currency, billingMode, margin)
	}

	_, err := exec.ExecContext(ctx, `
		INSERT INTO settlement_ledger (
			subsite_id, user_id, usage_request_id, billing_mode, currency,
			gross_amount, cost_amount, margin_amount, status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (subsite_id, usage_request_id) WHERE usage_request_id IS NOT NULL DO NOTHING
	`,
		entry.SubsiteID,
		nullableInt64(entry.UserID),
		requestID,
		nullableString(billingMode),
		currency,
		entry.RevenueAmount,
		entry.CostAmount,
		margin,
		SettlementStatusPending,
		"",
	)
	return err
}

// settlementRowQuerier is satisfied by *sql.DB and *sql.Tx, which the settlement
// writer prefers so it can detect an idempotency conflict and verify the
// existing row. Exec-only adapters fall back to the plain idempotent insert.
type settlementRowQuerier interface {
	SQLExecutor
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func verifyExistingUsageSettlement(
	ctx context.Context,
	exec settlementRowQuerier,
	entry UsageSettlement,
	currency, billingMode string,
	margin float64,
) error {
	var (
		existingSubsite int64
		existingUser    sql.NullInt64
		existingRevenue float64
		existingCost    float64
		existingMargin  float64
		existingMode    sql.NullString
		existingCurrency string
		existingStatus  string
	)
	err := exec.QueryRowContext(ctx, `
		SELECT subsite_id, user_id, gross_amount, cost_amount, margin_amount,
		       billing_mode, currency, status
		FROM settlement_ledger
		WHERE subsite_id = $1 AND usage_request_id = $2
	`, entry.SubsiteID, strings.TrimSpace(entry.UsageRequestID)).Scan(
		&existingSubsite,
		&existingUser,
		&existingRevenue,
		&existingCost,
		&existingMargin,
		&existingMode,
		&existingCurrency,
		&existingStatus,
	)
	if err != nil {
		return fmt.Errorf("downstream: verify existing usage settlement: %w", err)
	}
	if !settlementRowMatches(existingSubsite, existingUser, existingRevenue, existingCost, existingMargin, existingMode, existingCurrency, existingStatus, entry, currency, billingMode, margin) {
		return fmt.Errorf(
			"downstream: usage settlement conflict for request %q: existing row differs from new entry",
			strings.TrimSpace(entry.UsageRequestID),
		)
	}
	return nil
}

func settlementRowMatches(
	subsiteID int64,
	userID sql.NullInt64,
	revenue, cost, margin float64,
	mode sql.NullString,
	existingCurrency, existingStatus string,
	entry UsageSettlement,
	currency, billingMode string,
	expectedMargin float64,
) bool {
	if subsiteID != entry.SubsiteID {
		return false
	}
	if entry.UserID > 0 && (!userID.Valid || userID.Int64 != entry.UserID) {
		return false
	}
	if !amountsEqual(revenue, entry.RevenueAmount) ||
		!amountsEqual(cost, entry.CostAmount) ||
		!amountsEqual(margin, expectedMargin) {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(existingCurrency), strings.TrimSpace(currency)) {
		return false
	}
	if strings.TrimSpace(existingStatus) == SettlementStatusVoid {
		return false
	}
	if billingMode != "" {
		if !mode.Valid || strings.TrimSpace(mode.String) != billingMode {
			return false
		}
	}
	return true
}

func amountsEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func normalizeSettlementCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		// Usage is billed in USD; never inherit the recharge gateway currency.
		return "USD"
	}
	return currency
}

func isFiniteAmount(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func nullableInt64(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
