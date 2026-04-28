import { useState, useMemo, useCallback, useEffect, useRef, Fragment } from 'react'
import { RotateCcw, RefreshCw } from 'lucide-react'
import { DateInput } from '@/components/ui/date-input'
import { useQuery } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { getGanttViewApi } from '@/api/views'
import type { GanttMainItem } from '@/types'
import './gantt-overrides.css'
import { useToast } from '@/components/ui/toast'

// --- Constants ---

const DAY_WIDTH = 28
const ROW_HEIGHT = 40
const ITEMS_PER_PAGE = 20

// --- Helpers ---

function getDateRange(items: GanttMainItem[]): { start: Date; end: Date } {
  const now = new Date()

  let minDate: Date | null = null
  let maxDate: Date | null = null

  for (const item of items) {
    const dates = [item.startDate, item.expectedEndDate,
      ...item.subItems.map((s) => s.startDate),
      ...item.subItems.map((s) => s.expectedEndDate),
    ]
    for (const d of dates) {
      if (!d) continue
      const dt = new Date(d)
      if (!minDate || dt < minDate) minDate = dt
      if (!maxDate || dt > maxDate) maxDate = dt
    }
  }

  // Add padding: 14 days before earliest, 30 days after latest
  const start = minDate ? new Date(minDate) : new Date(now)
  start.setDate(start.getDate() - 14)

  const end = maxDate ? new Date(maxDate) : new Date(now)
  end.setDate(end.getDate() + 30)

  // Ensure minimum range of 60 days
  const MIN_DAYS = 60
  if (daysBetween(start, end) < MIN_DAYS) {
    end.setDate(start.getDate() + MIN_DAYS)
  }

  // Cap maximum range at 365 days to avoid an unreadably wide chart
  const MAX_DAYS = 365
  if (daysBetween(start, end) > MAX_DAYS) {
    end.setTime(start.getTime() + MAX_DAYS * 86400000)
  }

  return { start, end }
}

function daysBetween(a: Date, b: Date): number {
  return Math.round((b.getTime() - a.getTime()) / 86400000)
}

function formatDateInput(d: Date): string {
  return d.toISOString().slice(0, 10)
}

function getBarClass(item: { isOverdue?: boolean; itemStatus: string; startDate: string | null }): string {
  if (!item.startDate) return 'no-data'
  if (item.isOverdue) return 'overdue'
  if (item.itemStatus === 'completed') return 'completed'
  return ''
}

// --- Main Component ---

export default function GanttViewPage() {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const { addToast } = useToast()
  const [searchKeyword, setSearchKeyword] = useState('')
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set())
  const [loadedCount, setLoadedCount] = useState(ITEMS_PER_PAGE)
  const [dateRange, setDateRange] = useState<{ start: string; end: string } | null>(null)

  const { data, isLoading, isFetching, refetch } = useQuery({
    queryKey: ['ganttView', teamId],
    queryFn: () => getGanttViewApi(teamId!),
    enabled: !!teamId,
  })

  const allItems = data?.items ?? []

  // Compute date range from data or user input
  const rangeStart = useMemo(() => {
    if (dateRange?.start) return new Date(dateRange.start)
    if (allItems.length > 0) return getDateRange(allItems).start
    const d = new Date()
    d.setDate(d.getDate() - 14)
    return d
  }, [dateRange, allItems])

  const rangeEnd = useMemo(() => {
    if (dateRange?.end) return new Date(dateRange.end)
    if (allItems.length > 0) return getDateRange(allItems).end
    const d = new Date()
    d.setDate(d.getDate() + 14)
    return d
  }, [dateRange, allItems])

  // Filter items by search keyword
  const filteredItems = useMemo(() => {
    if (!searchKeyword.trim()) return allItems
    const kw = searchKeyword.toLowerCase()
    return allItems.filter((item) => {
      if (item.title.toLowerCase().includes(kw)) return true
      return item.subItems.some((sub) => sub.title.toLowerCase().includes(kw))
    })
  }, [allItems, searchKeyword])

  // Pagination
  const visibleItems = useMemo(() => {
    return filteredItems.slice(0, loadedCount)
  }, [filteredItems, loadedCount])

  const hasMore = filteredItems.length > loadedCount

  // Reset loaded count when filter changes
  useEffect(() => {
    setLoadedCount(ITEMS_PER_PAGE)
  }, [searchKeyword])

  // Toggle expand
  const toggleExpand = useCallback((itemId: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev)
      if (next.has(itemId)) next.delete(itemId)
      else next.add(itemId)
      return next
    })
  }, [])

  const totalDays = daysBetween(rangeStart, rangeEnd) + 1

  const handleDateStartChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const v = e.target.value
    setDateRange((prev) => prev ? { ...prev, start: v } : { start: v, end: '' })
  }, [])

  const handleDateEndChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const v = e.target.value
    setDateRange((prev) => prev ? { ...prev, end: v } : { start: '', end: v })
  }, [])

  const handleLoadMore = useCallback(() => {
    setLoadedCount((prev) => prev + ITEMS_PER_PAGE)
  }, [])

  if (!teamId) {
    return (
      <div data-testid="gantt-view-page">
        <div className="p-6 text-tertiary">请先选择团队</div>
      </div>
    )
  }

  return (
    <div data-testid="gantt-view-page" className="gantt-page">
      {/* Page Header */}
      <div className="gantt-page-header">
        <h1 className="text-xl font-semibold text-primary">整体进度</h1>
        <div className="flex items-center gap-2 text-[13px] text-secondary">
          <DateInput
            data-testid="date-start"
            className="h-8 w-36"
            value={dateRange?.start ?? formatDateInput(rangeStart)}
            onChange={handleDateStartChange}
          />
          <span>至</span>
          <DateInput
            data-testid="date-end"
            className="h-8 w-36"
            value={dateRange?.end ?? formatDateInput(rangeEnd)}
            onChange={handleDateEndChange}
          />
          <button
            data-testid="reset-date-btn"
            className="h-8 w-8 inline-flex items-center justify-center rounded-md border border-border bg-white text-secondary hover:text-primary-500 hover:border-primary-500 transition-colors"
            title="重置时间"
            onClick={() => setDateRange(null)}
          >
            <RotateCcw className="h-4 w-4" />
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
      ) : allItems.length === 0 ? (
        <div className="py-12 text-center">
          <p className="text-tertiary text-sm">暂无甘特图数据</p>
        </div>
      ) : (
        <GanttChart
          items={visibleItems}
          rangeStart={rangeStart}
          rangeEnd={rangeEnd}
          totalDays={totalDays}
          expandedGroups={expandedGroups}
          searchKeyword={searchKeyword}
          onToggleExpand={toggleExpand}
          onSearchChange={setSearchKeyword}
          hasMore={hasMore}
          onLoadMore={handleLoadMore}
          onRefresh={async () => { await refetch(); addToast('数据已刷新', 'success') }}
          isFetching={isFetching}
        />
      )}
    </div>
  )
}

// --- Gantt Chart Component ---

interface GanttChartProps {
  items: GanttMainItem[]
  rangeStart: Date
  rangeEnd: Date
  totalDays: number
  expandedGroups: Set<string>
  searchKeyword: string
  onToggleExpand: (id: string) => void
  onSearchChange: (kw: string) => void
  hasMore: boolean
  onLoadMore: () => void
  onRefresh: () => void
  isFetching: boolean
}

function GanttChart({
  items,
  rangeStart,
  rangeEnd,
  totalDays,
  expandedGroups,
  searchKeyword,
  onToggleExpand,
  onSearchChange,
  hasMore,
  onLoadMore,
  onRefresh,
  isFetching,
}: GanttChartProps) {
  const bodyWidth = totalDays * DAY_WIDTH
  const headerInnerRef = useRef<HTMLDivElement>(null)
  const bodyInnerRef = useRef<HTMLDivElement>(null)
  const hscrollRef = useRef<HTMLDivElement>(null)

  // Sync header & body horizontal position with scrollbar proxy
  const handleHScroll = useCallback(() => {
    if (hscrollRef.current) {
      const offset = -hscrollRef.current.scrollLeft
      if (headerInnerRef.current) headerInnerRef.current.style.transform = `translateX(${offset}px)`
      if (bodyInnerRef.current) bodyInnerRef.current.style.transform = `translateX(${offset}px)`
    }
  }, [])

  // Auto-expand groups when search is active
  const effectiveExpanded = useMemo(() => {
    if (!searchKeyword.trim()) return expandedGroups
    // When searching, expand all groups that have matching items
    const expanded = new Set(expandedGroups)
    for (const item of items) {
      const kw = searchKeyword.toLowerCase()
      const hasMatchingSub = item.subItems.some((sub) => sub.title.toLowerCase().includes(kw))
      if (hasMatchingSub) expanded.add(item.bizKey)
    }
    return expanded
  }, [expandedGroups, searchKeyword, items])

  return (
    <div data-testid="gantt-container" className="gantt-container">
      <div className="gantt-inner">
          {/* Label Panel (sticky left) */}
          <div className="gantt-labels">
            <div className="gantt-label-header">
              <input
                type="text"
                className="label-search"
                placeholder="搜索事项标题…"
                value={searchKeyword}
                onChange={(e) => onSearchChange(e.target.value)}
              />
              <button
                data-testid="refresh-btn"
                className="h-7 w-7 inline-flex items-center justify-center rounded border border-border bg-white text-secondary hover:text-primary-500 hover:border-primary-500 transition-colors flex-shrink-0"
                title="刷新"
                onClick={onRefresh}
              >
                <RefreshCw className={`h-3.5 w-3.5 ${isFetching ? 'animate-spin' : ''}`} />
              </button>
            </div>
            {items.map((item) => (
              <GanttLabelRow
                key={item.bizKey}
                item={item}
                isExpanded={effectiveExpanded.has(item.bizKey)}
                onToggle={() => onToggleExpand(item.bizKey)}
              />
            ))}
            {hasMore && (
              <button
                data-testid="load-more-btn"
                className="load-more-btn"
                onClick={onLoadMore}
              >
                加载更多
              </button>
            )}
          </div>

          {/* Timeline Panel */}
          <div className="gantt-timeline">
            <div className="gantt-header">
              <div className="gantt-header-inner" ref={headerInnerRef} style={{ width: bodyWidth }}>
                <GanttHeaderContent rangeStart={rangeStart} totalDays={totalDays} />
              </div>
            </div>
            <div className="gantt-body">
              <div className="gantt-body-inner" ref={bodyInnerRef} style={{ width: bodyWidth, position: 'relative' }}>
                {items.map((item) => (
                  <Fragment key={item.bizKey}>
                    <GanttTimelineRow
                      itemId={item.bizKey}
                      subId={undefined}
                      startDate={item.startDate}
                      endDate={item.expectedEndDate}
                      completion={item.completion}
                      barClass={getBarClass(item)}
                      rangeStart={rangeStart}
                      totalDays={totalDays}
                      isSub={false}
                      isExpanded={effectiveExpanded.has(item.bizKey)}
                    />
                    {item.subItems.map((sub) => (
                      <GanttTimelineRow
                        key={sub.bizKey}
                        itemId={item.bizKey}
                        subId={sub.bizKey}
                        startDate={sub.startDate}
                        endDate={sub.expectedEndDate}
                        completion={sub.completion}
                        barClass={getBarClass({ itemStatus: sub.itemStatus, startDate: sub.startDate })}
                        rangeStart={rangeStart}
                        totalDays={totalDays}
                        isSub={true}
                        isExpanded={effectiveExpanded.has(item.bizKey)}
                      />
                    ))}
                  </Fragment>
                ))}
                <TodayLine rangeStart={rangeStart} totalDays={totalDays} />
              </div>
            </div>
            <div className="gantt-hscroll" ref={hscrollRef} onScroll={handleHScroll}>
              <div style={{ width: bodyWidth, height: 1 }} />
            </div>
          </div>
        </div>
      </div>
  )
}

// --- Label Row ---

interface GanttLabelRowProps {
  item: GanttMainItem
  isExpanded: boolean
  onToggle: () => void
}

function GanttLabelRow({ item, isExpanded, onToggle }: GanttLabelRowProps) {
  const hasSubs = item.subItems.length > 0

  return (
    <>
      <div className="gantt-label-row main-item" onClick={hasSubs ? onToggle : undefined}>
        <button
          data-testid={`collapse-toggle-${item.bizKey}`}
          data-hidden={hasSubs ? undefined : 'true'}
          className="collapse-toggle"
          onClick={(e) => {
            e.stopPropagation()
            onToggle()
          }}
          style={{ visibility: hasSubs ? 'visible' : 'hidden' }}
        >
          {isExpanded ? '\u25BC' : '\u25B6'}
        </button>
        <span className="label-title" title={item.title}>{item.title}</span>
      </div>
      {item.subItems.map((sub) => (
        <div
          key={sub.bizKey}
          className={`gantt-label-row sub ${isExpanded ? 'visible' : ''}`}
          style={{ display: isExpanded ? undefined : 'none' }}
        >
          <span className="label-title" title={sub.title}>{sub.title}</span>
        </div>
      ))}
    </>
  )
}

// --- Timeline Row ---

interface GanttTimelineRowProps {
  itemId: string
  subId?: string
  startDate: string | null
  endDate: string | null
  completion: number
  barClass: string
  rangeStart: Date
  totalDays: number
  isSub: boolean
  isExpanded: boolean
}

function GanttTimelineRow({
  itemId,
  subId,
  startDate,
  endDate,
  completion,
  barClass,
  rangeStart,
  totalDays,
  isSub,
  isExpanded,
}: GanttTimelineRowProps) {
  const rowId = subId ?? itemId
  const hidden = isSub && !isExpanded

  const barStyle = useMemo(() => {
    if (!startDate || !endDate) return null
    const start = new Date(startDate)
    const end = new Date(endDate)
    let startOffset = daysBetween(rangeStart, start)
    let endOffset = daysBetween(rangeStart, end)
    if (startOffset < 0) startOffset = 0
    if (endOffset > totalDays) endOffset = totalDays
    return {
      left: startOffset * DAY_WIDTH,
      width: (endOffset - startOffset + 1) * DAY_WIDTH,
    }
  }, [startDate, endDate, rangeStart, totalDays])

  return (
    <div
      data-testid={`timeline-row-${rowId}`}
      className={`gantt-row ${hidden ? 'gantt-row-hidden' : ''}`}
      style={{ display: hidden ? 'none' : undefined }}
    >
      {barStyle ? (
        <div
          data-testid={`gantt-bar-${rowId}`}
          className={`gantt-bar ${barClass}`}
          style={{ left: barStyle.left, width: barStyle.width }}
        >
          <div className="gantt-bar-fill" style={{ width: `${completion}%` }} />
          <span className="gantt-bar-percent">{Math.round(completion)}%</span>
        </div>
      ) : (
        <div
          data-testid={`gantt-bar-${rowId}`}
          className="gantt-bar no-data"
          style={{ left: 14, width: 120 }}
        >
          <span className="gantt-bar-label">未设置时间</span>
        </div>
      )}
    </div>
  )
}

// --- Today Line ---

function TodayLine({ rangeStart, totalDays }: { rangeStart: Date; totalDays: number }) {
  const today = new Date()
  const offset = daysBetween(rangeStart, today)
  if (offset < 0 || offset >= totalDays) return null

  return (
    <div
      data-testid="gantt-today-line"
      className="gantt-today"
      style={{ left: offset * DAY_WIDTH + DAY_WIDTH / 2 }}
    >
      <span className="gantt-today-tip">今天</span>
    </div>
  )
}

// --- Gantt Header Content ---

function GanttHeaderContent({ rangeStart, totalDays }: { rangeStart: Date; totalDays: number }) {
  // Build month cells
  const months: { label: string; width: number }[] = useMemo(() => {
    const result: { label: string; width: number }[] = []
    const d = new Date(rangeStart)
    while (d <= new Date(rangeStart.getTime() + totalDays * 86400000)) {
      const currentMonth = d.getMonth()
      const currentYear = d.getFullYear()
      const monthStart = new Date(d)
      while (d.getMonth() === currentMonth) {
        d.setDate(d.getDate() + 1)
      }
      let monthEnd = new Date(d)
      const rangeEndDate = new Date(rangeStart.getTime() + totalDays * 86400000)
      if (monthEnd > rangeEndDate) monthEnd = new Date(rangeEndDate)
      const span = daysBetween(monthStart, monthEnd) + 1
      result.push({
        label: `${currentYear}/${String(currentMonth + 1).padStart(2, '0')}`,
        width: span * DAY_WIDTH,
      })
    }
    return result
  }, [rangeStart, totalDays])

  // Build day cells
  const days: { date: number; isWeekend: boolean }[] = useMemo(() => {
    return Array.from({ length: totalDays }, (_, i) => {
      const d = new Date(rangeStart)
      d.setDate(d.getDate() + i)
      const dow = d.getDay()
      return { date: d.getDate(), isWeekend: dow === 0 || dow === 6 }
    })
  }, [rangeStart, totalDays])

  return (
    <>
      {/* Month row */}
      <div className="gantt-month-row">
        {months.map((m, i) => (
          <div key={i} className="gantt-month-cell" style={{ width: m.width }}>
            {m.label}
          </div>
        ))}
      </div>
      {/* Day row */}
      <div className="gantt-day-row">
        {days.map((d, i) => (
          <div key={i} className={`gantt-day ${d.isWeekend ? 'weekend' : 'weekday'}`}>
            {d.date}
          </div>
        ))}
      </div>
    </>
  )
}
