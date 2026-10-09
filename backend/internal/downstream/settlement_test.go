package downstream

import (
	"context"
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// expectUsageSettlementInsert matches the idempotent insert. The margin is read
// from the same runtime expression the writer uses so float rounding matches.
func expectUsageSettlementInsert(mock sqlmock.Sqlmock, entry UsageSettlement, currency string) *sqlmock.ExpectedQuery {
	return mock.ExpectQuery(`(?s)INSERT INTO settlement_ledger`).WithArgs(
		entry.SubsiteID,
		nullableInt64(entry.UserID),
		entry.UsageRequestID,
		nullableString(entry.BillingMode),
		currency,
		entry.RevenueAmount,
		entry.CostAmount,
		entry.MarginAmount(),
		SettlementStatusPending,
		"",
	)
}

func existingSettlementRows(subsiteID, userID int64, revenue, cost float64, mode, currency, status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"subsite_id", "user_id", "gross_amount", "cost_amount",
		"margin_amount", "billing_mode", "currency", "status",
	}).AddRow(subsiteID, userID, revenue, cost, revenue-cost, mode, currency, status)
}

func TestRecordUsageSettlementInsertsPendingLedgerEntry(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	entry := UsageSettlement{
		SubsiteID:      7,
		UserID:         42,
		UsageRequestID: "req-abc",
		BillingMode:    "token",
		RevenueAmount:  1.20,
		CostAmount:     0.90,
	}
	expectUsageSettlementInsert(mock, entry, "USD").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	require.NoError(t, RecordUsageSettlement(context.Background(), db, entry))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordUsageSettlementDefaultsCurrencyToUSD(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	entry := UsageSettlement{
		SubsiteID:      7,
		UsageRequestID: "req-no-user",
		RevenueAmount:  0.5,
		CostAmount:     0.4,
	}
	expectUsageSettlementInsert(mock, entry, "USD").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(2)))

	require.NoError(t, RecordUsageSettlement(context.Background(), db, entry))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordUsageSettlementSkipsMainSiteUsage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	err = RecordUsageSettlement(context.Background(), db, UsageSettlement{
		SubsiteID:      0,
		UsageRequestID: "req-main",
		RevenueAmount:  1,
		CostAmount:     1,
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordUsageSettlementRejectsEmptyRequestID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	err = RecordUsageSettlement(context.Background(), db, UsageSettlement{
		SubsiteID:     7,
		RevenueAmount: 1,
		CostAmount:    1,
	})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordUsageSettlementVerifiesExistingRowOnConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	entry := UsageSettlement{
		SubsiteID:      7,
		UserID:         42,
		UsageRequestID: "req-abc",
		BillingMode:    "token",
		RevenueAmount:  1.20,
		CostAmount:     0.90,
	}
	// INSERT ... ON CONFLICT DO NOTHING returns no row on a replay.
	expectUsageSettlementInsert(mock, entry, "USD").WillReturnError(sql.ErrNoRows)
	// Existing identical row: replay must be a no-op.
	mock.ExpectQuery(`(?s)FROM settlement_ledger\s+WHERE subsite_id = \$1 AND usage_request_id = \$2`).
		WithArgs(int64(7), "req-abc").
		WillReturnRows(existingSettlementRows(7, 42, entry.RevenueAmount, entry.CostAmount, "token", "USD", SettlementStatusPending))

	require.NoError(t, RecordUsageSettlement(context.Background(), db, entry))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordUsageSettlementRejectsConflictingExistingRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	entry := UsageSettlement{
		SubsiteID:      7,
		UserID:         42,
		UsageRequestID: "req-abc",
		BillingMode:    "token",
		RevenueAmount:  1.20,
		CostAmount:     0.90,
	}
	expectUsageSettlementInsert(mock, entry, "USD").WillReturnError(sql.ErrNoRows)
	// Existing row disagrees on the charged revenue: must surface an error.
	mock.ExpectQuery(`(?s)FROM settlement_ledger\s+WHERE subsite_id = \$1 AND usage_request_id = \$2`).
		WithArgs(int64(7), "req-abc").
		WillReturnRows(existingSettlementRows(7, 42, 9.99, entry.CostAmount, "token", "USD", SettlementStatusPending))

	require.Error(t, RecordUsageSettlement(context.Background(), db, entry))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordUsageSettlementRejectsVoidedExistingRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	entry := UsageSettlement{
		SubsiteID:      7,
		UserID:         42,
		UsageRequestID: "req-abc",
		BillingMode:    "token",
		RevenueAmount:  1.20,
		CostAmount:     0.90,
	}
	expectUsageSettlementInsert(mock, entry, "USD").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`(?s)FROM settlement_ledger\s+WHERE subsite_id = \$1 AND usage_request_id = \$2`).
		WithArgs(int64(7), "req-abc").
		WillReturnRows(existingSettlementRows(7, 42, entry.RevenueAmount, entry.CostAmount, "token", "USD", SettlementStatusVoid))

	require.Error(t, RecordUsageSettlement(context.Background(), db, entry))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageSettlementMarginAmount(t *testing.T) {
	entry := UsageSettlement{RevenueAmount: 1.5, CostAmount: 1.0}
	require.InDelta(t, 0.5, entry.MarginAmount(), 1e-12)
}
