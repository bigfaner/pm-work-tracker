import { useState, useCallback, useRef, useEffect, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { listMainItemsApi, createMainItemApi } from '@/api/mainItems'
import { listMembersApi } from '@/api/teams'
import type { MainItem, SubItem } from '@/types'
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
import { Pagination, PaginationPageSize } from '@/components/ui/pagination'
import StatusBadge from '@/components/shared/StatusBadge'
import PriorityBadge from '@/components/shared/PriorityBadge'
import ProgressBar from '@/components/shared/ProgressBar'

// --- Constants ---

const STATUS_OPTIONS = ['未开始', '进行中', '待评审', '已完成', '已关闭', '阻塞中', '延期']
const SUMMARY_BATCH_SIZE = 5
const DEFAULT_PAGE_SIZE = 20

type ViewMode = 'summary' | 'detail'

// --- Main Component ---

export default function ItemViewPage() {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()

  // View mode
  const [viewMode, setViewMode] = useState<ViewMode>('summary')

  // Filter state (shared between views)
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [assigneeFilter, setAssigneeFilter] = useState<string>('')

  // Summary view: infinite scroll
  const [summaryCount, setSummaryCount] = useState(SUMMARY_BATCH_SIZE)
  const sentinelRef = useRef<HTMLDivElement>(null)

  // Detail view: pagination
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE)

  // Create dialog
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm, setCreateForm] = useState({ title: '', priority: 'P2', assigneeId: '', startDate: '', expectedEndDate: '' })

  // Expanded cards
  const [expandedCards, setExpandedCards] = useState<Set<number>>(new Set())

  // --- Data fetching ---

  const { data: membersData } = useQuery({
    queryKey: ['members', teamId],
    queryFn: () => listMembersApi(teamId!),
    enabled: !!teamId,
  })

  const { data: itemsData, isLoading } = useQuery({
    queryKey: ['mainItems', teamId],
    queryFn: () => listMainItemsApi(teamId!),
    enabled: !!teamId,
  })

  const members = membersData || []
  const allItems: (MainItem & { subItems?: SubItem[] })[] = itemsData?.items || []

  // --- Client-side filtering ---

  const filteredItems = useMemo(() => {
    let items = allItems
    if (searchText.trim()) {
      const q = searchText.trim().toLowerCase()
      items = items.filter(
        (item) =>
          item.title.toLowerCase().includes(q) ||
          item.code.toLowerCase().includes(q),
      )
    }
    if (statusFilter) {
      items = items.filter((item) => item.status === statusFilter)
    }
    if (assigneeFilter) {
      items = items.filter((item) => String(item.assignee_id) === assigneeFilter)
    }
    return items
  }, [allItems, searchText, statusFilter, assigneeFilter])

  // --- Summary view: visible items ---

  const summaryItems = filteredItems.slice(0, summaryCount)
  const hasMoreSummary = filteredItems.length > summaryCount

  // Infinite scroll observer
  useEffect(() => {
    if (viewMode !== 'summary') return
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMoreSummary) {
          setSummaryCount((prev) => prev + SUMMARY_BATCH_SIZE)
        }
      },
      { rootMargin: '200px' },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [viewMode, hasMoreSummary])

  // Reset summary count when filters change
  useEffect(() => {
    setSummaryCount(SUMMARY_BATCH_SIZE)
  }, [searchText, statusFilter, assigneeFilter])

  // --- Detail view: pagination ---

  const totalPages = Math.max(1, Math.ceil(filteredItems.length / pageSize))
  const paginatedItems = filteredItems.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize,
  )

  // Reset page when filters change
  useEffect(() => {
    setCurrentPage(1)
  }, [searchText, statusFilter, assigneeFilter, pageSize])

  // --- Mutations ---

  const createMutation = useMutation({
    mutationFn: (req: { title: string; priority: string; assigneeId?: number; startDate?: string; expectedEndDate?: string }) =>
      createMainItemApi(teamId!, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      setCreateOpen(false)
      setCreateForm({ title: '', priority: 'P2', assigneeId: '', startDate: '', expectedEndDate: '' })
    },
  })

  // --- Handlers ---

  const toggleExpand = useCallback((itemId: number) => {
    setExpandedCards((prev) => {
      const next = new Set(prev)
      if (next.has(itemId)) next.delete(itemId)
      else next.add(itemId)
      return next
    })
  }, [])

  const resetFilters = useCallback(() => {
    setSearchText('')
    setStatusFilter('')
    setAssigneeFilter('')
  }, [])

  const handleCreate = useCallback(() => {
    if (!createForm.title.trim()) return
    createMutation.mutate({
      title: createForm.title.trim(),
      priority: createForm.priority,
      ...(createForm.assigneeId && { assigneeId: Number(createForm.assigneeId) }),
      ...(createForm.startDate && { startDate: createForm.startDate }),
      ...(createForm.expectedEndDate && { expectedEndDate: createForm.expectedEndDate }),
    })
  }, [createForm, createMutation])

  const memberName = useCallback(
    (assigneeId: number | null) => {
      if (!assigneeId) return '-'
      const m = members.find((m) => m.userId === assigneeId)
      return m ? m.displayName : 'Unknown'
    },
    [members],
  )

  // --- Render helpers ---

  const formatDate = (date: string | null) => {
    if (!date) return '-'
    return date.replace(/-/g, '/')
  }

  // --- Render ---

  return (
    <div data-testid="item-view-page">
      {!teamId && <div className="p-6 text-tertiary">请先选择团队</div>}
      {teamId && (
        <>
      {/* Page Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold text-primary">事项清单</h1>
          <div className="inline-flex border border-border rounded-md overflow-hidden">
            <button
              className={`px-3 py-1 text-[13px] transition-colors ${
                viewMode === 'summary'
                  ? 'bg-primary-500 text-white'
                  : 'bg-transparent text-secondary hover:bg-bg-alt'
              }`}
              onClick={() => setViewMode('summary')}
              data-testid="toggle-summary"
            >
              汇总
            </button>
            <button
              className={`px-3 py-1 text-[13px] transition-colors ${
                viewMode === 'detail'
                  ? 'bg-primary-500 text-white'
                  : 'bg-transparent text-secondary hover:bg-bg-alt'
              }`}
              onClick={() => setViewMode('detail')}
              data-testid="toggle-detail"
            >
              明细
            </button>
          </div>
        </div>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          创建主事项
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
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="状态：全部" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="_all">状态：全部</SelectItem>
            {STATUS_OPTIONS.map((s) => (
              <SelectItem key={s} value={s}>{s}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={assigneeFilter} onValueChange={(v) => setAssigneeFilter(v === '_all' ? '' : v)}>
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="负责人：全部" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="_all">负责人：全部</SelectItem>
            {members.map((m) => (
              <SelectItem key={m.userId} value={String(m.userId)}>
                {m.displayName}
              </SelectItem>
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
          <p className="text-tertiary text-sm">暂无事项</p>
          <Button variant="secondary" size="sm" className="mt-3" onClick={() => setCreateOpen(true)}>
            创建第一个主事项
          </Button>
        </div>
      ) : viewMode === 'summary' ? (
        <SummaryView
          items={summaryItems}
          expandedCards={expandedCards}
          onToggleExpand={toggleExpand}
          memberName={memberName}
          formatDate={formatDate}
          hasMore={hasMoreSummary}
          sentinelRef={sentinelRef}
        />
      ) : (
        <DetailView
          items={paginatedItems}
          memberName={memberName}
          formatDate={formatDate}
          currentPage={currentPage}
          totalPages={totalPages}
          pageSize={pageSize}
          onPageChange={setCurrentPage}
          onPageSizeChange={(size) => {
            setPageSize(size)
            setCurrentPage(1)
          }}
          totalItems={filteredItems.length}
        />
      )}

      {/* Create Main Item Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent size="lg">
          <DialogHeader>
            <DialogTitle>新建主事项</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className="mb-4">
              <label className="block text-sm font-medium text-primary mb-1">
                标题 <span className="text-error">*</span>
              </label>
              <Input
                placeholder="请输入标题"
                maxLength={100}
                value={createForm.title}
                onChange={(e) => setCreateForm((f) => ({ ...f, title: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  优先级 <span className="text-error">*</span>
                </label>
                <Select value={createForm.priority} onValueChange={(v) => setCreateForm((f) => ({ ...f, priority: v }))}>
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
                <Select value={createForm.assigneeId || '_none'} onValueChange={(v) => setCreateForm((f) => ({ ...f, assigneeId: v === '_none' ? '' : v }))}>
                  <SelectTrigger><SelectValue placeholder="选择负责人" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="_none">不指定</SelectItem>
                    {members.map((m) => (
                      <SelectItem key={m.userId} value={String(m.userId)}>{m.displayName}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4 mt-4">
              <div>
                <label className="block text-sm font-medium text-primary mb-1">开始时间</label>
                <Input
                  type="date"
                  value={createForm.startDate}
                  onChange={(e) => setCreateForm((f) => ({ ...f, startDate: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">预期完成时间</label>
                <Input
                  type="date"
                  value={createForm.expectedEndDate}
                  onChange={(e) => setCreateForm((f) => ({ ...f, expectedEndDate: e.target.value }))}
                />
              </div>
            </div>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setCreateOpen(false)}>取消</Button>
            <Button onClick={handleCreate} disabled={!createForm.title.trim() || createMutation.isPending}>
              确认
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
        </>
      )}
    </div>
  )
}

// --- Summary View ---

interface SummaryViewProps {
  items: (MainItem & { subItems?: SubItem[] })[]
  expandedCards: Set<number>
  onToggleExpand: (id: number) => void
  memberName: (id: number | null) => string
  formatDate: (date: string | null) => string
  hasMore: boolean
  sentinelRef: React.RefObject<HTMLDivElement>
}

function SummaryView({
  items,
  expandedCards,
  onToggleExpand,
  memberName,
  formatDate,
  hasMore,
  sentinelRef,
}: SummaryViewProps) {
  return (
    <div>
      {items.map((item) => (
        <div key={item.id} className="mb-3">
          <div
            className="rounded-xl border border-border bg-white shadow-sm cursor-pointer"
            onClick={() => onToggleExpand(item.id)}
          >
            <div className="flex items-center gap-3 px-5 py-4">
              {/* Expand chevron */}
              <svg
                className={`w-4 h-4 flex-shrink-0 text-tertiary transition-transform ${
                  expandedCards.has(item.id) ? 'rotate-90' : ''
                }`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
              </svg>

              {/* Code */}
              <span className="font-mono text-xs text-tertiary bg-bg-alt px-1.5 py-0.5 rounded">
                {item.code}
              </span>

              {/* Priority */}
              <PriorityBadge priority={item.priority} />

              {/* Title + date range */}
              <div className="flex items-center gap-2 flex-1 min-w-0">
                <Link
                  to={`/items/${item.id}`}
                  className="text-sm font-medium text-primary hover:text-primary-600 truncate"
                  onClick={(e) => e.stopPropagation()}
                >
                  {item.title}
                </Link>
                <span className="text-xs text-tertiary whitespace-nowrap">
                  计划 {formatDate(item.start_date)}~{formatDate(item.expected_end_date)}
                  {item.actual_end_date ? ` 实际完成时间 ${formatDate(item.actual_end_date)}` : ''}
                </span>
              </div>

              {/* Assignee */}
              <span className="text-[13px] text-secondary whitespace-nowrap">
                {memberName(item.assignee_id)}
              </span>

              {/* Progress */}
              <div className="w-[112px] flex-shrink-0">
                <ProgressBar value={item.completion} size="sm" showPercentage />
              </div>

              {/* Status */}
              <StatusDropdown currentStatus={item.status} />

              {/* Actions */}
              <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
                <Link to={`/items/${item.id}`}>
                  <Button variant="ghost" size="sm">编辑</Button>
                </Link>
              </div>
            </div>

            {/* Expanded sub-items */}
            {expandedCards.has(item.id) && item.subItems && item.subItems.length > 0 && (
              <div className="border-t border-border px-5 py-3 pl-12">
                {item.subItems.map((sub) => (
                  <div
                    key={sub.id}
                    className="flex items-center gap-3 py-2 border-b border-border/50 last:border-b-0"
                  >
                    <span className="font-mono text-[11px] text-tertiary bg-bg-alt px-1.5 py-0.5 rounded">
                      SI-{String(item.id).padStart(3, '0')}-{String(sub.id).slice(-2)}
                    </span>
                    <PriorityBadge priority={sub.priority} className="text-[10px]" />
                    <Link
                      to={`/items/${item.id}/sub/${sub.id}`}
                      className="text-[13px] text-primary-600 hover:text-primary-700"
                    >
                      {sub.title}
                    </Link>
                    <span className="text-[11px] text-tertiary whitespace-nowrap">
                      计划 {formatDate(sub.start_date)}~{formatDate(sub.expected_end_date)}
                      {sub.actual_end_date ? ` 实际完成时间 ${formatDate(sub.actual_end_date)}` : ''}
                    </span>
                    <span className="ml-auto text-[13px] text-secondary">
                      {memberName(sub.assignee_id)}
                    </span>
                    <div className="w-[80px]">
                      <ProgressBar value={sub.completion} size="sm" />
                    </div>
                    <StatusBadge status={sub.status} />
                  </div>
                ))}
              </div>
            )}
          </div>
          {/* Expand button for test targeting */}
          <button
            data-testid={`expand-card-${item.id}`}
            className="hidden"
            onClick={() => onToggleExpand(item.id)}
          />
        </div>
      ))}

      {/* Infinite scroll sentinel */}
      <div ref={sentinelRef} className="py-4 text-center">
        {hasMore ? (
          <span className="text-xs text-tertiary">加载中...</span>
        ) : (
          items.length > 0 && (
            <span className="text-xs text-tertiary">-- 没有更多了 --</span>
          )
        )}
      </div>
    </div>
  )
}

// --- Detail View ---

interface DetailViewProps {
  items: (MainItem & { subItems?: SubItem[] })[]
  memberName: (id: number | null) => string
  formatDate: (date: string | null) => string
  currentPage: number
  totalPages: number
  pageSize: number
  onPageChange: (page: number) => void
  onPageSizeChange: (size: number) => void
  totalItems: number
}

function DetailView({
  items,
  memberName,
  formatDate,
  currentPage,
  totalPages,
  pageSize,
  onPageChange,
  onPageSizeChange,
  totalItems,
}: DetailViewProps) {
  return (
    <div className="rounded-xl border border-border bg-white shadow-sm">
      <div data-testid="detail-table">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>编号</TableHead>
              <TableHead>优先级</TableHead>
              <TableHead>标题</TableHead>
              <TableHead>负责人</TableHead>
              <TableHead>进度</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>开始时间</TableHead>
              <TableHead>预期完成时间</TableHead>
              <TableHead>实际完成时间</TableHead>
              <TableHead>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.id}>
                <TableCell>
                  <span className="font-mono text-xs">{item.code}</span>
                </TableCell>
                <TableCell>
                  <PriorityBadge priority={item.priority} />
                </TableCell>
                <TableCell>
                  <Link to={`/items/${item.id}`} className="font-medium text-primary hover:text-primary-600">
                    {item.title}
                  </Link>
                </TableCell>
                <TableCell>{memberName(item.assignee_id)}</TableCell>
                <TableCell>
                  <span className="text-xs">{item.completion}%</span>
                </TableCell>
                <TableCell>
                  <StatusDropdown currentStatus={item.status} />
                </TableCell>
                <TableCell className="text-xs">{formatDate(item.start_date)}</TableCell>
                <TableCell className="text-xs">{formatDate(item.expected_end_date)}</TableCell>
                <TableCell className="text-xs">{formatDate(item.actual_end_date)}</TableCell>
                <TableCell>
                  <div className="flex gap-1 whitespace-nowrap">
                    <Link to={`/items/${item.id}`}>
                      <Button variant="ghost" size="sm">编辑</Button>
                    </Link>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between px-5 py-3 border-t border-border">
        <PaginationPageSize
          pageSize={pageSize}
          onPageSizeChange={onPageSizeChange}
          options={[5, 10, 20, 50]}
        />
        <div className="flex items-center gap-3">
          <span className="text-[13px] text-tertiary">共 {totalItems} 条</span>
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={onPageChange}
          />
        </div>
      </div>
    </div>
  )
}

// --- Status Dropdown (inline status change) ---

function StatusDropdown({ currentStatus }: { currentStatus: string }) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="focus:outline-none">
          <StatusBadge status={currentStatus} className="cursor-pointer" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        {STATUS_OPTIONS.map((status) => (
          <DropdownMenuItem key={status} className="text-[13px]">
            {status}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
