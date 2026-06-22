package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

type UpstreamMonitorConfig struct {
	Enabled                    bool                      `json:"enabled"`
	AutoRefreshEnabled         bool                      `json:"auto_refresh_enabled"`
	RefreshIntervalMinutes     int                       `json:"refresh_interval_minutes"`
	DefaultExchangeRate        float64                   `json:"default_exchange_rate"`
	DefaultProfitRateThreshold float64                   `json:"default_profit_rate_threshold"`
	WarningRateThreshold       float64                   `json:"warning_rate_threshold"`
	CriticalRateThreshold      float64                   `json:"critical_rate_threshold"`
	NotifyOnCriticalOnly       bool                      `json:"notify_on_critical_only"`
	Sources                    []UpstreamMonitorSource   `json:"sources"`
	GroupMappings              []UpstreamMonitorGroupMap `json:"group_mappings"`
}

func (cfg *UpstreamMonitorConfig) UnmarshalJSON(data []byte) error {
	type Alias UpstreamMonitorConfig
	aux := Alias{
		NotifyOnCriticalOnly: true,
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*cfg = UpstreamMonitorConfig(aux)
	return nil
}

type UpstreamMonitorSource struct {
	ID                   string                               `json:"id"`
	Name                 string                               `json:"name"`
	Kind                 string                               `json:"kind"`
	Enabled              bool                                 `json:"enabled"`
	AutoSyncEnabled      bool                                 `json:"auto_sync_enabled"`
	AccountIDs           []int64                              `json:"account_ids"`
	FetchMode            string                               `json:"fetch_mode"`
	BaseURL              string                               `json:"base_url"`
	PricingURL           string                               `json:"pricing_url"`
	PricingPathHint      string                               `json:"pricing_path_hint"`
	AuthMode             string                               `json:"auth_mode"`
	AuthHeaderName       string                               `json:"auth_header_name"`
	AuthToken            string                               `json:"auth_token,omitempty"`
	AuthConfigured       bool                                 `json:"auth_configured"`
	Currency             string                               `json:"currency"`
	ExchangeRate         float64                              `json:"exchange_rate"`
	ReferenceMultiplier  float64                              `json:"reference_multiplier"`
	UpstreamGroupOptions []UpstreamMonitorUpstreamGroupOption `json:"upstream_group_options"`
	LastSyncAt           *time.Time                           `json:"last_sync_at,omitempty"`
	LastSyncStatus       string                               `json:"last_sync_status"`
	LastSyncError        string                               `json:"last_sync_error"`
	Notes                string                               `json:"notes"`
}

func (s *UpstreamMonitorSource) UnmarshalJSON(data []byte) error {
	type Alias UpstreamMonitorSource
	aux := struct {
		LastSyncAt json.RawMessage `json:"last_sync_at"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	aux.Alias.LastSyncAt = nil
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.LastSyncAt) == 0 || string(aux.LastSyncAt) == "null" {
		s.LastSyncAt = nil
		return nil
	}
	var raw string
	if err := json.Unmarshal(aux.LastSyncAt, &raw); err == nil {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			s.LastSyncAt = nil
			return nil
		}
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return err
		}
		s.LastSyncAt = &parsed
		return nil
	}
	var parsed time.Time
	if err := json.Unmarshal(aux.LastSyncAt, &parsed); err != nil {
		return err
	}
	s.LastSyncAt = &parsed
	return nil
}

type UpstreamMonitorGroupMap struct {
	ID                  string   `json:"id"`
	LocalGroupID        int64    `json:"local_group_id"`
	LocalGroup          string   `json:"local_group"`
	UpstreamGroupKey    string   `json:"upstream_group_key"`
	UpstreamGroup       string   `json:"upstream_group"`
	ModelFamily         string   `json:"model_family"`
	SourceIDs           []string `json:"source_ids"`
	ReferenceMultiplier float64  `json:"reference_multiplier"`
	Notes               string   `json:"notes"`
}

type UpstreamMonitorUpstreamGroupOption struct {
	Key                 string  `json:"key"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	ReferenceMultiplier float64 `json:"reference_multiplier"`
	RawID               string  `json:"raw_id"`
	Path                string  `json:"path"`
}

type upstreamMonitorGroupLister interface {
	ListActive(ctx context.Context) ([]Group, error)
}

type upstreamMonitorAccountLister interface {
	ListActive(ctx context.Context) ([]Account, error)
}

type UpstreamMonitorPreviewSnapshot struct {
	GeneratedAt    time.Time                           `json:"generated_at"`
	Summary        UpstreamMonitorPreviewSummary       `json:"summary"`
	SourceRows     []UpstreamMonitorPreviewSourceRow   `json:"source_rows"`
	GroupRows      []UpstreamMonitorPreviewGroupRow    `json:"group_rows"`
	AccountRows    []UpstreamMonitorPreviewAccountRow  `json:"account_rows"`
	AccountOptions []UpstreamMonitorPreviewAccountInfo `json:"account_options"`
	GroupOptions   []UpstreamMonitorPreviewGroupOption `json:"group_options"`
	UnmappedGroups []UpstreamMonitorPreviewUnmappedRow `json:"unmapped_groups"`
}

type UpstreamMonitorPreviewSummary struct {
	Enabled               bool    `json:"enabled"`
	AutoRefreshEnabled    bool    `json:"auto_refresh_enabled"`
	SourceCount           int     `json:"source_count"`
	EnabledSourceCount    int     `json:"enabled_source_count"`
	MappedGroupCount      int     `json:"mapped_group_count"`
	MonitoredAccountCount int     `json:"monitored_account_count"`
	UnmappedGroupCount    int     `json:"unmapped_group_count"`
	HealthyCount          int     `json:"healthy_count"`
	WarningCount          int     `json:"warning_count"`
	CriticalCount         int     `json:"critical_count"`
	UnknownCount          int     `json:"unknown_count"`
	AverageMarginRate     float64 `json:"average_margin_rate"`
	LowestMarginRate      float64 `json:"lowest_margin_rate"`
	HighestMarginRate     float64 `json:"highest_margin_rate"`
}

type UpstreamMonitorPreviewSourceRow struct {
	ID                   string                               `json:"id"`
	Name                 string                               `json:"name"`
	Enabled              bool                                 `json:"enabled"`
	AutoSyncEnabled      bool                                 `json:"auto_sync_enabled"`
	Kind                 string                               `json:"kind"`
	FetchMode            string                               `json:"fetch_mode"`
	Currency             string                               `json:"currency"`
	ExchangeRate         float64                              `json:"exchange_rate"`
	ReferenceMultiplier  float64                              `json:"reference_multiplier"`
	AccountIDs           []int64                              `json:"account_ids"`
	AccountCount         int                                  `json:"account_count"`
	UpstreamGroupOptions []UpstreamMonitorUpstreamGroupOption `json:"upstream_group_options"`
	AuthConfigured       bool                                 `json:"auth_configured"`
	PricingURL           string                               `json:"pricing_url"`
	LastSyncAt           *time.Time                           `json:"last_sync_at,omitempty"`
	LastSyncStatus       string                               `json:"last_sync_status"`
	LastSyncError        string                               `json:"last_sync_error"`
	Notes                string                               `json:"notes"`
}

type UpstreamMonitorPreviewAccountInfo struct {
	AccountID      int64    `json:"account_id"`
	AccountName    string   `json:"account_name"`
	Platform       string   `json:"platform"`
	Type           string   `json:"type"`
	RateMultiplier float64  `json:"rate_multiplier"`
	Status         string   `json:"status"`
	GroupIDs       []int64  `json:"group_ids"`
	GroupNames     []string `json:"group_names"`
}

type UpstreamMonitorPreviewGroupOption struct {
	GroupID          int64   `json:"group_id"`
	GroupName        string  `json:"group_name"`
	Platform         string  `json:"platform"`
	Multiplier       float64 `json:"multiplier"`
	IsExclusive      bool    `json:"is_exclusive"`
	SubscriptionType string  `json:"subscription_type"`
}

type UpstreamMonitorPreviewAccountRow struct {
	SourceID                string    `json:"source_id"`
	SourceName              string    `json:"source_name"`
	AccountID               int64     `json:"account_id"`
	AccountName             string    `json:"account_name"`
	AccountPlatform         string    `json:"account_platform"`
	AccountType             string    `json:"account_type"`
	AccountRateMultiplier   float64   `json:"account_rate_multiplier"`
	GroupIDs                []int64   `json:"group_ids"`
	GroupNames              []string  `json:"group_names"`
	GroupMultipliers        []float64 `json:"group_multipliers"`
	HighestGroupMultiplier  float64   `json:"highest_group_multiplier"`
	ReferenceMultiplier     float64   `json:"reference_multiplier"`
	EstimatedCostMultiplier float64   `json:"estimated_cost_multiplier"`
	EstimatedMarginRate     float64   `json:"estimated_margin_rate"`
	Status                  string    `json:"status"`
	Issues                  []string  `json:"issues"`
}

type UpstreamMonitorPreviewGroupRow struct {
	MappingID            string   `json:"mapping_id"`
	LocalGroup           string   `json:"local_group"`
	UpstreamGroupKey     string   `json:"upstream_group_key"`
	UpstreamGroup        string   `json:"upstream_group"`
	LocalGroupID         int64    `json:"local_group_id"`
	LocalGroupPlatform   string   `json:"local_group_platform"`
	LocalGroupMultiplier float64  `json:"local_group_multiplier"`
	ModelFamily          string   `json:"model_family"`
	SourceIDs            []string `json:"source_ids"`
	SourceNames          []string `json:"source_names"`
	SourceCount          int      `json:"source_count"`
	EnabledSourceCount   int      `json:"enabled_source_count"`
	ReferenceMultiplier  float64  `json:"reference_multiplier"`
	MappingMultiplier    float64  `json:"mapping_multiplier"`
	EstimatedMarginRate  float64  `json:"estimated_margin_rate"`
	Status               string   `json:"status"`
	Issues               []string `json:"issues"`
	Notes                string   `json:"notes"`
}

type UpstreamMonitorPreviewUnmappedRow struct {
	GroupID          int64   `json:"group_id"`
	GroupName        string  `json:"group_name"`
	Platform         string  `json:"platform"`
	Multiplier       float64 `json:"multiplier"`
	IsExclusive      bool    `json:"is_exclusive"`
	SubscriptionType string  `json:"subscription_type"`
}

type UpstreamMonitorRefreshSummary struct {
	AttemptedCount int `json:"attempted_count"`
	SuccessCount   int `json:"success_count"`
	FailedCount    int `json:"failed_count"`
	SkippedCount   int `json:"skipped_count"`
}

type UpstreamMonitorRefreshResult struct {
	RefreshedAt time.Time                       `json:"refreshed_at"`
	Summary     UpstreamMonitorRefreshSummary   `json:"summary"`
	Config      *UpstreamMonitorConfig          `json:"config,omitempty"`
	Snapshot    *UpstreamMonitorPreviewSnapshot `json:"snapshot,omitempty"`
}

type upstreamMonitorPricingSnapshot struct {
	ReferenceMultiplier float64
	HasReference        bool
	GroupMultipliers    map[string]float64
	GroupOptions        []UpstreamMonitorUpstreamGroupOption
}

type upstreamMonitorAlertRecipient struct {
	Email string
	Name  string
}

type upstreamMonitorRefreshOptions struct {
	force           bool
	persist         bool
	includeSnapshot bool
	sourceID        string
}

const (
	upstreamMonitorFetchModeAuto      = "auto"
	upstreamMonitorFetchModeJSONPath  = "json_path"
	upstreamMonitorFetchModePlainText = "plain_text"

	upstreamMonitorSyncStatusIdle    = "idle"
	upstreamMonitorSyncStatusSuccess = "success"
	upstreamMonitorSyncStatusError   = "error"
)

var newUpstreamMonitorHTTPClient = func() *req.Client {
	return req.C().SetTimeout(15 * time.Second)
}

func defaultUpstreamMonitorConfig() *UpstreamMonitorConfig {
	return &UpstreamMonitorConfig{
		Enabled:                    false,
		AutoRefreshEnabled:         true,
		RefreshIntervalMinutes:     10,
		DefaultExchangeRate:        7.2,
		DefaultProfitRateThreshold: 0.15,
		WarningRateThreshold:       0.08,
		CriticalRateThreshold:      0,
		NotifyOnCriticalOnly:       true,
		Sources:                    []UpstreamMonitorSource{},
		GroupMappings:              []UpstreamMonitorGroupMap{},
	}
}

func normalizeUpstreamMonitorConfig(cfg *UpstreamMonitorConfig) {
	if cfg == nil {
		return
	}
	if cfg.RefreshIntervalMinutes <= 0 {
		cfg.RefreshIntervalMinutes = 10
	}
	if cfg.DefaultExchangeRate <= 0 {
		cfg.DefaultExchangeRate = 7.2
	}
	if cfg.DefaultProfitRateThreshold < 0 {
		cfg.DefaultProfitRateThreshold = 0
	}
	if cfg.WarningRateThreshold < 0 {
		cfg.WarningRateThreshold = 0
	}
	if cfg.WarningRateThreshold == 0 && cfg.DefaultProfitRateThreshold > 0 {
		cfg.WarningRateThreshold = cfg.DefaultProfitRateThreshold
	}
	if cfg.CriticalRateThreshold < 0 {
		cfg.CriticalRateThreshold = 0
	}
	if cfg.WarningRateThreshold < cfg.CriticalRateThreshold {
		cfg.WarningRateThreshold = cfg.CriticalRateThreshold
	}
	if cfg.Sources == nil {
		cfg.Sources = []UpstreamMonitorSource{}
	}
	if cfg.GroupMappings == nil {
		cfg.GroupMappings = []UpstreamMonitorGroupMap{}
	}

	for i := range cfg.Sources {
		src := &cfg.Sources[i]
		src.ID = strings.TrimSpace(src.ID)
		src.Name = strings.TrimSpace(src.Name)
		src.Kind = normalizeUpstreamSourceKind(src.Kind)
		src.FetchMode = normalizeUpstreamFetchMode(src.FetchMode)
		src.BaseURL = strings.TrimSpace(src.BaseURL)
		src.PricingURL = strings.TrimSpace(src.PricingURL)
		if src.PricingURL == "" {
			src.PricingURL = defaultUpstreamPricingURL(src.Kind, src.BaseURL)
		}
		src.PricingPathHint = strings.TrimSpace(src.PricingPathHint)
		src.AuthMode = normalizeUpstreamSourceAuthMode(src.AuthMode)
		src.AuthHeaderName = strings.TrimSpace(src.AuthHeaderName)
		src.Currency = normalizeUpstreamCurrency(src.Currency)
		src.AccountIDs = cleanUpstreamMonitorInt64IDs(src.AccountIDs)
		src.UpstreamGroupOptions = normalizeUpstreamGroupOptions(src.UpstreamGroupOptions)
		src.LastSyncStatus = normalizeUpstreamSyncStatus(src.LastSyncStatus)
		src.LastSyncError = strings.TrimSpace(src.LastSyncError)
		src.Notes = strings.TrimSpace(src.Notes)
		if src.ExchangeRate <= 0 {
			src.ExchangeRate = cfg.DefaultExchangeRate
		}
		if src.ReferenceMultiplier < 0 {
			src.ReferenceMultiplier = 0
		}
		src.AuthConfigured = strings.TrimSpace(src.AuthToken) != ""
	}

	for i := range cfg.GroupMappings {
		mapping := &cfg.GroupMappings[i]
		mapping.ID = strings.TrimSpace(mapping.ID)
		if mapping.LocalGroupID < 0 {
			mapping.LocalGroupID = 0
		}
		mapping.LocalGroup = strings.TrimSpace(mapping.LocalGroup)
		mapping.UpstreamGroupKey = strings.TrimSpace(mapping.UpstreamGroupKey)
		mapping.UpstreamGroup = strings.TrimSpace(mapping.UpstreamGroup)
		if mapping.UpstreamGroup == "" {
			mapping.UpstreamGroup = mapping.LocalGroup
		}
		mapping.ModelFamily = normalizeUpstreamModelFamily(mapping.ModelFamily)
		if mapping.ReferenceMultiplier < 0 {
			mapping.ReferenceMultiplier = 0
		}
		mapping.Notes = strings.TrimSpace(mapping.Notes)
		if mapping.SourceIDs == nil {
			mapping.SourceIDs = []string{}
			continue
		}
		cleaned := make([]string, 0, len(mapping.SourceIDs))
		seen := make(map[string]struct{}, len(mapping.SourceIDs))
		for _, rawID := range mapping.SourceIDs {
			id := strings.TrimSpace(rawID)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			cleaned = append(cleaned, id)
		}
		mapping.SourceIDs = cleaned
	}
	cfg.GroupMappings = splitUpstreamMonitorGroupMappingsBySource(cfg.GroupMappings)
}

func splitUpstreamMonitorGroupMappingsBySource(values []UpstreamMonitorGroupMap) []UpstreamMonitorGroupMap {
	if len(values) == 0 {
		return []UpstreamMonitorGroupMap{}
	}
	out := make([]UpstreamMonitorGroupMap, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, mapping := range values {
		sourceIDs := mapping.SourceIDs
		if len(sourceIDs) == 0 {
			out = append(out, mapping)
			continue
		}
		for index, sourceID := range sourceIDs {
			sourceID = strings.TrimSpace(sourceID)
			if sourceID == "" {
				continue
			}
			item := mapping
			item.SourceIDs = []string{sourceID}
			baseID := strings.TrimSpace(item.ID)
			if baseID == "" {
				baseID = "mapping"
			}
			if len(sourceIDs) > 1 {
				item.ID = upstreamMonitorSourceMappingID(baseID, sourceID)
			} else if index == 0 {
				item.ID = baseID
			}
			dedupeKey := strings.Join([]string{
				strconv.FormatInt(item.LocalGroupID, 10),
				strings.ToLower(strings.TrimSpace(item.LocalGroup)),
				strings.ToLower(strings.TrimSpace(item.UpstreamGroupKey)),
				strings.ToLower(strings.TrimSpace(item.UpstreamGroup)),
				sourceID,
			}, "|")
			if _, ok := seen[dedupeKey]; ok {
				continue
			}
			seen[dedupeKey] = struct{}{}
			out = append(out, item)
		}
	}
	if out == nil {
		return []UpstreamMonitorGroupMap{}
	}
	return out
}

func upstreamMonitorSourceMappingID(baseID, sourceID string) string {
	baseID = strings.TrimSpace(baseID)
	sourceID = strings.TrimSpace(sourceID)
	if sourceID == "" {
		return baseID
	}
	if strings.HasSuffix(baseID, "__"+sourceID) {
		return baseID
	}
	return baseID + "__" + sourceID
}

func validateUpstreamMonitorConfig(cfg *UpstreamMonitorConfig) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if cfg.RefreshIntervalMinutes < 1 || cfg.RefreshIntervalMinutes > 1440 {
		return fmt.Errorf("refresh_interval_minutes must be between 1 and 1440")
	}
	if cfg.DefaultExchangeRate <= 0 || cfg.DefaultExchangeRate > 1000 {
		return fmt.Errorf("default_exchange_rate must be between 0 and 1000")
	}
	if cfg.DefaultProfitRateThreshold < 0 || cfg.DefaultProfitRateThreshold > 100 {
		return fmt.Errorf("default_profit_rate_threshold must be between 0 and 100")
	}
	if cfg.WarningRateThreshold < 0 || cfg.WarningRateThreshold > 100 {
		return fmt.Errorf("warning_rate_threshold must be between 0 and 100")
	}
	if cfg.CriticalRateThreshold < 0 || cfg.CriticalRateThreshold > 100 {
		return fmt.Errorf("critical_rate_threshold must be between 0 and 100")
	}
	if cfg.WarningRateThreshold < cfg.CriticalRateThreshold {
		return fmt.Errorf("warning_rate_threshold must be greater than or equal to critical_rate_threshold")
	}
	if len(cfg.Sources) > 50 {
		return fmt.Errorf("too many sources (max 50)")
	}
	if len(cfg.GroupMappings) > 100 {
		return fmt.Errorf("too many group mappings (max 100)")
	}

	sourceIDs := make(map[string]struct{}, len(cfg.Sources))
	for i, src := range cfg.Sources {
		if src.ID == "" {
			return fmt.Errorf("source[%d]: id is required", i)
		}
		if src.Name == "" {
			return fmt.Errorf("source[%d]: name is required", i)
		}
		if src.Kind == "" {
			return fmt.Errorf("source[%d]: kind is required", i)
		}
		switch src.FetchMode {
		case upstreamMonitorFetchModeAuto, upstreamMonitorFetchModeJSONPath, upstreamMonitorFetchModePlainText:
		default:
			return fmt.Errorf("source[%d]: unsupported fetch_mode %q", i, src.FetchMode)
		}
		if src.Enabled && src.AutoSyncEnabled && src.PricingURL == "" {
			return fmt.Errorf("source[%d]: pricing_url is required", i)
		}
		if src.ExchangeRate <= 0 || src.ExchangeRate > 1000 {
			return fmt.Errorf("source[%d]: exchange_rate must be between 0 and 1000", i)
		}
		if src.ReferenceMultiplier < 0 || src.ReferenceMultiplier > 1000 {
			return fmt.Errorf("source[%d]: reference_multiplier must be between 0 and 1000", i)
		}
		for _, accountID := range src.AccountIDs {
			if accountID <= 0 {
				return fmt.Errorf("source[%d]: account_ids must be positive", i)
			}
		}
		if src.Enabled && src.FetchMode == upstreamMonitorFetchModeJSONPath && src.AutoSyncEnabled && src.PricingPathHint == "" {
			return fmt.Errorf("source[%d]: pricing_path_hint is required for json_path fetch mode", i)
		}
		if src.AuthMode == "header" && src.AuthHeaderName == "" {
			src.AuthHeaderName = http.CanonicalHeaderKey("Authorization")
		}
		if _, exists := sourceIDs[src.ID]; exists {
			return fmt.Errorf("source[%d]: duplicate id %q", i, src.ID)
		}
		sourceIDs[src.ID] = struct{}{}
	}

	groupMapIDs := make(map[string]struct{}, len(cfg.GroupMappings))
	for i, mapping := range cfg.GroupMappings {
		if mapping.ID == "" {
			return fmt.Errorf("group_mappings[%d]: id is required", i)
		}
		if mapping.LocalGroup == "" {
			return fmt.Errorf("group_mappings[%d]: local_group is required", i)
		}
		if mapping.LocalGroupID < 0 {
			return fmt.Errorf("group_mappings[%d]: local_group_id must be non-negative", i)
		}
		if mapping.ModelFamily == "" {
			return fmt.Errorf("group_mappings[%d]: model_family is required", i)
		}
		if mapping.ReferenceMultiplier < 0 || mapping.ReferenceMultiplier > 1000 {
			return fmt.Errorf("group_mappings[%d]: reference_multiplier must be between 0 and 1000", i)
		}
		if len(mapping.SourceIDs) == 0 {
			return fmt.Errorf("group_mappings[%d]: at least one source_id is required", i)
		}
		if len(mapping.SourceIDs) > 1 {
			return fmt.Errorf("group_mappings[%d]: only one source_id is allowed per upstream group mapping", i)
		}
		if _, exists := groupMapIDs[mapping.ID]; exists {
			return fmt.Errorf("group_mappings[%d]: duplicate id %q", i, mapping.ID)
		}
		groupMapIDs[mapping.ID] = struct{}{}
		for _, sourceID := range mapping.SourceIDs {
			if _, ok := sourceIDs[sourceID]; !ok {
				return fmt.Errorf("group_mappings[%d]: source_id %q not found", i, sourceID)
			}
		}
	}

	return nil
}

func normalizeUpstreamSourceKind(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "manual":
		return "manual"
	case "newapi":
		return "newapi"
	case "sub2api":
		return "sub2api"
	case "openai_compatible", "openai-compatible":
		return "openai_compatible"
	case "custom":
		return "custom"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func normalizeUpstreamFetchMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", upstreamMonitorFetchModeAuto:
		return upstreamMonitorFetchModeAuto
	case upstreamMonitorFetchModeJSONPath, "json":
		return upstreamMonitorFetchModeJSONPath
	case upstreamMonitorFetchModePlainText, "text", "number":
		return upstreamMonitorFetchModePlainText
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func normalizeUpstreamSourceAuthMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "bearer":
		return "bearer"
	case "header":
		return "header"
	case "cookie":
		return "cookie"
	case "none", "":
		return "none"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func defaultUpstreamPricingURL(kind, baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}
	if !strings.Contains(baseURL, "://") {
		baseURL = "https://" + baseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if parsed, err := url.Parse(baseURL); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		baseURL = parsed.Scheme + "://" + parsed.Host
	}

	switch normalizeUpstreamSourceKind(kind) {
	case "sub2api":
		return baseURL + "/api/v1/channels/available"
	case "newapi":
		return baseURL + "/api/ratio_config"
	default:
		return ""
	}
}

func normalizeUpstreamCurrency(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "USD":
		return "USD"
	case "CNY":
		return "CNY"
	default:
		return strings.ToUpper(strings.TrimSpace(raw))
	}
}

func normalizeUpstreamSyncStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", upstreamMonitorSyncStatusIdle:
		return upstreamMonitorSyncStatusIdle
	case upstreamMonitorSyncStatusSuccess:
		return upstreamMonitorSyncStatusSuccess
	case upstreamMonitorSyncStatusError:
		return upstreamMonitorSyncStatusError
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func normalizeUpstreamModelFamily(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "gpt", "openai":
		return "gpt"
	case "claude", "anthropic":
		return "claude"
	case "gemini", "google":
		return "gemini"
	case "deepseek":
		return "deepseek"
	case "mixed", "all":
		return "mixed"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func (s *SettingService) GetUpstreamMonitorConfig(ctx context.Context) (*UpstreamMonitorConfig, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyUpstreamMonitorConfig)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			cfg := defaultUpstreamMonitorConfig()
			data, mErr := json.Marshal(cfg)
			if mErr == nil {
				_ = s.settingRepo.Set(ctx, SettingKeyUpstreamMonitorConfig, string(data))
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("get upstream monitor config: %w", err)
	}
	cfg := defaultUpstreamMonitorConfig()
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), cfg); err != nil {
			return nil, fmt.Errorf("parse upstream monitor config: %w", err)
		}
	}
	normalizeUpstreamMonitorConfig(cfg)
	maskUpstreamMonitorSecrets(cfg)
	return cfg, nil
}

func (s *SettingService) SaveUpstreamMonitorConfig(ctx context.Context, cfg *UpstreamMonitorConfig) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	s.mergeExistingUpstreamMonitorSecrets(ctx, cfg)
	normalizeUpstreamMonitorConfig(cfg)
	if err := validateUpstreamMonitorConfig(cfg); err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal upstream monitor config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyUpstreamMonitorConfig, string(data)); err != nil {
		return fmt.Errorf("save upstream monitor config: %w", err)
	}
	return nil
}

func (s *SettingService) RefreshUpstreamMonitorConfig(ctx context.Context, cfg *UpstreamMonitorConfig) (*UpstreamMonitorRefreshResult, error) {
	result, err := s.refreshUpstreamMonitorConfig(ctx, cfg, upstreamMonitorRefreshOptions{
		force:           true,
		persist:         true,
		includeSnapshot: true,
	})
	if err != nil {
		return nil, err
	}
	if result != nil && result.Config != nil {
		maskUpstreamMonitorSecrets(result.Config)
	}
	return result, nil
}

func (s *SettingService) RefreshStoredUpstreamMonitorConfig(ctx context.Context) (*UpstreamMonitorRefreshResult, error) {
	result, err := s.refreshUpstreamMonitorConfig(ctx, nil, upstreamMonitorRefreshOptions{
		force:           true,
		persist:         true,
		includeSnapshot: true,
	})
	if err != nil {
		return nil, err
	}
	if result != nil && result.Config != nil {
		maskUpstreamMonitorSecrets(result.Config)
	}
	return result, nil
}

func (s *SettingService) RefreshStoredUpstreamMonitorSource(ctx context.Context, sourceID string) (*UpstreamMonitorRefreshResult, error) {
	result, err := s.refreshUpstreamMonitorConfig(ctx, nil, upstreamMonitorRefreshOptions{
		force:           true,
		persist:         true,
		includeSnapshot: true,
		sourceID:        strings.TrimSpace(sourceID),
	})
	if err != nil {
		return nil, err
	}
	if result != nil && result.Config != nil {
		maskUpstreamMonitorSecrets(result.Config)
	}
	return result, nil
}

func (s *SettingService) RunDueUpstreamMonitorRefresh(ctx context.Context) (*UpstreamMonitorRefreshResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg, err := s.getUpstreamMonitorConfigRaw(ctx)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return &UpstreamMonitorRefreshResult{
				RefreshedAt: time.Now().UTC(),
				Summary:     UpstreamMonitorRefreshSummary{},
				Config:      defaultUpstreamMonitorConfig(),
			}, nil
		}
		return nil, fmt.Errorf("get upstream monitor config: %w", err)
	}
	if !cfg.Enabled || !cfg.AutoRefreshEnabled {
		return &UpstreamMonitorRefreshResult{
			RefreshedAt: time.Now().UTC(),
			Summary:     UpstreamMonitorRefreshSummary{},
			Config:      cfg,
		}, nil
	}

	return s.refreshUpstreamMonitorConfig(ctx, cfg, upstreamMonitorRefreshOptions{
		force:           false,
		persist:         true,
		includeSnapshot: false,
	})
}

func (s *SettingService) refreshUpstreamMonitorConfig(ctx context.Context, cfg *UpstreamMonitorConfig, opts upstreamMonitorRefreshOptions) (*UpstreamMonitorRefreshResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var working *UpstreamMonitorConfig
	if cfg == nil {
		current, err := s.getUpstreamMonitorConfigRaw(ctx)
		if err != nil {
			if !errors.Is(err, ErrSettingNotFound) {
				return nil, err
			}
			current = defaultUpstreamMonitorConfig()
		}
		working = current
	} else {
		working = cloneUpstreamMonitorConfig(cfg)
		s.mergeExistingUpstreamMonitorSecrets(ctx, working)
	}

	normalizeUpstreamMonitorConfig(working)
	if err := validateUpstreamMonitorConfig(working); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	summary := UpstreamMonitorRefreshSummary{}
	sourceFilter := strings.TrimSpace(opts.sourceID)
	sourceMatched := sourceFilter == ""

	for i := range working.Sources {
		source := &working.Sources[i]

		if sourceFilter != "" && source.ID != sourceFilter {
			continue
		}
		sourceMatched = true

		if sourceFilter == "" && (!source.Enabled || !source.AutoSyncEnabled) {
			summary.SkippedCount++
			continue
		}
		if !opts.force {
			if !working.AutoRefreshEnabled || !working.Enabled {
				summary.SkippedCount++
				continue
			}
			if !shouldRefreshUpstreamSource(source, working.RefreshIntervalMinutes, now) {
				summary.SkippedCount++
				continue
			}
		}

		summary.AttemptedCount++
		pricing, err := fetchUpstreamPricingSnapshot(ctx, source)
		ts := now
		source.LastSyncAt = &ts
		if err != nil {
			source.LastSyncStatus = upstreamMonitorSyncStatusError
			source.LastSyncError = limitUpstreamMonitorError(err.Error())
			summary.FailedCount++
			continue
		}
		if pricing.HasReference {
			source.ReferenceMultiplier = pricing.ReferenceMultiplier
		}
		source.UpstreamGroupOptions = normalizeUpstreamGroupOptions(pricing.GroupOptions)
		applyUpstreamGroupMultipliers(working, source.ID, pricing.GroupMultipliers, source.UpstreamGroupOptions)
		ensureUpstreamGroupOptionsFromMappings(working, source)
		source.LastSyncStatus = upstreamMonitorSyncStatusSuccess
		source.LastSyncError = ""
		summary.SuccessCount++
	}

	if !sourceMatched {
		return nil, fmt.Errorf("upstream monitor source %q not found", sourceFilter)
	}

	if opts.persist {
		if err := s.persistUpstreamMonitorConfigRaw(ctx, working); err != nil {
			return nil, err
		}
	}

	if opts.persist && working.Enabled {
		s.notifyUpstreamMonitorAlerts(ctx, working)
	}

	result := &UpstreamMonitorRefreshResult{
		RefreshedAt: now,
		Summary:     summary,
		Config:      working,
	}
	if opts.includeSnapshot {
		snapshot, err := s.PreviewUpstreamMonitorConfig(ctx, working)
		if err != nil {
			return nil, err
		}
		result.Snapshot = snapshot
	}

	return result, nil
}

func (s *SettingService) PreviewUpstreamMonitorConfig(ctx context.Context, cfg *UpstreamMonitorConfig) (*UpstreamMonitorPreviewSnapshot, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg == nil {
		current, err := s.GetUpstreamMonitorConfig(ctx)
		if err != nil {
			return nil, err
		}
		cfg = current
	}

	previewCfg := *cfg
	previewCfg.Sources = append([]UpstreamMonitorSource(nil), cfg.Sources...)
	previewCfg.GroupMappings = append([]UpstreamMonitorGroupMap(nil), cfg.GroupMappings...)
	normalizeUpstreamMonitorConfig(&previewCfg)
	if err := validateUpstreamMonitorConfig(&previewCfg); err != nil {
		return nil, err
	}

	if s.upstreamMonitorGroupLister == nil {
		return nil, fmt.Errorf("upstream monitor group lister not configured")
	}
	groups, err := s.upstreamMonitorGroupLister.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	accounts := []Account{}
	if s.upstreamMonitorAccountLister != nil {
		accounts, err = s.upstreamMonitorAccountLister.ListActive(ctx)
		if err != nil {
			return nil, fmt.Errorf("list active accounts: %w", err)
		}
	}

	groupByName := make(map[string]Group, len(groups))
	groupByID := make(map[int64]Group, len(groups))
	groupOptions := make([]UpstreamMonitorPreviewGroupOption, 0, len(groups))
	for _, group := range groups {
		groupByName[strings.ToLower(strings.TrimSpace(group.Name))] = group
		groupByID[group.ID] = group
		groupOptions = append(groupOptions, UpstreamMonitorPreviewGroupOption{
			GroupID:          group.ID,
			GroupName:        group.Name,
			Platform:         group.Platform,
			Multiplier:       group.RateMultiplier,
			IsExclusive:      group.IsExclusive,
			SubscriptionType: group.SubscriptionType,
		})
	}
	sort.SliceStable(groupOptions, func(i, j int) bool {
		leftPlatform := strings.ToLower(groupOptions[i].Platform)
		rightPlatform := strings.ToLower(groupOptions[j].Platform)
		if leftPlatform != rightPlatform {
			return leftPlatform < rightPlatform
		}
		leftName := strings.ToLower(groupOptions[i].GroupName)
		rightName := strings.ToLower(groupOptions[j].GroupName)
		if leftName != rightName {
			return leftName < rightName
		}
		return groupOptions[i].GroupID < groupOptions[j].GroupID
	})
	sourceByID := make(map[string]UpstreamMonitorSource, len(previewCfg.Sources))
	for _, source := range previewCfg.Sources {
		sourceByID[source.ID] = source
	}
	accountByID := make(map[int64]Account, len(accounts))
	accountOptions := make([]UpstreamMonitorPreviewAccountInfo, 0, len(accounts))
	for _, account := range accounts {
		accountByID[account.ID] = account
		info := upstreamMonitorAccountInfo(account, groupByID)
		accountOptions = append(accountOptions, info)
	}
	sort.SliceStable(accountOptions, func(i, j int) bool {
		left := strings.ToLower(accountOptions[i].AccountName)
		right := strings.ToLower(accountOptions[j].AccountName)
		if left != right {
			return left < right
		}
		return accountOptions[i].AccountID < accountOptions[j].AccountID
	})

	sourceRows := make([]UpstreamMonitorPreviewSourceRow, 0, len(previewCfg.Sources))
	enabledSources := 0
	for _, source := range previewCfg.Sources {
		if source.Enabled {
			enabledSources++
		}
		sourceRows = append(sourceRows, UpstreamMonitorPreviewSourceRow{
			ID:                   source.ID,
			Name:                 source.Name,
			Enabled:              source.Enabled,
			AutoSyncEnabled:      source.AutoSyncEnabled,
			Kind:                 source.Kind,
			FetchMode:            source.FetchMode,
			Currency:             source.Currency,
			ExchangeRate:         source.ExchangeRate,
			ReferenceMultiplier:  source.ReferenceMultiplier,
			AccountIDs:           append([]int64(nil), source.AccountIDs...),
			AccountCount:         len(source.AccountIDs),
			UpstreamGroupOptions: append([]UpstreamMonitorUpstreamGroupOption(nil), source.UpstreamGroupOptions...),
			AuthConfigured:       source.AuthConfigured,
			PricingURL:           source.PricingURL,
			LastSyncAt:           source.LastSyncAt,
			LastSyncStatus:       source.LastSyncStatus,
			LastSyncError:        source.LastSyncError,
			Notes:                source.Notes,
		})
	}
	sort.SliceStable(sourceRows, func(i, j int) bool {
		return strings.ToLower(sourceRows[i].Name) < strings.ToLower(sourceRows[j].Name)
	})

	groupRows := make([]UpstreamMonitorPreviewGroupRow, 0, len(previewCfg.GroupMappings))
	mappedGroupNames := make(map[string]struct{}, len(previewCfg.GroupMappings))
	var totalMargin float64
	var marginCount int
	var lowestMargin float64
	var highestMargin float64
	hasMargin := false
	healthyCount := 0
	warningCount := 0
	criticalCount := 0
	unknownCount := 0

	for _, mapping := range previewCfg.GroupMappings {
		row := UpstreamMonitorPreviewGroupRow{
			MappingID:         mapping.ID,
			LocalGroupID:      mapping.LocalGroupID,
			LocalGroup:        mapping.LocalGroup,
			UpstreamGroupKey:  mapping.UpstreamGroupKey,
			UpstreamGroup:     mapping.UpstreamGroup,
			ModelFamily:       mapping.ModelFamily,
			SourceIDs:         append([]string(nil), mapping.SourceIDs...),
			MappingMultiplier: mapping.ReferenceMultiplier,
			Notes:             mapping.Notes,
		}
		issues := make([]string, 0, 4)

		group, ok := groupByID[mapping.LocalGroupID]
		if !ok {
			group, ok = groupByName[strings.ToLower(strings.TrimSpace(mapping.LocalGroup))]
		}
		if ok {
			row.LocalGroupID = group.ID
			row.LocalGroupPlatform = group.Platform
			row.LocalGroupMultiplier = group.RateMultiplier
			row.SourceNames = row.SourceNames[:0]
			mappedGroupNames[strings.ToLower(strings.TrimSpace(group.Name))] = struct{}{}
		} else {
			issues = append(issues, "local group not found")
		}

		enabledSourceCount := 0
		bestReference := 0.0
		referenceFound := false
		for _, sourceID := range mapping.SourceIDs {
			source, ok := sourceByID[sourceID]
			if !ok {
				issues = append(issues, "source missing")
				continue
			}
			row.SourceNames = append(row.SourceNames, source.Name)
			if source.Enabled {
				enabledSourceCount++
				rawReference := mapping.ReferenceMultiplier
				if rawReference <= 0 {
					rawReference = source.ReferenceMultiplier
				}
				effectiveReference := upstreamMonitorEffectiveCostMultiplier(source, rawReference)
				if effectiveReference > 0 {
					if !referenceFound || effectiveReference > bestReference {
						bestReference = effectiveReference
					}
					referenceFound = true
				}
			}
		}
		row.EnabledSourceCount = enabledSourceCount
		row.SourceCount = len(row.SourceIDs)
		row.ReferenceMultiplier = bestReference
		if enabledSourceCount == 0 {
			issues = append(issues, "no enabled source")
		}

		if row.LocalGroupMultiplier <= 0 {
			issues = append(issues, "local multiplier missing")
		}
		if !referenceFound {
			issues = append(issues, "reference multiplier missing")
		}

		if row.LocalGroupMultiplier > 0 && referenceFound && bestReference > 0 {
			row.EstimatedMarginRate = upstreamMonitorProfitRate(row.LocalGroupMultiplier, bestReference)
			if !hasMargin {
				lowestMargin = row.EstimatedMarginRate
				highestMargin = row.EstimatedMarginRate
				hasMargin = true
			} else {
				if row.EstimatedMarginRate < lowestMargin {
					lowestMargin = row.EstimatedMarginRate
				}
				if row.EstimatedMarginRate > highestMargin {
					highestMargin = row.EstimatedMarginRate
				}
			}
			totalMargin += row.EstimatedMarginRate
			marginCount++
			row.Status = upstreamMonitorMarginStatus(row.EstimatedMarginRate, previewCfg.WarningRateThreshold, previewCfg.CriticalRateThreshold)
			healthyCount, warningCount, criticalCount, unknownCount = incrementUpstreamMonitorStatusCounts(row.Status, healthyCount, warningCount, criticalCount, unknownCount)
		} else {
			row.Status = "unknown"
			unknownCount++
		}

		sort.Strings(issues)
		row.Issues = dedupeUpstreamMonitorStrings(issues)
		groupRows = append(groupRows, row)
	}

	sort.SliceStable(groupRows, func(i, j int) bool {
		left := previewStatusWeight(groupRows[i].Status)
		right := previewStatusWeight(groupRows[j].Status)
		if left != right {
			return left < right
		}
		return strings.ToLower(groupRows[i].LocalGroup) < strings.ToLower(groupRows[j].LocalGroup)
	})

	accountRows := make([]UpstreamMonitorPreviewAccountRow, 0)
	monitoredAccountIDs := make(map[int64]struct{})
	for _, source := range previewCfg.Sources {
		effectiveSourceReference := upstreamMonitorEffectiveCostMultiplier(source, source.ReferenceMultiplier)
		for _, accountID := range source.AccountIDs {
			row := UpstreamMonitorPreviewAccountRow{
				SourceID:            source.ID,
				SourceName:          source.Name,
				AccountID:           accountID,
				ReferenceMultiplier: effectiveSourceReference,
			}
			issues := make([]string, 0, 4)
			if !source.Enabled {
				issues = append(issues, "source disabled")
			}
			if source.ReferenceMultiplier <= 0 {
				issues = append(issues, "reference multiplier missing")
			}

			account, ok := accountByID[accountID]
			if !ok {
				issues = append(issues, "account not found")
			} else {
				monitoredAccountIDs[accountID] = struct{}{}
				row.AccountName = account.Name
				row.AccountPlatform = account.Platform
				row.AccountType = account.Type
				row.AccountRateMultiplier = account.BillingRateMultiplier()
				row.GroupIDs = append([]int64(nil), account.GroupIDs...)
				for _, groupID := range account.GroupIDs {
					if group, ok := groupByID[groupID]; ok {
						row.GroupNames = append(row.GroupNames, group.Name)
						row.GroupMultipliers = append(row.GroupMultipliers, group.RateMultiplier)
						if group.RateMultiplier > row.HighestGroupMultiplier {
							row.HighestGroupMultiplier = group.RateMultiplier
						}
					}
				}
				if row.HighestGroupMultiplier <= 0 {
					issues = append(issues, "account has no active group")
				}
			}

			row.EstimatedCostMultiplier = row.AccountRateMultiplier
			if row.ReferenceMultiplier > row.EstimatedCostMultiplier {
				row.EstimatedCostMultiplier = row.ReferenceMultiplier
			}
			if row.HighestGroupMultiplier > 0 && row.EstimatedCostMultiplier > 0 && len(issues) == 0 {
				row.EstimatedMarginRate = upstreamMonitorProfitRate(row.HighestGroupMultiplier, row.EstimatedCostMultiplier)
				if !hasMargin {
					lowestMargin = row.EstimatedMarginRate
					highestMargin = row.EstimatedMarginRate
					hasMargin = true
				} else {
					if row.EstimatedMarginRate < lowestMargin {
						lowestMargin = row.EstimatedMarginRate
					}
					if row.EstimatedMarginRate > highestMargin {
						highestMargin = row.EstimatedMarginRate
					}
				}
				totalMargin += row.EstimatedMarginRate
				marginCount++
				row.Status = upstreamMonitorMarginStatus(row.EstimatedMarginRate, previewCfg.WarningRateThreshold, previewCfg.CriticalRateThreshold)
				healthyCount, warningCount, criticalCount, unknownCount = incrementUpstreamMonitorStatusCounts(row.Status, healthyCount, warningCount, criticalCount, unknownCount)
			} else {
				row.Status = "unknown"
				unknownCount++
			}
			sort.Strings(issues)
			row.Issues = dedupeUpstreamMonitorStrings(issues)
			accountRows = append(accountRows, row)
		}
	}
	sort.SliceStable(accountRows, func(i, j int) bool {
		left := previewStatusWeight(accountRows[i].Status)
		right := previewStatusWeight(accountRows[j].Status)
		if left != right {
			return left < right
		}
		if accountRows[i].SourceName != accountRows[j].SourceName {
			return strings.ToLower(accountRows[i].SourceName) < strings.ToLower(accountRows[j].SourceName)
		}
		if accountRows[i].AccountName != accountRows[j].AccountName {
			return strings.ToLower(accountRows[i].AccountName) < strings.ToLower(accountRows[j].AccountName)
		}
		return accountRows[i].AccountID < accountRows[j].AccountID
	})

	unmappedGroups := make([]UpstreamMonitorPreviewUnmappedRow, 0)
	for _, group := range groups {
		if _, ok := mappedGroupNames[strings.ToLower(strings.TrimSpace(group.Name))]; ok {
			continue
		}
		unmappedGroups = append(unmappedGroups, UpstreamMonitorPreviewUnmappedRow{
			GroupID:          group.ID,
			GroupName:        group.Name,
			Platform:         group.Platform,
			Multiplier:       group.RateMultiplier,
			IsExclusive:      group.IsExclusive,
			SubscriptionType: group.SubscriptionType,
		})
	}
	sort.SliceStable(unmappedGroups, func(i, j int) bool {
		return strings.ToLower(unmappedGroups[i].GroupName) < strings.ToLower(unmappedGroups[j].GroupName)
	})

	summary := UpstreamMonitorPreviewSummary{
		Enabled:               previewCfg.Enabled,
		AutoRefreshEnabled:    previewCfg.AutoRefreshEnabled,
		SourceCount:           len(previewCfg.Sources),
		EnabledSourceCount:    enabledSources,
		MappedGroupCount:      len(groupRows),
		MonitoredAccountCount: len(monitoredAccountIDs),
		UnmappedGroupCount:    len(unmappedGroups),
		HealthyCount:          healthyCount,
		WarningCount:          warningCount,
		CriticalCount:         criticalCount,
		UnknownCount:          unknownCount,
	}
	if marginCount > 0 {
		summary.AverageMarginRate = totalMargin / float64(marginCount)
		summary.LowestMarginRate = lowestMargin
		summary.HighestMarginRate = highestMargin
	}

	return &UpstreamMonitorPreviewSnapshot{
		GeneratedAt:    time.Now().UTC(),
		Summary:        summary,
		SourceRows:     sourceRows,
		GroupRows:      groupRows,
		AccountRows:    accountRows,
		AccountOptions: accountOptions,
		GroupOptions:   groupOptions,
		UnmappedGroups: unmappedGroups,
	}, nil
}

func previewStatusWeight(status string) int {
	switch status {
	case "critical":
		return 0
	case "warning":
		return 1
	case "unknown":
		return 2
	case "healthy":
		return 3
	default:
		return 4
	}
}

func upstreamMonitorMarginStatus(marginRate, warningThreshold, criticalThreshold float64) string {
	switch {
	case marginRate <= criticalThreshold:
		return "critical"
	case marginRate <= warningThreshold:
		return "warning"
	default:
		return "healthy"
	}
}

func upstreamMonitorProfitRate(localMultiplier, upstreamMultiplier float64) float64 {
	if localMultiplier <= 0 {
		return 0
	}
	return (localMultiplier - upstreamMultiplier) / localMultiplier
}

func upstreamMonitorEffectiveCostMultiplier(source UpstreamMonitorSource, referenceMultiplier float64) float64 {
	if referenceMultiplier <= 0 {
		return 0
	}
	currency := strings.ToUpper(strings.TrimSpace(source.Currency))
	if currency == "" || currency == "CNY" {
		return referenceMultiplier
	}
	exchangeRate := source.ExchangeRate
	if exchangeRate <= 0 {
		exchangeRate = 1
	}
	return referenceMultiplier * exchangeRate
}

func incrementUpstreamMonitorStatusCounts(status string, healthyCount, warningCount, criticalCount, unknownCount int) (int, int, int, int) {
	switch status {
	case "healthy":
		healthyCount++
	case "warning":
		warningCount++
	case "critical":
		criticalCount++
	default:
		unknownCount++
	}
	return healthyCount, warningCount, criticalCount, unknownCount
}

func upstreamMonitorAccountInfo(account Account, groupByID map[int64]Group) UpstreamMonitorPreviewAccountInfo {
	info := UpstreamMonitorPreviewAccountInfo{
		AccountID:      account.ID,
		AccountName:    account.Name,
		Platform:       account.Platform,
		Type:           account.Type,
		RateMultiplier: account.BillingRateMultiplier(),
		Status:         account.Status,
		GroupIDs:       append([]int64(nil), account.GroupIDs...),
		GroupNames:     make([]string, 0, len(account.GroupIDs)),
	}
	for _, groupID := range account.GroupIDs {
		if group, ok := groupByID[groupID]; ok {
			info.GroupNames = append(info.GroupNames, group.Name)
		}
	}
	return info
}

func dedupeUpstreamMonitorStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func cleanUpstreamMonitorInt64IDs(values []int64) []int64 {
	if len(values) == 0 {
		return []int64{}
	}
	seen := make(map[int64]struct{}, len(values))
	out := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func normalizeUpstreamGroupOptions(values []UpstreamMonitorUpstreamGroupOption) []UpstreamMonitorUpstreamGroupOption {
	if len(values) == 0 {
		return []UpstreamMonitorUpstreamGroupOption{}
	}
	out := make([]UpstreamMonitorUpstreamGroupOption, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		option := UpstreamMonitorUpstreamGroupOption{
			Key:                 strings.TrimSpace(value.Key),
			Name:                strings.TrimSpace(value.Name),
			Description:         strings.TrimSpace(value.Description),
			ReferenceMultiplier: value.ReferenceMultiplier,
			RawID:               strings.TrimSpace(value.RawID),
			Path:                strings.TrimSpace(value.Path),
		}
		if option.Name == "" || option.ReferenceMultiplier < 0 || option.ReferenceMultiplier > 1000 {
			continue
		}
		if option.Key == "" {
			option.Key = upstreamGroupOptionKey(option.RawID, option.Name, option.Path)
		}
		if option.Key == "" {
			continue
		}
		if _, ok := seen[option.Key]; ok {
			continue
		}
		seen[option.Key] = struct{}{}
		out = append(out, option)
	}
	sort.SliceStable(out, func(i, j int) bool {
		left := strings.ToLower(out[i].Name)
		right := strings.ToLower(out[j].Name)
		if left != right {
			return left < right
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func upstreamGroupOptionKey(rawID, name, path string) string {
	rawID = strings.TrimSpace(rawID)
	if rawID != "" {
		return "id:" + rawID
	}
	path = strings.TrimSpace(path)
	if path != "" {
		return "path:" + path
	}
	name = strings.ToLower(strings.TrimSpace(name))
	if name != "" {
		return "name:" + name
	}
	return ""
}

func cloneUpstreamMonitorConfig(cfg *UpstreamMonitorConfig) *UpstreamMonitorConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	if cfg.Sources != nil {
		cloned.Sources = make([]UpstreamMonitorSource, len(cfg.Sources))
		copy(cloned.Sources, cfg.Sources)
		for i := range cfg.Sources {
			if cfg.Sources[i].AccountIDs != nil {
				cloned.Sources[i].AccountIDs = append([]int64(nil), cfg.Sources[i].AccountIDs...)
			}
			if cfg.Sources[i].UpstreamGroupOptions != nil {
				cloned.Sources[i].UpstreamGroupOptions = append([]UpstreamMonitorUpstreamGroupOption(nil), cfg.Sources[i].UpstreamGroupOptions...)
			}
		}
	}
	if cfg.GroupMappings != nil {
		cloned.GroupMappings = make([]UpstreamMonitorGroupMap, len(cfg.GroupMappings))
		copy(cloned.GroupMappings, cfg.GroupMappings)
		for i := range cfg.GroupMappings {
			if cfg.GroupMappings[i].SourceIDs != nil {
				cloned.GroupMappings[i].SourceIDs = append([]string(nil), cfg.GroupMappings[i].SourceIDs...)
			}
		}
	}
	return &cloned
}

func shouldRefreshUpstreamSource(source *UpstreamMonitorSource, intervalMinutes int, now time.Time) bool {
	if source == nil {
		return false
	}
	if intervalMinutes <= 0 {
		intervalMinutes = 10
	}
	if source.LastSyncAt == nil || source.LastSyncAt.IsZero() {
		return true
	}
	return now.Sub(source.LastSyncAt.UTC()) >= time.Duration(intervalMinutes)*time.Minute
}

func fetchUpstreamPricingSnapshot(ctx context.Context, source *UpstreamMonitorSource) (*upstreamMonitorPricingSnapshot, error) {
	if source == nil {
		return nil, fmt.Errorf("source is nil")
	}
	if strings.TrimSpace(source.PricingURL) == "" {
		return nil, fmt.Errorf("pricing url is empty")
	}

	request := newUpstreamMonitorHTTPClient().
		R().
		SetContext(ctx).
		SetHeader("Accept", "application/json, text/plain;q=0.9, */*;q=0.8")

	switch source.AuthMode {
	case "bearer":
		if token := strings.TrimSpace(source.AuthToken); token != "" {
			request.SetBearerAuthToken(token)
		}
	case "header":
		if token := strings.TrimSpace(source.AuthToken); token != "" {
			headerName := strings.TrimSpace(source.AuthHeaderName)
			if headerName == "" {
				headerName = http.CanonicalHeaderKey("Authorization")
			}
			request.SetHeader(headerName, token)
		}
	case "cookie":
		if token := strings.TrimSpace(source.AuthToken); token != "" {
			request.SetHeader("Cookie", token)
		}
	}

	resp, err := request.Get(source.PricingURL)
	if err != nil {
		return nil, fmt.Errorf("request pricing endpoint: %w", err)
	}
	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("pricing endpoint returned status %d", resp.StatusCode)
	}
	pricing, err := parseUpstreamPricingSnapshot(
		resp.Bytes(),
		resp.Header.Get("Content-Type"),
		source.FetchMode,
		source.PricingPathHint,
	)
	if err != nil {
		return nil, err
	}
	if pricing.HasReference && (pricing.ReferenceMultiplier < 0 || pricing.ReferenceMultiplier > 1000) {
		return nil, fmt.Errorf("parsed reference multiplier %.4f out of range", pricing.ReferenceMultiplier)
	}
	for group, value := range pricing.GroupMultipliers {
		if value < 0 || value > 1000 {
			return nil, fmt.Errorf("parsed group %q multiplier %.4f out of range", group, value)
		}
	}
	for _, option := range pricing.GroupOptions {
		if option.ReferenceMultiplier < 0 || option.ReferenceMultiplier > 1000 {
			return nil, fmt.Errorf("parsed group %q multiplier %.4f out of range", option.Name, option.ReferenceMultiplier)
		}
	}
	return pricing, nil
}

func parseUpstreamPricingSnapshot(body []byte, contentType, fetchMode, pathHint string) (*upstreamMonitorPricingSnapshot, error) {
	snapshot := &upstreamMonitorPricingSnapshot{
		GroupMultipliers: map[string]float64{},
	}
	mode := normalizeUpstreamFetchMode(fetchMode)
	switch mode {
	case upstreamMonitorFetchModePlainText:
		value, err := parseUpstreamPlainTextMultiplier(body)
		if err != nil {
			return nil, err
		}
		snapshot.ReferenceMultiplier = value
		snapshot.HasReference = true
		return snapshot, nil
	case upstreamMonitorFetchModeJSONPath:
		value, err := parseUpstreamJSONMultiplier(body, pathHint)
		if err != nil {
			return nil, err
		}
		snapshot.ReferenceMultiplier = value
		snapshot.HasReference = true
		snapshot.GroupMultipliers = collectUpstreamGroupMultipliers(body)
		snapshot.GroupOptions = collectUpstreamGroupOptions(body)
		return snapshot, nil
	default:
		if strings.TrimSpace(pathHint) != "" && gjson.ValidBytes(body) {
			if value, err := parseUpstreamJSONMultiplier(body, pathHint); err == nil {
				snapshot.ReferenceMultiplier = value
				snapshot.HasReference = true
			}
		}
		if gjson.ValidBytes(body) || strings.Contains(strings.ToLower(contentType), "json") {
			snapshot.GroupMultipliers = collectUpstreamGroupMultipliers(body)
			snapshot.GroupOptions = collectUpstreamGroupOptions(body)
			if value, ok := tryAutoJSONMultiplier(body); ok {
				snapshot.ReferenceMultiplier = value
				snapshot.HasReference = true
			}
			if snapshot.HasReference || len(snapshot.GroupMultipliers) > 0 || len(snapshot.GroupOptions) > 0 {
				return snapshot, nil
			}
		}
		value, err := parseUpstreamPlainTextMultiplier(body)
		if err != nil {
			return nil, err
		}
		snapshot.ReferenceMultiplier = value
		snapshot.HasReference = true
		return snapshot, nil
	}
}

func parseUpstreamJSONMultiplier(body []byte, path string) (float64, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return 0, fmt.Errorf("pricing path hint is empty")
	}
	if !gjson.ValidBytes(body) {
		return 0, fmt.Errorf("pricing response is not valid json")
	}
	result := gjson.GetBytes(body, path)
	if !result.Exists() {
		return 0, fmt.Errorf("json path %q not found", path)
	}
	value, ok := gjsonResultToFloat(result)
	if !ok {
		return 0, fmt.Errorf("json path %q did not contain a numeric value", path)
	}
	return value, nil
}

func tryAutoJSONMultiplier(body []byte) (float64, bool) {
	paths := []string{
		"reference_multiplier",
		"multiplier",
		"cost_multiplier",
		"rate",
		"data.reference_multiplier",
		"data.multiplier",
		"data.cost_multiplier",
		"data.rate",
		"result.reference_multiplier",
		"result.multiplier",
		"result.cost_multiplier",
		"result.rate",
		"payload.reference_multiplier",
		"payload.multiplier",
		"payload.cost_multiplier",
		"payload.rate",
	}
	for _, path := range paths {
		result := gjson.GetBytes(body, path)
		if !result.Exists() {
			continue
		}
		value, ok := gjsonResultToFloat(result)
		if ok {
			return value, true
		}
	}
	return 0, false
}

func collectUpstreamGroupMultipliers(body []byte) map[string]float64 {
	out := map[string]float64{}
	if !gjson.ValidBytes(body) {
		return out
	}
	collectUpstreamGroupRatio(gjson.ParseBytes(body), out)
	collectUpstreamGroupMultipliersFromResult(gjson.ParseBytes(body), out, false)
	return out
}

func collectUpstreamGroupOptions(body []byte) []UpstreamMonitorUpstreamGroupOption {
	if !gjson.ValidBytes(body) {
		return []UpstreamMonitorUpstreamGroupOption{}
	}
	options := make([]UpstreamMonitorUpstreamGroupOption, 0)
	root := gjson.ParseBytes(body)
	collectUpstreamGroupRatioOptions(root, &options)
	collectUpstreamGroupOptionsFromResult(root, "$", &options, false)
	return normalizeUpstreamGroupOptions(options)
}

func collectUpstreamGroupRatio(result gjson.Result, out map[string]float64) {
	if out == nil || !result.Exists() {
		return
	}
	for _, path := range []string{"group_ratio", "groupRatio", "data.group_ratio", "data.groupRatio"} {
		ratios := result.Get(path)
		if !ratios.IsObject() {
			continue
		}
		ratios.ForEach(func(key, value gjson.Result) bool {
			groupName := strings.TrimSpace(key.String())
			if multiplier, ok := gjsonResultToFloat(value); ok && groupName != "" {
				out[strings.ToLower(groupName)] = multiplier
			}
			return true
		})
	}
}

func collectUpstreamGroupRatioOptions(result gjson.Result, out *[]UpstreamMonitorUpstreamGroupOption) {
	if out == nil || !result.Exists() {
		return
	}
	descriptions := collectUpstreamGroupDescriptions(result)
	for _, path := range []string{"group_ratio", "groupRatio", "data.group_ratio", "data.groupRatio"} {
		ratios := result.Get(path)
		if !ratios.IsObject() {
			continue
		}
		ratios.ForEach(func(key, value gjson.Result) bool {
			groupName := strings.TrimSpace(key.String())
			multiplier, ok := gjsonResultToFloat(value)
			if ok && groupName != "" {
				*out = append(*out, UpstreamMonitorUpstreamGroupOption{
					Key:                 "name:" + strings.ToLower(groupName),
					Name:                groupName,
					Description:         descriptions[strings.ToLower(groupName)],
					ReferenceMultiplier: multiplier,
					Path:                path + "." + groupName,
				})
			}
			return true
		})
	}
}

func collectUpstreamGroupDescriptions(result gjson.Result) map[string]string {
	out := map[string]string{}
	if !result.Exists() {
		return out
	}
	for _, path := range []string{"usable_group", "usableGroup", "data.usable_group", "data.usableGroup"} {
		groups := result.Get(path)
		if !groups.IsObject() {
			continue
		}
		groups.ForEach(func(key, value gjson.Result) bool {
			groupName := strings.TrimSpace(key.String())
			description := strings.TrimSpace(value.String())
			if groupName != "" && description != "" {
				out[strings.ToLower(groupName)] = description
			}
			return true
		})
	}
	for _, path := range []string{"usable_group_meta", "usableGroupMeta", "data.usable_group_meta", "data.usableGroupMeta"} {
		groups := result.Get(path)
		if !groups.IsObject() {
			continue
		}
		groups.ForEach(func(key, value gjson.Result) bool {
			groupName := strings.TrimSpace(key.String())
			if groupName == "" || !value.IsObject() {
				return true
			}
			for _, field := range []string{"desc", "description", "name", "title"} {
				description := strings.TrimSpace(value.Get(field).String())
				if description != "" {
					out[strings.ToLower(groupName)] = description
					break
				}
			}
			return true
		})
	}
	return out
}

func collectUpstreamGroupMultipliersFromResult(result gjson.Result, out map[string]float64, inGroupContainer bool) {
	if out == nil || !result.Exists() {
		return
	}
	if result.IsArray() {
		for _, item := range result.Array() {
			collectUpstreamGroupMultiplierFromObject(item, out)
			if item.IsArray() || item.IsObject() {
				collectUpstreamGroupMultipliersFromResult(item, out, inGroupContainer)
			}
		}
		return
	}
	if result.IsObject() {
		collectUpstreamGroupRatioLikeObject(result, "", out, nil)
		if !collectUpstreamGroupMultiplierFromObject(result, out) {
			result.ForEach(func(key, value gjson.Result) bool {
				groupName := strings.TrimSpace(key.String())
				if value.IsArray() || value.IsObject() {
					if isUpstreamMonitorIgnoredPricingContainerKey(groupName) {
						return true
					}
					if isUpstreamMonitorGroupRatioKey(groupName) {
						return true
					}
					childIsGroupContainer := inGroupContainer || isUpstreamMonitorGroupContainerKey(groupName)
					if multiplier, ok := firstUpstreamGroupMultiplier(value); ok && childIsGroupContainer && groupName != "" {
						out[strings.ToLower(groupName)] = multiplier
					}
					collectUpstreamGroupMultipliersFromResult(value, out, childIsGroupContainer)
					return true
				}
				if multiplier, ok := gjsonResultToFloat(value); ok && inGroupContainer && groupName != "" && !isUpstreamMonitorGenericMultiplierKey(groupName) {
					out[strings.ToLower(groupName)] = multiplier
				}
				return true
			})
		}
	}
}

func collectUpstreamGroupOptionsFromResult(result gjson.Result, path string, out *[]UpstreamMonitorUpstreamGroupOption, inGroupContainer bool) {
	if out == nil || !result.Exists() {
		return
	}
	if result.IsArray() {
		for index, item := range result.Array() {
			itemPath := fmt.Sprintf("%s[%d]", path, index)
			if item.IsObject() {
				collectUpstreamGroupOptionFromObject(item, itemPath, out)
			}
			if item.IsArray() || item.IsObject() {
				collectUpstreamGroupOptionsFromResult(item, itemPath, out, inGroupContainer)
			}
		}
		return
	}
	if result.IsObject() {
		collectUpstreamGroupRatioLikeObject(result, path, nil, out)
		if !collectUpstreamGroupOptionFromObject(result, path, out) {
			result.ForEach(func(key, value gjson.Result) bool {
				groupName := strings.TrimSpace(key.String())
				childPath := joinUpstreamGroupOptionPath(path, groupName)
				if value.IsArray() || value.IsObject() {
					if isUpstreamMonitorIgnoredPricingContainerKey(groupName) {
						return true
					}
					if isUpstreamMonitorGroupRatioKey(groupName) {
						return true
					}
					if value.IsObject() && collectUpstreamGroupOptionFromObject(value, childPath, out) {
						return true
					}
					childIsGroupContainer := inGroupContainer || isUpstreamMonitorGroupContainerKey(groupName)
					if multiplier, ok := firstUpstreamGroupMultiplier(value); ok && childIsGroupContainer && groupName != "" && !isUpstreamMonitorGenericContainerKey(groupName) {
						*out = append(*out, UpstreamMonitorUpstreamGroupOption{
							Key:                 upstreamGroupOptionKey("", groupName, childPath),
							Name:                groupName,
							ReferenceMultiplier: multiplier,
							Path:                childPath,
						})
					}
					collectUpstreamGroupOptionsFromResult(value, childPath, out, childIsGroupContainer)
					return true
				}
				if multiplier, ok := gjsonResultToFloat(value); ok && inGroupContainer && groupName != "" && !isUpstreamMonitorGenericMultiplierKey(groupName) {
					*out = append(*out, UpstreamMonitorUpstreamGroupOption{
						Key:                 upstreamGroupOptionKey("", groupName, childPath),
						Name:                groupName,
						ReferenceMultiplier: multiplier,
						Path:                childPath,
					})
				}
				return true
			})
		}
	}
}

func collectUpstreamGroupRatioLikeObject(result gjson.Result, path string, multipliers map[string]float64, options *[]UpstreamMonitorUpstreamGroupOption) {
	if !result.IsObject() {
		return
	}
	type ratioItem struct {
		name       string
		multiplier float64
		path       string
	}
	items := make([]ratioItem, 0)
	rejectedNumericKeys := 0
	result.ForEach(func(key, value gjson.Result) bool {
		groupName := strings.TrimSpace(key.String())
		if value.IsArray() || value.IsObject() {
			return true
		}
		multiplier, ok := gjsonResultToFloat(value)
		if !ok {
			return true
		}
		if !isLikelyUpstreamGroupName(groupName) {
			rejectedNumericKeys++
			return true
		}
		items = append(items, ratioItem{
			name:       groupName,
			multiplier: multiplier,
			path:       joinUpstreamGroupOptionPath(path, groupName),
		})
		return true
	})
	if len(items) == 0 || rejectedNumericKeys > 0 {
		return
	}
	for _, item := range items {
		if multipliers != nil {
			multipliers[strings.ToLower(item.name)] = item.multiplier
		}
		if options != nil {
			*options = append(*options, UpstreamMonitorUpstreamGroupOption{
				Key:                 upstreamGroupOptionKey("", item.name, item.path),
				Name:                item.name,
				ReferenceMultiplier: item.multiplier,
				Path:                item.path,
			})
		}
	}
}

func collectUpstreamGroupOptionFromObject(result gjson.Result, path string, out *[]UpstreamMonitorUpstreamGroupOption) bool {
	if out == nil || !result.IsObject() {
		return false
	}
	groupName := firstUpstreamGroupName(result)
	if groupName == "" {
		return false
	}
	multiplier, ok := firstUpstreamGroupMultiplier(result)
	if !ok {
		return false
	}
	rawID := firstUpstreamGroupRawID(result)
	*out = append(*out, UpstreamMonitorUpstreamGroupOption{
		Key:                 upstreamGroupOptionKey(rawID, groupName, path),
		Name:                groupName,
		ReferenceMultiplier: multiplier,
		RawID:               rawID,
		Path:                path,
	})
	return true
}

func collectUpstreamGroupMultiplierFromObject(result gjson.Result, out map[string]float64) bool {
	if out == nil || !result.IsObject() {
		return false
	}
	groupName := firstUpstreamGroupName(result)
	if groupName == "" {
		return false
	}
	multiplier, ok := firstUpstreamGroupMultiplier(result)
	if !ok {
		return false
	}
	out[strings.ToLower(groupName)] = multiplier
	return true
}

func firstUpstreamGroupName(result gjson.Result) string {
	for _, path := range []string{
		"group",
		"group_name",
		"groupName",
		"name",
		"title",
		"model_group",
		"modelGroup",
		"upstream_group",
		"upstreamGroup",
	} {
		value := strings.TrimSpace(result.Get(path).String())
		if value != "" {
			return value
		}
	}
	return ""
}

func firstUpstreamGroupRawID(result gjson.Result) string {
	for _, path := range []string{
		"id",
		"group_id",
		"groupId",
		"key",
		"slug",
		"code",
		"uid",
		"uuid",
	} {
		value := result.Get(path)
		if !value.Exists() {
			continue
		}
		raw := strings.TrimSpace(value.String())
		if raw != "" {
			return raw
		}
	}
	return ""
}

func firstUpstreamGroupMultiplier(result gjson.Result) (float64, bool) {
	for _, path := range []string{
		"reference_multiplier",
		"rate_multiplier",
		"rateMultiplier",
		"multiplier",
		"cost_multiplier",
		"rate",
		"price",
		"value",
	} {
		value, ok := gjsonResultToFloat(result.Get(path))
		if ok {
			return value, true
		}
	}
	return 0, false
}

func joinUpstreamGroupOptionPath(parent, child string) string {
	parent = strings.TrimSpace(parent)
	child = strings.TrimSpace(child)
	if parent == "" {
		return child
	}
	if child == "" {
		return parent
	}
	return parent + "." + child
}

func isUpstreamMonitorGenericMultiplierKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "reference_multiplier", "multiplier", "cost_multiplier", "rate", "price", "value",
		"rate_multiplier", "ratemultiplier",
		"input_price", "inputprice", "output_price", "outputprice", "cache_write_price", "cachewriteprice",
		"cache_read_price", "cachereadprice", "image_output_price", "imageoutputprice", "per_request_price", "perrequestprice",
		"model_ratio", "modelratio", "completion_ratio", "completionratio", "cache_ratio", "cacheratio",
		"create_cache_ratio", "createcacheratio", "model_price", "modelprice", "quota_type", "quotatype",
		"vendor_id", "vendorid":
		return true
	default:
		return false
	}
}

func isUpstreamMonitorGenericContainerKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "data", "result", "payload", "items", "list", "rows":
		return true
	default:
		return false
	}
}

func isUpstreamMonitorGroupContainerKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "groups", "group", "group_ratios", "groupratios", "group_multipliers", "groupmultipliers", "rates", "ratios":
		return true
	default:
		return false
	}
}

func isUpstreamMonitorGroupRatioKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "group_ratio", "groupratio":
		return true
	default:
		return false
	}
}

func isUpstreamMonitorIgnoredPricingContainerKey(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "model_ratio", "modelratio", "model_price", "modelprice", "model_prices", "modelprices",
		"completion_ratio", "completionratio", "cache_ratio", "cacheratio",
		"create_cache_ratio", "createcacheratio", "pricing", "prices",
		"model_pricing", "modelpricing", "supported_models", "supportedmodels":
		return true
	default:
		return false
	}
}

func isLikelyUpstreamGroupName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || isUpstreamMonitorGenericMultiplierKey(name) || isUpstreamMonitorGenericContainerKey(name) || isUpstreamMonitorGroupRatioKey(name) {
		return false
	}
	lower := strings.ToLower(name)
	if strings.Contains(lower, "倍率") || strings.Contains(lower, "ratio") || strings.Contains(lower, "group") || strings.Contains(lower, "vip") || strings.Contains(lower, "codex") || strings.Contains(lower, "relay") {
		return true
	}
	if strings.Contains(name, "-") || strings.Contains(name, "_") {
		return true
	}
	for _, r := range name {
		if r > 127 {
			return true
		}
	}
	return false
}

func ensureUpstreamGroupOptionsFromMappings(cfg *UpstreamMonitorConfig, source *UpstreamMonitorSource) {
	if cfg == nil || source == nil || strings.TrimSpace(source.ID) == "" {
		return
	}
	options := append([]UpstreamMonitorUpstreamGroupOption(nil), source.UpstreamGroupOptions...)
	existing := make(map[string]struct{}, len(options))
	for _, option := range normalizeUpstreamGroupOptions(options) {
		existing[option.Key] = struct{}{}
	}
	for _, mapping := range cfg.GroupMappings {
		if !stringSliceContains(mapping.SourceIDs, source.ID) {
			continue
		}
		name := strings.TrimSpace(mapping.UpstreamGroup)
		if name == "" {
			name = strings.TrimSpace(mapping.LocalGroup)
		}
		if name == "" {
			continue
		}
		multiplier := mapping.ReferenceMultiplier
		if multiplier <= 0 {
			multiplier = source.ReferenceMultiplier
		}
		if multiplier <= 0 {
			continue
		}
		key := strings.TrimSpace(mapping.UpstreamGroupKey)
		if key == "" {
			key = upstreamGroupOptionKey("", name, "configured:"+source.ID+":"+name)
		}
		if _, ok := existing[key]; ok {
			continue
		}
		existing[key] = struct{}{}
		options = append(options, UpstreamMonitorUpstreamGroupOption{
			Key:                 key,
			Name:                name,
			ReferenceMultiplier: multiplier,
			Path:                "configured:" + source.ID + ":" + name,
		})
	}
	source.UpstreamGroupOptions = normalizeUpstreamGroupOptions(options)
}

func applyUpstreamGroupMultipliers(cfg *UpstreamMonitorConfig, sourceID string, groupMultipliers map[string]float64, groupOptions []UpstreamMonitorUpstreamGroupOption) {
	if cfg == nil || strings.TrimSpace(sourceID) == "" || (len(groupMultipliers) == 0 && len(groupOptions) == 0) {
		return
	}
	normalizedOptions := normalizeUpstreamGroupOptions(groupOptions)
	optionByKey := make(map[string]UpstreamMonitorUpstreamGroupOption, len(normalizedOptions))
	optionsByName := make(map[string][]UpstreamMonitorUpstreamGroupOption, len(normalizedOptions))
	for _, option := range normalizedOptions {
		if option.Key != "" {
			optionByKey[option.Key] = option
		}
		nameKey := strings.ToLower(strings.TrimSpace(option.Name))
		if nameKey != "" {
			optionsByName[nameKey] = append(optionsByName[nameKey], option)
		}
	}
	for i := range cfg.GroupMappings {
		mapping := &cfg.GroupMappings[i]
		if !stringSliceContains(mapping.SourceIDs, sourceID) {
			continue
		}
		optionKey := strings.TrimSpace(mapping.UpstreamGroupKey)
		if optionKey != "" {
			if option, ok := optionByKey[optionKey]; ok {
				mapping.UpstreamGroupKey = option.Key
				mapping.UpstreamGroup = option.Name
				mapping.ReferenceMultiplier = option.ReferenceMultiplier
				continue
			}
		}
		groupKey := strings.ToLower(strings.TrimSpace(mapping.UpstreamGroup))
		if groupKey == "" {
			groupKey = strings.ToLower(strings.TrimSpace(mapping.LocalGroup))
		}
		if options := optionsByName[groupKey]; len(options) == 1 {
			mapping.UpstreamGroupKey = options[0].Key
			mapping.UpstreamGroup = options[0].Name
			mapping.ReferenceMultiplier = options[0].ReferenceMultiplier
			continue
		}
		if options := optionsByName[groupKey]; len(options) > 1 {
			continue
		}
		if value, ok := groupMultipliers[groupKey]; ok {
			mapping.ReferenceMultiplier = value
		}
	}
}

func stringSliceContains(values []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}

func gjsonResultToFloat(result gjson.Result) (float64, bool) {
	switch result.Type {
	case gjson.Number:
		return result.Float(), true
	case gjson.String:
		value, err := strconv.ParseFloat(strings.TrimSpace(result.String()), 64)
		if err != nil {
			return 0, false
		}
		return value, true
	default:
		return 0, false
	}
}

func parseUpstreamPlainTextMultiplier(body []byte) (float64, error) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return 0, fmt.Errorf("pricing response is empty")
	}
	if len(trimmed) > 64 {
		return 0, fmt.Errorf("plain text response is too long; use json_path mode instead")
	}
	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("plain text response is not a valid number")
	}
	return value, nil
}

func limitUpstreamMonitorError(message string) string {
	message = strings.TrimSpace(message)
	if len(message) <= 240 {
		return message
	}
	return message[:240]
}

func (s *SettingService) persistUpstreamMonitorConfigRaw(ctx context.Context, cfg *UpstreamMonitorConfig) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal upstream monitor config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyUpstreamMonitorConfig, string(data)); err != nil {
		return fmt.Errorf("save upstream monitor config: %w", err)
	}
	return nil
}

func (s *SettingService) mergeExistingUpstreamMonitorSecrets(ctx context.Context, cfg *UpstreamMonitorConfig) {
	existing, err := s.getUpstreamMonitorConfigRaw(ctx)
	if err != nil || existing == nil {
		return
	}
	existingByID := make(map[string]string, len(existing.Sources))
	for _, src := range existing.Sources {
		if token := strings.TrimSpace(src.AuthToken); token != "" {
			existingByID[src.ID] = token
		}
	}
	for i := range cfg.Sources {
		cfg.Sources[i].ID = strings.TrimSpace(cfg.Sources[i].ID)
		if strings.TrimSpace(cfg.Sources[i].AuthToken) == "" {
			if token, ok := existingByID[cfg.Sources[i].ID]; ok {
				cfg.Sources[i].AuthToken = token
			}
		}
	}
}

func (s *SettingService) getUpstreamMonitorConfigRaw(ctx context.Context) (*UpstreamMonitorConfig, error) {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyUpstreamMonitorConfig)
	if err != nil {
		return nil, err
	}
	cfg := defaultUpstreamMonitorConfig()
	if strings.TrimSpace(raw) == "" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, err
	}
	normalizeUpstreamMonitorConfig(cfg)
	return cfg, nil
}

func maskUpstreamMonitorSecrets(cfg *UpstreamMonitorConfig) {
	if cfg == nil {
		return
	}
	for i := range cfg.Sources {
		cfg.Sources[i].AuthConfigured = strings.TrimSpace(cfg.Sources[i].AuthToken) != ""
		cfg.Sources[i].AuthToken = ""
	}
}

func (s *SettingService) notifyUpstreamMonitorAlerts(ctx context.Context, cfg *UpstreamMonitorConfig) {
	if s == nil || s.notificationEmailService == nil || cfg == nil {
		return
	}

	snapshot, err := s.PreviewUpstreamMonitorConfig(ctx, cfg)
	if err != nil || snapshot == nil {
		return
	}

	recipients := s.getUpstreamMonitorAlertRecipients(ctx)
	if len(recipients) == 0 {
		return
	}

	for _, row := range snapshot.GroupRows {
		multiplierChanged, multiplierStateKey, multiplierState := s.shouldSendUpstreamMonitorMultiplierChangeAlert(ctx, row)
		severity := upstreamMonitorAlertSeverity(cfg, row)
		shouldSend, stateKey := s.shouldSendUpstreamMonitorAlert(ctx, cfg, row, severity)

		if !shouldSend || severity == "" {
			if multiplierChanged {
				changeRow := upstreamMonitorAlertRowWithIssue(row, "change", "reference multiplier changed")
				allSent := true
				for _, recipient := range recipients {
					if err := s.notificationEmailService.Send(ctx, NotificationEmailSendInput{
						Event:          NotificationEmailEventUpstreamMonitorAlert,
						RecipientEmail: recipient.Email,
						RecipientName:  recipient.Name,
						Variables:      upstreamMonitorAlertEmailVariables(changeRow, "change"),
					}); err != nil {
						allSent = false
						continue
					}
				}
				if allSent && multiplierStateKey != "" {
					_ = s.settingRepo.Set(ctx, multiplierStateKey, multiplierState)
				}
			}
			continue
		}

		alertRow := row
		if multiplierChanged {
			alertRow = upstreamMonitorAlertRowWithIssue(row, row.Status, "reference multiplier changed")
		}
		allSent := true
		for _, recipient := range recipients {
			if err := s.notificationEmailService.Send(ctx, NotificationEmailSendInput{
				Event:          NotificationEmailEventUpstreamMonitorAlert,
				RecipientEmail: recipient.Email,
				RecipientName:  recipient.Name,
				Variables:      upstreamMonitorAlertEmailVariables(alertRow, severity),
			}); err != nil {
				allSent = false
				continue
			}
		}
		if allSent && stateKey != "" {
			_ = s.settingRepo.Set(ctx, stateKey, severity)
		}
		if allSent && multiplierChanged && multiplierStateKey != "" {
			_ = s.settingRepo.Set(ctx, multiplierStateKey, multiplierState)
		}
	}
	for _, row := range snapshot.AccountRows {
		severity := upstreamMonitorAccountAlertSeverity(cfg, row)
		shouldSend, stateKey := s.shouldSendUpstreamMonitorAccountAlert(ctx, cfg, row, severity)
		if !shouldSend {
			continue
		}
		if severity == "" {
			continue
		}
		allSent := true
		for _, recipient := range recipients {
			if err := s.notificationEmailService.Send(ctx, NotificationEmailSendInput{
				Event:          NotificationEmailEventUpstreamMonitorAlert,
				RecipientEmail: recipient.Email,
				RecipientName:  recipient.Name,
				Variables:      upstreamMonitorAccountAlertEmailVariables(row, severity),
			}); err != nil {
				allSent = false
				continue
			}
		}
		if allSent && stateKey != "" {
			_ = s.settingRepo.Set(ctx, stateKey, severity)
		}
	}
}

func upstreamMonitorAlertSeverity(cfg *UpstreamMonitorConfig, row UpstreamMonitorPreviewGroupRow) string {
	if cfg == nil {
		return ""
	}
	if row.LocalGroupMultiplier > 0 && row.ReferenceMultiplier > 0 && row.LocalGroupMultiplier < row.ReferenceMultiplier {
		return "critical"
	}
	if !cfg.NotifyOnCriticalOnly && row.Status == "warning" {
		return "warning"
	}
	return ""
}

func upstreamMonitorAccountAlertSeverity(cfg *UpstreamMonitorConfig, row UpstreamMonitorPreviewAccountRow) string {
	if cfg == nil {
		return ""
	}
	if row.HighestGroupMultiplier > 0 && row.EstimatedCostMultiplier > 0 && row.HighestGroupMultiplier < row.EstimatedCostMultiplier {
		return "critical"
	}
	if !cfg.NotifyOnCriticalOnly && row.Status == "warning" {
		return "warning"
	}
	return ""
}

func upstreamMonitorAlertRowWithIssue(row UpstreamMonitorPreviewGroupRow, status, issue string) UpstreamMonitorPreviewGroupRow {
	out := row
	if strings.TrimSpace(status) != "" {
		out.Status = status
	}
	out.Issues = append([]string{}, row.Issues...)
	if strings.TrimSpace(issue) != "" {
		out.Issues = append(out.Issues, strings.TrimSpace(issue))
	}
	out.Issues = dedupeUpstreamMonitorStrings(out.Issues)
	return out
}

func (s *SettingService) shouldSendUpstreamMonitorAlert(ctx context.Context, cfg *UpstreamMonitorConfig, row UpstreamMonitorPreviewGroupRow, severity string) (bool, string) {
	if s == nil || s.settingRepo == nil || cfg == nil {
		return false, ""
	}
	key := upstreamMonitorAlertStateKey(row.MappingID)
	currentState := strings.TrimSpace(severity)
	previous, err := s.settingRepo.GetValue(ctx, key)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return false, ""
	}

	if currentState == "" {
		if previous != "" {
			_ = s.settingRepo.Delete(ctx, key)
		}
		return false, ""
	}
	if previous == currentState {
		return false, ""
	}
	return true, key
}

func (s *SettingService) shouldSendUpstreamMonitorMultiplierChangeAlert(ctx context.Context, row UpstreamMonitorPreviewGroupRow) (bool, string, string) {
	if s == nil || s.settingRepo == nil {
		return false, "", ""
	}
	if row.MappingID == "" || row.ReferenceMultiplier <= 0 {
		return false, "", ""
	}

	key := upstreamMonitorMultiplierStateKey(row.MappingID)
	currentState := upstreamMonitorMultiplierState(row.ReferenceMultiplier)
	previous, err := s.settingRepo.GetValue(ctx, key)
	if errors.Is(err, ErrSettingNotFound) {
		_ = s.settingRepo.Set(ctx, key, currentState)
		return false, "", ""
	}
	if err != nil {
		return false, "", ""
	}
	if previous == currentState {
		return false, "", ""
	}
	return true, key, currentState
}

func (s *SettingService) shouldSendUpstreamMonitorAccountAlert(ctx context.Context, cfg *UpstreamMonitorConfig, row UpstreamMonitorPreviewAccountRow, severity string) (bool, string) {
	if s == nil || s.settingRepo == nil || cfg == nil {
		return false, ""
	}
	key := upstreamMonitorAccountAlertStateKey(row.SourceID, row.AccountID)
	currentState := strings.TrimSpace(severity)
	previous, err := s.settingRepo.GetValue(ctx, key)
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return false, ""
	}

	if currentState == "" {
		if previous != "" {
			_ = s.settingRepo.Delete(ctx, key)
		}
		return false, ""
	}
	if previous == currentState {
		return false, ""
	}
	return true, key
}

func upstreamMonitorAlertStateKey(mappingID string) string {
	mappingID = strings.TrimSpace(mappingID)
	if mappingID == "" {
		mappingID = "unknown"
	}
	return "upstream_monitor_alert_state:" + mappingID
}

func upstreamMonitorMultiplierStateKey(mappingID string) string {
	mappingID = strings.TrimSpace(mappingID)
	if mappingID == "" {
		mappingID = "unknown"
	}
	return "upstream_monitor_multiplier_state:" + mappingID
}

func upstreamMonitorMultiplierState(referenceMultiplier float64) string {
	return strconv.FormatFloat(referenceMultiplier, 'f', 8, 64)
}

func upstreamMonitorAccountAlertStateKey(sourceID string, accountID int64) string {
	sourceID = strings.TrimSpace(sourceID)
	if sourceID == "" {
		sourceID = "unknown"
	}
	return "upstream_monitor_alert_state:account:" + sourceID + ":" + strconv.FormatInt(accountID, 10)
}

func (s *SettingService) getUpstreamMonitorAlertRecipients(ctx context.Context) []upstreamMonitorAlertRecipient {
	seen := make(map[string]struct{})
	recipients := make([]upstreamMonitorAlertRecipient, 0, 4)

	if s.settingRepo != nil {
		if raw, err := s.settingRepo.GetValue(ctx, SettingKeyAccountQuotaNotifyEmails); err == nil {
			for _, entry := range ParseNotifyEmails(raw) {
				if entry.Disabled || !entry.Verified {
					continue
				}
				email := strings.TrimSpace(entry.Email)
				if email == "" {
					continue
				}
				lower := strings.ToLower(email)
				if _, ok := seen[lower]; ok {
					continue
				}
				seen[lower] = struct{}{}
				recipients = append(recipients, upstreamMonitorAlertRecipient{
					Email: email,
					Name:  emailRecipientName(email),
				})
			}
		}
	}

	if len(recipients) == 0 && s.upstreamMonitorAdminReader != nil {
		admin, err := s.upstreamMonitorAdminReader.GetFirstAdmin(ctx)
		if err == nil && admin != nil {
			email := strings.TrimSpace(admin.Email)
			if email != "" {
				recipients = append(recipients, upstreamMonitorAlertRecipient{
					Email: email,
					Name:  firstNonEmpty(strings.TrimSpace(admin.Username), emailRecipientName(email)),
				})
			}
		}
	}

	return recipients
}

func upstreamMonitorAlertEmailVariables(row UpstreamMonitorPreviewGroupRow, severity string) map[string]string {
	sourceNames := "-"
	if len(row.SourceNames) > 0 {
		sourceNames = strings.Join(row.SourceNames, ", ")
	}
	issues := "-"
	if len(row.Issues) > 0 {
		issues = strings.Join(row.Issues, ", ")
	}
	return map[string]string{
		"group_name":            firstNonEmpty(strings.TrimSpace(row.LocalGroup), "-"),
		"model_family":          firstNonEmpty(strings.TrimSpace(row.ModelFamily), "-"),
		"severity":              severity,
		"status":                firstNonEmpty(strings.TrimSpace(row.Status), "-"),
		"local_multiplier":      formatUpstreamMonitorFloat(row.LocalGroupMultiplier),
		"reference_multiplier":  formatUpstreamMonitorFloat(row.ReferenceMultiplier),
		"estimated_margin_rate": formatUpstreamMonitorPercent(row.EstimatedMarginRate),
		"source_names":          sourceNames,
		"issues":                issues,
		"triggered_at":          time.Now().UTC().Format(time.RFC3339),
	}
}

func upstreamMonitorAccountAlertEmailVariables(row UpstreamMonitorPreviewAccountRow, severity string) map[string]string {
	sourceNames := firstNonEmpty(strings.TrimSpace(row.SourceName), "-")
	issues := "-"
	if len(row.Issues) > 0 {
		issues = strings.Join(row.Issues, ", ")
	}
	groupNames := "-"
	if len(row.GroupNames) > 0 {
		groupNames = strings.Join(row.GroupNames, ", ")
	}
	return map[string]string{
		"group_name":            firstNonEmpty(strings.TrimSpace(row.AccountName), "Account #"+strconv.FormatInt(row.AccountID, 10)),
		"model_family":          firstNonEmpty(strings.TrimSpace(row.AccountPlatform), "-"),
		"severity":              severity,
		"status":                firstNonEmpty(strings.TrimSpace(row.Status), "-"),
		"local_multiplier":      formatUpstreamMonitorFloat(row.HighestGroupMultiplier),
		"reference_multiplier":  formatUpstreamMonitorFloat(row.EstimatedCostMultiplier),
		"estimated_margin_rate": formatUpstreamMonitorPercent(row.EstimatedMarginRate),
		"source_names":          sourceNames + " / " + groupNames,
		"issues":                issues,
		"triggered_at":          time.Now().UTC().Format(time.RFC3339),
	}
}

func formatUpstreamMonitorFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 4, 64)
}

func formatUpstreamMonitorPercent(value float64) string {
	return strconv.FormatFloat(value*100, 'f', 2, 64) + "%"
}
