//nolint:misspell // "cancelled" is a domain status value per PRD/API contract.
package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- MilestoneMapStatuses ---

func TestMilestoneMapStatuses_Count(t *testing.T) {
	assert.Len(t, MilestoneMapStatuses, 6)
}

func TestMilestoneMapStatuses_ContainsAll(t *testing.T) {
	expected := []string{"planning", "reviewed", "ready", "executing", "completed", "cancelled"}
	for _, code := range expected {
		_, ok := MilestoneMapStatuses[code]
		assert.True(t, ok, "MilestoneMapStatuses should contain %q", code)
	}
}

func TestMilestoneMapStatuses_Definitions(t *testing.T) {
	tests := []struct {
		code     string
		name     string
		terminal bool
	}{
		{"planning", "规划中", false},
		{"reviewed", "已评审", false},
		{"ready", "待实施", false},
		{"executing", "实施中", false},
		{"completed", "已完成", true},
		{"cancelled", "已取消", true},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			def, ok := MilestoneMapStatuses[tt.code]
			assert.True(t, ok)
			assert.Equal(t, tt.code, def.Code)
			assert.Equal(t, tt.name, def.Name)
			assert.Equal(t, tt.terminal, def.Terminal)
		})
	}
}

// --- MilestoneStatuses ---

func TestMilestoneStatuses_Count(t *testing.T) {
	assert.Len(t, MilestoneStatuses, 4)
}

func TestMilestoneStatuses_ContainsAll(t *testing.T) {
	expected := []string{"not_started", "in_progress", "completed", "cancelled"}
	for _, code := range expected {
		_, ok := MilestoneStatuses[code]
		assert.True(t, ok, "MilestoneStatuses should contain %q", code)
	}
}

func TestMilestoneStatuses_Definitions(t *testing.T) {
	tests := []struct {
		code     string
		name     string
		terminal bool
	}{
		{"not_started", "未开始", false},
		{"in_progress", "进行中", false},
		{"completed", "已完成", true},
		{"cancelled", "已取消", true},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			def, ok := MilestoneStatuses[tt.code]
			assert.True(t, ok)
			assert.Equal(t, tt.code, def.Code)
			assert.Equal(t, tt.name, def.Name)
			assert.Equal(t, tt.terminal, def.Terminal)
		})
	}
}

// --- Getter functions ---

func TestGetMilestoneMapStatus(t *testing.T) {
	def, ok := GetMilestoneMapStatus("planning")
	assert.True(t, ok)
	assert.Equal(t, "规划中", def.Name)

	_, ok = GetMilestoneMapStatus("unknown")
	assert.False(t, ok)
}

func TestGetMilestoneStatus(t *testing.T) {
	def, ok := GetMilestoneStatus("in_progress")
	assert.True(t, ok)
	assert.Equal(t, "进行中", def.Name)

	_, ok = GetMilestoneStatus("unknown")
	assert.False(t, ok)
}

// --- MilestoneMapTransitions ---

func TestMilestoneMapTransitions_EntryCount(t *testing.T) {
	// 4 non-terminal states have transition entries (planning, reviewed, ready, executing)
	assert.Len(t, MilestoneMapTransitions, 4)
}

func TestMilestoneMapTransitions_TotalPaths(t *testing.T) {
	// planning:2 + reviewed:3 + ready:3 + executing:3 = 11
	total := 0
	for _, targets := range MilestoneMapTransitions {
		total += len(targets)
	}
	assert.Equal(t, 11, total, "MilestoneMapTransitions should define exactly 11 paths")
}

func TestMilestoneMapTransitions_AllPaths(t *testing.T) {
	expected := map[string][]string{
		"planning":  {"reviewed", "cancelled"},
		"reviewed":  {"ready", "planning", "cancelled"},
		"ready":     {"executing", "reviewed", "cancelled"},
		"executing": {"completed", "ready", "cancelled"},
	}
	for from, expectedTargets := range expected {
		t.Run(from, func(t *testing.T) {
			assert.Equal(t, expectedTargets, MilestoneMapTransitions[from])
		})
	}
}

func TestMilestoneMapTransitions_TerminalStatesHaveNoEntries(t *testing.T) {
	_, ok := MilestoneMapTransitions["completed"]
	assert.False(t, ok, "completed is terminal, should have no transitions")
	_, ok = MilestoneMapTransitions["cancelled"]
	assert.False(t, ok, "cancelled is terminal, should have no transitions")
}

// --- MilestoneTransitions ---

func TestMilestoneTransitions_EntryCount(t *testing.T) {
	// 3 non-terminal states have entries (not_started, in_progress, completed)
	assert.Len(t, MilestoneTransitions, 3)
}

func TestMilestoneTransitions_TotalPaths(t *testing.T) {
	// not_started:2 + in_progress:2 + completed:2 = 6
	total := 0
	for _, targets := range MilestoneTransitions {
		total += len(targets)
	}
	assert.Equal(t, 6, total, "MilestoneTransitions should define exactly 6 paths")
}

func TestMilestoneTransitions_AllPaths(t *testing.T) {
	expected := map[string][]string{
		"not_started": {"in_progress", "cancelled"},
		"in_progress": {"completed", "cancelled"},
		"completed":   {"cancelled", "in_progress"},
	}
	for from, expectedTargets := range expected {
		t.Run(from, func(t *testing.T) {
			assert.Equal(t, expectedTargets, MilestoneTransitions[from])
		})
	}
}

func TestMilestoneTransitions_CancelledIsTerminal(t *testing.T) {
	_, ok := MilestoneTransitions["cancelled"]
	assert.False(t, ok, "cancelled is terminal, should have no transitions")
}

// --- Transition validation with milestone maps ---

func TestIsValidTransition_MilestoneMap(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		// Legal forward
		{"planning->reviewed", "planning", "reviewed", true},
		{"reviewed->ready", "reviewed", "ready", true},
		{"ready->executing", "ready", "executing", true},
		{"executing->completed", "executing", "completed", true},
		// Legal backward
		{"reviewed->planning", "reviewed", "planning", true},
		{"ready->reviewed", "ready", "reviewed", true},
		{"executing->ready", "executing", "ready", true},
		// Cancel from any non-terminal
		{"planning->cancelled", "planning", "cancelled", true},
		{"reviewed->cancelled", "reviewed", "cancelled", true},
		{"ready->cancelled", "ready", "cancelled", true},
		{"executing->cancelled", "executing", "cancelled", true},
		// Illegal
		{"planning->completed", "planning", "completed", false},
		{"planning->executing", "planning", "executing", false},
		{"completed->planning", "completed", "planning", false},
		{"cancelled->planning", "cancelled", "planning", false},
		// Self
		{"planning->planning", "planning", "planning", false},
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
		// Legal forward
		{"not_started->in_progress", "not_started", "in_progress", true},
		{"in_progress->completed", "in_progress", "completed", true},
		// Legal cancel
		{"not_started->cancelled", "not_started", "cancelled", true},
		{"in_progress->cancelled", "in_progress", "cancelled", true},
		{"completed->cancelled", "completed", "cancelled", true},
		// Legal reopen
		{"completed->in_progress", "completed", "in_progress", true},
		// Illegal
		{"not_started->completed", "not_started", "completed", false},
		{"cancelled->in_progress", "cancelled", "in_progress", false},
		// Self
		{"in_progress->in_progress", "in_progress", "in_progress", false},
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
		{"planning", "planning", []string{"reviewed", "cancelled"}},
		{"reviewed", "reviewed", []string{"ready", "planning", "cancelled"}},
		{"ready", "ready", []string{"executing", "reviewed", "cancelled"}},
		{"executing", "executing", []string{"completed", "ready", "cancelled"}},
		{"completed", "completed", []string{}},
		{"cancelled", "cancelled", []string{}},
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
		{"completed", "completed", []string{"cancelled", "in_progress"}},
		{"cancelled", "cancelled", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAvailableTransitions(MilestoneTransitions, tt.current)
			assert.Equal(t, tt.want, got)
		})
	}
}
