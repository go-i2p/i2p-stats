// Package listener constructs a SAMv3-managed I2P hidden service listener.
package listener

import (
	"fmt"
	"net"

	// onramp — thin wrapper over sam3 that exposes a net.Listener.
	"github.com/go-i2p/onramp"
)

// Garlic is the subset of *onramp.Garlic we depend on, enabling fakes.
type Garlic interface {
	Listen() (net.Listener, error)
	Close() error
	B32() string
}

// Config configures the hidden service.
type Config struct {
	TunnelName string
	SAMAddr    string
	KeyDir     string
	Options    []string // sam3-style tunnel options
}

type garlicWrapper struct {
	*onramp.Garlic
}

func (w *garlicWrapper) Listen() (net.Listener, error) {
	return w.Garlic.ListenStream()
}

func (w *garlicWrapper) B32() string {
	keys, err := w.Garlic.Keys()
	if err != nil || keys == nil {
		return ""
	}
	return keys.Addr().Base32()
}

// New creates and starts a SAMv3 hidden service. The returned Garlic
// must be Closed during shutdown. Listen() yields a net.Listener for http.Serve.
func New(cfg Config) (Garlic, net.Listener, error) {
	if cfg.Options == nil {
		// Sensible defaults from plan §9.4.
		cfg.Options = []string{
			"inbound.length=3",
			"outbound.length=3",
			"inbound.quantity=2",
			"outbound.quantity=2",
			"inbound.backupQuantity=1",
			"outbound.backupQuantity=1",
		}
	}

	// Persist keys in custom key directory
	if cfg.KeyDir != "" {
		onramp.I2P_KEYSTORE_PATH = cfg.KeyDir
	}

	g, err := onramp.NewGarlic(cfg.TunnelName, cfg.SAMAddr, cfg.Options)
	if err != nil {
		return nil, nil, fmt.Errorf("onramp.NewGarlic: %w", err)
	}

	w := &garlicWrapper{Garlic: g}
	ln, err := w.Listen()
	if err != nil {
		_ = g.Close()
		return nil, nil, fmt.Errorf("garlic.Listen: %w", err)
	}

	return w, ln, nil
}
