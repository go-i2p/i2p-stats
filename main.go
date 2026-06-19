// i2pstats — JS-free I2P router stats dashboard exposed as a SAM hidden service.
//
// Third-party libraries (see THIRD_PARTY_LICENSES.md):
//   - github.com/go-i2p/onramp                  (MIT)
//   - github.com/go-i2p/go-i2pcontrol           (MIT)
//   - golang.org/x/sync/singleflight            (BSD-3-Clause)
package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-i2p/i2p-stats/pkg/cache"
	"github.com/go-i2p/i2p-stats/pkg/client"
	"github.com/go-i2p/i2p-stats/pkg/handlers"
	"github.com/go-i2p/i2p-stats/pkg/listener"
)

//go:embed templates static
var assets embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg := loadConfig()
	logger.Info("starting i2pstats",
		"sam", cfg.SAMAddr,
		"i2pcontrol", cfg.I2PControlEndpoint,
		"tunnel", cfg.TunnelName,
		"refresh", cfg.RefreshInterval,
	)

	// --- I2PControl client ---
	rc := client.New(client.Config{
		Endpoint: cfg.I2PControlEndpoint,
		Password: cfg.I2PControlPassword,
		Timeout:  3 * time.Second,
	})
	defer rc.Close()

	// --- Stats cache ---
	cacheCtx, cancelCache := context.WithCancel(context.Background())
	defer cancelCache()
	statsCache := cache.New(rc, cfg.RefreshInterval, logger)
	go statsCache.Run(cacheCtx)

	// --- Handlers ---
	h, err := handlers.New(statsCache, assets, logger, cfg.RefreshInterval)
	if err != nil {
		logger.Error("handler init", "err", err)
		os.Exit(1)
	}

	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		logger.Error("static FS", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Index)
	mux.HandleFunc("/healthz", h.Health)
	mux.Handle("/static/",
		http.StripPrefix("/static/",
			cacheControl(http.FileServer(http.FS(staticFS)), "public, max-age=86400"),
		))

	// --- Hidden service listener (no TCP bind ever) ---
	garlic, ln, err := listener.New(listener.Config{
		TunnelName: cfg.TunnelName,
		SAMAddr:    cfg.SAMAddr,
		KeyDir:     cfg.KeyDir,
	})
	if err != nil {
		logger.Error("hidden service init", "err", err)
		os.Exit(1)
	}
	logger.Info("hidden service ready", "b32", garlic.B32())

	// --- HTTP server ---
	var serverLn net.Listener = ln
	srv := &http.Server{
		Handler:           gzipMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// --- Signal-driven graceful shutdown ---
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		if err := srv.Serve(serverLn); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown signal received")
	case err := <-serveErr:
		if err != nil {
			logger.Error("http serve", "err", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("server shutdown", "err", err)
	}
	if err := garlic.Close(); err != nil {
		logger.Warn("garlic close", "err", err)
	}
	logger.Info("bye")
}
