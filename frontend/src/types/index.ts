// Domain models

export interface User {
  id: number
  username: string
  display_name: string
  is_super_admin: boolean
  can_create_team: boolean
  created_at: string
  updated_at: string
}

export interface Team {
  id: number
  name: string
  description: string
  pm_id: number
  created_at: string
  updated_at: string
}

export interface TeamMember {
  id: number
  team_id: number
  user_id: number
  role: string
  joined_at: string
  created_at: string
  updated_at: string
}

export interface MainItem {
  id: number
  team_id: number
  code: string
  title: string
  priority: string
  proposer_id: number
  assignee_id: number | null
  start_date: string | null
  expected_end_date: string | null
  actual_end_date: string | null
  status: string
  completion: number
  is_key_item: boolean
  delay_count: number
  archived_at: string | null
  created_at: string
  updated_at: string
}

export interface SubItem {
  id: number
  team_id: number
  main_item_id: number
  title: string
  description: string
  priority: string
  assignee_id: number | null
  start_date: string | null
  expected_end_date: string | null
  actual_end_date: string | null
  status: string
  completion: number
  is_key_item: boolean
  delay_count: number
  weight: number
  created_at: string
  updated_at: string
}

export interface ProgressRecord {
  id: number
  sub_item_id: number
  team_id: number
  author_id: number
  completion: number
  achievement: string
  blocker: string
  lesson: string
  is_pm_correct: boolean
  created_at: string
}

export interface ItemPool {
  id: number
  team_id: number
  title: string
  background: string
  expected_output: string
  submitter_id: number
  status: string
  assigned_main_id: number | null
  assigned_sub_id: number | null
  assignee_id: number | null
  reject_reason: string
  reviewed_at: string | null
  reviewer_id: number | null
  created_at: string
  updated_at: string
}

// Paginated result
export interface PageResult<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
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
  role: string
}

export interface TransferPMReq {
  newPmUserId: number
}

export interface TeamDetailResp {
  id: number
  name: string
  description: string
  pmId: number
  pm: { displayName: string }
  memberCount: number
  mainItemCount: number
  createdAt: string
}

export interface TeamMemberResp {
  userId: number
  displayName: string
  username: string
  role: string
  joinedAt: string
}

// MainItems
export interface CreateMainItemReq {
  title: string
  priority: string
  assigneeId?: number
  startDate?: string
  expectedEndDate?: string
}

export interface UpdateMainItemReq {
  title?: string
  priority?: string
  assigneeId?: number | null
  startDate?: string | null
  expectedEndDate?: string | null
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

export interface AssignSubItemReq {
  assigneeId: number
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

export interface CorrectCompletionReq {
  completion: number
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
  subItemId: number
}

// Views
export interface WeeklyViewResp {
  weekStart: string
  weekEnd: string
  groups: WeeklyGroup[]
}

export interface WeeklyGroup {
  mainItem: { id: number; title: string; completion: number }
  newlyCompleted: SubItemWithProgress[]
  hasProgress: SubItemWithProgress[]
  noChangeFromLastWeek: { id: number; title: string; status: string; completion: number }[]
}

export interface SubItemWithProgress {
  subItem: SubItem
  progressThisWeek: ProgressRecord[]
}

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
  achievements: string[]
  blockers: string[]
}

// Admin
export interface AdminUser {
  id: number
  username: string
  displayName: string
  canCreateTeam: boolean
  isSuperAdmin: boolean
}

export interface SetCanCreateTeamReq {
  canCreateTeam: boolean
}

export interface AdminTeam {
  id: number
  name: string
  pm: { displayName: string }
  memberCount: number
  mainItemCount: number
  createdAt: string
}
