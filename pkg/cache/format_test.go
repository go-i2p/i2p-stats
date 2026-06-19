package cache

import (
	"testing"
	"time"
)

func TestFormatBytesPerSec(t *testing.T) {
	tests := []struct {
		val  float64
		want string
	}{
		{500, "500 B/s"},
		{1024, "1.00 KB/s"},
		{1536, "1.50 KB/s"},
		{1024 * 1024, "1.00 MB/s"},
		{1024 * 1024 * 1024, "1.00 GB/s"},
	}

	for _, tt := range tests {
		got := formatBytesPerSec(tt.val)
		if got != tt.want {
			t.Errorf("formatBytesPerSec(%f) = %q; want %q", tt.val, got, tt.want)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		val  float64
		want string
	}{
		{0, "0"},
		{42, "42"},
		{100.4, "100"},
		{99.9, "100"}, // Rounds
	}

	for _, tt := range tests {
		got := formatInt(tt.val)
		if got != tt.want {
			t.Errorf("formatInt(%f) = %q; want %q", tt.val, got, tt.want)
		}
	}
}

func TestFormatMillis(t *testing.T) {
	tests := []struct {
		val  float64
		want string
	}{
		{50, "50 ms"},
		{999, "999 ms"},
		{1000, "1.00 s"},
		{1500, "1.50 s"},
	}

	for _, tt := range tests {
		got := formatMillis(tt.val)
		if got != tt.want {
			t.Errorf("formatMillis(%f) = %q; want %q", tt.val, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		val  time.Duration
		want string
	}{
		{0, "—"},
		{-10 * time.Second, "—"},
		{10 * time.Second, "0m"}, // rounds down but stays positive
		{45 * time.Second, "1m"}, // rounded to nearest minute
		{2 * time.Minute, "2m"},
		{90 * time.Minute, "1h 30m"},
		{25 * time.Hour, "1d 1h 0m"},
		{25*time.Hour + 30*time.Minute, "1d 1h 30m"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.val)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q; want %q", tt.val, got, tt.want)
		}
	}
}
