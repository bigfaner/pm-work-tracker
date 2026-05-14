import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plus,
  RefreshCw,
  AlertCircle,
  MapPin,
  ChevronLeft,
  ArrowRight,
} from 'lucide-react'
import { useTeamStore } from '@/store/team'
import { usePermission } from '@/hooks/usePermission'
import {
  listMilestoneMapsApi,
  createMilestoneMapApi,
  listMilestonesByMapApi,
  createMilestoneApi,
  updateMilestoneApi,
} from '@/api/milestones'
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
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Toggle } from '@/components/ui/toggle'
import ProgressBar from '@/components/shared/ProgressBar'
import { useToast } from '@/components/ui/toast'
import MilestoneDetailPanel from './milestones/MilestoneDetailPanel'
import CreateEditMilestoneDialog from './milestones/CreateEditMilestoneDialog'
import type { MilestoneMap, Milestone } from '@/types'

// --- Constants ---

const MAP_STATUSES = [
  { code: '', name: '全部' },
  { code: 'planning', name: '规划中' },
  { code: 'reviewed', name: '已评审' },
  { code: 'ready', name: '待实施' },
  { code: 'executing', name: '实施中' },
  { code: 'completed', name: '已完成' },
] as const

const MAP_STATUS_BADGE_VARIANT: Record<string, string> = {
  planning: 'default',
  reviewed: 'warning',
  ready: 'warning',
  executing: 'primary',
  completed: 'success',
}

const ZOOM_LEVELS = [
  { key: 'week', label: '周', pxPerDay: 80, tickDays: 7 },
  { key: 'month', label: '月', pxPerDay: 20, tickDays: 30 },
  { key: 'quarter', label: '季', pxPerDay: 5, tickDays: 90 },
] as const

type ZoomLevel = (typeof ZOOM_LEVELS)[number]['key']

// --- Skeleton ---

function SkeletonCards() {
  return (
    <div data-testid="maps-loading" className="grid grid-cols-[repeat(auto-fill,minmax(340px,1fr))] gap-4">
      {Array.from({ length: 3 }).map((_, i) => (
        <div
          key={i}
          className="h-48 rounded-xl border border-border bg-white animate-pulse"
        />
      ))}
    </div>
  )
}

// --- Map Card ---

function MapCard({
  map,
  onClick,
}: {
  map: MilestoneMap
  onClick: () => void
}) {
  const badgeVariant =
    MAP_STATUS_BADGE_VARIANT[map.mapStatus] || 'default'

  return (
    <button
      type="button"
      onClick={onClick}
      className="flex flex-col gap-2.5 p-4 rounded-xl border border-border bg-white text-left hover:bg-bg-alt transition-colors cursor-pointer w-full"
      data-testid="map-card"
    >
      <div className="flex items-center justify-between">
        <span className="text-[15px] font-medium text-primary truncate" data-testid="map-card-title">
          {map.mapName}
        </span>
        <Badge variant={badgeVariant as 'default' | 'warning' | 'primary' | 'success'} data-testid="badge-status">
          {map.statusName}
        </Badge>
      </div>
      {map.mapDesc && (
        <p className="text-[13px] text-tertiary line-clamp-1">{map.mapDesc}</p>
      )}
      <div className="flex items-center gap-3 text-[13px] text-secondary">
        <span>
          {map.milestoneCount} 里程碑
        </span>
        <span>
          {map.itemCount} 事项
        </span>
      </div>
      <ProgressBar value={map.overallProgress} size="sm" showPercentage />
      {/* Mini timeline thumbnail */}
      <div className="flex items-center gap-1 mt-1">
        {Array.from({ length: Math.min(map.milestoneCount, 6) }).map((_, i) => (
          <span key={i} className="flex items-center">
            <span
              className={`w-2 h-2 rounded-full ${
                i < Math.ceil((map.overallProgress / 100) * map.milestoneCount)
                  ? 'bg-success'
                  : 'bg-border'
              }`}
            />
            {i < Math.min(map.milestoneCount, 6) - 1 && (
              <span className="w-3 h-[1px] bg-border" />
            )}
          </span>
        ))}
      </div>
    </button>
  )
}

// --- Create Map Dialog ---

function CreateMapDialog({
  open,
  onOpenChange,
  teamId,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  teamId: string
}) {
  const [mapName, setMapName] = useState('')
  const [mapDesc, setMapDesc] = useState('')
  const [nameError, setNameError] = useState('')
  const queryClient = useQueryClient()
  const { addToast } = useToast()

  const createMutation = useMutation({
    mutationFn: () =>
      createMilestoneMapApi(teamId, {
        mapName: mapName.trim(),
        mapDesc: mapDesc.trim() || undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['milestone-maps', teamId] })
      addToast('里程碑图创建成功', 'success')
      resetAndClose()
    },
    onError: () => {
      addToast('创建失败，请重试', 'error')
    },
  })

  function resetAndClose() {
    setMapName('')
    setMapDesc('')
    setNameError('')
    onOpenChange(false)
  }

  function handleConfirm() {
    if (!mapName.trim()) {
      setNameError('请输入名称')
      return
    }
    createMutation.mutate()
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>创建里程碑图</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-[13px] font-medium text-secondary mb-1.5">
              名称 <span className="text-error">*</span>
            </label>
            <Input
              value={mapName}
              onChange={(e) => {
                setMapName(e.target.value)
                if (e.target.value.trim()) setNameError('')
              }}
              placeholder="请输入里程碑图名称"
              maxLength={100}
              data-testid="input-map-name"
            />
            {nameError && (
              <p className="text-[12px] text-error-text mt-1">{nameError}</p>
            )}
          </div>
          <div>
            <label className="block text-[13px] font-medium text-secondary mb-1.5">
              描述
            </label>
            <Textarea
              value={mapDesc}
              onChange={(e) => setMapDesc(e.target.value)}
              placeholder="请输入描述（可选）"
              className="min-h-[80px]"
              data-testid="input-map-description"
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={resetAndClose}>
            取消
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={createMutation.isPending}
            data-testid="btn-confirm"
          >
            {createMutation.isPending ? '保存中...' : '确认'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// --- Timeline View ---

function TimelineView({
  mapId,
  mapName,
  onBack,
}: {
  mapId: string
  mapName: string
  onBack: () => void
}) {
  const teamId = useTeamStore((s) => s.currentTeamId) ?? ''
  const canCreate = usePermission('milestone:create')
  const [zoom, setZoom] = useState<ZoomLevel>('month')
  const zoomConfig = ZOOM_LEVELS.find((z) => z.key === zoom)!

  // Detail panel state
  const [selectedMilestoneKey, setSelectedMilestoneKey] = useState<string | null>(null)

  // Create/edit milestone dialog state
  const [milestoneDialogOpen, setMilestoneDialogOpen] = useState(false)
  const [milestoneDialogMode, setMilestoneDialogMode] = useState<'create' | 'edit'>('create')
  const [editingMilestone, setEditingMilestone] = useState<Milestone | null>(null)

  const qc = useQueryClient()
  const { addToast } = useToast()

  const {
    data: milestonesData,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['milestones-by-map', teamId, mapId],
    queryFn: () => listMilestonesByMapApi(teamId, mapId),
    enabled: !!teamId && !!mapId,
  })

  const milestones = milestonesData?.items ?? []

  // Create milestone mutation
  const createMilestoneMutation = useMutation({
    mutationFn: (data: { milestoneName: string, expectedEndDate: string }) =>
      createMilestoneApi(teamId, mapId, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestones-by-map', teamId, mapId] })
      addToast('里程碑创建成功', 'success')
      setMilestoneDialogOpen(false)
    },
    onError: () => {
      addToast('创建失败，请重试', 'error')
    },
  })

  // Update milestone mutation
  const updateMilestoneMutation = useMutation({
    mutationFn: (data: { bizKey: string, milestoneName: string, expectedEndDate: string }) =>
      updateMilestoneApi(teamId, data.bizKey, {
        milestoneName: data.milestoneName,
        expectedEndDate: data.expectedEndDate,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestones-by-map', teamId, mapId] })
      qc.invalidateQueries({ queryKey: ['milestone-detail', teamId, selectedMilestoneKey] })
      addToast('里程碑更新成功', 'success')
      setMilestoneDialogOpen(false)
      setEditingMilestone(null)
    },
    onError: () => {
      addToast('更新失败，请重试', 'error')
    },
  })

  function handleMilestoneSubmit(data: { milestoneName: string, expectedEndDate: string }) {
    if (milestoneDialogMode === 'create') {
      createMilestoneMutation.mutate(data)
    } else if (editingMilestone) {
      updateMilestoneMutation.mutate({ bizKey: editingMilestone.bizKey, ...data })
    }
  }

  // Compute timeline positioning
  const timelineData = computeTimeline(milestones, zoomConfig)

  return (
    <div>
      {/* Breadcrumb + toolbar */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Breadcrumb>
            <BreadcrumbItem href="#" onClick={(e) => { e.preventDefault(); onBack() }}>
              里程碑图
            </BreadcrumbItem>
            <BreadcrumbSeparator />
            <BreadcrumbItem isCurrent>{mapName}</BreadcrumbItem>
          </Breadcrumb>
        </div>
        <div className="flex items-center gap-2">
          {canCreate && (
            <Button
              size="sm"
              onClick={() => {
                setMilestoneDialogMode('create')
                setEditingMilestone(null)
                setMilestoneDialogOpen(true)
              }}
              data-testid="btn-create-milestone"
            >
              <Plus className="w-4 h-4 mr-1" />
              创建里程碑
            </Button>
          )}
          <Button variant="secondary" size="sm" onClick={onBack}>
            <ChevronLeft className="w-4 h-4 mr-1" />
            返回列表
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => refetch()}
          >
            <RefreshCw className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Zoom controls */}
      <div className="flex items-center gap-1 mb-4">
        {ZOOM_LEVELS.map((z) => (
          <Toggle
            key={z.key}
            pressed={zoom === z.key}
            onPressedChange={() => setZoom(z.key)}
            size="sm"
            variant="outline"
            data-testid={`zoom-${z.key}`}
          >
            {z.label}
          </Toggle>
        ))}
      </div>

      {/* Timeline content */}
      {isLoading && (
        <div data-testid="timeline-loading" className="flex items-center justify-center py-20">
          <div className="flex flex-col items-center gap-2">
            <RefreshCw className="w-6 h-6 animate-spin text-tertiary" />
            <span className="text-[13px] text-tertiary">加载里程碑数据中</span>
          </div>
        </div>
      )}

      {isError && (
        <div className="flex flex-col items-center justify-center py-20">
          <AlertCircle className="w-10 h-10 text-error mb-2" />
          <p className="text-[13px] text-secondary mb-3">加载失败，请重试</p>
          <Button variant="secondary" size="sm" onClick={() => refetch()}>
            重试
          </Button>
        </div>
      )}

      {!isLoading && !isError && milestones.length === 0 && (
        <div className="flex flex-col items-center justify-center py-20">
          <MapPin className="w-10 h-10 text-tertiary mb-2" />
          <p className="text-[14px] text-secondary mb-3">暂无里程碑</p>
        </div>
      )}

      {!isLoading && !isError && milestones.length > 0 && (
        <div
          className="relative overflow-x-auto min-h-[400px] border border-border rounded-xl bg-white p-6"
          role="application"
          aria-label="里程碑时间线"
          data-testid="timeline-view"
        >
          {/* Time axis */}
          {timelineData.ticks.map((tick, i) => (
            <div
              key={i}
              className="absolute top-0 bottom-0 border-l border-border/40"
              style={{ left: tick.x }}
            >
              <span className="text-[11px] text-tertiary ml-1 whitespace-nowrap" data-testid="axis-label">
                {tick.label}
              </span>
            </div>
          ))}

          {/* Milestone nodes */}
          <div className="relative pt-8" style={{ minHeight: 400 }}>
            {timelineData.nodes.map((node, i) => (
              <div
                key={node.milestone.bizKey}
                className="absolute w-40"
                style={{ left: node.x, top: 8 }}
                data-testid="milestone-node"
                role="button"
                tabIndex={0}
                aria-label={`${node.milestone.milestoneName}，${node.milestone.statusName}，完成度 ${Math.round(node.milestone.completion)}%，${node.milestone.relatedMICount} 个事项`}
                onClick={() => setSelectedMilestoneKey(node.milestone.bizKey)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') setSelectedMilestoneKey(node.milestone.bizKey)
                }}
              >
                <div className={`border rounded-xl p-3 bg-white hover:bg-bg-alt transition-colors cursor-pointer ${
                  selectedMilestoneKey === node.milestone.bizKey ? 'border-primary ring-2 ring-blue-200' : 'border-border'
                }`}>
                  <div className="flex items-center gap-1.5 mb-1">
                    <span
                      className={`w-2 h-2 rounded-full shrink-0 ${
                        node.milestone.milestoneStatus === 'completed'
                          ? 'bg-primary-700'
                          : node.milestone.milestoneStatus === 'in_progress'
                          ? 'bg-primary'
                          : node.milestone.milestoneStatus === 'cancelled'
                          ? 'bg-border'
                          : 'bg-secondary'
                      }`}
                    />
                    <span className="text-[13px] font-medium text-primary truncate">
                      {node.milestone.milestoneName}
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-[12px] text-tertiary mb-1">
                    <span>{node.milestone.expectedEndDate}</span>
                    <span>{Math.round(node.milestone.completion)}%</span>
                  </div>
                  <ProgressBar value={node.milestone.completion} size="sm" />
                  <div className="text-[12px] text-tertiary mt-1">
                    {node.milestone.relatedMICount} 个事项
                  </div>
                </div>

                {/* Arrow to next node */}
                {i < timelineData.nodes.length - 1 && (
                  <div className="absolute top-1/2 -right-3 flex items-center">
                    <ArrowRight className="w-3 h-3 text-tertiary" />
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Milestone detail panel */}
      {selectedMilestoneKey && (
        <MilestoneDetailPanel
          milestoneBizKey={selectedMilestoneKey}
          onClose={() => setSelectedMilestoneKey(null)}
          onDeleted={() => setSelectedMilestoneKey(null)}
        />
      )}

      {/* Create/edit milestone dialog */}
      <CreateEditMilestoneDialog
        open={milestoneDialogOpen}
        onOpenChange={setMilestoneDialogOpen}
        mode={milestoneDialogMode}
        milestone={editingMilestone}
        onSubmit={handleMilestoneSubmit}
        isPending={createMilestoneMutation.isPending || updateMilestoneMutation.isPending}
      />
    </div>
  )
}

// --- Timeline computation ---

interface TimelineNode {
  milestone: Milestone
  x: number
}

interface TimelineTick {
  x: number
  label: string
}

function computeTimeline(
  milestones: Milestone[],
  zoomConfig: (typeof ZOOM_LEVELS)[number],
): { nodes: TimelineNode[], ticks: TimelineTick[] } {
  if (milestones.length === 0) return { nodes: [], ticks: [] }

  const sorted = [...milestones].sort(
    (a, b) =>
      new Date(a.expectedEndDate).getTime() -
      new Date(b.expectedEndDate).getTime(),
  )

  const originDate = new Date(sorted[0].expectedEndDate)
  originDate.setDate(originDate.getDate() - 14)

  const nodes: TimelineNode[] = []
  let lastX = 0

  for (const ms of sorted) {
    const msDate = new Date(ms.expectedEndDate)
    const daysDiff =
      (msDate.getTime() - originDate.getTime()) / (1000 * 60 * 60 * 24)
    let x = daysDiff * zoomConfig.pxPerDay

    // Minimum spacing: 40px
    if (x - lastX < 40) {
      x = lastX + 40
    }

    nodes.push({ milestone: ms, x })
    lastX = x
  }

  // Generate ticks
  const ticks: TimelineTick[] = []
  const endDate = new Date(sorted[sorted.length - 1].expectedEndDate)
  endDate.setDate(endDate.getDate() + 14)

  const tickDate = new Date(originDate)
  while (tickDate <= endDate) {
    const daysDiff =
      (tickDate.getTime() - originDate.getTime()) / (1000 * 60 * 60 * 24)
    const x = daysDiff * zoomConfig.pxPerDay
    const label = tickDate.toISOString().slice(0, 7)
    ticks.push({ x, label })
    tickDate.setDate(tickDate.getDate() + zoomConfig.tickDays)
  }

  return { nodes, ticks }
}

// --- Main Component ---

export default function MilestonesPage() {
  const teamId = useTeamStore((s) => s.currentTeamId) ?? ''
  const canCreate = usePermission('milestone:create')

  // View state
  const [selectedMapId, setSelectedMapId] = useState<string | null>(null)
  const [selectedMapName, setSelectedMapName] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [dialogOpen, setDialogOpen] = useState(false)

  // Fetch milestone maps
  const {
    data: mapsData,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['milestone-maps', teamId, statusFilter],
    queryFn: () =>
      listMilestoneMapsApi(teamId, {
        status: statusFilter || undefined,
      }),
    enabled: !!teamId,
  })

  const maps = mapsData?.items ?? []

  // Handlers
  const handleCardClick = useCallback(
    (map: MilestoneMap) => {
      setSelectedMapId(map.bizKey)
      setSelectedMapName(map.mapName)
    },
    [],
  )

  const handleBackToList = useCallback(() => {
    setSelectedMapId(null)
    setSelectedMapName('')
  }, [])

  // --- Render ---

  // Timeline view (second level)
  if (selectedMapId) {
    return (
      <div className="p-6">
        <TimelineView
          mapId={selectedMapId}
          mapName={selectedMapName}
          onBack={handleBackToList}
        />
      </div>
    )
  }

  // List view (first level)
  return (
    <div data-testid="milestones-page" className="p-6">
      {/* Page header */}
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-lg font-semibold text-primary">里程碑图</h1>
        <div className="flex items-center gap-2">
          <Button
            onClick={() => setDialogOpen(true)}
            disabled={!canCreate}
            data-testid="btn-create-map"
          >
            <Plus className="w-4 h-4 mr-1" />
            创建里程碑图
          </Button>
          <Button
            variant="secondary"
            onClick={() => refetch()}
            aria-label="刷新"
          >
            <RefreshCw className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Status filter */}
      <div className="mb-4" data-testid="filter-status">
        <Select
          value={statusFilter || '_all'}
          onValueChange={(val) => setStatusFilter(val === '_all' ? '' : val)}
        >
          <SelectTrigger
            data-testid="status-filter-trigger"
            className="w-[160px]"
          >
            <SelectValue placeholder="全部" />
          </SelectTrigger>
          <SelectContent>
            {MAP_STATUSES.map((s) => (
              <SelectItem key={s.code || '_all'} value={s.code || '_all'}>
                {s.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Content */}
      {isLoading && <SkeletonCards />}

      {isError && (
        <div className="flex flex-col items-center justify-center py-20" data-testid="error-state">
          <AlertCircle className="w-10 h-10 text-error mb-2" />
          <p className="text-[14px] text-secondary mb-3">加载失败，请重试</p>
          <Button variant="secondary" size="sm" onClick={() => refetch()} data-testid="btn-retry">
            重试
          </Button>
        </div>
      )}

      {!isLoading && !isError && maps.length === 0 && (
        <div className="flex flex-col items-center justify-center py-20" data-testid="empty-state">
          <MapPin className="w-10 h-10 text-tertiary mb-2" />
          <p className="text-[14px] text-secondary mb-3">暂无里程碑图</p>
          {canCreate && (
            <Button onClick={() => setDialogOpen(true)}>
              <Plus className="w-4 h-4 mr-1" />
              创建里程碑图
            </Button>
          )}
        </div>
      )}

      {!isLoading && !isError && maps.length > 0 && (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(340px,1fr))] gap-4">
          {maps.map((map) => (
            <MapCard
              key={map.bizKey}
              map={map}
              onClick={() => handleCardClick(map)}
            />
          ))}
        </div>
      )}

      {/* Create map dialog */}
      <CreateMapDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        teamId={teamId}
      />
    </div>
  )
}
