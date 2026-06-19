// Package model holds display-ready, pre-formatted stats for templates.
package model

import "time"

// Stats is the single struct rendered by templates. All values are
// pre-formatted strings so handlers do zero formatting work per request.
type Stats struct {
	// Router identity / status
	Version       string
	Uptime        string
	NetworkStatus string
	Status        string

	// Bandwidth (formatted, e.g. "12.4 KB/s")
	SendRate string
	RecvRate string

	// Peers
	ActivePeers       string
	NetDBActivePeers  string
	FastPeers         string
	HighCapacityPeers string

	// Tunnels
	ParticipatingTunnels string
	BuildSuccess         string
	BuildFailure         string

	// Latency / health proxies
	SendAckTime             string
	JobLag                  string
	TransportSendProcessing string

	// Meta
	LastUpdated time.Time
	Unhealthy   bool
	StaleReason string
}

// Age returns how long ago the stats were refreshed.
func (s *Stats) Age() time.Duration {
	if s.LastUpdated.IsZero() {
		return 0
	}
	return time.Since(s.LastUpdated)
}
