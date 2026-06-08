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
import { updateMainItemApi } from '@/api/mainItems'
import { MILESTONE_STATUSES, MILESTONE_MAP_STATUSES } from '@/lib/status'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useToast } from '@/components/ui/toast'
import StatusTransitionDropdown from '@/components/shared/StatusTransitionDropdown'
import ConfirmDialog from '@/components/shared/ConfirmDialog'
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
const NODE_WIDTH = 160 // w-40 = 160px
const NODE_GAP = 24
const NODE_UNIT = NODE_WIDTH + NODE_GAP // 184px per node
const MIN_TOTAL_DAYS = 30
const ZOOM_CONFIGS = {
  week: {
    interval: 7,
    formatLabel: (d: dayjs.Dayjs) => `${d.month() + 1}/${d.date()}`,
  },
  month: {
    interval: 30,
    formatLabel: (d: dayjs.Dayjs) => `${d.month() + 1}月`,
  },
  quarter: {
    interval: 90,
    formatLabel: (d: dayjs.Dayjs) => `Q${Math.ceil((d.month() + 1) / 3)}`,
  },
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
}

// --- Helper: calculate node position ---

interface NodePosition {
  bizKey: string
  x: number
  milestone: Milestone
}

function calculateNodePositions(
  milestones: Milestone[],
  containerWidth: number,
): { positions: NodePosition[], originDate: dayjs.Dayjs, totalDays: number } {
  if (milestones.length === 0) {
    return { positions: [], originDate: dayjs(), totalDays: MIN_TOTAL_DAYS }
  }

  // Filter milestones with valid dates
  const dated = milestones.filter((m) => m.expectedEndDate)
  if (dated.length === 0) {
    // No dates: spread evenly
    const positions = milestones.map((m, i) => ({
      bizKey: m.bizKey,
      x: i * NODE_UNIT + 40,
      milestone: m,
    }))
    return { positions, originDate: dayjs(), totalDays: MIN_TOTAL_DAYS }
  }

  const dates = dated.map((m) => dayjs(m.expectedEndDate))
  const earliest = dates.reduce((a, b) => (a.isBefore(b) ? a : b))
  const latest = dates.reduce((a, b) => (a.isAfter(b) ? a : b))

  // Extend by 15 days on each side for padding
  const originDate = earliest.subtract(15, 'day')
  const endDate = latest.add(15, 'day')
  let totalDays = endDate.diff(originDate, 'day')
  totalDays = Math.max(totalDays, MIN_TOTAL_DAYS)

  const positions = milestones.map((m) => {
    let x: number
    if (m.expectedEndDate) {
      const days = dayjs(m.expectedEndDate).diff(originDate, 'day')
      x = (days / totalDays) * containerWidth
    } else {
      // No date: place at end
      x = containerWidth
    }
    return { bizKey: m.bizKey, x, milestone: m }
  })

  return { positions, originDate, totalDays }
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
}: MilestoneTimelineProps) {
  const teamId = useTeamStore((s) => s.currentTeamId) ?? ''
  const qc = useQueryClient()
  const navigate = useNavigate()
  const canCreate = usePermission('milestone:create')
  const { addToast } = useToast()

  // State
  const [zoom, setZoom] = useState<ZoomLevel>('month')
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

  // Measure container width
  useEffect(() => {
    const el = timelineContainerRef.current
    if (!el) return
    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setContainerWidth(entry.contentRect.width)
      }
    })
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

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

  // Node positions
  const minWidth = Math.max(allMilestones.length * NODE_UNIT, containerWidth)
  const { positions, originDate, totalDays } = calculateNodePositions(
    allMilestones,
    minWidth - 80, // 40px padding on each side
  )

  // Tick marks
  const tickMarks = useMemo(() => {
    if (allMilestones.length === 0) return []
    const config = ZOOM_CONFIGS[zoom]
    const ticks: { x: number, label: string }[] = []
    const end = originDate.add(totalDays, 'day')

    for (
      let d = originDate;
      d.isBefore(end) || d.isSame(end, 'day');
      d = d.add(config.interval, 'day')
    ) {
      const days = d.diff(originDate, 'day')
      const x = (days / totalDays) * (minWidth - 80) + 40
      ticks.push({ x, label: config.formatLabel(d) })
    }
    return ticks
  }, [originDate, totalDays, minWidth, zoom, allMilestones.length])

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
          <StatusTransitionDropdown
            currentStatus={mapData?.mapStatus ?? ''}
            itemType="milestone-map"
            teamId={teamId}
            itemId={mapId}
            onStatusChanged={handleMapStatusChanged}
            disabled={isMapTerminal}
          />
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => onEditMap(mapData!)}
          >
            编辑
          </Button>
          {canDeleteMap && (
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
          ref={timelineContainerRef}
        >
          {/* Zoom controls */}
          <div className="absolute top-3 right-3 z-10 flex items-center gap-1.5 rounded-md border border-border bg-white px-2 py-1">
            <span className="text-xs text-tertiary">刻度</span>
            <div className="flex">
              {(['week', 'month', 'quarter'] as const).map((level) => (
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
                  {level === 'week' ? '周' : level === 'month' ? '月' : '季'}
                </button>
              ))}
            </div>
          </div>

          {/* Scrollable timeline track */}
          <div className="overflow-x-auto" style={{ minHeight: 400 }}>
            <div
              className="relative"
              style={{
                minWidth: `${minWidth}px`,
                padding: '28px 40px 40px',
              }}
            >
              {/* Tick marks */}
              <div
                data-testid="tick-marks"
                className="relative border-b-2 border-border"
                style={{ height: 40, marginBottom: 8 }}
              >
                {tickMarks.map((tick, i) => (
                  <div
                    key={i}
                    className="absolute flex flex-col items-center"
                    style={{
                      left: tick.x,
                      transform: 'translateX(-50%)',
                    }}
                  >
                    <div className="w-px h-4 bg-border-dark" />
                    <span className="text-[11px] text-tertiary mt-1 whitespace-nowrap">
                      {tick.label}
                    </span>
                  </div>
                ))}
              </div>

              {/* Milestone nodes layer */}
              <div className="relative" style={{ minHeight: 110 }}>
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
