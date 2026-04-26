import { useState, useCallback, useMemo } from 'react'
import { Pencil, ToggleRight, ToggleLeft, RefreshCw } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listUsersApi,
  createUserApi,
  updateUserApi,
  toggleUserStatusApi,
  listAdminTeamsApi,
} from '@/api/admin'
import type { AdminUser, CreateUserReq, UpdateUserReq } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
  DialogDescription,
} from '@/components/ui/dialog'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Pagination, PaginationPageSize } from '@/components/ui/pagination'
import PaginationBar from '@/components/shared/PaginationBar'
import UserAvatar from '@/components/shared/UserAvatar'
import { useToast } from '@/components/ui/toast'

// --- Constants ---

const DEFAULT_PAGE_SIZE = 10

// --- Main Component ---

export default function UserManagementPage() {
  const qc = useQueryClient()
  const { addToast } = useToast()

  // Filter state
  const [searchText, setSearchText] = useState('')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE)

  // Create dialog
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm, setCreateForm] = useState<CreateUserReq & { email: string }>({
    username: '',
    displayName: '',
    email: '',
  })
  const [createError, setCreateError] = useState('')

  // Password display dialog
  const [passwordOpen, setPasswordOpen] = useState(false)
  const [initialPassword, setInitialPassword] = useState('')

  // Edit dialog
  const [editOpen, setEditOpen] = useState(false)
  const [editUserId, setEditUserId] = useState<string | null>(null)
  const [editForm, setEditForm] = useState<UpdateUserReq & { displayName: string; email: string }>({
    displayName: '',
    email: '',
  })

  // Toggle status dialog
  const [statusOpen, setStatusOpen] = useState(false)
  const [statusUser, setStatusUser] = useState<AdminUser | null>(null)
  const [statusError, setStatusError] = useState('')

  // --- Data fetching ---

  const { data: usersData, isLoading, isFetching, refetch } = useQuery({
    queryKey: ['adminUsers', currentPage, pageSize, searchText],
    queryFn: () =>
      listUsersApi({
        page: currentPage,
        pageSize,
        ...(searchText.trim() && { search: searchText.trim() }),
      }),
  })

  const { data: teamsData } = useQuery({
    queryKey: ['adminTeams'],
    queryFn: () => listAdminTeamsApi(1, 100),
  })

  const teams = teamsData?.items || []
  const users = usersData?.items || []
  const total = usersData?.total || 0
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  // Reset page when filters change
  const handleSearchChange = useCallback((value: string) => {
    setSearchText(value)
    setCurrentPage(1)
  }, [])

  const handlePageSizeChange = useCallback((size: number) => {
    setPageSize(size)
    setCurrentPage(1)
  }, [])

  // --- Mutations ---

  const createMutation = useMutation({
    mutationFn: (req: CreateUserReq) => createUserApi(req),
    onSuccess: (resp) => {
      qc.invalidateQueries({ queryKey: ['adminUsers'] })
      setCreateOpen(false)
      setCreateForm({ username: '', displayName: '', email: '' })
      setCreateError('')
      setInitialPassword(resp.initialPassword)
      setPasswordOpen(true)
    },
    onError: (err: any) => {
      const code = err?.response?.data?.code
      if (code === 'USER_EXISTS') {
        setCreateError('该账号名已存在')
      } else if (code === 'VALIDATION_ERROR') {
        setCreateError('请检查输入信息')
      } else {
        setCreateError('创建失败，请稍后重试')
      }
    },
  })

  const editMutation = useMutation({
    mutationFn: ({ userId, req }: { userId: string; req: UpdateUserReq }) =>
      updateUserApi(userId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['adminUsers'] })
      setEditOpen(false)
      setEditUserId(null)
    },
  })

  const statusMutation = useMutation({
    mutationFn: ({ userId, status }: { userId: string; status: 'enabled' | 'disabled' }) =>
      toggleUserStatusApi(userId, { status }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['adminUsers'] })
      setStatusOpen(false)
      setStatusUser(null)
      setStatusError('')
    },
    onError: (err: any) => {
      const code = err?.response?.data?.code
      if (code === 'CANNOT_DISABLE_SELF') {
        setStatusError('不能禁用自己')
      } else {
        setStatusError('操作失败，请稍后重试')
      }
    },
  })

  // --- Handlers ---

  const handleCreate = useCallback(() => {
    setCreateError('')
    if (!createForm.username.trim() || !createForm.displayName.trim()) {
      setCreateError('请填写必填字段')
      return
    }
    if (createForm.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(createForm.email)) {
      setCreateError('邮箱格式不正确')
      return
    }
    createMutation.mutate({
      username: createForm.username.trim(),
      displayName: createForm.displayName.trim(),
      ...(createForm.email.trim() && { email: createForm.email.trim() }),
      ...(createForm.teamKey && { teamKey: createForm.teamKey }),
    })
  }, [createForm, createMutation])

  const openEdit = useCallback((user: AdminUser) => {
    setEditUserId(user.bizKey)
    setEditForm({
      displayName: user.displayName,
      email: user.email || '',
    })
    setEditOpen(true)
  }, [])

  const handleEdit = useCallback(() => {
    if (!editUserId) return
    editMutation.mutate({
      userId: editUserId,
      req: {
        displayName: editForm.displayName.trim(),
        ...(editForm.email.trim() && { email: editForm.email.trim() }),
      },
    })
  }, [editUserId, editForm, editMutation])

  const openToggleStatus = useCallback((user: AdminUser) => {
    setStatusUser(user)
    setStatusError('')
    setStatusOpen(true)
  }, [])

  const handleToggleStatus = useCallback(() => {
    if (!statusUser) return
    const newStatus = statusUser.userStatus === 'enabled' ? 'disabled' : 'enabled'
    statusMutation.mutate({ userId: statusUser.bizKey, status: newStatus })
  }, [statusUser, statusMutation])

  // --- Render ---

  return (
    <div data-testid="user-management-page">
      {/* Page Header */}
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-semibold text-primary">用户管理</h1>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          创建用户
        </Button>
      </div>

      {/* Filter Bar */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <Input
          placeholder="搜索用户名/姓名"
          value={searchText}
          onChange={(e) => handleSearchChange(e.target.value)}
          className="w-[240px]"
        />
        <Button variant="secondary" size="sm" onClick={async () => { await refetch(); addToast('数据已刷新', 'success') }} disabled={isFetching} data-testid="refresh-btn">
          <RefreshCw size={14} className={isFetching ? 'animate-spin' : ''} />
          刷新
        </Button>
      </div>

      {/* User Table */}
      {isLoading || (isFetching && !usersData) ? (
        <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
      ) : users.length === 0 ? (
        <div className="py-12 text-center">
          <p className="text-tertiary text-sm">暂无用户</p>
        </div>
      ) : (
        <div className="rounded-xl border border-border bg-white shadow-sm">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>姓名</TableHead>
                <TableHead>账号</TableHead>
                <TableHead>邮箱</TableHead>
                <TableHead>所属团队</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.map((user) => (
                <TableRow key={user.bizKey}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <UserAvatar name={user.displayName} size="sm" />
                      <span className="font-medium text-primary">{user.displayName}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className="font-mono text-xs">{user.username}</span>
                  </TableCell>
                  <TableCell>
                    <span className="text-[13px]">{user.email || '-'}</span>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {user.teams && user.teams.length > 0 ? (
                        user.teams.map((t) => (
                          <Badge key={t.bizKey} variant="primary">{t.name}</Badge>
                        ))
                      ) : (
                        <span className="text-tertiary text-xs">-</span>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant={user.userStatus === 'enabled' ? 'success' : 'warning'}>
                      {user.userStatus === 'enabled' ? '启用' : '停用'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button variant="ghost" size="sm" className="text-primary-600" onClick={() => openEdit(user)}>
                        <Pencil className="w-3.5 h-3.5" />
                        编辑
                      </Button>
                      <Button variant="ghost" size="sm" className="text-primary-600" onClick={() => openToggleStatus(user)}>
                        {user.userStatus === 'enabled' ? (
                          <ToggleRight className="w-3.5 h-3.5" />
                        ) : (
                          <ToggleLeft className="w-3.5 h-3.5" />
                        )}
                        修改状态
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          {/* Pagination */}
          <PaginationBar
            currentPage={currentPage}
            totalPages={totalPages}
            total={total}
            onPageChange={setCurrentPage}
            pageSize={pageSize}
            onPageSizeChange={handlePageSizeChange}
            pageSizeOptions={[10, 20, 50]}
          />
        </div>
      )}

      {/* Create User Dialog */}
      <Dialog open={createOpen} onOpenChange={(open) => { setCreateOpen(open); if (!open) setCreateError('') }}>
        <DialogContent size="md">
          <DialogHeader>
            <DialogTitle>创建用户</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className="flex flex-col gap-4">
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  姓名 <span className="text-error">*</span>
                </label>
                <Input
                  placeholder="请输入姓名"
                  value={createForm.displayName}
                  onChange={(e) => setCreateForm((f) => ({ ...f, displayName: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  账号 <span className="text-error">*</span>
                </label>
                <Input
                  placeholder="请输入账号"
                  value={createForm.username}
                  onChange={(e) => setCreateForm((f) => ({ ...f, username: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">邮箱</label>
                <Input
                  type="email"
                  placeholder="请输入邮箱"
                  value={createForm.email}
                  onChange={(e) => setCreateForm((f) => ({ ...f, email: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">所属团队</label>
                <Select
                  value={createForm.teamKey || '_none'}
                  onValueChange={(v) =>
                    setCreateForm((f) => ({ ...f, teamKey: v === '_none' ? undefined : v }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择团队" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_none">不指定</SelectItem>
                    {teams.map((t) => (
                      <SelectItem key={t.bizKey} value={t.bizKey}>{t.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
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
              disabled={!createForm.username.trim() || !createForm.displayName.trim() || createMutation.isPending}
            >
              确认创建
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Initial Password Dialog */}
      <Dialog open={passwordOpen} onOpenChange={setPasswordOpen}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>用户创建成功</DialogTitle>
            <DialogDescription>请妥善保管，关闭后无法再次查看</DialogDescription>
          </DialogHeader>
          <DialogBody>
            <div className="bg-bg-alt rounded-lg p-3 text-center">
              <p className="text-xs text-tertiary mb-1">初始密码</p>
              <p className="font-mono text-lg font-semibold text-primary" data-testid="initial-password">
                {initialPassword}
              </p>
            </div>
          </DialogBody>
          <DialogFooter>
            <Button onClick={() => setPasswordOpen(false)}>我知道了</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit User Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent size="md">
          <DialogHeader>
            <DialogTitle>编辑用户</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className="flex flex-col gap-4">
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  姓名 <span className="text-error">*</span>
                </label>
                <Input
                  value={editForm.displayName}
                  onChange={(e) => setEditForm((f) => ({ ...f, displayName: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">邮箱</label>
                <Input
                  type="email"
                  value={editForm.email}
                  onChange={(e) => setEditForm((f) => ({ ...f, email: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">所属团队</label>
                <Select
                  value={editForm.teamKey || '_none'}
                  onValueChange={(v) =>
                    setEditForm((f) => ({ ...f, teamKey: v === '_none' ? undefined : v }))
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择团队" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_none">不指定</SelectItem>
                    {teams.map((t) => (
                      <SelectItem key={t.bizKey} value={t.bizKey}>{t.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setEditOpen(false)}>取消</Button>
            <Button onClick={handleEdit} disabled={editMutation.isPending}>保存修改</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Toggle Status Dialog */}
      <Dialog open={statusOpen} onOpenChange={(open) => { setStatusOpen(open); if (!open) setStatusError('') }}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>修改用户状态</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className="flex flex-col gap-4">
              <div>
                <p className="text-sm text-tertiary mb-1">用户</p>
                <p className="font-medium text-primary">{statusUser?.displayName}</p>
              </div>
              <div>
                <p className="text-sm text-tertiary mb-1">当前状态</p>
                <Badge variant={statusUser?.userStatus === 'enabled' ? 'success' : 'warning'}>
                  {statusUser?.userStatus === 'enabled' ? '启用' : '停用'}
                </Badge>
              </div>
              <div>
                <p className="text-sm text-tertiary mb-1">新状态</p>
                <Badge variant={statusUser?.userStatus === 'enabled' ? 'warning' : 'success'}>
                  {statusUser?.userStatus === 'enabled' ? '停用' : '启用'}
                </Badge>
              </div>
              {statusUser?.userStatus === 'enabled' && (
                <div className="bg-warning-bg border border-warning-text/20 rounded-lg p-3 text-sm text-warning-text">
                  <strong>注意：</strong>禁用用户后该用户将无法登录系统，但数据不会被删除。
                </div>
              )}
            </div>
            {statusError && (
              <p className="mt-3 text-sm text-error">{statusError}</p>
            )}
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => { setStatusOpen(false); setStatusError('') }}>
              取消
            </Button>
            <Button onClick={handleToggleStatus} disabled={statusMutation.isPending}>
              确认修改
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
