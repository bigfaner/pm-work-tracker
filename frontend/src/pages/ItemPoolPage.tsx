import { useState, useCallback, useRef, useEffect, useMemo } from 'react'
import { ArrowUpCircle, ArrowDownCircle, XCircle } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { listItemPoolApi, submitItemPoolApi, assignItemPoolApi, convertToMainApi, rejectItemPoolApi } from '@/api/itemPool'
import { listMainItemsApi } from '@/api/mainItems'
import { listMembersApi } from '@/api/teams'
import type { ItemPool, AssignItemPoolReq, ConvertToMainItemReq } from '@/types'
import { PermissionGuard } from '@/components/PermissionGuard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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

// --- Constants ---

const POOL_BATCH_SIZE = 5
const POOL_STATUS_OPTIONS = [
  { value: '待分配', label: '待分配' },
  { value: '已分配', label: '已分配' },
  { value: '已拒绝', label: '已拒绝' },
]

const STATUS_BORDER: Record<string, string> = {
  '待分配': 'border-l-4 border-l-blue-500',
  '已分配': 'border-l-4 border-l-tertiary opacity-70',
  '已拒绝': 'border-l-4 border-l-error opacity-70',
}

const STATUS_BADGE_VARIANT: Record<string, 'primary' | 'success' | 'default'> = {
  '待分配': 'primary',
  '已分配': 'success',
  '已拒绝': 'default',
}

const STATUS_LABEL: Record<string, string> = {
  '待分配': '待分配',
  '已分配': '已分配',
  '已拒绝': '已拒绝',
}

// --- Pool Item Card ---

interface PoolItemCardProps {
  item: ItemPool
  onConvertToMain: (item: ItemPool) => void
  onConvertToSub: (item: ItemPool) => void
  onReject: (item: ItemPool) => void
}

function PoolItemCard({ item, onConvertToMain, onConvertToSub, onReject }: PoolItemCardProps) {
  const isPending = item.status === '待分配'

  return (
    <div
      data-testid={`pool-item-${item.id}`}
      className={`rounded-xl border border-border bg-white shadow-sm ${STATUS_BORDER[item.status] || ''}`}
    >
      {/* Header */}
      <div className="flex items-center justify-between px-5 py-3">
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs text-tertiary bg-bg-alt px-1.5 py-0.5 rounded">
            POOL-{String(item.id).padStart(3, '0')}
          </span>
          <span className="text-sm font-medium text-primary">{item.title}</span>
          <Badge variant={STATUS_BADGE_VARIANT[item.status]}>{STATUS_LABEL[item.status]}</Badge>
        </div>
        <span className="text-xs text-tertiary">{formatRelativeTime(item.createdAt)}</span>
      </div>

      {/* Body */}
      <div className="px-5 pb-3">
        {item.background && (
          <p className="text-[13px] text-secondary">
            <span className="text-tertiary">背景：</span>{item.background}
          </p>
        )}
        {item.expectedOutput && (
          <p className="text-[13px] text-secondary mt-1">
            <span className="text-tertiary">预期产出：</span>{item.expectedOutput}
          </p>
        )}
        {item.status === '已分配' && item.assignedMainId && (
          <div className="mt-2 text-[13px] text-secondary">
            {item.assignedSubId ? '已转为子事项挂载至：' : '已转为主事项：'}
            <Link to={`/items/${item.assignedMainId}`} className="font-medium text-primary-600 hover:text-primary-700 hover:underline">
              {item.assignedMainCode ? `${item.assignedMainCode} ${item.assignedMainTitle}` : `主事项 #${item.assignedMainId}`}
            </Link>
          </div>
        )}
        {item.status === '已拒绝' && item.rejectReason && (
          <div className="mt-2 text-[13px] text-tertiary">
            拒绝原因：{item.rejectReason}
          </div>
        )}
      </div>

      {/* Actions (only for pending items) */}
      {isPending && (
        <PermissionGuard code="item_pool:review">
          <div className="flex justify-end gap-2 px-5 py-2 border-t border-border/50">
            <Button variant="ghost" size="sm" className="text-primary-600" data-testid={`to-main-${item.id}`} onClick={() => onConvertToMain(item)}>
              <ArrowUpCircle className="w-3.5 h-3.5" />
              转为主事项
            </Button>
            <Button variant="ghost" size="sm" className="text-primary-600" data-testid={`to-sub-${item.id}`} onClick={() => onConvertToSub(item)}>
              <ArrowDownCircle className="w-3.5 h-3.5" />
              转为子事项
            </Button>
            <Button variant="ghost" size="sm" className="text-error" data-testid={`reject-${item.id}`} onClick={() => onReject(item)}>
              <XCircle className="w-3.5 h-3.5" />
              拒绝
            </Button>
          </div>
        </PermissionGuard>
      )}
    </div>
  )
}

// --- Main Component ---

export default function ItemPoolPage() {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()

  // Filter state
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')

  // Infinite scroll
  const [visibleCount, setVisibleCount] = useState(POOL_BATCH_SIZE)
  const sentinelRef = useRef<HTMLDivElement>(null)

  // Dialogs
  const [submitOpen, setSubmitOpen] = useState(false)
  const [toMainOpen, setToMainOpen] = useState(false)
  const [toSubOpen, setToSubOpen] = useState(false)
  const [rejectOpen, setRejectOpen] = useState(false)
  const [selectedItem, setSelectedItem] = useState<ItemPool | null>(null)

  // Form states
  const [submitForm, setSubmitForm] = useState({ title: '', background: '', expectedOutput: '' })
  const [toMainForm, setToMainForm] = useState({ priority: 'P2', assigneeId: '', startDate: '', expectedEndDate: '' })
  const [toSubForm, setToSubForm] = useState({ parentItemId: '', priority: 'P2', assigneeId: '', startDate: '', expectedEndDate: '' })
  const [rejectForm, setRejectForm] = useState({ reason: '' })

  // --- Data fetching ---

  const { data: poolData, isLoading } = useQuery({
    queryKey: ['itemPool', teamId],
    queryFn: () => listItemPoolApi(teamId!),
    enabled: !!teamId,
  })

  const { data: membersData } = useQuery({
    queryKey: ['members', teamId],
    queryFn: () => listMembersApi(teamId!),
    enabled: !!teamId,
  })

  const { data: mainItemsData } = useQuery({
    queryKey: ['mainItems', teamId],
    queryFn: () => listMainItemsApi(teamId!),
    enabled: !!teamId,
  })

  const members = membersData || []
  const mainItems = mainItemsData?.items || []
  const allItems: ItemPool[] = poolData?.items || []

  // --- Client-side filtering ---

  const filteredItems = useMemo(() => {
    let items = allItems
    if (searchText.trim()) {
      const q = searchText.trim().toLowerCase()
      items = items.filter(
        (item) =>
          item.title.toLowerCase().includes(q) ||
          `pool-${String(item.id).padStart(3, '0')}`.includes(q),
      )
    }
    if (statusFilter) {
      items = items.filter((item) => item.status === statusFilter)
    }
    return items
  }, [allItems, searchText, statusFilter])

  // --- Infinite scroll ---

  const visibleItems = filteredItems.slice(0, visibleCount)
  const hasMore = filteredItems.length > visibleCount

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore) {
          setVisibleCount((prev) => prev + POOL_BATCH_SIZE)
        }
      },
      { rootMargin: '200px' },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [hasMore])

  // Reset visible count when filters change
  useEffect(() => {
    setVisibleCount(POOL_BATCH_SIZE)
  }, [searchText, statusFilter])

  // --- Mutations ---

  const submitMutation = useMutation({
    mutationFn: (req: { title: string; background?: string; expectedOutput?: string }) =>
      submitItemPoolApi(teamId!, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['itemPool', teamId] })
      setSubmitOpen(false)
      setSubmitForm({ title: '', background: '', expectedOutput: '' })
    },
  })

  const assignMutation = useMutation({
    mutationFn: ({ poolId, req }: { poolId: number; req: AssignItemPoolReq }) =>
      assignItemPoolApi(teamId!, poolId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['itemPool', teamId] })
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      setToSubOpen(false)
      setSelectedItem(null)
    },
  })

  const convertToMainMutation = useMutation({
    mutationFn: ({ poolId, req }: { poolId: number; req: ConvertToMainItemReq }) =>
      convertToMainApi(teamId!, poolId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['itemPool', teamId] })
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      setToMainOpen(false)
      setSelectedItem(null)
    },
  })

  const rejectMutation = useMutation({
    mutationFn: ({ poolId, req }: { poolId: number; req: { reason: string } }) =>
      rejectItemPoolApi(teamId!, poolId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['itemPool', teamId] })
      setRejectOpen(false)
      setRejectForm({ reason: '' })
      setSelectedItem(null)
    },
  })

  // --- Handlers ---

  const openConvertToMain = useCallback((item: ItemPool) => {
    setSelectedItem(item)
    setToMainForm({ priority: 'P2', assigneeId: '', startDate: '', expectedEndDate: '' })
    setToMainOpen(true)
  }, [])

  const openConvertToSub = useCallback((item: ItemPool) => {
    setSelectedItem(item)
    setToSubForm({ parentItemId: '', priority: 'P2', assigneeId: '', startDate: '', expectedEndDate: '' })
    setToSubOpen(true)
  }, [])

  const openReject = useCallback((item: ItemPool) => {
    setSelectedItem(item)
    setRejectForm({ reason: '' })
    setRejectOpen(true)
  }, [])

  const resetFilters = useCallback(() => {
    setSearchText('')
    setStatusFilter('')
  }, [])

  const handleSubmit = useCallback(() => {
    if (!submitForm.title.trim()) return
    submitMutation.mutate({
      title: submitForm.title.trim(),
      ...(submitForm.background && { background: submitForm.background }),
      ...(submitForm.expectedOutput && { expectedOutput: submitForm.expectedOutput }),
    })
  }, [submitForm, submitMutation])

  const handleToMain = useCallback(() => {
    if (!selectedItem) return
    convertToMainMutation.mutate({
      poolId: selectedItem.id,
      req: {
        priority: toMainForm.priority || 'P2',
        assigneeId: toMainForm.assigneeId ? Number(toMainForm.assigneeId) : 0,
        startDate: toMainForm.startDate || '',
        expectedEndDate: toMainForm.expectedEndDate || '',
      },
    })
  }, [selectedItem, toMainForm, convertToMainMutation])

  const handleToSub = useCallback(() => {
    if (!selectedItem || !toSubForm.parentItemId) return
    assignMutation.mutate({
      poolId: selectedItem.id,
      req: {
        mainItemId: Number(toSubForm.parentItemId),
        assigneeId: toSubForm.assigneeId ? Number(toSubForm.assigneeId) : 0,
        priority: toSubForm.priority || 'P2',
        startDate: toSubForm.startDate || '',
        expectedEndDate: toSubForm.expectedEndDate || '',
      },
    })
  }, [selectedItem, toSubForm, assignMutation])

  const handleReject = useCallback(() => {
    if (!selectedItem || !rejectForm.reason.trim()) return
    rejectMutation.mutate({
      poolId: selectedItem.id,
      req: { reason: rejectForm.reason.trim() },
    })
  }, [selectedItem, rejectForm, rejectMutation])

  // --- Render ---

  return (
    <div data-testid="item-pool-page">
      {!teamId && <div className="p-6 text-tertiary">请先选择团队</div>}
      {teamId && (
        <>
          {/* Page Header */}
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-xl font-semibold text-primary">待办事项</h1>
            <Button size="sm" onClick={() => setSubmitOpen(true)}>
              <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
              </svg>
              新增待办事项
            </Button>
          </div>

          {/* Filter Bar */}
          <div className="flex items-center gap-3 mb-4 flex-wrap">
            <Input
              placeholder="搜索标题或编号..."
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              className="w-[360px]"
            />
            <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v === '_all' ? '' : v)}>
              <SelectTrigger className="w-[140px]" data-testid="pool-status-filter">
                <SelectValue placeholder="状态：全部" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="_all">状态：全部</SelectItem>
                {POOL_STATUS_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button variant="secondary" size="sm" onClick={resetFilters}>
              重置
            </Button>
          </div>

          {/* Content */}
          {isLoading ? (
            <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
          ) : filteredItems.length === 0 ? (
            <div className="py-12 text-center">
              <p className="text-tertiary text-sm">暂无待办事项</p>
              <Button variant="secondary" size="sm" className="mt-3" onClick={() => setSubmitOpen(true)}>
                提交第一个待办事项
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              {visibleItems.map((item) => (
                <PoolItemCard
                  key={item.id}
                  item={item}
                  onConvertToMain={openConvertToMain}
                  onConvertToSub={openConvertToSub}
                  onReject={openReject}
                />
              ))}

              {/* Infinite scroll sentinel */}
              <div ref={sentinelRef} className="py-4 text-center">
                {hasMore ? (
                  <span className="text-xs text-tertiary">加载中...</span>
                ) : (
                  <span className="text-xs text-tertiary">-- 没有更多了 --</span>
                )}
              </div>
            </div>
          )}

          {/* Submit Dialog */}
          <Dialog open={submitOpen} onOpenChange={setSubmitOpen}>
            <DialogContent size="md">
              <DialogHeader>
                <DialogTitle>新增待办事项</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    标题 <span className="text-error">*</span>
                  </label>
                  <Input
                    placeholder="请输入事项标题"
                    value={submitForm.title}
                    onChange={(e) => setSubmitForm((f) => ({ ...f, title: e.target.value }))}
                  />
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">背景</label>
                  <Textarea
                    rows={3}
                    placeholder="描述提交该事项的背景和原因"
                    value={submitForm.background}
                    onChange={(e) => setSubmitForm((f) => ({ ...f, background: e.target.value }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">预期产出</label>
                  <Textarea
                    rows={3}
                    placeholder="描述希望达成的产出或目标"
                    value={submitForm.expectedOutput}
                    onChange={(e) => setSubmitForm((f) => ({ ...f, expectedOutput: e.target.value }))}
                  />
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setSubmitOpen(false)}>取消</Button>
                <Button onClick={handleSubmit} disabled={!submitForm.title.trim() || submitMutation.isPending}>
                  提交
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Convert to Main Item Dialog */}
          <Dialog open={toMainOpen} onOpenChange={setToMainOpen}>
            <DialogContent size="lg">
              <DialogHeader>
                <DialogTitle>转为主事项</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    标题 <span className="text-error">*</span>
                  </label>
                  <Input value={selectedItem?.title || ''} readOnly />
                </div>
                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">优先级</label>
                    <Select value={toMainForm.priority} onValueChange={(v) => setToMainForm((f) => ({ ...f, priority: v }))}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <PrioritySelectItems />
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">负责人</label>
                    <Select value={toMainForm.assigneeId || '_none'} onValueChange={(v) => setToMainForm((f) => ({ ...f, assigneeId: v === '_none' ? '' : v }))}>
                      <SelectTrigger><SelectValue placeholder="请选择" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="_none">请选择</SelectItem>
                        {members.map((m) => (
                          <SelectItem key={m.userId} value={String(m.userId)}>{m.displayName}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">开始时间 <span className="text-error">*</span></label>
                    <Input
                      type="date"
                      value={toMainForm.startDate}
                      onChange={(e) => setToMainForm((f) => ({ ...f, startDate: e.target.value }))}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">预期完成时间 <span className="text-error">*</span></label>
                    <Input
                      type="date"
                      value={toMainForm.expectedEndDate}
                      onChange={(e) => setToMainForm((f) => ({ ...f, expectedEndDate: e.target.value }))}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">描述</label>
                  <Textarea
                    rows={3}
                    value={selectedItem?.background || ''}
                    readOnly
                  />
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setToMainOpen(false)}>取消</Button>
                <Button onClick={handleToMain} disabled={!toMainForm.startDate || !toMainForm.expectedEndDate || convertToMainMutation.isPending}>确认转换</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Convert to Sub Item Dialog */}
          <Dialog open={toSubOpen} onOpenChange={setToSubOpen}>
            <DialogContent size="lg">
              <DialogHeader>
                <DialogTitle>转为子事项</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    挂载主事项 <span className="text-error">*</span>
                  </label>
                  <Select value={toSubForm.parentItemId || '_none'} onValueChange={(v) => setToSubForm((f) => ({ ...f, parentItemId: v === '_none' ? '' : v }))}>
                    <SelectTrigger><SelectValue placeholder="请选择主事项" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="_none">请选择主事项</SelectItem>
                      {mainItems.map((mi) => (
                        <SelectItem key={mi.id} value={String(mi.id)}>
                          {mi.code} {mi.title}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-primary mb-1">
                    标题 <span className="text-error">*</span>
                  </label>
                  <Input value={selectedItem?.title || ''} readOnly />
                </div>
                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">优先级</label>
                    <Select value={toSubForm.priority} onValueChange={(v) => setToSubForm((f) => ({ ...f, priority: v }))}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <PrioritySelectItems />
                      </SelectContent>
                    </Select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">负责人</label>
                    <Select value={toSubForm.assigneeId || '_none'} onValueChange={(v) => setToSubForm((f) => ({ ...f, assigneeId: v === '_none' ? '' : v }))}>
                      <SelectTrigger><SelectValue placeholder="请选择" /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="_none">请选择</SelectItem>
                        {members.map((m) => (
                          <SelectItem key={m.userId} value={String(m.userId)}>{m.displayName}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">开始时间 <span className="text-error">*</span></label>
                    <Input
                      type="date"
                      value={toSubForm.startDate}
                      onChange={(e) => setToSubForm((f) => ({ ...f, startDate: e.target.value }))}
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-primary mb-1">预期完成时间 <span className="text-error">*</span></label>
                    <Input
                      type="date"
                      value={toSubForm.expectedEndDate}
                      onChange={(e) => setToSubForm((f) => ({ ...f, expectedEndDate: e.target.value }))}
                    />
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">描述</label>
                  <Textarea
                    rows={3}
                    value={selectedItem?.background || ''}
                    readOnly
                  />
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setToSubOpen(false)}>取消</Button>
                <Button onClick={handleToSub} disabled={!toSubForm.parentItemId || !toSubForm.startDate || !toSubForm.expectedEndDate || assignMutation.isPending}>确认转换</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* Reject Dialog */}
          <Dialog open={rejectOpen} onOpenChange={setRejectOpen}>
            <DialogContent size="md">
              <DialogHeader>
                <DialogTitle>拒绝事项</DialogTitle>
              </DialogHeader>
              <DialogBody>
                <div>
                  <label className="block text-sm font-medium text-primary mb-1">
                    拒绝原因 <span className="text-error">*</span>
                  </label>
                  <Textarea
                    className="min-h-[96px]"
                    rows={4}
                    placeholder="请说明拒绝原因"
                    value={rejectForm.reason}
                    onChange={(e) => setRejectForm({ reason: e.target.value })}
                  />
                </div>
              </DialogBody>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setRejectOpen(false)}>取消</Button>
                <Button variant="danger" onClick={handleReject} disabled={!rejectForm.reason.trim() || rejectMutation.isPending}>确认拒绝</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </>
      )}
    </div>
  )
}

// --- Utility ---

function formatRelativeTime(dateStr: string): string {
  const now = new Date()
  const date = new Date(dateStr)
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  if (diffDays === 0) return '今天提交'
  if (diffDays === 1) return '1天前提交'
  if (diffDays < 7) return `${diffDays}天前提交`
  if (diffDays < 30) {
    const weeks = Math.floor(diffDays / 7)
    return `${weeks}周前提交`
  }
  return date.toLocaleDateString('zh-CN')
}
