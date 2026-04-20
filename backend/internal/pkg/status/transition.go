package status

// MainItemTransitions defines the legal status transition paths for main items.
var MainItemTransitions = map[string][]string{
	"pending":     {"progressing", "closed"},
	"progressing": {"blocking", "pausing", "reviewing", "closed"},
	"blocking":    {"progressing"},
	"pausing":     {"progressing", "closed"},
	"reviewing":   {"completed", "progressing"},
}

// SubItemTransitions defines the legal status transition paths for sub items.
var SubItemTransitions = map[string][]string{
	"pending":     {"progressing", "closed"},
	"progressing": {"blocking", "pausing", "completed", "closed"},
	"blocking":    {"progressing"},
	"pausing":     {"progressing", "closed"},
}

// IsValidTransition checks whether transitioning from one status to another is allowed.
// Returns false for self-transitions or unknown status codes.
func IsValidTransition(transitions map[string][]string, from, to string) bool {
	if from == to {
		return false
	}
	targets, ok := transitions[from]
	if !ok {
		return false
	}
	for _, target := range targets {
		if target == to {
			return true
		}
	}
	return false
}

// GetAvailableTransitions returns the list of valid target statuses from the current status.
// Returns an empty slice for terminal or unknown statuses.
// Returns a copy to prevent mutation of the original data.
func GetAvailableTransitions(transitions map[string][]string, current string) []string {
	targets, ok := transitions[current]
	if !ok {
		return []string{}
	}
	result := make([]string, len(targets))
	copy(result, targets)
	return result
}
