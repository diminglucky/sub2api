package downstream

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
)

const SubsiteBaseDomain = "superai.sbs"

var subsiteSlugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

// SubsiteUpsert is the admin-facing request shape for creating or updating a
// subsite. Domain is optional on create and defaults to <slug>.superai.sbs.
type SubsiteUpsert struct {
	Slug        string
	Domain      string
	Name        string
	LogoURL     string
	ThemeColor  string
	Status      string
	AdminUserID *int64
}

// PriceOverrideUpsert is the admin-facing request shape for a subsite price
// override. RateMultiplier nil means "keep default 1.0".
type PriceOverrideUpsert struct {
	Scope           string
	GroupID         *int64
	Model           *string
	RateMultiplier  *float64
	InputPrice      *float64
	OutputPrice     *float64
	CacheWritePrice *float64
	CacheReadPrice  *float64
	PerRequestPrice *float64
	ImagePrice1K    *float64
	ImagePrice2K    *float64
	ImagePrice4K    *float64
	Status          string
}

func NormalizeSubsiteUpsert(input SubsiteUpsert) (SubsiteUpsert, error) {
	input.Slug = NormalizeSlug(input.Slug)
	input.Domain = strings.ToLower(strings.TrimSpace(input.Domain))
	input.Name = strings.TrimSpace(input.Name)
	input.LogoURL = strings.TrimSpace(input.LogoURL)
	input.ThemeColor = strings.TrimSpace(input.ThemeColor)
	input.Status = normalizeSubsiteStatus(input.Status)
	if input.Status != SubsiteStatusActive && input.Status != SubsiteStatusDisabled {
		return SubsiteUpsert{}, fmt.Errorf("invalid subsite status")
	}

	if !subsiteSlugPattern.MatchString(input.Slug) {
		return SubsiteUpsert{}, fmt.Errorf("invalid subsite slug")
	}
	if input.Name == "" {
		return SubsiteUpsert{}, fmt.Errorf("subsite name is required")
	}
	if input.Domain == "" {
		input.Domain = input.Slug + "." + SubsiteBaseDomain
	}
	if input.Domain != input.Slug+"."+SubsiteBaseDomain {
		return SubsiteUpsert{}, fmt.Errorf("subsite domain must be %s.%s", input.Slug, SubsiteBaseDomain)
	}
	if input.AdminUserID != nil && *input.AdminUserID <= 0 {
		return SubsiteUpsert{}, fmt.Errorf("admin_user_id must be positive")
	}
	return input, nil
}

func NormalizePriceOverrideUpsert(input PriceOverrideUpsert) (PriceOverrideUpsert, error) {
	input.Scope = strings.ToLower(strings.TrimSpace(input.Scope))
	input.Status = normalizePriceStatus(input.Status)
	if input.Status != SubsiteStatusActive && input.Status != SubsiteStatusDisabled {
		return PriceOverrideUpsert{}, fmt.Errorf("invalid price status")
	}
	if input.Model != nil {
		model := strings.TrimSpace(*input.Model)
		if model == "" {
			input.Model = nil
		} else {
			input.Model = &model
		}
	}

	switch input.Scope {
	case PriceOverrideScopeGroup:
		if input.GroupID == nil || *input.GroupID <= 0 || input.Model != nil {
			return PriceOverrideUpsert{}, fmt.Errorf("group override requires group_id and no model")
		}
	case PriceOverrideScopeModel:
		if input.Model == nil || input.GroupID != nil {
			return PriceOverrideUpsert{}, fmt.Errorf("model override requires model and no group_id")
		}
	default:
		return PriceOverrideUpsert{}, fmt.Errorf("unsupported price override scope")
	}

	for _, value := range []*float64{
		input.RateMultiplier,
		input.InputPrice,
		input.OutputPrice,
		input.CacheWritePrice,
		input.CacheReadPrice,
		input.PerRequestPrice,
		input.ImagePrice1K,
		input.ImagePrice2K,
		input.ImagePrice4K,
	} {
		if value != nil && (math.IsNaN(*value) || math.IsInf(*value, 0) || *value < 0) {
			return PriceOverrideUpsert{}, fmt.Errorf("price values must be non-negative")
		}
	}
	return input, nil
}

// ListSubsites returns all subsites for the main-site admin control plane.
func (r *Repository) ListSubsites(ctx context.Context) ([]Subsite, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+subsiteColumns+` FROM subsites ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("downstream: list subsites: %w", err)
	}
	defer rows.Close()

	subsites := make([]Subsite, 0)
	for rows.Next() {
		subsite, err := scanSubsite(rows)
		if err != nil {
			return nil, err
		}
		subsites = append(subsites, *subsite)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("downstream: iterate subsites: %w", err)
	}
	return subsites, nil
}

// GetSubsiteByID returns one subsite for admin editing.
func (r *Repository) GetSubsiteByID(ctx context.Context, id int64) (*Subsite, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if id <= 0 {
		return nil, ErrSubsiteNotFound
	}
	row := r.db.QueryRowContext(ctx, `SELECT `+subsiteColumns+` FROM subsites WHERE id = $1`, id)
	return scanSubsite(row)
}

// CreateSubsite creates one pilot or production subsite.
func (r *Repository) CreateSubsite(ctx context.Context, input SubsiteUpsert) (*Subsite, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	normalized, err := NormalizeSubsiteUpsert(input)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO subsites (slug, domain, name, logo_url, theme_color, status, admin_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+subsiteColumns,
		normalized.Slug,
		normalized.Domain,
		normalized.Name,
		normalized.LogoURL,
		normalized.ThemeColor,
		normalized.Status,
		nullableInt64Ptr(normalized.AdminUserID),
	)
	return scanSubsite(row)
}

// UpdateSubsite updates all editable fields.
func (r *Repository) UpdateSubsite(ctx context.Context, id int64, input SubsiteUpsert) (*Subsite, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if id <= 0 {
		return nil, ErrSubsiteNotFound
	}
	normalized, err := NormalizeSubsiteUpsert(input)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRowContext(ctx, `
		UPDATE subsites
		SET slug = $2,
		    domain = $3,
		    name = $4,
		    logo_url = $5,
		    theme_color = $6,
		    status = $7,
		    admin_user_id = $8,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING `+subsiteColumns,
		id,
		normalized.Slug,
		normalized.Domain,
		normalized.Name,
		normalized.LogoURL,
		normalized.ThemeColor,
		normalized.Status,
		nullableInt64Ptr(normalized.AdminUserID),
	)
	return scanSubsite(row)
}

// DisableSubsite disables a subsite without deleting historical data.
func (r *Repository) DisableSubsite(ctx context.Context, id int64) (*Subsite, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if id <= 0 {
		return nil, ErrSubsiteNotFound
	}
	row := r.db.QueryRowContext(ctx, `
		UPDATE subsites
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING `+subsiteColumns,
		id,
		SubsiteStatusDisabled,
	)
	return scanSubsite(row)
}

const priceOverrideColumns = `
	id, subsite_id, scope, group_id, model, rate_multiplier,
	input_price, output_price, cache_write_price, cache_read_price,
	per_request_price, image_price_1k, image_price_2k, image_price_4k,
	status, created_at, updated_at`

// ListSubsitePrices returns only price overrides owned by subsiteID.
func (r *Repository) ListSubsitePrices(ctx context.Context, subsiteID int64) ([]PriceOverride, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if subsiteID <= 0 {
		return nil, ErrSubsiteNotFound
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+priceOverrideColumns+`
		FROM subsite_prices
		WHERE subsite_id = $1
		ORDER BY id DESC`,
		subsiteID,
	)
	if err != nil {
		return nil, fmt.Errorf("downstream: list subsite prices: %w", err)
	}
	defer rows.Close()

	prices := make([]PriceOverride, 0)
	for rows.Next() {
		price, err := scanPriceOverride(rows)
		if err != nil {
			return nil, err
		}
		prices = append(prices, *price)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("downstream: iterate subsite prices: %w", err)
	}
	return prices, nil
}

// CreateSubsitePrice creates one price override under the requested subsite.
func (r *Repository) CreateSubsitePrice(ctx context.Context, subsiteID int64, input PriceOverrideUpsert) (*PriceOverride, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if subsiteID <= 0 {
		return nil, ErrSubsiteNotFound
	}
	normalized, err := NormalizePriceOverrideUpsert(input)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO subsite_prices (
			subsite_id, scope, group_id, model, rate_multiplier,
			input_price, output_price, cache_write_price, cache_read_price,
			per_request_price, image_price_1k, image_price_2k, image_price_4k, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING `+priceOverrideColumns,
		subsiteID,
		normalized.Scope,
		nullableInt64Ptr(normalized.GroupID),
		nullableStringPtr(normalized.Model),
		rateMultiplierOrDefault(normalized.RateMultiplier),
		nullableFloat64Ptr(normalized.InputPrice),
		nullableFloat64Ptr(normalized.OutputPrice),
		nullableFloat64Ptr(normalized.CacheWritePrice),
		nullableFloat64Ptr(normalized.CacheReadPrice),
		nullableFloat64Ptr(normalized.PerRequestPrice),
		nullableFloat64Ptr(normalized.ImagePrice1K),
		nullableFloat64Ptr(normalized.ImagePrice2K),
		nullableFloat64Ptr(normalized.ImagePrice4K),
		normalized.Status,
	)
	return scanPriceOverride(row)
}

// UpdateSubsitePrice replaces one price override after verifying subsite scope.
func (r *Repository) UpdateSubsitePrice(ctx context.Context, subsiteID, priceID int64, input PriceOverrideUpsert) (*PriceOverride, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("downstream: nil repository database")
	}
	if subsiteID <= 0 || priceID <= 0 {
		return nil, ErrSubsiteNotFound
	}
	normalized, err := NormalizePriceOverrideUpsert(input)
	if err != nil {
		return nil, err
	}
	row := r.db.QueryRowContext(ctx, `
		UPDATE subsite_prices
		SET scope = $3,
		    group_id = $4,
		    model = $5,
		    rate_multiplier = $6,
		    input_price = $7,
		    output_price = $8,
		    cache_write_price = $9,
		    cache_read_price = $10,
		    per_request_price = $11,
		    image_price_1k = $12,
		    image_price_2k = $13,
		    image_price_4k = $14,
		    status = $15,
		    updated_at = NOW()
		WHERE id = $2 AND subsite_id = $1
		RETURNING `+priceOverrideColumns,
		subsiteID,
		priceID,
		normalized.Scope,
		nullableInt64Ptr(normalized.GroupID),
		nullableStringPtr(normalized.Model),
		rateMultiplierOrDefault(normalized.RateMultiplier),
		nullableFloat64Ptr(normalized.InputPrice),
		nullableFloat64Ptr(normalized.OutputPrice),
		nullableFloat64Ptr(normalized.CacheWritePrice),
		nullableFloat64Ptr(normalized.CacheReadPrice),
		nullableFloat64Ptr(normalized.PerRequestPrice),
		nullableFloat64Ptr(normalized.ImagePrice1K),
		nullableFloat64Ptr(normalized.ImagePrice2K),
		nullableFloat64Ptr(normalized.ImagePrice4K),
		normalized.Status,
	)
	return scanPriceOverride(row)
}

// DeleteSubsiteModelPricesExcept removes model-scoped overrides whose model is
// no longer offered (used after a group is unassigned).
func (r *Repository) DeleteSubsiteModelPricesExcept(ctx context.Context, subsiteID int64, keep []string) error {
	if r == nil || r.db == nil {
		return errors.New("downstream: nil repository database")
	}
	if subsiteID <= 0 {
		return ErrSubsiteNotFound
	}
	if len(keep) == 0 {
		_, err := r.db.ExecContext(ctx,
			`DELETE FROM subsite_prices WHERE subsite_id = $1 AND scope = 'model'`,
			subsiteID,
		)
		return err
	}
	placeholders := make([]string, 0, len(keep))
	args := make([]any, 0, len(keep)+1)
	args = append(args, subsiteID)
	for i, model := range keep {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		args = append(args, model)
	}
	query := `DELETE FROM subsite_prices WHERE subsite_id = $1 AND scope = 'model' AND model NOT IN (` +
		strings.Join(placeholders, ", ") + `)`
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteSubsitePrice deletes one price override only inside subsiteID.
func (r *Repository) DeleteSubsitePrice(ctx context.Context, subsiteID, priceID int64) error {
	if r == nil || r.db == nil {
		return errors.New("downstream: nil repository database")
	}
	if subsiteID <= 0 || priceID <= 0 {
		return ErrSubsiteNotFound
	}
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM subsite_prices
		WHERE id = $2 AND subsite_id = $1`,
		subsiteID,
		priceID,
	)
	if err != nil {
		return fmt.Errorf("downstream: delete subsite price: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("downstream: subsite price delete rows: %w", err)
	}
	if affected == 0 {
		return ErrSubsiteNotFound
	}
	return nil
}

func scanPriceOverride(row rowScanner) (*PriceOverride, error) {
	var (
		price           PriceOverride
		groupID         sql.NullInt64
		model           sql.NullString
		inputPrice      sql.NullFloat64
		outputPrice     sql.NullFloat64
		cacheWritePrice sql.NullFloat64
		cacheReadPrice  sql.NullFloat64
		perRequestPrice sql.NullFloat64
		imagePrice1K    sql.NullFloat64
		imagePrice2K    sql.NullFloat64
		imagePrice4K    sql.NullFloat64
	)
	if err := row.Scan(
		&price.ID,
		&price.SubsiteID,
		&price.Scope,
		&groupID,
		&model,
		&price.RateMultiplier,
		&inputPrice,
		&outputPrice,
		&cacheWritePrice,
		&cacheReadPrice,
		&perRequestPrice,
		&imagePrice1K,
		&imagePrice2K,
		&imagePrice4K,
		&price.Status,
		&price.CreatedAt,
		&price.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubsiteNotFound
		}
		return nil, fmt.Errorf("downstream: scan subsite price: %w", err)
	}
	if groupID.Valid {
		value := groupID.Int64
		price.GroupID = &value
	}
	if model.Valid {
		value := model.String
		price.Model = &value
	}
	price.InputPrice = nullableFloat64(inputPrice)
	price.OutputPrice = nullableFloat64(outputPrice)
	price.CacheWritePrice = nullableFloat64(cacheWritePrice)
	price.CacheReadPrice = nullableFloat64(cacheReadPrice)
	price.PerRequestPrice = nullableFloat64(perRequestPrice)
	price.ImagePrice1K = nullableFloat64(imagePrice1K)
	price.ImagePrice2K = nullableFloat64(imagePrice2K)
	price.ImagePrice4K = nullableFloat64(imagePrice4K)
	return &price, nil
}

func normalizeSubsiteStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", SubsiteStatusActive:
		return SubsiteStatusActive
	case SubsiteStatusDisabled:
		return SubsiteStatusDisabled
	default:
		return strings.ToLower(strings.TrimSpace(status))
	}
}

func normalizePriceStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", SubsiteStatusActive:
		return SubsiteStatusActive
	case SubsiteStatusDisabled:
		return SubsiteStatusDisabled
	default:
		return strings.ToLower(strings.TrimSpace(status))
	}
}

func rateMultiplierOrDefault(value *float64) float64 {
	if value == nil {
		return 1
	}
	return *value
}

func nullableInt64Ptr(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableStringPtr(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableFloat64Ptr(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}
