package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSettingHandler_RefreshUpstreamMonitorConfig_UsesStoredConfigWithoutBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	repo := &upstreamMonitorHandlerSettingRepo{values: map[string]string{}}
	stored := &service.UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []service.UpstreamMonitorSource{
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
		GroupMappings: []service.UpstreamMonitorGroupMap{},
	}
	raw, err := json.Marshal(stored)
	require.NoError(t, err)
	repo.values[service.SettingKeyUpstreamMonitorConfig] = string(raw)

	svc := service.NewSettingService(repo, &config.Config{})
	svc.SetUpstreamMonitorGroupLister(upstreamMonitorHandlerGroupLister{})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings/upstream-monitor/refresh", nil)

	handler.RefreshUpstreamMonitorConfig(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	persisted, err := svc.GetUpstreamMonitorConfig(ctx)
	require.NoError(t, err)
	require.Len(t, persisted.Sources, 1)
	require.Equal(t, "manual", persisted.Sources[0].ID)
	require.InDelta(t, 1.2, persisted.Sources[0].ReferenceMultiplier, 0.0001)
}

func TestSettingHandler_RefreshUpstreamMonitorConfig_AcceptsSourceIDBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &upstreamMonitorHandlerSettingRepo{values: map[string]string{}}
	stored := &service.UpstreamMonitorConfig{
		Enabled:                true,
		AutoRefreshEnabled:     true,
		RefreshIntervalMinutes: 10,
		DefaultExchangeRate:    7.2,
		Sources: []service.UpstreamMonitorSource{
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
		GroupMappings: []service.UpstreamMonitorGroupMap{},
	}
	raw, err := json.Marshal(stored)
	require.NoError(t, err)
	repo.values[service.SettingKeyUpstreamMonitorConfig] = string(raw)

	svc := service.NewSettingService(repo, &config.Config{})
	svc.SetUpstreamMonitorGroupLister(upstreamMonitorHandlerGroupLister{})
	handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

	body := bytes.NewBufferString(`{"source_id":"manual"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings/upstream-monitor/refresh", body)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.RefreshUpstreamMonitorConfig(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
}

type upstreamMonitorHandlerGroupLister struct{}

func (upstreamMonitorHandlerGroupLister) ListActive(context.Context) ([]service.Group, error) {
	return []service.Group{}, nil
}

type upstreamMonitorHandlerSettingRepo struct {
	values map[string]string
}

func (r *upstreamMonitorHandlerSettingRepo) Get(_ context.Context, key string) (*service.Setting, error) {
	value, ok := r.values[key]
	if !ok {
		return nil, service.ErrSettingNotFound
	}
	return &service.Setting{Key: key, Value: value}, nil
}

func (r *upstreamMonitorHandlerSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r.values[key]
	if !ok {
		return "", service.ErrSettingNotFound
	}
	return value, nil
}

func (r *upstreamMonitorHandlerSettingRepo) Set(_ context.Context, key, value string) error {
	if r.values == nil {
		r.values = map[string]string{}
	}
	r.values[key] = value
	return nil
}

func (r *upstreamMonitorHandlerSettingRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (r *upstreamMonitorHandlerSettingRepo) SetMultiple(_ context.Context, settings map[string]string) error {
	if r.values == nil {
		r.values = map[string]string{}
	}
	for key, value := range settings {
		r.values[key] = value
	}
	return nil
}

func (r *upstreamMonitorHandlerSettingRepo) GetAll(_ context.Context) (map[string]string, error) {
	out := make(map[string]string, len(r.values))
	for key, value := range r.values {
		out[key] = value
	}
	return out, nil
}

func (r *upstreamMonitorHandlerSettingRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}
