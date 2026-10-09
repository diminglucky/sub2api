package downstream

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// SQLExecutor is the small database surface needed for subsite attribution.
// *sql.DB, *sql.Tx, and ent clients all satisfy it.
type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// SubsiteIDFromContext returns the resolved subsite ID for the current request.
// Existing main-site requests return false and must keep writing NULL.
func SubsiteIDFromContext(ctx context.Context) (int64, bool) {
	subsite, ok := FromContext(ctx)
	if !ok || subsite == nil || subsite.ID <= 0 {
		return 0, false
	}
	return subsite.ID, true
}

// RecordSubsiteMember links a main-site user to the current subsite.
func RecordSubsiteMember(ctx context.Context, exec SQLExecutor, subsiteID, userID int64, source string) error {
	if exec == nil {
		return errors.New("downstream: nil SQL executor")
	}
	if subsiteID <= 0 || userID <= 0 {
		return errors.New("downstream: invalid subsite member attribution")
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = SubsiteMemberSourceRegistration
	}
	_, err := exec.ExecContext(ctx, `
		INSERT INTO subsite_members (subsite_id, user_id, role, source)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (subsite_id, user_id) DO NOTHING
	`, subsiteID, userID, SubsiteMemberRoleMember, source)
	return err
}

// AttributePaymentOrder stamps a newly-created recharge order.
func AttributePaymentOrder(ctx context.Context, exec SQLExecutor, subsiteID, orderID int64) error {
	if exec == nil {
		return errors.New("downstream: nil SQL executor")
	}
	if subsiteID <= 0 || orderID <= 0 {
		return errors.New("downstream: invalid payment order attribution")
	}
	_, err := exec.ExecContext(ctx,
		`UPDATE payment_orders SET subsite_id = $1 WHERE id = $2`,
		subsiteID, orderID,
	)
	return err
}

// AttributeAPIKey stamps a newly-created API key.
func AttributeAPIKey(ctx context.Context, exec SQLExecutor, subsiteID, keyID int64) error {
	if exec == nil {
		return errors.New("downstream: nil SQL executor")
	}
	if subsiteID <= 0 || keyID <= 0 {
		return errors.New("downstream: invalid API key attribution")
	}
	_, err := exec.ExecContext(ctx,
		`UPDATE api_keys SET subsite_id = $1 WHERE id = $2`,
		subsiteID, keyID,
	)
	return err
}
