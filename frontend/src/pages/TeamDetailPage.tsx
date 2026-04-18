import { useState, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  getTeamApi,
  listMembersApi,
  transferPmApi,
  removeMemberApi,
  deleteTeamApi,
  inviteMemberApi,
} from '@/api/teams'
import type { TeamMemberResp } from '@/types'
import { Breadcrumb, BreadcrumbItem, BreadcrumbSeparator } from '@/components/ui/breadcrumb'
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

// --- Helpers ---

function formatDate(dateStr: string): string {
  return dateStr.slice(0, 10).replace(/-/g, '/')
}

// --- Main Component ---

export default function TeamDetailPage() {
  const { teamId } = useParams<{ teamId: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { addToast } = useToast()
  const numericTeamId = Number(teamId)

  // --- Data fetching ---

  const { data: team, isLoading: teamLoading } = useQuery({
    queryKey: ['team', numericTeamId],
    queryFn: () => getTeamApi(numericTeamId),
    enabled: !!numericTeamId,
  })

  const { data: members = [], isLoading: membersLoading } = useQuery({
    queryKey: ['teamMembers', numericTeamId],
    queryFn: () => listMembersApi(numericTeamId),
    enabled: !!numericTeamId,
  })

  const isLoading = teamLoading || membersLoading

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
  const [inviteRole, setInviteRole] = useState('member')
  const [disbandOpen, setDisbandOpen] = useState(false)
  const [disbandInput, setDisbandInput] = useState('')

  // --- Mutations ---

  const transferMutation = useMutation({
    mutationFn: () => transferPmApi(numericTeamId, { newPmUserId: transferTarget!.userId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['team', numericTeamId] })
      qc.invalidateQueries({ queryKey: ['teamMembers', numericTeamId] })
      setTransferTarget(null)
      addToast('已设为PM', 'success')
    },
  })

  const removeMutation = useMutation({
    mutationFn: () => removeMemberApi(numericTeamId, removeTarget!.userId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['team', numericTeamId] })
      qc.invalidateQueries({ queryKey: ['teamMembers', numericTeamId] })
      setRemoveTarget(null)
      addToast('已移除成员', 'success')
    },
  })

  const inviteMutation = useMutation({
    mutationFn: () => inviteMemberApi(numericTeamId, { username: inviteUsername, role: inviteRole }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['team', numericTeamId] })
      qc.invalidateQueries({ queryKey: ['teamMembers', numericTeamId] })
      setInviteOpen(false)
      setInviteUsername('')
      setInviteRole('member')
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

  // Find non-members from available users (for invite dialog)
  // We'll use a simple approach: show users not in current member list
  // The actual available users would come from a search API, but for simplicity
  // we allow typing a username
  const memberUsernames = new Set(members.map((m) => m.username))

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
            <div className="text-xs text-tertiary mb-1">PM</div>
            <div className="flex items-center gap-2">
              <UserAvatar name={team.pm.displayName} size="sm" />
              <span className="text-sm font-medium text-primary">{team.pm.displayName}</span>
            </div>
          </div>
          <div>
            <div className="text-xs text-tertiary mb-1">成员数</div>
            <span className="text-sm font-medium text-primary">{team.memberCount}</span>
          </div>
          <div>
            <div className="text-xs text-tertiary mb-1">创建时间</div>
            <span className="text-sm text-primary">{formatDate(team.createdAt)}</span>
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
        <Button size="sm" onClick={() => setInviteOpen(true)}>添加成员</Button>
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
              <TableRow key={member.userId}>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <UserAvatar name={member.displayName} size="sm" />
                    <span className="font-medium text-primary">{member.displayName}</span>
                  </div>
                </TableCell>
                <TableCell>
                  {member.role === 'pm' ? (
                    <Badge variant="primary">PM</Badge>
                  ) : (
                    <Badge variant="default">成员</Badge>
                  )}
                </TableCell>
                <TableCell>
                  <span className="text-sm text-primary">{formatDate(member.joinedAt)}</span>
                </TableCell>
                <TableCell>
                  {member.role === 'pm' ? (
                    <span className="text-tertiary">&mdash;</span>
                  ) : (
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-primary-600"
                        onClick={() => setTransferTarget(member)}
                      >
                        设为PM
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-error"
                        onClick={() => setRemoveTarget(member)}
                      >
                        移除
                      </Button>
                    </div>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {/* Danger Zone */}
      <div className="rounded-xl border border-red-300 bg-white shadow-sm">
        <div className="px-5 py-3 border-b border-red-200">
          <h3 className="text-[15px] font-semibold text-error">危险操作</h3>
        </div>
        <div className="px-5 py-4 flex items-center justify-between">
          <div>
            <div className="font-medium text-primary">解散团队</div>
            <div className="text-[13px] text-secondary">解散后所有数据将无法恢复，所有成员将被移除。</div>
          </div>
          <Button variant="danger" onClick={() => setDisbandOpen(true)}>解散团队</Button>
        </div>
      </div>

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
      <Dialog open={inviteOpen} onOpenChange={(open) => { setInviteOpen(open); if (!open) { setInviteUsername(''); setInviteRole('member') } }}>
        <DialogContent size="md">
          <DialogHeader>
            <DialogTitle>添加成员</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className="flex flex-col gap-4">
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  搜索用户 <span className="text-error">*</span>
                </label>
                <Input
                  placeholder="请输入用户名"
                  value={inviteUsername}
                  onChange={(e) => setInviteUsername(e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  角色 <span className="text-error">*</span>
                </label>
                <Select value={inviteRole} onValueChange={setInviteRole}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="pm">PM</SelectItem>
                    <SelectItem value="member">成员</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setInviteOpen(false)}>取消</Button>
            <Button
              onClick={() => inviteMutation.mutate()}
              disabled={!inviteUsername.trim() || inviteMutation.isPending}
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
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 text-sm text-red-800 mb-4">
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
