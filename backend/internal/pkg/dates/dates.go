package dates

import "time"

// DateFormat is the standard date format used throughout the application.
const DateFormat = "2006-01-02"

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateFormat, s)
}
