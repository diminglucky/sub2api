package service

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type LotteryDrawRunner struct {
	lotterySvc *LotteryService
	interval   time.Duration
	stopCh     chan struct{}
	stopOnce   sync.Once
	wg         sync.WaitGroup
}

func NewLotteryDrawRunner(lotterySvc *LotteryService, interval time.Duration) *LotteryDrawRunner {
	return &LotteryDrawRunner{
		lotterySvc: lotterySvc,
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
}

func (r *LotteryDrawRunner) Start() {
	if r == nil || r.lotterySvc == nil || r.interval <= 0 {
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

func (r *LotteryDrawRunner) Stop() {
	if r == nil {
		return
	}
	r.stopOnce.Do(func() {
		close(r.stopCh)
	})
	r.wg.Wait()
}

func (r *LotteryDrawRunner) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	count, err := r.lotterySvc.DrawDue(ctx, 20)
	if err != nil {
		slog.Warn("[LotteryDrawRunner] draw due failed", "error", err)
		return
	}
	if count > 0 {
		slog.Info("[LotteryDrawRunner] drew due lotteries", "count", count)
	}
}
