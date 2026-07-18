package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDetectBalanceProvider(t *testing.T) {
	require.Equal(t, UpstreamBalanceProviderDeepSeek, detectBalanceProvider("https://api.deepseek.com/v1"))
	require.Equal(t, UpstreamBalanceProviderStepFun, detectBalanceProvider("https://api.stepfun.com/v1"))
	require.Equal(t, UpstreamBalanceProviderSiliconFlow, detectBalanceProvider("https://api.siliconflow.cn/v1"))
	require.Empty(t, detectBalanceProvider("https://relay.example.com/v1"))
}

func TestResolveBalanceRequestURLCustomStaysOnAccountOrigin(t *testing.T) {
	got, err := resolveBalanceRequestURL("https://relay.example.com/v1", UpstreamBalanceProviderCustom, "/api/user/balance")
	require.NoError(t, err)
	require.Equal(t, "https://relay.example.com/api/user/balance", got)

	_, err = resolveBalanceRequestURL("https://relay.example.com/v1", UpstreamBalanceProviderCustom, "https://evil.example.com/balance")
	require.Error(t, err)
	_, err = resolveBalanceRequestURL("https://relay.example.com/v1", UpstreamBalanceProviderCustom, "//evil.example.com/balance")
	require.Error(t, err)
}

func TestResolveBalanceRequestURLBuiltInRequiresMatchingAccountOrigin(t *testing.T) {
	got, err := resolveBalanceRequestURL("https://api.deepseek.com/v1", UpstreamBalanceProviderDeepSeek, "")
	require.NoError(t, err)
	require.Equal(t, "https://api.deepseek.com/user/balance", got)

	_, err = resolveBalanceRequestURL("https://relay.example.com/v1", UpstreamBalanceProviderDeepSeek, "")
	require.Error(t, err)
}

func TestResolveBalanceRequestURLNewAPIStaysOnAccountOrigin(t *testing.T) {
	got, err := resolveBalanceRequestURL("https://relay.example.com/v1", UpstreamBalanceProviderNewAPI, "")
	require.NoError(t, err)
	require.Equal(t, "https://relay.example.com/api/user/self", got)

	got, err = resolveBalanceRequestURL("https://relay.example.com/v1", UpstreamBalanceProviderSub2API, "")
	require.NoError(t, err)
	require.Equal(t, "https://relay.example.com/api/v1/user/profile", got)
}

func TestParseUpstreamBalanceAmount(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		payload  any
		path     string
		divisor  float64
		want     float64
	}{
		{
			name:     "deepseek selects cny",
			provider: UpstreamBalanceProviderDeepSeek,
			payload: map[string]any{"balance_infos": []any{
				map[string]any{"currency": "USD", "total_balance": "9.5"},
				map[string]any{"currency": "CNY", "total_balance": "88.25"},
			}},
			want: 88.25,
		},
		{
			name:     "stepfun",
			provider: UpstreamBalanceProviderStepFun,
			payload:  map[string]any{"balance": 20.5},
			want:     20.5,
		},
		{
			name:     "siliconflow",
			provider: UpstreamBalanceProviderSiliconFlow,
			payload:  map[string]any{"data": map[string]any{"totalBalance": "30.75"}},
			want:     30.75,
		},
		{
			name:     "custom path and divisor",
			provider: UpstreamBalanceProviderCustom,
			payload:  map[string]any{"data": []any{map[string]any{"remaining": 12345.0}}},
			path:     "data.0.remaining",
			divisor:  100,
			want:     123.45,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUpstreamBalanceAmount(tt.provider, tt.payload, tt.path, tt.divisor)
			require.NoError(t, err)
			require.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestParseNewAPIWalletCNYBalance(t *testing.T) {
	status := []byte(`{"success":true,"data":{"quota_display_type":"CNY","quota_per_unit":500000,"usd_exchange_rate":7.2}}`)
	profile := []byte(`{"success":true,"data":{"quota":10000000}}`)
	got, err := parseNewAPIWalletCNYBalance(status, profile)
	require.NoError(t, err)
	require.InDelta(t, 144, got, 0.0001)
}

func TestParseNewAPIWalletCNYBalanceRejectsNonCNYSite(t *testing.T) {
	status := []byte(`{"success":true,"data":{"quota_display_type":"USD","quota_per_unit":500000,"usd_exchange_rate":7.2}}`)
	profile := []byte(`{"success":true,"data":{"quota":10000000}}`)
	_, err := parseNewAPIWalletCNYBalance(status, profile)
	require.ErrorContains(t, err, "does not report balances in CNY")
}

func TestParseNewAPIWalletCNYBalanceRequiresConversionSettings(t *testing.T) {
	status := []byte(`{"success":true,"data":{"quota_display_type":"CNY"}}`)
	profile := []byte(`{"success":true,"data":{"quota":10000000}}`)
	_, err := parseNewAPIWalletCNYBalance(status, profile)
	require.ErrorContains(t, err, "quota_per_unit")
}

func TestFetchBalanceNewAPIUsesWalletLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/user/login":
			require.Equal(t, http.MethodPost, r.Method)
			w.Header().Add("Set-Cookie", "session=wallet-session; Path=/; HttpOnly")
			_, _ = w.Write([]byte(`{"success":true,"data":{"id":123}}`))
		case "/api/status":
			_, _ = w.Write([]byte(`{"success":true,"data":{"quota_display_type":"CNY","quota_per_unit":500000,"usd_exchange_rate":7.2}}`))
		case "/api/user/self":
			require.Equal(t, "123", r.Header.Get("New-Api-User"))
			cookie, err := r.Cookie("session")
			require.NoError(t, err)
			require.Equal(t, "wallet-session", cookie.Value)
			_, _ = w.Write([]byte(`{"success":true,"data":{"quota":10000000}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	account := &Account{Credentials: map[string]any{
		"api_key":                             "model-key-not-used-for-wallet",
		"base_url":                            server.URL + "/v1",
		upstreamBalanceAuthTokenCredentialKey: "password",
	}}
	service := NewUpstreamBalanceService(nil, nil, nil, nil)
	defer service.Stop()
	got, err := service.fetchBalance(context.Background(), account, UpstreamBalanceConfig{
		Provider:     UpstreamBalanceProviderNewAPI,
		AuthMode:     "login",
		AuthUsername: "wallet@example.com",
	})
	require.NoError(t, err)
	require.InDelta(t, 144, got, 0.0001)
}

func TestFetchBalanceSub2APIUsesWalletLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/auth/login":
			require.Equal(t, http.MethodPost, r.Method)
			_, _ = w.Write([]byte(`{"code":0,"message":"success","data":{"access_token":"wallet-token"}}`))
		case "/api/v1/user/profile":
			require.Equal(t, "Bearer wallet-token", r.Header.Get("Authorization"))
			_, _ = w.Write([]byte(`{"code":0,"message":"success","data":{"balance":88.25}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	account := &Account{Credentials: map[string]any{
		"api_key":                             "model-key-not-used-for-wallet",
		"base_url":                            server.URL + "/v1",
		upstreamBalanceAuthTokenCredentialKey: "password",
	}}
	service := NewUpstreamBalanceService(nil, nil, nil, nil)
	defer service.Stop()
	got, err := service.fetchBalance(context.Background(), account, UpstreamBalanceConfig{
		Provider:     UpstreamBalanceProviderSub2API,
		AuthMode:     "login",
		AuthUsername: "wallet@example.com",
	})
	require.NoError(t, err)
	require.InDelta(t, 88.25, got, 0.0001)
}

func TestValidateUpstreamBalanceConfigRequiresCustomFields(t *testing.T) {
	account := &Account{
		Type: AccountTypeUpstream,
		Credentials: map[string]any{
			"api_key":  "secret",
			"base_url": "https://relay.example.com/v1",
		},
	}
	cfg := &UpstreamBalanceConfig{
		Enabled:      true,
		Provider:     UpstreamBalanceProviderCustom,
		PlatformName: "Relay",
		Endpoint:     "/api/balance",
	}
	require.Error(t, validateUpstreamBalanceConfig(account, cfg))
	cfg.JSONPath = "data.balance"
	require.NoError(t, validateUpstreamBalanceConfig(account, cfg))
}

func TestValidateUpstreamBalanceConfigRequiresWalletAuthentication(t *testing.T) {
	account := &Account{
		Type: AccountTypeUpstream,
		Credentials: map[string]any{
			"api_key":  "model-key",
			"base_url": "https://relay.example.com/v1",
		},
	}
	cfg := &UpstreamBalanceConfig{
		Enabled:      true,
		Provider:     UpstreamBalanceProviderNewAPI,
		PlatformName: "Relay",
		AuthMode:     "login",
		AuthUsername: "wallet@example.com",
	}
	require.ErrorContains(t, validateUpstreamBalanceConfig(account, cfg), "credential is required")

	cfg.AuthToken = "password"
	require.NoError(t, validateUpstreamBalanceConfig(account, cfg))

	account.Credentials[upstreamBalanceAuthTokenCredentialKey] = "stored-password"
	cfg.AuthToken = ""
	require.NoError(t, validateUpstreamBalanceConfig(account, cfg))
	cfg.AuthCleared = true
	require.ErrorContains(t, validateUpstreamBalanceConfig(account, cfg), "credential is required")
}

func TestUpstreamBalanceConfigMasksWalletCredential(t *testing.T) {
	account := &Account{
		Credentials: map[string]any{upstreamBalanceAuthTokenCredentialKey: "secret"},
		Extra: map[string]any{
			upstreamBalanceAuthModeKey:     "login",
			upstreamBalanceAuthUsernameKey: "wallet@example.com",
		},
	}
	cfg := upstreamBalanceConfigFromAccount(account)
	require.True(t, cfg.AuthConfigured)
	require.Empty(t, cfg.AuthToken)
	require.Equal(t, "wallet@example.com", cfg.AuthUsername)
}

func TestShouldNotifyUpstreamBalanceRetriesUntilDelivered(t *testing.T) {
	now := time.Now().UTC()
	alerted := now.Add(-time.Hour)
	recovered := now.Add(-time.Minute)
	require.True(t, shouldNotifyUpstreamBalance(true, false, &alerted, nil))
	require.True(t, shouldNotifyUpstreamBalance(true, true, nil, nil))
	require.True(t, shouldNotifyUpstreamBalance(true, true, &alerted, &recovered))
	require.False(t, shouldNotifyUpstreamBalance(true, true, &alerted, nil))
	require.False(t, shouldNotifyUpstreamBalance(false, true, nil, nil))
}

func TestBalanceFundingIdentityNormalizesAutoProvider(t *testing.T) {
	account := &Account{Credentials: map[string]any{
		"api_key":  "shared-secret",
		"base_url": "https://api.deepseek.com/v1",
	}}
	auto := balanceFundingIdentity(account, UpstreamBalanceConfig{Provider: UpstreamBalanceProviderAuto})
	explicit := balanceFundingIdentity(account, UpstreamBalanceConfig{Provider: UpstreamBalanceProviderDeepSeek})
	require.Equal(t, explicit, auto)
}

func TestBalanceFundingIdentityScopesCustomKeysByOrigin(t *testing.T) {
	first := &Account{Credentials: map[string]any{
		"api_key":  "shared-secret",
		"base_url": "https://relay-one.example.com/v1",
	}}
	second := &Account{Credentials: map[string]any{
		"api_key":  "shared-secret",
		"base_url": "https://relay-two.example.com/v1",
	}}
	cfg := UpstreamBalanceConfig{Provider: UpstreamBalanceProviderCustom}
	require.NotEqual(t, balanceFundingIdentity(first, cfg), balanceFundingIdentity(second, cfg))
}

func TestBalanceFundingIdentityDeduplicatesWalletAcrossModelKeys(t *testing.T) {
	first := &Account{Credentials: map[string]any{
		"api_key":                             "model-key-one",
		"base_url":                            "https://relay.example.com/v1",
		upstreamBalanceAuthTokenCredentialKey: "wallet-password",
	}}
	second := &Account{Credentials: map[string]any{
		"api_key":                             "model-key-two",
		"base_url":                            "https://relay.example.com/v1",
		upstreamBalanceAuthTokenCredentialKey: "wallet-password",
	}}
	cfg := UpstreamBalanceConfig{Provider: UpstreamBalanceProviderNewAPI, AuthMode: "login", AuthUsername: "wallet@example.com"}
	require.Equal(t, balanceFundingIdentity(first, cfg), balanceFundingIdentity(second, cfg))
}

type upstreamBalanceOverviewRepoStub struct {
	AccountRepository
	accounts []Account
}

func (r *upstreamBalanceOverviewRepoStub) ListAllWithFilters(context.Context, string, string, string, string, int64, string) ([]Account, error) {
	return r.accounts, nil
}

func TestUpstreamBalanceOverviewExcludesStaleAmount(t *testing.T) {
	staleSuccess := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	repo := &upstreamBalanceOverviewRepoStub{accounts: []Account{
		{
			ID:       1,
			Name:     "stale-wallet",
			Type:     AccountTypeAPIKey,
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"api_key":  "model-key",
				"base_url": "https://api.deepseek.com/v1",
			},
			Extra: map[string]any{
				upstreamBalanceEnabledKey:       true,
				upstreamBalanceProviderKey:      UpstreamBalanceProviderDeepSeek,
				upstreamBalancePlatformNameKey:  "DeepSeek",
				upstreamBalanceLastAmountKey:    100.0,
				upstreamBalanceLastSuccessAtKey: staleSuccess,
			},
		},
	}}
	service := NewUpstreamBalanceService(repo, nil, nil, nil)
	defer service.Stop()
	overview, err := service.Overview(context.Background())
	require.NoError(t, err)
	require.Zero(t, overview.TotalAmount)
	require.Len(t, overview.Platforms, 1)
	require.Zero(t, overview.Platforms[0].Amount)
	require.Equal(t, 1, overview.Platforms[0].StaleCount)
}

func TestUpstreamBalanceOverviewPrefersFreshDuplicateWallet(t *testing.T) {
	staleSuccess := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
	freshSuccess := time.Now().UTC().Add(-time.Minute).Format(time.RFC3339)
	account := func(id int64, name, success string, amount float64) Account {
		return Account{
			ID:       id,
			Name:     name,
			Type:     AccountTypeAPIKey,
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"api_key":  "shared-wallet-key",
				"base_url": "https://api.deepseek.com/v1",
			},
			Extra: map[string]any{
				upstreamBalanceEnabledKey:       true,
				upstreamBalanceProviderKey:      UpstreamBalanceProviderDeepSeek,
				upstreamBalancePlatformNameKey:  "DeepSeek",
				upstreamBalanceLastAmountKey:    amount,
				upstreamBalanceLastSuccessAtKey: success,
			},
		}
	}
	repo := &upstreamBalanceOverviewRepoStub{accounts: []Account{
		account(1, "stale-model-key", staleSuccess, 100),
		account(2, "fresh-model-key", freshSuccess, 50),
	}}
	service := NewUpstreamBalanceService(repo, nil, nil, nil)
	defer service.Stop()
	overview, err := service.Overview(context.Background())
	require.NoError(t, err)
	require.Equal(t, 50.0, overview.TotalAmount)
	require.Len(t, overview.Platforms, 1)
	require.Equal(t, 50.0, overview.Platforms[0].Amount)
	require.Equal(t, 1, overview.Platforms[0].FundingCount)
}
