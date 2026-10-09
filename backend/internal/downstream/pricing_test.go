package downstream

import (
	"context"
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestRepositoryResolveSubsitePriceModelOverridePrecedence(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	input := 1.0
	output := 2.0
	explicitInput := 0.5
	mock.ExpectQuery(`(?s)FROM subsite_prices\s+WHERE subsite_id = \$1\s+AND scope = 'model'\s+AND model = \$2`).
		WithArgs(int64(7), "gpt-4o").
		WillReturnRows(sqlmock.NewRows([]string{
			"rate_multiplier", "input_price", "output_price",
			"cache_write_price", "cache_read_price", "per_request_price",
		}).AddRow(2.0, explicitInput, nil, nil, nil, nil))

	repo := NewRepository(db)
	got, err := repo.ResolveSubsitePrice(context.Background(), 7, "gpt-4o", Price{
		InputPrice:  &input,
		OutputPrice: &output,
	})
	require.NoError(t, err)
	require.NotNil(t, got.InputPrice)
	require.InDelta(t, 0.5, *got.InputPrice, 1e-12)
	require.NotNil(t, got.OutputPrice)
	require.InDelta(t, 4.0, *got.OutputPrice, 1e-12)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryResolveSubsitePriceGroupOverrideFallsBackFromModel(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	input := 1.0
	output := 2.0
	mock.ExpectQuery(`(?s)FROM subsite_prices\s+WHERE subsite_id = \$1\s+AND scope = 'model'\s+AND model = \$2`).
		WithArgs(int64(7), "gpt-4o").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`(?s)FROM subsite_prices\s+WHERE subsite_id = \$1\s+AND scope = 'group'\s+AND group_id = \$2`).
		WithArgs(int64(7), int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{
			"rate_multiplier", "input_price", "output_price",
			"cache_write_price", "cache_read_price", "per_request_price",
		}).AddRow(1.5, nil, nil, nil, nil, nil))

	repo := NewRepository(db)
	got, err := repo.ResolveSubsitePrice(context.Background(), 7, "gpt-4o", Price{
		InputPrice:  &input,
		OutputPrice: &output,
	}, 9)
	require.NoError(t, err)
	require.InDelta(t, 1.5, *got.InputPrice, 1e-12)
	require.InDelta(t, 3.0, *got.OutputPrice, 1e-12)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryResolveSubsitePriceFallsBackToBaseForMainSite(t *testing.T) {
	input := 1.25
	base := Price{InputPrice: &input}

	repo := NewRepository(nil)
	got, err := repo.ResolveSubsitePrice(context.Background(), 0, "gpt-4o", base)
	require.NoError(t, err)
	require.NotNil(t, got.InputPrice)
	require.InDelta(t, 1.25, *got.InputPrice, 1e-12)
}
