//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// V1 keeps the settlement ledger usage-based. Refunding a recharge order must
// not touch the ledger (there is no order-based settlement row to reverse), and
// it must not require the settlement_ledger table to exist. This test runs the
// refund completion path against a client without that table: any settlement
// write added to refunds would fail here.
func TestRefundCompletionDoesNotWriteUsageSettlement(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	order := createPaymentFulfillmentSubscriptionOrder(
		t,
		ctx,
		client,
		OrderStatusCompleted,
		time.Now().Add(-time.Minute),
	)

	svc := &PaymentService{entClient: client}
	result, err := svc.markRefundOk(ctx, &RefundPlan{
		OrderID:      order.ID,
		Order:        order,
		RefundAmount: order.Amount,
		Reason:       "test refund",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.Success)

	updated, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, updated.Status)
	require.InDelta(t, order.Amount, updated.RefundAmount, 1e-9)

	// The settlement table intentionally does not exist in this client: a
	// refund that tried to write a ledger row would have failed above.
	rows, err := client.QueryContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE name = 'settlement_ledger'")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	require.True(t, rows.Next())
	var count int
	require.NoError(t, rows.Scan(&count))
	require.NoError(t, rows.Err())
	require.Equal(t, 0, count)
}
