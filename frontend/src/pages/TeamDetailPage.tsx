import { useState, useMemo } from 'react'
import { Crown, UserMinus, Edit, RefreshCw } from 'lucide-react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getTeamApi,
  listMembersApi,
  transferPmApi,
  removeMemberApi,
  deleteTeamApi,
  inviteMemberApi,
  changeMemberRoleApi,
  searchAvailableUsersApi,
  type UserSearchResult,
} from '@/api/teams'
import { listRolesApi } from '@/api/roles'
import type { TeamMemberResp } from '@/types'
import { Breadcrumb, BreadcrumbItem, BreadcrumbSeparator } from '@/components/ui/breadcrumb'
import { PermissionGuard } from '@/components/PermissionGuard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import UserAvatar from '@/components/shared/UserAvatar'
import { useToast } from '@/components/ui/toast'
import { useAuthStore } from '@/store/auth'
import { formatDate as _formatDate } from '@/lib/format'

// TeamDetailPage receives ISO datetime strings; truncate to date before formatting
function formatDate(dateStr: string): string {
  return _formatDate(dateStr.slice(0, 10))
}

// --- Main Component ---

export default function TeamDetailPage() {
  const { teamId } = useParams<{ teamId: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { addToast } = useToast()
  const numericTeamId = teamId!

  // --- Data fetching ---

  const { data: team, isLoading: teamLoading } = useQuery({
    queryKey: ['team', numericTeamId],
    queryFn: () => getTeamApi(numericTeamId),
    enabled: !!numericTeamId,
  })

  const { data: members = [], isLoading: membersLoading, isFetching: membersFetching, refetch: refetchMembers } = useQuery({
    queryKey: ['teamMembers', numericTeamId],
    queryFn: () => listMembersApi(numericTeamId),
    enabled: !!numericTeamId,
  })

  // Roles list (exclude superadmin)
  const { data: rolesData } = useQuery({
    queryKey: ['roles'],
    queryFn: () => listRolesApi({ pageSize: 100 }),
    staleTime: 5 * 60 * 1000,
  })

  const roles = useMemo(() => {
    if (!rolesData?.items) return []
    return rolesData.items.filter((r) => r.roleName !== 'superadmin' && r.roleName !== 'pm')
  }, [rolesData])

  const defaultRoleId = roles.find((r) => r.roleName === 'member')?.id ?? roles[0]?.id

  const isLoading = teamLoading || membersLoading

  // Current user info for PM self-check
  const currentUser = useAuthStore((s) => s.user)

  // --- Filter state ---

  const [searchText, setSearchText] = useState('')
  const [roleFilter, setRoleFilter] = useState('')

  const filteredMembers = useMemo(() => {
    let result = members
    if (searchText.trim()) {
      const q = searchText.trim().toLowerCase()
      result = result.filter((m) => m.displayName.toLowerCase().includes(q))
    }
    if (roleFilter) {
      result = result.filter((m) => m.role === roleFilter)
    }
    return result
  }, [members, searchText, roleFilter])

  // --- Dialog state ---

  const [transferTarget, setTransferTarget] = useState<TeamMemberResp | null>(null)
  const [removeTarget, setRemoveTarget] = useState<TeamMemberResp | null>(null)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [inviteUsername, setInviteUsername] = useState('')
  const [inviteRoleId, setInviteRoleId] = useState<number | undefined>(undefined)
  const [userSearch, setUserSearch] = useState('')
  const [selectedUser, setSelectedUser] = useState<UserSearchResult | null>(null)
  const [userDropdownOpen, setUserDropdownOpen] = useState(false)
  const [disbandOpen, setDisbandOpen] = useState(false)
  const [disbandInput, setDisbandInput] = useState('')

  // Role edit dialog state
  const [roleEditTarget, setRoleEditTarget] = useState<TeamMemberResp | null>(null)
  const [roleEditRoleId, setRoleEditRoleId] = useState<number | undefined>(undefined)

  // User search for invite dialog
  const { data: userSearchResults = [] } = useQuery({
    queryKey: ['searchUsers', numericTeamId, userSearch],
    queryFn: () => searchAvailableUsersApi(numericTeamId, userSearch),
    enabled: inviteOpen && !!numericTeamId,
  })

  // --- Mutations ---

  const transferMutation = useMutation({
    mutationFn: () => transferPmApi(numericTeamId, { newPmUserKey: String(transferTarget!.userBizKey) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['team', numericTeamId] })
      qc.invalidateQueries({ queryKey: ['teamMembers', numericTeamId] })
      setTransferTarget(null)
      addToast('已设为PM', 'success')
    },
  })

  const removeMutation = useMutation({
    mutationFn: () => removeMemberApi(numericTeamId, removeTarget!.userBizKey),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['team', numericTeamId] })
      qc.invalidateQueries({ queryKey: ['teamMembers', numericTeamId] })
      setRemoveTarget(null)
      addToast('已移除成员', 'success')
    },
  })

  const inviteMutation = useMutation({
    mutationFn: () => inviteMemberApi(numericTeamId, { username: inviteUsername, roleKey: String(inviteRoleId!) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['team', numericTeamId] })
      qc.invalidateQueries({ queryKey: ['teamMembers', numericTeamId] })
      setInviteOpen(false)
      setInviteUsername('')
      setInviteRoleId(undefined)
      setUserSearch('')
      setSelectedUser(null)
      addToast('成员已添加', 'success')
    },
    onError: (err: any) => {
      const code = err?.response?.data?.code
      if (code === 'ALREADY_MEMBER') {
        addToast('该用户已是团队成员', 'error')
      } else {
        addToast('添加失败，请稍后重试', 'error')
      }
    },
  })

  const changeRoleMutation = useMutation({
    mutationFn: ({ memberId, roleId }: { memberId: string; roleId: number }) =>
      changeMemberRoleApi(numericTeamId, memberId, { roleKey: String(roleId) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['teamMembers', numericTeamId] })
      setRoleEditTarget(null)
      addToast('角色已更新', 'success')
    },
    onError: () => {
      setRoleEditTarget(null)
      addToast('角色变更失败，请稍后重试', 'error')
    },
  })

  const disbandMutation = useMutation({
    mutationFn: () => deleteTeamApi(numericTeamId, { confirmName: disbandInput }),
    onSuccess: () => {
      setDisbandOpen(false)
      setDisbandInput('')
      addToast('团队已解散', 'success')
      navigate('/teams')
    },
    onError: (err: any) => {
      const code = err?.response?.data?.code
      if (code === 'NAME_MISMATCH') {
        addToast('团队名称不匹配', 'error')
      } else {
        addToast('解散失败，请稍后重试', 'error')
      }
    },
  })

  // --- Render ---

  if (isLoading) {
    return (
      <div data-testid="team-detail-page" className="py-8 text-center text-tertiary text-sm">
        加载中...
      </div>
    )
  }

  if (!team) {
    return (
      <div data-testid="team-detail-page" className="py-8 text-center text-tertiary text-sm">
        团队不存在
      </div>
    )
  }

  const isPm = (member: TeamMemberResp) => member.role === 'pm'
  const isSelf = (member: TeamMemberResp) => currentUser != null && String(member.userBizKey) === currentUser.bizKey

  return (
    <div data-testid="team-detail-page">
      {/* Breadcrumb */}
      <Breadcrumb className="mb-4">
        <BreadcrumbItem href="/teams">团队管理</BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem isCurrent>{team.name}</BreadcrumbItem>
      </Breadcrumb>

      {/* Page Header */}
      <div className="mb-4">
        <h1 className="text-xl font-semibold text-primary">{team.name}</h1>
      </div>

      {/* Team Info Card */}
      <div className="rounded-xl border border-border bg-white shadow-sm p-5 mb-5">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="text-xs text-tertiary mb-1">团队名称</div>
            <span className="text-sm font-medium text-primary">{team.name}</span>
          </div>
          <div>
            <div className="text-xs text-tertiary mb-1">CODE</div>
            <span className="text-sm font-medium text-primary">{team.code}</span>
          </div>
          <div>
            <div className="text-xs text-tertiary mb-1">PM</div>
            <div className="flex items-center gap-2">
              <UserAvatar name={team.pmDisplayName} size="sm" />
              <span className="text-sm font-medium text-primary">{team.pmDisplayName}</span>
            </div>
          </div>
          <div>
            <div className="text-xs text-tertiary mb-1">成员数</div>
            <span className="text-sm font-medium text-primary">{team.memberCount}</span>
          </div>
          <div className="col-span-2">
            <div className="text-xs text-tertiary mb-1">简介</div>
            <span className="text-sm text-secondary">{team.description || '暂无简介'}</span>
          </div>
        </div>
      </div>

      {/* Member List Section */}
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-[15px] font-semibold text-primary">成员列表</h3>
        <PermissionGuard code="team:invite">
          <Button size="sm" onClick={() => { setInviteOpen(true); setInviteRoleId(defaultRoleId); setUserSearch(''); setSelectedUser(null) }}>添加成员</Button>
        </PermissionGuard>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 mb-3">
        <Input
          placeholder="搜索姓名..."
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          className="w-[180px]"
        />
        <Select value={roleFilter || '_all'} onValueChange={(v) => setRoleFilter(v === '_all' ? '' : v)}>
          <SelectTrigger className="w-[120px]">
            <SelectValue placeholder="角色：全部" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="_all">角色：全部</SelectItem>
            <SelectItem value="pm">PM</SelectItem>
            <SelectItem value="member">成员</SelectItem>
          </SelectContent>
        </Select>
        <Button variant="secondary" size="sm" onClick={async () => { await refetchMembers(); addToast('数据已刷新', 'success') }} disabled={membersFetching} data-testid="refresh-btn">
          <RefreshCw size={14} className={membersFetching ? 'animate-spin' : ''} />
          刷新
        </Button>
      </div>

      {/* Member Table */}
      <div className="rounded-xl border border-border bg-white shadow-sm mb-5">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>姓名</TableHead>
              <TableHead>角色</TableHead>
              <TableHead>加入时间</TableHead>
              <TableHead>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredMembers.map((member) => (
              <TableRow key={member.userBizKey}>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <UserAvatar name={member.displayName} size="sm" />
                    <span className="font-medium text-primary">{member.displayName}</span>
                  </div>
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    {isPm(member) ? (
                      <Badge variant="primary">PM</Badge>
                    ) : (
                      <Badge variant="default">{member.roleName || '成员'}</Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell>
                  <span className="text-sm text-primary">{formatDate(member.joinedAt)}</span>
                </TableCell>
                <TableCell>
                  {isPm(member) ? (
                    <span className="text-tertiary">&mdash;</span>
                  ) : (
                    <div className="flex gap-1">
                      <PermissionGuard code="team:invite">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-primary-600"
                          onClick={() => { setRoleEditTarget(member); setRoleEditRoleId(member.roleId) }}
                          data-testid="change-role-btn"
                        >
                          <Edit className="w-3.5 h-3.5" />
                          修改角色
                        </Button>
                      </PermissionGuard>
                      <PermissionGuard code="team:transfer">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-primary-600"
                          onClick={() => setTransferTarget(member)}
                        >
                          <Crown className="w-3.5 h-3.5" />
                          设为PM
                        </Button>
                      </PermissionGuard>
                      <PermissionGuard code="team:remove">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-primary-600"
                          onClick={() => setRemoveTarget(member)}
                        >
                          <UserMinus className="w-3.5 h-3.5" />
                          移除
                        </Button>
                      </PermissionGuard>
                    </div>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {/* Danger Zone */}
      <div className="rounded-xl border border-error-text/40 bg-white shadow-sm">
        <div className="px-5 py-3 border-b border-error-text/20">
          <h3 className="text-[15px] font-semibold text-error">危险操作</h3>
        </div>
        <div className="px-5 py-4 flex items-center justify-between">
          <div>
            <div className="font-medium text-primary">解散团队</div>
            <div className="text-[13px] text-secondary">解散后所有数据将无法恢复，所有成员将被移除。</div>
          </div>
          <PermissionGuard code="team:delete">
            <Button variant="danger" onClick={() => setDisbandOpen(true)}>解散团队</Button>
          </PermissionGuard>
        </div>
      </div>

      {/* Change Role Dialog */}
      <Dialog open={!!roleEditTarget} onOpenChange={(open) => { if (!open) setRoleEditTarget(null) }}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>修改角色</DialogTitle>
            <DialogDescription>
              修改「{roleEditTarget?.displayName}」的角色
            </DialogDescription>
          </DialogHeader>
          <DialogBody>
            <Select
              value={roleEditRoleId != null ? String(roleEditRoleId) : ''}
              onValueChange={(v) => setRoleEditRoleId(Number(v))}
            >
              <SelectTrigger data-testid="role-edit-select">
                <SelectValue placeholder="选择角色" />
              </SelectTrigger>
              <SelectContent>
                {roles.map((role) => (
                  <SelectItem key={role.id} value={String(role.id)}>
                    {role.roleName}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setRoleEditTarget(null)}>取消</Button>
            <Button
              onClick={() => changeRoleMutation.mutate({ memberId: roleEditTarget!.userBizKey, roleId: roleEditRoleId! })}
              disabled={roleEditRoleId == null || roleEditRoleId === roleEditTarget?.roleId || changeRoleMutation.isPending}
            >
              确认修改
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Transfer PM Dialog */}
      <Dialog open={!!transferTarget} onOpenChange={(open) => { if (!open) setTransferTarget(null) }}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>设为PM</DialogTitle>
            <DialogDescription>
              确认将「{transferTarget?.displayName}」设为PM？当前PM将变为普通成员。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setTransferTarget(null)}>取消</Button>
            <Button
              onClick={() => transferMutation.mutate()}
              disabled={transferMutation.isPending}
            >
              确认设为PM
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Remove Member Dialog */}
      <Dialog open={!!removeTarget} onOpenChange={(open) => { if (!open) setRemoveTarget(null) }}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>移除成员</DialogTitle>
            <DialogDescription>
              确认移除成员「{removeTarget?.displayName}」？此操作不可撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setRemoveTarget(null)}>取消</Button>
            <Button
              variant="danger"
              onClick={() => removeMutation.mutate()}
              disabled={removeMutation.isPending}
            >
              确认移除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Invite Member Dialog */}
      <Dialog open={inviteOpen} onOpenChange={(open) => { setInviteOpen(open); if (!open) { setInviteUsername(''); setInviteRoleId(undefined); setUserSearch(''); setSelectedUser(null) } }}>
        <DialogContent size="md">
          <DialogHeader>
            <DialogTitle>添加成员</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className="flex flex-col gap-4">
              <div className="relative">
                <label className="block text-sm font-medium text-primary mb-1">
                  搜索用户 <span className="text-error">*</span>
                </label>
                <Input
                  placeholder="输入用户名或姓名搜索..."
                  value={selectedUser ? `${selectedUser.displayName} (${selectedUser.username})` : userSearch}
                  onChange={(e) => {
                    setUserSearch(e.target.value)
                    setSelectedUser(null)
                    setInviteUsername('')
                    setUserDropdownOpen(true)
                  }}
                  onFocus={() => setUserDropdownOpen(true)}
                  data-testid="invite-user-search"
                />
                {userDropdownOpen && !selectedUser && userSearchResults.length > 0 && (
                  <div className="absolute z-50 mt-1 w-full max-h-48 overflow-auto rounded-md border border-border bg-white shadow-lg" data-testid="invite-user-dropdown">
                    {userSearchResults.map((u) => (
                      <button
                        key={u.bizKey}
                        type="button"
                        className="w-full px-3 py-2 text-left text-sm hover:bg-bg-alt focus:bg-bg-alt focus:outline-none"
                        onClick={() => {
                          setSelectedUser(u)
                          setInviteUsername(u.username)
                          setUserSearch('')
                          setUserDropdownOpen(false)
                        }}
                        data-testid={`invite-user-option-${u.bizKey}`}
                      >
                        <span className="font-medium text-primary">{u.displayName}</span>
                        <span className="ml-2 text-tertiary">{u.username}</span>
                      </button>
                    ))}
                  </div>
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  角色 <span className="text-error">*</span>
                </label>
                <Select
                  value={inviteRoleId != null ? String(inviteRoleId) : ''}
                  onValueChange={(v) => setInviteRoleId(Number(v))}
                >
                  <SelectTrigger data-testid="invite-role-select">
                    <SelectValue placeholder="选择角色" />
                  </SelectTrigger>
                  <SelectContent>
                    {roles.map((role) => (
                      <SelectItem key={role.id} value={String(role.id)}>
                        {role.roleName}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setInviteOpen(false)}>取消</Button>
            <Button
              onClick={() => inviteMutation.mutate()}
              disabled={!selectedUser || inviteRoleId == null || inviteMutation.isPending}
              data-testid="invite-submit-btn"
            >
              确认添加
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Disband Team Dialog */}
      <Dialog open={disbandOpen} onOpenChange={(open) => { setDisbandOpen(open); if (!open) setDisbandInput('') }}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>解散团队</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className="bg-error-bg border border-error-text/20 rounded-lg p-3 text-sm text-error-text mb-4">
              此操作不可恢复，团队下所有事项和数据将被永久删除。
            </div>
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                请输入团队名称确认
              </label>
              <Input
                placeholder={`请输入「${team.name}」`}
                value={disbandInput}
                onChange={(e) => setDisbandInput(e.target.value)}
              />
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setDisbandOpen(false)}>取消</Button>
            <Button
              variant="danger"
              disabled={disbandInput !== team.name || disbandMutation.isPending}
              onClick={() => disbandMutation.mutate()}
            >
              解散团队
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
