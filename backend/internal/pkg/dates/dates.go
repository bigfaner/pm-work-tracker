package dates

import "time"

// DateFormat is the standard date format used throughout the application.
const DateFormat = "2006-01-02"

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateFormat, s)
}

// FormatTimePtr formats a *time.Time to a *string using DateFormat.
// Returns nil if the input is nil.
func FormatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(DateFormat)
	return &s
}
