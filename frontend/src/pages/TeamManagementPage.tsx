import { useState, useCallback, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { UserPlus, RefreshCw } from 'lucide-react'
import { listTeamsApi, createTeamApi, inviteMemberApi, searchAvailableUsersApi, type UserSearchResult } from '@/api/teams'
import { listRolesApi } from '@/api/roles'
import type { Team } from '@/types'
import { useAuthStore } from '@/store/auth'
import { PermissionGuard } from '@/components/PermissionGuard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
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
import PaginationBar from '@/components/shared/PaginationBar'
import { useToast } from '@/components/ui/toast'
import { formatDateOnly } from '@/lib/format'

const PAGE_SIZE = 10

export default function TeamManagementPage() {
  const qc = useQueryClient()
  const { addToast } = useToast()

  // Create dialog state
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm, setCreateForm] = useState({ name: '', code: '', description: '' })
  const [createError, setCreateError] = useState('')
  const [codeError, setCodeError] = useState('')

  // Add member dialog state
  const [addMemberOpen, setAddMemberOpen] = useState(false)
  const [addMemberTeamId, setAddMemberTeamId] = useState<string | null>(null)
  const [userSearch, setUserSearch] = useState('')
  const [selectedUser, setSelectedUser] = useState<UserSearchResult | null>(null)
  const [userDropdownOpen, setUserDropdownOpen] = useState(false)
  const [inviteRoleId, setInviteRoleId] = useState<string | undefined>(undefined)

  // Search + pagination state
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  // Data fetching
  const { data, isLoading, isFetching, refetch } = useQuery({
    queryKey: ['teams', search, page],
    queryFn: () => listTeamsApi({ search: search || undefined, page, pageSize: PAGE_SIZE }),
  })

  const teamList = data?.items ?? []
  const total = data?.total ?? 0
  const totalPages = Math.ceil(total / PAGE_SIZE)

  // Roles for invite dialog — only fetch if user can invite members
  const hasPermission = useAuthStore((s) => s.hasPermission)
  const { data: rolesData } = useQuery({
    queryKey: ['roles'],
    queryFn: () => listRolesApi({ pageSize: 100 }),
    staleTime: 5 * 60 * 1000,
    enabled: hasPermission('team:invite'),
  })

  const roles = useMemo(() => {
    if (!rolesData?.items) return []
    return rolesData.items.filter((r) => r.roleName !== 'superadmin' && r.roleName !== 'pm')
  }, [rolesData])

  const defaultRoleId = roles.find((r) => r.roleName === 'member')?.bizKey ?? roles[0]?.bizKey

  // User search for add member dialog
  const { data: userSearchResults = [] } = useQuery({
    queryKey: ['searchUsers', addMemberTeamId, userSearch],
    queryFn: () => searchAvailableUsersApi(addMemberTeamId!, userSearch),
    enabled: addMemberOpen && !!addMemberTeamId,
  })

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (req: { name: string; code: string; description?: string }) => createTeamApi(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['teams'] })
      setCreateOpen(false)
      setCreateForm({ name: '', code: '', description: '' })
      setCreateError('')
      setCodeError('')
    },
    onError: (err: any) => {
      const code = err?.response?.data?.code
      if (code === 'TEAM_CODE_DUPLICATE') {
        setCodeError('该CODE已被使用')
      } else {
        setCreateError('创建失败，请稍后重试')
      }
    },
  })

  // Add member mutation
  const inviteMutation = useMutation({
    mutationFn: () => inviteMemberApi(addMemberTeamId!, { username: selectedUser!.username, roleKey: inviteRoleId! }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['teams'] })
      closeAddMemberDialog()
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

  const closeAddMemberDialog = useCallback(() => {
    setAddMemberOpen(false)
    setAddMemberTeamId(null)
    setUserSearch('')
    setSelectedUser(null)
    setUserDropdownOpen(false)
    setInviteRoleId(undefined)
  }, [])

  const openAddMemberDialog = useCallback((teamId: string) => {
    setAddMemberTeamId(teamId)
    setInviteRoleId(defaultRoleId)
    setUserSearch('')
    setSelectedUser(null)
    setUserDropdownOpen(false)
    setAddMemberOpen(true)
  }, [defaultRoleId])

  const handleCreate = useCallback(() => {
    setCreateError('')
    setCodeError('')
    if (!createForm.name.trim()) {
      setCreateError('请填写团队名称')
      return
    }
    if (!/^[A-Za-z]{2,6}$/.test(createForm.code)) {
      setCodeError('CODE须为 2~6 位英文字母')
      return
    }
    createMutation.mutate({
      name: createForm.name.trim(),
      code: createForm.code.trim(),
      ...(createForm.description.trim() && { description: createForm.description.trim() }),
    })
  }, [createForm, createMutation])

  return (
    <div data-testid="team-management-page">
      {/* Page Header */}
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-semibold text-primary">团队管理</h1>
        <PermissionGuard code="team:create">
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            创建团队
          </Button>
        </PermissionGuard>
      </div>

      {/* Filter Bar */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <Input
          placeholder="搜索团队名称或CODE..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
          className="w-[240px]"
          data-testid="team-search-input"
        />
        <Button
          variant="secondary"
          size="sm"
          onClick={() => refetch()}
          disabled={isFetching}
          data-testid="refresh-btn"
        >
          <RefreshCw size={14} className={isFetching ? 'animate-spin' : ''} />
          刷新
        </Button>
      </div>

      {/* Team Table */}
      {isLoading ? (
        <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
      ) : teamList.length === 0 ? (
        <div className="py-12 text-center">
          <p className="text-tertiary text-sm">暂无团队</p>
          <p className="text-tertiary text-xs mt-1">点击上方按钮创建第一个团队</p>
        </div>
      ) : (
        <div className="rounded-xl border border-border bg-white shadow-sm">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>团队名称</TableHead>
                <TableHead>Code</TableHead>
                <TableHead>项目经理</TableHead>
                <TableHead>简介</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {teamList.map((team) => (
                <TableRow key={team.bizKey}>
                  <TableCell>
                    <Link
                      to={`/teams/${team.bizKey}`}
                      className="font-medium text-primary-600 hover:text-primary-700 hover:underline"
                    >
                      {team.name}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <span className="text-[13px] text-secondary font-mono">{team.code}</span>
                  </TableCell>
                  <TableCell>
                    <span className="text-[13px] text-secondary">{team.pmDisplayName}</span>
                  </TableCell>
                  <TableCell>
                    <span className="text-[13px] text-secondary" style={{ maxWidth: 200 }}>
                      {team.description || '-'}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-[13px] text-secondary">
                      {formatDateOnly(team.createdAt)}
                    </span>
                  </TableCell>
                  <TableCell>
                    <PermissionGuard code="team:invite">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-primary-600"
                        onClick={() => openAddMemberDialog(team.bizKey)}
                        data-testid={`add-member-btn-${team.bizKey}`}
                      >
                        <UserPlus className="w-3.5 h-3.5" />
                        添加成员
                      </Button>
                    </PermissionGuard>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <PaginationBar
            currentPage={page}
            totalPages={totalPages}
            total={total}
            onPageChange={setPage}
          />
        </div>
      )}

      {/* Create Team Dialog */}
      <Dialog open={createOpen} onOpenChange={(open) => { setCreateOpen(open); if (!open) setCreateError('') }}>
        <DialogContent size="md">
          <DialogHeader>
            <DialogTitle>创建团队</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className="flex flex-col gap-4">
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  团队名称 <span className="text-error">*</span>
                </label>
                <Input
                  placeholder="请输入团队名称"
                  value={createForm.name}
                  onChange={(e) => setCreateForm((f) => ({ ...f, name: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  Code <span className="text-error">*</span>
                </label>
                <Input
                  placeholder="如 FEAT、CORE"
                  value={createForm.code}
                  onChange={(e) => { setCreateForm((f) => ({ ...f, code: e.target.value })); setCodeError('') }}
                  onBlur={() => {
                    if (createForm.code && !/^[A-Za-z]{2,6}$/.test(createForm.code)) {
                      setCodeError('CODE须为 2~6 位英文字母')
                    }
                  }}
                />
                {codeError && <p className="mt-1 text-sm text-error">{codeError}</p>}
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">简介</label>
                <Textarea
                  className="text-sm resize-none"
                  rows={3}
                  placeholder="请输入团队简介（选填）"
                  value={createForm.description}
                  onChange={(e) => setCreateForm((f) => ({ ...f, description: e.target.value }))}
                />
              </div>
            </div>
            {createError && (
              <p className="mt-3 text-sm text-error">{createError}</p>
            )}
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => { setCreateOpen(false); setCreateError('') }}>
              取消
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!createForm.name.trim() || !createForm.code.trim() || createMutation.isPending}
            >
              确认创建
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Add Member Dialog */}
      <Dialog open={addMemberOpen} onOpenChange={(open) => { if (!open) closeAddMemberDialog() }}>
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
                    setUserDropdownOpen(true)
                  }}
                  onFocus={() => setUserDropdownOpen(true)}
                  data-testid="add-member-user-search"
                />
                {userDropdownOpen && !selectedUser && userSearchResults.length > 0 && (
                  <div className="absolute z-50 mt-1 w-full max-h-48 overflow-auto rounded-md border border-border bg-white shadow-lg" data-testid="add-member-user-dropdown">
                    {userSearchResults.map((u) => (
                      <button
                        key={u.bizKey}
                        type="button"
                        className="w-full px-3 py-2 text-left text-sm hover:bg-bg-alt focus:bg-bg-alt focus:outline-none"
                        onClick={() => {
                          setSelectedUser(u)
                          setUserSearch('')
                          setUserDropdownOpen(false)
                        }}
                        data-testid={`add-member-user-option-${u.bizKey}`}
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
                  value={inviteRoleId ?? ''}
                  onValueChange={(v) => setInviteRoleId(v)}
                >
                  <SelectTrigger data-testid="add-member-role-select">
                    <SelectValue placeholder="选择角色" />
                  </SelectTrigger>
                  <SelectContent>
                    {roles.map((role) => (
                      <SelectItem key={role.bizKey} value={role.bizKey}>
                        {role.roleName}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={closeAddMemberDialog}>取消</Button>
            <Button
              onClick={() => inviteMutation.mutate()}
              disabled={!selectedUser || inviteRoleId == null || inviteMutation.isPending}
              data-testid="add-member-submit-btn"
            >
              确认添加
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
