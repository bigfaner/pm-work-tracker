// Package permissions defines the system-wide permission code registry.
// All permission codes follow the resource:action format and serve as the
// single source of truth for RBAC middleware, services, and handlers.
package permissions

// Permission describes a single permission code.
type Permission struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// ResourcePermissions groups permission codes by resource.
type ResourcePermissions struct {
	Resource    string       `json:"resource"`
	Permissions []Permission `json:"permissions"`
}

// Registry is the ordered list of all resources and their permissions.
var Registry = []ResourcePermissions{
	{
		Resource: "team",
		Permissions: []Permission{
			{Code: "team:create", Description: "创建团队（全局权限，非团队上下文）"},
			{Code: "team:read", Description: "查看团队信息"},
			{Code: "team:update", Description: "编辑团队信息"},
			{Code: "team:delete", Description: "解散团队"},
			{Code: "team:invite", Description: "邀请成员加入"},
			{Code: "team:remove", Description: "移除团队成员"},
			{Code: "team:transfer", Description: "转让PM身份"},
		},
	},
	{
		Resource: "main_item",
		Permissions: []Permission{
			{Code: "main_item:create", Description: "创建主事项"},
			{Code: "main_item:read", Description: "查看主事项"},
			{Code: "main_item:update", Description: "编辑主事项"},
			{Code: "main_item:archive", Description: "归档主事项"},
		},
	},
	{
		Resource: "sub_item",
		Permissions: []Permission{
			{Code: "sub_item:create", Description: "创建子事项"},
			{Code: "sub_item:read", Description: "查看子事项"},
			{Code: "sub_item:update", Description: "编辑子事项"},
			{Code: "sub_item:assign", Description: "分配子事项负责人"},
			{Code: "sub_item:change_status", Description: "变更子事项状态"},
		},
	},
	{
		Resource: "progress",
		Permissions: []Permission{
			{Code: "progress:create", Description: "追加进度记录"},
			{Code: "progress:read", Description: "查看进度记录"},
			{Code: "progress:update", Description: "修正进度记录"},
		},
	},
	{
		Resource: "item_pool",
		Permissions: []Permission{
			{Code: "item_pool:submit", Description: "提交事项到事项池"},
			{Code: "item_pool:review", Description: "审核/分配/拒绝事项池事项"},
		},
	},
	{
		Resource: "view",
		Permissions: []Permission{
			{Code: "view:weekly", Description: "查看周视图"},
			{Code: "view:gantt", Description: "查看甘特图"},
			{Code: "view:table", Description: "查看表格视图"},
		},
	},
	{
		Resource: "report",
		Permissions: []Permission{
			{Code: "report:export", Description: "导出周报"},
		},
	},
	{
		Resource: "user",
		Permissions: []Permission{
			{Code: "user:read", Description: "查看用户信息"},
			{Code: "user:update", Description: "编辑用户信息"},
			{Code: "user:manage_role", Description: "管理角色定义"},
		},
	},
}

// allCodes is the lazily populated flat set of all valid permission codes.
var allCodes map[string]bool

func init() {
	allCodes = make(map[string]bool, countCodes())
	for _, rp := range Registry {
		for _, p := range rp.Permissions {
			allCodes[p.Code] = true
		}
	}
}

// AllCodes returns a flat map[string]bool containing every registered permission code.
func AllCodes() map[string]bool {
	result := make(map[string]bool, len(allCodes))
	for k, v := range allCodes {
		result[k] = v
	}
	return result
}

// ValidateCode returns true if the given code is a registered permission code.
func ValidateCode(code string) bool {
	return allCodes[code]
}

// countCodes counts total permission codes across all resources.
func countCodes() int {
	n := 0
	for _, rp := range Registry {
		n += len(rp.Permissions)
	}
	return n
}
