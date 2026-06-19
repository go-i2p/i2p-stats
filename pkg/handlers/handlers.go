// Package handlers contains read-only HTTP handlers backed by the stats cache.
package handlers

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-i2p/i2p-stats/pkg/cache"
)

// Handlers wires templates and the cache into http.Handlers.
type Handlers struct {
	cache  *cache.Cache
	tmpl   *template.Template
	logger *slog.Logger
	ttl    time.Duration
}

// New parses templates from the provided FS and returns a Handlers instance.
// Templates are parsed once at startup per plan §6.
func New(c *cache.Cache, tmplFS fs.FS, logger *slog.Logger, ttl time.Duration) (*Handlers, error) {
	t, err := template.ParseFS(tmplFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Handlers{cache: c, tmpl: t, logger: logger, ttl: ttl}, nil
}

// Index renders the dashboard.
func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	stats := h.cache.Snapshot()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")

	data := struct {
		Stats          any
		RefreshSeconds int
		HasData        bool
	}{
		Stats:          stats,
		RefreshSeconds: int(h.ttl.Seconds()),
		HasData:        !stats.LastUpdated.IsZero(),
	}
	if err := h.tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		h.logger.Error("template render", "err", err)
	}
}

// Health returns 200 when cache freshness is within 5x TTL, else 503.
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	if h.cache.IsHealthy(5 * h.ttl) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("stale"))
}
