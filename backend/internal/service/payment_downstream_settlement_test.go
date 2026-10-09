//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/downstream"
	"github.com/stretchr/testify/require"
)

func TestMarkCompletedWithSubsiteSettlementWritesLedgerAtomically(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	_, err := client.ExecContext(ctx, `
		CREATE TABLE settlement_ledger (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			subsite_id    INTEGER NOT NULL,
			order_id      INTEGER,
			user_id       INTEGER,
			currency      TEXT NOT NULL DEFAULT 'CNY',
			gross_amount  REAL NOT NULL DEFAULT 0,
			cost_amount   REAL NOT NULL DEFAULT 0,
			margin_amount REAL NOT NULL DEFAULT 0,
			status        TEXT NOT NULL DEFAULT 'pending',
			notes         TEXT NOT NULL DEFAULT '',
			settled_at    DATETIME,
			created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	require.NoError(t, err)
	_, err = client.ExecContext(ctx, `
		CREATE UNIQUE INDEX idx_settlement_ledger_order_id
		ON settlement_ledger(order_id)
		WHERE order_id IS NOT NULL`)
	require.NoError(t, err)

	order := createPaymentFulfillmentSubscriptionOrder(
		t,
		ctx,
		client,
		OrderStatusPaid,
		time.Now().Add(-time.Minute),
	)
	_, err = client.PaymentOrder.UpdateOneID(order.ID).
		SetAmount(100).
		SetPayAmount(105).
		SetProviderSnapshot(map[string]any{
			downstream.ProviderSnapshotSubsiteIDKey: int64(7),
		}).
		Save(ctx)
	require.NoError(t, err)
	order, err = client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)

	svc := &PaymentService{entClient: client}
	lease, err := svc.acquirePaymentFulfillmentLease(ctx, order)
	require.NoError(t, err)
	require.NotNil(t, lease)
	err = svc.markCompleted(ctx, order, lease, "SUBSCRIPTION_SUCCESS")
	require.NoError(t, err)

	rows, err := client.QueryContext(ctx, `
		SELECT subsite_id, order_id, user_id, currency,
		       gross_amount, cost_amount, margin_amount, status
		FROM settlement_ledger
		WHERE order_id = $1`, order.ID)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	require.True(t, rows.Next())
	var (
		subsiteID int64
		orderID   int64
		userID    int64
		currency  string
		gross     float64
		cost      float64
		margin    float64
		status    string
	)
	require.NoError(t, rows.Scan(&subsiteID, &orderID, &userID, &currency, &gross, &cost, &margin, &status))
	require.NoError(t, rows.Err())
	require.Equal(t, int64(7), subsiteID)
	require.Equal(t, order.ID, orderID)
	require.Equal(t, order.UserID, userID)
	require.Equal(t, "CNY", currency)
	require.InDelta(t, 105, gross, 1e-9)
	require.InDelta(t, 100, cost, 1e-9)
	require.InDelta(t, 5, margin, 1e-9)
	require.Equal(t, downstream.SettlementStatusPending, status)
}
