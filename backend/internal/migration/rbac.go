package migration

import (
	"fmt"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/permissions"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

const rbacMigrationVersion = "rbac_001"

// MigrateToRBAC runs the RBAC data migration in a single database transaction.
// It creates new tables, seeds preset roles, migrates team_members.role strings
// to role_id,.
// It is idempotent: re-running produces no side effects (tracked via schema_migrations).
func MigrateToRBAC(db *gorm.DB) error {
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	// Check if already applied
	var count int64
	if err := db.Raw("SELECT count(*) FROM schema_migrations WHERE version = ?", rbacMigrationVersion).Scan(&count).Error; err != nil {
		return fmt.Errorf("check rbac migration status: %w", err)
	}
	if count > 0 {
		return nil // already applied
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := runRBACMigration(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Mark as applied
	if err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", rbacMigrationVersion).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("mark rbac migration: %w", err)
	}

	return tx.Commit().Error
}

func runRBACMigration(tx *gorm.DB) error {
	// 1. Create new tables (roles, role_permissions)
	if err := tx.AutoMigrate(&model.Role{}, &model.RolePermission{}); err != nil {
		return fmt.Errorf("create rbac tables: %w", err)
	}

	// 2. Seed preset roles
	if err := seedPresetRoles(tx); err != nil {
		return fmt.Errorf("seed preset roles: %w", err)
	}

	// 3. Rebuild team_members table: migrate role string -> role_id and drop role column
	// Must be done in one step (SQLite rebuild) because the existing unique index on role_id
	// would prevent updating multiple members to the same role.
	roleMap, err := getRoleIDMap(tx)
	if err != nil {
		return fmt.Errorf("build role map: %w", err)
	}
	if err := rebuildTeamMembersTable(tx, roleMap); err != nil {
		return fmt.Errorf("rebuild team_members: %w", err)
	}

	// 4. Users table: can_create_team column removed (handled by RBAC team:create permission).

	return nil
}

func seedPresetRoles(tx *gorm.DB) error {
	// Seed superadmin (id=1, no permission codes)
	if err := seedRole(tx, "superadmin", "系统超级管理员，绕过所有权限检查", true, nil); err != nil {
		return err
	}

	// Seed pm (id=2, 22 codes)
	pmCodes := []string{
		"team:create", "team:read", "team:update", "team:delete",
		"team:invite", "team:remove", "team:transfer",
		"main_item:create", "main_item:read", "main_item:update", "main_item:archive",
		"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:assign", "sub_item:change_status",
		"progress:create", "progress:read", "progress:update",
		"item_pool:submit", "item_pool:review",
		"view:weekly", "view:gantt", "view:table",
		"report:export",
		"user:read",
	}
	if err := seedRole(tx, "pm", "项目经理，团队管理权限", true, pmCodes); err != nil {
		return err
	}

	// Seed member (id=3, 11 codes)
	memberCodes := []string{
		"main_item:read",
		"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:change_status",
		"progress:create", "progress:read",
		"item_pool:submit",
		"view:weekly", "view:table",
		"report:export",
	}
	if err := seedRole(tx, "member", "团队成员，基础操作权限", true, memberCodes); err != nil {
		return err
	}

	return nil
}

func seedRole(tx *gorm.DB, name, description string, isPreset bool, codes []string) error {
	// Check if role already exists
	var existing model.Role
	result := tx.Where("name = ?", name).First(&existing)
	if result.RowsAffected > 0 {
		// Backfill bizKey if it was left at default 0 from a pre-migration database
		if existing.BizKey == 0 {
			if err := tx.Model(&existing).Update("biz_key", snowflake.Generate()).Error; err != nil {
				return fmt.Errorf("backfill biz_key for role %s: %w", name, err)
			}
		}
		return nil // already seeded
	}

	role := model.Role{
		BaseModel:   model.BaseModel{BizKey: snowflake.Generate()},
		Name:        name,
		Description: description,
		IsPreset:    isPreset,
	}
	if err := tx.Create(&role).Error; err != nil {
		return fmt.Errorf("create role %s: %w", name, err)
	}

	// Insert permission codes
	for _, code := range codes {
		rp := model.RolePermission{
			RoleID:         role.ID,
			PermissionCode: code,
		}
		if err := tx.Create(&rp).Error; err != nil {
			return fmt.Errorf("create permission %s for role %s: %w", code, name, err)
		}
	}

	return nil
}

func getRoleIDMap(tx *gorm.DB) (map[string]uint, error) {
	var roles []model.Role
	if err := tx.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("fetch roles: %w", err)
	}
	m := make(map[string]uint, len(roles))
	for _, r := range roles {
		m[r.Name] = r.ID
	}
	return m, nil
}

func rebuildTeamMembersTable(tx *gorm.DB, roleMap map[string]uint) error {
	// Check if legacy team_members table exists (pre-rename schema)
	legacyExists := tableExists(tx, "team_members")
	newExists := tableExists(tx, "pmw_team_members")

	if !legacyExists && !newExists {
		// Fresh install: create the new table directly
		return tx.Exec(`
			CREATE TABLE pmw_team_members (
				id              INTEGER PRIMARY KEY AUTOINCREMENT,
				biz_key         INTEGER NOT NULL DEFAULT 0,
				create_time     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				db_update_time  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				deleted_flag    INTEGER NOT NULL DEFAULT 0,
				deleted_time    DATETIME NOT NULL DEFAULT '1970-01-01 08:00:00',
				team_key        INTEGER NOT NULL,
				user_key        INTEGER NOT NULL,
				role_key        INTEGER,
				joined_at       DATETIME NOT NULL,
				UNIQUE(team_key, user_key)
			)
		`).Error
	}

	if newExists {
		// Already migrated to new schema — nothing to do
		return nil
	}

	// Legacy table exists: rebuild as pmw_team_members with new column names.
	memberRoleID := roleMap["member"]
	pmRoleID := roleMap["pm"]

	if err := tx.Exec(`
		CREATE TABLE pmw_team_members (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			biz_key         INTEGER NOT NULL DEFAULT 0,
			create_time     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			db_update_time  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_flag    INTEGER NOT NULL DEFAULT 0,
			deleted_time    DATETIME NOT NULL DEFAULT '1970-01-01 08:00:00',
			team_key        INTEGER NOT NULL,
			user_key        INTEGER NOT NULL,
			role_key        INTEGER,
			joined_at       DATETIME NOT NULL,
			UNIQUE(team_key, user_key)
		)
	`).Error; err != nil {
		return fmt.Errorf("create pmw_team_members: %w", err)
	}

	if columnExists(tx, "team_members", "role") {
		if err := tx.Exec(`
			INSERT INTO pmw_team_members (id, team_key, user_key, role_key, joined_at)
			SELECT id, team_id, user_id,
				CASE
					WHEN role = 'pm' THEN ?
					WHEN role = 'member' THEN ?
					ELSE ?
				END,
				joined_at
			FROM team_members
		`, pmRoleID, memberRoleID, memberRoleID).Error; err != nil {
			return fmt.Errorf("copy team_members data: %w", err)
		}
	} else {
		if err := tx.Exec(`
			INSERT INTO pmw_team_members (id, team_key, user_key, role_key, joined_at)
			SELECT id, team_id, user_id, ?, joined_at
			FROM team_members
		`, memberRoleID).Error; err != nil {
			return fmt.Errorf("copy team_members data: %w", err)
		}
	}

	if err := tx.Exec("DROP TABLE team_members").Error; err != nil {
		return fmt.Errorf("drop old team_members: %w", err)
	}

	return nil
}


// CountPermissionsForRole returns the number of permission codes bound to a role.
// Exported for testing convenience.
func CountPermissionsForRole(db *gorm.DB, roleID uint) (int64, error) {
	var count int64
	err := db.Model(&model.RolePermission{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

// HasColumn checks if a table has a specific column.
// Exported for testing convenience.
func HasColumn(db *gorm.DB, table, column string) bool {
	var count int64
	db.Raw("SELECT count(*) FROM pragma_table_info(?) WHERE name = ?", table, column).Scan(&count)
	return count > 0
}

// columnExists checks if a column exists in a SQLite table.
func columnExists(db *gorm.DB, table, column string) bool {
	var count int64
	db.Raw("SELECT count(*) FROM pragma_table_info(?) WHERE name = ?", table, column).Scan(&count)
	return count > 0
}

// tableExists checks if a table exists in the SQLite database.
func tableExists(db *gorm.DB, table string) bool {
	var count int64
	db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name = ?", table).Scan(&count)
	return count > 0
}

// ensureSchemaMigrationsTable is the package-level version used by rbac.go.
func ensureSchemaMigrationsTable(db *gorm.DB) error {
	sql := `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version  VARCHAR(255) PRIMARY KEY,
    applied  DATETIME NOT NULL DEFAULT (datetime('now'))
);
`
	return db.Exec(sql).Error
}

// VerifyPresetRoleCodes checks that pm and member preset roles have the expected
// permission codes from the permissions registry.
func VerifyPresetRoleCodes(db *gorm.DB) error {
	expected := map[string][]string{
		"pm": {
			"team:create", "team:read", "team:update", "team:delete",
			"team:invite", "team:remove", "team:transfer",
			"main_item:create", "main_item:read", "main_item:update", "main_item:archive",
			"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:assign", "sub_item:change_status",
			"progress:create", "progress:read", "progress:update",
			"item_pool:submit", "item_pool:review",
			"view:weekly", "view:gantt", "view:table",
			"report:export",
			"user:read",
		},
		"member": {
			"main_item:read",
			"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:change_status",
			"progress:create", "progress:read",
			"item_pool:submit",
			"view:weekly", "view:table",
			"report:export",
		},
	}

	allCodes := permissions.AllCodes()

	for roleName, codes := range expected {
		var role model.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			return fmt.Errorf("find role %s: %w", roleName, err)
		}

		var rolePerms []model.RolePermission
		if err := db.Where("role_id = ?", role.ID).Find(&rolePerms).Error; err != nil {
			return fmt.Errorf("find permissions for role %s: %w", roleName, err)
		}

		actual := make(map[string]bool, len(rolePerms))
		for _, rp := range rolePerms {
			if !allCodes[rp.PermissionCode] {
				return fmt.Errorf("invalid permission code %s in role %s", rp.PermissionCode, roleName)
			}
			actual[rp.PermissionCode] = true
		}

		for _, code := range codes {
			if !actual[code] {
				return fmt.Errorf("missing permission %s in role %s", code, roleName)
			}
		}

		if len(actual) != len(codes) {
			return fmt.Errorf("role %s has %d permissions, expected %d", roleName, len(actual), len(codes))
		}
	}

	return nil
}
