// Permission types

export interface PermissionData {
  isSuperAdmin: boolean
  teamPermissions: Record<string, string[]>
}

export interface Role {
  bizKey: string
  roleName: string
  roleDesc: string
  isPreset: boolean
  permissionCount: number
  memberCount: number
  createTime: string
}

export interface PermissionItem {
  code: string
  description: string
}

export interface PermissionGroup {
  resource: string
  actions: PermissionItem[]
}

export interface RoleDetail {
  bizKey: string
  roleName: string
  roleDesc: string
  isPreset: boolean
  permissions: PermissionItem[]
  memberCount: number
  createTime: string
}

export interface CreateRoleReq {
  name: string
  description?: string
  permissionCodes: string[]
}

export interface UpdateRoleReq {
  name?: string
  description?: string
  permissionCodes?: string[]
}

export interface RoleListParams {
  search?: string
  isPreset?: string
  page?: number
  pageSize?: number
}

// Domain models

export interface TeamSummary {
  bizKey: string
  name: string
  role: string
}

export interface User {
  bizKey: string
  username: string
  displayName: string
  email?: string
  isSuperAdmin: boolean
  userStatus?: 'enabled' | 'disabled'
  teams?: TeamSummary[]
  createTime: string
}

export interface Team {
  bizKey: string
  name: string
  code?: string
  description: string
  pmKey: string
  pmDisplayName?: string
  createdAt: string
  updatedAt: string
}

export interface MainItem {
  bizKey: string
  teamKey: string
  code: string
  title: string
  itemDesc?: string
  priority: string
  proposerKey: string
  assigneeKey: string | null
  planStartDate: string | null
  expectedEndDate: string | null
  actualEndDate: string | null
  itemStatus: string
  statusName?: string
  completion: number
  createTime: string
  dbUpdateTime: string
}

export interface SubItem {
  bizKey: string
  teamKey: string
  mainItemKey: string
  code: string
  title: string
  itemDesc: string
  priority: string
  assigneeKey: string | null
  planStartDate: string | null
  expectedEndDate: string | null
  actualEndDate: string | null
  itemStatus: string
  statusName?: string
  completion: number
  weight: number
  createTime: string
  dbUpdateTime: string
}

export interface ProgressRecord {
  subItemKey: string
  teamKey: string
  authorKey: string
  authorName?: string
  completion: number
  achievement: string
  blocker: string
  lesson: string
  isPMCorrect: boolean
  createTime: string
}

export interface WeeklyProgressRecord {
  bizKey: string
  completion: number
  achievement: string
  blocker: string
  createdAt: string
}

export interface ItemPool {
  bizKey: string
  teamKey: string
  title: string
  background: string
  expectedOutput: string
  submitterKey: string
  submitterName?: string
  poolStatus: string
  assignedMainKey: string | null
  assignedSubKey: string | null
  assignedMainCode: string
  assignedMainTitle: string
  assigneeKey: string | null
  rejectReason: string
  reviewedAt: string | null
  reviewerKey: string | null
  createTime: string
  dbUpdateTime: string
}

// Paginated result
export interface PageResult<T> {
  items: T[]
  total: number
  page: number
  size: number
}

// API request types

// Auth
export interface LoginReq {
  username: string
  password: string
}

export interface LoginResp {
  token: string
  user: User
}

// Teams
export interface CreateTeamReq {
  name: string
  code: string
  description?: string
}

export interface UpdateTeamReq {
  name?: string
  description?: string
}

export interface DeleteTeamReq {
  confirmName: string
}

export interface InviteMemberReq {
  username: string
  roleKey: string
}

export interface TransferPMReq {
  newPmUserKey: string
}

export interface TeamDetailResp {
  bizKey: string
  name: string
  code: string
  description: string
  pmKey: string
  pmDisplayName: string
  memberCount: number
  mainItemCount: number
  createTime: string
  dbUpdateTime: string
}

export interface TeamMemberResp {
  bizKey: string
  teamKey: string
  userKey: string
  role: string
  roleKey: string
  roleName: string
  joinedAt: string
  displayName: string
  username: string
}

// MainItems
export interface CreateMainItemReq {
  title: string
  description?: string
  priority: string
  assigneeKey: string
  startDate: string
  expectedEndDate: string
}

export interface UpdateMainItemReq {
  title?: string
  description?: string
  priority?: string
  assigneeKey?: string | null
  startDate?: string | null
  expectedEndDate?: string | null
  actualEndDate?: string | null
}

export interface MainItemFilter {
  priority?: string
  status?: string
  assigneeKey?: string
  archived?: boolean
  page?: number
  pageSize?: number
}

// SubItems
export interface CreateSubItemReq {
  title: string
  description?: string
  priority: string
  assigneeKey: string
  startDate?: string
  expectedEndDate?: string
}

export interface UpdateSubItemReq {
  title?: string
  description?: string
  priority?: string
  expectedEndDate?: string
}

export interface ChangeStatusReq {
  status: string
}

export interface SubItemFilter {
  priority?: string
  status?: string
  assigneeKey?: string
  page?: number
  pageSize?: number
}

// Progress
export interface AppendProgressReq {
  completion: number
  achievement?: string
  blocker?: string
  lesson?: string
}

// ItemPool
export interface UpdateItemPoolReq {
  title?: string
  background?: string
  expectedOutput?: string
}

export interface SubmitItemPoolReq {
  title: string
  background?: string
  expectedOutput?: string
}

export interface AssignItemPoolReq {
  mainItemKey: string
  assigneeKey: string
  priority?: string
  startDate: string
  expectedEndDate: string
}

export interface ConvertToMainItemReq {
  priority: string
  assigneeKey: string
  startDate: string
  expectedEndDate: string
}

export interface RejectItemPoolReq {
  reason: string
}

export interface ItemPoolFilter {
  status?: string
  page?: number
  pageSize?: number
}

export interface AssignItemPoolResp {
  mainItemBizKey: string
  subItemBizKey: string
}

// Views
export interface GanttViewResp {
  items: GanttMainItem[]
}

export interface GanttMainItem {
  bizKey: string
  title: string
  priority: string
  startDate: string | null
  expectedEndDate: string | null
  completion: number
  itemStatus: string
  isOverdue: boolean
  subItems: GanttSubItem[]
}

export interface GanttSubItem {
  bizKey: string
  title: string
  startDate: string | null
  expectedEndDate: string | null
  completion: number
  itemStatus: string
}

export interface TableFilter {
  type?: string
  priority?: string
  status?: string
  assigneeKey?: string
  sortBy?: string
  sortOrder?: string
  page?: number
  pageSize?: number
}

export interface TableRow {
  bizKey: string
  type: string
  code: string
  title: string
  priority: string
  assigneeKey: string | null
  assigneeName: string
  itemStatus: string
  completion: number
  expectedEndDate: string | null
  actualEndDate: string | null
  mainItemId?: string | null
}

// Reports
export interface ReportPreviewResp {
  weekStart: string
  weekEnd: string
  sections: ReportSection[]
}

export interface ReportSection {
  mainItem: { bizKey: string; title: string; completion: number; isKeyItem?: boolean }
  subItems: ReportSubItem[]
}

export interface ReportSubItem {
  bizKey: string
  title: string
  completion: number
  assigneeKey?: string | null
  assigneeName?: string
  achievements: string[]
  blockers: string[]
}

// Admin
export interface AdminUser {
  bizKey: string
  username: string
  displayName: string
  email?: string
  isSuperAdmin: boolean
  userStatus?: 'enabled' | 'disabled'
  teams?: TeamSummary[]
}

export interface AdminTeam {
  bizKey: string
  name: string
  pmDisplayName: string
  memberCount: number
  mainItemCount: number
  createTime: string
}

export interface CreateUserReq {
  username: string
  displayName: string
  email?: string
  teamKey?: string
}

export interface CreateUserResp {
  bizKey: string
  username: string
  displayName: string
  email: string
  userStatus: string
  teams: TeamSummary[]
  initialPassword: string
}

export interface UpdateUserReq {
  displayName?: string
  email?: string
  teamKey?: string
}

export interface UpdateUserResp {
  bizKey: string
  username: string
  displayName: string
  email: string
  userStatus: string
  teams: TeamSummary[]
}

export interface ToggleUserStatusReq {
  status: 'enabled' | 'disabled'
}

export interface ToggleUserStatusResp {
  bizKey: string
  username: string
  userStatus: string
}

export interface GetUserResp {
  bizKey: string
  username: string
  displayName: string
  email: string
  isSuperAdmin: boolean
  userStatus: string
  teams: TeamSummary[]
}

export interface ResetPasswordReq {
  newPassword: string
}

export interface ResetPasswordResp {
  bizKey: string
  username: string
  displayName: string
}

// Weekly view (enhanced)
export interface MainItemSummary {
  bizKey: string
  code: string
  title: string
  priority: string
  itemStatus: string
  startDate: string
  expectedEndDate: string
  actualEndDate: string | null
  completion: number
  subItemCount: number
}

export interface WeeklyViewResponse {
  weekStart: string
  weekEnd: string
  stats: {
    activeSubItems: number
    newlyCompleted: number
    inProgress: number
    blocked: number
    pending: number
    pausing: number
    overdue: number
  }
  groups: WeeklyComparisonGroup[]
}

export interface WeeklyComparisonGroup {
  mainItem: MainItemSummary
  lastWeek: SubItemSnapshot[]
  thisWeek: SubItemSnapshot[]
  completedNoChange: SubItemSnapshot[]
}

export interface SubItemSnapshot {
  bizKey: string
  code: string
  title: string
  priority: string
  itemStatus: string
  assigneeName: string
  startDate: string
  expectedEndDate: string
  actualEndDate?: string | null
  completion: number
  progressDescription: string
  progressRecords: WeeklyProgressRecord[]
  delta?: number
  isNew?: boolean
  justCompleted?: boolean
}
