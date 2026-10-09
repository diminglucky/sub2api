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
