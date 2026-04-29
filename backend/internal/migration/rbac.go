package migration

import (
	"fmt"

	"gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/permissions"
	"pm-work-tracker/backend/internal/pkg/snowflake"
)

const rbacMigrationVersion = "rbac_001"
const permissionGranularityVersion = "permission_granularity_001"

// MigrateToRBAC runs the RBAC data migration in a single database transaction.
// It creates new tables, seeds preset roles, migrates team_members.role strings
// to role_key,.
// It is idempotent: re-running produces no side effects (tracked via schema_migrations).
// When autoSchema is false, DDL and schema_migrations tracking are skipped.
// The underlying DML steps (seed roles, rebuild team_members) are idempotent on their own.
func MigrateToRBAC(db *gorm.DB, autoSchema bool) error {
	if autoSchema {
		if err := ensureSchemaMigrationsTable(db); err != nil {
			return fmt.Errorf("ensure schema_migrations: %w", err)
		}

		var count int64
		if err := db.Raw("SELECT count(*) FROM schema_migrations WHERE version = ?", rbacMigrationVersion).Scan(&count).Error; err != nil {
			return fmt.Errorf("check rbac migration status: %w", err)
		}
		if count > 0 {
			return nil // already applied
		}
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := runRBACMigration(tx, autoSchema); err != nil {
		tx.Rollback()
		return err
	}

	if autoSchema {
		if err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", rbacMigrationVersion).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("mark rbac migration: %w", err)
		}
	}

	return tx.Commit().Error
}

// SyncPresetRoles ensures preset role permission codes are up to date.
// Unlike MigrateToRBAC, this runs on every startup to pick up any new codes
// added to the seed list after the initial migration.
func SyncPresetRoles(db *gorm.DB) error {
	return seedPresetRoles(db)
}

func runRBACMigration(tx *gorm.DB, autoSchema bool) error {
	if err := createRBACTables(tx, autoSchema); err != nil {
		return fmt.Errorf("create rbac tables: %w", err)
	}

	// 2. Seed preset roles
	if err := seedPresetRoles(tx); err != nil {
		return fmt.Errorf("seed preset roles: %w", err)
	}

	// 3. Rebuild team_members table: migrate role string -> role_key and drop role column
	// Must be done in one step (SQLite rebuild) because the existing unique index on role_key
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

func createRBACTables(tx *gorm.DB, autoSchema bool) error {
	if !autoSchema {
		return nil
	}
	for _, stmt := range rbacTableDDL(tx) {
		if err := tx.Exec(stmt).Error; err != nil {
			if tableExists(tx, "pmw_roles") && tableExists(tx, "pmw_role_permissions") {
				return nil
			}
			return fmt.Errorf("create RBAC tables: %w (hint: run schema SQL as root, or set auto_schema: true)", err)
		}
	}
	return nil
}

func rbacTableDDL(tx *gorm.DB) []string {
	if isMySQL(tx) {
		return []string{
			`CREATE TABLE IF NOT EXISTS pmw_roles (
	    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	    biz_key         BIGINT          NOT NULL,
	    create_time     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
	    db_update_time  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	    deleted_flag    TINYINT(1)      NOT NULL DEFAULT 0,
	    deleted_time    DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
	    role_name       VARCHAR(50)     NOT NULL,
	    role_desc       VARCHAR(200)    NOT NULL DEFAULT '',
	    is_preset       TINYINT(1)      NOT NULL DEFAULT 0,
	    PRIMARY KEY (id),
	    UNIQUE KEY uk_roles_name (role_name, deleted_flag, deleted_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
			`CREATE TABLE IF NOT EXISTS pmw_role_permissions (
	    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	    deleted_flag     TINYINT(1)      NOT NULL DEFAULT 0,
	    deleted_time     DATETIME        NOT NULL DEFAULT '1970-01-01 08:00:00',
	    role_key         BIGINT          NOT NULL,
	    permission_code  VARCHAR(50)     NOT NULL,
	    PRIMARY KEY (id),
	    UNIQUE KEY uk_role_permission (role_key, permission_code, deleted_flag, deleted_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		}
	}
	return []string{
		`CREATE TABLE IF NOT EXISTS pmw_roles (
	    id              INTEGER PRIMARY KEY AUTOINCREMENT,
	    biz_key         INTEGER       NOT NULL,
	    create_time     DATETIME      NOT NULL DEFAULT (datetime('now')),
	    db_update_time  DATETIME      NOT NULL DEFAULT (datetime('now')),
	    deleted_flag    INTEGER       NOT NULL DEFAULT 0,
	    deleted_time    DATETIME      NOT NULL DEFAULT '1970-01-01 08:00:00',
	    role_name       VARCHAR(50)   NOT NULL,
	    role_desc       VARCHAR(200)  NOT NULL DEFAULT '',
	    is_preset       INTEGER       NOT NULL DEFAULT 0
)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_roles_name ON pmw_roles(role_name, deleted_flag, deleted_time)`,
		`CREATE TABLE IF NOT EXISTS pmw_role_permissions (
	    id               INTEGER PRIMARY KEY AUTOINCREMENT,
	    deleted_flag     INTEGER      NOT NULL DEFAULT 0,
	    deleted_time     DATETIME     NOT NULL DEFAULT '1970-01-01 08:00:00',
	    role_key         INTEGER      NOT NULL,
	    permission_code  VARCHAR(50)  NOT NULL,
	    UNIQUE(role_key, permission_code, deleted_flag, deleted_time)
)`,
	}
}

func seedPresetRoles(tx *gorm.DB) error {
	// Seed superadmin (id=1, no permission codes)
	if err := seedRole(tx, "superadmin", "系统超级管理员，绕过所有权限检查", true, nil); err != nil {
		return err
	}

	// Seed pm (id=2, 32 codes)
	pmCodes := []string{
		"team:create", "team:read", "team:update", "team:delete",
		"team:invite", "team:remove", "team:transfer",
		"main_item:create", "main_item:read", "main_item:update", "main_item:archive",
		"sub_item:create", "sub_item:read", "sub_item:update", "sub_item:assign", "sub_item:change_status",
		"progress:create", "progress:read", "progress:update",
		"item_pool:submit", "item_pool:review",
		"view:weekly", "view:gantt", "view:table",
		"report:export",
		"user:list", "user:read", "user:assign_role",
		"role:read", "role:create", "role:update", "role:delete",
	}
	if err := seedRole(tx, "pm", "项目经理，团队管理权限", true, pmCodes); err != nil {
		return err
	}

	// Seed member (id=3, 14 codes)
	memberCodes := []string{
		"team:read",
		"main_item:create", "main_item:read", "main_item:update",
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
	result := tx.Where("role_name = ?", name).First(&existing)
	if result.RowsAffected > 0 {
		// Backfill bizKey if it was left at default 0 from a pre-migration database
		if existing.BizKey == 0 {
			if err := tx.Model(&existing).Update("biz_key", snowflake.Generate()).Error; err != nil {
				return fmt.Errorf("backfill biz_key for role %s: %w", name, err)
			}
		}
		// Sync permissions: add any missing codes
		var existingPerms []model.RolePermission
		if err := tx.Where("role_key = ? AND deleted_flag = 0", existing.BizKey).Find(&existingPerms).Error; err != nil {
			return fmt.Errorf("fetch permissions for role %s: %w", name, err)
		}
		existingSet := make(map[string]bool, len(existingPerms))
		for _, p := range existingPerms {
			existingSet[p.PermissionCode] = true
		}
		for _, code := range codes {
			if !existingSet[code] {
				rp := model.RolePermission{RoleKey: existing.BizKey, PermissionCode: code}
				if err := tx.Create(&rp).Error; err != nil {
					return fmt.Errorf("add permission %s to role %s: %w", code, name, err)
				}
			}
		}
		return nil
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
			RoleKey:        role.BizKey,
			PermissionCode: code,
		}
		if err := tx.Create(&rp).Error; err != nil {
			return fmt.Errorf("create permission %s for role %s: %w", code, name, err)
		}
	}

	return nil
}

func getRoleIDMap(tx *gorm.DB) (map[string]int64, error) {
	var roles []model.Role
	if err := tx.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("fetch roles: %w", err)
	}
	m := make(map[string]int64, len(roles))
	for _, r := range roles {
		m[r.Name] = r.BizKey
	}
	return m, nil
}

func rebuildTeamMembersTable(tx *gorm.DB, roleMap map[string]int64) error {
	// Check if legacy team_members table exists (pre-rename schema)
	legacyExists := tableExists(tx, "team_members")
	newExists := tableExists(tx, "pmw_team_members")

	if !legacyExists && !newExists {
		// Fresh install: create the new table directly
		for _, stmt := range teamMembersDDL(tx) {
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("create pmw_team_members: %w", err)
			}
		}
		return nil
	}

	if newExists {
		// Already migrated to new schema — nothing to do
		return nil
	}

	// Legacy table exists: rebuild as pmw_team_members with new column names.
	memberRoleID := roleMap["member"]
	pmRoleID := roleMap["pm"]

	for _, stmt := range teamMembersDDL(tx) {
		if err := tx.Exec(stmt).Error; err != nil {
			return fmt.Errorf("create pmw_team_members: %w", err)
		}
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

// teamMembersDDL returns DDL statements to create the pmw_team_members table.
// Returns 1 statement for SQLite (inline UNIQUE) or 2 statements for MySQL
// (CREATE TABLE + CREATE UNIQUE INDEX).
func teamMembersDDL(db *gorm.DB) []string {
	if isMySQL(db) {
		return teamMembersDDLMySQL()
	}
	return []string{
		`CREATE TABLE pmw_team_members (
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
		)`,
	}
}

// teamMembersDDLMySQL returns MySQL-compatible DDL for pmw_team_members.
func teamMembersDDLMySQL() []string {
	return []string{
		`CREATE TABLE pmw_team_members (
			id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			biz_key         BIGINT NOT NULL DEFAULT 0,
			create_time     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			db_update_time  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_flag    TINYINT(1) NOT NULL DEFAULT 0,
			deleted_time    DATETIME NOT NULL DEFAULT '1970-01-01 08:00:00',
			team_key        BIGINT NOT NULL,
			user_key        BIGINT NOT NULL,
			role_key        BIGINT,
			joined_at       DATETIME NOT NULL,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE UNIQUE INDEX uk_team_members_team_user ON pmw_team_members(team_key, user_key)`,
	}
}

// CountPermissionsForRole returns the number of permission codes bound to a role.
// Exported for testing convenience.
func CountPermissionsForRole(db *gorm.DB, roleKey int64) (int64, error) {
	var count int64
	err := db.Model(&model.RolePermission{}).Where("role_key = ?", roleKey).Count(&count).Error
	return count, err
}

// HasColumn checks if a table has a specific column.
// Exported for testing convenience.
func HasColumn(db *gorm.DB, table, column string) bool {
	return columnExists(db, table, column)
}

// isMySQL returns true if the underlying database is MySQL.
func isMySQL(db *gorm.DB) bool {
	return db.Dialector.Name() == "mysql"
}

// columnExists checks if a column exists in a table.
func columnExists(db *gorm.DB, table, column string) bool {
	var count int64
	if isMySQL(db) {
		db.Raw("SELECT count(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?", table, column).Scan(&count)
	} else {
		db.Raw("SELECT count(*) FROM pragma_table_info(?) WHERE name = ?", table, column).Scan(&count)
	}
	return count > 0
}

// tableExists checks if a table exists in the database.
func tableExists(db *gorm.DB, table string) bool {
	var count int64
	if isMySQL(db) {
		db.Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", table).Scan(&count)
	} else {
		db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name = ?", table).Scan(&count)
	}
	return count > 0
}

// ensureSchemaMigrationsTable creates the schema_migrations table if it does not exist.
// Tolerates missing CREATE privilege when the table already exists.
func ensureSchemaMigrationsTable(db *gorm.DB) error {
	var ddl string
	if isMySQL(db) {
		ddl = `CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(255) PRIMARY KEY, applied DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`
	} else {
		ddl = `CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(255) PRIMARY KEY, applied DATETIME NOT NULL DEFAULT (datetime('now')))`
	}
	if err := db.Exec(ddl).Error; err != nil {
		if tableExists(db, "schema_migrations") {
			return nil
		}
		return fmt.Errorf("create schema_migrations table: %w (hint: run schema SQL as root, or set auto_schema: true)", err)
	}
	return nil
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
			"user:list", "user:read", "user:assign_role",
			"role:read", "role:create", "role:update", "role:delete",
		},
		"member": {
			"team:read",
			"main_item:create", "main_item:read", "main_item:update",
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
		if err := db.Where("role_name = ?", roleName).First(&role).Error; err != nil {
			return fmt.Errorf("find role %s: %w", roleName, err)
		}

		var rolePerms []model.RolePermission
		if err := db.Where("role_key = ?", role.BizKey).Find(&rolePerms).Error; err != nil {
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

// MigratePermissionGranularity migrates old permission codes to the new
// granular permission model. It is idempotent: re-running produces no side
// effects (tracked via schema_migrations).
//
// Migration logic:
//  1. user:manage_role → role:create + role:update + role:delete (old code deleted)
//  2. Roles with old user:read → add user:list (old user:read preserved)
//  3. Mark schema_migrations version = permission_granularity_001
func MigratePermissionGranularity(db *gorm.DB) error {
	if err := ensureSchemaMigrationsTable(db); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	var count int64
	if err := db.Raw("SELECT count(*) FROM schema_migrations WHERE version = ?", permissionGranularityVersion).Scan(&count).Error; err != nil {
		return fmt.Errorf("check permission granularity migration status: %w", err)
	}
	if count > 0 {
		return nil // already applied
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := runPermissionGranularityMigration(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", permissionGranularityVersion).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("mark permission granularity migration: %w", err)
	}

	return tx.Commit().Error
}

func runPermissionGranularityMigration(tx *gorm.DB) error {
	// 1. Find all role_keys that have user:manage_role
	var manageRoleKeys []int64
	if err := tx.Model(&model.RolePermission{}).
		Where("permission_code = ?", "user:manage_role").
		Pluck("role_key", &manageRoleKeys).Error; err != nil {
		return fmt.Errorf("find roles with user:manage_role: %w", err)
	}

	// 2. For each role with user:manage_role, insert role:create, role:update, role:delete
	//    and delete user:manage_role
	newCodes := []string{"role:create", "role:update", "role:delete"}
	for _, rk := range manageRoleKeys {
		for _, code := range newCodes {
			// INSERT IGNORE pattern: check if already exists
			var exists int64
			tx.Model(&model.RolePermission{}).
				Where("role_key = ? AND permission_code = ?", rk, code).
				Count(&exists)
			if exists == 0 {
				rp := model.RolePermission{RoleKey: rk, PermissionCode: code}
				if err := tx.Create(&rp).Error; err != nil {
					return fmt.Errorf("insert %s for role_key %d: %w", code, rk, err)
				}
			}
		}
		// Delete the old user:manage_role code
		if err := tx.Where("role_key = ? AND permission_code = ?", rk, "user:manage_role").
			Delete(&model.RolePermission{}).Error; err != nil {
			return fmt.Errorf("delete user:manage_role for role_key %d: %w", rk, err)
		}
	}

	// 3. Find all role_keys that have user:read and add user:list
	var userReadKeys []int64
	if err := tx.Model(&model.RolePermission{}).
		Where("permission_code = ?", "user:read").
		Pluck("role_key", &userReadKeys).Error; err != nil {
		return fmt.Errorf("find roles with user:read: %w", err)
	}

	for _, rk := range userReadKeys {
		var exists int64
		tx.Model(&model.RolePermission{}).
			Where("role_key = ? AND permission_code = ?", rk, "user:list").
			Count(&exists)
		if exists == 0 {
			rp := model.RolePermission{RoleKey: rk, PermissionCode: "user:list"}
			if err := tx.Create(&rp).Error; err != nil {
				return fmt.Errorf("insert user:list for role_key %d: %w", rk, err)
			}
		}
	}

	return nil
}
