package db

import "time"

// TimeLayout is the canonical ISO-8601 UTC layout (millisecond precision) used
// for all TEXT timestamp columns, matching the SQLite `strftime` default.
const TimeLayout = "2006-01-02T15:04:05.000Z"

// FormatTime renders t as a canonical UTC timestamp string.
func FormatTime(t time.Time) string { return t.UTC().Format(TimeLayout) }

// ParseTime parses a stored timestamp, tolerating a few common layouts. Returns
// the zero time if none match.
func ParseTime(s string) time.Time {
	for _, layout := range []string{TimeLayout, time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
