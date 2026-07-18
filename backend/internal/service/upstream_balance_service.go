package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

const (
	UpstreamBalanceProviderAuto        = "auto"
	UpstreamBalanceProviderDeepSeek    = "deepseek"
	UpstreamBalanceProviderStepFun     = "stepfun"
	UpstreamBalanceProviderSiliconFlow = "siliconflow"
	UpstreamBalanceProviderNewAPI      = "newapi"
	UpstreamBalanceProviderSub2API     = "sub2api"
	UpstreamBalanceProviderCustom      = "custom"

	upstreamBalanceRefreshInterval = 10 * time.Minute
	upstreamBalanceRequestTimeout  = 15 * time.Second
	upstreamBalanceRefreshTimeout  = 3 * time.Minute
	upstreamBalanceStaleAfter      = 30 * time.Minute
	upstreamBalanceLockKey         = "sub2api:upstream_balance:refresh_lock"
	upstreamBalanceLockTTL         = 9 * time.Minute
)

const (
	upstreamBalanceEnabledKey         = "upstream_balance_enabled"
	upstreamBalanceProviderKey        = "upstream_balance_provider"
	upstreamBalancePlatformNameKey    = "upstream_balance_platform_name"
	upstreamBalanceThresholdKey       = "upstream_balance_threshold"
	upstreamBalanceNotifyEnabledKey   = "upstream_balance_notify_enabled"
	upstreamBalanceEndpointKey        = "upstream_balance_endpoint"
	upstreamBalanceJSONPathKey        = "upstream_balance_json_path"
	upstreamBalanceDivisorKey         = "upstream_balance_divisor"
	upstreamBalanceFundingKey         = "upstream_balance_funding_key"
	upstreamBalanceAuthModeKey        = "upstream_balance_auth_mode"
	upstreamBalanceAuthUsernameKey    = "upstream_balance_auth_username"
	upstreamBalanceLastAmountKey      = "upstream_balance_last_amount"
	upstreamBalanceLastCheckedAtKey   = "upstream_balance_last_checked_at"
	upstreamBalanceLastSuccessAtKey   = "upstream_balance_last_success_at"
	upstreamBalanceLastErrorKey       = "upstream_balance_last_error"
	upstreamBalanceAlertActiveKey     = "upstream_balance_alert_active"
	upstreamBalanceLastAlertedAtKey   = "upstream_balance_last_alerted_at"
	upstreamBalanceLastRecoveredAtKey = "upstream_balance_last_recovered_at"
)

const upstreamBalanceAuthTokenCredentialKey = "upstream_balance_auth_token"

var upstreamBalanceReleaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)

type UpstreamBalanceConfig struct {
	Enabled        bool    `json:"enabled"`
	Provider       string  `json:"provider"`
	PlatformName   string  `json:"platform_name"`
	Threshold      float64 `json:"threshold"`
	NotifyEnabled  bool    `json:"notify_enabled"`
	Endpoint       string  `json:"endpoint,omitempty"`
	JSONPath       string  `json:"json_path,omitempty"`
	Divisor        float64 `json:"divisor"`
	FundingKey     string  `json:"funding_key,omitempty"`
	AuthMode       string  `json:"auth_mode,omitempty"`
	AuthUsername   string  `json:"auth_username,omitempty"`
	AuthToken      string  `json:"auth_token,omitempty"`
	AuthCleared    bool    `json:"auth_cleared,omitempty"`
	AuthConfigured bool    `json:"auth_configured"`
}

type UpstreamBalanceSnapshot struct {
	Amount        *float64   `json:"amount,omitempty"`
	Currency      string     `json:"currency"`
	LastCheckedAt *time.Time `json:"last_checked_at,omitempty"`
	LastSuccessAt *time.Time `json:"last_success_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
	AlertActive   bool       `json:"alert_active"`
	Stale         bool       `json:"stale"`
}

type UpstreamBalanceSource struct {
	AccountID   int64                   `json:"account_id"`
	AccountName string                  `json:"account_name"`
	Protocol    string                  `json:"protocol"`
	AccountType string                  `json:"account_type"`
	BaseURL     string                  `json:"base_url"`
	Config      UpstreamBalanceConfig   `json:"config"`
	Snapshot    UpstreamBalanceSnapshot `json:"snapshot"`
}

type UpstreamBalancePlatformSummary struct {
	PlatformName    string     `json:"platform_name"`
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency"`
	AccountCount    int        `json:"account_count"`
	FundingCount    int        `json:"funding_count"`
	LowBalanceCount int        `json:"low_balance_count"`
	ErrorCount      int        `json:"error_count"`
	StaleCount      int        `json:"stale_count"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type UpstreamBalanceOverview struct {
	TotalAmount float64                          `json:"total_amount"`
	Currency    string                           `json:"currency"`
	Platforms   []UpstreamBalancePlatformSummary `json:"platforms"`
	Sources     []UpstreamBalanceSource          `json:"sources"`
}

type UpstreamBalanceRefreshResult struct {
	Refreshed int `json:"refreshed"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

type upstreamBalanceFundingSnapshot struct {
	summary  *UpstreamBalancePlatformSummary
	snapshot UpstreamBalanceSnapshot
}

type UpstreamBalanceService struct {
	accountRepo   AccountRepository
	proxyRepo     ProxyRepository
	notifyService *BalanceNotifyService
	redisClient   *redis.Client

	instanceID string
	ctx        context.Context
	cancel     context.CancelFunc
	stopCh     chan struct{}
	startOnce  sync.Once
	stopOnce   sync.Once
	wg         sync.WaitGroup
	refreshMu  sync.Mutex
}

func NewUpstreamBalanceService(
	accountRepo AccountRepository,
	proxyRepo ProxyRepository,
	notifyService *BalanceNotifyService,
	redisClient *redis.Client,
) *UpstreamBalanceService {
	serviceCtx, cancel := context.WithCancel(context.Background())
	return &UpstreamBalanceService{
		accountRepo:   accountRepo,
		proxyRepo:     proxyRepo,
		notifyService: notifyService,
		redisClient:   redisClient,
		instanceID:    uuid.NewString(),
		ctx:           serviceCtx,
		cancel:        cancel,
		stopCh:        make(chan struct{}),
	}
}

func (s *UpstreamBalanceService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		s.wg.Add(1)
		go s.run()
	})
}

func (s *UpstreamBalanceService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		s.cancel()
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *UpstreamBalanceService) run() {
	defer s.wg.Done()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			ctx, cancel := context.WithTimeout(s.ctx, upstreamBalanceRefreshTimeout)
			if release, ok := s.acquireDistributedLock(ctx); ok {
				_, _ = s.Refresh(ctx, nil)
				release()
			}
			cancel()
			timer.Reset(upstreamBalanceRefreshInterval)
		case <-s.stopCh:
			return
		}
	}
}

func (s *UpstreamBalanceService) acquireDistributedLock(ctx context.Context) (func(), bool) {
	if s.redisClient == nil {
		return func() {}, true
	}
	ok, err := s.redisClient.SetNX(ctx, upstreamBalanceLockKey, s.instanceID, upstreamBalanceLockTTL).Result()
	if err != nil {
		slog.Warn("upstream balance lock unavailable; continuing with local lock", "error", err)
		return func() {}, true
	}
	if !ok {
		return nil, false
	}
	return func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, _ = upstreamBalanceReleaseScript.Run(releaseCtx, s.redisClient, []string{upstreamBalanceLockKey}, s.instanceID).Result()
	}, true
}

func (s *UpstreamBalanceService) Overview(ctx context.Context) (*UpstreamBalanceOverview, error) {
	accounts, err := s.accountRepo.ListAllWithFilters(ctx, "", "", "", "", 0, "")
	if err != nil {
		return nil, fmt.Errorf("list balance accounts: %w", err)
	}

	now := time.Now().UTC()
	overview := &UpstreamBalanceOverview{Currency: "CNY", Sources: make([]UpstreamBalanceSource, 0)}
	platforms := make(map[string]*UpstreamBalancePlatformSummary)
	fundingSnapshots := make(map[string]upstreamBalanceFundingSnapshot)

	for i := range accounts {
		account := &accounts[i]
		if !isUpstreamBalanceEligible(account) {
			continue
		}
		cfg := upstreamBalanceConfigFromAccount(account)
		snapshot := upstreamBalanceSnapshotFromAccount(account, now)
		source := UpstreamBalanceSource{
			AccountID:   account.ID,
			AccountName: account.Name,
			Protocol:    account.Platform,
			AccountType: account.Type,
			BaseURL:     strings.TrimSpace(account.GetCredential("base_url")),
			Config:      cfg,
			Snapshot:    snapshot,
		}
		overview.Sources = append(overview.Sources, source)
		if !cfg.Enabled {
			continue
		}

		platformName := balancePlatformName(account, cfg)
		summary := platforms[platformName]
		if summary == nil {
			summary = &UpstreamBalancePlatformSummary{PlatformName: platformName, Currency: "CNY"}
			platforms[platformName] = summary
		}
		summary.AccountCount++
		if snapshot.AlertActive {
			summary.LowBalanceCount++
		}
		if snapshot.LastError != "" {
			summary.ErrorCount++
		}
		if snapshot.Stale {
			summary.StaleCount++
		}
		if snapshot.LastSuccessAt != nil && (summary.UpdatedAt == nil || snapshot.LastSuccessAt.Before(*summary.UpdatedAt)) {
			ts := *snapshot.LastSuccessAt
			summary.UpdatedAt = &ts
		}

		fundingKey := balanceFundingIdentity(account, cfg)
		candidate, exists := fundingSnapshots[fundingKey]
		if !exists || preferUpstreamBalanceSnapshot(snapshot, candidate.snapshot) {
			fundingSnapshots[fundingKey] = upstreamBalanceFundingSnapshot{summary: summary, snapshot: snapshot}
		}
	}

	for _, funding := range fundingSnapshots {
		funding.summary.FundingCount++
		if funding.snapshot.Amount != nil && !funding.snapshot.Stale {
			funding.summary.Amount += *funding.snapshot.Amount
			overview.TotalAmount += *funding.snapshot.Amount
		}
	}

	overview.Platforms = make([]UpstreamBalancePlatformSummary, 0, len(platforms))
	for _, summary := range platforms {
		overview.Platforms = append(overview.Platforms, *summary)
	}
	sortBalanceOverview(overview)
	return overview, nil
}

func preferUpstreamBalanceSnapshot(candidate, current UpstreamBalanceSnapshot) bool {
	if candidate.Stale != current.Stale {
		return !candidate.Stale
	}
	if candidate.LastSuccessAt == nil {
		return false
	}
	return current.LastSuccessAt == nil || candidate.LastSuccessAt.After(*current.LastSuccessAt)
}

func (s *UpstreamBalanceService) Configure(ctx context.Context, accountID int64, cfg UpstreamBalanceConfig) error {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if !isUpstreamBalanceEligible(account) {
		return infraerrors.BadRequest("UPSTREAM_BALANCE_ACCOUNT_UNSUPPORTED", "account must use an API key or upstream credential with api_key")
	}
	if err := validateUpstreamBalanceConfig(account, &cfg); err != nil {
		return infraerrors.BadRequest("UPSTREAM_BALANCE_CONFIG_INVALID", err.Error())
	}
	updates := map[string]any{
		upstreamBalanceEnabledKey:       cfg.Enabled,
		upstreamBalanceProviderKey:      normalizeBalanceProvider(cfg.Provider),
		upstreamBalancePlatformNameKey:  strings.TrimSpace(cfg.PlatformName),
		upstreamBalanceThresholdKey:     cfg.Threshold,
		upstreamBalanceNotifyEnabledKey: cfg.NotifyEnabled,
		upstreamBalanceEndpointKey:      strings.TrimSpace(cfg.Endpoint),
		upstreamBalanceJSONPathKey:      strings.TrimSpace(cfg.JSONPath),
		upstreamBalanceDivisorKey:       normalizeBalanceDivisor(cfg.Divisor),
		upstreamBalanceFundingKey:       strings.TrimSpace(cfg.FundingKey),
		upstreamBalanceAuthModeKey:      normalizeBalanceAuthMode(cfg.AuthMode),
		upstreamBalanceAuthUsernameKey:  strings.TrimSpace(cfg.AuthUsername),
	}
	if !cfg.Enabled {
		updates[upstreamBalanceAlertActiveKey] = false
	}
	if cfg.AuthCleared || strings.TrimSpace(cfg.AuthToken) != "" {
		credentials := shallowCopyMap(account.Credentials)
		if cfg.AuthCleared {
			delete(credentials, upstreamBalanceAuthTokenCredentialKey)
		} else {
			credentials[upstreamBalanceAuthTokenCredentialKey] = strings.TrimSpace(cfg.AuthToken)
		}
		if err := persistAccountCredentials(ctx, s.accountRepo, account, credentials); err != nil {
			return fmt.Errorf("save upstream balance credentials: %w", err)
		}
	}
	return s.accountRepo.UpdateExtra(ctx, accountID, updates)
}

func (s *UpstreamBalanceService) Refresh(ctx context.Context, accountID *int64) (*UpstreamBalanceRefreshResult, error) {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	accounts, err := s.accountRepo.ListAllWithFilters(ctx, "", "", "", "", 0, "")
	if err != nil {
		return nil, fmt.Errorf("list balance accounts: %w", err)
	}
	targets := make([]*Account, 0)
	for i := range accounts {
		account := &accounts[i]
		if accountID != nil && account.ID != *accountID {
			continue
		}
		if isUpstreamBalanceEligible(account) && upstreamBalanceConfigFromAccount(account).Enabled {
			targets = append(targets, account)
		}
	}
	result := &UpstreamBalanceRefreshResult{Refreshed: len(targets)}
	var resultMu sync.Mutex
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(4)
	for _, account := range targets {
		account := account
		group.Go(func() error {
			err := s.refreshAccount(groupCtx, account)
			resultMu.Lock()
			if err != nil {
				result.Failed++
			} else {
				result.Succeeded++
			}
			resultMu.Unlock()
			return nil
		})
	}
	_ = group.Wait()
	return result, nil
}

func (s *UpstreamBalanceService) refreshAccount(ctx context.Context, account *Account) error {
	cfg := upstreamBalanceConfigFromAccount(account)
	now := time.Now().UTC()
	amount, err := s.fetchBalance(ctx, account, cfg)
	if err != nil {
		updates := map[string]any{
			upstreamBalanceLastCheckedAtKey: now.Format(time.RFC3339),
			upstreamBalanceLastErrorKey:     truncateBalanceError(err.Error()),
		}
		if updateErr := s.accountRepo.UpdateExtra(ctx, account.ID, updates); updateErr != nil {
			return updateErr
		}
		return err
	}

	wasAlertActive := account.getExtraBool(upstreamBalanceAlertActiveKey)
	isAlertActive := cfg.NotifyEnabled && cfg.Threshold > 0 && amount <= cfg.Threshold
	lastAlertedAt := parseBalanceTime(account.getExtraString(upstreamBalanceLastAlertedAtKey))
	lastRecoveredAt := parseBalanceTime(account.getExtraString(upstreamBalanceLastRecoveredAtKey))
	shouldNotify := shouldNotifyUpstreamBalance(isAlertActive, wasAlertActive, lastAlertedAt, lastRecoveredAt)
	updates := map[string]any{
		upstreamBalanceLastAmountKey:    amount,
		upstreamBalanceLastCheckedAtKey: now.Format(time.RFC3339),
		upstreamBalanceLastSuccessAtKey: now.Format(time.RFC3339),
		upstreamBalanceLastErrorKey:     "",
		upstreamBalanceAlertActiveKey:   isAlertActive,
	}
	if !isAlertActive && wasAlertActive {
		updates[upstreamBalanceLastRecoveredAtKey] = now.Format(time.RFC3339)
	}
	if err := s.accountRepo.UpdateExtra(ctx, account.ID, updates); err != nil {
		return err
	}
	if shouldNotify {
		if s.notifyService == nil {
			slog.Warn("upstream balance alert is pending because notification service is unavailable", "account_id", account.ID)
		} else if err := s.notifyService.NotifyUpstreamBalanceLow(ctx, account.ID, account.Name, balancePlatformName(account, cfg), amount, cfg.Threshold); err != nil {
			slog.Warn("upstream balance alert delivery failed; will retry", "account_id", account.ID, "error", err)
		} else if err := s.accountRepo.UpdateExtra(ctx, account.ID, map[string]any{
			upstreamBalanceLastAlertedAtKey: now.Format(time.RFC3339),
		}); err != nil {
			return err
		}
	}
	return nil
}

func shouldNotifyUpstreamBalance(isActive, wasActive bool, lastAlertedAt, lastRecoveredAt *time.Time) bool {
	return isActive && (!wasActive || lastAlertedAt == nil || (lastRecoveredAt != nil && lastRecoveredAt.After(*lastAlertedAt)))
}

func (s *UpstreamBalanceService) fetchBalance(ctx context.Context, account *Account, cfg UpstreamBalanceConfig) (float64, error) {
	provider := normalizeBalanceProvider(cfg.Provider)
	baseURL := strings.TrimSpace(account.GetCredential("base_url"))
	if provider == UpstreamBalanceProviderAuto {
		provider = detectBalanceProvider(baseURL)
	}
	requestURL, err := resolveBalanceRequestURL(baseURL, provider, cfg.Endpoint)
	if err != nil {
		return 0, err
	}
	client := req.C().SetTimeout(upstreamBalanceRequestTimeout)
	if proxyURL := s.resolveProxyURL(ctx, account); proxyURL != "" {
		client.SetProxyURL(proxyURL)
	}
	if provider == UpstreamBalanceProviderNewAPI {
		statusURL, err := resolveBalanceSameOriginURL(baseURL, "/api/status")
		if err != nil {
			return 0, err
		}
		headers, err := upstreamBalanceAuthHeaders(ctx, client, account, cfg, UpstreamBalanceProviderNewAPI)
		if err != nil {
			return 0, err
		}
		return fetchNewAPIWalletCNYBalance(ctx, client, statusURL, requestURL, headers)
	}
	if provider == UpstreamBalanceProviderSub2API {
		headers, err := upstreamBalanceAuthHeaders(ctx, client, account, cfg, UpstreamBalanceProviderSub2API)
		if err != nil {
			return 0, err
		}
		return fetchSub2APIWalletCNYBalance(ctx, client, requestURL, headers)
	}
	apiKey := strings.TrimSpace(account.GetCredential("api_key"))
	if apiKey == "" {
		return 0, fmt.Errorf("api_key is missing")
	}
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Accept", "application/json").
		Get(requestURL)
	if err != nil {
		return 0, fmt.Errorf("balance request failed: %w", err)
	}
	if !resp.IsSuccessState() {
		return 0, fmt.Errorf("balance endpoint returned HTTP %d", resp.StatusCode)
	}
	var payload any
	if err := json.Unmarshal(resp.Bytes(), &payload); err != nil {
		return 0, fmt.Errorf("invalid balance response JSON: %w", err)
	}
	return parseUpstreamBalanceAmount(provider, payload, cfg.JSONPath, cfg.Divisor)
}

func (s *UpstreamBalanceService) resolveProxyURL(ctx context.Context, account *Account) string {
	if account == nil || account.ProxyID == nil || s.proxyRepo == nil {
		return ""
	}
	proxy, err := s.proxyRepo.GetByID(ctx, *account.ProxyID)
	if err != nil || proxy == nil || !proxy.IsActive() {
		return ""
	}
	return proxy.URL()
}

func upstreamBalanceConfigFromAccount(account *Account) UpstreamBalanceConfig {
	cfg := UpstreamBalanceConfig{Provider: UpstreamBalanceProviderAuto, Divisor: 1}
	if account == nil {
		return cfg
	}
	cfg.Enabled = account.getExtraBool(upstreamBalanceEnabledKey)
	cfg.Provider = normalizeBalanceProvider(account.getExtraString(upstreamBalanceProviderKey))
	cfg.PlatformName = strings.TrimSpace(account.getExtraString(upstreamBalancePlatformNameKey))
	cfg.Threshold = account.getExtraFloat64(upstreamBalanceThresholdKey)
	cfg.NotifyEnabled = account.getExtraBool(upstreamBalanceNotifyEnabledKey)
	cfg.Endpoint = strings.TrimSpace(account.getExtraString(upstreamBalanceEndpointKey))
	cfg.JSONPath = strings.TrimSpace(account.getExtraString(upstreamBalanceJSONPathKey))
	cfg.Divisor = normalizeBalanceDivisor(account.getExtraFloat64(upstreamBalanceDivisorKey))
	cfg.FundingKey = strings.TrimSpace(account.getExtraString(upstreamBalanceFundingKey))
	cfg.AuthMode = normalizeBalanceAuthMode(account.getExtraString(upstreamBalanceAuthModeKey))
	cfg.AuthUsername = strings.TrimSpace(account.getExtraString(upstreamBalanceAuthUsernameKey))
	cfg.AuthConfigured = balanceAuthConfigured(account, cfg)
	return cfg
}

func upstreamBalanceSnapshotFromAccount(account *Account, now time.Time) UpstreamBalanceSnapshot {
	snapshot := UpstreamBalanceSnapshot{Currency: "CNY"}
	if account == nil {
		return snapshot
	}
	if account.Extra != nil {
		if raw, ok := account.Extra[upstreamBalanceLastAmountKey]; ok {
			if amount, ok := numberFromAny(raw); ok {
				snapshot.Amount = &amount
			}
		}
	}
	snapshot.LastCheckedAt = parseBalanceTime(account.getExtraString(upstreamBalanceLastCheckedAtKey))
	snapshot.LastSuccessAt = parseBalanceTime(account.getExtraString(upstreamBalanceLastSuccessAtKey))
	snapshot.LastError = strings.TrimSpace(account.getExtraString(upstreamBalanceLastErrorKey))
	snapshot.AlertActive = account.getExtraBool(upstreamBalanceAlertActiveKey)
	snapshot.Stale = snapshot.LastSuccessAt == nil || now.Sub(*snapshot.LastSuccessAt) > upstreamBalanceStaleAfter
	return snapshot
}

func validateUpstreamBalanceConfig(account *Account, cfg *UpstreamBalanceConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	cfg.Provider = normalizeBalanceProvider(cfg.Provider)
	cfg.PlatformName = strings.TrimSpace(cfg.PlatformName)
	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	cfg.JSONPath = strings.TrimSpace(cfg.JSONPath)
	cfg.FundingKey = strings.TrimSpace(cfg.FundingKey)
	cfg.AuthMode = normalizeBalanceAuthMode(cfg.AuthMode)
	cfg.AuthUsername = strings.TrimSpace(cfg.AuthUsername)
	cfg.AuthToken = strings.TrimSpace(cfg.AuthToken)
	cfg.Divisor = normalizeBalanceDivisor(cfg.Divisor)
	if cfg.Threshold < 0 || math.IsNaN(cfg.Threshold) || math.IsInf(cfg.Threshold, 0) {
		return fmt.Errorf("threshold must be a finite number greater than or equal to 0")
	}
	if cfg.PlatformName == "" && cfg.Enabled {
		return fmt.Errorf("platform_name is required when balance collection is enabled")
	}
	switch cfg.Provider {
	case UpstreamBalanceProviderAuto, UpstreamBalanceProviderDeepSeek, UpstreamBalanceProviderStepFun, UpstreamBalanceProviderSiliconFlow, UpstreamBalanceProviderNewAPI, UpstreamBalanceProviderSub2API:
	case UpstreamBalanceProviderCustom:
		if cfg.Endpoint == "" || cfg.JSONPath == "" {
			return fmt.Errorf("custom provider requires endpoint and json_path")
		}
	default:
		return fmt.Errorf("unsupported balance provider")
	}
	if cfg.Enabled {
		provider := cfg.Provider
		if provider == UpstreamBalanceProviderAuto {
			provider = detectBalanceProvider(account.GetCredential("base_url"))
			if provider == "" {
				return fmt.Errorf("provider cannot be detected; select custom or a built-in provider")
			}
		}
		if _, err := resolveBalanceRequestURL(account.GetCredential("base_url"), provider, cfg.Endpoint); err != nil {
			return err
		}
		if provider == UpstreamBalanceProviderNewAPI || provider == UpstreamBalanceProviderSub2API {
			if err := validateBalanceWalletAuth(account, cfg, provider); err != nil {
				return err
			}
		}
	}
	return nil
}

func normalizeBalanceProvider(provider string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return UpstreamBalanceProviderAuto
	}
	return provider
}

func normalizeBalanceAuthMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
}

func balanceAuthConfigured(account *Account, cfg UpstreamBalanceConfig) bool {
	token := ""
	if account != nil {
		token = strings.TrimSpace(account.GetCredential(upstreamBalanceAuthTokenCredentialKey))
	}
	if strings.TrimSpace(cfg.AuthToken) != "" {
		token = strings.TrimSpace(cfg.AuthToken)
	}
	if cfg.AuthCleared || token == "" {
		return false
	}
	switch normalizeBalanceAuthMode(cfg.AuthMode) {
	case "login":
		return strings.TrimSpace(cfg.AuthUsername) != ""
	case "bearer", "cookie":
		return true
	default:
		return false
	}
}

func validateBalanceWalletAuth(account *Account, cfg *UpstreamBalanceConfig, provider string) error {
	mode := normalizeBalanceAuthMode(cfg.AuthMode)
	if provider == UpstreamBalanceProviderNewAPI && mode != "login" && mode != "cookie" {
		return fmt.Errorf("New API wallet balance requires login or cookie authentication")
	}
	if provider == UpstreamBalanceProviderSub2API && mode != "login" && mode != "bearer" {
		return fmt.Errorf("Sub2API wallet balance requires login or bearer authentication")
	}
	if mode == "login" && strings.TrimSpace(cfg.AuthUsername) == "" {
		return fmt.Errorf("login email is required")
	}
	if !balanceAuthConfigured(account, *cfg) {
		return fmt.Errorf("wallet authentication credential is required")
	}
	return nil
}

func normalizeBalanceDivisor(divisor float64) float64 {
	if divisor <= 0 || math.IsNaN(divisor) || math.IsInf(divisor, 0) {
		return 1
	}
	return divisor
}

func isUpstreamBalanceEligible(account *Account) bool {
	if account == nil || (account.Type != AccountTypeAPIKey && account.Type != AccountTypeUpstream) {
		return false
	}
	return strings.TrimSpace(account.GetCredential("api_key")) != ""
}

func detectBalanceProvider(baseURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	switch {
	case host == "api.deepseek.com" || strings.HasSuffix(host, ".deepseek.com"):
		return UpstreamBalanceProviderDeepSeek
	case host == "api.stepfun.com" || host == "api.stepfun.ai" || strings.HasSuffix(host, ".stepfun.com") || strings.HasSuffix(host, ".stepfun.ai"):
		return UpstreamBalanceProviderStepFun
	case host == "api.siliconflow.cn" || strings.HasSuffix(host, ".siliconflow.cn"):
		return UpstreamBalanceProviderSiliconFlow
	default:
		return ""
	}
}

func resolveBalanceRequestURL(baseURL, provider, customEndpoint string) (string, error) {
	if provider != UpstreamBalanceProviderCustom && provider != UpstreamBalanceProviderNewAPI && provider != UpstreamBalanceProviderSub2API && detectBalanceProvider(baseURL) != provider {
		return "", fmt.Errorf("account base_url does not match the selected balance provider")
	}
	switch provider {
	case UpstreamBalanceProviderDeepSeek:
		return "https://api.deepseek.com/user/balance", nil
	case UpstreamBalanceProviderStepFun:
		return "https://api.stepfun.com/v1/accounts", nil
	case UpstreamBalanceProviderSiliconFlow:
		return "https://api.siliconflow.cn/v1/user/info", nil
	case UpstreamBalanceProviderNewAPI:
		return resolveBalanceSameOriginURL(baseURL, "/api/user/self")
	case UpstreamBalanceProviderSub2API:
		return resolveBalanceSameOriginURL(baseURL, "/api/v1/user/profile")
	case UpstreamBalanceProviderCustom:
		base, err := url.Parse(strings.TrimSpace(baseURL))
		if err != nil || base.Scheme == "" || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
			return "", fmt.Errorf("account base_url must be an http or https URL")
		}
		endpoint := strings.TrimSpace(customEndpoint)
		if endpoint == "" || !strings.HasPrefix(endpoint, "/") || strings.HasPrefix(endpoint, "//") {
			return "", fmt.Errorf("custom endpoint must be an absolute path on the account base_url host")
		}
		rel, err := url.Parse(endpoint)
		if err != nil || rel.IsAbs() || rel.Host != "" {
			return "", fmt.Errorf("custom endpoint must not contain another host")
		}
		base.Path = ""
		base.RawPath = ""
		base.RawQuery = ""
		base.Fragment = ""
		return base.ResolveReference(rel).String(), nil
	default:
		return "", fmt.Errorf("balance provider is not supported")
	}
}

func resolveBalanceSameOriginURL(baseURL, endpoint string) (string, error) {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || base.Scheme == "" || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return "", fmt.Errorf("account base_url must be an http or https URL")
	}
	base.Path = ""
	base.RawPath = ""
	base.RawQuery = ""
	base.Fragment = ""
	rel, err := url.Parse(endpoint)
	if err != nil || !strings.HasPrefix(endpoint, "/") || rel.IsAbs() || rel.Host != "" {
		return "", fmt.Errorf("balance endpoint must be an absolute path on the account base_url host")
	}
	return base.ResolveReference(rel).String(), nil
}

func upstreamBalanceAuthHeaders(ctx context.Context, client *req.Client, account *Account, cfg UpstreamBalanceConfig, provider string) (map[string]string, error) {
	source := &UpstreamMonitorSource{
		Kind:         provider,
		BaseURL:      account.GetCredential("base_url"),
		AuthMode:     normalizeBalanceAuthMode(cfg.AuthMode),
		AuthUsername: strings.TrimSpace(cfg.AuthUsername),
		AuthToken:    strings.TrimSpace(account.GetCredential(upstreamBalanceAuthTokenCredentialKey)),
	}
	headers, err := upstreamMonitorRequestHeaders(ctx, client, source)
	if err != nil {
		return nil, fmt.Errorf("%s wallet authentication failed: %w", provider, err)
	}
	if len(headers) == 0 {
		return nil, fmt.Errorf("%s wallet authentication produced no credentials", provider)
	}
	return headers, nil
}

func fetchNewAPIWalletCNYBalance(ctx context.Context, client *req.Client, statusURL, profileURL string, headers map[string]string) (float64, error) {
	statusBody, err := requestUpstreamMonitorJSON(ctx, client, statusURL, nil)
	if err != nil {
		return 0, fmt.Errorf("New API status: %w", err)
	}
	profileBody, err := requestUpstreamMonitorJSON(ctx, client, profileURL, headers)
	if err != nil {
		return 0, fmt.Errorf("New API wallet profile: %w", err)
	}
	return parseNewAPIWalletCNYBalance(statusBody, profileBody)
}

func parseNewAPIWalletCNYBalance(statusBody, profileBody []byte) (float64, error) {
	var statusPayload struct {
		Success bool `json:"success"`
		Data    struct {
			QuotaDisplayType string  `json:"quota_display_type"`
			QuotaPerUnit     float64 `json:"quota_per_unit"`
			USDExchangeRate  float64 `json:"usd_exchange_rate"`
		} `json:"data"`
	}
	if err := json.Unmarshal(statusBody, &statusPayload); err != nil {
		return 0, fmt.Errorf("invalid New API status response JSON: %w", err)
	}
	if !statusPayload.Success {
		return 0, fmt.Errorf("New API status response is unsuccessful")
	}
	if !strings.EqualFold(strings.TrimSpace(statusPayload.Data.QuotaDisplayType), "CNY") {
		return 0, fmt.Errorf("New API site does not report balances in CNY")
	}
	if statusPayload.Data.QuotaPerUnit <= 0 || math.IsNaN(statusPayload.Data.QuotaPerUnit) || math.IsInf(statusPayload.Data.QuotaPerUnit, 0) {
		return 0, fmt.Errorf("New API status response has invalid quota_per_unit")
	}
	if statusPayload.Data.USDExchangeRate <= 0 || math.IsNaN(statusPayload.Data.USDExchangeRate) || math.IsInf(statusPayload.Data.USDExchangeRate, 0) {
		return 0, fmt.Errorf("New API status response has invalid usd_exchange_rate")
	}
	var profilePayload struct {
		Success bool `json:"success"`
		Data    struct {
			Quota float64 `json:"quota"`
		} `json:"data"`
	}
	if err := json.Unmarshal(profileBody, &profilePayload); err != nil {
		return 0, fmt.Errorf("invalid New API wallet profile JSON: %w", err)
	}
	if !profilePayload.Success {
		return 0, fmt.Errorf("New API wallet profile response is unsuccessful")
	}
	amount := profilePayload.Data.Quota / statusPayload.Data.QuotaPerUnit * statusPayload.Data.USDExchangeRate
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, fmt.Errorf("New API wallet balance is not finite")
	}
	return amount, nil
}

func fetchSub2APIWalletCNYBalance(ctx context.Context, client *req.Client, profileURL string, headers map[string]string) (float64, error) {
	body, err := requestUpstreamMonitorJSON(ctx, client, profileURL, headers)
	if err != nil {
		return 0, fmt.Errorf("Sub2API wallet profile: %w", err)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Balance float64 `json:"balance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, fmt.Errorf("invalid Sub2API wallet profile JSON: %w", err)
	}
	if payload.Code != 0 {
		return 0, fmt.Errorf("Sub2API wallet profile response is unsuccessful")
	}
	if math.IsNaN(payload.Data.Balance) || math.IsInf(payload.Data.Balance, 0) {
		return 0, fmt.Errorf("Sub2API wallet balance is not finite")
	}
	return payload.Data.Balance, nil
}

func parseUpstreamBalanceAmount(provider string, payload any, jsonPath string, divisor float64) (float64, error) {
	var raw any
	switch provider {
	case UpstreamBalanceProviderDeepSeek:
		root, ok := payload.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("unexpected DeepSeek balance response")
		}
		infos, ok := root["balance_infos"].([]any)
		if !ok {
			return 0, fmt.Errorf("DeepSeek response is missing balance_infos")
		}
		for _, item := range infos {
			info, ok := item.(map[string]any)
			if !ok || !strings.EqualFold(strings.TrimSpace(fmt.Sprint(info["currency"])), "CNY") {
				continue
			}
			raw = info["total_balance"]
			break
		}
		if raw == nil {
			return 0, fmt.Errorf("DeepSeek response has no CNY balance")
		}
	case UpstreamBalanceProviderStepFun:
		raw = lookupJSONPath(payload, "balance")
	case UpstreamBalanceProviderSiliconFlow:
		raw = lookupJSONPath(payload, "data.totalBalance")
	case UpstreamBalanceProviderCustom:
		raw = lookupJSONPath(payload, jsonPath)
	default:
		return 0, fmt.Errorf("balance provider is not supported")
	}
	amount, ok := numberFromAny(raw)
	if !ok || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, fmt.Errorf("balance field is missing or not numeric")
	}
	return amount / normalizeBalanceDivisor(divisor), nil
}

func lookupJSONPath(value any, path string) any {
	current := value
	for _, segment := range strings.Split(strings.TrimSpace(path), ".") {
		if segment == "" {
			return nil
		}
		switch node := current.(type) {
		case map[string]any:
			current = node[segment]
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(node) {
				return nil
			}
			current = node[index]
		default:
			return nil
		}
	}
	return current
}

func numberFromAny(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func balancePlatformName(account *Account, cfg UpstreamBalanceConfig) string {
	if name := strings.TrimSpace(cfg.PlatformName); name != "" {
		return name
	}
	provider := normalizeBalanceProvider(cfg.Provider)
	if provider == UpstreamBalanceProviderAuto && account != nil {
		provider = detectBalanceProvider(account.GetCredential("base_url"))
	}
	if provider == "" || provider == UpstreamBalanceProviderAuto {
		return "Unknown"
	}
	return provider
}

func balanceFundingIdentity(account *Account, cfg UpstreamBalanceConfig) string {
	if cfg.FundingKey != "" {
		return "explicit:" + strings.ToLower(cfg.FundingKey)
	}
	provider := normalizeBalanceProvider(cfg.Provider)
	if provider == UpstreamBalanceProviderAuto {
		provider = detectBalanceProvider(account.GetCredential("base_url"))
	}
	identity := provider
	if provider == UpstreamBalanceProviderCustom || provider == UpstreamBalanceProviderNewAPI || provider == UpstreamBalanceProviderSub2API {
		if parsed, err := url.Parse(strings.TrimSpace(account.GetCredential("base_url"))); err == nil {
			identity += "\x00" + strings.ToLower(parsed.Scheme+"://"+parsed.Host)
		}
	}
	credentialIdentity := account.GetCredential("api_key")
	if provider == UpstreamBalanceProviderNewAPI || provider == UpstreamBalanceProviderSub2API {
		credentialIdentity = normalizeBalanceAuthMode(cfg.AuthMode) + "\x00" + strings.ToLower(strings.TrimSpace(cfg.AuthUsername)) + "\x00" + account.GetCredential(upstreamBalanceAuthTokenCredentialKey)
	}
	h := sha256.Sum256([]byte(identity + "\x00" + credentialIdentity))
	return hex.EncodeToString(h[:])
}

func parseBalanceTime(value string) *time.Time {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	parsed = parsed.UTC()
	return &parsed
}

func truncateBalanceError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 240 {
		return value
	}
	return value[:240]
}

func sortBalanceOverview(overview *UpstreamBalanceOverview) {
	if overview == nil {
		return
	}
	for i := 0; i < len(overview.Platforms); i++ {
		for j := i + 1; j < len(overview.Platforms); j++ {
			if strings.ToLower(overview.Platforms[j].PlatformName) < strings.ToLower(overview.Platforms[i].PlatformName) {
				overview.Platforms[i], overview.Platforms[j] = overview.Platforms[j], overview.Platforms[i]
			}
		}
	}
	for i := 0; i < len(overview.Sources); i++ {
		for j := i + 1; j < len(overview.Sources); j++ {
			if overview.Sources[j].AccountID < overview.Sources[i].AccountID {
				overview.Sources[i], overview.Sources[j] = overview.Sources[j], overview.Sources[i]
			}
		}
	}
}
