// Package status provides status definitions and lookup helpers for work items.
package status

//nolint:misspell // "cancelled" is intentional British English spelling per design spec

// MilestoneMapStatuses is the single source of truth for milestone map status definitions.
// 5 states: planning, reviewed, ready, executing, completed (terminal).
var MilestoneMapStatuses = map[string]StatusDef{
	"planning":  {Code: "planning", Name: "规划中", Terminal: false},
	"reviewed":  {Code: "reviewed", Name: "已评审", Terminal: false},
	"ready":     {Code: "ready", Name: "就绪", Terminal: false},
	"executing": {Code: "executing", Name: "执行中", Terminal: false},
	"completed": {Code: "completed", Name: "已完成", Terminal: true},
}

// MilestoneStatuses is the single source of truth for milestone status definitions.
// 4 states: not_started, in_progress, completed (non-terminal), cancelled (terminal).
var MilestoneStatuses = map[string]StatusDef{
	"not_started": {Code: "not_started", Name: "未开始", Terminal: false},
	"in_progress": {Code: "in_progress", Name: "进行中", Terminal: false},
	"completed":   {Code: "completed", Name: "已完成", Terminal: false},
	"cancelled":   {Code: "cancelled", Name: "已取消", Terminal: true},
}

// MilestoneMapTransitions defines the legal status transition paths for milestone maps.
var MilestoneMapTransitions = map[string][]string{
	"planning":  {"reviewed"},
	"reviewed":  {"ready", "planning"},
	"ready":     {"executing", "reviewed"},
	"executing": {"completed", "ready"},
}

// MilestoneTransitions defines the legal status transition paths for milestones.
var MilestoneTransitions = map[string][]string{
	"not_started": {"in_progress", "cancelled"},
	"in_progress": {"completed", "cancelled"},
	"completed":   {"cancelled"},
}

// GetMilestoneMapStatus returns the StatusDef for a milestone map status code.
func GetMilestoneMapStatus(code string) (StatusDef, bool) {
	def, ok := MilestoneMapStatuses[code]
	return def, ok
}

// GetMilestoneStatus returns the StatusDef for a milestone status code.
func GetMilestoneStatus(code string) (StatusDef, bool) {
	def, ok := MilestoneStatuses[code]
	return def, ok
}

// IsMilestoneMapTerminal returns true if the code is a terminal status for milestone maps.
func IsMilestoneMapTerminal(code string) bool {
	def, ok := MilestoneMapStatuses[code]
	return ok && def.Terminal
}

// IsMilestoneTerminal returns true if the code is a terminal status for milestones.
func IsMilestoneTerminal(code string) bool {
	def, ok := MilestoneStatuses[code]
	return ok && def.Terminal
}
