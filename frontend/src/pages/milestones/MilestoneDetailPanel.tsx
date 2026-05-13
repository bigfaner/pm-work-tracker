import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { X, Plus, Trash2, Pencil } from 'lucide-react'
import { useTeamStore } from '@/store/team'
import { usePermission } from '@/hooks/usePermission'
import {
  getMilestoneApi,
  getMilestoneTransitionsApi,
  changeMilestoneStatusApi,
  deleteMilestoneApi,
} from '@/api/milestones'
import { updateMainItemApi } from '@/api/mainItems'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import ProgressBar from '@/components/shared/ProgressBar'
import CreateEditMilestoneDialog from './CreateEditMilestoneDialog'
import QuickAddMIDialog from './QuickAddMIDialog'
import type { MainItem } from '@/types'
import { getStatusName } from '@/lib/status'
import { useToast } from '@/components/ui/toast'

interface MilestoneDetailPanelProps {
  milestoneBizKey: string
  onClose: () => void
  onDeleted: () => void
}

export default function MilestoneDetailPanel({
  milestoneBizKey,
  onClose,
  onDeleted,
}: MilestoneDetailPanelProps) {
  const teamId = useTeamStore((s) => s.currentTeamId) ?? ''
  const canUpdate = usePermission('milestone:update')
  const canDelete = usePermission('milestone:delete')
  const qc = useQueryClient()
  const { addToast } = useToast()

  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [quickAddOpen, setQuickAddOpen] = useState(false)
  const [statusDropdownOpen, setStatusDropdownOpen] = useState(false)
  const [showNoTransitions] = useState(false)

  // Fetch milestone detail
  const {
    data: milestone,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['milestone-detail', teamId, milestoneBizKey],
    queryFn: () => getMilestoneApi(teamId, milestoneBizKey),
    enabled: !!teamId && !!milestoneBizKey,
  })

  // Fetch related MIs (use listMainItemsApi with milestoneKey filter)
  const { data: relatedMIs } = useQuery({
    queryKey: ['milestone-related-mis', teamId, milestoneBizKey],
    queryFn: async () => {
      // Import dynamically to avoid circular deps
      const { listMainItemsApi } = await import('@/api/mainItems')
      const result = await listMainItemsApi(teamId, {
        page: 1,
        pageSize: 100,
      })
      // Filter MIs that belong to this milestone
      return result.items.filter(
        (mi: MainItem) => mi.milestoneKey === milestoneBizKey,
      )
    },
    enabled: !!teamId && !!milestoneBizKey && milestone?.milestoneStatus !== 'cancelled',
  })

  // Status transitions
  const { data: transitions = [] } = useQuery({
    queryKey: ['milestone-transitions', teamId, milestoneBizKey],
    queryFn: () => getMilestoneTransitionsApi(teamId, milestoneBizKey),
    enabled: !!teamId && !!milestoneBizKey && statusDropdownOpen,
  })

  // Change status mutation
  const statusMutation = useMutation({
    mutationFn: (newStatus: string) =>
      changeMilestoneStatusApi(teamId, milestoneBizKey, { status: newStatus }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestone-detail', teamId, milestoneBizKey] })
      qc.invalidateQueries({ queryKey: ['milestones-by-map', teamId] })
      qc.invalidateQueries({ queryKey: ['milestone-related-mis', teamId, milestoneBizKey] })
      setStatusDropdownOpen(false)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: () => deleteMilestoneApi(teamId, milestoneBizKey),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestones-by-map', teamId] })
      onDeleted()
    },
  })

  // Unbind MI mutation
  const unbindMutation = useMutation({
    mutationFn: (miBizKey: string) =>
      updateMainItemApi(teamId, miBizKey, { milestoneKey: null }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestone-related-mis', teamId, milestoneBizKey] })
      qc.invalidateQueries({ queryKey: ['milestone-detail', teamId, milestoneBizKey] })
      addToast('已解绑事项', 'success')
    },
    onError: () => {
      addToast('解绑失败，请重试', 'error')
    },
  })

  const isCancelled = milestone?.milestoneStatus === 'cancelled'
  const mis = relatedMIs ?? []

  const handleStatusSelect = useCallback(
    (status: string) => {
      statusMutation.mutate(status)
    },
    [statusMutation],
  )

  // Loading state
  if (isLoading) {
    return (
      <div
        className="fixed right-0 top-0 h-full w-[360px] bg-white shadow-[0_10px_15px_-3px_rgb(0_0_0/0.1),0_4px_6px_-4px_rgb(0_0_0/0.1)] z-40 flex flex-col"
        role="dialog"
        aria-label="里程碑详情"
        aria-modal="true"
        data-testid="milestone-detail-panel"
      >
        <div className="flex items-center justify-between p-4 border-b border-border">
          <div className="h-6 w-32 bg-bg-alt animate-pulse rounded" />
          <button type="button" onClick={onClose} className="text-tertiary hover:text-primary">
            <X className="w-5 h-5" />
          </button>
        </div>
        <div className="p-4 space-y-4">
          <div className="h-4 w-24 bg-bg-alt animate-pulse rounded" />
          <div className="h-8 w-full bg-bg-alt animate-pulse rounded" />
          <div className="h-4 w-20 bg-bg-alt animate-pulse rounded" />
          <div className="h-2 w-full bg-bg-alt animate-pulse rounded" />
          <div className="space-y-2">
            <div className="h-8 bg-bg-alt animate-pulse rounded" />
            <div className="h-8 bg-bg-alt animate-pulse rounded" />
            <div className="h-8 bg-bg-alt animate-pulse rounded" />
          </div>
        </div>
      </div>
    )
  }

  if (isError || !milestone) {
    return (
      <div
        className="fixed right-0 top-0 h-full w-[360px] bg-white shadow-[0_10px_15px_-3px_rgb(0_0_0/0.1),0_4px_6px_-4px_rgb(0_0_0/0.1)] z-40 flex flex-col"
        role="dialog"
        aria-label="里程碑详情"
        aria-modal="true"
      >
        <div className="flex items-center justify-between p-4 border-b border-border">
          <span className="text-[15px] font-medium text-primary">里程碑详情</span>
          <button type="button" onClick={onClose} className="text-tertiary hover:text-primary">
            <X className="w-5 h-5" />
          </button>
        </div>
        <div className="flex-1 flex flex-col items-center justify-center gap-3">
          <p className="text-[13px] text-secondary">加载失败</p>
          <Button variant="secondary" size="sm" onClick={() => refetch()}>
            重试
          </Button>
        </div>
      </div>
    )
  }

  return (
    <>
      <div
        className={`fixed right-0 top-0 h-full w-[360px] bg-white shadow-[0_10px_15px_-3px_rgb(0_0_0/0.1),0_4px_6px_-4px_rgb(0_0_0/0.1)] z-40 flex flex-col transition-transform duration-300 ease-out ${isCancelled ? 'opacity-70' : ''}`}
        role="dialog"
        aria-label="里程碑详情"
        aria-modal="true"
        data-testid="milestone-detail-panel"
      >
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <span className="text-[15px] font-medium text-primary truncate">
            {milestone.milestoneName}
          </span>
          <button
            type="button"
            onClick={onClose}
            className="text-tertiary hover:text-primary"
            aria-label="关闭"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {/* Meta row: date + status + edit */}
          <div className="flex items-start justify-between">
            <div>
              <span className="block text-[12px] text-tertiary">计划完成时间</span>
              <span className="text-[13px] text-primary">{milestone.expectedEndDate}</span>
            </div>
            <div className="flex items-center gap-2">
              {/* Status badge dropdown */}
              <div className="relative inline-flex">
                {showNoTransitions && (
                  <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1.5 whitespace-nowrap text-xs px-2 py-1 rounded-md bg-primary text-white shadow-md pointer-events-none z-50">
                    暂无可用流转
                  </div>
                )}
                <DropdownMenu
                  open={statusDropdownOpen}
                  onOpenChange={(open) => {
                    if (!canUpdate || isCancelled) return
                    setStatusDropdownOpen(open)
                  }}
                >
                  <DropdownMenuTrigger asChild>
                    <button
                      className="focus:outline-none"
                      disabled={!canUpdate || isCancelled}
                    >
                      <Badge
                        variant={
                          `status-${
                            milestone.milestoneStatus === 'not_started'
                              ? 'planning'
                              : milestone.milestoneStatus === 'in_progress'
                                ? 'in-progress'
                                : milestone.milestoneStatus === 'completed'
                                  ? 'completed'
                                  : 'cancelled'
                          }` as 'status-planning' | 'status-in-progress' | 'status-completed' | 'status-cancelled'
                        }
                        className={`text-[11px] ${canUpdate && !isCancelled ? 'cursor-pointer' : ''}`}
                      >
                        {milestone.statusName}
                      </Badge>
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent className="min-w-0 w-auto">
                    {transitions.map((status) => (
                      <DropdownMenuItem
                        key={status}
                        className="text-[13px] justify-center"
                        onSelect={(e) => {
                          e.preventDefault()
                          handleStatusSelect(status)
                        }}
                      >
                        {getStatusName(status) || status}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>

              {/* Edit button */}
              {canUpdate && !isCancelled && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setEditDialogOpen(true)}
                  aria-label="编辑里程碑"
                >
                  <Pencil className="w-4 h-4" />
                </Button>
              )}
            </div>
          </div>

          {/* Completion */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <span className="text-[13px] text-secondary">完成度</span>
              <span className="text-[13px] text-secondary">{Math.round(milestone.completion)}%</span>
            </div>
            <ProgressBar value={milestone.completion} size="sm" />
          </div>

          {/* Related MIs */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-[13px] font-medium text-primary">
                关联事项 ({milestone.relatedMICount})
              </span>
              {canUpdate && !isCancelled && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setQuickAddOpen(true)}
                >
                  <Plus className="w-3 h-3 mr-1" />
                  添加
                </Button>
              )}
            </div>

            {isCancelled ? (
              <p className="text-[12px] text-tertiary">里程碑已取消</p>
            ) : mis.length === 0 ? (
              <p className="text-[12px] text-tertiary">暂无关联事项</p>
            ) : (
              <div className="space-y-1">
                {mis.map((mi: MainItem) => (
                  <div
                    key={mi.bizKey}
                    className="group flex items-center justify-between px-2 py-1.5 rounded-md hover:bg-bg-alt text-[13px]"
                  >
                    <span className="text-primary truncate">
                      {mi.code} {mi.title}
                    </span>
                    {canUpdate && !isCancelled && (
                      <button
                        type="button"
                        className="opacity-0 group-hover:opacity-100 text-tertiary hover:text-error-text transition-opacity shrink-0 ml-2"
                        onClick={() => unbindMutation.mutate(mi.bizKey)}
                        aria-label={`解绑 ${mi.code}`}
                        disabled={unbindMutation.isPending}
                      >
                        <X className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Danger zone */}
          {canDelete && !isCancelled && (
            <div className="pt-4 border-t border-border">
              <Button
                variant="danger"
                size="sm"
                onClick={() => setDeleteConfirmOpen(true)}
              >
                <Trash2 className="w-4 h-4 mr-1" />
                删除里程碑
              </Button>
            </div>
          )}
        </div>
      </div>

      {/* Edit dialog */}
      <CreateEditMilestoneDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        mode="edit"
        milestone={milestone}
        isPending={false}
        onSubmit={() => {
          // Handled by parent mutation
        }}
      />

      {/* Quick add MI dialog */}
      <QuickAddMIDialog
        open={quickAddOpen}
        onOpenChange={setQuickAddOpen}
        milestoneBizKey={milestoneBizKey}
        milestoneName={milestone.milestoneName}
        onCreated={() => {
          qc.invalidateQueries({ queryKey: ['milestone-related-mis', teamId, milestoneBizKey] })
          qc.invalidateQueries({ queryKey: ['milestone-detail', teamId, milestoneBizKey] })
        }}
      />

      {/* Delete confirmation dialog */}
      <Dialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>
              确定删除里程碑「{milestone.milestoneName}」？
            </DialogTitle>
          </DialogHeader>
          <DialogBody>
            <p className="text-[13px] text-secondary">
              关联的 {milestone.relatedMICount} 个事项将解除绑定，里程碑数据不可恢复。
            </p>
          </DialogBody>
          <DialogFooter>
            <Button
              variant="secondary"
              onClick={() => setDeleteConfirmOpen(false)}
              disabled={deleteMutation.isPending}
            >
              取消
            </Button>
            <Button
              variant="danger"
              onClick={() => deleteMutation.mutate()}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? '删除中...' : '确认删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
