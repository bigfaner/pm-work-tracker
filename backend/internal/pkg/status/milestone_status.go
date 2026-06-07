// Package status provides milestone status definitions and transition maps.
//
//nolint:misspell // "cancelled" is a domain status value per PRD/API contract, not a misspelling.
package status

// MilestoneMapStatuses defines the 6-state lifecycle for milestone maps.
var MilestoneMapStatuses = map[string]StatusDef{
	"planning":  {Code: "planning", Name: "规划中", Terminal: false},
	"reviewed":  {Code: "reviewed", Name: "已评审", Terminal: false},
	"ready":     {Code: "ready", Name: "待实施", Terminal: false},
	"executing": {Code: "executing", Name: "实施中", Terminal: false},
	"completed": {Code: "completed", Name: "已完成", Terminal: true},
	"cancelled": {Code: "cancelled", Name: "已取消", Terminal: true},
}

// MilestoneStatuses defines the 4-state lifecycle for milestones.
var MilestoneStatuses = map[string]StatusDef{
	"not_started": {Code: "not_started", Name: "未开始", Terminal: false},
	"in_progress": {Code: "in_progress", Name: "进行中", Terminal: false},
	"completed":   {Code: "completed", Name: "已完成", Terminal: true},
	"cancelled":   {Code: "cancelled", Name: "已取消", Terminal: true},
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

// MilestoneMapTransitions defines legal status transitions for milestone maps.
// Terminal states (completed, cancelled) have no entries.
var MilestoneMapTransitions = map[string][]string{
	"planning":  {"reviewed", "cancelled"},
	"reviewed":  {"ready", "planning", "cancelled"},
	"ready":     {"executing", "reviewed", "cancelled"},
	"executing": {"completed", "ready", "cancelled"},
}

// MilestoneTransitions defines legal status transitions for milestones.
// Terminal state (cancelled) has no entries.
var MilestoneTransitions = map[string][]string{
	"not_started": {"in_progress", "cancelled"},
	"in_progress": {"completed", "cancelled"},
	"completed":   {"cancelled", "in_progress"},
}
