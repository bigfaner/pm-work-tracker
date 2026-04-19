import { useState, useCallback, useMemo } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { getMainItemApi, updateMainItemApi } from '@/api/mainItems'
import { createSubItemApi } from '@/api/subItems'
import { changeSubItemStatusApi } from '@/api/subItems'
import { listMembersApi } from '@/api/teams'
import type { SubItem } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
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
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { PermissionGuard } from '@/components/PermissionGuard'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import StatusBadge from '@/components/shared/StatusBadge'
import PriorityBadge from '@/components/shared/PriorityBadge'
import ProgressBar from '@/components/shared/ProgressBar'
import UserAvatar from '@/components/shared/UserAvatar'

// --- Constants ---

const STATUS_OPTIONS = ['未开始', '进行中', '待评审', '已完成', '已关闭', '阻塞中', '延期']

// --- Main Component ---

export default function MainItemDetailPage() {
  const { mainItemId } = useParams<{ mainItemId: string }>()
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()
  const itemId = Number(mainItemId)

  // State
  const [expanded, setExpanded] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [createSubOpen, setCreateSubOpen] = useState(false)
  const [editForm, setEditForm] = useState({ title: '', priority: '', assigneeId: '', status: '', expectedEndDate: '', actualEndDate: '' })
  const [subForm, setSubForm] = useState({ title: '', priority: '', assigneeId: '', startDate: '', expectedEndDate: '', description: '' })

  // --- Data fetching ---

  const { data: item, isLoading } = useQuery({
    queryKey: ['mainItem', teamId, itemId],
    queryFn: () => getMainItemApi(teamId!, itemId),
    enabled: !!teamId && !!itemId,
  })

  const { data: members } = useQuery({
    queryKey: ['members', teamId],
    queryFn: () => listMembersApi(teamId!),
    enabled: !!teamId,
  })

  const memberName = useCallback(
    (assigneeId: number | null) => {
      if (!assigneeId || !members) return '-'
      const m = members.find((m) => m.userId === assigneeId)
      return m ? m.displayName : 'Unknown'
    },
    [members],
  )

  // Populate edit form when data loads
  useMemo(() => {
    if (item) {
      setEditForm({
        title: item.title,
        priority: item.priority,
        assigneeId: item.assigneeId ? String(item.assigneeId) : '',
        status: item.status,
        expectedEndDate: item.expectedEndDate || '',
        actualEndDate: item.actualEndDate || '',
      })
    }
  }, [item])

  // --- Mutations ---

  const updateMutation = useMutation({
    mutationFn: (req: { title?: string; priority?: string; assigneeId?: number | null; status?: string; expectedEndDate?: string | null; actualEndDate?: string | null }) =>
      updateMainItemApi(teamId!, itemId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItem', teamId, itemId] })
      setEditOpen(false)
    },
  })

  const createSubMutation = useMutation({
    mutationFn: (req: { title: string; priority: string; assigneeId: number; startDate?: string; expectedEndDate?: string; description?: string }) =>
      createSubItemApi(teamId!, itemId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItem', teamId, itemId] })
      setCreateSubOpen(false)
      setSubForm({ title: '', priority: '', assigneeId: '', startDate: '', expectedEndDate: '', description: '' })
    },
  })

  const statusChangeMutation = useMutation({
    mutationFn: ({ subItemId, status }: { subItemId: number; status: string }) =>
      changeSubItemStatusApi(teamId!, subItemId, { status }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItem', teamId, itemId] })
    },
  })

  // --- Handlers ---

  const handleEdit = useCallback(() => {
    if (!editForm.title.trim()) return
    updateMutation.mutate({
      title: editForm.title.trim(),
      priority: editForm.priority,
      assigneeId: editForm.assigneeId ? Number(editForm.assigneeId) : null,
      status: editForm.status,
      expectedEndDate: editForm.expectedEndDate || null,
      actualEndDate: editForm.actualEndDate || null,
    })
  }, [editForm, updateMutation])

  const handleCreateSub = useCallback(() => {
    if (!subForm.title.trim() || !subForm.priority || !subForm.assigneeId || !subForm.startDate || !subForm.expectedEndDate) return
    createSubMutation.mutate({
      title: subForm.title.trim(),
      priority: subForm.priority,
      assigneeId: subForm.assigneeId ? Number(subForm.assigneeId) : 0,
      startDate: subForm.startDate,
      expectedEndDate: subForm.expectedEndDate,
      ...(subForm.description && { description: subForm.description }),
    })
  }, [subForm, createSubMutation])

  const formatDate = (date: string | null) => {
    if (!date) return '-'
    return date
  }

  // --- Computed ---

  const subItems: SubItem[] = (item as any)?.subItems || []
  const completedCount = subItems.filter((s) => s.status === '已完成').length
  const completion = item?.completion ?? 0

  // --- Render ---

  if (!teamId) return <div className="p-6 text-tertiary">请先选择团队</div>

  return (
    <div data-testid="main-item-detail-page">
      {isLoading ? (
        <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
      ) : !item ? (
        <div className="py-8 text-center text-tertiary text-sm">事项不存在</div>
      ) : (
        <>
          {/* Breadcrumb */}
          <Breadcrumb className="mb-4">
            <BreadcrumbItem href="/items">事项清单</BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem isCurrent>{item.title}</BreadcrumbItem>
          </Breadcrumb>

          {/* Title Bar */}
          <div className="flex items-center gap-3 mb-6 flex-wrap">
            <Badge variant="default" className="font-mono">{item.code}</Badge>
            <h1 className="text-xl font-semibold text-primary m-0">{item.title}</h1>
            <PriorityBadge priority={item.priority} />
            <StatusBadge status={item.status} />
            <div className="flex-1" />
            <PermissionGuard code="main_item:update">
              <Button variant="secondary" onClick={() => setEditOpen(true)}>
                <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                编辑
              </Button>
            </PermissionGuard>
          </div>

          {/* Info Grid */}
          <Card className="mb-5">
            <CardContent>
              <div className="grid grid-cols-2 gap-4 mb-4">
                <div>
                  <div className="text-xs text-tertiary mb-1">负责人</div>
                  <div className="flex items-center gap-2">
                    <UserAvatar name={memberName(item.assigneeId)} size="sm" />
                    <span className="text-[13px] font-medium">{memberName(item.assigneeId)}</span>
                  </div>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">预期完成时间</div>
                  <span className="text-[13px] font-medium text-error">{formatDate(item.expectedEndDate)}</span>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">开始时间</div>
                  <span className="text-[13px]">{formatDate(item.startDate)}</span>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">实际完成时间</div>
                  <span className="text-[13px] text-tertiary">{formatDate(item.actualEndDate)}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Progress & Summary Card */}
          <Card className="mb-5">
            <CardHeader
              className="cursor-pointer select-none"
              onClick={() => setExpanded(!expanded)}
            >
              <div className="flex items-center gap-4">
                {/* Circular progress SVG */}
                <svg width="56" height="56" viewBox="0 0 80 80">
                  <circle cx="40" cy="40" r="34" fill="none" stroke="#e2e8f0" strokeWidth="6" />
                  <circle
                    cx="40" cy="40" r="34" fill="none" stroke="var(--color-primary-500, #3b82f6)"
                    strokeWidth="6" strokeLinecap="round"
                    strokeDasharray="213.6"
                    strokeDashoffset={213.6 * (1 - completion / 100)}
                    transform="rotate(-90 40 40)"
                  />
                  <text
                    x="40" y="40"
                    textAnchor="middle" dominantBaseline="central"
                    fontSize="18" fontWeight="600" fill="var(--color-primary, #1e293b)"
                  >
                    {completion}%
                  </text>
                </svg>
                <div>
                  <h3 className="text-sm font-semibold text-primary m-0">进度与汇总</h3>
                  <div className="text-[13px] text-secondary mt-0.5">
                    已完成 {completedCount} 个子事项 / 共 {subItems.length} 个子事项
                  </div>
                </div>
              </div>
              <svg
                className={`w-4 h-4 text-tertiary transition-transform ${expanded ? 'rotate-90' : ''}`}
                fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2"
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
              </svg>
            </CardHeader>
            {expanded && (
              <CardContent>
                <ProgressBar value={completion} size="default" showPercentage className="mb-5" />
                <div className="grid grid-cols-2 gap-5">
                  <div>
                    <div className="text-[13px] font-medium mb-2 text-emerald-600">成果汇总</div>
                    <ul className="text-[13px] text-secondary pl-4 list-disc leading-relaxed">
                      {(item as any).achievements?.map((a: string, i: number) => (
                        <li key={i}>{a}</li>
                      )) || <li className="text-tertiary">暂无</li>}
                    </ul>
                  </div>
                  <div>
                    <div className="text-[13px] font-medium mb-2 text-red-600">卡点汇总</div>
                    <ul className="text-[13px] text-secondary pl-4 list-disc leading-relaxed">
                      {(item as any).blockers?.map((b: string, i: number) => (
                        <li key={i}>{b}</li>
                      )) || <li className="text-tertiary">暂无</li>}
                    </ul>
                  </div>
                </div>
              </CardContent>
            )}
          </Card>

          {/* Sub-items Table */}
          <Card>
            <CardHeader>
              <h3 className="text-sm font-semibold text-primary m-0">子事项列表</h3>
              <Button size="sm" onClick={() => setCreateSubOpen(true)}>+ 新增子事项</Button>
            </CardHeader>
            {subItems.length > 0 && (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>编号</TableHead>
                    <TableHead>标题</TableHead>
                    <TableHead>负责人</TableHead>
                    <TableHead>完成度</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>开始时间</TableHead>
                    <TableHead>预期完成</TableHead>
                    <TableHead>实际完成</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {subItems.map((sub) => {
                    const subCode = `SI-${String(item.id).padStart(3, '0')}-${String(sub.id).slice(-2)}`
                    return (
                      <TableRow key={sub.id}>
                        <TableCell>
                          <Badge variant="default" className="font-mono text-[11px]">{subCode}</Badge>
                        </TableCell>
                        <TableCell>
                          <Link
                            to={`/items/${item.id}/sub/${sub.id}`}
                            className="text-[13px] font-medium text-primary hover:text-primary-600"
                          >
                            {sub.title}
                          </Link>
                        </TableCell>
                        <TableCell>{memberName(sub.assigneeId)}</TableCell>
                        <TableCell>
                          <ProgressBar value={sub.completion} size="sm" showPercentage />
                        </TableCell>
                        <TableCell>
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <button className="focus:outline-none">
                                <StatusBadge status={sub.status} className="cursor-pointer" />
                              </button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent>
                              {STATUS_OPTIONS.map((status) => (
                                <DropdownMenuItem
                                  key={status}
                                  className="text-[13px]"
                                  onSelect={() => statusChangeMutation.mutate({ subItemId: sub.id, status })}
                                >
                                  {status}
                                </DropdownMenuItem>
                              ))}
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                        <TableCell className="text-xs">{formatDate(sub.startDate)}</TableCell>
                        <TableCell className="text-xs">{formatDate(sub.expectedEndDate)}</TableCell>
                        <TableCell className="text-xs text-tertiary">{formatDate(sub.actualEndDate)}</TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            )}
          </Card>

          {/* Edit Dialog */}
          <Dialog open={editOpen} onOpenChange={setEditOpen}>
            <DialogContent size="md">
              <DialogHeader>
                <DialogTitle>编辑主事项</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    标题 <span className="text-error">*</span>
                  </label>
                  <Input
                    value={editForm.title}
                    onChange={(e) => setEditForm((f) => ({ ...f, title: e.target.value }))}
                  />
                </div>
                <div className="grid grid-cols-2 gap-3 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">优先级</label>
                    <Select value={editForm.priority} onValueChange={(v) => setEditForm((f) => ({ ...f, priority: v }))}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="P1">P1</SelectItem>
                        <SelectItem value="P2">P2</SelectItem>
                        <SelectItem value="P3">P3</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">负责人</label>
                    <Select value={editForm.assigneeId || '_none'} onValueChange={(v) => setEditForm((f) => ({ ...f, assigneeId: v === '_none' ? '' : v }))}>
                      <SelectTrigger><SelectValue placeholder="选择负责人" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="_none">不指定</SelectItem>
                        {members?.map((m) => (
                          <SelectItem key={m.userId} value={String(m.userId)}>{m.displayName}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">状态</label>
                  <Select value={editForm.status} onValueChange={(v) => setEditForm((f) => ({ ...f, status: v }))}>
                    <SelectTrigger><SelectValue /></SelectTrigger>
                    <SelectContent>
                      {STATUS_OPTIONS.map((s) => (
                        <SelectItem key={s} value={s}>{s}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">预期完成时间</label>
                    <Input
                      type="date"
                      value={editForm.expectedEndDate}
                      onChange={(e) => setEditForm((f) => ({ ...f, expectedEndDate: e.target.value }))}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">实际完成时间</label>
                    <Input
                      type="date"
                      value={editForm.actualEndDate}
                      onChange={(e) => setEditForm((f) => ({ ...f, actualEndDate: e.target.value }))}
                    />
                  </div>
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setEditOpen(false)}>取消</Button>
                <Button onClick={handleEdit} disabled={!editForm.title.trim() || updateMutation.isPending}>保存</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Create Sub-item Dialog */}
          <Dialog open={createSubOpen} onOpenChange={setCreateSubOpen}>
            <DialogContent size="md">
              <DialogHeader>
                <DialogTitle>新增子事项</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    标题 <span className="text-error">*</span>
                  </label>
                  <Input
                    placeholder="请输入子事项标题"
                    value={subForm.title}
                    onChange={(e) => setSubForm((f) => ({ ...f, title: e.target.value }))}
                  />
                </div>
                <div className="grid grid-cols-2 gap-3 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">
                      优先级 <span className="text-error">*</span>
                    </label>
                    <Select value={subForm.priority} onValueChange={(v) => setSubForm((f) => ({ ...f, priority: v }))}>
                      <SelectTrigger><SelectValue placeholder="请选择优先级" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="P1">P1</SelectItem>
                        <SelectItem value="P2">P2</SelectItem>
                        <SelectItem value="P3">P3</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">
                      负责人 <span className="text-error">*</span>
                    </label>
                    <Select value={subForm.assigneeId || '_none'} onValueChange={(v) => setSubForm((f) => ({ ...f, assigneeId: v === '_none' ? '' : v }))}>
                      <SelectTrigger><SelectValue placeholder="选择负责人" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="_none">不指定</SelectItem>
                        {members?.map((m) => (
                          <SelectItem key={m.userId} value={String(m.userId)}>{m.displayName}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-3 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">
                      开始时间 <span className="text-error">*</span>
                    </label>
                    <Input
                      type="date"
                      value={subForm.startDate}
                      onChange={(e) => setSubForm((f) => ({ ...f, startDate: e.target.value }))}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">
                      预期完成时间 <span className="text-error">*</span>
                    </label>
                    <Input
                      type="date"
                      value={subForm.expectedEndDate}
                      onChange={(e) => setSubForm((f) => ({ ...f, expectedEndDate: e.target.value }))}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">描述</label>
                  <textarea
                    className="flex w-full rounded-md border border-border-dark bg-white px-3 py-2 text-[13px] text-primary shadow-sm placeholder:text-tertiary focus:outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-200"
                    rows={3}
                    placeholder="请输入子事项描述（可选）"
                    value={subForm.description}
                    onChange={(e) => setSubForm((f) => ({ ...f, description: e.target.value }))}
                  />
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setCreateSubOpen(false)}>取消</Button>
                <Button onClick={handleCreateSub} disabled={!subForm.title.trim() || !subForm.priority || !subForm.assigneeId || !subForm.startDate || !subForm.expectedEndDate || createSubMutation.isPending}>确认</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </>
      )}
    </div>
  )
}
