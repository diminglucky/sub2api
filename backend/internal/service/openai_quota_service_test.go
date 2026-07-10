package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

func TestOpenAIQuotaServiceQueryUsage(t *testing.T) {
	var gotAuth, gotAccountID, gotOriginator, gotProxyURL string
	upstream := newLocalHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		switch r.URL.Path {
		case "/backend-api/wham/usage":
			gotAuth = r.Header.Get("authorization")
			gotAccountID = r.Header.Get("chatgpt-account-id")
			gotOriginator = r.Header.Get("originator")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"user_id":"u1","account_id":"acct_1","rate_limit_reset_credits":{"available_count":2}}`))
		case "/backend-api/wham/rate-limit-reset-credits":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"rate_limit_reset_credits":[{"expires_at":"2026-07-09T00:00:00Z"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	expiresAt := time.Now().Add(time.Hour).Format(time.RFC3339)
	proxyID := int64(9)
	svc := NewOpenAIQuotaService(
		openAIQuotaAccountRepoStub{account: &Account{
			ID:       11,
			Platform: PlatformOpenAI,
			Type:     AccountTypeOAuth,
			Credentials: map[string]any{
				"access_token":       "access-123",
				"expires_at":         expiresAt,
				"chatgpt_account_id": "acct_1",
			},
			ProxyID: &proxyID,
			Proxy: &Proxy{
				Protocol: "http",
				Host:     "127.0.0.1",
				Port:     8888,
			},
		}},
		nil,
		NewOpenAITokenProvider(nil, nil, nil),
		func(proxyURL string) (*req.Client, error) {
			gotProxyURL = proxyURL
			return newOpenAIQuotaTestClient(upstream.URL), nil
		},
	)

	got, err := svc.QueryUsage(context.Background(), 11)
	require.NoError(t, err)
	require.Equal(t, "u1", got.UserID)
	require.Equal(t, "acct_1", got.AccountID)
	require.Equal(t, 2, got.RateLimitResetCredits.AvailableCount)
	require.Equal(t, []OpenAIRateLimitResetCreditDetail{{ExpiresAt: "2026-07-09T00:00:00Z"}}, got.RateLimitResetCredits.Credits)
	require.NotZero(t, got.FetchedAt)
	require.Equal(t, "Bearer access-123", gotAuth)
	require.Equal(t, "acct_1", gotAccountID)
	require.Equal(t, "Codex Desktop", gotOriginator)
	require.Equal(t, "http://127.0.0.1:8888", gotProxyURL)
}

func TestOpenAIQuotaServiceResetCredit(t *testing.T) {
	var gotBody map[string]string
	upstream := newLocalHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/wham/rate-limit-reset-credits/consume", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("content-type"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotBody))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"ok","windows_reset":1,"credit":{"id":"credit_1","status":"redeemed"}}`))
	}))
	defer upstream.Close()

	expiresAt := time.Now().Add(time.Hour).Format(time.RFC3339)
	svc := NewOpenAIQuotaService(
		openAIQuotaAccountRepoStub{account: &Account{
			ID:       12,
			Platform: PlatformOpenAI,
			Type:     AccountTypeOAuth,
			Credentials: map[string]any{
				"access_token":    "access-456",
				"expires_at":      expiresAt,
				"organization_id": "legacy_account",
			},
		}},
		nil,
		NewOpenAITokenProvider(nil, nil, nil),
		func(proxyURL string) (*req.Client, error) {
			require.Empty(t, proxyURL)
			return newOpenAIQuotaTestClient(upstream.URL), nil
		},
	)

	got, err := svc.ResetCredit(context.Background(), 12)
	require.NoError(t, err)
	require.Equal(t, "ok", got.Code)
	require.Equal(t, 1, got.WindowsReset)
	require.NotNil(t, got.Credit)
	require.Equal(t, "credit_1", got.Credit.ID)
	require.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, gotBody["redeem_request_id"])
}

func TestOpenAIQuotaServiceValidation(t *testing.T) {
	svc := NewOpenAIQuotaService(
		openAIQuotaAccountRepoStub{account: &Account{
			ID:          13,
			Platform:    PlatformAnthropic,
			Type:        AccountTypeOAuth,
			Credentials: map[string]any{},
		}},
		nil,
		NewOpenAITokenProvider(nil, nil, nil),
		func(proxyURL string) (*req.Client, error) {
			t.Fatalf("privacy client should not be constructed for invalid account")
			return nil, nil
		},
	)

	_, err := svc.QueryUsage(context.Background(), 13)
	require.Error(t, err)
	require.Contains(t, err.Error(), "OPENAI_QUOTA_INVALID_PLATFORM")
}

func newOpenAIQuotaTestClient(baseURL string) *req.Client {
	baseURL = strings.TrimRight(baseURL, "/")
	client := req.C()
	client.GetTransport().WrapRoundTripFunc(func(rt http.RoundTripper) req.HttpRoundTripFunc {
		return func(r *http.Request) (*http.Response, error) {
			if r.URL.Host == "chatgpt.com" {
				replacement, err := http.NewRequestWithContext(r.Context(), r.Method, baseURL+r.URL.RequestURI(), r.Body)
				if err != nil {
					return nil, err
				}
				replacement.Header = r.Header.Clone()
				r = replacement
			}
			return rt.RoundTrip(r)
		}
	})
	return client
}

type openAIQuotaAccountRepoStub struct {
	AccountRepository
	account *Account
	err     error
}

func (r openAIQuotaAccountRepoStub) GetByID(ctx context.Context, id int64) (*Account, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.account, nil
}
