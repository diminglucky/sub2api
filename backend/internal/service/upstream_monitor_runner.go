package service

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	upstreamMonitorRunnerTimeout        = 45 * time.Second
	upstreamMonitorRefreshLeaderLockKey = "upstream:monitor:refresh:leader"
	upstreamMonitorRefreshLeaderLockTTL = 2 * time.Minute
)

type UpstreamMonitorRunner struct {
	settingService *SettingService
	interval       time.Duration
	stopCh         chan struct{}
	stopOnce       sync.Once
	wg             sync.WaitGroup

	lockCache  LeaderLockCache
	db         *sql.DB
	instanceID string
}

func NewUpstreamMonitorRunner(settingService *SettingService, interval time.Duration) *UpstreamMonitorRunner {
	return &UpstreamMonitorRunner{
		settingService: settingService,
		interval:       interval,
		stopCh:         make(chan struct{}),
		instanceID:     uuid.NewString(),
	}
}

func (r *UpstreamMonitorRunner) SetLeaderLock(lockCache LeaderLockCache, db *sql.DB) {
	if r == nil {
		return
	}
	r.lockCache = lockCache
	r.db = db
}

func (r *UpstreamMonitorRunner) Start() {
	if r == nil || r.settingService == nil || r.interval <= 0 {
		return
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		r.runOnce()
		for {
			select {
			case <-ticker.C:
				r.runOnce()
			case <-r.stopCh:
				return
			}
		}
	}()
}

func (r *UpstreamMonitorRunner) Stop() {
	if r == nil {
		return
	}
	r.stopOnce.Do(func() {
		close(r.stopCh)
	})
	r.wg.Wait()
}

func (r *UpstreamMonitorRunner) runOnce() {
	lockCtx, lockCancel := context.WithTimeout(context.Background(), 2*time.Second)
	release, ok := tryAcquireSingletonLeaderLock(lockCtx, r.lockCache, r.db, upstreamMonitorRefreshLeaderLockKey, r.instanceID, upstreamMonitorRefreshLeaderLockTTL)
	lockCancel()
	if !ok {
		return
	}
	defer release()

	runCtx, cancel := context.WithTimeout(context.Background(), upstreamMonitorRunnerTimeout)
	defer cancel()

	result, err := r.settingService.RunDueUpstreamMonitorRefresh(runCtx)
	if err != nil {
		slog.Warn("[UpstreamMonitorRunner] refresh due sources failed", "error", err)
		return
	}
	if result == nil {
		return
	}
	if result.Summary.AttemptedCount == 0 && result.Summary.FailedCount == 0 {
		return
	}

	slog.Info(
		"[UpstreamMonitorRunner] refresh cycle completed",
		"attempted", result.Summary.AttemptedCount,
		"success", result.Summary.SuccessCount,
		"failed", result.Summary.FailedCount,
		"skipped", result.Summary.SkippedCount,
	)
}
