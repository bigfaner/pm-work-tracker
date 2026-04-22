import { useState, useCallback, useMemo, useEffect } from 'react'
import { Pencil, Plus } from 'lucide-react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { getMainItemApi, updateMainItemApi, changeMainItemStatusApi, getMainItemTransitionsApi } from '@/api/mainItems'
import { createSubItemApi, updateSubItemApi, changeSubItemStatusApi, getSubItemTransitionsApi } from '@/api/subItems'
import { appendProgressApi } from '@/api/progress'
import { listMembersApi } from '@/api/teams'
import type { SubItem } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { DateInput } from '@/components/ui/date-input'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { PrioritySelectItems } from '@/components/shared/PrioritySelect'
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
import { STATUS_OPTIONS, MAIN_ITEM_STATUSES, SUB_ITEM_STATUSES } from '@/lib/status'
import { getStatusName, isOverdue } from '@/lib/status'
import { useToast } from '@/components/ui/toast'

// --- Main Component ---

export default function MainItemDetailPage() {
  const { mainItemId } = useParams<{ mainItemId: string }>()
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()
  const itemId = Number(mainItemId)

  // State
  const [expanded, setExpanded] = useState(false)
  const today = () => new Date().toISOString().slice(0, 10)

  const [editOpen, setEditOpen] = useState(false)
  const [createSubOpen, setCreateSubOpen] = useState(false)
  const [editForm, setEditForm] = useState({ title: '', priority: '', assigneeId: '', expectedEndDate: '', description: '' })
  const [subForm, setSubForm] = useState({ title: '', priority: 'P2', assigneeId: '', startDate: today(), expectedEndDate: '', description: '' })

  const [editSubOpen, setEditSubOpen] = useState(false)
  const [editSubTarget, setEditSubTarget] = useState<SubItem | null>(null)
  const [editSubForm, setEditSubForm] = useState({ title: '', priority: '', expectedEndDate: '', description: '' })

  const [appendProgressOpen, setAppendProgressOpen] = useState(false)
  const [appendProgressTarget, setAppendProgressTarget] = useState<SubItem | null>(null)
  const [appendProgressForm, setAppendProgressForm] = useState({ completion: '', achievement: '', blocker: '' })

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
        expectedEndDate: item.expectedEndDate || '',
        description: item.description || '',
      })
    }
  }, [item])

  // --- Mutations ---

  const updateMutation = useMutation({
    mutationFn: (req: { title?: string; priority?: string; assigneeId?: number | null; expectedEndDate?: string | null; actualEndDate?: string | null; description?: string }) =>
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
      setSubForm({ title: '', priority: 'P2', assigneeId: '', startDate: today(), expectedEndDate: '', description: '' })
    },
  })

  const statusChangeMutation = useMutation({
    mutationFn: ({ subItemId, status }: { subItemId: number; status: string }) =>
      changeSubItemStatusApi(teamId!, subItemId, { status }),
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ['mainItem', teamId, itemId] })
    },
  })

  const updateSubMutation = useMutation({
    mutationFn: ({ subId, req }: { subId: number; req: { title?: string; priority?: string; expectedEndDate?: string; description?: string } }) =>
      updateSubItemApi(teamId!, subId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItem', teamId, itemId] })
      setEditSubOpen(false)
      setEditSubTarget(null)
    },
  })

  const appendProgressMutation = useMutation({
    mutationFn: ({ subId, req }: { subId: number; req: { completion: number; achievement?: string; blocker?: string } }) =>
      appendProgressApi(teamId!, subId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItem', teamId, itemId] })
      setAppendProgressOpen(false)
      setAppendProgressTarget(null)
      setAppendProgressForm({ completion: '', achievement: '', blocker: '' })
    },
  })

  // --- Handlers ---

  const handleEdit = useCallback(() => {
    if (!editForm.title.trim()) return
    updateMutation.mutate({
      title: editForm.title.trim(),
      priority: editForm.priority,
      assigneeId: editForm.assigneeId ? Number(editForm.assigneeId) : null,
      expectedEndDate: editForm.expectedEndDate || null,
      description: editForm.description,
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

  const openEditSub = useCallback((sub: SubItem) => {
    setEditSubTarget(sub)
    setEditSubForm({
      title: sub.title,
      priority: sub.priority,
      expectedEndDate: sub.expectedEndDate || '',
      description: sub.description || '',
    })
    setEditSubOpen(true)
  }, [])

  const handleEditSub = useCallback(() => {
    if (!editSubTarget || !editSubForm.title.trim()) return
    updateSubMutation.mutate({
      subId: editSubTarget.id,
      req: {
        title: editSubForm.title.trim(),
        priority: editSubForm.priority,
        expectedEndDate: editSubForm.expectedEndDate || undefined,
        description: editSubForm.description,
      },
    })
  }, [editSubTarget, editSubForm, updateSubMutation])

  const openAppendProgress = useCallback((sub: SubItem) => {
    setAppendProgressTarget(sub)
    setAppendProgressForm({ completion: String(sub.completion), achievement: '', blocker: '' })
    setAppendProgressOpen(true)
  }, [])

  const handleAppendProgress = useCallback(() => {
    if (!appendProgressTarget || appendProgressForm.completion === '') return
    appendProgressMutation.mutate({
      subId: appendProgressTarget.id,
      req: {
        completion: Number(appendProgressForm.completion),
        ...(appendProgressForm.achievement && { achievement: appendProgressForm.achievement }),
        ...(appendProgressForm.blocker && { blocker: appendProgressForm.blocker }),
      },
    })
  }, [appendProgressTarget, appendProgressForm, appendProgressMutation])

  const formatDate = (date: string | null) => {
    if (!date) return '-'
    return date
  }

  // --- Computed ---

  const subItems: SubItem[] = (item as any)?.subItems || []
  const completedCount = subItems.filter((s) => s.status === 'completed').length
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
            <MainItemStatusDropdown itemId={item.id} currentStatus={item.status} />
            <div className="flex-1" />
            <PermissionGuard code="main_item:update">
              <Button variant="secondary" disabled={!!MAIN_ITEM_STATUSES[item.status as keyof typeof MAIN_ITEM_STATUSES]?.terminal} onClick={() => setEditOpen(true)}>
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
              <div className="grid grid-cols-4 gap-4 mb-4">
                <div>
                  <div className="text-xs text-tertiary mb-1">负责人</div>
                  <span className="text-[13px] font-medium">{memberName(item.assigneeId)}</span>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">开始时间</div>
                  <span className="text-[13px] font-medium">{formatDate(item.startDate)}</span>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">预期完成时间</div>
                  <div className="flex items-center gap-1.5">
                    <span className="text-[13px] font-medium">{formatDate(item.expectedEndDate)}</span>
                    {isOverdue(item.expectedEndDate ?? undefined, item.status) && (
                      <Badge variant="error">延期</Badge>
                    )}
                  </div>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">结束时间</div>
                  <span className="text-[13px] font-medium">{formatDate(item.actualEndDate)}</span>
                </div>
              </div>
              <div>
                <div className="text-xs text-tertiary mb-1">描述</div>
                <span className="text-[13px] text-secondary leading-relaxed">{item.description || '暂无描述'}</span>
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
                    <div className="text-[13px] font-medium mb-2 text-success-text">成果汇总</div>
                    <ul className="text-[13px] text-secondary pl-4 list-disc leading-relaxed">
                      {(item as any).achievements?.map((a: string, i: number) => (
                        <li key={i}>{a}</li>
                      )) || <li className="text-tertiary">暂无</li>}
                    </ul>
                  </div>
                  <div>
                    <div className="text-[13px] font-medium mb-2 text-error-text">卡点汇总</div>
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
              <Button size="sm" disabled={!!MAIN_ITEM_STATUSES[item.status as keyof typeof MAIN_ITEM_STATUSES]?.terminal} onClick={() => setCreateSubOpen(true)}>+ 新增子事项</Button>
            </CardHeader>
            {subItems.length > 0 && (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>编号</TableHead>
                    <TableHead>标题</TableHead>
                    <TableHead>负责人</TableHead>
                    <TableHead>进度</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>开始时间</TableHead>
                    <TableHead>预期完成时间</TableHead>
                    <TableHead>结束时间</TableHead>
                    <TableHead>操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {(() => {
                    const mainTerminal = (MAIN_ITEM_STATUSES as Record<string, { terminal: boolean }>)[item.status]?.terminal ?? false
                    return subItems.map((sub) => {
                    return (
                      <TableRow key={sub.id}>
                        <TableCell>
                          <Badge variant="default" className="font-mono text-[11px]">{sub.code}</Badge>
                        </TableCell>
                        <TableCell>
                          <Link
                            to={`/items/${item.id}/sub/${sub.id}`}
                            className="text-[13px] font-medium text-primary-600 hover:text-primary-700 hover:underline truncate block max-w-xs"
                            title={sub.title}
                          >
                            {sub.title}
                          </Link>
                        </TableCell>
                        <TableCell>{memberName(sub.assigneeId)}</TableCell>
                        <TableCell>
                          <span>{sub.completion}%</span>
                        </TableCell>
                        <TableCell>
                          <SubItemStatusDropdown
                            subId={sub.id}
                            currentStatus={sub.status}
                            onStatusChange={(status) => statusChangeMutation.mutate({ subItemId: sub.id, status })}
                          />
                        </TableCell>
                        <TableCell className="text-xs">{formatDate(sub.startDate)}</TableCell>
                        <TableCell className="text-xs">{formatDate(sub.expectedEndDate)}</TableCell>
                        <TableCell className="text-xs">{formatDate(sub.actualEndDate)}</TableCell>
                        <TableCell>
                          <div className="flex gap-0.5">
                            <Button variant="ghost" size="sm" disabled={mainTerminal} onClick={() => openEditSub(sub)}><Pencil size={12} />编辑</Button>
                            <Button variant="ghost" size="sm" disabled={mainTerminal} onClick={() => openAppendProgress(sub)}><Plus size={12} />追加进度</Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    )
                  })
                  })()}
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
                        <PrioritySelectItems />
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
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">预期完成时间</label>
                  <DateInput
                    value={editForm.expectedEndDate}
                    onChange={(e) => setEditForm((f) => ({ ...f, expectedEndDate: e.target.value }))}
                  />
                </div>
                <div className="mt-3">
                  <label className="block text-sm font-medium text-primary mb-1">描述</label>
                  <Textarea
                    rows={3}
                    value={editForm.description}
                    onChange={(e) => setEditForm((f) => ({ ...f, description: e.target.value }))}
                  />
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
                        <PrioritySelectItems />
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
                    <DateInput
                      value={subForm.startDate}
                      onChange={(e) => setSubForm((f) => ({ ...f, startDate: e.target.value }))}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">
                      预期完成时间 <span className="text-error">*</span>
                    </label>
                    <DateInput
                      value={subForm.expectedEndDate}
                      onChange={(e) => setSubForm((f) => ({ ...f, expectedEndDate: e.target.value }))}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">描述</label>
                  <Textarea
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

          {/* Edit Sub-item Dialog */}
          <Dialog open={editSubOpen} onOpenChange={setEditSubOpen}>
            <DialogContent size="md">
              <DialogHeader>
                <DialogTitle>编辑子事项</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    标题 <span className="text-error">*</span>
                  </label>
                  <Input
                    value={editSubForm.title}
                    onChange={(e) => setEditSubForm((f) => ({ ...f, title: e.target.value }))}
                  />
                </div>
                <div className="grid grid-cols-2 gap-3 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">优先级</label>
                    <Select value={editSubForm.priority} onValueChange={(v) => setEditSubForm((f) => ({ ...f, priority: v }))}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <PrioritySelectItems />
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">预期完成时间</label>
                    <DateInput
                      value={editSubForm.expectedEndDate}
                      onChange={(e) => setEditSubForm((f) => ({ ...f, expectedEndDate: e.target.value }))}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">描述</label>
                  <Textarea
                    rows={3}
                    value={editSubForm.description}
                    onChange={(e) => setEditSubForm((f) => ({ ...f, description: e.target.value }))}
                  />
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setEditSubOpen(false)}>取消</Button>
                <Button onClick={handleEditSub} disabled={!editSubForm.title.trim() || updateSubMutation.isPending}>保存</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Append Progress Dialog */}
          <Dialog open={appendProgressOpen} onOpenChange={setAppendProgressOpen}>
            <DialogContent size="sm">
              <DialogHeader>
                <DialogTitle>追加进度</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    完成度 (%) <span className="text-error">*</span>
                  </label>
                  <Input
                    type="number"
                    min={0}
                    max={100}
                    value={appendProgressForm.completion}
                    onChange={(e) => setAppendProgressForm((f) => ({ ...f, completion: e.target.value }))}
                  />
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">成果</label>
                  <Textarea
                    rows={2}
                    placeholder="本次进展成果（可选）"
                    value={appendProgressForm.achievement}
                    onChange={(e) => setAppendProgressForm((f) => ({ ...f, achievement: e.target.value }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">卡点</label>
                  <Textarea
                    rows={2}
                    placeholder="遇到的阻碍（可选）"
                    value={appendProgressForm.blocker}
                    onChange={(e) => setAppendProgressForm((f) => ({ ...f, blocker: e.target.value }))}
                  />
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setAppendProgressOpen(false)}>取消</Button>
                <Button onClick={handleAppendProgress} disabled={appendProgressForm.completion === '' || appendProgressMutation.isPending}>确认</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </>
      )}
    </div>
  )
}

// --- Main Item StatusDropdown ---

const MAIN_ITEM_TERMINAL_STATUSES = new Set(
  Object.entries(MAIN_ITEM_STATUSES)
    .filter(([, v]) => v.terminal)
    .map(([k]) => k)
)

function MainItemStatusDropdown({
  itemId,
  currentStatus,
}: {
  itemId: number
  currentStatus: string
}) {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()
  const { addToast } = useToast()
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingStatus, setPendingStatus] = useState<string | null>(null)
  const [open, setOpen] = useState(false)
  const [showTip, setShowTip] = useState(false)

  const { data: transitions = [], isFetched, isFetching } = useQuery({
    queryKey: ['mainItemTransitions', teamId, itemId],
    queryFn: () => getMainItemTransitionsApi(teamId!, itemId),
    enabled: !!teamId && open,
  })

  useEffect(() => {
    if (open && isFetched && !isFetching && transitions.length === 0) {
      setOpen(false)
      setShowTip(true)
      setTimeout(() => setShowTip(false), 2000)
    }
  }, [open, isFetched, isFetching, transitions.length])

  const statusChangeMutation = useMutation({
    mutationFn: ({ newStatus }: { newStatus: string }) =>
      changeMainItemStatusApi(teamId!, itemId, { status: newStatus }),
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ['mainItem', teamId, itemId] })
      qc.invalidateQueries({ queryKey: ['mainItemTransitions', teamId, itemId] })
      setOpen(false)
      setConfirmOpen(false)
      setPendingStatus(null)
      if ((data as any)?.linkageWarning) {
        addToast((data as any).linkageWarning, 'warning')
      }
    },
  })

  const handleSelect = useCallback((status: string) => {
    if (MAIN_ITEM_TERMINAL_STATUSES.has(status)) {
      setPendingStatus(status)
      setConfirmOpen(true)
    } else {
      statusChangeMutation.mutate({ newStatus: status })
    }
  }, [statusChangeMutation])

  const handleConfirm = useCallback(() => {
    if (pendingStatus) {
      statusChangeMutation.mutate({ newStatus: pendingStatus })
    }
  }, [pendingStatus, statusChangeMutation])

  return (
    <>
      <div className="relative inline-flex">
        {showTip && (
          <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1.5 whitespace-nowrap text-xs px-2 py-1 rounded-md bg-primary text-white shadow-md pointer-events-none z-50">
            暂无可用流转
          </div>
        )}
        <DropdownMenu open={open} onOpenChange={setOpen}>
          <DropdownMenuTrigger asChild>
            <button className="focus:outline-none">
              <StatusBadge status={currentStatus} className="cursor-pointer" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="min-w-0 w-auto">
            {transitions.map((status) => (
              <DropdownMenuItem
                key={status}
                className="text-[13px] justify-center"
              onSelect={(e) => {
                e.preventDefault()
                handleSelect(status)
              }}
            >
              {getStatusName(status) || status}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
      </div>

      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>确认变更状态</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <p className="text-sm text-secondary">
              确认将状态变更为「{getStatusName(pendingStatus || '') || pendingStatus}」？此操作可能不可逆。
            </p>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => { setConfirmOpen(false); setPendingStatus(null) }}>取消</Button>
            <Button onClick={handleConfirm} disabled={statusChangeMutation.isPending}>确认</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

// --- Sub-item StatusDropdown with transitions ---

const SUB_ITEM_TERMINAL_STATUSES = new Set(
  Object.entries(SUB_ITEM_STATUSES)
    .filter(([, v]) => v.terminal)
    .map(([k]) => k)
)

function SubItemStatusDropdown({
  subId,
  currentStatus,
  onStatusChange,
}: {
  subId: number
  currentStatus: string
  onStatusChange: (status: string) => void
}) {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()
  const { addToast } = useToast()
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingStatus, setPendingStatus] = useState<string | null>(null)
  const [open, setOpen] = useState(false)
  const [showTip, setShowTip] = useState(false)

  const { data: transitions = [], isFetched, isFetching } = useQuery({
    queryKey: ['subItemTransitions', teamId, subId],
    queryFn: () => getSubItemTransitionsApi(teamId!, subId),
    enabled: !!teamId && open,
  })

  useEffect(() => {
    if (open && isFetched && !isFetching && transitions.length === 0) {
      setOpen(false)
      setShowTip(true)
      setTimeout(() => setShowTip(false), 2000)
    }
  }, [open, isFetched, isFetching, transitions.length])

  const statusChangeMutation = useMutation({
    mutationFn: ({ newStatus }: { newStatus: string }) =>
      changeSubItemStatusApi(teamId!, subId, { status: newStatus }),
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ['mainItem', teamId] })
      qc.invalidateQueries({ queryKey: ['subItemTransitions', teamId, subId] })
      setOpen(false)
      setConfirmOpen(false)
      setPendingStatus(null)
      onStatusChange(pendingStatus || '')
    },
  })

  const handleSelect = useCallback((status: string) => {
    if (SUB_ITEM_TERMINAL_STATUSES.has(status)) {
      setPendingStatus(status)
      setConfirmOpen(true)
    } else {
      statusChangeMutation.mutate({ newStatus: status })
    }
  }, [statusChangeMutation])

  const handleConfirm = useCallback(() => {
    if (pendingStatus) {
      statusChangeMutation.mutate({ newStatus: pendingStatus })
    }
  }, [pendingStatus, statusChangeMutation])

  return (
    <>
      <div className="relative inline-flex">
        {showTip && (
          <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1.5 whitespace-nowrap text-xs px-2 py-1 rounded-md bg-primary text-white shadow-md pointer-events-none z-50">
            暂无可用流转
          </div>
        )}
        <DropdownMenu open={open} onOpenChange={setOpen}>
          <DropdownMenuTrigger asChild>
            <button className="focus:outline-none">
              <StatusBadge status={currentStatus} className="cursor-pointer" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="min-w-0 w-auto">
            {transitions.map((status) => (
              <DropdownMenuItem
                key={status}
                className="text-[13px] justify-center"
              onSelect={(e) => {
                e.preventDefault()
                handleSelect(status)
              }}
            >
              {getStatusName(status) || status}
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
      </div>

      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>确认变更状态</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <p className="text-sm text-secondary">
              确认将状态变更为「{getStatusName(pendingStatus || '') || pendingStatus}」？此操作可能不可逆。
            </p>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => { setConfirmOpen(false); setPendingStatus(null) }}>取消</Button>
            <Button onClick={handleConfirm} disabled={statusChangeMutation.isPending}>确认</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
