package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeeklyStats_NewFields(t *testing.T) {
	s := WeeklyStats{
		ActiveSubItems: 10,
		NewlyCompleted: 2,
		InProgress:     3,
		Blocked:        1,
		Pending:        4,
		Pausing:        5,
		Overdue:        6,
	}
	assert.Equal(t, 4, s.Pending)
	assert.Equal(t, 5, s.Pausing)
	assert.Equal(t, 6, s.Overdue)
}

func TestWeeklyStats_JSONTags(t *testing.T) {
	s := WeeklyStats{Pending: 1, Pausing: 2, Overdue: 3}
	b, err := json.Marshal(s)
	require.NoError(t, err)
	var m map[string]int
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, 1, m["pending"])
	assert.Equal(t, 2, m["pausing"])
	assert.Equal(t, 3, m["overdue"])
}

func TestWeeklyStats_ZeroDefaults(t *testing.T) {
	s := WeeklyStats{}
	assert.Equal(t, 0, s.Pending)
	assert.Equal(t, 0, s.Pausing)
	assert.Equal(t, 0, s.Overdue)
}
