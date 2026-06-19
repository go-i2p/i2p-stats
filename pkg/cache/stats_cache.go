// Package cache holds a TTL'd, mutex-protected snapshot of router stats and
// uses singleflight to coalesce concurrent refresh requests.
package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/singleflight" // BSD-3-Clause

	"github.com/go-i2p/i2p-stats/pkg/client"
	"github.com/go-i2p/i2p-stats/pkg/model"
)

// Stat keys requested per refresh — see plan §4.1.
var statKeys = []string{
	"bw.sendRate",
	"bw.recvRate",
	"router.activePeers",
	"router.netDbActivePeers",
	"router.fastPeers",
	"router.highCapacityPeers",
	"tunnel.participatingTunnels",
	"tunnel.buildSuccess",
	"tunnel.buildFailure",
	"client.sendAckTime",
	"jobQueue.jobLag",
	"transport.sendProcessingTime",
}

// Cache holds the latest pre-formatted stats.
type Cache struct {
	rc     client.RouterClient
	ttl    time.Duration
	logger *slog.Logger

	mu       sync.RWMutex
	current  model.Stats
	hasData  bool

	sf singleflight.Group
}

// New constructs a Cache. ttl is the refresh interval.
func New(rc client.RouterClient, ttl time.Duration, logger *slog.Logger) *Cache {
	return &Cache{
		rc:     rc,
		ttl:    ttl,
		logger: logger,
	}
}

// Snapshot returns a copy of the latest stats; safe for templates.
func (c *Cache) Snapshot() model.Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current
}

// Run refreshes immediately, then on every TTL tick, until ctx is cancelled.
func (c *Cache) Run(ctx context.Context) {
	// First refresh is best-effort; missing data renders as placeholders.
	if _, err := c.refresh(ctx); err != nil {
		c.logger.Warn("initial refresh failed", "err", err)
	}
	t := time.NewTicker(c.ttl)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if _, err := c.refresh(ctx); err != nil {
				c.logger.Warn("refresh failed", "err", err)
			}
		}
	}
}

// Refresh forces a refresh, coalescing concurrent calls via singleflight.
func (c *Cache) Refresh(ctx context.Context) error {
	_, err := c.refresh(ctx)
	return err
}

func (c *Cache) refresh(ctx context.Context) (model.Stats, error) {
	v, err, _ := c.sf.Do("stats", func() (interface{}, error) {
		return c.doRefresh(ctx)
	})
	if err != nil {
		return model.Stats{}, err
	}
	return v.(model.Stats), nil
}

func (c *Cache) doRefresh(ctx context.Context) (model.Stats, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Start from the previous snapshot so partial failures keep last-good values.
	c.mu.RLock()
	next := c.current
	had := c.hasData
	c.mu.RUnlock()

	var firstErr error

	info, err := c.rc.RouterInfo(ctx)
	if err != nil {
		firstErr = fmt.Errorf("router info: %w", err)
	} else {
		next.Version = info.Version
		next.Uptime = formatDuration(info.Uptime)
		next.NetworkStatus = info.NetworkStatus
		next.Status = info.Status
	}

	rates, err := c.rc.RateStats(ctx, statKeys)
	if err != nil && firstErr == nil {
		firstErr = err
	}

	if v, ok := rates["bw.sendRate"]; ok {
		next.SendRate = formatBytesPerSec(v)
	}
	if v, ok := rates["bw.recvRate"]; ok {
		next.RecvRate = formatBytesPerSec(v)
	}
	if v, ok := rates["router.activePeers"]; ok {
		next.ActivePeers = formatInt(v)
	}
	if v, ok := rates["router.netDbActivePeers"]; ok {
		next.NetDBActivePeers = formatInt(v)
	}
	if v, ok := rates["router.fastPeers"]; ok {
		next.FastPeers = formatInt(v)
	}
	if v, ok := rates["router.highCapacityPeers"]; ok {
		next.HighCapacityPeers = formatInt(v)
	}
	if v, ok := rates["tunnel.participatingTunnels"]; ok {
		next.ParticipatingTunnels = formatInt(v)
	}
	if v, ok := rates["tunnel.buildSuccess"]; ok {
		next.BuildSuccess = formatInt(v)
	}
	if v, ok := rates["tunnel.buildFailure"]; ok {
		next.BuildFailure = formatInt(v)
	}
	if v, ok := rates["client.sendAckTime"]; ok {
		next.SendAckTime = formatMillis(v)
	}
	if v, ok := rates["jobQueue.jobLag"]; ok {
		next.JobLag = formatMillis(v)
	}
	if v, ok := rates["transport.sendProcessingTime"]; ok {
		next.TransportSendProcessing = formatMillis(v)
	}

	if firstErr != nil {
		next.Unhealthy = true
		next.StaleReason = firstErr.Error()
		// Only update LastUpdated if we got *something* fresh.
		if !had {
			next.LastUpdated = time.Time{}
		}
	} else {
		next.Unhealthy = false
		next.StaleReason = ""
		next.LastUpdated = time.Now()
	}

	c.mu.Lock()
	c.current = next
	c.hasData = c.hasData || firstErr == nil
	c.mu.Unlock()

	return next, firstErr
}

// IsHealthy reports whether the cache holds data fresher than maxAge.
func (c *Cache) IsHealthy(maxAge time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.hasData || c.current.LastUpdated.IsZero() {
		return false
	}
	return time.Since(c.current.LastUpdated) <= maxAge
}
