package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

type upstreamMonitorTestGroupLister struct {
	groups []Group
	err    error
}

type upstreamMonitorTestAccountLister struct {
	accounts []Account
	err      error
}

type upstreamMonitorSettingRepo struct {
	mu   sync.Mutex
	data map[string]string
}

func newUpstreamMonitorSettingRepo() *upstreamMonitorSettingRepo {
	return &upstreamMonitorSettingRepo{data: make(map[string]string)}
}

func (r *upstreamMonitorSettingRepo) Get(_ context.Context, key string) (*Setting, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.data[key]
	if !ok {
		return nil, ErrSettingNotFound
	}
	return &Setting{Key: key, Value: value}, nil
}

func (r *upstreamMonitorSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.data[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}

func (r *upstreamMonitorSettingRepo) Set(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = value
	return nil
}

func (r *upstreamMonitorSettingRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.data[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (r *upstreamMonitorSettingRepo) SetMultiple(_ context.Context, settings map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, value := range settings {
		r.data[key] = value
	}
	return nil
}

func (r *upstreamMonitorSettingRepo) GetAll(_ context.Context) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]string, len(r.data))
	for key, value := range r.data {
		out[key] = value
	}
	return out, nil
}

func (r *upstreamMonitorSettingRepo) Delete(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, key)
	return nil
}

func (l upstreamMonitorTestGroupLister) ListActive(context.Context) ([]Group, error) {
	return l.groups, l.err
}

func (l upstreamMonitorTestAccountLister) ListActive(context.Context) ([]Account, error) {
	return l.accounts, l.err
}

func stubUpstreamMonitorClient(
	t *testing.T,
	statusCode int,
	contentType string,
	body string,
) {
	t.Helper()

	original := newUpstreamMonitorHTTPClient
	newUpstreamMonitorHTTPClient = func() *req.Client {
		client := req.C().SetTimeout(5 * time.Second)
		client.GetTransport().WrapRoundTripFunc(func(http.RoundTripper) req.HttpRoundTripFunc {
			return func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: statusCode,
					Header:     http.Header{"Content-Type": []string{contentType}},
					Body:       io.NopCloser(strings.NewReader(body)),
					Request:    r,
				}, nil
			}
		})
		return client
	}
	t.Cleanup(func() {
		newUpstreamMonitorHTTPClient = original
	})
}

func stubUpstreamMonitorClientFunc(
	t *testing.T,
	handler func(*http.Request) (int, string, string),
) {
	t.Helper()
	stubUpstreamMonitorClientHeaderFunc(t, func(r *http.Request) (int, http.Header, string) {
		statusCode, contentType, body := handler(r)
		return statusCode, http.Header{"Content-Type": []string{contentType}}, body
	})
}

func stubUpstreamMonitorClientHeaderFunc(
	t *testing.T,
	handler func(*http.Request) (int, http.Header, string),
) {
	t.Helper()

	original := newUpstreamMonitorHTTPClient
	newUpstreamMonitorHTTPClient = func() *req.Client {
		client := req.C().SetTimeout(5 * time.Second)
		client.GetTransport().WrapRoundTripFunc(func(http.RoundTripper) req.HttpRoundTripFunc {
			return func(r *http.Request) (*http.Response, error) {
				statusCode, header, body := handler(r)
				return &http.Response{
					StatusCode: statusCode,
					Header:     header,
					Body:       io.NopCloser(strings.NewReader(body)),
					Request:    r,
				}, nil
			}
		})
		return client
	}
	t.Cleanup(func() {
		newUpstreamMonitorHTTPClient = original
	})
}

func TestSettingServiceGetUpstreamMonitorConfig_DoesNotRefreshOnRead(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "nbility",
				Name:            "Nbility",
				Kind:            "custom",
				Enabled:         true,
				AutoSyncEnabled: true,
				FetchMode:       upstreamMonitorFetchModeJSONPath,
				BaseURL:         "https://example.com",
				PricingURL:      "https://example.com/pricing",
				PricingPathHint: "data.reference_multiplier",
				AuthMode:        "none",
				Currency:        "CNY",
				ExchangeRate:    7.2,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{},
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	repo := newUpstreamMonitorSettingRepo()
	repo.data[SettingKeyUpstreamMonitorConfig] = string(raw)
	svc := NewUpstreamMonitorService(repo)

	got, err := svc.GetUpstreamMonitorConfig(context.Background())
	require.NoError(t, err)
	require.Len(t, got.Sources, 1)
	require.Zero(t, got.Sources[0].ReferenceMultiplier)
	require.Equal(t, upstreamMonitorSyncStatusIdle, got.Sources[0].LastSyncStatus)
	require.Nil(t, got.Sources[0].LastSyncAt)

	persisted := &UpstreamMonitorConfig{}
	require.NoError(t, json.Unmarshal([]byte(repo.data[SettingKeyUpstreamMonitorConfig]), persisted))
	require.Zero(t, persisted.Sources[0].ReferenceMultiplier)
	require.Empty(t, persisted.Sources[0].LastSyncStatus)
}

func TestUpstreamMonitorService_ReturnsErrorWithoutSettingRepo(t *testing.T) {
	svc := NewUpstreamMonitorService(nil)

	_, err := svc.GetUpstreamMonitorConfig(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "settings repository not configured")

	err = svc.SaveUpstreamMonitorConfig(context.Background(), &UpstreamMonitorConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "settings repository not configured")

	_, err = svc.RefreshUpstreamMonitorConfig(context.Background(), &UpstreamMonitorConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "settings repository not configured")

	var nilSvc *UpstreamMonitorService
	_, err = nilSvc.PreviewUpstreamMonitorConfig(context.Background(), &UpstreamMonitorConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "service not configured")
}

func TestSettingServiceGetUpstreamMonitorConfig_ReturnsErrorOnInvalidJSON(t *testing.T) {
	repo := newUpstreamMonitorSettingRepo()
	repo.data[SettingKeyUpstreamMonitorConfig] = `{bad json`
	svc := NewUpstreamMonitorService(repo)

	got, err := svc.GetUpstreamMonitorConfig(context.Background())
	require.Error(t, err)
	require.Nil(t, got)
	require.Contains(t, err.Error(), "parse upstream monitor config")
	require.Equal(t, `{bad json`, repo.data[SettingKeyUpstreamMonitorConfig])
}

func TestSettingServiceGetUpstreamMonitorConfig_AllowsEmptyLastSyncAt(t *testing.T) {
	repo := newUpstreamMonitorSettingRepo()
	repo.data[SettingKeyUpstreamMonitorConfig] = `{
		"enabled": true,
		"auto_refresh_enabled": true,
		"refresh_interval_minutes": 10,
		"default_exchange_rate": 7.2,
		"sources": [{
			"id": "manual",
			"name": "Manual",
			"kind": "manual",
			"enabled": true,
			"auto_sync_enabled": false,
			"auth_mode": "none",
			"currency": "CNY",
			"exchange_rate": 1,
			"reference_multiplier": 0.08,
			"last_sync_at": ""
		}],
		"group_mappings": []
	}`
	svc := NewUpstreamMonitorService(repo)

	got, err := svc.GetUpstreamMonitorConfig(context.Background())
	require.NoError(t, err)
	require.Len(t, got.Sources, 1)
	require.Nil(t, got.Sources[0].LastSyncAt)
}

func TestSettingServiceSaveUpstreamMonitorConfig_AllowsManualSourceWithoutURL(t *testing.T) {
	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "manual",
				Name:                "Manual Reference",
				Kind:                "manual",
				Enabled:             true,
				AutoSyncEnabled:     false,
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 1.2,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{},
	}

	require.NoError(t, svc.SaveUpstreamMonitorConfig(context.Background(), cfg))
}

func TestSettingServiceSaveUpstreamMonitorConfig_AllowsClearingStoredAuthToken(t *testing.T) {
	repo := newUpstreamMonitorSettingRepo()
	repo.data[SettingKeyUpstreamMonitorConfig] = `{
		"enabled": true,
		"auto_refresh_enabled": true,
		"refresh_interval_minutes": 10,
		"default_exchange_rate": 7.2,
		"sources": [{
			"id": "relay",
			"name": "Relay",
			"kind": "custom",
			"enabled": true,
			"auto_sync_enabled": true,
			"fetch_mode": "plain_text",
			"pricing_url": "https://example.com/pricing",
			"auth_mode": "header",
			"auth_header_name": "Authorization",
			"auth_token": "stored-token",
			"currency": "CNY",
			"exchange_rate": 7.2
		}],
		"group_mappings": []
	}`

	svc := NewUpstreamMonitorService(repo)
	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "relay",
				Name:                "Relay",
				Kind:                "custom",
				Enabled:             true,
				AutoSyncEnabled:     true,
				FetchMode:           upstreamMonitorFetchModePlainText,
				PricingURL:          "https://example.com/pricing",
				AuthMode:            "header",
				AuthHeaderName:      "Authorization",
				AuthTokenCleared:    true,
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 0,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{},
	}

	require.NoError(t, svc.SaveUpstreamMonitorConfig(context.Background(), cfg))
	raw, err := repo.GetValue(context.Background(), SettingKeyUpstreamMonitorConfig)
	require.NoError(t, err)

	persisted := &UpstreamMonitorConfig{}
	require.NoError(t, json.Unmarshal([]byte(raw), persisted))
	require.Len(t, persisted.Sources, 1)
	require.Empty(t, persisted.Sources[0].AuthToken)
	require.False(t, persisted.Sources[0].AuthConfigured)
}

func TestNormalizeUpstreamMonitorConfig_DerivesKnownPricingURLs(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "pool",
				Name:            "Pool",
				Kind:            "sub2api",
				Enabled:         true,
				AutoSyncEnabled: true,
				BaseURL:         "pool.gptstore.club/",
				AuthMode:        "bearer",
				Currency:        "CNY",
				ExchangeRate:    1,
			},
			{
				ID:              "newapi",
				Name:            "New API",
				Kind:            "newapi",
				Enabled:         true,
				AutoSyncEnabled: true,
				BaseURL:         "https://relay.example.com",
				AuthMode:        "none",
				Currency:        "CNY",
				ExchangeRate:    1,
			},
		},
	}

	normalizeUpstreamMonitorConfig(cfg)

	require.Equal(t, "https://pool.gptstore.club/api/v1/groups/available", cfg.Sources[0].PricingURL)
	require.Equal(t, "https://relay.example.com/api/user/self/groups", cfg.Sources[1].PricingURL)
	require.NoError(t, validateUpstreamMonitorConfig(cfg))
}

func TestValidateUpstreamMonitorConfig_AllowsLegacyJSONPathOnStandardSources(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "newapi",
				Name:            "New API",
				Kind:            "newapi",
				Enabled:         true,
				AutoSyncEnabled: true,
				FetchMode:       upstreamMonitorFetchModeJSONPath,
				BaseURL:         "https://relay.example.com",
				AuthMode:        "bearer",
				AuthToken:       "token",
				Currency:        "CNY",
				ExchangeRate:    1,
			},
		},
	}

	normalizeUpstreamMonitorConfig(cfg)

	require.Equal(t, "https://relay.example.com/api/user/self/groups", cfg.Sources[0].PricingURL)
	require.NoError(t, validateUpstreamMonitorConfig(cfg))
}

func TestFetchUpstreamPricingSnapshot_SendsBearerTokenForSub2API(t *testing.T) {
	stubUpstreamMonitorClientFunc(t, func(r *http.Request) (int, string, string) {
		require.Equal(t, "Bearer user-login-token", r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/api/v1/groups/available":
			return http.StatusOK, "application/json", `[
				{"id": 1, "name": "codex-cheap", "description": "cheap", "rate_multiplier": 0.08}
			]`
		case "/api/v1/groups/rates":
			return http.StatusOK, "application/json", `{"1":0.09}`
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return http.StatusNotFound, "text/plain", ""
		}
	})

	snapshot, err := fetchUpstreamPricingSnapshot(context.Background(), &UpstreamMonitorSource{
		ID:        "pool",
		Name:      "Pool",
		Kind:      "sub2api",
		BaseURL:   "https://pool.gptstore.club",
		AuthMode:  "bearer",
		AuthToken: "user-login-token",
	})

	require.NoError(t, err)
	require.Len(t, snapshot.GroupOptions, 1)
	require.Equal(t, "codex-cheap", snapshot.GroupOptions[0].Name)
	require.InDelta(t, 0.09, snapshot.GroupOptions[0].ReferenceMultiplier, 0.0001)
	require.Len(t, snapshot.GroupMultipliers, 1)
}

func TestFetchUpstreamPricingSnapshot_SendsJSONAccessTokenForSub2API(t *testing.T) {
	stubUpstreamMonitorClientFunc(t, func(r *http.Request) (int, string, string) {
		require.Equal(t, "Bearer json-token", r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/api/v1/groups/available":
			return http.StatusOK, "application/json", `[
				{"id": 1, "name": "codex-cheap", "description": "cheap", "rate_multiplier": 0.08}
			]`
		case "/api/v1/groups/rates":
			return http.StatusOK, "application/json", `{}`
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return http.StatusNotFound, "text/plain", ""
		}
	})

	snapshot, err := fetchUpstreamPricingSnapshot(context.Background(), &UpstreamMonitorSource{
		ID:        "pool",
		Name:      "Pool",
		Kind:      "sub2api",
		BaseURL:   "https://pool.gptstore.club",
		AuthMode:  "bearer",
		AuthToken: `{"access_token":"json-token"}`,
	})

	require.NoError(t, err)
	require.Len(t, snapshot.GroupOptions, 1)
	require.Equal(t, "codex-cheap", snapshot.GroupOptions[0].Name)
}

func TestFetchUpstreamPricingSnapshot_SendsNewAPICookieCredential(t *testing.T) {
	stubUpstreamMonitorClientFunc(t, func(r *http.Request) (int, string, string) {
		require.Equal(t, "session=abc; other=def", r.Header.Get("Cookie"))
		require.Equal(t, "123", r.Header.Get("New-Api-User"))
		require.Equal(t, "/api/user/self/groups", r.URL.Path)
		return http.StatusOK, "application/json", `{
			"default": {"ratio": 1, "desc": "default group"},
			"vip": {"ratio": 0.2, "desc": "vip group"}
		}`
	})

	snapshot, err := fetchUpstreamPricingSnapshot(context.Background(), &UpstreamMonitorSource{
		ID:        "newapi",
		Name:      "NewAPI",
		Kind:      "newapi",
		BaseURL:   "https://relay.example.com",
		AuthMode:  "cookie",
		AuthToken: `{"cookie":"session=abc; other=def","user_id":"123"}`,
	})

	require.NoError(t, err)
	require.Len(t, snapshot.GroupOptions, 2)
	require.InDelta(t, 1, snapshot.GroupMultipliers["default"], 0.0001)
	require.InDelta(t, 0.2, snapshot.GroupMultipliers["vip"], 0.0001)
}

func TestFetchUpstreamPricingSnapshot_NewAPICookieRequiresUserID(t *testing.T) {
	_, err := fetchUpstreamPricingSnapshot(context.Background(), &UpstreamMonitorSource{
		ID:        "newapi",
		Name:      "NewAPI",
		Kind:      "newapi",
		BaseURL:   "https://relay.example.com",
		AuthMode:  "cookie",
		AuthToken: "session=abc",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "newapi cookie auth requires JSON credential")
	require.Contains(t, err.Error(), "user_id")
}

func TestFetchUpstreamPricingSnapshot_Sub2APICookieExplainsAccessToken(t *testing.T) {
	_, err := fetchUpstreamPricingSnapshot(context.Background(), &UpstreamMonitorSource{
		ID:        "pool",
		Name:      "Pool",
		Kind:      "sub2api",
		BaseURL:   "https://pool.gptstore.club",
		AuthMode:  "cookie",
		AuthToken: "session=abc",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "sub2api standard endpoints use Bearer access_token")
}

func TestFetchUpstreamPricingSnapshot_LoginModeFetchesNewAPIGroups(t *testing.T) {
	stubUpstreamMonitorClientHeaderFunc(t, func(r *http.Request) (int, http.Header, string) {
		switch r.URL.Path {
		case "/api/user/login":
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			return http.StatusOK, http.Header{
				"Content-Type": []string{"application/json"},
				"Set-Cookie":   []string{"session=test-session; Path=/; HttpOnly"},
			}, `{"success":true,"data":{"id":123}}`
		case "/api/user/self/groups":
			require.Equal(t, "123", r.Header.Get("New-Api-User"))
			require.NotEmpty(t, r.Header.Get("Cookie"))
			return http.StatusOK, http.Header{"Content-Type": []string{"application/json"}}, `{
				"default": {"ratio": 1, "desc": "default group"},
				"vip": {"ratio": 0.2, "desc": "vip group"},
				"auto": {"ratio": "自动", "desc": "skip"}
			}`
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return http.StatusNotFound, http.Header{"Content-Type": []string{"text/plain"}}, ""
		}
	})

	snapshot, err := fetchUpstreamPricingSnapshot(context.Background(), &UpstreamMonitorSource{
		ID:           "newapi",
		Name:         "NewAPI",
		Kind:         "newapi",
		BaseURL:      "https://relay.example.com",
		AuthMode:     "login",
		AuthUsername: "monitor@example.com",
		AuthToken:    "secret-password",
	})

	require.NoError(t, err)
	require.Len(t, snapshot.GroupOptions, 2)
	require.InDelta(t, 1, snapshot.GroupMultipliers["default"], 0.0001)
	require.InDelta(t, 0.2, snapshot.GroupMultipliers["vip"], 0.0001)
}

func TestFetchUpstreamPricingSnapshot_LoginModeFetchesSessionForSub2API(t *testing.T) {
	stubUpstreamMonitorClientFunc(t, func(r *http.Request) (int, string, string) {
		switch r.URL.Path {
		case "/api/v1/auth/login":
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))
			return http.StatusOK, "application/json", `{"code":0,"message":"ok","data":{"access_token":"session-token","expires_in":3600}}`
		case "/api/v1/groups/available":
			require.Equal(t, "Bearer session-token", r.Header.Get("Authorization"))
			return http.StatusOK, "application/json", `[
				{"id": 12, "name": "codex-cheap", "description": "cheap", "rate_multiplier": 0.08}
			]`
		case "/api/v1/groups/rates":
			require.Equal(t, "Bearer session-token", r.Header.Get("Authorization"))
			return http.StatusOK, "application/json", `{}`
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return http.StatusNotFound, "text/plain", ""
		}
	})

	snapshot, err := fetchUpstreamPricingSnapshot(context.Background(), &UpstreamMonitorSource{
		ID:           "pool",
		Name:         "Pool",
		Kind:         "sub2api",
		BaseURL:      "https://pool.gptstore.club",
		PricingURL:   "https://pool.gptstore.club/api/v1/groups/available",
		AuthMode:     "login",
		AuthUsername: "monitor@example.com",
		AuthToken:    "secret-password",
	})

	require.NoError(t, err)
	require.Len(t, snapshot.GroupOptions, 1)
	require.Equal(t, "codex-cheap", snapshot.GroupOptions[0].Name)
	require.InDelta(t, 0.08, snapshot.GroupOptions[0].ReferenceMultiplier, 0.0001)
}

func TestFetchUpstreamPricingSnapshot_401ShowsAuthHint(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusUnauthorized, "application/json", `{"error":"unauthorized"}`)

	_, err := fetchUpstreamPricingSnapshot(context.Background(), &UpstreamMonitorSource{
		ID:         "pool",
		Name:       "Pool",
		Kind:       "sub2api",
		PricingURL: "https://pool.gptstore.club/api/v1/groups/available",
		AuthMode:   "none",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "401 unauthorized")
	require.Contains(t, err.Error(), "credential")
}

func TestFetchUpstreamPricingSnapshot_403ShowsForbiddenHint(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusForbidden, "application/json", `{"error":"forbidden"}`)

	_, err := fetchUpstreamPricingSnapshot(context.Background(), &UpstreamMonitorSource{
		ID:         "pool",
		Name:       "Pool",
		Kind:       "sub2api",
		PricingURL: "https://pool.gptstore.club/api/v1/groups/available",
		AuthMode:   "bearer",
		AuthToken:  "user-login-token",
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "403 forbidden")
}

func TestSettingServicePreviewUpstreamMonitorConfig_UsesDefaultProfitThresholdAsWarningFallback(t *testing.T) {
	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "VIP",
				Platform:       "openai",
				RateMultiplier: 2.00,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                    true,
		AutoRefreshEnabled:         true,
		DefaultExchangeRate:        7.2,
		DefaultProfitRateThreshold: 0.25,
		WarningRateThreshold:       0,
		CriticalRateThreshold:      0.10,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "manual",
				Name:                "Manual Reference",
				Kind:                "manual",
				Enabled:             true,
				AutoSyncEnabled:     false,
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 1.60,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:          "map_1",
				LocalGroup:  "VIP",
				ModelFamily: "gpt",
				SourceIDs:   []string{"manual"},
			},
		},
	}

	snapshot, err := svc.PreviewUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, snapshot.GroupOptions, 1)
	require.Equal(t, "VIP", snapshot.GroupOptions[0].GroupName)
	require.InDelta(t, 2.00, snapshot.GroupOptions[0].Multiplier, 0.0001)
	require.Len(t, snapshot.GroupRows, 1)
	require.InDelta(t, 0.20, snapshot.GroupRows[0].EstimatedMarginRate, 0.0001)
	require.Equal(t, "warning", snapshot.GroupRows[0].Status)
}

func TestSettingServicePreviewUpstreamMonitorConfig_MergesStoredSecretsForAuthSources(t *testing.T) {
	repo := newUpstreamMonitorSettingRepo()
	repo.data[SettingKeyUpstreamMonitorConfig] = `{
		"enabled": true,
		"auto_refresh_enabled": true,
		"refresh_interval_minutes": 10,
		"default_exchange_rate": 7.2,
		"sources": [{
			"id": "relay",
			"name": "Relay",
			"kind": "custom",
			"enabled": true,
			"auto_sync_enabled": true,
			"fetch_mode": "plain_text",
			"pricing_url": "https://example.com/pricing",
			"auth_mode": "header",
			"auth_header_name": "Authorization",
			"auth_token": "stored-token",
			"currency": "CNY",
			"exchange_rate": 7.2
		}],
		"group_mappings": [{
			"id": "map_1",
			"local_group": "VIP",
			"model_family": "gpt",
			"source_ids": ["relay"]
		}]
	}`

	svc := NewUpstreamMonitorService(repo)
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "VIP",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.2,
			},
		},
	})

	input := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "relay",
				Name:            "Relay",
				Kind:            "custom",
				Enabled:         true,
				AutoSyncEnabled: true,
				FetchMode:       upstreamMonitorFetchModePlainText,
				PricingURL:      "https://example.com/pricing",
				AuthMode:        "header",
				AuthHeaderName:  "Authorization",
				Currency:        "CNY",
				ExchangeRate:    7.2,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:          "map_1",
				LocalGroup:  "VIP",
				ModelFamily: "gpt",
				SourceIDs:   []string{"relay"},
			},
		},
	}

	snapshot, err := svc.PreviewUpstreamMonitorConfig(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, snapshot)
	require.Len(t, snapshot.SourceRows, 1)
	require.True(t, snapshot.SourceRows[0].AuthConfigured)
}

func TestSettingServiceRefreshUpstreamMonitorConfig_BuildsSnapshot(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusOK, "text/plain", "1.88")

	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "VIP",
				Platform:       "openai",
				RateMultiplier: 2.20,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.08,
		CriticalRateThreshold:  0,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "plain",
				Name:            "Plain Text Upstream",
				Kind:            "custom",
				Enabled:         true,
				AutoSyncEnabled: true,
				FetchMode:       upstreamMonitorFetchModePlainText,
				BaseURL:         "https://example.com",
				PricingURL:      "https://example.com/pricing",
				AuthMode:        "none",
				Currency:        "CNY",
				ExchangeRate:    7.2,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:          "map_1",
				LocalGroup:  "VIP",
				ModelFamily: "gpt",
				SourceIDs:   []string{"plain"},
			},
		},
	}

	result, err := svc.RefreshUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Config)
	require.NotNil(t, result.Snapshot)
	require.Equal(t, 1, result.Summary.AttemptedCount)
	require.Equal(t, 1, result.Summary.SuccessCount)
	require.Equal(t, 0, result.Summary.FailedCount)
	require.InDelta(t, 1.88, result.Config.Sources[0].ReferenceMultiplier, 0.0001)
	require.Equal(t, upstreamMonitorSyncStatusSuccess, result.Config.Sources[0].LastSyncStatus)
	require.InDelta(t, 1.88, result.Snapshot.GroupRows[0].ReferenceMultiplier, 0.0001)
	require.Equal(t, "healthy", result.Snapshot.GroupRows[0].Status)
}

func TestSettingServiceRefreshUpstreamMonitorConfig_UpdatesMappingByUpstreamGroup(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusOK, "application/json", `{
		"reference_multiplier": 0.50,
		"groups": [
			{"group_name": "relay-0.2", "multiplier": 0.10},
			{"group_name": "relay-0.5", "multiplier": 0.35}
		]
	}`)

	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "0.2倍率",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.20,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.08,
		CriticalRateThreshold:  0,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "relay",
				Name:                "Relay",
				Kind:                "custom",
				Enabled:             true,
				AutoSyncEnabled:     true,
				FetchMode:           upstreamMonitorFetchModeAuto,
				BaseURL:             "https://example.com",
				PricingURL:          "https://example.com/pricing",
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 0.08,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:            "map_1",
				LocalGroupID:  1,
				LocalGroup:    "0.2倍率",
				UpstreamGroup: "relay-0.2",
				ModelFamily:   "gpt",
				SourceIDs:     []string{"relay"},
			},
		},
	}

	result, err := svc.RefreshUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.InDelta(t, 0.50, result.Config.Sources[0].ReferenceMultiplier, 0.0001)
	require.InDelta(t, 0.10, result.Config.GroupMappings[0].ReferenceMultiplier, 0.0001)
	require.InDelta(t, 0.10, result.Snapshot.GroupRows[0].ReferenceMultiplier, 0.0001)
	require.InDelta(t, 0.50, result.Snapshot.GroupRows[0].EstimatedMarginRate, 0.0001)
}

func TestSettingServiceRefreshUpstreamMonitorConfig_UpdatesDuplicateUpstreamGroupByKey(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusOK, "application/json", `{
		"groups": [
			{"id": "cheap", "group_name": "0.2倍率", "multiplier": 0.08},
			{"id": "premium", "group_name": "0.2倍率", "multiplier": 0.10}
		]
	}`)

	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "0.2倍率",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.20,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.08,
		CriticalRateThreshold:  0,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "relay",
				Name:            "Relay",
				Kind:            "custom",
				Enabled:         true,
				AutoSyncEnabled: true,
				FetchMode:       upstreamMonitorFetchModeAuto,
				PricingURL:      "https://example.com/pricing",
				AuthMode:        "none",
				Currency:        "CNY",
				ExchangeRate:    7.2,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:               "map_1",
				LocalGroupID:     1,
				LocalGroup:       "0.2倍率",
				UpstreamGroupKey: "id:premium",
				UpstreamGroup:    "0.2倍率",
				ModelFamily:      "gpt",
				SourceIDs:        []string{"relay"},
			},
		},
	}

	result, err := svc.RefreshUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Config.Sources[0].UpstreamGroupOptions, 2)
	require.Equal(t, "id:premium", result.Config.GroupMappings[0].UpstreamGroupKey)
	require.Equal(t, "0.2倍率", result.Config.GroupMappings[0].UpstreamGroup)
	require.InDelta(t, 0.10, result.Config.GroupMappings[0].ReferenceMultiplier, 0.0001)
	require.Equal(t, "id:premium", result.Snapshot.GroupRows[0].UpstreamGroupKey)
	require.InDelta(t, 0.10, result.Snapshot.GroupRows[0].ReferenceMultiplier, 0.0001)
}

func TestSettingServiceRefreshUpstreamMonitorConfig_AddsConfiguredGroupOptionWhenNoUpstreamGroups(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusOK, "application/json", `{
		"reference_multiplier": 0.08
	}`)

	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "0.2倍率",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.20,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "relay",
				Name:            "Relay",
				Kind:            "custom",
				Enabled:         true,
				AutoSyncEnabled: true,
				FetchMode:       upstreamMonitorFetchModeAuto,
				PricingURL:      "https://example.com/pricing",
				AuthMode:        "none",
				Currency:        "CNY",
				ExchangeRate:    7.2,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:            "map_1",
				LocalGroupID:  1,
				LocalGroup:    "0.2倍率",
				UpstreamGroup: "0.2倍率",
				ModelFamily:   "gpt",
				SourceIDs:     []string{"relay"},
			},
		},
	}

	result, err := svc.RefreshUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, result.Config.Sources[0].UpstreamGroupOptions, 1)
	option := result.Config.Sources[0].UpstreamGroupOptions[0]
	require.Equal(t, "0.2倍率", option.Name)
	require.InDelta(t, 0.08, option.ReferenceMultiplier, 0.0001)
	require.Contains(t, option.Path, "configured:relay")
}

func TestNormalizeUpstreamMonitorConfig_SplitsMappingsBySource(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{ID: "source_a", Name: "A", Kind: "custom", Currency: "CNY", ExchangeRate: 7.2},
			{ID: "source_b", Name: "B", Kind: "custom", Currency: "CNY", ExchangeRate: 7.2},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:            "map_1",
				LocalGroupID:  1,
				LocalGroup:    "0.2倍率",
				UpstreamGroup: "0.2倍率",
				ModelFamily:   "gpt",
				SourceIDs:     []string{"source_a", "source_b"},
			},
		},
	}

	normalizeUpstreamMonitorConfig(cfg)
	require.Len(t, cfg.GroupMappings, 2)
	require.Equal(t, "map_1__source_a", cfg.GroupMappings[0].ID)
	require.Equal(t, []string{"source_a"}, cfg.GroupMappings[0].SourceIDs)
	require.Equal(t, "map_1__source_b", cfg.GroupMappings[1].ID)
	require.Equal(t, []string{"source_b"}, cfg.GroupMappings[1].SourceIDs)
	require.NoError(t, validateUpstreamMonitorConfig(cfg))
}

func TestNormalizeUpstreamMonitorConfig_KeepsMultipleUpstreamGroupsForSameLocalGroup(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{ID: "source_a", Name: "A", Kind: "custom", Currency: "CNY", ExchangeRate: 7.2},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:               "map_cheap",
				LocalGroupID:     1,
				LocalGroup:       "0.2倍率",
				UpstreamGroupKey: "id:cheap",
				UpstreamGroup:    "codex-cheap",
				ModelFamily:      "gpt",
				SourceIDs:        []string{"source_a"},
			},
			{
				ID:               "map_fast",
				LocalGroupID:     1,
				LocalGroup:       "0.2倍率",
				UpstreamGroupKey: "id:fast",
				UpstreamGroup:    "codex-fast",
				ModelFamily:      "gpt",
				SourceIDs:        []string{"source_a"},
			},
		},
	}

	normalizeUpstreamMonitorConfig(cfg)
	require.Len(t, cfg.GroupMappings, 2)
	require.Equal(t, "id:cheap", cfg.GroupMappings[0].UpstreamGroupKey)
	require.Equal(t, "id:fast", cfg.GroupMappings[1].UpstreamGroupKey)
	require.NoError(t, validateUpstreamMonitorConfig(cfg))
}

func TestNormalizeUpstreamMonitorConfig_KeepsDistinctModelFamiliesForSameLocalGroup(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{ID: "source_a", Name: "A", Kind: "custom", Currency: "CNY", ExchangeRate: 7.2},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:                  "map_gpt",
				LocalGroupID:        1,
				LocalGroup:          "0.2倍率",
				UpstreamGroupKey:    "id:cheap",
				UpstreamGroup:       "codex-cheap",
				ModelFamily:         "gpt",
				ReferenceMultiplier: 0.08,
				SourceIDs:           []string{"source_a"},
			},
			{
				ID:                  "map_claude",
				LocalGroupID:        1,
				LocalGroup:          "0.2倍率",
				UpstreamGroupKey:    "id:cheap",
				UpstreamGroup:       "codex-cheap",
				ModelFamily:         "claude",
				ReferenceMultiplier: 0.10,
				SourceIDs:           []string{"source_a"},
			},
		},
	}

	normalizeUpstreamMonitorConfig(cfg)
	require.Len(t, cfg.GroupMappings, 2)
	require.Equal(t, "gpt", cfg.GroupMappings[0].ModelFamily)
	require.Equal(t, "claude", cfg.GroupMappings[1].ModelFamily)
	require.InDelta(t, 0.08, cfg.GroupMappings[0].ReferenceMultiplier, 0.00000001)
	require.InDelta(t, 0.10, cfg.GroupMappings[1].ReferenceMultiplier, 0.00000001)
	require.NoError(t, validateUpstreamMonitorConfig(cfg))
}

func TestCollectUpstreamGroupOptions_IgnoresGenericNumericFields(t *testing.T) {
	body := []byte(`{
		"success": true,
		"data": {
			"total": 100,
			"balance": 52.5,
			"groups": [
				{"id": "stable", "name": "0.2倍率", "multiplier": 0.08}
			]
		}
	}`)

	options := collectUpstreamGroupOptions(body)
	require.Len(t, options, 1)
	require.Equal(t, "id:stable", options[0].Key)
	require.Equal(t, "0.2倍率", options[0].Name)
	require.InDelta(t, 0.08, options[0].ReferenceMultiplier, 0.0001)

	multipliers := collectUpstreamGroupMultipliers(body)
	require.Len(t, multipliers, 1)
	require.InDelta(t, 0.08, multipliers["0.2倍率"], 0.0001)
}

func TestCollectUpstreamGroupOptions_ParsesRatioLikeRootObject(t *testing.T) {
	body := []byte(`{
		"0.2倍率": 0.08,
		"0.5倍率": 0.20
	}`)

	options := collectUpstreamGroupOptions(body)
	require.Len(t, options, 2)
	require.Equal(t, "0.2倍率", options[0].Name)
	require.InDelta(t, 0.08, options[0].ReferenceMultiplier, 0.0001)

	multipliers := collectUpstreamGroupMultipliers(body)
	require.Len(t, multipliers, 2)
	require.InDelta(t, 0.08, multipliers["0.2倍率"], 0.0001)
	require.InDelta(t, 0.20, multipliers["0.5倍率"], 0.0001)
}

func TestCollectUpstreamGroupOptions_ParsesSub2APIAvailableChannels(t *testing.T) {
	body := []byte(`[
		{
			"name": "MagicAI",
			"description": "OpenAI relay",
			"platforms": [
				{
					"platform": "openai",
					"groups": [
						{
							"id": 12,
							"name": "codex-cheap",
							"platform": "openai",
							"rate_multiplier": 0.08,
							"is_exclusive": false
						},
						{
							"id": 13,
							"name": "codex-pro",
							"platform": "openai",
							"rate_multiplier": 1,
							"is_exclusive": false
						}
					],
					"supported_models": [
						{
							"name": "gpt-5.5",
							"platform": "openai",
							"pricing": {
								"billing_mode": "token",
								"input_price": 0.000001,
								"output_price": 0.000008
							}
						}
					]
				}
			]
		}
	]`)

	options := collectUpstreamGroupOptions(body)
	require.Len(t, options, 2)
	require.Equal(t, "id:12", options[0].Key)
	require.Equal(t, "codex-cheap", options[0].Name)
	require.InDelta(t, 0.08, options[0].ReferenceMultiplier, 0.0001)
	require.Equal(t, "id:13", options[1].Key)
	require.Equal(t, "codex-pro", options[1].Name)
	require.InDelta(t, 1, options[1].ReferenceMultiplier, 0.0001)

	multipliers := collectUpstreamGroupMultipliers(body)
	require.Len(t, multipliers, 2)
	require.InDelta(t, 0.08, multipliers["codex-cheap"], 0.0001)
	require.InDelta(t, 1, multipliers["codex-pro"], 0.0001)
}

func TestCollectUpstreamGroupOptions_ParsesNbilityPricingGroups(t *testing.T) {
	body := []byte(`{
		"success": true,
		"group_ratio": {
			"codex-cheap": 0.08,
			"codex-pro": 1
		},
		"usable_group": {
			"codex-cheap": "限时特惠 | OpenAI GPT 超低价分组",
			"codex-pro": "稳定首推 | OpenAI GPT Pro专用分组"
		},
			"data": [
				{
					"model_name": "gpt-5.5",
					"model_ratio": 2.5,
					"completion_ratio": 6,
					"cache_ratio": 0.1,
					"enable_groups": ["codex-pro", "codex-cheap"]
				}
			]
	}`)

	options := collectUpstreamGroupOptions(body)
	require.Len(t, options, 2)
	require.Equal(t, "codex-cheap", options[0].Name)
	require.Equal(t, "限时特惠 | OpenAI GPT 超低价分组", options[0].Description)
	require.InDelta(t, 0.08, options[0].ReferenceMultiplier, 0.0001)
	require.NotContains(t, []string{options[0].Name, options[1].Name}, "cache_ratio")
	require.NotContains(t, []string{options[0].Name, options[1].Name}, "completion_ratio")
	require.NotContains(t, []string{options[0].Name, options[1].Name}, "model_ratio")

	multipliers := collectUpstreamGroupMultipliers(body)
	require.Len(t, multipliers, 2)
	require.InDelta(t, 0.08, multipliers["codex-cheap"], 0.0001)
	require.InDelta(t, 1, multipliers["codex-pro"], 0.0001)
}

func TestCollectUpstreamGroupOptions_IgnoresModelPricingMaps(t *testing.T) {
	body := []byte(`{
		"success": true,
		"data": {
			"group_ratio": {
				"codex-cheap": 0.08
			},
			"model_ratio": {
				"gpt-5.5": 2.5,
				"claude-opus-4": 3
			},
			"model_price": {
				"gpt-image-1": 0.04
			}
		}
	}`)

	options := collectUpstreamGroupOptions(body)
	require.Len(t, options, 1)
	require.Equal(t, "codex-cheap", options[0].Name)
	require.InDelta(t, 0.08, options[0].ReferenceMultiplier, 0.0001)

	multipliers := collectUpstreamGroupMultipliers(body)
	require.Len(t, multipliers, 1)
	require.InDelta(t, 0.08, multipliers["codex-cheap"], 0.0001)
	require.NotContains(t, multipliers, "gpt-5.5")
	require.NotContains(t, multipliers, "claude-opus-4")
	require.NotContains(t, multipliers, "gpt-image-1")
}

func TestSettingServiceRefreshUpstreamMonitorConfig_UpdatesMappingFromGroupRatio(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusOK, "application/json", `{
		"success": true,
		"group_ratio": {
			"codex-618": 0.2,
			"codex-cheap": 0.08
		}
	}`)

	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "0.2倍率",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.20,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.08,
		CriticalRateThreshold:  0,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "nbility",
				Name:            "Nbility",
				Kind:            "custom",
				Enabled:         true,
				AutoSyncEnabled: true,
				FetchMode:       upstreamMonitorFetchModeAuto,
				BaseURL:         "https://nbility.dev",
				PricingURL:      "https://nbility.dev/api/pricing",
				AuthMode:        "none",
				Currency:        "CNY",
				ExchangeRate:    1,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:            "map_1",
				LocalGroupID:  1,
				LocalGroup:    "0.2倍率",
				UpstreamGroup: "codex-618",
				ModelFamily:   "gpt",
				SourceIDs:     []string{"nbility"},
			},
		},
	}

	result, err := svc.RefreshUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Zero(t, result.Config.Sources[0].ReferenceMultiplier)
	require.InDelta(t, 0.20, result.Config.GroupMappings[0].ReferenceMultiplier, 0.0001)
	require.InDelta(t, 0.20, result.Snapshot.GroupRows[0].ReferenceMultiplier, 0.0001)
}

func TestSettingServicePreviewUpstreamMonitorConfig_UsesMappingReferenceMultiplier(t *testing.T) {
	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "0.2倍率",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.20,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.08,
		CriticalRateThreshold:  0,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "nb",
				Name:                "Nbility",
				Kind:                "custom",
				Enabled:             true,
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 0.08,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:                  "map_1",
				LocalGroup:          "0.2倍率",
				UpstreamGroup:       "nb-0.08",
				ModelFamily:         "gpt",
				SourceIDs:           []string{"nb"},
				ReferenceMultiplier: 0.12,
			},
		},
	}

	snapshot, err := svc.PreviewUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, snapshot.GroupRows, 1)
	row := snapshot.GroupRows[0]
	require.Equal(t, "0.2倍率", row.LocalGroup)
	require.Equal(t, "nb-0.08", row.UpstreamGroup)
	require.InDelta(t, 0.20, row.LocalGroupMultiplier, 0.0001)
	require.InDelta(t, 0.12, row.ReferenceMultiplier, 0.0001)
	require.InDelta(t, 0.12, row.MappingMultiplier, 0.0001)
	require.InDelta(t, (0.20-0.12)/0.20, row.EstimatedMarginRate, 0.0001)
	require.Equal(t, "healthy", row.Status)
}

func TestSettingServicePreviewUpstreamMonitorConfig_AppliesExchangeRateForForeignCurrency(t *testing.T) {
	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "GPT",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.50,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.08,
		CriticalRateThreshold:  0,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "usd_pool",
				Name:                "USD Pool",
				Kind:                "custom",
				Enabled:             true,
				AuthMode:            "none",
				Currency:            "USD",
				ExchangeRate:        7,
				ReferenceMultiplier: 0.08,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:                  "map_usd",
				LocalGroup:          "GPT",
				UpstreamGroup:       "usd-gpt",
				ModelFamily:         "gpt",
				SourceIDs:           []string{"usd_pool"},
				ReferenceMultiplier: 0.08,
			},
		},
	}

	snapshot, err := svc.PreviewUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, snapshot.GroupRows, 1)
	row := snapshot.GroupRows[0]
	require.InDelta(t, 0.56, row.ReferenceMultiplier, 0.0001)
	require.InDelta(t, (0.50-0.56)/0.50, row.EstimatedMarginRate, 0.0001)
	require.Equal(t, "critical", row.Status)
}

func TestSettingServicePreviewUpstreamMonitorConfig_UsesLocalGroupID(t *testing.T) {
	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "共享倍率",
				Platform:       PlatformOpenAI,
				RateMultiplier: 0.10,
			},
			{
				ID:             2,
				Name:           "共享倍率",
				Platform:       PlatformAnthropic,
				RateMultiplier: 0.30,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.08,
		CriticalRateThreshold:  0,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "relay",
				Name:                "Relay",
				Kind:                "custom",
				Enabled:             true,
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 0.12,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:            "map_1",
				LocalGroupID:  2,
				LocalGroup:    "共享倍率",
				UpstreamGroup: "relay-claude",
				ModelFamily:   "claude",
				SourceIDs:     []string{"relay"},
			},
		},
	}

	snapshot, err := svc.PreviewUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, snapshot.GroupRows, 1)
	row := snapshot.GroupRows[0]
	require.Equal(t, int64(2), row.LocalGroupID)
	require.Equal(t, PlatformAnthropic, row.LocalGroupPlatform)
	require.InDelta(t, 0.30, row.LocalGroupMultiplier, 0.0001)
	require.InDelta(t, 0.60, row.EstimatedMarginRate, 0.0001)
}

func TestValidateUpstreamMonitorConfig_RejectsDuplicateBusinessMappingOnSameSource(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "source_a",
				Name:                "Source A",
				Kind:                "custom",
				Enabled:             true,
				AutoSyncEnabled:     true,
				FetchMode:           upstreamMonitorFetchModeAuto,
				PricingURL:          "https://example.com/a",
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 0.08,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:                  "map_1",
				LocalGroupID:        1,
				LocalGroup:          "GPT",
				UpstreamGroupKey:    "up:gpt",
				UpstreamGroup:       "GPT upstream",
				ModelFamily:         "gpt",
				SourceIDs:           []string{"source_a"},
				ReferenceMultiplier: 0.08,
			},
			{
				ID:                  "map_2",
				LocalGroupID:        1,
				LocalGroup:          "GPT",
				UpstreamGroupKey:    "up:gpt",
				UpstreamGroup:       "GPT upstream",
				ModelFamily:         "gpt",
				SourceIDs:           []string{"source_a"},
				ReferenceMultiplier: 0.10,
			},
		},
	}

	normalizeUpstreamMonitorConfig(cfg)
	err := validateUpstreamMonitorConfig(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate local/upstream mapping for source")
}

func TestValidateUpstreamMonitorConfig_AllowsSameLocalAndUpstreamForDifferentModelFamilies(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "source_a",
				Name:                "Source A",
				Kind:                "custom",
				Enabled:             true,
				AutoSyncEnabled:     true,
				FetchMode:           upstreamMonitorFetchModeAuto,
				PricingURL:          "https://example.com/a",
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 0.08,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:                  "map_1",
				LocalGroupID:        1,
				LocalGroup:          "GPT",
				UpstreamGroupKey:    "up:gpt",
				UpstreamGroup:       "GPT upstream",
				ModelFamily:         "gpt",
				SourceIDs:           []string{"source_a"},
				ReferenceMultiplier: 0.08,
			},
			{
				ID:                  "map_2",
				LocalGroupID:        1,
				LocalGroup:          "GPT",
				UpstreamGroupKey:    "up:gpt",
				UpstreamGroup:       "GPT upstream",
				ModelFamily:         "claude",
				SourceIDs:           []string{"source_a"},
				ReferenceMultiplier: 0.10,
			},
		},
	}

	normalizeUpstreamMonitorConfig(cfg)
	require.NoError(t, validateUpstreamMonitorConfig(cfg))
}

func TestSettingServiceRefreshStoredUpstreamMonitorConfig_InitializesMissingConfig(t *testing.T) {
	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{})

	result, err := svc.RefreshStoredUpstreamMonitorConfig(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Config)
	require.False(t, result.Config.Enabled)
	require.NotNil(t, result.Snapshot)
	require.Equal(t, 0, result.Summary.AttemptedCount)
}

func TestSettingServiceRefreshStoredUpstreamMonitorSource_RefreshesOnlyRequestedSource(t *testing.T) {
	var requestedPaths []string
	stubUpstreamMonitorClientFunc(t, func(r *http.Request) (int, string, string) {
		requestedPaths = append(requestedPaths, r.URL.Path)
		return http.StatusOK, "application/json", `{
			"groups": [
				{"id": "vip", "group_name": "VIP", "multiplier": 0.08}
			]
		}`
	})

	repo := newUpstreamMonitorSettingRepo()
	stored := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "source_a",
				Name:                "Source A",
				Kind:                "custom",
				Enabled:             true,
				AutoSyncEnabled:     true,
				FetchMode:           upstreamMonitorFetchModeAuto,
				PricingURL:          "https://example.com/source-a",
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 0.20,
			},
			{
				ID:                  "source_b",
				Name:                "Source B",
				Kind:                "custom",
				Enabled:             false,
				AutoSyncEnabled:     false,
				FetchMode:           upstreamMonitorFetchModeAuto,
				PricingURL:          "https://example.com/source-b",
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 0.30,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{},
	}
	raw, err := json.Marshal(stored)
	require.NoError(t, err)
	repo.data[SettingKeyUpstreamMonitorConfig] = string(raw)

	svc := NewUpstreamMonitorService(repo)
	svc.SetGroupLister(upstreamMonitorTestGroupLister{})

	result, err := svc.RefreshStoredUpstreamMonitorSource(context.Background(), "source_b")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, []string{"/source-b"}, requestedPaths)
	require.Equal(t, 1, result.Summary.AttemptedCount)
	require.Equal(t, 1, result.Summary.SuccessCount)
	require.Len(t, result.Config.Sources, 2)
	require.Equal(t, upstreamMonitorSyncStatusIdle, result.Config.Sources[0].LastSyncStatus)
	require.Empty(t, result.Config.Sources[0].UpstreamGroupOptions)
	require.False(t, result.Config.Sources[1].Enabled)
	require.Equal(t, upstreamMonitorSyncStatusSuccess, result.Config.Sources[1].LastSyncStatus)
	require.Len(t, result.Config.Sources[1].UpstreamGroupOptions, 1)
	require.Equal(t, "VIP", result.Config.Sources[1].UpstreamGroupOptions[0].Name)
}

func TestSettingServiceRefreshUpstreamMonitorConfig_RecordsSyncError(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusOK, "text/plain", "not-a-number")

	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{})

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "bad",
				Name:                "Bad Upstream",
				Kind:                "custom",
				Enabled:             true,
				AutoSyncEnabled:     true,
				FetchMode:           upstreamMonitorFetchModePlainText,
				BaseURL:             "https://example.com",
				PricingURL:          "https://example.com/pricing",
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 2.34,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{},
	}

	result, err := svc.RefreshUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 1, result.Summary.AttemptedCount)
	require.Equal(t, 0, result.Summary.SuccessCount)
	require.Equal(t, 1, result.Summary.FailedCount)
	require.InDelta(t, 2.34, result.Config.Sources[0].ReferenceMultiplier, 0.0001)
	require.Equal(t, upstreamMonitorSyncStatusError, result.Config.Sources[0].LastSyncStatus)
	require.NotEmpty(t, result.Config.Sources[0].LastSyncError)
}

func TestSettingServicePreviewUpstreamMonitorConfig_IgnoresDisabledSourceMultiplier(t *testing.T) {
	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{
				ID:             1,
				Name:           "VIP",
				Platform:       "openai",
				RateMultiplier: 3.00,
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:               true,
		AutoRefreshEnabled:    true,
		DefaultExchangeRate:   7.2,
		WarningRateThreshold:  0.08,
		CriticalRateThreshold: 0,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "disabled_high",
				Name:                "Disabled High",
				Kind:                "custom",
				Enabled:             false,
				BaseURL:             "https://example.com",
				PricingURL:          "https://example.com/high",
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 9.90,
			},
			{
				ID:                  "enabled_low",
				Name:                "Enabled Low",
				Kind:                "custom",
				Enabled:             true,
				BaseURL:             "https://example.com",
				PricingURL:          "https://example.com/low",
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 1.50,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{
			{
				ID:          "map_1",
				LocalGroup:  "VIP",
				ModelFamily: "gpt",
				SourceIDs:   []string{"disabled_high", "enabled_low"},
			},
		},
	}

	snapshot, err := svc.PreviewUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, snapshot.GroupRows, 2)
	var enabledRow *UpstreamMonitorPreviewGroupRow
	var disabledRow *UpstreamMonitorPreviewGroupRow
	for i := range snapshot.GroupRows {
		switch snapshot.GroupRows[i].SourceIDs[0] {
		case "enabled_low":
			enabledRow = &snapshot.GroupRows[i]
		case "disabled_high":
			disabledRow = &snapshot.GroupRows[i]
		}
	}
	require.NotNil(t, enabledRow)
	require.NotNil(t, disabledRow)
	require.InDelta(t, 1.50, enabledRow.ReferenceMultiplier, 0.0001)
	require.Equal(t, "healthy", enabledRow.Status)
	require.Equal(t, "unknown", disabledRow.Status)
	require.Contains(t, disabledRow.Issues, "no enabled source")
}

func TestSettingServicePreviewUpstreamMonitorConfig_BuildsAccountRiskRows(t *testing.T) {
	rate := 1.15
	svc := NewUpstreamMonitorService(newUpstreamMonitorSettingRepo())
	svc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{ID: 10, Name: "VIP", Platform: PlatformOpenAI, RateMultiplier: 1.30, Status: StatusActive},
		},
	})
	svc.SetAccountLister(upstreamMonitorTestAccountLister{
		accounts: []Account{
			{
				ID:             20,
				Name:           "OpenAI upstream A",
				Platform:       PlatformOpenAI,
				Type:           AccountTypeOAuth,
				Status:         StatusActive,
				RateMultiplier: &rate,
				GroupIDs:       []int64{10},
			},
		},
	})

	cfg := &UpstreamMonitorConfig{
		Enabled:                    true,
		AutoRefreshEnabled:         true,
		RefreshIntervalMinutes:     10,
		DefaultExchangeRate:        7.2,
		DefaultProfitRateThreshold: 0.15,
		WarningRateThreshold:       0.12,
		CriticalRateThreshold:      0.02,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "nbility",
				Name:                "Nbility",
				Kind:                "custom",
				Enabled:             true,
				AutoSyncEnabled:     false,
				AccountIDs:          []int64{20},
				FetchMode:           upstreamMonitorFetchModeAuto,
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 1.18,
				LastSyncStatus:      upstreamMonitorSyncStatusIdle,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{},
	}

	snapshot, err := svc.PreviewUpstreamMonitorConfig(context.Background(), cfg)
	require.NoError(t, err)
	require.Len(t, snapshot.AccountRows, 1)
	row := snapshot.AccountRows[0]
	require.Equal(t, int64(20), row.AccountID)
	require.Equal(t, "OpenAI upstream A", row.AccountName)
	require.Equal(t, []string{"VIP"}, row.GroupNames)
	require.InDelta(t, 1.15, row.AccountRateMultiplier, 0.0001)
	require.InDelta(t, 1.18, row.ReferenceMultiplier, 0.0001)
	require.InDelta(t, 1.18, row.EstimatedCostMultiplier, 0.0001)
	require.InDelta(t, (1.30-1.18)/1.30, row.EstimatedMarginRate, 0.0001)
	require.Equal(t, "warning", row.Status)
	require.Equal(t, 1, snapshot.Summary.MonitoredAccountCount)
	require.Len(t, snapshot.AccountOptions, 1)
}
