package handler

import (
	"testing"
)

// assertPanicMessage is a helper that runs f and verifies it panics with the expected message.
func assertPanicMessage(t *testing.T, name string, f func(), want string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("%s: expected panic containing %q, but did not panic", name, want)
		}
		got, ok := r.(string)
		if !ok {
			t.Fatalf("%s: panic value is %T (%v), want string containing %q", name, r, r, want)
		}
		if got != want {
			t.Fatalf("%s: panic message = %q, want %q", name, got, want)
		}
	}()
	f()
}

// ========== SubItemHandler ==========

func TestNewSubItemHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "SubItemHandler/nil-svc", func() {
		NewSubItemHandler(nil, &StubMainItemSvc{})
	}, "sub_item_handler: subItemService must not be nil")
}

func TestNewSubItemHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewSubItemHandler(&StubSubItemSvc{}, &StubMainItemSvc{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ========== MainItemHandler ==========

func TestNewMainItemHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "MainItemHandler/nil-svc", func() {
		NewMainItemHandler(nil, &StubRouterRepoUser{}, &StubRouterRepoSubItem{})
	}, "main_item_handler: mainItemService must not be nil")
}

func TestNewMainItemHandler_PanicsOnNilUserRepo(t *testing.T) {
	assertPanicMessage(t, "MainItemHandler/nil-userRepo", func() {
		NewMainItemHandler(&StubMainItemSvc{}, nil, &StubRouterRepoSubItem{})
	}, "main_item_handler: userRepo must not be nil")
}

func TestNewMainItemHandler_PanicsOnNilSubItemRepo(t *testing.T) {
	assertPanicMessage(t, "MainItemHandler/nil-subItemRepo", func() {
		NewMainItemHandler(&StubMainItemSvc{}, &StubRouterRepoUser{}, nil)
	}, "main_item_handler: subItemRepo must not be nil")
}

func TestNewMainItemHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewMainItemHandler(&StubMainItemSvc{}, &StubRouterRepoUser{}, &StubRouterRepoSubItem{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ========== TeamHandler ==========

func TestNewTeamHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "TeamHandler/nil-svc", func() {
		NewTeamHandler(nil, &StubRouterRepoUser{})
	}, "team_handler: teamService must not be nil")
}

func TestNewTeamHandler_PanicsOnNilUserRepo(t *testing.T) {
	assertPanicMessage(t, "TeamHandler/nil-userRepo", func() {
		NewTeamHandler(&StubTeamSvc{}, nil)
	}, "team_handler: userRepo must not be nil")
}

func TestNewTeamHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewTeamHandler(&StubTeamSvc{}, &StubRouterRepoUser{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ========== ItemPoolHandler ==========

func TestNewItemPoolHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "ItemPoolHandler/nil-svc", func() {
		NewItemPoolHandler(nil, &StubRouterRepoUser{}, &StubRouterRepoMainItem{})
	}, "item_pool_handler: itemPoolService must not be nil")
}

func TestNewItemPoolHandler_PanicsOnNilUserRepo(t *testing.T) {
	assertPanicMessage(t, "ItemPoolHandler/nil-userRepo", func() {
		NewItemPoolHandler(&StubItemPoolSvc{}, nil, &StubRouterRepoMainItem{})
	}, "item_pool_handler: userRepo must not be nil")
}

func TestNewItemPoolHandler_PanicsOnNilMainItemRepo(t *testing.T) {
	assertPanicMessage(t, "ItemPoolHandler/nil-mainItemRepo", func() {
		NewItemPoolHandler(&StubItemPoolSvc{}, &StubRouterRepoUser{}, nil)
	}, "item_pool_handler: mainItemRepo must not be nil")
}

func TestNewItemPoolHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewItemPoolHandler(&StubItemPoolSvc{}, &StubRouterRepoUser{}, &StubRouterRepoMainItem{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ========== RoleHandler ==========

func TestNewRoleHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "RoleHandler/nil-svc", func() {
		NewRoleHandler(nil)
	}, "role_handler: roleService must not be nil")
}

func TestNewRoleHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewRoleHandler(&StubRoleSvc{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ========== AdminHandler ==========

func TestNewAdminHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "AdminHandler/nil-svc", func() {
		NewAdminHandler(nil)
	}, "admin_handler: adminService must not be nil")
}

func TestNewAdminHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewAdminHandler(&StubAdminSvc{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ========== ReportHandler ==========

func TestNewReportHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "ReportHandler/nil-svc", func() {
		NewReportHandler(nil)
	}, "report_handler: reportService must not be nil")
}

func TestNewReportHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewReportHandler(&StubReportSvc{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ========== ViewHandler ==========

func TestNewViewHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "ViewHandler/nil-svc", func() {
		NewViewHandler(nil)
	}, "view_handler: viewService must not be nil")
}

func TestNewViewHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewViewHandler(&StubViewSvc{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ========== ProgressHandler ==========

func TestNewProgressHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "ProgressHandler/nil-svc", func() {
		NewProgressHandler(nil, &StubRouterRepoUser{}, &StubSubItemSvc{})
	}, "progress_handler: progressService must not be nil")
}

func TestNewProgressHandler_PanicsOnNilUserRepo(t *testing.T) {
	assertPanicMessage(t, "ProgressHandler/nil-userRepo", func() {
		NewProgressHandler(&StubProgressSvc{}, nil, &StubSubItemSvc{})
	}, "progress_handler: userRepo must not be nil")
}

func TestNewProgressHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewProgressHandler(&StubProgressSvc{}, &StubRouterRepoUser{}, &StubSubItemSvc{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

// ========== PermissionHandler ==========

func TestNewPermissionHandler_PanicsOnNilService(t *testing.T) {
	assertPanicMessage(t, "PermissionHandler/nil-svc", func() {
		NewPermissionHandler(nil)
	}, "permission_handler: roleService must not be nil")
}

func TestNewPermissionHandler_SucceedsWithValidDeps(t *testing.T) {
	h := NewPermissionHandler(&StubRoleSvc{})
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}
