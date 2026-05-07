// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package http

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DataDog/datadog-agent/pkg/logs/client"
)

var defaultMinWorkers = 4
var defaultMaxWorkers = defaultMinWorkers * 10

func defaultPool() *workerPool {
	return newWorkerPool(0, ewmaAlpha, defaultMinWorkers, defaultMaxWorkers, targetLatency, client.NewNoopDestinationMetadata())
}

// driveUntil submits constant-latency results to the pool until predicate
// returns true, or the timeout elapses.
//
// pool.run() updates the EWMA samples via a goroutine spawned per call. Under
// CI scheduler pressure those goroutines can bunch up, so a fixed iteration
// count produces a non-deterministic number of EWMA steps and makes "submit N
// tasks; assert exact worker count" flaky. Driving with a predicate decouples
// the test from the scheduler: we keep feeding work until the algorithm has
// actually converged to the expected state. Returns true if the predicate
// became true within the timeout.
func driveUntil(pool *workerPool, latency time.Duration, predicate func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if predicate() {
			return true
		}
		pool.run(func() destinationResult {
			return destinationResult{latency: latency}
		})
	}
	return predicate()
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func TestRetryableError(t *testing.T) {
	pool := defaultPool()
	require.True(t,
		driveUntil(pool, targetLatency*2, func() bool {
			return pool.inUseWorkers == defaultMinWorkers*2
		}, 2*time.Second),
		"pool failed to grow to %d workers; got %d", defaultMinWorkers*2, pool.inUseWorkers)

	pool.run(func() destinationResult {
		return destinationResult{latency: targetLatency * 2, err: client.NewRetryableError(errors.New(""))}
	})

	start := time.Now()

	pool.Lock()
	backoff := pool.shouldBackoff
	pool.Unlock()
	for !backoff && time.Since(start) < 500*time.Millisecond {
		time.Sleep(time.Millisecond)
		pool.Lock()
		backoff = pool.shouldBackoff
		pool.Unlock()
	}

	assert.Equal(t, true, pool.shouldBackoff)

	// The next pool run will detect the backoff flag to
	// 1. Reduce the worker load to the minimum
	// 2. Reset the tracked latency to force a gradual increase.
	pool.run(func() destinationResult {
		return destinationResult{latency: targetLatency * 2}
	})

	assert.Equal(t, defaultMinWorkers, pool.inUseWorkers)
	assert.Equal(t, time.Duration(0), pool.virtualLatency)

	// Confirm that we do in fact recover over time.
	require.True(t,
		driveUntil(pool, targetLatency*2, func() bool {
			return pool.inUseWorkers == defaultMinWorkers*2
		}, 2*time.Second),
		"pool failed to recover to %d workers after backoff; got %d", defaultMinWorkers*2, pool.inUseWorkers)
}

func TestNonRetryableError(t *testing.T) {
	pool := defaultPool()
	require.True(t,
		driveUntil(pool, targetLatency*2, func() bool {
			return pool.inUseWorkers == defaultMinWorkers*2
		}, 2*time.Second),
		"pool failed to grow to %d workers; got %d", defaultMinWorkers*2, pool.inUseWorkers)

	pool.run(func() destinationResult {
		return destinationResult{latency: targetLatency * 2, err: errors.New("")}
	})
	pool.resizeUnsafe()
	pool.resizeUnsafe()
	assert.Equal(t, defaultMinWorkers*2, pool.inUseWorkers)
}

func TestWorkerCounts(t *testing.T) {
	scenarios := []struct {
		name                string
		latency             time.Duration
		expectedWorkerCount int
	}{
		{
			name:                "Mininum Workers chosen if latency below target",
			latency:             0,
			expectedWorkerCount: defaultMinWorkers,
		},
		{
			name:                "Reasonable number of workers added at higher than target latency",
			latency:             targetLatency * 2,
			expectedWorkerCount: defaultMinWorkers * 2,
		},
		{
			name:                "Maximum number of workers not exceeded",
			latency:             targetLatency * 20,
			expectedWorkerCount: defaultMaxWorkers,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			pool := defaultPool()

			// Drive convergence on three things at once: worker count,
			// pool channel length (= goroutines have all returned their
			// worker), and EWMA virtual latency. The 10ms tolerance on
			// virtualLatency is loose enough to absorb CI scheduler
			// jitter but still tight enough to catch real regressions
			// (~0.3% of the largest scenario value).
			converged := driveUntil(pool, s.latency, func() bool {
				return pool.inUseWorkers == s.expectedWorkerCount &&
					len(pool.pool) == s.expectedWorkerCount &&
					absDuration(pool.virtualLatency-s.latency) <= 10*time.Millisecond
			}, 3*time.Second)
			require.Truef(t, converged,
				"pool failed to converge: workers=%d (want %d), len(pool)=%d, virtualLatency=%s (want %s)",
				pool.inUseWorkers, s.expectedWorkerCount, len(pool.pool), pool.virtualLatency, s.latency)
		})
	}
}
