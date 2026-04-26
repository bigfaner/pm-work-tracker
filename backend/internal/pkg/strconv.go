package pkg

import (
	"strconv"
)

// FormatID converts an int64 snowflake ID to its string representation for JSON serialization.
func FormatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

// FormatIDPtr converts a nullable int64 snowflake ID to *string.
// Returns nil when input is nil.
func FormatIDPtr(id *int64) *string {
	if id == nil {
		return nil
	}
	s := strconv.FormatInt(*id, 10)
	return &s
}

// ParseID parses a string snowflake ID back to int64.
func ParseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ParseIDPtr parses a nullable string snowflake ID back to *int64.
// Returns nil when input is nil or empty.
func ParseIDPtr(s *string) (*int64, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(*s, 10, 64)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
