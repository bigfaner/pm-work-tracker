package integration

import (
	"sort"
	"testing"

	"pm-work-tracker/backend/internal/pkg/permissions"
)

// testedCodes is the explicit contract: every code listed here must have
// a corresponding test in permission_matrix_test.go or rbac_permission_test.go.
// When a new code is added to permissions/codes.go, add it here AND write a test.
var testedCodes = map[string]bool{
	// team
	"team:create": true, "team:read": true, "team:update": true,
	"team:delete": true, "team:invite": true, "team:remove": true, "team:transfer": true,
	// main_item
	"main_item:create": true, "main_item:read": true, "main_item:update": true,
	"main_item:archive": true, "main_item:change_status": true,
	// sub_item
	"sub_item:create": true, "sub_item:read": true, "sub_item:update": true,
	"sub_item:assign": true, "sub_item:change_status": true,
	// progress
	"progress:create": true, "progress:read": true, "progress:update": true,
	// item_pool
	"item_pool:submit": true, "item_pool:review": true,
	// view
	"view:weekly": true, "view:gantt": true, "view:table": true,
	// report
	"report:export": true,
	// user
	"user:read": true, "user:update": true, "user:manage_role": true,
}

func TestPermissionCodeCoverage(t *testing.T) {
	allCodes := permissions.AllCodes()
	var missing []string
	for code := range allCodes {
		if !testedCodes[code] {
			missing = append(missing, code)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("missing test coverage for: %v", missing)
	}
}
