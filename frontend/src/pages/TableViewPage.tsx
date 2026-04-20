import { useState, useMemo, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { getTableViewApi, exportTableCsvApi } from '@/api/views'
import { listMembersApi } from '@/api/teams'
import type { TableRow, TableFilter } from '@/types'
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
  TableRow as TableRowComp,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { Pagination, PaginationPageSize } from '@/components/ui/pagination'
import StatusBadge from '@/components/shared/StatusBadge'
import PriorityBadge from '@/components/shared/PriorityBadge'
import ProgressBar from '@/components/shared/ProgressBar'
import UserAvatar from '@/components/shared/UserAvatar'

// --- Constants ---

const STATUS_OPTIONS = ['未开始', '进行中', '待评审', '已完成', '已关闭', '阻塞中', '延期']
const PRIORITY_OPTIONS = ['P1', 'P2', 'P3']
const TYPE_OPTIONS = [
  { value: '', label: '类型：全部' },
  { value: 'main', label: '主事项' },
  { value: 'sub', label: '子事项' },
]
const PAGE_SIZE_OPTIONS = [5, 10, 20, 50]
const DEFAULT_PAGE_SIZE = 10

// --- Main Component ---

export default function TableViewPage() {
  const teamId = useTeamStore((s) => s.currentTeamId)

  // Filter state
  const [searchText, setSearchText] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('')
  const [priorityFilter, setPriorityFilter] = useState<string>('')
  const [assigneeFilter, setAssigneeFilter] = useState<string>('')
  const [statusFilter, setStatusFilter] = useState<string>('')

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE)

  // --- Data fetching ---

  const { data: membersData } = useQuery({
    queryKey: ['members', teamId],
    queryFn: () => listMembersApi(teamId!),
    enabled: !!teamId,
  })

  const members = membersData || []

  // Build server-side filter
  const serverFilter: TableFilter = useMemo(() => {
    const filter: TableFilter = {
      page: currentPage,
      pageSize,
    }
    if (typeFilter) filter.type = typeFilter
    if (priorityFilter) filter.priority = priorityFilter
    if (statusFilter) filter.status = statusFilter
    if (assigneeFilter) filter.assigneeId = Number(assigneeFilter)
    return filter
  }, [typeFilter, priorityFilter, statusFilter, assigneeFilter, currentPage, pageSize])

  const { data: tableData, isLoading } = useQuery({
    queryKey: ['tableView', teamId, serverFilter],
    queryFn: () => getTableViewApi(teamId!, serverFilter),
    enabled: !!teamId,
  })

  // Handle backend returning `size` instead of `pageSize`
  const apiItems: TableRow[] = tableData?.items || []
  const apiTotal = tableData?.total || 0

  // Client-side title search
  const filteredItems = useMemo(() => {
    if (!searchText.trim()) return apiItems
    const q = searchText.trim().toLowerCase()
    return apiItems.filter(
      (row) =>
        row.title.toLowerCase().includes(q) ||
        row.code.toLowerCase().includes(q),
    )
  }, [apiItems, searchText])

  const totalItems = searchText.trim() ? filteredItems.length : apiTotal

  // --- Reset page when filters change ---
  const handleTypeChange = useCallback((v: string) => {
    setTypeFilter(v === '_all' ? '' : v)
    setCurrentPage(1)
  }, [])
  const handlePriorityChange = useCallback((v: string) => {
    setPriorityFilter(v === '_all' ? '' : v)
    setCurrentPage(1)
  }, [])
  const handleAssigneeChange = useCallback((v: string) => {
    setAssigneeFilter(v === '_all' ? '' : v)
    setCurrentPage(1)
  }, [])
  const handleStatusChange = useCallback((v: string) => {
    setStatusFilter(v === '_all' ? '' : v)
    setCurrentPage(1)
  }, [])
  const handlePageSizeChange = useCallback((size: number) => {
    setPageSize(size)
    setCurrentPage(1)
  }, [])
  const resetFilters = useCallback(() => {
    setSearchText('')
    setTypeFilter('')
    setPriorityFilter('')
    setAssigneeFilter('')
    setStatusFilter('')
    setCurrentPage(1)
  }, [])

  // --- Helpers ---

  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize))

  const isOverdue = (expectedEndDate: string | null, status: string): boolean => {
    if (!expectedEndDate) return false
    if (status === '已完成' || status === '已关闭') return false
    return new Date(expectedEndDate) < new Date()
  }

  const formatDate = (date: string | null) => {
    if (!date) return '-'
    return date
  }

  const getItemLink = (row: TableRow): string => {
    if (row.type === 'main') return `/items/${row.id}`
    if (row.mainItemId) return `/items/${row.mainItemId}/sub/${row.id}`
    // Fallback: can't determine parent
    return `/items/${row.id}`
  }

  // --- CSV export ---

  const handleExportCsv = async () => {
    if (!teamId) return
    try {
      const exportFilter: TableFilter = {}
      if (typeFilter) exportFilter.type = typeFilter
      if (priorityFilter) exportFilter.priority = priorityFilter
      if (statusFilter) exportFilter.status = statusFilter
      if (assigneeFilter) exportFilter.assigneeId = Number(assigneeFilter)

      const blob = await exportTableCsvApi(teamId, exportFilter)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'items-export.csv'
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      // Silently fail - could add toast later
    }
  }

  // --- Render ---

  return (
    <div data-testid="table-view-page">
      {!teamId && <div className="p-6 text-tertiary">请先选择团队</div>}
      {teamId && (
        <>
          {/* Page Header */}
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-xl font-semibold text-primary">表格视图</h1>
            <Button size="sm" onClick={handleExportCsv} data-testid="export-csv-btn">
              导出 CSV
            </Button>
          </div>

          {/* Filter Bar */}
          <div className="flex items-center gap-3 mb-4 flex-wrap">
            <Input
              placeholder="搜索标题..."
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              className="w-[180px]"
            />
            <Select value={typeFilter || '_all'} onValueChange={handleTypeChange}>
              <SelectTrigger className="w-[120px]" data-testid="type-filter">
                <SelectValue placeholder="类型：全部" />
              </SelectTrigger>
              <SelectContent>
                {TYPE_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value || '_all'}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={priorityFilter || '_all'} onValueChange={handlePriorityChange}>
              <SelectTrigger className="w-[120px]" data-testid="priority-filter">
                <SelectValue placeholder="优先级：全部" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="_all">优先级：全部</SelectItem>
                {PRIORITY_OPTIONS.map((p) => (
                  <SelectItem key={p} value={p}>{p}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={assigneeFilter || '_all'} onValueChange={handleAssigneeChange}>
              <SelectTrigger className="w-[120px]" data-testid="assignee-filter">
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
            <Select value={statusFilter || '_all'} onValueChange={handleStatusChange}>
              <SelectTrigger className="w-[120px]" data-testid="status-filter">
                <SelectValue placeholder="状态：全部" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="_all">状态：全部</SelectItem>
                {STATUS_OPTIONS.map((s) => (
                  <SelectItem key={s} value={s}>{s}</SelectItem>
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
              <p className="text-tertiary text-sm">暂无数据</p>
            </div>
          ) : (
            <div className="rounded-xl border border-border bg-white shadow-sm">
              <div data-testid="table-content">
                <Table>
                  <TableHeader>
                    <TableRowComp>
                      <TableHead>类型</TableHead>
                      <TableHead>编号</TableHead>
                      <TableHead>标题</TableHead>
                      <TableHead>优先级</TableHead>
                      <TableHead>负责人</TableHead>
                      <TableHead>进度</TableHead>
                      <TableHead>状态</TableHead>
                      <TableHead>预期完成</TableHead>
                      <TableHead>实际完成</TableHead>
                    </TableRowComp>
                  </TableHeader>
                  <TableBody>
                    {filteredItems.map((row) => (
                      <TableRowComp key={`${row.type}-${row.id}`}>
                        <TableCell>
                          <span
                            className={
                              row.type === 'main'
                                ? 'inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium bg-blue-50 text-blue-700'
                                : 'inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium bg-bg-alt text-secondary'
                            }
                          >
                            {row.type === 'main' ? '主事项' : '子事项'}
                          </span>
                        </TableCell>
                        <TableCell>
                          <span className="font-mono text-xs">{row.code}</span>
                        </TableCell>
                        <TableCell>
                          <Link
                            to={getItemLink(row)}
                            className="font-medium text-primary-600 hover:text-primary-700 hover:underline"
                          >
                            {row.title}
                          </Link>
                        </TableCell>
                        <TableCell>
                          <PriorityBadge priority={row.priority} />
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1.5">
                            <UserAvatar name={row.assigneeName} size="sm" />
                            <span className="text-[13px]">{row.assigneeName || '-'}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <ProgressBar value={row.completion} size="sm" showPercentage />
                        </TableCell>
                        <TableCell>
                          <StatusBadge status={row.status} />
                        </TableCell>
                        <TableCell>
                          <span
                            data-testid={`expected-date-${row.id}`}
                            className={isOverdue(row.expectedEndDate, row.status) ? 'text-error text-xs' : 'text-xs'}
                          >
                            {formatDate(row.expectedEndDate)}
                          </span>
                        </TableCell>
                        <TableCell>
                          <span className="text-xs">{formatDate(row.actualEndDate)}</span>
                        </TableCell>
                      </TableRowComp>
                    ))}
                  </TableBody>
                </Table>
              </div>
              <div className="flex items-center justify-between px-5 py-3 border-t border-border">
                <PaginationPageSize
                  pageSize={pageSize}
                  onPageSizeChange={handlePageSizeChange}
                  options={PAGE_SIZE_OPTIONS}
                  data-testid="pagination-page-size"
                />
                <div className="flex items-center gap-3">
                  <span className="text-[13px] text-tertiary" data-testid="total-count">
                    共 {totalItems} 条
                  </span>
                  <Pagination
                    currentPage={currentPage}
                    totalPages={totalPages}
                    onPageChange={setCurrentPage}
                  />
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
