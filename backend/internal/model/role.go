package model

import "time"

// Role represents a named set of permissions that can be assigned to team members.
type Role struct {
	BaseModel
	Name        string `gorm:"column:role_name;type:varchar(50);uniqueIndex:uk_roles_name;not null" json:"name"`
	Description string `gorm:"column:role_desc;type:varchar(200);not null;default:''" json:"description"`
	IsPreset    bool   `gorm:"not null;default:false" json:"isPreset"`
}

func (Role) TableName() string {
	return "pmw_roles"
}

// RolePermission binds a permission code to a role.
type RolePermission struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	DeletedFlag    int       `gorm:"not null;default:0;uniqueIndex:uk_role_permission" json:"-"`
	DeletedTime    time.Time `gorm:"not null;default:'1970-01-01 08:00:00';uniqueIndex:uk_role_permission" json:"-"`
	RoleKey        int64     `gorm:"not null;uniqueIndex:uk_role_permission" json:"roleKey"`
	PermissionCode string    `gorm:"type:varchar(50);not null;uniqueIndex:uk_role_permission" json:"permissionCode"`
}

func (RolePermission) TableName() string {
	return "pmw_role_permissions"
}
