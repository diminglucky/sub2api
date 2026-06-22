//go:build unit

package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type tokenRefreshCandidateRepo struct {
	AccountRepository
	accounts             []Account
	updatedCredentialIDs []int64
	listActiveCalls      int
}

func (r *tokenRefreshCandidateRepo) ListActive(context.Context) ([]Account, error) {
	r.listActiveCalls++
	return r.accounts, nil
}

func (r *tokenRefreshCandidateRepo) ListOAuthRefreshCandidates(context.Context) ([]Account, error) {
	candidates := make([]Account, 0, len(r.accounts))
	now := time.Now()
	for _, account := range r.accounts {
		refreshToken, _ := account.Credentials["refresh_token"].(string)
		inRetryCooldown := account.TempUnschedulableUntil != nil &&
			account.TempUnschedulableUntil.After(now) &&
			strings.HasPrefix(account.TempUnschedulableReason, "token refresh retry exhausted:")
		if account.Status != StatusActive ||
			account.Type != AccountTypeOAuth ||
			!isOAuthRefreshPlatform(account.Platform) ||
			strings.TrimSpace(refreshToken) == "" ||
			inRetryCooldown {
			continue
		}
		candidates = append(candidates, account)
	}
	return candidates, nil
}

func (r *tokenRefreshCandidateRepo) UpdateCredentials(_ context.Context, id int64, _ map[string]any) error {
	r.updatedCredentialIDs = append(r.updatedCredentialIDs, id)
	return nil
}

func isOAuthRefreshPlatform(platform string) bool {
	switch platform {
	case PlatformAnthropic, PlatformOpenAI, PlatformGemini, PlatformAntigravity:
		return true
	default:
		return false
	}
}

type tokenRefreshTestRefresher struct{}

func (r *tokenRefreshTestRefresher) CanRefresh(*Account) bool { return true }

func (r *tokenRefreshTestRefresher) NeedsRefresh(*Account, time.Duration) bool { return true }

func (r *tokenRefreshTestRefresher) Refresh(context.Context, *Account) (map[string]any, error) {
	return map[string]any{"access_token": "new-access-token", "refresh_token": "new-refresh-token"}, nil
}

func TestTokenRefreshService_ProcessRefreshUsesOAuthRefreshCandidates(t *testing.T) {
	future := time.Now().Add(10 * time.Minute)
	repo := &tokenRefreshCandidateRepo{
		accounts: []Account{
			{
				ID:          1,
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Credentials: map[string]any{"refresh_token": "refresh-token"},
			},
			{
				ID:          2,
				Platform:    PlatformOpenAI,
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Credentials: map[string]any{},
			},
			{
				ID:          3,
				Platform:    PlatformGemini,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Credentials: map[string]any{"refresh_token": "refresh-token"},
			},
			{
				ID:                      4,
				Platform:                PlatformAntigravity,
				Type:                    AccountTypeOAuth,
				Status:                  StatusActive,
				Credentials:             map[string]any{"refresh_token": "refresh-token"},
				TempUnschedulableUntil:  &future,
				TempUnschedulableReason: "token refresh retry exhausted: network timeout",
			},
			{
				ID:          5,
				Platform:    "other",
				Type:        AccountTypeOAuth,
				Status:      StatusActive,
				Credentials: map[string]any{"refresh_token": "refresh-token"},
			},
		},
	}
	svc := &TokenRefreshService{
		accountRepo:   repo,
		refreshers:    []TokenRefresher{&tokenRefreshTestRefresher{}},
		refreshPolicy: DefaultBackgroundRefreshPolicy(),
		cfg:           &config.TokenRefreshConfig{RefreshBeforeExpiryHours: 1, MaxRetries: 1},
	}

	svc.processRefresh()

	require.Zero(t, repo.listActiveCalls, "TokenRefreshService should not use the broad active-account query")
	require.Equal(t, []int64{1}, repo.updatedCredentialIDs)
}
