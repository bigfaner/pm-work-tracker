package service

import (
	"context"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	apperrors "pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/permissions"
	"pm-work-tracker/backend/internal/repository"
)

// Role-level application errors.
var (
	ErrRoleNotFound          = &apperrors.AppError{Code: "ERR_ROLE_NOT_FOUND", Status: 404, Message: "角色不存在"}
	ErrRoleNameExists        = &apperrors.AppError{Code: "ERR_ROLE_NAME_EXISTS", Status: 409, Message: "角色名称已存在"}
	ErrRoleInUse             = &apperrors.AppError{Code: "ERR_ROLE_IN_USE", Status: 422, Message: "角色正在被使用，无法删除"}
	ErrPresetRoleImmutable   = &apperrors.AppError{Code: "ERR_PRESET_ROLE_IMMUTABLE", Status: 403, Message: "预置角色不可编辑或删除"}
	ErrInvalidPermissionCode = &apperrors.AppError{Code: "ERR_INVALID_PERMISSION_CODE", Status: 400, Message: "无效的权限码"}
	ErrValidation            = &apperrors.AppError{Code: "ERR_VALIDATION", Status: 422, Message: "请求验证失败"}
)

// RoleListItem is the response shape for a role in list view.
type RoleListItem struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	IsPreset        bool   `json:"isPreset"`
	PermissionCount int    `json:"permissionCount"`
	MemberCount     int64  `json:"memberCount"`
	CreatedAt       string `json:"createdAt"`
}

// RoleDetail is the response shape for a single role with full permission list.
type RoleDetail struct {
	ID          uint               `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	IsPreset    bool               `json:"isPreset"`
	Permissions []PermissionItem   `json:"permissions"`
	MemberCount int64              `json:"memberCount"`
	CreatedAt   string             `json:"createdAt"`
}

// PermissionItem describes a single permission code in a role.
type PermissionItem struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// UserPermissions is the response shape for a user's permission map.
type UserPermissions struct {
	IsSuperAdmin    bool              `json:"isSuperAdmin"`
	TeamPermissions map[uint][]string `json:"teamPermissions"`
}

// RoleService defines business logic for role management.
type RoleService interface {
	ListRoles(ctx context.Context) ([]RoleListItem, error)
	GetRole(ctx context.Context, roleID uint) (*RoleDetail, error)
	CreateRole(ctx context.Context, req dto.CreateRoleReq) (*RoleListItem, error)
	UpdateRole(ctx context.Context, roleID uint, req dto.UpdateRoleReq) (*RoleDetail, error)
	DeleteRole(ctx context.Context, roleID uint) error
	ListPermissionCodes(ctx context.Context) []permissions.ResourcePermissions
	GetUserPermissions(ctx context.Context, userID uint) (*UserPermissions, error)
}

type roleService struct {
	roleRepo repository.RoleRepo
	userRepo repository.UserRepo
}

// NewRoleService creates a new RoleService.
func NewRoleService(roleRepo repository.RoleRepo, userRepo repository.UserRepo) RoleService {
	return &roleService{roleRepo: roleRepo, userRepo: userRepo}
}

func (s *roleService) ListRoles(ctx context.Context) ([]RoleListItem, error) {
	roles, err := s.roleRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]RoleListItem, 0, len(roles))
	for _, r := range roles {
		var permCount int
		if r.Name == "superadmin" {
			permCount = permissions.TotalCodeCount()
		} else {
			var err error
			permCount, err = s.permCount(ctx, r.ID)
			if err != nil {
				return nil, err
			}
		}
		memberCount, err := s.roleRepo.CountMembersByRoleID(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, RoleListItem{
			ID:              r.ID,
			Name:            r.Name,
			Description:     r.Description,
			IsPreset:        r.IsPreset,
			PermissionCount: permCount,
			MemberCount:     memberCount,
			CreatedAt:       r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return items, nil
}

func (s *roleService) GetRole(ctx context.Context, roleID uint) (*RoleDetail, error) {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, ErrRoleNotFound)
	}

	var codes []string
	if role.Name == "superadmin" {
		codes = allCodes()
	} else {
		var err error
		codes, err = s.roleRepo.ListPermissions(ctx, roleID)
		if err != nil {
			return nil, err
		}
	}

	memberCount, err := s.roleRepo.CountMembersByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return &RoleDetail{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsPreset:    role.IsPreset,
		Permissions: codesToItems(codes),
		MemberCount: memberCount,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *roleService) CreateRole(ctx context.Context, req dto.CreateRoleReq) (*RoleListItem, error) {
	if err := validateRoleName(req.Name); err != nil {
		return nil, err
	}
	if len(req.PermissionCodes) == 0 {
		return nil, ErrValidation
	}
	if err := validatePermissionCodes(req.PermissionCodes); err != nil {
		return nil, err
	}

	// Check name uniqueness
	if _, err := s.roleRepo.FindByName(ctx, req.Name); err == nil {
		return nil, ErrRoleNameExists
	}

	role := &model.Role{
		Name:        req.Name,
		Description: req.Description,
		IsPreset:    false,
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	if err := s.roleRepo.SetPermissions(ctx, role.ID, req.PermissionCodes); err != nil {
		return nil, err
	}

	return &RoleListItem{
		ID:              role.ID,
		Name:            role.Name,
		Description:     role.Description,
		IsPreset:        false,
		PermissionCount: len(req.PermissionCodes),
		MemberCount:     0,
		CreatedAt:       role.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *roleService) UpdateRole(ctx context.Context, roleID uint, req dto.UpdateRoleReq) (*RoleDetail, error) {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, apperrors.MapNotFound(err, ErrRoleNotFound)
	}

	// Preset role checks
	if role.IsPreset {
		if role.Name == "superadmin" {
			return nil, ErrPresetRoleImmutable
		}
		// pm/member: can change description + permissions, but not name
		if req.Name != nil {
			return nil, ErrPresetRoleImmutable
		}
	}

	// Validate new name if provided
	if req.Name != nil {
		if err := validateRoleName(*req.Name); err != nil {
			return nil, err
		}
		// Check name uniqueness (skip if name unchanged)
		if *req.Name != role.Name {
			if _, err := s.roleRepo.FindByName(ctx, *req.Name); err == nil {
				return nil, ErrRoleNameExists
			}
		}
	}

	// Validate permission codes if provided
	if req.PermissionCodes != nil {
		if len(req.PermissionCodes) == 0 {
			return nil, ErrValidation
		}
		if err := validatePermissionCodes(req.PermissionCodes); err != nil {
			return nil, err
		}
	}

	// Apply updates
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	// Update permissions if provided
	if req.PermissionCodes != nil {
		if err := s.roleRepo.SetPermissions(ctx, roleID, req.PermissionCodes); err != nil {
			return nil, err
		}
	}

	// Reload permissions for response
	codes, err := s.roleRepo.ListPermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	memberCount, err := s.roleRepo.CountMembersByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return &RoleDetail{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsPreset:    role.IsPreset,
		Permissions: codesToItems(codes),
		MemberCount: memberCount,
		CreatedAt:   role.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *roleService) DeleteRole(ctx context.Context, roleID uint) error {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return apperrors.MapNotFound(err, ErrRoleNotFound)
	}

	if role.IsPreset {
		return ErrPresetRoleImmutable
	}

	count, err := s.roleRepo.CountMembersByRoleID(ctx, roleID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrRoleInUse
	}

	return s.roleRepo.Delete(ctx, roleID)
}

func (s *roleService) ListPermissionCodes(_ context.Context) []permissions.ResourcePermissions {
	return permissions.Registry
}

func (s *roleService) GetUserPermissions(ctx context.Context, userID uint) (*UserPermissions, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	teamPerms, err := s.roleRepo.GetUserTeamPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	if teamPerms == nil {
		teamPerms = map[uint][]string{}
	}

	return &UserPermissions{
		IsSuperAdmin:    user.IsSuperAdmin,
		TeamPermissions: teamPerms,
	}, nil
}

// --- internal helpers ---

func (s *roleService) permCount(ctx context.Context, roleID uint) (int, error) {
	codes, err := s.roleRepo.ListPermissions(ctx, roleID)
	if err != nil {
		return 0, err
	}
	return len(codes), nil
}

func validateRoleName(name string) error {
	if len(name) < 2 || len(name) > 50 {
		return ErrValidation
	}
	return nil
}

func validatePermissionCodes(codes []string) error {
	for _, code := range codes {
		if !permissions.ValidateCode(code) {
			return ErrInvalidPermissionCode
		}
	}
	return nil
}

func codesToItems(codes []string) []PermissionItem {
	items := make([]PermissionItem, 0, len(codes))
	descMap := buildCodeDescMap()
	for _, code := range codes {
		desc := ""
		if d, ok := descMap[code]; ok {
			desc = d
		}
		items = append(items, PermissionItem{Code: code, Description: desc})
	}
	return items
}

func allCodes() []string {
	codes := make([]string, 0, permissions.TotalCodeCount())
	for _, rp := range permissions.Registry {
		for _, p := range rp.Permissions {
			codes = append(codes, p.Code)
		}
	}
	return codes
}

func buildCodeDescMap() map[string]string {
	m := make(map[string]string)
	for _, rp := range permissions.Registry {
		for _, p := range rp.Permissions {
			m[p.Code] = p.Description
		}
	}
	return m
}

