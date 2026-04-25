// Permission types

export interface PermissionData {
  isSuperAdmin: boolean
  teamPermissions: Record<number, string[]>
}

export interface Role {
  id: number
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
  id: number
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
  id: number
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
  teamName: string
  code?: string
  teamDesc: string
  pmKey: string
  pmDisplayName?: string
  createTime: string
  dbUpdateTime: string
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
  roleId: number
}

export interface TransferPMReq {
  newPmUserId: number
}

export interface TeamDetailResp {
  id: number
  name: string
  code: string
  description: string
  pmId: number
  pmDisplayName: string
  memberCount: number
  mainItemCount: number
  createdAt: string
}

export interface TeamMemberResp {
  id: number
  teamId: number
  userId: number
  role: string
  roleId: number
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
  assigneeId: number
  startDate: string
  expectedEndDate: string
}

export interface UpdateMainItemReq {
  title?: string
  description?: string
  priority?: string
  assigneeId?: number | null
  startDate?: string | null
  expectedEndDate?: string | null
  actualEndDate?: string | null
}

export interface MainItemFilter {
  priority?: string
  status?: string
  assigneeId?: number
  archived?: boolean
  page?: number
  pageSize?: number
}

// SubItems
export interface CreateSubItemReq {
  title: string
  description?: string
  priority: string
  assigneeId: number
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
  assigneeId?: number
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
export interface SubmitItemPoolReq {
  title: string
  background?: string
  expectedOutput?: string
}

export interface AssignItemPoolReq {
  mainItemId: number
  assigneeId: number
  priority?: string
  startDate: string
  expectedEndDate: string
}

export interface ConvertToMainItemReq {
  priority: string
  assigneeId: number
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
  mainItemId: number
  subItemId: number
}

// Views
export interface GanttViewResp {
  items: GanttMainItem[]
}

export interface GanttMainItem {
  id: number
  title: string
  priority: string
  startDate: string | null
  expectedEndDate: string | null
  completion: number
  status: string
  isOverdue: boolean
  subItems: GanttSubItem[]
}

export interface GanttSubItem {
  id: number
  title: string
  startDate: string | null
  expectedEndDate: string | null
  completion: number
  status: string
}

export interface TableFilter {
  type?: string
  priority?: string
  status?: string
  assigneeId?: number
  sortBy?: string
  sortOrder?: string
  page?: number
  pageSize?: number
}

export interface TableRow {
  id: number
  type: string
  code: string
  title: string
  priority: string
  assigneeId: number | null
  assigneeName: string
  status: string
  completion: number
  expectedEndDate: string | null
  actualEndDate: string | null
  mainItemId?: number | null
}

// Reports
export interface ReportPreviewResp {
  weekStart: string
  weekEnd: string
  sections: ReportSection[]
}

export interface ReportSection {
  mainItem: { id: number; title: string; completion: number; isKeyItem?: boolean }
  subItems: ReportSubItem[]
}

export interface ReportSubItem {
  id: number
  title: string
  completion: number
  assigneeId?: number | null
  assigneeName?: string
  achievements: string[]
  blockers: string[]
}

// Admin
export interface AdminUser {
  id: number
  username: string
  displayName: string
  email?: string
  isSuperAdmin: boolean
  status?: 'enabled' | 'disabled'
  teams?: TeamSummary[]
}

export interface AdminTeam {
  id: number
  name: string
  pmDisplayName: string
  memberCount: number
  mainItemCount: number
  createdAt: string
}

export interface CreateUserReq {
  username: string
  displayName: string
  email?: string
  teamId?: number
}

export interface CreateUserResp {
  id: number
  username: string
  displayName: string
  email: string
  status: string
  teams: TeamSummary[]
  initialPassword: string
}

export interface UpdateUserReq {
  displayName?: string
  email?: string
  teamId?: number
}

export interface UpdateUserResp {
  id: number
  username: string
  displayName: string
  email: string
  status: string
  teams: TeamSummary[]
}

export interface ToggleUserStatusReq {
  status: 'enabled' | 'disabled'
}

export interface ToggleUserStatusResp {
  id: number
  username: string
  status: string
}

export interface GetUserResp {
  id: number
  username: string
  displayName: string
  email: string
  isSuperAdmin: boolean
  status: string
  teams: TeamSummary[]
}

// Weekly view (enhanced)
export interface MainItemSummary {
  id: number
  code: string
  title: string
  priority: string
  status: string
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
  id: number
  code: string
  title: string
  priority: string
  status: string
  assigneeName: string
  startDate: string
  expectedEndDate: string
  actualEndDate?: string | null
  completion: number
  progressDescription: string
  progressRecords: ProgressRecord[]
  delta?: number
  isNew?: boolean
  justCompleted?: boolean
}
