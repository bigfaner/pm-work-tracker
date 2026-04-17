import { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Select, Empty, Spin } from 'antd'
import { useTeamStore } from '@/store/team'
import { getGanttViewApi } from '@/api/views'
import type { GanttMainItem } from '@/types'
import Gantt from 'frappe-gantt'
import dayjs from 'dayjs'

const STATUS_OPTIONS = [
  { value: '', label: '全部状态' },
  { value: '待开始', label: '待开始' },
  { value: '进行中', label: '进行中' },
  { value: '阻塞中', label: '阻塞中' },
  { value: '挂起', label: '挂起' },
  { value: '已延期', label: '已延期' },
  { value: '待验收', label: '待验收' },
  { value: '已完成', label: '已完成' },
  { value: '已关闭', label: '已关闭' },
]

const PRIORITY_CLASS_MAP: Record<string, string> = {
  P1: 'bar-p1',
  P2: 'bar-p2',
  P3: 'bar-p3',
}

const today = () => dayjs().format('YYYY-MM-DD')

function toGanttTask(item: GanttMainItem) {
  const classes: string[] = []
  if (item.isOverdue) {
    classes.push('bar-overdue')
  } else {
    const cls = PRIORITY_CLASS_MAP[item.priority]
    if (cls) classes.push(cls)
  }

  return {
    id: String(item.id),
    name: item.title,
    start: item.startDate || today(),
    end: item.expectedEndDate || today(),
    progress: Math.round(item.completion),
    custom_class: classes.join(' '),
  }
}

function toGanttSubTask(subItem: GanttMainItem['subItems'][number], parentId: number) {
  return {
    id: `sub-${subItem.id}`,
    name: `  └ ${subItem.title}`,
    start: subItem.startDate || today(),
    end: subItem.expectedEndDate || today(),
    progress: Math.round(subItem.completion),
    custom_class: 'bar-sub',
    dependencies: [String(parentId)],
  }
}

function buildTaskList(items: GanttMainItem[], expandedIds: Set<number>) {
  const tasks: ReturnType<typeof toGanttTask>[] = []
  for (const item of items) {
    tasks.push(toGanttTask(item))
    if (expandedIds.has(item.id) && item.subItems.length > 0) {
      for (const sub of item.subItems) {
        tasks.push(toGanttSubTask(sub, item.id))
      }
    }
  }
  return tasks
}

export default function GanttViewPage() {
  const { currentTeamId } = useTeamStore()
  const [statusFilter, setStatusFilter] = useState<string | undefined>(undefined)
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
  const ganttRef = useRef<Gantt | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['ganttView', currentTeamId, statusFilter],
    queryFn: () => getGanttViewApi(currentTeamId!, statusFilter || undefined),
    enabled: !!currentTeamId,
  })

  const items = data?.items ?? []

  // Base tasks: main items only (collapsed state)
  const baseTasks = useMemo(() => buildTaskList(items, new Set()), [items])

  // Full task list with expanded sub-items
  const ganttTasks = useMemo(() => buildTaskList(items, expandedIds), [items, expandedIds])

  // Initialize frappe-gantt when base data arrives
  useEffect(() => {
    if (baseTasks.length === 0 || !containerRef.current) return

    const container = containerRef.current
    container.innerHTML = '<svg></svg>'

    ganttRef.current = new Gantt(container, baseTasks, {
      view_mode: 'Day',
      date_format: 'YYYY-MM-DD',
      custom_popup_html: undefined,
    })

    return () => {
      if (container) {
        container.innerHTML = ''
      }
      ganttRef.current = null
    }
  }, [baseTasks])

  // Refresh when expand/collapse changes (after initial render)
  const isInitialMount = useRef(true)
  useEffect(() => {
    if (isInitialMount.current) {
      isInitialMount.current = false
      return
    }
    if (ganttRef.current) {
      ganttRef.current.refresh(ganttTasks)
    }
  }, [ganttTasks])

  const toggleExpand = useCallback((mainItemId: number) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(mainItemId)) {
        next.delete(mainItemId)
      } else {
        next.add(mainItemId)
      }
      return next
    })
  }, [])

  const handleStatusChange = useCallback((value: string) => {
    setStatusFilter(value || undefined)
    setExpandedIds(new Set())
  }, [])

  return (
    <div data-testid="gantt-view-page">
      {/* Page Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h2 style={{ margin: 0 }}>甘特图</h2>
        <Select
          data-testid="gantt-status-filter"
          style={{ width: 160 }}
          placeholder="筛选状态"
          allowClear
          options={STATUS_OPTIONS}
          onChange={handleStatusChange}
        />
      </div>

      {/* Content */}
      {isLoading ? (
        <div data-testid="gantt-loading" style={{ textAlign: 'center', padding: 60 }}>
          <Spin size="large" />
        </div>
      ) : items.length === 0 ? (
        <div data-testid="gantt-empty-state">
          <Empty description="暂无事项数据" />
        </div>
      ) : (
        <div>
          {/* Main item rows with expand/collapse */}
          <div style={{ marginBottom: 16 }}>
            {items.map((item) => (
              <div
                key={item.id}
                data-testid={`gantt-row-${item.id}`}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  padding: '6px 12px',
                  borderBottom: '1px solid #f0f0f0',
                  cursor: item.subItems.length > 0 ? 'pointer' : 'default',
                }}
                onClick={() => item.subItems.length > 0 && toggleExpand(item.id)}
              >
                {item.subItems.length > 0 && (
                  <span data-testid={`gantt-expand-${item.id}`} style={{ fontSize: 12, color: '#999' }}>
                    {expandedIds.has(item.id) ? '▼' : '▶'}
                  </span>
                )}
                <span style={{ flex: 1, fontWeight: 500 }}>{item.title}</span>
                <span style={{ fontSize: 13, color: '#666' }}>{item.completion}%</span>
              </div>
            ))}
          </div>

          {/* Gantt chart */}
          <div
            ref={containerRef}
            data-testid="gantt-chart-container"
            style={{ overflow: 'auto' }}
          />
        </div>
      )}

      {/* Inline styles for frappe-gantt and bar coloring */}
      <style>{`
        .gantt .grid-background{fill:none}.gantt .grid-header{fill:#fff;stroke:#e0e0e0;stroke-width:1.4}.gantt .grid-row{fill:#fff}.gantt .grid-row:nth-child(even){fill:#f5f5f5}.gantt .row-line{stroke:#ebeff2}.gantt .tick{stroke:#e0e0e0;stroke-width:.2}.gantt .tick.thick{stroke-width:.4}.gantt .today-highlight{fill:#fcf8e3;opacity:.5}.gantt .arrow{fill:none;stroke:#666;stroke-width:1.4}.gantt .bar{fill:#b8c2cc;stroke:#8D99A6;stroke-width:0;transition:stroke-width .3s ease;user-select:none}.gantt .bar-progress{fill:#a3a3ff}.gantt .bar-invalid{fill:transparent;stroke:#8D99A6;stroke-width:1;stroke-dasharray:5}.gantt .bar-invalid~.bar-label{fill:#555}.gantt .bar-label{fill:#fff;dominant-baseline:central;text-anchor:middle;font-size:12px;font-weight:lighter}.gantt .bar-label.big{fill:#555;text-anchor:start}.gantt .handle{fill:#ddd;cursor:ew-resize;opacity:0;visibility:hidden;transition:opacity .3s ease}.gantt .bar-wrapper{cursor:pointer;outline:none}.gantt .bar-wrapper:hover .bar{fill:#a9b5c1}.gantt .bar-wrapper:hover .bar-progress{fill:#8a8aff}.gantt .bar-wrapper:hover .handle{visibility:visible;opacity:1}.gantt .bar-wrapper.active .bar{fill:#a9b5c1}.gantt .bar-wrapper.active .bar-progress{fill:#8a8aff}.gantt .lower-text,.gantt .upper-text{font-size:12px;text-anchor:middle}.gantt .upper-text{fill:#555}.gantt .lower-text{fill:#333}.gantt .hide{display:none}.gantt-container{position:relative;overflow:auto;font-size:12px}.gantt-container .popup-wrapper{position:absolute;top:0;left:0;background:rgba(0,0,0,.8);padding:0;color:#959da5;border-radius:3px}.gantt-container .popup-wrapper .title{border-bottom:3px solid #a3a3ff;padding:10px}.gantt-container .popup-wrapper .subtitle{padding:10px;color:#dfe2e5}.gantt-container .popup-wrapper .pointer{position:absolute;height:5px;margin:0 0 0 -5px;border:5px solid transparent;border-top-color:rgba(0,0,0,.8)}
        .bar-p1 .bar{fill:#fa8c16!important}.bar-p2 .bar{fill:#1677ff!important}.bar-p3 .bar{fill:#d9d9d9!important}.bar-overdue .bar{fill:#ff4d4f!important}.bar-sub .bar{fill:#95de64!important}
      `}</style>
    </div>
  )
}
