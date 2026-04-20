package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainItemTransitions_Count(t *testing.T) {
	// 11 legal transition paths per PRD transition table
	// (AC-2 states 10 but the explicit transition table lists 11:
	// pending:2 + progressing:4 + blocking:1 + pausing:2 + reviewing:2 = 11)
	total := 0
	for _, targets := range MainItemTransitions {
		total += len(targets)
	}
	assert.Equal(t, 11, total, "MainItemTransitions should define exactly 11 paths")
}

func TestSubItemTransitions_Count(t *testing.T) {
	// 9 legal transition paths per PRD AC-3
	total := 0
	for _, targets := range SubItemTransitions {
		total += len(targets)
	}
	assert.Equal(t, 9, total, "SubItemTransitions should define exactly 9 paths")
}

func TestMainItemTransitions_AllPaths(t *testing.T) {
	expected := map[string][]string{
		"pending":     {"progressing", "closed"},
		"progressing": {"blocking", "pausing", "reviewing", "closed"},
		"blocking":    {"progressing"},
		"pausing":     {"progressing", "closed"},
		"reviewing":   {"completed", "progressing"},
	}
	for from, expectedTargets := range expected {
		t.Run(from, func(t *testing.T) {
			assert.Equal(t, expectedTargets, MainItemTransitions[from])
		})
	}
}

func TestSubItemTransitions_AllPaths(t *testing.T) {
	expected := map[string][]string{
		"pending":     {"progressing", "closed"},
		"progressing": {"blocking", "pausing", "completed", "closed"},
		"blocking":    {"progressing"},
		"pausing":     {"progressing", "closed"},
	}
	for from, expectedTargets := range expected {
		t.Run(from, func(t *testing.T) {
			assert.Equal(t, expectedTargets, SubItemTransitions[from])
		})
	}
}

func TestIsValidTransition_MainItem(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		// Legal paths
		{"pending->progressing", "pending", "progressing", true},
		{"pending->closed", "pending", "closed", true},
		{"progressing->blocking", "progressing", "blocking", true},
		{"progressing->pausing", "progressing", "pausing", true},
		{"progressing->reviewing", "progressing", "reviewing", true},
		{"progressing->closed", "progressing", "closed", true},
		{"blocking->progressing", "blocking", "progressing", true},
		{"pausing->progressing", "pausing", "progressing", true},
		{"pausing->closed", "pausing", "closed", true},
		{"reviewing->completed", "reviewing", "completed", true},
		{"reviewing->progressing", "reviewing", "progressing", true},
		// Illegal paths
		{"pending->completed", "pending", "completed", false},
		{"pending->reviewing", "pending", "reviewing", false},
		{"completed->pending", "completed", "pending", false},
		{"closed->pending", "closed", "pending", false},
		{"blocking->closed", "blocking", "closed", false},
		// Self transitions
		{"pending->pending", "pending", "pending", false},
		{"progressing->progressing", "progressing", "progressing", false},
		// Unknown status
		{"unknown->pending", "unknown", "pending", false},
		{"pending->unknown", "pending", "unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTransition(MainItemTransitions, tt.from, tt.to)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidTransition_SubItem(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		// Legal paths
		{"pending->progressing", "pending", "progressing", true},
		{"pending->closed", "pending", "closed", true},
		{"progressing->blocking", "progressing", "blocking", true},
		{"progressing->pausing", "progressing", "pausing", true},
		{"progressing->completed", "progressing", "completed", true},
		{"progressing->closed", "progressing", "closed", true},
		{"blocking->progressing", "blocking", "progressing", true},
		{"pausing->progressing", "pausing", "progressing", true},
		{"pausing->closed", "pausing", "closed", true},
		// Illegal paths
		{"pending->completed", "pending", "completed", false},
		{"completed->pending", "completed", "pending", false},
		{"blocking->closed", "blocking", "closed", false},
		// Self transitions
		{"pending->pending", "pending", "pending", false},
		// Unknown status
		{"unknown->pending", "unknown", "pending", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTransition(SubItemTransitions, tt.from, tt.to)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAvailableTransitions_MainItem(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    []string
	}{
		{"pending", "pending", []string{"progressing", "closed"}},
		{"progressing", "progressing", []string{"blocking", "pausing", "reviewing", "closed"}},
		{"blocking", "blocking", []string{"progressing"}},
		{"pausing", "pausing", []string{"progressing", "closed"}},
		{"reviewing", "reviewing", []string{"completed", "progressing"}},
		{"completed", "completed", []string{}},
		{"closed", "closed", []string{}},
		{"unknown", "unknown", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAvailableTransitions(MainItemTransitions, tt.current)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAvailableTransitions_SubItem(t *testing.T) {
	tests := []struct {
		name    string
		current string
		want    []string
	}{
		{"pending", "pending", []string{"progressing", "closed"}},
		{"progressing", "progressing", []string{"blocking", "pausing", "completed", "closed"}},
		{"blocking", "blocking", []string{"progressing"}},
		{"pausing", "pausing", []string{"progressing", "closed"}},
		{"completed", "completed", []string{}},
		{"closed", "closed", []string{}},
		{"unknown", "unknown", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAvailableTransitions(SubItemTransitions, tt.current)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAvailableTransitions_ReturnsCopy(t *testing.T) {
	original := GetAvailableTransitions(MainItemTransitions, "pending")
	modified := GetAvailableTransitions(MainItemTransitions, "pending")

	// Mutate the returned slice
	if len(modified) > 0 {
		modified[0] = "mutated"
	}

	// Original should be unchanged
	unchanged := GetAvailableTransitions(MainItemTransitions, "pending")
	assert.Equal(t, original, unchanged)
}
