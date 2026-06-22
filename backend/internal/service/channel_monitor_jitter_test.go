//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateJitter(t *testing.T) {
	require.NoError(t, validateJitter(0, 60))
	require.NoError(t, validateJitter(45, 60))

	require.ErrorIs(t, validateJitter(-1, 60), ErrChannelMonitorInvalidJitter)
	require.ErrorIs(t, validateJitter(46, 60), ErrChannelMonitorInvalidJitter)
}

func TestScheduledMonitorNextDelayWithJitter(t *testing.T) {
	task := &scheduledMonitor{
		interval: 60 * time.Second,
		jitter:   10 * time.Second,
	}

	for i := 0; i < 100; i++ {
		delay := task.nextDelay()
		require.GreaterOrEqual(t, delay, 50*time.Second)
		require.LessOrEqual(t, delay, 70*time.Second)
	}
}

func TestScheduledMonitorNextDelayClampsFloor(t *testing.T) {
	task := &scheduledMonitor{
		interval: 16 * time.Second,
		jitter:   10 * time.Second,
	}

	for i := 0; i < 100; i++ {
		delay := task.nextDelay()
		require.GreaterOrEqual(t, delay, monitorMinIntervalSeconds*time.Second)
	}
}
