package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSettingServiceRunDueUpstreamMonitorRefresh_UsesStoredConfig(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusOK, "application/json", `{"data":{"reference_multiplier":1.42}}`)

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
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	repo := newUpstreamMonitorSettingRepo()
	repo.data[SettingKeyUpstreamMonitorConfig] = string(raw)

	svc := NewUpstreamMonitorService(repo)
	result, err := svc.RunDueUpstreamMonitorRefresh(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 1, result.Summary.AttemptedCount)
	require.Equal(t, 1, result.Summary.SuccessCount)
	require.InDelta(t, 1.42, result.Config.Sources[0].ReferenceMultiplier, 0.0001)
}

func TestSettingServiceRunDueUpstreamMonitorRefresh_SkipsWhenDisabled(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		Enabled:                false,
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
				FetchMode:       upstreamMonitorFetchModePlainText,
				BaseURL:         "https://example.com",
				PricingURL:      "https://example.com/pricing",
				AuthMode:        "none",
				Currency:        "CNY",
				ExchangeRate:    7.2,
			},
		},
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	repo := newUpstreamMonitorSettingRepo()
	repo.data[SettingKeyUpstreamMonitorConfig] = string(raw)

	svc := NewUpstreamMonitorService(repo)
	result, err := svc.RunDueUpstreamMonitorRefresh(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 0, result.Summary.AttemptedCount)
	require.Equal(t, string(raw), repo.data[SettingKeyUpstreamMonitorConfig])
}

func TestUpstreamMonitorRunnerRunOnce_RefreshesWhenLeader(t *testing.T) {
	stubUpstreamMonitorClient(t, http.StatusOK, "text/plain", "1.77")

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "plain",
				Name:            "Plain Upstream",
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
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	repo := newUpstreamMonitorSettingRepo()
	repo.data[SettingKeyUpstreamMonitorConfig] = string(raw)

	svc := NewUpstreamMonitorService(repo)
	runner := NewUpstreamMonitorRunner(svc, time.Minute)
	runner.SetLeaderLock(&fakeLeaderLockCache{}, nil)
	runner.runOnce()

	persisted := &UpstreamMonitorConfig{}
	require.NoError(t, json.Unmarshal([]byte(repo.data[SettingKeyUpstreamMonitorConfig]), persisted))
	require.InDelta(t, 1.77, persisted.Sources[0].ReferenceMultiplier, 0.0001)
	require.Equal(t, upstreamMonitorSyncStatusSuccess, persisted.Sources[0].LastSyncStatus)
}

func TestUpstreamMonitorRunnerRunOnce_SkipsWhenPeerHoldsLock(t *testing.T) {
	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "plain",
				Name:                "Plain Upstream",
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
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	repo := newUpstreamMonitorSettingRepo()
	repo.data[SettingKeyUpstreamMonitorConfig] = string(raw)

	lockCache := &fakeLeaderLockCache{}
	acquired, err := lockCache.TryAcquireLeaderLock(context.Background(), upstreamMonitorRefreshLeaderLockKey, "peer", time.Minute)
	require.NoError(t, err)
	require.True(t, acquired)

	svc := NewUpstreamMonitorService(repo)
	runner := NewUpstreamMonitorRunner(svc, time.Minute)
	runner.SetLeaderLock(lockCache, nil)
	runner.runOnce()

	persisted := &UpstreamMonitorConfig{}
	require.NoError(t, json.Unmarshal([]byte(repo.data[SettingKeyUpstreamMonitorConfig]), persisted))
	require.InDelta(t, 2.34, persisted.Sources[0].ReferenceMultiplier, 0.0001)
	require.Empty(t, persisted.Sources[0].LastSyncStatus)
}

func TestSettingServiceRunDueUpstreamMonitorRefresh_SendsUpstreamAlertEmail(t *testing.T) {
	ctx := context.Background()
	repo := newNotificationEmailMemorySettingRepo()
	repo.values[SettingKeySMTPHost] = "smtp.example.com"
	repo.values[SettingKeySMTPPort] = "587"
	repo.values[SettingKeySMTPUsername] = "demo"
	repo.values[SettingKeySMTPPassword] = "secret"
	repo.values[SettingKeySMTPFrom] = "noreply@example.com"
	repo.values[SettingKeyAccountQuotaNotifyEmails] = `[{"email":"ops@example.com","disabled":false,"verified":true}]`

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.4,
		CriticalRateThreshold:  0.2,
		NotifyOnCriticalOnly:   true,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "plain",
				Name:            "Plain Upstream",
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
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	repo.values[SettingKeyUpstreamMonitorConfig] = string(raw)

	emailSvc := NewEmailService(repo, nil)
	notificationSvc := NewNotificationEmailService(repo, emailSvc)
	upstreamMonitorSvc := NewUpstreamMonitorService(repo)
	upstreamMonitorSvc.SetNotificationEmailService(notificationSvc)
	upstreamMonitorSvc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{ID: 1, Name: "VIP", Platform: "openai", RateMultiplier: 2.0},
		},
	})
	var sent []string
	emailSvc.sendWithConfig = func(config *SMTPConfig, to, subject, body string) error {
		sent = append(sent, to+"\n"+subject+"\n"+body)
		return nil
	}

	runRefresh := func(body string) {
		repo.mu.Lock()
		if rawCfg, ok := repo.values[SettingKeyUpstreamMonitorConfig]; ok {
			stored := &UpstreamMonitorConfig{}
			require.NoError(t, json.Unmarshal([]byte(rawCfg), stored))
			for i := range stored.Sources {
				stored.Sources[i].LastSyncAt = nil
			}
			updated, err := json.Marshal(stored)
			require.NoError(t, err)
			repo.values[SettingKeyUpstreamMonitorConfig] = string(updated)
		}
		repo.mu.Unlock()
		stubUpstreamMonitorClient(t, http.StatusOK, "text/plain", body)
		result, err := upstreamMonitorSvc.RunDueUpstreamMonitorRefresh(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)
	}

	runRefresh("1.80")
	require.Len(t, sent, 0)
	_, err = repo.GetValue(ctx, upstreamMonitorAlertStateKey("map_1"))
	require.ErrorIs(t, err, ErrSettingNotFound)

	runRefresh("2.10")
	require.Len(t, sent, 1)
	require.Contains(t, sent[0], "ops@example.com")
	require.Contains(t, sent[0], "VIP")
	state, err := repo.GetValue(ctx, upstreamMonitorAlertStateKey("map_1"))
	require.NoError(t, err)
	require.Equal(t, "critical", state)
	values, err := repo.GetAll(ctx)
	require.NoError(t, err)
	for key := range values {
		require.NotContains(t, key, notificationEmailDeliveryKeyPrefix)
	}

	runRefresh("2.10")
	require.Len(t, sent, 1)

	runRefresh("1.00")
	require.Len(t, sent, 2)
	require.Contains(t, sent[1], "[change]")
	_, err = repo.GetValue(ctx, upstreamMonitorAlertStateKey("map_1"))
	require.ErrorIs(t, err, ErrSettingNotFound)

	runRefresh("2.10")
	require.Len(t, sent, 3)
	require.Contains(t, sent[2], "[critical]")
	state, err = repo.GetValue(ctx, upstreamMonitorAlertStateKey("map_1"))
	require.NoError(t, err)
	require.Equal(t, "critical", state)
}

func TestSettingServiceRunDueUpstreamMonitorRefresh_SendsEmailWhenUpstreamMultiplierChanges(t *testing.T) {
	ctx := context.Background()
	repo := newNotificationEmailMemorySettingRepo()
	repo.values[SettingKeySMTPHost] = "smtp.example.com"
	repo.values[SettingKeySMTPPort] = "587"
	repo.values[SettingKeySMTPUsername] = "demo"
	repo.values[SettingKeySMTPPassword] = "secret"
	repo.values[SettingKeySMTPFrom] = "noreply@example.com"
	repo.values[SettingKeyAccountQuotaNotifyEmails] = `[{"email":"ops@example.com","disabled":false,"verified":true}]`

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.4,
		CriticalRateThreshold:  0.2,
		NotifyOnCriticalOnly:   true,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "plain",
				Name:            "Plain Upstream",
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
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	repo.values[SettingKeyUpstreamMonitorConfig] = string(raw)

	emailSvc := NewEmailService(repo, nil)
	notificationSvc := NewNotificationEmailService(repo, emailSvc)
	upstreamMonitorSvc := NewUpstreamMonitorService(repo)
	upstreamMonitorSvc.SetNotificationEmailService(notificationSvc)
	upstreamMonitorSvc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{ID: 1, Name: "VIP", Platform: "openai", RateMultiplier: 2.0},
		},
	})
	var sent []string
	emailSvc.sendWithConfig = func(config *SMTPConfig, to, subject, body string) error {
		sent = append(sent, to+"\n"+subject+"\n"+body)
		return nil
	}

	runRefresh := func(body string) {
		repo.mu.Lock()
		if rawCfg, ok := repo.values[SettingKeyUpstreamMonitorConfig]; ok {
			stored := &UpstreamMonitorConfig{}
			require.NoError(t, json.Unmarshal([]byte(rawCfg), stored))
			for i := range stored.Sources {
				stored.Sources[i].LastSyncAt = nil
			}
			updated, err := json.Marshal(stored)
			require.NoError(t, err)
			repo.values[SettingKeyUpstreamMonitorConfig] = string(updated)
		}
		repo.mu.Unlock()
		stubUpstreamMonitorClient(t, http.StatusOK, "text/plain", body)
		result, err := upstreamMonitorSvc.RunDueUpstreamMonitorRefresh(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)
	}

	runRefresh("1.20")
	require.Len(t, sent, 0)
	state, err := repo.GetValue(ctx, upstreamMonitorMultiplierStateKey("map_1"))
	require.NoError(t, err)
	require.Equal(t, upstreamMonitorMultiplierState(1.20), state)

	runRefresh("1.30")
	require.Len(t, sent, 1)
	require.Contains(t, sent[0], "[change]")
	require.Contains(t, sent[0], "reference multiplier changed")
	state, err = repo.GetValue(ctx, upstreamMonitorMultiplierStateKey("map_1"))
	require.NoError(t, err)
	require.Equal(t, upstreamMonitorMultiplierState(1.30), state)

	runRefresh("1.30")
	require.Len(t, sent, 1)
}

func TestSettingServiceRunDueUpstreamMonitorRefresh_DoesNotEmailAccountWarningUnlessCostExceedsLocal(t *testing.T) {
	ctx := context.Background()
	repo := newNotificationEmailMemorySettingRepo()
	repo.values[SettingKeySMTPHost] = "smtp.example.com"
	repo.values[SettingKeySMTPPort] = "587"
	repo.values[SettingKeySMTPUsername] = "demo"
	repo.values[SettingKeySMTPPassword] = "secret"
	repo.values[SettingKeySMTPFrom] = "noreply@example.com"
	repo.values[SettingKeyAccountQuotaNotifyEmails] = `[{"email":"ops@example.com","disabled":false,"verified":true}]`

	rate := 1.15
	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.12,
		CriticalRateThreshold:  0.02,
		NotifyOnCriticalOnly:   true,
		Sources: []UpstreamMonitorSource{
			{
				ID:                  "manual",
				Name:                "Manual Upstream",
				Kind:                "manual",
				Enabled:             true,
				AutoSyncEnabled:     false,
				AccountIDs:          []int64{20},
				FetchMode:           upstreamMonitorFetchModeAuto,
				AuthMode:            "none",
				Currency:            "CNY",
				ExchangeRate:        7.2,
				ReferenceMultiplier: 1.18,
			},
		},
		GroupMappings: []UpstreamMonitorGroupMap{},
	}
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	repo.values[SettingKeyUpstreamMonitorConfig] = string(raw)

	emailSvc := NewEmailService(repo, nil)
	notificationSvc := NewNotificationEmailService(repo, emailSvc)
	upstreamMonitorSvc := NewUpstreamMonitorService(repo)
	upstreamMonitorSvc.SetNotificationEmailService(notificationSvc)
	upstreamMonitorSvc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{ID: 10, Name: "VIP", Platform: PlatformOpenAI, RateMultiplier: 1.30, Status: StatusActive},
		},
	})
	upstreamMonitorSvc.SetAccountLister(upstreamMonitorTestAccountLister{
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
	var sent []string
	emailSvc.sendWithConfig = func(config *SMTPConfig, to, subject, body string) error {
		sent = append(sent, to+"\n"+subject+"\n"+body)
		return nil
	}

	result, err := upstreamMonitorSvc.RunDueUpstreamMonitorRefresh(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, sent, 0)

	stored := &UpstreamMonitorConfig{}
	require.NoError(t, json.Unmarshal([]byte(repo.values[SettingKeyUpstreamMonitorConfig]), stored))
	stored.Sources[0].LastSyncAt = nil
	stored.Sources[0].ReferenceMultiplier = 1.40
	raw, err = json.Marshal(stored)
	require.NoError(t, err)
	repo.values[SettingKeyUpstreamMonitorConfig] = string(raw)

	result, err = upstreamMonitorSvc.RunDueUpstreamMonitorRefresh(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, sent, 1)
	require.Contains(t, sent[0], "[critical]")
	require.Contains(t, sent[0], "OpenAI upstream A")
}

func TestSettingServiceRunDueUpstreamMonitorRefresh_SendsWarningOnlyWhenConfigured(t *testing.T) {
	ctx := context.Background()

	buildService := func(notifyOnCriticalOnly bool) (*UpstreamMonitorService, *notificationEmailMemorySettingRepo, *[]string) {
		repo := newNotificationEmailMemorySettingRepo()
		repo.values[SettingKeySMTPHost] = "smtp.example.com"
		repo.values[SettingKeySMTPPort] = "587"
		repo.values[SettingKeySMTPUsername] = "demo"
		repo.values[SettingKeySMTPPassword] = "secret"
		repo.values[SettingKeySMTPFrom] = "noreply@example.com"
		repo.values[SettingKeyAccountQuotaNotifyEmails] = `[{"email":"ops@example.com","disabled":false,"verified":true}]`

		cfg := &UpstreamMonitorConfig{
			Enabled:                true,
			AutoRefreshEnabled:     true,
			RefreshIntervalMinutes: 10,
			DefaultExchangeRate:    7.2,
			WarningRateThreshold:   0.25,
			CriticalRateThreshold:  0,
			NotifyOnCriticalOnly:   notifyOnCriticalOnly,
			Sources: []UpstreamMonitorSource{
				{
					ID:              "plain",
					Name:            "Plain Upstream",
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
					ID:          "map_warning",
					LocalGroup:  "VIP",
					ModelFamily: "gpt",
					SourceIDs:   []string{"plain"},
				},
			},
		}
		raw, err := json.Marshal(cfg)
		require.NoError(t, err)
		repo.values[SettingKeyUpstreamMonitorConfig] = string(raw)

		emailSvc := NewEmailService(repo, nil)
		notificationSvc := NewNotificationEmailService(repo, emailSvc)
		upstreamMonitorSvc := NewUpstreamMonitorService(repo)
		upstreamMonitorSvc.SetNotificationEmailService(notificationSvc)
		upstreamMonitorSvc.SetGroupLister(upstreamMonitorTestGroupLister{
			groups: []Group{
				{ID: 1, Name: "VIP", Platform: PlatformOpenAI, RateMultiplier: 2.0},
			},
		})
		sent := []string{}
		emailSvc.sendWithConfig = func(config *SMTPConfig, to, subject, body string) error {
			sent = append(sent, to+"\n"+subject+"\n"+body)
			return nil
		}
		return upstreamMonitorSvc, repo, &sent
	}

	runRefresh := func(upstreamMonitorSvc *UpstreamMonitorService) {
		stubUpstreamMonitorClient(t, http.StatusOK, "text/plain", "1.60")
		result, err := upstreamMonitorSvc.RunDueUpstreamMonitorRefresh(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)
	}

	criticalOnlySvc, criticalOnlyRepo, criticalOnlySent := buildService(true)
	runRefresh(criticalOnlySvc)
	require.Len(t, *criticalOnlySent, 0)
	_, err := criticalOnlyRepo.GetValue(ctx, upstreamMonitorAlertStateKey("map_warning"))
	require.ErrorIs(t, err, ErrSettingNotFound)

	warningSvc, warningRepo, warningSent := buildService(false)
	runRefresh(warningSvc)
	require.Len(t, *warningSent, 1)
	require.Contains(t, (*warningSent)[0], "[warning]")
	state, err := warningRepo.GetValue(ctx, upstreamMonitorAlertStateKey("map_warning"))
	require.NoError(t, err)
	require.Equal(t, "warning", state)
}

func TestSettingServiceRunDueUpstreamMonitorRefresh_RetriesUpstreamAlertWhenEmailFails(t *testing.T) {
	ctx := context.Background()
	repo := newNotificationEmailMemorySettingRepo()
	repo.values[SettingKeySMTPHost] = "smtp.example.com"
	repo.values[SettingKeySMTPPort] = "587"
	repo.values[SettingKeySMTPUsername] = "demo"
	repo.values[SettingKeySMTPPassword] = "secret"
	repo.values[SettingKeySMTPFrom] = "noreply@example.com"
	repo.values[SettingKeyAccountQuotaNotifyEmails] = `[{"email":"ops@example.com","disabled":false,"verified":true}]`

	cfg := &UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		WarningRateThreshold:   0.4,
		CriticalRateThreshold:  0.2,
		NotifyOnCriticalOnly:   true,
		Sources: []UpstreamMonitorSource{
			{
				ID:              "plain",
				Name:            "Plain Upstream",
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
	raw, err := json.Marshal(cfg)
	require.NoError(t, err)
	repo.values[SettingKeyUpstreamMonitorConfig] = string(raw)

	emailSvc := NewEmailService(repo, nil)
	notificationSvc := NewNotificationEmailService(repo, emailSvc)
	upstreamMonitorSvc := NewUpstreamMonitorService(repo)
	upstreamMonitorSvc.SetNotificationEmailService(notificationSvc)
	upstreamMonitorSvc.SetGroupLister(upstreamMonitorTestGroupLister{
		groups: []Group{
			{ID: 1, Name: "VIP", Platform: "openai", RateMultiplier: 2.0},
		},
	})

	attempts := 0
	emailSvc.sendWithConfig = func(config *SMTPConfig, to, subject, body string) error {
		attempts++
		if attempts == 1 {
			return errors.New("smtp temporarily unavailable")
		}
		return nil
	}

	runRefresh := func() {
		repo.mu.Lock()
		if rawCfg, ok := repo.values[SettingKeyUpstreamMonitorConfig]; ok {
			stored := &UpstreamMonitorConfig{}
			require.NoError(t, json.Unmarshal([]byte(rawCfg), stored))
			for i := range stored.Sources {
				stored.Sources[i].LastSyncAt = nil
			}
			updated, err := json.Marshal(stored)
			require.NoError(t, err)
			repo.values[SettingKeyUpstreamMonitorConfig] = string(updated)
		}
		repo.mu.Unlock()
		stubUpstreamMonitorClient(t, http.StatusOK, "text/plain", "2.10")
		result, err := upstreamMonitorSvc.RunDueUpstreamMonitorRefresh(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)
	}

	runRefresh()
	require.Equal(t, 1, attempts)
	_, err = repo.GetValue(ctx, upstreamMonitorAlertStateKey("map_1"))
	require.ErrorIs(t, err, ErrSettingNotFound)

	runRefresh()
	require.Equal(t, 2, attempts)
	state, err := repo.GetValue(ctx, upstreamMonitorAlertStateKey("map_1"))
	require.NoError(t, err)
	require.Equal(t, "critical", state)
}
