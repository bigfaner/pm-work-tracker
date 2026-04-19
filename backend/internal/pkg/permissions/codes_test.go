package permissions

import (
	"testing"
)

func TestRegistryContainsAllResources(t *testing.T) {
	wantResources := []string{
		"team", "main_item", "sub_item", "progress",
		"item_pool", "view", "report", "user",
	}
	if len(Registry) != len(wantResources) {
		t.Fatalf("Registry has %d resources, want %d", len(Registry), len(wantResources))
	}
	for i, want := range wantResources {
		if Registry[i].Resource != want {
			t.Errorf("Registry[%d].Resource = %q, want %q", i, Registry[i].Resource, want)
		}
	}
}

func TestRegistryResourceCodeCounts(t *testing.T) {
	wantCounts := map[string]int{
		"team":      7,
		"main_item": 4,
		"sub_item":  5,
		"progress":  3,
		"item_pool": 2,
		"view":      3,
		"report":    1,
		"user":      3,
	}
	for _, rp := range Registry {
		want, ok := wantCounts[rp.Resource]
		if !ok {
			t.Errorf("unexpected resource %q in Registry", rp.Resource)
			continue
		}
		if len(rp.Permissions) != want {
			t.Errorf("resource %q has %d permissions, want %d", rp.Resource, len(rp.Permissions), want)
		}
	}
}

func TestAllCodesReturnsExpectedCount(t *testing.T) {
	codes := AllCodes()
	wantCount := 28 // 7+4+5+3+2+3+1+3
	if len(codes) != wantCount {
		t.Fatalf("AllCodes() returned %d codes, want %d", len(codes), wantCount)
	}
}

func TestAllCodesContainsSpecificCodes(t *testing.T) {
	codes := AllCodes()
	mustHave := []string{
		"team:create", "team:read", "team:update", "team:delete", "team:invite", "team:remove", "team:transfer",
		"main_item:create", "main_item:read", "main_item:update", "main_item:archive",
		"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:assign", "sub_item:change_status",
		"progress:create", "progress:read", "progress:update",
		"item_pool:submit", "item_pool:review",
		"view:weekly", "view:gantt", "view:table",
		"report:export",
		"user:read", "user:update", "user:manage_role",
	}
	for _, code := range mustHave {
		if !codes[code] {
			t.Errorf("AllCodes() missing code %q", code)
		}
	}
}

func TestAllCodesReturnsCopy(t *testing.T) {
	codes1 := AllCodes()
	codes2 := AllCodes()
	codes1["fake:code"] = true
	if codes2["fake:code"] {
		t.Error("AllCodes() should return independent copies, not the same map")
	}
}

func TestValidateCodeValidCodes(t *testing.T) {
	validCodes := []string{
		"team:create", "team:read", "team:update", "team:delete",
		"main_item:archive", "sub_item:assign", "progress:update",
		"item_pool:review", "view:gantt", "report:export", "user:manage_role",
	}
	for _, code := range validCodes {
		if !ValidateCode(code) {
			t.Errorf("ValidateCode(%q) = false, want true", code)
		}
	}
}

func TestValidateCodeInvalidCodes(t *testing.T) {
	invalidCodes := []string{
		"",
		"team",
		"team:",
		":create",
		"fake:action",
		"TEAM:CREATE",
		"team:create ",
		" team:create",
		"nonexistent:code",
		"main_item:create_extra",
	}
	for _, code := range invalidCodes {
		if ValidateCode(code) {
			t.Errorf("ValidateCode(%q) = true, want false", code)
		}
	}
}

func TestPermissionDescriptions(t *testing.T) {
	for _, rp := range Registry {
		for _, p := range rp.Permissions {
			if p.Code == "" {
				t.Errorf("resource %q has empty permission code", rp.Resource)
			}
			if p.Description == "" {
				t.Errorf("permission %q has empty description", p.Code)
			}
		}
	}
}

func TestPermissionCodesFollowFormat(t *testing.T) {
	for _, rp := range Registry {
		for _, p := range rp.Permissions {
			// Must follow resource:action format
			if p.Code != rp.Resource+":"+extractAction(p.Code) {
				t.Errorf("permission code %q does not match resource %q", p.Code, rp.Resource)
			}
		}
	}
}

func extractAction(code string) string {
	for i := len(code) - 1; i >= 0; i-- {
		if code[i] == ':' {
			return code[i+1:]
		}
	}
	return ""
}
