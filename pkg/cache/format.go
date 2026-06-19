package cache

import (
	"fmt"
	"time"
)

const placeholder = "—"

func formatBytesPerSec(v float64) string {
	const (
		kb = 1024.0
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case v >= gb:
		return fmt.Sprintf("%.2f GB/s", v/gb)
	case v >= mb:
		return fmt.Sprintf("%.2f MB/s", v/mb)
	case v >= kb:
		return fmt.Sprintf("%.2f KB/s", v/kb)
	default:
		return fmt.Sprintf("%.0f B/s", v)
	}
}

func formatInt(v float64) string {
	return fmt.Sprintf("%.0f", v)
}

func formatMillis(v float64) string {
	if v >= 1000 {
		return fmt.Sprintf("%.2f s", v/1000)
	}
	return fmt.Sprintf("%.0f ms", v)
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return placeholder
	}
	d = d.Round(time.Minute)
	if d < 0 {
		return placeholder
	}
	days := int(d / (24 * time.Hour))
	d -= time.Duration(days) * 24 * time.Hour
	hours := int(d / time.Hour)
	d -= time.Duration(hours) * time.Hour
	mins := int(d / time.Minute)
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
