package dates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDate_ValidDate(t *testing.T) {
	result, err := ParseDate("2024-01-15")
	require.NoError(t, err)
	expected := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	assert.True(t, result.Equal(expected))
}

func TestParseDate_InvalidFormat(t *testing.T) {
	result, err := ParseDate("invalid")
	assert.Error(t, err)
	assert.True(t, result.IsZero())
}

func TestParseDate_EmptyString(t *testing.T) {
	result, err := ParseDate("")
	assert.Error(t, err)
	assert.True(t, result.IsZero())
}

func TestParseDate_WrongFormat(t *testing.T) {
	result, err := ParseDate("15-01-2024")
	assert.Error(t, err)
	assert.True(t, result.IsZero())
}

func TestDateFormat(t *testing.T) {
	assert.Equal(t, "2006-01-02", DateFormat)
}

func TestFormatTimePtr_Nil(t *testing.T) {
	result := FormatTimePtr(nil)
	assert.Nil(t, result)
}

func TestFormatTimePtr_Valid(t *testing.T) {
	input := time.Date(2026, 4, 15, 10, 30, 0, 0, time.UTC)
	result := FormatTimePtr(&input)
	require.NotNil(t, result)
	assert.Equal(t, "2026-04-15", *result)
}

func TestFormatTimePtr_ZeroTime(t *testing.T) {
	input := time.Time{}
	result := FormatTimePtr(&input)
	require.NotNil(t, result)
	assert.Equal(t, "0001-01-01", *result)
}
