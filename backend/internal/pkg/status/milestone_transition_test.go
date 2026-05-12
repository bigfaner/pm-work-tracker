package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMilestoneMapTransitions_AllPaths(t *testing.T) {
	expected := map[string][]string{
		"planning":  {"reviewed"},
		"reviewed":  {"ready", "planning"},
		"ready":     {"executing", "reviewed"},
		"executing": {"completed", "ready"},
	}
	for from, expectedTargets := range expected {
		t.Run(from, func(t *testing.T) {
			assert.Equal(t, expectedTargets, MilestoneMapTransitions[from])
		})
	}
}

func TestMilestoneTransitions_AllPaths(t *testing.T) {
	expected := map[string][]string{
		"not_started": {"in_progress", "cancelled"},
		"in_progress": {"completed", "cancelled"},
		"completed":   {"cancelled"},
	}
	for from, expectedTargets := range expected {
		t.Run(from, func(t *testing.T) {
			assert.Equal(t, expectedTargets, MilestoneTransitions[from])
		})
	}
}

func TestMilestoneMapTransitions_Count(t *testing.T) {
	// planning:1 + reviewed:2 + ready:2 + executing:2 = 7
	total := 0
	for _, targets := range MilestoneMapTransitions {
		total += len(targets)
	}
	assert.Equal(t, 7, total)
}

func TestMilestoneTransitions_Count(t *testing.T) {
	// not_started:2 + in_progress:2 + completed:1 = 5
	total := 0
	for _, targets := range MilestoneTransitions {
		total += len(targets)
	}
	assert.Equal(t, 5, total)
}

func TestIsValidTransition_MilestoneMap(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		// Legal paths
		{"planning->reviewed", "planning", "reviewed", true},
		{"reviewed->ready", "reviewed", "ready", true},
		{"reviewed->planning", "reviewed", "planning", true},
		{"ready->executing", "ready", "executing", true},
		{"ready->reviewed", "ready", "reviewed", true},
		{"executing->completed", "executing", "completed", true},
		{"executing->ready", "executing", "ready", true},
		// Illegal paths
		{"planning->executing", "planning", "executing", false},
		{"completed->planning", "completed", "planning", false},
		{"planning->planning", "planning", "planning", false},
		// Unknown status
		{"unknown->planning", "unknown", "planning", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTransition(MilestoneMapTransitions, tt.from, tt.to)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidTransition_Milestone(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		// Legal paths
		{"not_started->in_progress", "not_started", "in_progress", true},
		{"not_started->cancelled", "not_started", "cancelled", true},
		{"in_progress->completed", "in_progress", "completed", true},
		{"in_progress->cancelled", "in_progress", "cancelled", true},
		{"completed->cancelled", "completed", "cancelled", true},
		// Illegal paths
		{"not_started->completed", "not_started", "completed", false},
		{"cancelled->not_started", "cancelled", "not_started", false},
		{"completed->in_progress", "completed", "in_progress", false},
		// Self transitions
		{"not_started->not_started", "not_started", "not_started", false},
		// Unknown status
		{"unknown->not_started", "unknown", "not_started", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTransition(MilestoneTransitions, tt.from, tt.to)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAvailableTransitions_MilestoneMap(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    []string
	}{
		{"planning", "planning", []string{"reviewed"}},
		{"reviewed", "reviewed", []string{"ready", "planning"}},
		{"ready", "ready", []string{"executing", "reviewed"}},
		{"executing", "executing", []string{"completed", "ready"}},
		{"completed", "completed", []string{}},
		{"unknown", "unknown", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAvailableTransitions(MilestoneMapTransitions, tt.current)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAvailableTransitions_Milestone(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    []string
	}{
		{"not_started", "not_started", []string{"in_progress", "cancelled"}},
		{"in_progress", "in_progress", []string{"completed", "cancelled"}},
		{"completed", "completed", []string{"cancelled"}},
		{"cancelled", "cancelled", []string{}},
		{"unknown", "unknown", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAvailableTransitions(MilestoneTransitions, tt.current)
			assert.Equal(t, tt.want, got)
		})
	}
}
