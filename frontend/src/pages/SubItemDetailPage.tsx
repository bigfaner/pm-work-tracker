import { useState, useCallback, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { getMainItemApi } from '@/api/mainItems'
import { getSubItemApi, updateSubItemApi, changeSubItemStatusApi, getSubItemTransitionsApi } from '@/api/subItems'
import { listProgressApi, appendProgressApi } from '@/api/progress'
import { listMembersApi } from '@/api/teams'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { DateInput } from '@/components/ui/date-input'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { PermissionGuard } from '@/components/PermissionGuard'
import { Progress } from '@/components/ui/progress'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import StatusBadge from '@/components/shared/StatusBadge'
import PriorityBadge from '@/components/shared/PriorityBadge'
import { PrioritySelectItems } from '@/components/shared/PrioritySelect'
import { SUB_ITEM_STATUSES, getStatusName, isOverdue } from '@/lib/status'

// --- Main Component ---

export default function SubItemDetailPage() {
  const { mainItemId, subItemId } = useParams<{ mainItemId: string; subItemId: string }>()
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()
  const mId = Number(mainItemId)
  const sId = Number(subItemId)

  // State
  const [appendOpen, setAppendOpen] = useState(false)
  const [appendForm, setAppendForm] = useState({ completion: '', achievement: '', blocker: '' })
  const [editOpen, setEditOpen] = useState(false)
  const [editForm, setEditForm] = useState({ title: '', priority: '', assigneeId: '', expectedEndDate: '', description: '' })

  // --- Data fetching ---

  const { data: subItem, isLoading: subLoading } = useQuery({
    queryKey: ['subItem', teamId, sId],
    queryFn: () => getSubItemApi(teamId!, sId),
    enabled: !!teamId && !!sId,
  })

  const { data: mainItem } = useQuery({
    queryKey: ['mainItem', teamId, mId],
    queryFn: () => getMainItemApi(teamId!, mId),
    enabled: !!teamId && !!mId,
  })

  const { data: progressRecords } = useQuery({
    queryKey: ['progress', teamId, sId],
    queryFn: () => listProgressApi(teamId!, sId),
    enabled: !!teamId && !!sId,
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

  // Last completion value for validation
  const lastCompletion = progressRecords && progressRecords.length > 0
    ? progressRecords[progressRecords.length - 1].completion
    : 0

  // --- Mutations ---

  const editMutation = useMutation({
    mutationFn: (req: { title: string; priority: string; assigneeId?: number; expectedEndDate?: string; description?: string }) =>
      updateSubItemApi(teamId!, sId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['subItem', teamId, sId] })
      setEditOpen(false)
    },
  })

  const appendMutation = useMutation({
    mutationFn: (req: { completion: number; achievement?: string; blocker?: string }) =>
      appendProgressApi(teamId!, sId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['progress', teamId, sId] })
      qc.invalidateQueries({ queryKey: ['subItem', teamId, sId] })
      setAppendOpen(false)
      setAppendForm({ completion: '', achievement: '', blocker: '' })
    },
  })

  // --- Handlers ---

  const handleEdit = useCallback(() => {
    if (!editForm.title.trim()) return
    editMutation.mutate({
      title: editForm.title.trim(),
      priority: editForm.priority,
      assigneeId: editForm.assigneeId ? Number(editForm.assigneeId) : undefined,
      expectedEndDate: editForm.expectedEndDate || undefined,
      description: editForm.description,
    })
  }, [editForm, editMutation])

  const handleAppend = useCallback(() => {
    const val = Number(appendForm.completion)
    if (isNaN(val) || val < 0 || val > 100) return
    if (val < lastCompletion) return
    appendMutation.mutate({
      completion: val,
      ...(appendForm.achievement && { achievement: appendForm.achievement }),
      ...(appendForm.blocker && { blocker: appendForm.blocker }),
    })
  }, [appendForm, lastCompletion, appendMutation])

  // Sorted progress records: reverse chronological
  const sortedRecords = [...(progressRecords || [])].reverse()

  // Sub item code
  const subCode = subItem ? `SI-${String(mId).padStart(3, '0')}-${String(sId).slice(-2)}` : ''

  // Terminal status guard
  const isTerminalStatus = subItem
    ? !!(SUB_ITEM_STATUSES[subItem.status as keyof typeof SUB_ITEM_STATUSES]?.terminal)
    : false

  // --- Render ---

  if (!teamId) return <div className="p-6 text-tertiary">请先选择团队</div>

  const isLoading = subLoading

  return (
    <div data-testid="sub-item-detail-page">
      {isLoading ? (
        <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
      ) : !subItem ? (
        <div className="py-8 text-center text-tertiary text-sm">子事项不存在</div>
      ) : (
        <>
          {/* Breadcrumb */}
          <Breadcrumb className="mb-4">
            <BreadcrumbItem href="/items">事项清单</BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem href={`/items/${mId}`}>
              {mainItem?.title || `主事项 #${mId}`}
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem isCurrent>{subItem.title}</BreadcrumbItem>
          </Breadcrumb>

          {/* Title Bar */}
          <div className="flex items-center gap-3 mb-6 flex-wrap">
            <Badge variant="default" className="font-mono">{subCode}</Badge>
            <h1 className="text-xl font-semibold text-primary m-0">{subItem.title}</h1>
            <PriorityBadge priority={subItem.priority} />
            <SubItemStatusDropdown subId={sId} currentStatus={subItem.status} />
            <div className="flex-1" />
            <PermissionGuard code="main_item:update">
              <Button variant="secondary" disabled={isTerminalStatus} onClick={() => {
                setEditForm({
                  title: subItem.title,
                  priority: subItem.priority,
                  assigneeId: subItem.assigneeId ? String(subItem.assigneeId) : '',
                  expectedEndDate: subItem.expectedEndDate || '',
                  description: subItem.description || '',
                })
                setEditOpen(true)
              }}>
                <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                编辑
              </Button>
            </PermissionGuard>
          </div>

          {/* Info Card */}
          <Card className="mb-5">
            <CardContent>
              <div className="grid grid-cols-4 gap-4 mb-4">
                <div>
                  <div className="text-xs text-tertiary mb-1">负责人</div>
                  <span className="text-[13px] font-medium">{memberName(subItem.assigneeId)}</span>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">开始时间</div>
                  <span className="text-[13px] font-medium">{subItem.startDate || '-'}</span>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">预期完成时间</div>
                  <div className="flex items-center gap-1.5">
                    <span className="text-[13px] font-medium">{subItem.expectedEndDate || '-'}</span>
                    {isOverdue(subItem.expectedEndDate ?? undefined, subItem.status) && (
                      <Badge variant="error">延期</Badge>
                    )}
                  </div>
                </div>
                <div>
                  <div className="text-xs text-tertiary mb-1">结束时间</div>
                  <span className="text-[13px] font-medium">{subItem.actualEndDate || '-'}</span>
                </div>
              </div>
              <div>
                <div className="text-xs text-tertiary mb-1">描述</div>
                <span className="text-[13px] text-secondary leading-relaxed">{subItem.description || '暂无描述'}</span>
              </div>
            </CardContent>
          </Card>

          {/* Progress Bar */}
          <Card className="mb-5">
            <CardContent>
              <div>
                <span className="text-[13px] text-secondary">总进度</span>
                <Progress value={subItem.completion} className="mt-2" />
                <span className="text-[13px] font-semibold mt-1 block text-center">{subItem.completion}%</span>
              </div>
            </CardContent>
          </Card>

          {/* Progress Timeline */}
          <Card>
            <CardHeader>
              <h3 className="text-sm font-semibold text-primary m-0">进度记录</h3>
              <PermissionGuard code="progress:update">
                <Button size="sm" disabled={isTerminalStatus} onClick={() => setAppendOpen(true)}>追加进度</Button>
              </PermissionGuard>
            </CardHeader>
            <CardContent>
              {sortedRecords.length === 0 ? (
                <div className="text-tertiary text-sm">暂无进度记录</div>
              ) : (
                <div className="relative pl-6 border-l-2 border-border space-y-5">
                  {sortedRecords.map((record) => {
                    const date = new Date(record.createdAt)
                    const dateStr = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
                    return (
                      <div key={record.id} className="relative">
                        {/* Timeline dot */}
                        <div className="absolute -left-6.25 top-1 w-2.5 h-2.5 rounded-full bg-primary-500 border-2 border-white" />
                        <div className="mb-1">
                          <span className="text-xs text-tertiary">{dateStr}</span>
                          <span className="text-xs text-tertiary ml-2">{record.completion}%</span>
                        </div>
                        <div className="text-[13px] text-secondary">
                          {record.achievement && (
                            <div className="mt-1">
                              <strong className="text-success-text">成果：</strong>{record.achievement}
                            </div>
                          )}
                          {record.blocker && (
                            <div className="mt-1">
                              <strong className="text-error-text">卡点：</strong>{record.blocker}
                            </div>
                          )}
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Edit Dialog */}
          <Dialog open={editOpen} onOpenChange={setEditOpen}>
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
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">预期完成时间</label>
                  <DateInput
                    value={editForm.expectedEndDate}
                    onChange={(e) => setEditForm((f) => ({ ...f, expectedEndDate: e.target.value }))}
                  />
                </div>
                <div>
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
                <Button onClick={handleEdit} disabled={!editForm.title.trim() || editMutation.isPending}>保存</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Append Progress Dialog */}
          <Dialog open={appendOpen} onOpenChange={setAppendOpen}>
            <DialogContent size="md">
              <DialogHeader>
                <DialogTitle>追加进度</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    进度 <span className="text-error">*</span>
                  </label>
                  <Input
                    type="number"
                    min={0}
                    max={100}
                    placeholder="请输入 0-100 的整数"
                    value={appendForm.completion}
                    onChange={(e) => setAppendForm((f) => ({ ...f, completion: e.target.value }))}
                  />
                  <div className="text-xs text-tertiary mt-1">
                    不能低于上一条记录的进度（当前：{lastCompletion}%）
                  </div>
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">成果</label>
                  <Textarea
                    rows={3}
                    placeholder="描述本次取得的成果"
                    value={appendForm.achievement}
                    onChange={(e) => setAppendForm((f) => ({ ...f, achievement: e.target.value }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">卡点</label>
                  <Textarea
                    rows={3}
                    placeholder="描述遇到的问题或阻碍"
                    value={appendForm.blocker}
                    onChange={(e) => setAppendForm((f) => ({ ...f, blocker: e.target.value }))}
                  />
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setAppendOpen(false)}>取消</Button>
                <Button
                  onClick={handleAppend}
                  disabled={!appendForm.completion || appendMutation.isPending}
                >
                  提交
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </>
      )}
    </div>
  )
}

// --- Sub-item StatusDropdown ---

const SUB_ITEM_TERMINAL_STATUSES = new Set(
  Object.entries(SUB_ITEM_STATUSES)
    .filter(([, v]) => v.terminal)
    .map(([k]) => k)
)

function SubItemStatusDropdown({
  subId,
  currentStatus,
}: {
  subId: number
  currentStatus: string
}) {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [achievementOpen, setAchievementOpen] = useState(false)
  const [achievementText, setAchievementText] = useState('')
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
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['subItem', teamId, subId] })
      qc.invalidateQueries({ queryKey: ['subItemTransitions', teamId, subId] })
      setOpen(false)
      setConfirmOpen(false)
      setAchievementOpen(false)
      setAchievementText('')
      setPendingStatus(null)
    },
  })

  const appendMutation = useMutation({
    mutationFn: (achievement: string) =>
      appendProgressApi(teamId!, subId, { completion: 100, ...(achievement && { achievement }) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['progress', teamId, subId] })
      qc.invalidateQueries({ queryKey: ['subItem', teamId, subId] })
    },
  })

  const handleSelect = useCallback((status: string) => {
    if (status === 'completed') {
      setPendingStatus(status)
      setAchievementOpen(true)
    } else if (SUB_ITEM_TERMINAL_STATUSES.has(status)) {
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

  const handleAchievementConfirm = useCallback(async () => {
    if (!pendingStatus) return
    const achievement = achievementText.trim()
    try {
      await statusChangeMutation.mutateAsync({ newStatus: pendingStatus })
      if (achievement) {
        await appendMutation.mutateAsync(achievement)
      }
    } catch {
      // errors handled by individual mutations
    }
  }, [pendingStatus, achievementText, statusChangeMutation, appendMutation])

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

      {/* Achievement dialog for completed */}
      <Dialog open={achievementOpen} onOpenChange={(open) => {
        if (!open) { setAchievementOpen(false); setAchievementText(''); setPendingStatus(null) }
      }}>
        <DialogContent size="md">
          <DialogHeader>
            <DialogTitle>填写完成成果</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <p className="text-sm text-secondary mb-4">
              即将标记为「已完成」，进度将自动设为 100%。
            </p>
            <div>
              <label className="block text-sm font-medium text-primary mb-1">成果</label>
              <Textarea
                rows={4}
                placeholder="描述本次取得的成果（选填）"
                value={achievementText}
                onChange={(e) => setAchievementText(e.target.value)}
              />
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => { setAchievementOpen(false); setAchievementText(''); setPendingStatus(null) }}>取消</Button>
            <Button
              onClick={handleAchievementConfirm}
              disabled={statusChangeMutation.isPending || appendMutation.isPending}
            >
              确认完成
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
