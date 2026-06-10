import { useState, useCallback, useMemo, useRef, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, RefreshCw } from 'lucide-react'
import dayjs from 'dayjs'
import { useTeamStore } from '@/store/team'
import { usePermission } from '@/hooks/usePermission'
import {
  getMilestoneMapApi,
  listMilestonesByMapApi,
  createMilestoneApi,
  deleteMilestoneMapApi,
} from '@/api/milestones'
import { listMainItemsApi, updateMainItemApi } from '@/api/mainItems'
import { MILESTONE_STATUSES, MILESTONE_MAP_STATUSES } from '@/lib/status'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useToast } from '@/components/ui/toast'
import StatusTransitionDropdown from '@/components/shared/StatusTransitionDropdown'
import ConfirmDialog from '@/components/shared/ConfirmDialog'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { StatusTagFilter } from '@/components/shared/StatusTagFilter'
import MilestoneNode from './MilestoneNode'
import MilestoneDetailPanel from './MilestoneDetailPanel'
import CreateMilestoneDialog, {
  type MilestoneFormState,
} from './CreateMilestoneDialog'
import type { Milestone } from '@/types'

/** Data shape stored in dataTransfer during MI drag */
interface DragMI {
  miBizKey: string
  miCode: string
  sourceMilestoneKey: string
}

/** Exposed for test access — sets drag data on the global window */
export function setDragMI(data: DragMI | null) {
  ;(window as unknown as Record<string, unknown>).__dragMI = data
}

/** Read current drag MI data */
export function getDragMI(): DragMI | null {
  return (window as unknown as Record<string, unknown>).__dragMI as DragMI | null
}

// --- Constants ---

const DEBOUNCE_MS = 300
const NODE_WIDTH = 240 // w-60 = 240px
const ZOOM_CONFIGS = {
  compact: { gap: 8, label: '紧凑' },
  standard: { gap: 24, label: '标准' },
  relaxed: { gap: 48, label: '宽松' },
} as const

type ZoomLevel = keyof typeof ZOOM_CONFIGS;

const MILESTONE_STATUS_OPTIONS = Object.entries(MILESTONE_STATUSES).map(
  ([value, def]) => ({ value, label: def.name }),
)

const EMPTY_MILESTONE_FORM: MilestoneFormState = {
  milestoneName: '',
  expectedEndDate: '',
  milestoneDesc: '',
}

// --- Interfaces ---

interface MilestoneTimelineProps {
  mapId: string
  onEditMap: (map: import('@/types').MilestoneMap) => void
  onEditMilestone: (milestone: Milestone) => void
  onQuickAdd: (milestone: Milestone) => void
  onBindExisting: (milestone: Milestone) => void
}

// --- Helper: calculate node position ---

interface NodePosition {
  bizKey: string
  x: number
  milestone: Milestone
}

function calculateNodePositions(
  milestones: Milestone[],
  gap: number,
  containerWidth: number,
): { positions: NodePosition[], contentWidth: number } {
  if (milestones.length === 0) {
    return { positions: [], contentWidth: 0 }
  }

  // Sort by date (earliest first, no-date at end)
  const sorted = [...milestones].sort((a, b) => {
    if (a.expectedEndDate && b.expectedEndDate) {
      return a.expectedEndDate.localeCompare(b.expectedEndDate)
    }
    if (a.expectedEndDate) return -1
    if (b.expectedEndDate) return 1
    return 0
  })

  const nodeUnit = NODE_WIDTH + gap
  const dated = sorted.filter((m) => m.expectedEndDate)

  // Try date-proportional layout if ≥2 dated milestones with different dates
  if (dated.length >= 2) {
    const dates = dated.map((m) => dayjs(m.expectedEndDate!))
    const earliest = dates.reduce((a, b) => (a.isBefore(b) ? a : b))
    const latest = dates.reduce((a, b) => (a.isAfter(b) ? a : b))
    const totalDays = latest.diff(earliest, 'day')

    if (totalDays > 0 && containerWidth > 0) {
      const spanWidth = containerWidth - NODE_WIDTH
      const propXs = sorted.map((m) => {
        if (!m.expectedEndDate) return containerWidth - NODE_WIDTH / 2
        const days = dayjs(m.expectedEndDate).diff(earliest, 'day')
        return (days / totalDays) * spanWidth + NODE_WIDTH / 2
      })

      // Check for overlaps
      let hasOverlap = false
      for (let i = 1; i < propXs.length; i++) {
        if (propXs[i] - propXs[i - 1] < nodeUnit) {
          hasOverlap = true
          break
        }
      }

      if (!hasOverlap) {
        const positions = sorted.map((m, i) => ({
          bizKey: m.bizKey,
          x: propXs[i],
          milestone: m,
        }))
        return { positions, contentWidth: containerWidth }
      }
    }
  }

  // Fall back: even spacing across container width
  const count = sorted.length
  const startX = NODE_WIDTH / 2
  const endX = Math.max(containerWidth - NODE_WIDTH / 2, startX + nodeUnit * (count - 1))
  const positions = sorted.map((m, i) => ({
    bizKey: m.bizKey,
    x: count === 1 ? startX : startX + i * ((endX - startX) / (count - 1)),
    milestone: m,
  }))
  const contentWidth = Math.max(containerWidth, endX + NODE_WIDTH / 2)
  return { positions, contentWidth }
}

// --- Skeleton ---

function TimelineSkeleton() {
  return (
    <div className="animate-pulse space-y-4">
      <div className="flex items-center gap-2">
        <div className="h-4 w-24 bg-bg-alt rounded" />
        <div className="h-4 w-16 bg-bg-alt rounded" />
      </div>
      <div className="h-24 bg-bg-alt rounded-xl" />
      <div className="h-60 bg-bg-alt rounded-xl" />
    </div>
  )
}

// --- Main Component ---

export default function MilestoneTimeline({
  mapId,
  onEditMap,
  onEditMilestone,
  onQuickAdd,
  onBindExisting,
}: MilestoneTimelineProps) {
  const teamId = useTeamStore((s) => s.currentTeamId) ?? ''
  const qc = useQueryClient()
  const navigate = useNavigate()
  const canCreate = usePermission('milestone:create')
  const canUpdate = usePermission('milestone:update')
  const canDeletePerm = usePermission('milestone:delete')
  const { addToast } = useToast()

  // State
  const [zoom, setZoom] = useState<ZoomLevel>('standard')
  const [searchText, setSearchText] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<string[]>([])
  const [selectedMilestoneId, setSelectedMilestoneId] = useState<string | null>(
    null,
  )
  const [detailPanelOpen, setDetailPanelOpen] = useState(false)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [createForm, setCreateForm] =
    useState<MilestoneFormState>(EMPTY_MILESTONE_FORM)
  const [deleteMapConfirmOpen, setDeleteMapConfirmOpen] = useState(false)
  const [containerWidth, setContainerWidth] = useState(800)
  const timelineContainerRef = useRef<HTMLDivElement>(null)
  const scrollTrackRef = useRef<HTMLDivElement>(null)
  const [canScrollLeft, setCanScrollLeft] = useState(false)
  const [canScrollRight, setCanScrollRight] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout>>()
  const [dragOverMilestoneKey, setDragOverMilestoneKey] = useState<
    string | null
  >(null)

  // Debounce search
  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(
      () => setDebouncedSearch(searchText),
      DEBOUNCE_MS,
    )
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [searchText])

  // Scroll tracking helpers (defined before useEffect that uses them)
  const updateScrollState = useCallback(() => {
    const el = scrollTrackRef.current
    if (!el) return
    setCanScrollLeft(el.scrollLeft > 4)
    setCanScrollRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 4)
  }, [])

  const scrollBy = useCallback((direction: 'left' | 'right') => {
    const el = scrollTrackRef.current
    if (!el) return
    el.scrollBy({ left: direction === 'left' ? -300 : 300, behavior: 'smooth' })
  }, [])

  // Measure container width + update scroll state on resize
  useEffect(() => {
    const el = timelineContainerRef.current
    if (!el) return
    const observer = new ResizeObserver(() => {
      setContainerWidth(el.clientWidth)
      requestAnimationFrame(() => updateScrollState())
    })
    observer.observe(el)
    return () => observer.disconnect()
  }, [updateScrollState])

  // Data fetching
  const { data: mapData, isLoading: mapLoading } = useQuery({
    queryKey: ['milestoneMap', teamId, mapId],
    queryFn: () => getMilestoneMapApi(teamId, mapId),
    enabled: !!teamId && !!mapId,
  })

  const {
    data: milestonesData,
    isLoading: milestonesLoading,
    refetch: refetchMilestones,
    isFetching: isFetchingMilestones,
  } = useQuery({
    queryKey: ['milestones', teamId, mapId],
    queryFn: () => listMilestonesByMapApi(teamId, mapId),
    enabled: !!teamId && !!mapId,
  })

  const allMilestones: Milestone[] = milestonesData?.items ?? []

  // Track scroll for arrow buttons
  useEffect(() => {
    const el = scrollTrackRef.current
    if (!el) return
    // Use rAF to ensure DOM has reflowed after layout changes
    const raf = requestAnimationFrame(() => updateScrollState())
    el.addEventListener('scroll', updateScrollState, { passive: true })
    return () => {
      cancelAnimationFrame(raf)
      el.removeEventListener('scroll', updateScrollState)
    }
  }, [updateScrollState, allMilestones, zoom])

  // Fetch all MIs for this map (grouped by milestoneKey client-side)
  const { data: miData } = useQuery({
    queryKey: ['mapMIs', teamId, mapId],
    queryFn: () =>
      listMainItemsApi(teamId, { pageSize: 200 }).then((res) => res.items),
    enabled: !!teamId && !!mapId,
  })

  // Group MIs by milestoneKey
  const misByMilestone = useMemo(() => {
    const map: Record<string, NonNullable<typeof miData>> = {}
    if (!miData) return map
    for (const mi of miData) {
      if (mi.milestoneKey) {
        if (!map[mi.milestoneKey]) map[mi.milestoneKey] = []
        map[mi.milestoneKey]!.push(mi)
      }
    }
    return map
  }, [miData])

  // Client-side filtering
  const filteredMilestones = useMemo(() => {
    let items = allMilestones
    if (debouncedSearch.trim()) {
      const q = debouncedSearch.trim().toLowerCase()
      items = items.filter((m) => m.milestoneName.toLowerCase().includes(q))
    }
    if (statusFilter.length > 0) {
      items = items.filter((m) => statusFilter.includes(m.milestoneStatus))
    }
    return items
  }, [allMilestones, debouncedSearch, statusFilter])

  // Node positions — based on filtered milestones so timeline re-renders on filter
  const config = ZOOM_CONFIGS[zoom]
  const paddingH = 40
  const usableWidth = Math.max(containerWidth - paddingH * 2, NODE_WIDTH)
  const { positions, contentWidth } = calculateNodePositions(filteredMilestones, config.gap, usableWidth)
  const minWidth = Math.max(contentWidth + paddingH * 2, containerWidth)

  // Tick marks: one per milestone, aligned with node center
  const tickMarks = useMemo(() => {
    return positions.map((pos) => {
      const label = pos.milestone.expectedEndDate ?? '未设置'
      return { x: pos.x, label, bizKey: pos.bizKey }
    })
  }, [positions])

  // Map terminal state check
  const isMapTerminal = mapData
    ? ((MILESTONE_MAP_STATUSES as Record<string, { terminal: boolean }>)[
        mapData.mapStatus
      ]?.terminal ?? false)
    : false

  const canDeleteMap = mapData
    ? ['planning', 'reviewed', 'ready'].includes(mapData.mapStatus)
    : false

  // Mutations
  const createMilestoneMutation = useMutation({
    mutationFn: (form: MilestoneFormState) =>
      createMilestoneApi(teamId, mapId, {
        milestoneName: form.milestoneName.trim(),
        expectedEndDate: form.expectedEndDate,
        ...(form.milestoneDesc && { milestoneDesc: form.milestoneDesc }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestones', teamId, mapId] })
      qc.invalidateQueries({ queryKey: ['milestoneMap', teamId, mapId] })
      setCreateDialogOpen(false)
      setCreateForm(EMPTY_MILESTONE_FORM)
    },
  })

  const deleteMapMutation = useMutation({
    mutationFn: () => deleteMilestoneMapApi(teamId, mapId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestoneMaps', teamId] })
      navigate('/milestones')
    },
  })

  // Rebind MI to a different milestone via drag-and-drop (AC-5)
  const rebindMutation = useMutation({
    mutationFn: ({
      miBizKey,
      targetMilestoneKey,
    }: {
      miBizKey: string
      targetMilestoneKey: string
    }) => updateMainItemApi(teamId, miBizKey, { milestoneKey: targetMilestoneKey }),
    onSuccess: (_data, variables) => {
      const { miBizKey, targetMilestoneKey } = variables
      qc.invalidateQueries({ queryKey: ['milestones', teamId, mapId] })
      qc.invalidateQueries({ queryKey: ['milestoneMIs', teamId] })
      qc.invalidateQueries({ queryKey: ['milestone', teamId] })
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })

      // Find target milestone name for toast
      const targetName =
        allMilestones.find((m) => m.bizKey === targetMilestoneKey)
          ?.milestoneName ?? targetMilestoneKey
      addToast(`已将事项移至「${targetName}」`, 'success')

      // Undo toast — rebind back to source within 5s
      const dragData = getDragMI()
      if (dragData && dragData.sourceMilestoneKey !== targetMilestoneKey) {
        const undoTimer = setTimeout(() => {
          setDragMI(null)
        }, 5000)
        addToast('点击撤销可恢复原绑定', 'default')
        // Store undo info for test access
        ;(window as unknown as Record<string, unknown>).__lastUndoInfo = {
          miBizKey,
          sourceMilestoneKey: dragData.sourceMilestoneKey,
          targetMilestoneKey,
          timer: undoTimer,
        }
      }
    },
  })

  const handleMapStatusChanged = useCallback(() => {
    qc.invalidateQueries({ queryKey: ['milestoneMap', teamId, mapId] })
    qc.invalidateQueries({ queryKey: ['milestones', teamId, mapId] })
    qc.invalidateQueries({
      queryKey: ['milestoneMapTransitions', teamId, mapId],
    })
  }, [qc, teamId, mapId])

  const handleNodeClick = useCallback((milestone: Milestone) => {
    setSelectedMilestoneId(milestone.bizKey)
    setDetailPanelOpen(true)
  }, [])

  const handleCloseDetailPanel = useCallback(() => {
    setDetailPanelOpen(false)
    // Keep selectedId for focus return
  }, [])

  const handleResetFilters = useCallback(() => {
    setSearchText('')
    setDebouncedSearch('')
    setStatusFilter([])
  }, [])

  const handleRefresh = useCallback(() => {
    refetchMilestones()
  }, [refetchMilestones])

  // DnD handlers for MI rebinding (AC-5)
  const handleNodeDragOver = useCallback(
    (e: React.DragEvent, milestoneBizKey: string) => {
      if (!getDragMI()) return
      if (getDragMI()?.sourceMilestoneKey === milestoneBizKey) return
      e.preventDefault()
      e.dataTransfer.dropEffect = 'move'
      setDragOverMilestoneKey(milestoneBizKey)
    },
    [],
  )

  const handleNodeDragLeave = useCallback(() => {
    setDragOverMilestoneKey(null)
  }, [])

  const handleNodeDrop = useCallback(
    (e: React.DragEvent, targetMilestoneKey: string) => {
      e.preventDefault()
      setDragOverMilestoneKey(null)
      const dragData = getDragMI()
      if (!dragData) return
      if (dragData.sourceMilestoneKey === targetMilestoneKey) return
      rebindMutation.mutate({
        miBizKey: dragData.miBizKey,
        targetMilestoneKey,
      })
    },
    [rebindMutation],
  )

  const isLoading = mapLoading || milestonesLoading

  if (isLoading) return <TimelineSkeleton />

  // Error state
  if (!mapData && !mapLoading) {
    return (
      <div className="py-12 text-center">
        <p className="text-tertiary text-sm">加载失败，请重试</p>
        <Button
          variant="secondary"
          size="sm"
          className="mt-3"
          onClick={() => refetchMilestones()}
        >
          重试
        </Button>
      </div>
    )
  }

  return (
    <div data-testid="milestone-timeline">
      {/* Breadcrumb */}
      <nav className="mb-4 text-sm text-secondary" aria-label="breadcrumb">
        <Link to="/milestones" className="hover:text-primary">
          里程碑图
        </Link>
        <span className="mx-1.5">/</span>
        <span className="text-primary">{mapData?.mapName}</span>
      </nav>

      {/* Title area */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <h1 className="text-lg font-semibold text-primary">
            {mapData?.mapName}
          </h1>
          {canUpdate && (
            <StatusTransitionDropdown
              currentStatus={mapData?.mapStatus ?? ''}
              itemType="milestone-map"
              teamId={teamId}
              itemId={mapId}
              onStatusChanged={handleMapStatusChanged}
              disabled={isMapTerminal}
            />
          )}
        </div>
        <div className="flex items-center gap-2">
          {canUpdate && (
            <Button
              variant="secondary"
              size="sm"
              onClick={() => onEditMap(mapData!)}
            >
              编辑
            </Button>
          )}
          {canDeleteMap && canDeletePerm && (
            <Button
              variant="danger"
              size="sm"
              onClick={() => setDeleteMapConfirmOpen(true)}
              data-testid="delete-map-btn"
            >
              删除
            </Button>
          )}
        </div>
      </div>

      {/* Basic info card */}
      <div className="rounded-xl border border-border bg-white p-5 mb-4">
        <div className="flex items-center justify-between">
          <div>
            <span className="text-xs text-tertiary block mb-1">负责人</span>
            <span className="text-sm font-medium text-primary">
              {mapData?.assigneeName || '-'}
            </span>
          </div>
          <div>
            <span className="text-xs text-tertiary block mb-1">计划开始</span>
            <span className="text-sm text-primary">
              {mapData?.planStartDate
                ? mapData.planStartDate.slice(0, 10)
                : '-'}
            </span>
          </div>
          <div>
            <span className="text-xs text-tertiary block mb-1">计划完成</span>
            <span className="text-sm text-primary">
              {mapData?.expectedEndDate
                ? mapData.expectedEndDate.slice(0, 10)
                : '-'}
            </span>
          </div>
          <div>
            <span className="text-xs text-tertiary block mb-1">整体进度</span>
            <span className="text-sm font-medium text-primary">
              {Math.round(mapData?.overallProgress ?? 0)}%
            </span>
          </div>
        </div>
        {mapData?.mapDesc && (
          <div
            className="mt-4 pt-4 border-t border-border text-sm text-secondary"
            style={{
              display: '-webkit-box',
              WebkitLineClamp: 3,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
            }}
          >
            {mapData.mapDesc}
          </div>
        )}
      </div>

      {/* Filter bar + actions */}
      <div className="flex items-center justify-between mb-4 flex-wrap gap-2">
        <div className="flex items-center gap-3 flex-wrap">
          <Input
            placeholder="搜索里程碑..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            className="w-48"
            data-testid="search-milestones-input"
          />
          <StatusTagFilter
            options={MILESTONE_STATUS_OPTIONS}
            selected={statusFilter}
            onChange={setStatusFilter}
          />
          <Button
            variant="secondary"
            size="sm"
            onClick={handleResetFilters}
            data-testid="reset-filters-btn"
          >
            重置
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleRefresh}
            disabled={isFetchingMilestones}
            data-testid="refresh-btn"
          >
            <RefreshCw
              size={14}
              className={isFetchingMilestones ? 'animate-spin' : ''}
            />
            刷新
          </Button>
        </div>
        <div className="flex items-center gap-2">
          {canCreate && !isMapTerminal && (
            <Button
              size="sm"
              onClick={() => setCreateDialogOpen(true)}
              data-testid="create-milestone-btn"
            >
              <Plus className="w-4 h-4" />
              创建里程碑
            </Button>
          )}
        </div>
      </div>

      {/* Timeline area */}
      {allMilestones.length === 0 ? (
        <div className="py-12 text-center">
          <p className="text-tertiary text-sm">暂无里程碑</p>
          {canCreate && !isMapTerminal && (
            <Button
              size="sm"
              className="mt-3"
              onClick={() => setCreateDialogOpen(true)}
            >
              <Plus className="w-4 h-4" />
              创建里程碑
            </Button>
          )}
        </div>
      ) : (
        <div
          className="relative border border-border rounded-xl bg-white"
          style={{ height: 'calc(100vh - 320px)', minHeight: 360, maxHeight: 560 }}
          ref={timelineContainerRef}
        >
          {/* Zoom controls — fully transparent, inside container top-right */}
          <div className="absolute top-1 right-3 z-10 flex items-center gap-1.5 rounded-md px-2 py-1 bg-transparent">
            <span className="text-xs text-tertiary">间距</span>
            <div className="flex">
              {(['compact', 'standard', 'relaxed'] as const).map((level) => (
                <button
                  key={level}
                  data-testid={`zoom-${level}`}
                  onClick={() => setZoom(level)}
                  className={`px-2.5 py-0.5 text-xs transition-colors ${
                    zoom === level
                      ? 'text-primary font-medium'
                      : 'text-tertiary hover:text-primary'
                  }`}
                >
                  {ZOOM_CONFIGS[level].label}
                </button>
              ))}
            </div>
          </div>

          {/* Scroll arrow buttons */}
          {canScrollLeft && (
            <button
              data-testid="scroll-left"
              onClick={() => scrollBy('left')}
              className="absolute left-0 top-1/2 -translate-y-1/2 z-20 w-14 h-14 flex items-center justify-center rounded-full bg-white/80 shadow cursor-pointer"
            >
              <svg width="40" height="40" viewBox="0 0 20 20" fill="none">
                <path d="M12 4L6 10L12 16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-tertiary" />
              </svg>
            </button>
          )}
          {canScrollRight && (
            <button
              data-testid="scroll-right"
              onClick={() => scrollBy('right')}
              className="absolute right-0 top-1/2 -translate-y-1/2 z-20 w-14 h-14 flex items-center justify-center rounded-full bg-white/80 shadow cursor-pointer"
            >
              <svg width="40" height="40" viewBox="0 0 20 20" fill="none">
                <path d="M8 4L14 10L8 16" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-tertiary" />
              </svg>
            </button>
          )}

          {/* Scrollable timeline track */}
          <div className="overflow-x-auto h-full" ref={scrollTrackRef}>
            <div
              className="relative"
              style={{
                minWidth: `${minWidth}px`,
                padding: '28px 40px 40px',
              }}
            >
              {/* Tick marks: date labels + dots + connectors to nodes */}
              <div
                data-testid="tick-marks"
                className="relative overflow-visible"
                style={{ height: 48, marginBottom: 0 }}
              >
                {/* Timeline line: full width edge to edge */}
                <div className="absolute left-0 right-0 border-b-2 border-border" style={{ top: 24 }} />

                {/* Date above dot, then dot on the line */}
                {tickMarks.map((tick, i) => (
                  <div
                    key={i}
                    className="absolute flex flex-col items-center"
                    style={{
                      left: tick.x,
                      transform: 'translateX(-50%)',
                    }}
                  >
                    <span className="text-[11px] text-tertiary whitespace-nowrap leading-4">
                      {tick.label}
                    </span>
                    <svg width="16" height="16" className="shrink-0">
                      <circle cx="8" cy="8" r="6" data-testid="tick-dot" className="fill-border-dark" />
                    </svg>
                  </div>
                ))}
                {/* Connector lines from dot bottom to node top border */}
                {tickMarks.map((tick) => (
                  <div
                    key={`conn-${tick.bizKey}`}
                    data-testid="tick-connector"
                    className="absolute left-0 w-px bg-border-dark"
                    style={{
                      left: tick.x,
                      top: 32,
                      height: 24,
                    }}
                  />
                ))}
              </div>

              {/* Milestone nodes layer */}
              <div className="relative" style={{ minHeight: 160 }}>
                {/* Arrows between consecutive nodes */}
                {positions.length > 1 && (
                  <svg
                    className="absolute top-0 left-0 w-full pointer-events-none"
                    style={{ height: 160, overflow: 'visible' }}
                  >
                    <defs>
                      <marker id="arrowhead" markerWidth="6" markerHeight="4" refX="5" refY="2" orient="auto">
                        <polygon points="0 0, 6 2, 0 4" className="fill-tertiary" />
                      </marker>
                    </defs>
                    {positions.slice(0, -1).map((pos, i) => {
                      const next = positions[i + 1]
                      const startX = pos.x + NODE_WIDTH / 2 + 2
                      const endX = next.x - NODE_WIDTH / 2 - 2
                      if (endX <= startX + 8) return null
                      return (
                        <line
                          key={`arrow-${pos.bizKey}`}
                          data-testid="timeline-arrow"
                          x1={startX}
                          y1={80}
                          x2={endX}
                          y2={80}
                          className="stroke-tertiary"
                          strokeWidth={1}
                          markerEnd="url(#arrowhead)"
                        />
                      )
                    })}
                  </svg>
                )}
                {filteredMilestones.map((m) => {
                  const pos = positions.find((p) => p.bizKey === m.bizKey)
                  if (!pos) return null
                  return (
                    <div
                      key={m.bizKey}
                      className="absolute top-0"
                      style={{
                        left: pos.x - NODE_WIDTH / 2,
                        transition: 'left 200ms ease',
                      }}
                    >
                      <MilestoneNode
                        milestone={m}
                        selected={selectedMilestoneId === m.bizKey}
                        onClick={() => handleNodeClick(m)}
                        isDragOver={dragOverMilestoneKey === m.bizKey}
                        onDragOver={(e) => handleNodeDragOver(e, m.bizKey)}
                        onDragLeave={handleNodeDragLeave}
                        onDrop={(e) => handleNodeDrop(e, m.bizKey)}
                      />
                    </div>
                  )
                })}
              </div>

              {/* MI layer: items below each milestone */}
              <div className="relative" data-testid="mi-layer">
                {filteredMilestones.map((m) => {
                  const pos = positions.find((p) => p.bizKey === m.bizKey)
                  if (!pos) return null
                  const items = misByMilestone[m.bizKey]
                  if (!items || items.length === 0) return null
                  return (
                    <div
                      key={`mi-${m.bizKey}`}
                      className="absolute top-0"
                      style={{
                        left: pos.x - NODE_WIDTH / 2,
                        width: NODE_WIDTH,
                      }}
                    >
                      <div
                        style={{
                          width: 1,
                          height: 16,
                          borderLeft: '1px dashed var(--text-tertiary)',
                          margin: '0 auto 8px',
                        }}
                      />
                      {items.map((mi) => (
                        <div
                          key={mi.bizKey}
                          data-testid={`mi-item-${mi.bizKey}`}
                          draggable
                          onDragStart={() => {
                            setDragMI({
                              miBizKey: mi.bizKey,
                              miCode: mi.code,
                              sourceMilestoneKey: m.bizKey,
                            })
                          }}
                          onDragEnd={() => {
                            setTimeout(() => setDragMI(null), 200)
                          }}
                          onClick={() => navigate(`/items/${mi.bizKey}`)}
                          className="flex items-center gap-1.5 rounded-md px-2 py-1 text-xs cursor-pointer hover:bg-bg-alt"
                          style={{ marginBottom: 3 }}
                        >
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <span className="text-primary truncate">
                                {mi.title}
                              </span>
                            </TooltipTrigger>
                            <TooltipContent side="bottom" className="max-w-sm !overflow-visible whitespace-normal break-words">
                              {mi.title}
                            </TooltipContent>
                          </Tooltip>
                          <span className="text-tertiary text-[11px] shrink-0">
                            {Math.round(mi.completion)}%
                          </span>
                        </div>
                      ))}
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Create milestone dialog */}
      <CreateMilestoneDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        form={createForm}
        onFormChange={setCreateForm}
        onSubmit={() => createMilestoneMutation.mutate(createForm)}
        isPending={createMilestoneMutation.isPending}
      />

      {/* Detail panel */}
      {selectedMilestoneId && (
        <MilestoneDetailPanel
          open={detailPanelOpen}
          onClose={handleCloseDetailPanel}
          milestoneId={selectedMilestoneId}
          onEdit={onEditMilestone}
          onQuickAdd={onQuickAdd}
          onBindExisting={onBindExisting}
          onDeleted={() => {
            setDetailPanelOpen(false)
            refetchMilestones()
          }}
        />
      )}

      {/* Delete map confirmation */}
      <ConfirmDialog
        open={deleteMapConfirmOpen}
        onOpenChange={setDeleteMapConfirmOpen}
        title={`确定删除里程碑图 "${mapData?.mapName ?? ''}"？`}
        description="所有里程碑及其关联事项绑定将被清除，数据不可恢复。"
        confirmLabel="确认删除"
        confirmVariant="danger"
        onConfirm={() => deleteMapMutation.mutate()}
      />
    </div>
  )
}
