import type { CheckboxOption } from '@/components/ui/checkbox-group'

export interface PermissionGroupDef {
  key: string
  label: string
  permissions: CheckboxOption[]
}

export const PERMISSION_GROUPS: PermissionGroupDef[] = [
  {
    key: 'team',
    label: '团队管理',
    permissions: [
      { value: 'team:create', label: '创建团队' },
      { value: 'team:read', label: '查看团队信息' },
      { value: 'team:update', label: '编辑团队信息' },
      { value: 'team:delete', label: '解散团队' },
      { value: 'team:invite', label: '邀请成员' },
      { value: 'team:remove', label: '移除成员' },
      { value: 'team:transfer', label: '转让 PM' },
    ],
  },
  {
    key: 'main_item',
    label: '主事项',
    permissions: [
      { value: 'main_item:create', label: '创建主事项' },
      { value: 'main_item:read', label: '查看主事项' },
      { value: 'main_item:update', label: '编辑主事项' },
      { value: 'main_item:archive', label: '归档主事项' },
    ],
  },
  {
    key: 'sub_item',
    label: '子事项',
    permissions: [
      { value: 'sub_item:create', label: '创建子事项' },
      { value: 'sub_item:read', label: '查看子事项' },
      { value: 'sub_item:update', label: '编辑子事项' },
      { value: 'sub_item:assign', label: '分配负责人' },
      { value: 'sub_item:change_status', label: '变更状态' },
    ],
  },
  {
    key: 'progress',
    label: '进度管理',
    permissions: [
      { value: 'progress:create', label: '提交进度' },
      { value: 'progress:read', label: '查看进度' },
      { value: 'progress:update', label: '修正进度' },
    ],
  },
  {
    key: 'item_pool',
    label: '事项池',
    permissions: [
      { value: 'item_pool:submit', label: '提交待办' },
      { value: 'item_pool:review', label: '审核待办' },
    ],
  },
  {
    key: 'view',
    label: '视图',
    permissions: [
      { value: 'view:weekly', label: '每周进展' },
      { value: 'view:gantt', label: '甘特图' },
      { value: 'view:table', label: '表格视图' },
    ],
  },
  {
    key: 'report',
    label: '周报',
    permissions: [
      { value: 'report:export', label: '导出周报' },
    ],
  },
  {
    key: 'user',
    label: '用户管理',
    permissions: [
      { value: 'user:read', label: '查看用户' },
      { value: 'user:update', label: '编辑用户' },
      { value: 'user:manage_role', label: '管理角色' },
    ],
  },
]

/** Flat list of all permission codes */
export const ALL_PERMISSION_CODES = PERMISSION_GROUPS.flatMap((g) =>
  g.permissions.map((p) => p.value),
)
