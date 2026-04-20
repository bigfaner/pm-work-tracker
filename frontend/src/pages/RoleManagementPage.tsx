import { useState, useCallback } from 'react'
import { Pencil, Trash2 } from 'lucide-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listRolesApi, deleteRoleApi } from '@/api/roles'
import type { Role } from '@/types'
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
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Pagination } from '@/components/ui/pagination'
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
  TooltipProvider,
} from '@/components/ui/tooltip'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import ConfirmDialog from '@/components/shared/ConfirmDialog'
import RoleEditDialog from '@/components/RoleEditDialog'
import PermissionBrowseDialog from '@/components/PermissionBrowseDialog'

const PAGE_SIZE = 20

export default function RoleManagementPage() {
  const qc = useQueryClient()

  // Filters
  const [searchText, setSearchText] = useState('')
  const [presetFilter, setPresetFilter] = useState('all')
  const [currentPage, setCurrentPage] = useState(1)

  // Dialogs
  const [editOpen, setEditOpen] = useState(false)
  const [editRoleId, setEditRoleId] = useState<number | null>(null)
  const [browseOpen, setBrowseOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteRole, setDeleteRole] = useState<Role | null>(null)
  const [deleteError, setDeleteError] = useState('')

  // Data
  const { data: rolesData, isLoading, isFetching } = useQuery({
    queryKey: ['roles', currentPage, searchText, presetFilter],
    queryFn: () =>
      listRolesApi({
        page: currentPage,
        pageSize: PAGE_SIZE,
        ...(searchText.trim() && { search: searchText.trim() }),
        ...(presetFilter !== 'all' && { isPreset: presetFilter }),
      }),
  })

  const roles = rolesData?.items || []
  const total = rolesData?.total || 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  // Handlers
  const handleSearchChange = useCallback((value: string) => {
    setSearchText(value)
    setCurrentPage(1)
  }, [])

  const handlePresetFilterChange = useCallback((value: string) => {
    setPresetFilter(value)
    setCurrentPage(1)
  }, [])

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: number) => deleteRoleApi(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['roles'] })
      setDeleteOpen(false)
      setDeleteRole(null)
      setDeleteError('')
    },
    onError: (err: any) => {
      const code = err?.response?.data?.code
      if (code === 'ERR_PRESET_ROLE_IMMUTABLE') {
        setDeleteError('预置角色不可删除')
      } else if (code === 'ERR_ROLE_IN_USE') {
        setDeleteError('角色正在被使用，无法删除')
      } else {
        setDeleteError('删除失败，请稍后重试')
      }
    },
  })

  const openEdit = useCallback((role: Role) => {
    setEditRoleId(role.id)
    setEditOpen(true)
  }, [])

  const openCreate = useCallback(() => {
    setEditRoleId(null)
    setEditOpen(true)
  }, [])

  const openDelete = useCallback((role: Role) => {
    setDeleteRole(role)
    setDeleteError('')
    setDeleteOpen(true)
  }, [])

  const handleDeleteConfirm = useCallback(() => {
    if (deleteRole) {
      deleteMutation.mutate(deleteRole.id)
    }
  }, [deleteRole, deleteMutation])

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr)
    return `${d.getFullYear()}/${String(d.getMonth() + 1).padStart(2, '0')}/${String(d.getDate()).padStart(2, '0')}`
  }

  return (
    <TooltipProvider>
      <div data-testid="role-management-page">
        {/* Breadcrumb */}
        <Breadcrumb className="mb-4">
          <BreadcrumbItem href="/items">首页</BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem isCurrent>角色管理</BreadcrumbItem>
        </Breadcrumb>

        {/* Page Header */}
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-xl font-semibold text-primary">角色管理</h1>
          <div className="flex gap-2">
            <Button variant="secondary" size="sm" onClick={() => setBrowseOpen(true)}>
              权限列表
            </Button>
            <Button size="sm" onClick={openCreate}>
              <svg
                width="16"
                height="16"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M12 4v16m8-8H4"
                />
              </svg>
              创建角色
            </Button>
          </div>
        </div>

        {/* Filter Bar */}
        <div className="flex items-center gap-3 mb-4 flex-wrap">
          <Input
            placeholder="搜索角色名称"
            value={searchText}
            onChange={(e) => handleSearchChange(e.target.value)}
            className="w-[240px]"
          />
          <Select value={presetFilter} onValueChange={handlePresetFilterChange}>
            <SelectTrigger className="w-[160px]">
              <SelectValue placeholder="类型筛选" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部类型</SelectItem>
              <SelectItem value="preset">预置角色</SelectItem>
              <SelectItem value="custom">自定义角色</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Table */}
        {isLoading || (isFetching && !rolesData) ? (
          <div className="rounded-xl border border-border bg-white shadow-sm p-8">
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-10 bg-bg-alt rounded animate-pulse" />
              ))}
            </div>
          </div>
        ) : roles.length === 0 ? (
          <div className="py-12 text-center">
            <p className="text-tertiary text-sm">
              {searchText || presetFilter !== 'all'
                ? '没有匹配的角色'
                : '暂无自定义角色'}
            </p>
            {!searchText && presetFilter === 'all' && (
              <Button className="mt-3" size="sm" onClick={openCreate}>
                创建角色
              </Button>
            )}
          </div>
        ) : (
          <div className="rounded-xl border border-border bg-white shadow-sm">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>角色名称</TableHead>
                  <TableHead>描述</TableHead>
                  <TableHead>权限数量</TableHead>
                  <TableHead>使用人数</TableHead>
                  <TableHead>类型</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {roles.map((role) => (
                  <TableRow key={role.id}>
                    <TableCell>
                      <span className="font-medium text-primary">{role.name}</span>
                    </TableCell>
                    <TableCell>
                      <span className="text-secondary">
                        {role.description || '-'}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span>{role.permissionCount}</span>
                    </TableCell>
                    <TableCell>
                      <span>{role.memberCount}</span>
                    </TableCell>
                    <TableCell>
                      <Badge variant={role.isPreset ? 'primary' : 'default'}>
                        {role.isPreset ? '预置' : '自定义'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <span className="text-tertiary">{formatDate(role.createdAt)}</span>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openEdit(role)}
                        >
                          <Pencil className="w-3.5 h-3.5" />
                          编辑
                        </Button>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <span>
                              <Button
                                variant="ghost"
                                size="sm"
                                disabled={role.isPreset || role.memberCount > 0}
                                onClick={() => openDelete(role)}
                              >
                                <Trash2 className="w-3.5 h-3.5" />
                                删除
                              </Button>
                            </span>
                          </TooltipTrigger>
                          {(role.isPreset || role.memberCount > 0) && (
                            <TooltipContent>
                              {role.isPreset
                                ? '预置角色不可删除'
                                : `该角色正在被 ${role.memberCount} 个用户使用`}
                            </TooltipContent>
                          )}
                        </Tooltip>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>

            {/* Pagination */}
            <div className="flex items-center justify-center px-5 py-3 border-t border-border">
              <div className="flex items-center gap-3">
                <span className="text-[13px] text-tertiary">共 {total} 条</span>
                {totalPages > 1 && (
                  <Pagination
                    currentPage={currentPage}
                    totalPages={totalPages}
                    onPageChange={setCurrentPage}
                  />
                )}
              </div>
            </div>
          </div>
        )}

        {/* Edit/Create Dialog */}
        <RoleEditDialog
          open={editOpen}
          onOpenChange={setEditOpen}
          roleId={editRoleId}
        />

        {/* Permission Browse Dialog */}
        <PermissionBrowseDialog open={browseOpen} onOpenChange={setBrowseOpen} />

        {/* Delete Confirm Dialog */}
        <ConfirmDialog
          open={deleteOpen}
          onOpenChange={(v) => { setDeleteOpen(v); if (!v) setDeleteError('') }}
          title="删除角色"
          description={
            deleteRole
              ? `确定要删除角色"${deleteRole.name}"吗？此操作不可撤销。`
              : ''
          }
          confirmLabel="确认删除"
          confirmVariant="danger"
          onConfirm={handleDeleteConfirm}
        />
        {deleteError && (
          <p className="text-sm text-error text-center mt-2">{deleteError}</p>
        )}
      </div>
    </TooltipProvider>
  )
}
