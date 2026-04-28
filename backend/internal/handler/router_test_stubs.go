package handler

import (
	"context"
	"fmt"
	"time"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/permissions"
	"pm-work-tracker/backend/internal/repository"
	"pm-work-tracker/backend/internal/service"
)

// Stub services for use in test wiring where handlers are never actually invoked.
// Exported so they can be used from other test packages (e.g. cmd/server).

type StubTeamSvc struct{}

func (s *StubTeamSvc) CreateTeam(_ context.Context, _ uint, _ dto.CreateTeamReq) (*model.Team, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) GetTeam(_ context.Context, _ uint) (*model.Team, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) GetTeamDetail(_ context.Context, _ uint) (*dto.TeamDetailResp, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) ListTeams(_ context.Context, _ uint, _ bool, _ string, _, _ int) ([]*dto.TeamListResp, int64, error) {
	return nil, 0, fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) UpdateTeam(_ context.Context, _, _ uint, _ dto.UpdateTeamReq) (*model.Team, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) InviteMember(_ context.Context, _, _ uint, _ dto.InviteMemberReq) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) RemoveMember(_ context.Context, _, _, _ uint) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) TransferPM(_ context.Context, _, _, _ uint) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) DisbandTeam(_ context.Context, _ uint, _ uint, _ string) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) UpdateMemberRole(_ context.Context, _, _, _, _ uint) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) ListMembers(_ context.Context, _ uint) ([]*dto.TeamMemberDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubTeamSvc) SearchAvailableUsers(_ context.Context, _ uint, _ string) ([]*dto.UserSearchDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.TeamService = (*StubTeamSvc)(nil)

type StubMainItemSvc struct{}

func (s *StubMainItemSvc) Create(_ context.Context, _, _ uint, _ dto.MainItemCreateReq) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubMainItemSvc) Update(_ context.Context, _ uint, _ uint, _ dto.MainItemUpdateReq) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubMainItemSvc) Archive(_ context.Context, _ uint, _ uint) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubMainItemSvc) List(_ context.Context, _ uint, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubMainItemSvc) Get(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubMainItemSvc) GetByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubMainItemSvc) RecalcCompletion(_ context.Context, _ uint) error {
	return fmt.Errorf("stub: not implemented")
}

func (s *StubMainItemSvc) ChangeStatus(_ context.Context, _, _, _ uint, _ string) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

func (s *StubMainItemSvc) AvailableTransitions(_ context.Context, _, _, _ uint) ([]string, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

func (s *StubMainItemSvc) EvaluateLinkage(_ context.Context, _ uint, _ uint) (*service.LinkageResult, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.MainItemService = (*StubMainItemSvc)(nil)

type StubSubItemSvc struct{}

func (s *StubSubItemSvc) Create(_ context.Context, _, _ uint, _ dto.SubItemCreateReq) (*model.SubItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubSubItemSvc) Update(_ context.Context, _, _ uint, _ dto.SubItemUpdateReq) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubSubItemSvc) ChangeStatus(_ context.Context, _, _, _ uint, _ string) (*service.SubItemChangeResult, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubSubItemSvc) Get(_ context.Context, _, _ uint) (*model.SubItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubSubItemSvc) List(_ context.Context, _ uint, _ *uint, _ dto.SubItemFilter, _ dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubSubItemSvc) Assign(_ context.Context, _, _, _, _ uint) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubSubItemSvc) AvailableTransitions(_ context.Context, _, _ uint) ([]string, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubSubItemSvc) Delete(_ context.Context, _, _, _ uint) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubSubItemSvc) GetByBizKey(_ context.Context, _ int64) (*model.SubItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.SubItemService = (*StubSubItemSvc)(nil)

type StubItemPoolSvc struct{}

func (s *StubItemPoolSvc) Submit(_ context.Context, _, _ uint, _ dto.SubmitItemPoolReq) (*model.ItemPool, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubItemPoolSvc) Assign(_ context.Context, _, _, _ uint, _ dto.AssignItemPoolReq) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubItemPoolSvc) ConvertToMain(_ context.Context, _, _, _ uint, _ dto.ConvertToMainItemReq) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubItemPoolSvc) Reject(_ context.Context, _, _, _ uint, _ string) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubItemPoolSvc) List(_ context.Context, _ uint, _ dto.ItemPoolFilter, _ dto.Pagination) (*dto.PageResult[model.ItemPool], error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubItemPoolSvc) Get(_ context.Context, _, _ uint) (*model.ItemPool, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubItemPoolSvc) GetByBizKey(_ context.Context, _ int64) (*model.ItemPool, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubItemPoolSvc) Update(_ context.Context, _, _ uint, _ dto.UpdateItemPoolReq) (*model.ItemPool, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.ItemPoolService = (*StubItemPoolSvc)(nil)

type StubAdminSvc struct{}

func (s *StubAdminSvc) ListUsers(_ context.Context, _ string, _, _ int) ([]*dto.AdminUserDTO, int, error) {
	return nil, 0, fmt.Errorf("stub: not implemented")
}
func (s *StubAdminSvc) GetUser(_ context.Context, _ int64) (*dto.AdminUserDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubAdminSvc) CreateUser(_ context.Context, _ *dto.CreateUserReq) (*dto.AdminUserDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubAdminSvc) UpdateUser(_ context.Context, _ int64, _ *dto.UpdateUserReq) (*dto.AdminUserDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubAdminSvc) ToggleUserStatus(_ context.Context, _ uint, _ int64, _ string) (*dto.AdminUserDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubAdminSvc) ListAllTeams(_ context.Context) ([]*dto.AdminTeamDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubAdminSvc) ResetPassword(_ context.Context, _ int64, _ string) (*dto.ResetPasswordResp, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubAdminSvc) SoftDeleteUser(_ context.Context, _ uint, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}

var _ service.AdminService = (*StubAdminSvc)(nil)

type StubRoleSvc struct{}

func (s *StubRoleSvc) ListRoles(_ context.Context) ([]service.RoleListItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubRoleSvc) GetRole(_ context.Context, _ int64) (*service.RoleDetail, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubRoleSvc) CreateRole(_ context.Context, _ dto.CreateRoleReq) (*service.RoleListItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubRoleSvc) UpdateRole(_ context.Context, _ int64, _ dto.UpdateRoleReq) (*service.RoleDetail, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubRoleSvc) DeleteRole(_ context.Context, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubRoleSvc) ListPermissionCodes(_ context.Context) []permissions.ResourcePermissions {
	return nil
}
func (s *StubRoleSvc) GetUserPermissions(_ context.Context, _ uint) (*service.UserPermissions, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.RoleService = (*StubRoleSvc)(nil)

// Stub repos for constructor args

type StubRouterRepoUser struct{}

func (s *StubRouterRepoUser) FindByID(_ context.Context, _ uint) (*model.User, error) {
	return nil, nil
}
func (s *StubRouterRepoUser) FindByUsername(_ context.Context, _ string) (*model.User, error) {
	return nil, nil
}
func (s *StubRouterRepoUser) List(_ context.Context) ([]*model.User, error) { return nil, nil }
func (s *StubRouterRepoUser) Update(_ context.Context, _ *model.User) error { return nil }
func (s *StubRouterRepoUser) Create(_ context.Context, _ *model.User) error { return nil }
func (s *StubRouterRepoUser) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.User, error) {
	return nil, nil
}
func (s *StubRouterRepoUser) FindByBizKey(_ context.Context, _ int64) (*model.User, error) {
	return nil, nil
}
func (s *StubRouterRepoUser) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
	return nil, 0, nil
}
func (s *StubRouterRepoUser) SearchAvailable(_ context.Context, _ uint, _ string, _ int) ([]*model.User, error) {
	return nil, nil
}
func (s *StubRouterRepoUser) SoftDelete(_ context.Context, _ *model.User) error { return nil }

var _ repository.UserRepo = (*StubRouterRepoUser)(nil)

type StubRouterRepoSubItem struct{}

func (s *StubRouterRepoSubItem) Create(_ context.Context, _ *model.SubItem) error { return nil }
func (s *StubRouterRepoSubItem) FindByID(_ context.Context, _ uint) (*model.SubItem, error) {
	return nil, nil
}
func (s *StubRouterRepoSubItem) Update(_ context.Context, _ *model.SubItem, _ map[string]interface{}) error {
	return nil
}
func (s *StubRouterRepoSubItem) List(_ context.Context, _, _ uint, _ dto.SubItemFilter, _ dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}
func (s *StubRouterRepoSubItem) ListByMainItem(_ context.Context, _ uint) ([]*model.SubItem, error) {
	return nil, nil
}
func (s *StubRouterRepoSubItem) ListByTeam(_ context.Context, _ uint) ([]model.SubItem, error) {
	return nil, nil
}
func (s *StubRouterRepoSubItem) SoftDelete(_ context.Context, _ uint) error { return nil }
func (s *StubRouterRepoSubItem) FindByBizKey(_ context.Context, _ int64) (*model.SubItem, error) {
	return nil, nil
}
func (s *StubRouterRepoSubItem) NextSubCode(_ context.Context, _ uint) (string, error) {
	return "", nil
}

var _ repository.SubItemRepo = (*StubRouterRepoSubItem)(nil)

type StubRouterRepoMainItem struct{}

func (s *StubRouterRepoMainItem) Create(_ context.Context, _ *model.MainItem) error { return nil }
func (s *StubRouterRepoMainItem) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}
func (s *StubRouterRepoMainItem) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}

func (s *StubRouterRepoMainItem) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}
func (s *StubRouterRepoMainItem) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}
func (s *StubRouterRepoMainItem) List(_ context.Context, _ uint, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}
func (s *StubRouterRepoMainItem) NextCode(_ context.Context, _ uint) (string, error) {
	return "", nil
}
func (s *StubRouterRepoMainItem) CountByTeam(_ context.Context, _ uint) (int64, error) {
	return 0, nil
}
func (s *StubRouterRepoMainItem) ListNonArchivedByTeam(_ context.Context, _ uint) ([]model.MainItem, error) {
	return nil, nil
}
func (s *StubRouterRepoMainItem) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}
func (s *StubRouterRepoMainItem) ListByTeamAndStatus(_ context.Context, _ uint, _ string) ([]model.MainItem, error) {
	return nil, nil
}

var _ repository.MainItemRepo = (*StubRouterRepoMainItem)(nil)

// StubReportSvc is a stub ReportService for test wiring.
type StubReportSvc struct{}

func (s *StubReportSvc) Preview(_ context.Context, _ uint, _ time.Time) (*dto.ReportPreview, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubReportSvc) ExportMarkdown(_ context.Context, _ uint, _ time.Time) ([]byte, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.ReportService = (*StubReportSvc)(nil)

// StubViewSvc is a stub ViewService for test wiring.
type StubViewSvc struct{}

func (s *StubViewSvc) WeeklyComparison(_ context.Context, _ uint, _ time.Time) (*dto.WeeklyViewResponse, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubViewSvc) GanttView(_ context.Context, _ uint, _ dto.GanttFilter) (*dto.GanttResult, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubViewSvc) TableView(_ context.Context, _ uint, _ dto.TableFilter, _ dto.Pagination) (*dto.PageResult[dto.TableRow], error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubViewSvc) TableExportCSV(_ context.Context, _ uint, _ dto.TableFilter) ([]byte, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.ViewService = (*StubViewSvc)(nil)

// StubProgressSvc is a stub ProgressService for test wiring.
type StubProgressSvc struct{}

func (s *StubProgressSvc) Append(_ context.Context, _, _, _ uint, _ float64, _, _, _ string, _ bool) (*model.ProgressRecord, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubProgressSvc) CorrectCompletion(_ context.Context, _, _ uint, _ float64) error {
	return fmt.Errorf("stub: not implemented")
}
func (s *StubProgressSvc) List(_ context.Context, _, _ uint) ([]model.ProgressRecord, error) {
	return nil, fmt.Errorf("stub: not implemented")
}
func (s *StubProgressSvc) GetByBizKey(_ context.Context, _ int64) (*model.ProgressRecord, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.ProgressService = (*StubProgressSvc)(nil)
