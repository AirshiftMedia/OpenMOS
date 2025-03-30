package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateID generates a random ID
func GenerateID(prefix string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	id := hex.EncodeToString(bytes)
	if prefix != "" {
		id = fmt.Sprintf("%s_%s", prefix, id)
	}

	return id, nil
}

// FormatDuration formats a duration in seconds to HH:MM:SS format
func FormatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second

	h := int(duration.Hours())
	m := int(duration.Minutes()) % 60
	s := int(duration.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// ParseDuration parses a duration in HH:MM:SS format to seconds
func ParseDuration(duration string) (int, error) {
	var h, m, s int

	n, err := fmt.Sscanf(duration, "%d:%d:%d", &h, &m, &s)
	if err != nil || n != 3 {
		// Try MM:SS format
		h = 0
		n, err = fmt.Sscanf(duration, "%d:%d", &m, &s)
		if err != nil || n != 2 {
			// Try seconds format
			m, s = 0, 0
			n, err = fmt.Sscanf(duration, "%d", &h)
			if err != nil || n != 1 {
				return 0, fmt.Errorf("invalid duration format: %s", duration)
			}
			return h, nil
		}
	}

	return h*3600 + m*60 + s, nil
}

// FormatTimestamp formats a time to ISO 8601 format
func FormatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTimestamp parses an ISO 8601 timestamp
func ParseTimestamp(timestamp string) (time.Time, error) {
	return time.Parse(time.RFC3339, timestamp)
}

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return len(s) == 0 || s == ""
}

// Truncate truncates a string to the specified length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

// Coalesce returns the first non-empty string
func Coalesce(values ...string) string {
	for _, v := range values {
		if !IsEmpty(v) {
			return v
		}
	}
	return ""
}
