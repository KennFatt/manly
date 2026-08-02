package analytics

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ParseSince converts a user-facing duration into an absolute UTC cutoff.
// It accepts time.ParseDuration values and day values such as 7d.
func ParseSince(value string, now time.Time) (*time.Time, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return nil, nil
	}

	var duration time.Duration
	var err error
	if strings.HasSuffix(value, "d") {
		days, parseErr := strconv.ParseFloat(strings.TrimSuffix(value, "d"), 64)
		if parseErr != nil || days < 0 || math.IsNaN(days) || math.IsInf(days, 0) || days > float64(math.MaxInt64)/float64(24*time.Hour) {
			if parseErr == nil && days < 0 {
				err = fmt.Errorf("duration must not be negative")
			} else if parseErr == nil {
				err = fmt.Errorf("duration is too large")
			} else {
				err = parseErr
			}
		} else {
			duration = time.Duration(days * float64(24*time.Hour))
		}
	} else {
		duration, err = time.ParseDuration(value)
		if err == nil && duration < 0 {
			err = fmt.Errorf("duration must not be negative")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("invalid --since duration %q (want values such as 24h or 7d): %w", value, err)
	}
	cutoff := now.UTC().Add(-duration)
	return &cutoff, nil
}
