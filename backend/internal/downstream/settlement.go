package downstream

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
)

// ProviderSnapshotSubsiteIDKey stores the resolved subsite on the payment order
// snapshot so webhook fulfillment can create settlement rows without coupling
// the main payment service to downstream tables.
const ProviderSnapshotSubsiteIDKey = "downstream_subsite_id"

// SettlementOrder is the immutable order snapshot used to create a ledger row.
// GrossAmount is what the main site collected; CostAmount is the platform cost
// attributed to the order.
type SettlementOrder struct {
	ID          int64
	SubsiteID   int64
	UserID      int64
	Currency    string
	GrossAmount float64
	CostAmount  float64
}

// RecordSettlementForOrder records a settlement entry using the repository DB.
func (r *Repository) RecordSettlementForOrder(ctx context.Context, order SettlementOrder) error {
	if r == nil || r.db == nil {
		if order.SubsiteID <= 0 {
			return nil
		}
		return errors.New("downstream: nil repository database")
	}
	return RecordSettlementForOrder(ctx, r.db, order)
}

// SubsiteIDFromProviderSnapshot returns the subsite ID stored on the payment
// order snapshot. Main-site orders have no marker and return false.
func SubsiteIDFromProviderSnapshot(snapshot map[string]any) (int64, bool) {
	if len(snapshot) == 0 {
		return 0, false
	}
	switch value := snapshot[ProviderSnapshotSubsiteIDKey].(type) {
	case int64:
		return positiveInt64(value)
	case int:
		return positiveInt64(int64(value))
	case float64:
		return positiveInt64(int64(value))
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return 0, false
		}
		return positiveInt64(parsed)
	default:
		return 0, false
	}
}

// RecordSettlementForOrder records a payable ledger entry for one successful
// downstream recharge order. Main-site orders (SubsiteID <= 0) are a no-op.
// The ORDER_ID uniqueness constraint makes repeated webhook fulfillment safe.
func RecordSettlementForOrder(ctx context.Context, exec SQLExecutor, order SettlementOrder) error {
	if order.SubsiteID <= 0 {
		return nil
	}
	if exec == nil {
		return errors.New("downstream: nil settlement executor")
	}
	if order.ID <= 0 {
		return errors.New("downstream: invalid settlement order id")
	}
	if !isFiniteAmount(order.GrossAmount) || !isFiniteAmount(order.CostAmount) {
		return errors.New("downstream: invalid settlement amount")
	}
	currency := strings.TrimSpace(order.Currency)
	if currency == "" {
		currency = "CNY"
	}
	margin := order.GrossAmount - order.CostAmount
	_, err := exec.ExecContext(ctx, `
		INSERT INTO settlement_ledger (
			subsite_id, order_id, user_id, currency,
			gross_amount, cost_amount, margin_amount, status, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (order_id) WHERE order_id IS NOT NULL DO NOTHING
	`,
		order.SubsiteID,
		order.ID,
		order.UserID,
		currency,
		order.GrossAmount,
		order.CostAmount,
		margin,
		SettlementStatusPending,
		"",
	)
	return err
}

func isFiniteAmount(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func positiveInt64(value int64) (int64, bool) {
	if value <= 0 {
		return 0, false
	}
	return value, true
}
