import { useState, useCallback, useMemo, useRef, useEffect } from 'react'
import { Plus, RefreshCw } from 'lucide-react'
import { useToast } from '@/components/ui/toast'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { listMilestoneMapsApi, createMilestoneMapApi } from '@/api/milestones'
import { listMembersApi } from '@/api/teams'
import type { MilestoneMap } from '@/types'
import { MILESTONE_MAP_STATUSES } from '@/lib/status'
import { usePermission } from '@/hooks/usePermission'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { StatusTagFilter } from '@/components/shared/StatusTagFilter'
import MilestoneMapCard from './MilestoneMapCard'
import CreateMilestoneMapDialog, {
  type MilestoneMapFormState,
} from './CreateMilestoneMapDialog'

// --- Constants ---

const DEBOUNCE_MS = 300
const SKELETON_COUNT = 3

const MILESTONE_MAP_STATUS_OPTIONS = Object.entries(MILESTONE_MAP_STATUSES).map(
  ([value, def]) => ({ value, label: def.name }),
)

const EMPTY_FORM: MilestoneMapFormState = {
  mapName: '',
  assigneeKey: '',
  planStartDate: '',
  expectedEndDate: '',
  mapDesc: '',
}

// --- Skeleton Card ---

function SkeletonCard() {
  return (
    <div className="rounded-xl border border-border bg-white p-4 animate-pulse">
      <div className="flex items-center justify-between mb-2">
        <div className="h-4 w-32 bg-bg-alt rounded" />
        <div className="h-5 w-12 bg-bg-alt rounded" />
      </div>
      <div className="flex items-center justify-between mb-2">
        <div className="h-3 w-28 bg-bg-alt rounded" />
        <div className="h-3 w-10 bg-bg-alt rounded" />
      </div>
      <div className="flex items-center justify-between mb-2">
        <div className="h-3 w-20 bg-bg-alt rounded" />
        <div className="h-3 w-24 bg-bg-alt rounded" />
      </div>
      <div className="h-2 w-full bg-bg-alt rounded" />
    </div>
  )
}

// --- Main Component ---

export default function MilestoneMapList() {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const qc = useQueryClient()
  const { addToast } = useToast()

  // Permission
  const canCreate = usePermission('milestone:create')

  // Filter state
  const [searchText, setSearchText] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [assigneeFilter, setAssigneeFilter] = useState<string>('')
  const [statusFilter, setStatusFilter] = useState<string[]>([])

  // Dialog state
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [createForm, setCreateForm] =
    useState<MilestoneMapFormState>(EMPTY_FORM)

  // Debounce search
  const debounceRef = useRef<ReturnType<typeof setTimeout>>()
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

  // --- Data fetching ---

  const {
    data: mapsData,
    isLoading,
    isError,
    refetch,
    isFetching,
  } = useQuery({
    queryKey: ['milestoneMaps', teamId],
    queryFn: () => listMilestoneMapsApi(teamId!),
    enabled: !!teamId,
  })

  const { data: membersData } = useQuery({
    queryKey: ['members', teamId],
    queryFn: () => listMembersApi(teamId!),
    enabled: !!teamId,
  })

  const members = useMemo(
    () =>
      (membersData || []).map(
        (m: { userKey: string, displayName: string }) => ({
          userKey: m.userKey,
          displayName: m.displayName,
        }),
      ),
    [membersData],
  )

  const allMaps: MilestoneMap[] = mapsData?.items ?? []

  // --- Client-side filtering ---

  const filteredMaps = useMemo(() => {
    let items = allMaps
    if (debouncedSearch.trim()) {
      const q = debouncedSearch.trim().toLowerCase()
      items = items.filter((m) => m.mapName.toLowerCase().includes(q))
    }
    if (assigneeFilter) {
      items = items.filter((m) => m.assigneeKey === assigneeFilter)
    }
    if (statusFilter.length > 0) {
      items = items.filter((m) => statusFilter.includes(m.mapStatus))
    }
    return items
  }, [allMaps, debouncedSearch, assigneeFilter, statusFilter])

  // --- Reset filters ---

  const resetFilters = useCallback(() => {
    setSearchText('')
    setDebouncedSearch('')
    setAssigneeFilter('')
    setStatusFilter([])
  }, [])

  // --- Create mutation ---

  const createMutation = useMutation({
    mutationFn: (form: MilestoneMapFormState) =>
      createMilestoneMapApi(teamId!, {
        mapName: form.mapName.trim(),
        assigneeBizKey: form.assigneeKey,
        ...(form.planStartDate && { planStartDate: form.planStartDate }),
        ...(form.expectedEndDate && { expectedEndDate: form.expectedEndDate }),
        ...(form.mapDesc && { mapDesc: form.mapDesc }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestoneMaps', teamId] })
      setCreateDialogOpen(false)
      setCreateForm(EMPTY_FORM)
    },
  })

  // --- Render ---

  return (
    <div data-testid="milestone-map-list">
      {/* Page Header */}
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-lg font-semibold text-primary">里程碑图</h1>
        <div className="flex items-center gap-2">
          {canCreate && (
            <Button
              size="sm"
              onClick={() => setCreateDialogOpen(true)}
              data-testid="create-map-btn"
            >
              <Plus className="w-4 h-4" />
              创建里程碑图
            </Button>
          )}
        </div>
      </div>

      {/* Filter Bar */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <Input
          placeholder="搜索名称..."
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          className="w-60"
          data-testid="search-input"
        />
        <Select
          value={assigneeFilter}
          onValueChange={(v) => setAssigneeFilter(v === '_all' ? '' : v)}
        >
          <SelectTrigger className="w-35" data-testid="assignee-filter">
            <SelectValue placeholder="负责人：全部" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="_all">负责人：全部</SelectItem>
            {members.map((m) => (
              <SelectItem key={m.userKey} value={m.userKey}>
                {m.displayName}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <StatusTagFilter
          options={MILESTONE_MAP_STATUS_OPTIONS}
          selected={statusFilter}
          onChange={setStatusFilter}
          label="状态"
        />
        <Button variant="secondary" size="sm" onClick={resetFilters}>
          重置
        </Button>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => {
            refetch()
            addToast('已刷新里程碑图列表', 'success')
          }}
          disabled={isFetching}
          data-testid="refresh-btn"
        >
          <RefreshCw size={14} className={isFetching ? 'animate-spin' : ''} />
          刷新
        </Button>
      </div>

      {/* Content */}
      {isLoading ? (
        <div
          className="grid grid-cols-[repeat(auto-fill,minmax(340px,1fr))] gap-4"
          data-testid="loading-skeleton"
        >
          {Array.from({ length: SKELETON_COUNT }, (_, i) => (
            <SkeletonCard key={i} />
          ))}
        </div>
      ) : isError ? (
        <div className="py-12 text-center" data-testid="error-state">
          <p className="text-tertiary text-sm">加载失败，请重试</p>
          <Button
            variant="secondary"
            size="sm"
            className="mt-3"
            onClick={() => refetch()}
          >
            重试
          </Button>
        </div>
      ) : allMaps.length === 0 ? (
        <div className="py-12 text-center" data-testid="empty-state">
          <p className="text-tertiary text-sm">暂无里程碑图</p>
          {canCreate && (
            <Button
              size="sm"
              className="mt-3"
              onClick={() => setCreateDialogOpen(true)}
            >
              <Plus className="w-4 h-4" />
              创建里程碑图
            </Button>
          )}
        </div>
      ) : filteredMaps.length === 0 ? (
        <div className="py-12 text-center">
          <p className="text-tertiary text-sm">没有符合条件的事项</p>
          <Button
            variant="secondary"
            size="sm"
            className="mt-3"
            onClick={resetFilters}
          >
            清除过滤条件
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(340px,1fr))] gap-4">
          {filteredMaps.map((map) => (
            <MilestoneMapCard key={map.bizKey} map={map} />
          ))}
          {/* Dashed create card */}
          {canCreate && (
            <button
              type="button"
              onClick={() => setCreateDialogOpen(true)}
              data-testid="create-map-card"
              className="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-border bg-transparent p-8 text-tertiary transition-colors hover:border-primary-300 hover:text-secondary cursor-pointer min-h-[160px]"
            >
              <Plus className="w-6 h-6 mb-1" />
              <span className="text-sm">创建里程碑图</span>
            </button>
          )}
        </div>
      )}

      {/* Create Dialog */}
      <CreateMilestoneMapDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        form={createForm}
        onFormChange={setCreateForm}
        members={members}
        onSubmit={() => createMutation.mutate(createForm)}
        isPending={createMutation.isPending}
      />
    </div>
  )
}
