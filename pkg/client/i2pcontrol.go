// Package client wraps go-i2pcontrol behind a narrow RouterClient interface
// so handlers and tests are decoupled from the concrete RPC library.
package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	// go-i2pcontrol — MIT; JSON-RPC client for I2PControl.
	i2pctl "github.com/go-i2p/go-i2pcontrol"
)

// RouterClient is the minimal surface our cache depends on.
type RouterClient interface {
	RouterInfo(ctx context.Context) (RouterInfo, error)
	RateStats(ctx context.Context, keys []string) (map[string]float64, error)
	Close() error
}

// RouterInfo is the subset of router metadata we render.
type RouterInfo struct {
	Version       string
	Uptime        time.Duration
	NetworkStatus string
	Status        string
}

// Config configures a new I2PControl client.
type Config struct {
	Endpoint string // e.g. "http://127.0.0.1:7650"
	Password string
	Timeout  time.Duration
}

// Global package mutex to serialize access to the go-i2pcontrol package-level state.
var i2pcontrolMu sync.Mutex

// i2pControlClient implements RouterClient.
type i2pControlClient struct {
	cfg Config

	tokenAge  time.Time
	authValid bool
}

// New returns a RouterClient. Authentication is lazy on first call so that
// startup does not fail if the router is briefly unreachable.
func New(cfg Config) RouterClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 3 * time.Second
	}
	return &i2pControlClient{cfg: cfg}
}

// ensureAuth (re)authenticates if the session is missing or expired.
// Caller must hold i2pcontrolMu.
func (c *i2pControlClient) ensureAuth(ctx context.Context) error {
	// I2PControl tokens expire after ~7 minutes; reauth proactively at 5.
	if c.authValid && time.Since(c.tokenAge) < 5*time.Minute {
		return nil
	}

	u, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return fmt.Errorf("parse endpoint: %w", err)
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	path := strings.Trim(u.Path, "/")
	if path == "" {
		path = "jsonrpc"
	}

	// Initialize the package-level RPC client
	i2pctl.Initialize(host, port, path)

	// Authenticate with I2PControl
	_, err = i2pctl.Authenticate(c.cfg.Password)
	if err != nil {
		c.authValid = false
		return fmt.Errorf("i2pcontrol authenticate: %w", err)
	}

	c.authValid = true
	c.tokenAge = time.Now()
	return nil
}

// RouterInfo returns merged router status fields.
func (c *i2pControlClient) RouterInfo(ctx context.Context) (RouterInfo, error) {
	i2pcontrolMu.Lock()
	defer i2pcontrolMu.Unlock()

	if err := ctx.Err(); err != nil {
		return RouterInfo{}, err
	}

	if err := c.ensureAuth(ctx); err != nil {
		return RouterInfo{}, err
	}

	var info RouterInfo
	var err error

	info.Version, err = i2pctl.Version()
	if err != nil {
		c.authValid = false
		return RouterInfo{}, fmt.Errorf("router version: %w", err)
	}

	uptimeMs, err := i2pctl.UpTime()
	if err != nil {
		c.authValid = false
		return RouterInfo{}, fmt.Errorf("router uptime: %w", err)
	}
	info.Uptime = time.Duration(uptimeMs) * time.Millisecond

	info.NetworkStatus, err = i2pctl.NetStatus()
	if err != nil {
		c.authValid = false
		return RouterInfo{}, fmt.Errorf("router netstatus: %w", err)
	}

	info.Status, err = i2pctl.Status()
	if err != nil {
		c.authValid = false
		return RouterInfo{}, fmt.Errorf("router status: %w", err)
	}

	return info, nil
}

// RateStats issues sequential requests for the listed stat keys.
func (c *i2pControlClient) RateStats(ctx context.Context, keys []string) (map[string]float64, error) {
	i2pcontrolMu.Lock()
	defer i2pcontrolMu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := c.ensureAuth(ctx); err != nil {
		return nil, err
	}

	results := make(map[string]float64, len(keys))
	for _, k := range keys {
		if err := ctx.Err(); err != nil {
			return results, err
		}
		// Period of 60000ms (60 seconds)
		v, err := i2pctl.RateStatExact(k, 60000)
		if err != nil {
			c.authValid = false
			return results, fmt.Errorf("ratestat %q: %w", k, err)
		}
		results[k] = v
	}
	return results, nil
}

// Close releases any underlying resources.
func (c *i2pControlClient) Close() error {
	i2pcontrolMu.Lock()
	defer i2pcontrolMu.Unlock()
	c.authValid = false
	return nil
}
