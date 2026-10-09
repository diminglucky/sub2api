package downstream

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestRecordSettlementForOrderInsertsPendingLedgerEntry(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectExec(`(?s)INSERT INTO settlement_ledger`).
		WithArgs(
			int64(7),
			int64(295),
			int64(42),
			"CNY",
			105.0,
			100.0,
			5.0,
			SettlementStatusPending,
			"",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = RecordSettlementForOrder(context.Background(), db, SettlementOrder{
		ID:          295,
		SubsiteID:   7,
		UserID:      42,
		Currency:    "CNY",
		GrossAmount: 105,
		CostAmount:  100,
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordSettlementForOrderSkipsMainSiteOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	err = RecordSettlementForOrder(context.Background(), db, SettlementOrder{
		ID:          295,
		SubsiteID:   0,
		GrossAmount: 105,
		CostAmount:  100,
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordSettlementForOrderUsesIdempotentInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectExec(`(?s)ON CONFLICT \(order_id\).*DO NOTHING`).
		WithArgs(
			int64(7),
			int64(295),
			int64(42),
			"USD",
			5.0,
			10.0,
			-5.0,
			SettlementStatusPending,
			"",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = RecordSettlementForOrder(context.Background(), db, SettlementOrder{
		ID:          295,
		SubsiteID:   7,
		UserID:      42,
		Currency:    "USD",
		GrossAmount: 5,
		CostAmount:  10,
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
