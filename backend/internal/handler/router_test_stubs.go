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

// StubTeamSvc is a stub implementation of service.TeamService for testing.
type StubTeamSvc struct{}

// CreateTeam is a stub for service.TeamService.CreateTeam.
func (s *StubTeamSvc) CreateTeam(_ context.Context, _ int64, _ dto.CreateTeamReq) (*model.Team, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// GetTeam is a stub for service.TeamService.GetTeam.
func (s *StubTeamSvc) GetTeam(_ context.Context, _ int64) (*model.Team, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// GetTeamDetail is a stub for service.TeamService.GetTeamDetail.
func (s *StubTeamSvc) GetTeamDetail(_ context.Context, _ int64) (*dto.TeamDetailResp, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// ListTeams is a stub for service.TeamService.ListTeams.
func (s *StubTeamSvc) ListTeams(_ context.Context, _ int64, _ bool, _ string, _, _ int) ([]*dto.TeamListResp, int64, error) {
	return nil, 0, fmt.Errorf("stub: not implemented")
}

// UpdateTeam is a stub for service.TeamService.UpdateTeam.
func (s *StubTeamSvc) UpdateTeam(_ context.Context, _ int64, _ dto.UpdateTeamReq) (*model.Team, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// InviteMember is a stub for service.TeamService.InviteMember.
func (s *StubTeamSvc) InviteMember(_ context.Context, _, _ int64, _ dto.InviteMemberReq) error {
	return fmt.Errorf("stub: not implemented")
}

// RemoveMember is a stub for service.TeamService.RemoveMember.
func (s *StubTeamSvc) RemoveMember(_ context.Context, _, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}

// TransferPM is a stub for service.TeamService.TransferPM.
func (s *StubTeamSvc) TransferPM(_ context.Context, _, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}

// DisbandTeam is a stub for service.TeamService.DisbandTeam.
func (s *StubTeamSvc) DisbandTeam(_ context.Context, _ int64, _ string) error {
	return fmt.Errorf("stub: not implemented")
}

// UpdateMemberRole is a stub for service.TeamService.UpdateMemberRole.
func (s *StubTeamSvc) UpdateMemberRole(_ context.Context, _, _, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}

// ListMembers is a stub for service.TeamService.ListMembers.
func (s *StubTeamSvc) ListMembers(_ context.Context, _ int64) ([]*dto.TeamMemberDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// SearchAvailableUsers is a stub for service.TeamService.SearchAvailableUsers.
func (s *StubTeamSvc) SearchAvailableUsers(_ context.Context, _ int64, _ string) ([]*dto.UserSearchDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.TeamService = (*StubTeamSvc)(nil)

// StubMainItemSvc is a stub implementation of service.MainItemService for testing.
type StubMainItemSvc struct{}

// Create is a stub for service.MainItemService.Create.
func (s *StubMainItemSvc) Create(_ context.Context, _, _ int64, _ dto.MainItemCreateReq) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Update is a stub for service.MainItemService.Update.
func (s *StubMainItemSvc) Update(_ context.Context, _ int64, _ uint, _ dto.MainItemUpdateReq) error {
	return fmt.Errorf("stub: not implemented")
}

// Archive is a stub for service.MainItemService.Archive.
func (s *StubMainItemSvc) Archive(_ context.Context, _ int64, _ uint) error {
	return fmt.Errorf("stub: not implemented")
}

// List is a stub for service.MainItemService.List.
func (s *StubMainItemSvc) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], map[int64]*dto.MainItemMatchInfo, error) {
	return nil, nil, fmt.Errorf("stub: not implemented")
}

// Get is a stub for service.MainItemService.Get.
func (s *StubMainItemSvc) Get(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// GetByBizKey is a stub for service.MainItemService.GetByBizKey.
func (s *StubMainItemSvc) GetByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// RecalcCompletion is a stub for service.MainItemService.RecalcCompletion.
func (s *StubMainItemSvc) RecalcCompletion(_ context.Context, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}

// ChangeStatus is a stub for service.MainItemService.ChangeStatus.
func (s *StubMainItemSvc) ChangeStatus(_ context.Context, _, _ int64, _ uint, _ string) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// AvailableTransitions is a stub for service.MainItemService.AvailableTransitions.
func (s *StubMainItemSvc) AvailableTransitions(_ context.Context, _, _ int64, _ uint) ([]string, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// EvaluateLinkage is a stub for service.MainItemService.EvaluateLinkage.
func (s *StubMainItemSvc) EvaluateLinkage(_ context.Context, _, _ int64) (*service.LinkageResult, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Delete is a stub for service.MainItemService.Delete.
func (s *StubMainItemSvc) Delete(_ context.Context, _, _, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}

var _ service.MainItemService = (*StubMainItemSvc)(nil)

// StubSubItemSvc is a stub implementation of service.SubItemService for testing.
type StubSubItemSvc struct{}

// Create is a stub for service.SubItemService.Create.
func (s *StubSubItemSvc) Create(_ context.Context, _, _ int64, _ dto.SubItemCreateReq) (*model.SubItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Update is a stub for service.SubItemService.Update.
func (s *StubSubItemSvc) Update(_ context.Context, _ int64, _ uint, _ dto.SubItemUpdateReq) error {
	return fmt.Errorf("stub: not implemented")
}

// ChangeStatus is a stub for service.SubItemService.ChangeStatus.
func (s *StubSubItemSvc) ChangeStatus(_ context.Context, _, _ int64, _ uint, _ string) (*service.SubItemChangeResult, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Get is a stub for service.SubItemService.Get.
func (s *StubSubItemSvc) Get(_ context.Context, _ int64, _ uint) (*model.SubItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// List is a stub for service.SubItemService.List.
func (s *StubSubItemSvc) List(_ context.Context, _ int64, _ *int64, _ dto.SubItemFilter, _ dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Assign is a stub for service.SubItemService.Assign.
func (s *StubSubItemSvc) Assign(_ context.Context, _, _ int64, _ uint, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}

// AvailableTransitions is a stub for service.SubItemService.AvailableTransitions.
func (s *StubSubItemSvc) AvailableTransitions(_ context.Context, _ int64, _ uint) ([]string, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Delete is a stub for service.SubItemService.Delete.
func (s *StubSubItemSvc) Delete(_ context.Context, _, _ int64, _ uint) error {
	return fmt.Errorf("stub: not implemented")
}

// GetByBizKey is a stub for service.SubItemService.GetByBizKey.
func (s *StubSubItemSvc) GetByBizKey(_ context.Context, _ int64) (*model.SubItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Move is a stub for service.SubItemService.Move.
func (s *StubSubItemSvc) Move(_ context.Context, _, _, _, _ int64) (*service.MoveResult, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.SubItemService = (*StubSubItemSvc)(nil)

// StubItemPoolSvc is a stub implementation of service.ItemPoolService for testing.
type StubItemPoolSvc struct{}

// Submit is a stub for service.ItemPoolService.Submit.
func (s *StubItemPoolSvc) Submit(_ context.Context, _, _ int64, _ dto.SubmitItemPoolReq) (*model.ItemPool, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Assign is a stub for service.ItemPoolService.Assign.
func (s *StubItemPoolSvc) Assign(_ context.Context, _, _ int64, _ uint, _ dto.AssignItemPoolReq) error {
	return fmt.Errorf("stub: not implemented")
}

// ConvertToMain is a stub for service.ItemPoolService.ConvertToMain.
func (s *StubItemPoolSvc) ConvertToMain(_ context.Context, _, _ int64, _ uint, _ dto.ConvertToMainItemReq) (*model.MainItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Reject is a stub for service.ItemPoolService.Reject.
func (s *StubItemPoolSvc) Reject(_ context.Context, _, _ int64, _ uint, _ string) error {
	return fmt.Errorf("stub: not implemented")
}

// List is a stub for service.ItemPoolService.List.
func (s *StubItemPoolSvc) List(_ context.Context, _ int64, _ dto.ItemPoolFilter, _ dto.Pagination) (*dto.PageResult[model.ItemPool], error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Get is a stub for service.ItemPoolService.Get.
func (s *StubItemPoolSvc) Get(_ context.Context, _ int64, _ uint) (*model.ItemPool, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// GetByBizKey is a stub for service.ItemPoolService.GetByBizKey.
func (s *StubItemPoolSvc) GetByBizKey(_ context.Context, _ int64) (*model.ItemPool, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// Update is a stub for service.ItemPoolService.Update.
func (s *StubItemPoolSvc) Update(_ context.Context, _ int64, _ uint, _ dto.UpdateItemPoolReq) (*model.ItemPool, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.ItemPoolService = (*StubItemPoolSvc)(nil)

// StubAdminSvc is a stub implementation of service.AdminService for testing.
type StubAdminSvc struct{}

// ListUsers is a stub for service.AdminService.ListUsers.
func (s *StubAdminSvc) ListUsers(_ context.Context, _ string, _, _ int) ([]*dto.AdminUserDTO, int, error) {
	return nil, 0, fmt.Errorf("stub: not implemented")
}

// GetUser is a stub for service.AdminService.GetUser.
func (s *StubAdminSvc) GetUser(_ context.Context, _ int64) (*dto.AdminUserDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// CreateUser is a stub for service.AdminService.CreateUser.
func (s *StubAdminSvc) CreateUser(_ context.Context, _ *dto.CreateUserReq) (*dto.AdminUserDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// UpdateUser is a stub for service.AdminService.UpdateUser.
func (s *StubAdminSvc) UpdateUser(_ context.Context, _ int64, _ *dto.UpdateUserReq) (*dto.AdminUserDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// ToggleUserStatus is a stub for service.AdminService.ToggleUserStatus.
func (s *StubAdminSvc) ToggleUserStatus(_ context.Context, _ uint, _ int64, _ string) (*dto.AdminUserDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// ListAllTeams is a stub for service.AdminService.ListAllTeams.
func (s *StubAdminSvc) ListAllTeams(_ context.Context) ([]*dto.AdminTeamDTO, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// ResetPassword is a stub for service.AdminService.ResetPassword.
func (s *StubAdminSvc) ResetPassword(_ context.Context, _ int64, _ string) (*dto.ResetPasswordResp, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// SoftDeleteUser is a stub for service.AdminService.SoftDeleteUser.
func (s *StubAdminSvc) SoftDeleteUser(_ context.Context, _ uint, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}

var _ service.AdminService = (*StubAdminSvc)(nil)

// StubRoleSvc is a stub implementation of service.RoleService for testing.
type StubRoleSvc struct{}

// ListRoles is a stub for service.RoleService.ListRoles.
func (s *StubRoleSvc) ListRoles(_ context.Context, _ string) ([]service.RoleListItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// GetRole is a stub for service.RoleService.GetRole.
func (s *StubRoleSvc) GetRole(_ context.Context, _ int64) (*service.RoleDetail, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// CreateRole is a stub for service.RoleService.CreateRole.
func (s *StubRoleSvc) CreateRole(_ context.Context, _ dto.CreateRoleReq) (*service.RoleListItem, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// UpdateRole is a stub for service.RoleService.UpdateRole.
func (s *StubRoleSvc) UpdateRole(_ context.Context, _ int64, _ dto.UpdateRoleReq) (*service.RoleDetail, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// DeleteRole is a stub for service.RoleService.DeleteRole.
func (s *StubRoleSvc) DeleteRole(_ context.Context, _ int64) error {
	return fmt.Errorf("stub: not implemented")
}

// ListPermissionCodes is a stub for service.RoleService.ListPermissionCodes.
func (s *StubRoleSvc) ListPermissionCodes(_ context.Context) []permissions.ResourcePermissions {
	return nil
}

// GetUserPermissions is a stub for service.RoleService.GetUserPermissions.
func (s *StubRoleSvc) GetUserPermissions(_ context.Context, _ uint) (*service.UserPermissions, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.RoleService = (*StubRoleSvc)(nil)

// StubRouterRepoUser is a stub implementation of repository.UserRepo for testing.
type StubRouterRepoUser struct{}

// FindByID is a stub for repository.UserRepo.FindByID.
func (s *StubRouterRepoUser) FindByID(_ context.Context, _ uint) (*model.User, error) {
	return nil, nil
}

// FindByUsername is a stub for repository.UserRepo.FindByUsername.
func (s *StubRouterRepoUser) FindByUsername(_ context.Context, _ string) (*model.User, error) {
	return nil, nil
}

// List is a stub for repository.UserRepo.List.
func (s *StubRouterRepoUser) List(_ context.Context) ([]*model.User, error) { return nil, nil }

// Update is a stub for repository.UserRepo.Update.
func (s *StubRouterRepoUser) Update(_ context.Context, _ *model.User) error { return nil }

// Create is a stub for repository.UserRepo.Create.
func (s *StubRouterRepoUser) Create(_ context.Context, _ *model.User) error { return nil }

// FindByIDs is a stub for repository.UserRepo.FindByIDs.
func (s *StubRouterRepoUser) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.User, error) {
	return nil, nil
}

// FindByBizKey is a stub for repository.UserRepo.FindByBizKey.
func (s *StubRouterRepoUser) FindByBizKey(_ context.Context, _ int64) (*model.User, error) {
	return nil, nil
}

// FindByBizKeys is a stub for repository.UserRepo.FindByBizKeys.
func (s *StubRouterRepoUser) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.User, error) {
	return nil, nil
}

// ListFiltered is a stub for repository.UserRepo.ListFiltered.
func (s *StubRouterRepoUser) ListFiltered(_ context.Context, _ string, _, _ int) ([]*model.User, int64, error) {
	return nil, 0, nil
}

// SearchAvailable is a stub for repository.UserRepo.SearchAvailable.
func (s *StubRouterRepoUser) SearchAvailable(_ context.Context, _ int64, _ string, _ int) ([]*model.User, error) {
	return nil, nil
}

// SoftDelete is a stub for repository.UserRepo.SoftDelete.
func (s *StubRouterRepoUser) SoftDelete(_ context.Context, _ *model.User) error { return nil }

var _ repository.UserRepo = (*StubRouterRepoUser)(nil)

// StubRouterRepoSubItem is a stub implementation of repository.SubItemRepo for testing.
type StubRouterRepoSubItem struct{}

// Create is a stub for repository.SubItemRepo.Create.
func (s *StubRouterRepoSubItem) Create(_ context.Context, _ *model.SubItem) error { return nil }

// FindByID is a stub for repository.SubItemRepo.FindByID.
func (s *StubRouterRepoSubItem) FindByID(_ context.Context, _ uint) (*model.SubItem, error) {
	return nil, nil
}

// Update is a stub for repository.SubItemRepo.Update.
func (s *StubRouterRepoSubItem) Update(_ context.Context, _ *model.SubItem, _ map[string]interface{}) error {
	return nil
}

// List is a stub for repository.SubItemRepo.List.
func (s *StubRouterRepoSubItem) List(_ context.Context, _, _ int64, _ dto.SubItemFilter, _ dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	return nil, nil
}

// ListByMainItem is a stub for repository.SubItemRepo.ListByMainItem.
func (s *StubRouterRepoSubItem) ListByMainItem(_ context.Context, _ int64) ([]*model.SubItem, error) {
	return nil, nil
}

// ListByTeam is a stub for repository.SubItemRepo.ListByTeam.
func (s *StubRouterRepoSubItem) ListByTeam(_ context.Context, _ int64) ([]model.SubItem, error) {
	return nil, nil
}

// SoftDelete is a stub for repository.SubItemRepo.SoftDelete.
func (s *StubRouterRepoSubItem) SoftDelete(_ context.Context, _ uint) error { return nil }

// FindByBizKey is a stub for repository.SubItemRepo.FindByBizKey.
func (s *StubRouterRepoSubItem) FindByBizKey(_ context.Context, _ int64) (*model.SubItem, error) {
	return nil, nil
}

// NextSubCode is a stub for repository.SubItemRepo.NextSubCode.
func (s *StubRouterRepoSubItem) NextSubCode(_ context.Context, _ int64) (string, error) {
	return "", nil
}

var _ repository.SubItemRepo = (*StubRouterRepoSubItem)(nil)

// StubRouterRepoMainItem is a stub implementation of repository.MainItemRepo for testing.
type StubRouterRepoMainItem struct{}

// Create is a stub for repository.MainItemRepo.Create.
func (s *StubRouterRepoMainItem) Create(_ context.Context, _ *model.MainItem) error { return nil }

// FindByID is a stub for repository.MainItemRepo.FindByID.
func (s *StubRouterRepoMainItem) FindByID(_ context.Context, _ uint) (*model.MainItem, error) {
	return nil, nil
}

// FindByBizKey is a stub for repository.MainItemRepo.FindByBizKey.
func (s *StubRouterRepoMainItem) FindByBizKey(_ context.Context, _ int64) (*model.MainItem, error) {
	return nil, nil
}

// FindByBizKeys is a stub for repository.MainItemRepo.FindByBizKeys.
func (s *StubRouterRepoMainItem) FindByBizKeys(_ context.Context, _ []int64) (map[int64]*model.MainItem, error) {
	return nil, nil
}

// Update is a stub for repository.MainItemRepo.Update.
func (s *StubRouterRepoMainItem) Update(_ context.Context, _ *model.MainItem, _ map[string]interface{}) error {
	return nil
}

// List is a stub for repository.MainItemRepo.List.
func (s *StubRouterRepoMainItem) List(_ context.Context, _ int64, _ dto.MainItemFilter, _ dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	return nil, nil
}

// NextCode is a stub for repository.MainItemRepo.NextCode.
func (s *StubRouterRepoMainItem) NextCode(_ context.Context, _ int64) (string, error) {
	return "", nil
}

// CountByTeam is a stub for repository.MainItemRepo.CountByTeam.
func (s *StubRouterRepoMainItem) CountByTeam(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

// ListNonArchivedByTeam is a stub for repository.MainItemRepo.ListNonArchivedByTeam.
func (s *StubRouterRepoMainItem) ListNonArchivedByTeam(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}

// FindByIDs is a stub for repository.MainItemRepo.FindByIDs.
func (s *StubRouterRepoMainItem) FindByIDs(_ context.Context, _ []uint) (map[uint]*model.MainItem, error) {
	return nil, nil
}

// ListByTeamAndStatus is a stub for repository.MainItemRepo.ListByTeamAndStatus.
func (s *StubRouterRepoMainItem) ListByTeamAndStatus(_ context.Context, _ int64, _ string) ([]model.MainItem, error) {
	return nil, nil
}

// SoftDelete is a stub for repository.MainItemRepo.SoftDelete.
func (s *StubRouterRepoMainItem) SoftDelete(_ context.Context, _ uint) error { return nil }

// CascadeSoftDelete is a stub for repository.MainItemRepo.CascadeSoftDelete.
func (s *StubRouterRepoMainItem) CascadeSoftDelete(_ context.Context, _ uint, _ []uint, _ []model.StatusHistory) error {
	return nil
}

// FindByMilestoneKey is a stub for repository.MainItemRepo.FindByMilestoneKey.
func (s *StubRouterRepoMainItem) FindByMilestoneKey(_ context.Context, _ int64) ([]model.MainItem, error) {
	return nil, nil
}

// CountByMilestoneKey is a stub for repository.MainItemRepo.CountByMilestoneKey.
func (s *StubRouterRepoMainItem) CountByMilestoneKey(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

// ClearMilestoneKeyByMilestone is a stub for repository.MainItemRepo.ClearMilestoneKeyByMilestone.
func (s *StubRouterRepoMainItem) ClearMilestoneKeyByMilestone(_ context.Context, _ int64) error {
	return nil
}

// ClearMilestoneKeyByMap is a stub for repository.MainItemRepo.ClearMilestoneKeyByMap.
func (s *StubRouterRepoMainItem) ClearMilestoneKeyByMap(_ context.Context, _ []int64) error {
	return nil
}

var _ repository.MainItemRepo = (*StubRouterRepoMainItem)(nil)

// StubReportSvc is a stub ReportService for test wiring.
type StubReportSvc struct{}

// Preview is a stub for service.ReportService.Preview.
func (s *StubReportSvc) Preview(_ context.Context, _ int64, _ time.Time) (*dto.ReportPreview, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// ExportMarkdown is a stub for service.ReportService.ExportMarkdown.
func (s *StubReportSvc) ExportMarkdown(_ context.Context, _ int64, _ time.Time) ([]byte, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.ReportService = (*StubReportSvc)(nil)

// StubViewSvc is a stub ViewService for test wiring.
type StubViewSvc struct{}

// WeeklyComparison is a stub for service.ViewService.WeeklyComparison.
func (s *StubViewSvc) WeeklyComparison(_ context.Context, _ int64, _ time.Time) (*dto.WeeklyViewResponse, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// GanttView is a stub for service.ViewService.GanttView.
func (s *StubViewSvc) GanttView(_ context.Context, _ int64, _ dto.GanttFilter) (*dto.GanttResult, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// TableView is a stub for service.ViewService.TableView.
func (s *StubViewSvc) TableView(_ context.Context, _ int64, _ dto.TableFilter, _ dto.Pagination) (*dto.PageResult[dto.TableRow], error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// TableExportCSV is a stub for service.ViewService.TableExportCSV.
func (s *StubViewSvc) TableExportCSV(_ context.Context, _ int64, _ dto.TableFilter) ([]byte, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.ViewService = (*StubViewSvc)(nil)

// StubProgressSvc is a stub ProgressService for test wiring.
type StubProgressSvc struct{}

// Append is a stub for service.ProgressService.Append.
func (s *StubProgressSvc) Append(_ context.Context, _, _ int64, _ uint, _ float64, _, _, _ string, _ bool) (*model.ProgressRecord, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// CorrectCompletion is a stub for service.ProgressService.CorrectCompletion.
func (s *StubProgressSvc) CorrectCompletion(_ context.Context, _ int64, _ uint, _ float64) error {
	return fmt.Errorf("stub: not implemented")
}

// List is a stub for service.ProgressService.List.
func (s *StubProgressSvc) List(_ context.Context, _ int64, _ uint) ([]model.ProgressRecord, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

// GetByBizKey is a stub for service.ProgressService.GetByBizKey.
func (s *StubProgressSvc) GetByBizKey(_ context.Context, _ int64) (*model.ProgressRecord, error) {
	return nil, fmt.Errorf("stub: not implemented")
}

var _ service.ProgressService = (*StubProgressSvc)(nil)
