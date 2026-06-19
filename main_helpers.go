package main

import (
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// We import net here so the alias in main.go compiles cleanly.
var _ net.Listener

// Config holds environment-driven runtime configuration.
type Config struct {
	I2PControlEndpoint string
	I2PControlPassword string
	SAMAddr            string
	TunnelName         string
	KeyDir             string
	RefreshInterval    time.Duration
}

func loadConfig() Config {
	return Config{
		I2PControlEndpoint: env("I2PCONTROL_ENDPOINT", "http://127.0.0.1:7650"),
		I2PControlPassword: env("I2PCONTROL_PASSWORD", "itoopie"),
		SAMAddr:            env("SAM_ADDR", "127.0.0.1:7656"),
		TunnelName:         env("TUNNEL_NAME", "i2pstats"),
		KeyDir:             env("KEY_DIR", "./keys"),
		RefreshInterval:    envDuration("REFRESH_INTERVAL", 30*time.Second),
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// cacheControl sets a Cache-Control header on the wrapped handler.
func cacheControl(h http.Handler, value string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", value)
		h.ServeHTTP(w, r)
	})
}

// gzipMiddleware compresses text responses for clients that accept it.
// Kept inline rather than pulling a dependency for ~30 lines of code.
var gzPool = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) { return g.gz.Write(b) }

func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		gz := gzPool.Get().(*gzip.Writer)
		defer gzPool.Put(gz)
		gz.Reset(w)
		defer gz.Close()
		h.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, gz: gz}, r)
	})
}
