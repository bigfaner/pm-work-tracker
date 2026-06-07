import { useState, useEffect, useRef, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { X, Pencil, Plus, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Button } from '@/components/ui/button'
import StatusBadge from '@/components/shared/StatusBadge'
import StatusTransitionDropdown from '@/components/shared/StatusTransitionDropdown'
import ProgressBar from '@/components/shared/ProgressBar'
import ConfirmDialog from '@/components/shared/ConfirmDialog'
import { useToast } from '@/components/ui/toast'
import { useTeamStore } from '@/store/team'
import { getMilestoneApi, deleteMilestoneApi } from '@/api/milestones'
import { listMainItemsApi, updateMainItemApi } from '@/api/mainItems'
import { MILESTONE_STATUSES } from '@/lib/status'
import type { Milestone, MainItem } from '@/types'

/** Related MI displayed in the panel */
interface RelatedMI {
  bizKey: string;
  code: string;
  title: string;
  completion: number;
  itemStatus: string;
}

interface MilestoneDetailPanelProps {
  open: boolean;
  onClose: () => void;
  milestoneId: string | null;
  /** Called to open the edit milestone dialog */
  onEdit: (milestone: Milestone) => void;
  /** Called to open the quick-add main item dialog */
  onQuickAdd: (milestone: Milestone) => void;
  /** Called after the milestone is deleted */
  onDeleted?: () => void;
  /** Trigger element ref for returning focus on close */
  triggerRef?: React.RefObject<HTMLElement | null>;
}

/** Statuses that allow deletion per BR-4 */
const DELETABLE_STATUSES = ['not_started', 'cancelled']

/** Terminal milestone statuses */
const MILESTONE_TERMINAL_STATUSES = Object.entries(MILESTONE_STATUSES)
  .filter(([, v]) => v.terminal)
  .map(([k]) => k)

/** Main item terminal statuses */
const MAIN_TERMINAL_STATUSES = ['completed', 'closed']

export default function MilestoneDetailPanel({
  open,
  onClose,
  milestoneId,
  onEdit,
  onQuickAdd,
  onDeleted,
  triggerRef,
}: MilestoneDetailPanelProps) {
  const teamId = useTeamStore((s) => s.currentTeamId) ?? ''
  const qc = useQueryClient()
  const { addToast } = useToast()
  const panelRef = useRef<HTMLDivElement>(null)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)

  // Fetch milestone data
  const { data: milestone, isLoading } = useQuery({
    queryKey: ['milestone', teamId, milestoneId],
    queryFn: () => getMilestoneApi(teamId, milestoneId!),
    enabled: open && !!teamId && !!milestoneId,
  })

  // Fetch related MIs (filtered by milestoneKey)
  const { data: relatedMIsResult } = useQuery({
    queryKey: ['milestoneMIs', teamId, milestoneId],
    queryFn: () =>
      listMainItemsApi(teamId, { milestoneKey: milestoneId!, pageSize: 200 }),
    enabled: open && !!teamId && !!milestoneId,
  })

  const relatedMIs: RelatedMI[] = (relatedMIsResult?.items ?? []).map(
    (item: MainItem) => ({
      bizKey: item.bizKey,
      code: item.code,
      title: item.title,
      completion: item.completion,
      itemStatus: item.itemStatus,
    }),
  )

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: () => deleteMilestoneApi(teamId, milestoneId!),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestones', teamId] })
      qc.invalidateQueries({ queryKey: ['milestoneMap', teamId] })
      setDeleteConfirmOpen(false)
      onDeleted?.()
      onClose()
    },
  })

  // Unbind mutation
  const unbindMutation = useMutation({
    mutationFn: (miBizKey: string) =>
      updateMainItemApi(teamId, miBizKey, { milestoneKey: '' }),
    onSuccess: (_data, miBizKey) => {
      qc.invalidateQueries({ queryKey: ['milestoneMIs', teamId, milestoneId] })
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      // Find the unbound MI for undo toast
      const unboundMI = relatedMIs.find((mi) => mi.bizKey === miBizKey)
      addToast(`已解除事项 ${unboundMI?.code ?? ''} 的绑定`, 'success')
    },
  })

  // Escape key handler
  useEffect(() => {
    if (!open) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault()
        onClose()
      }
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [open, onClose])

  // Return focus to trigger element on close
  useEffect(() => {
    if (!open && triggerRef?.current) {
      triggerRef.current.focus()
    }
  }, [open, triggerRef])

  const isCancelled = milestone?.milestoneStatus === 'cancelled'
  const isTerminal = milestone
    ? MILESTONE_TERMINAL_STATUSES.includes(milestone.milestoneStatus)
    : false
  const canDelete = milestone
    ? DELETABLE_STATUSES.includes(milestone.milestoneStatus)
    : false

  const handleStatusChanged = useCallback(() => {
    qc.invalidateQueries({ queryKey: ['milestone', teamId, milestoneId] })
    qc.invalidateQueries({ queryKey: ['milestones', teamId] })
  }, [qc, teamId, milestoneId])

  const handleEditClick = useCallback(() => {
    if (milestone) onEdit(milestone)
  }, [milestone, onEdit])

  const handleQuickAddClick = useCallback(() => {
    if (milestone) onQuickAdd(milestone)
  }, [milestone, onQuickAdd])

  // Don't render anything when closed (no overlay)
  if (!open) return null

  return (
    <>
      {/* Overlay */}
      <div
        className="fixed inset-0 z-40 bg-black/20"
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Panel */}
      <div
        ref={panelRef}
        role="dialog"
        aria-label={`里程碑详情: ${milestone?.milestoneName ?? ''}`}
        className="fixed right-0 top-0 z-40 h-full w-[360px] bg-white shadow-lg transition-transform duration-300 ease-out translate-x-0 flex flex-col"
      >
        {/* Close button */}
        <div className="flex items-center justify-between px-5 pt-4 pb-2">
          <button
            onClick={onClose}
            className="text-tertiary hover:text-primary transition-colors"
            aria-label="关闭面板"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-5 pb-6">
          {isLoading ? (
            <div className="space-y-4 animate-pulse">
              <div className="h-6 bg-bg-alt rounded w-2/3" />
              <div className="h-4 bg-bg-alt rounded w-1/2" />
              <div className="h-4 bg-bg-alt rounded w-full" />
              <div className="h-4 bg-bg-alt rounded w-3/4" />
            </div>
          ) : milestone ? (
            <>
              {/* Name */}
              <h2
                className={`text-lg font-semibold mb-4 ${
                  isCancelled ? 'text-tertiary line-through' : 'text-primary'
                }`}
              >
                {milestone.milestoneName}
              </h2>

              {/* Description area: row 1 = label + status + edit; row 2 = text */}
              <div className="mb-4">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs text-tertiary">描述</span>
                  <div className="flex items-center gap-2 flex-shrink-0">
                    <StatusTransitionDropdown
                      currentStatus={milestone.milestoneStatus}
                      itemType="main"
                      teamId={teamId}
                      itemId={milestoneId!}
                      onStatusChanged={handleStatusChanged}
                      disabled={isTerminal}
                    />
                    {!isTerminal && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={handleEditClick}
                        aria-label="编辑里程碑"
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                </div>
                {milestone.milestoneDesc ? (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <p
                        className={`text-sm ${
                          isCancelled ? 'text-tertiary' : 'text-secondary'
                        }`}
                        style={{
                          display: '-webkit-box',
                          WebkitLineClamp: 6,
                          WebkitBoxOrient: 'vertical',
                          overflow: 'hidden',
                        }}
                      >
                        {milestone.milestoneDesc}
                      </p>
                    </TooltipTrigger>
                    <TooltipContent
                      side="left"
                      className="max-w-[280px] whitespace-pre-wrap"
                    >
                      {milestone.milestoneDesc}
                    </TooltipContent>
                  </Tooltip>
                ) : (
                  <p className="text-sm text-tertiary">暂无描述</p>
                )}
              </div>

              {/* Expected end date */}
              <div className="mb-4">
                <span className="text-xs text-tertiary block mb-1">
                  计划完成时间
                </span>
                <span
                  className={`text-sm ${isCancelled ? 'text-tertiary' : 'text-primary'}`}
                >
                  {milestone.expectedEndDate ?? '未设置'}
                </span>
              </div>

              {/* Progress */}
              <div className="mb-4">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs text-tertiary">进度</span>
                  <span className="text-xs text-tertiary">
                    {Math.round(milestone.completion)}%
                  </span>
                </div>
                <ProgressBar value={milestone.completion} size="sm" />
              </div>

              {/* Related MIs */}
              {!isCancelled && (
                <div className="mb-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs text-tertiary">
                      关联事项 ({relatedMIs.length})
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={handleQuickAddClick}
                      className="text-xs text-primary-700 hover:text-primary"
                    >
                      <Plus className="h-3.5 w-3.5 mr-1" />
                      添加
                    </Button>
                  </div>

                  {relatedMIs.length === 0 ? (
                    <p className="text-xs text-tertiary py-2">暂无关联事项</p>
                  ) : (
                    <div className="space-y-1">
                      {relatedMIs.map((mi) => {
                        const isMITerminal = MAIN_TERMINAL_STATUSES.includes(
                          mi.itemStatus,
                        )
                        return (
                          <div
                            key={mi.bizKey}
                            draggable={!isMITerminal}
                            onDragStart={() => {
                              ;(window as unknown as Record<string, unknown>).__dragMI = {
                                miBizKey: mi.bizKey,
                                miCode: mi.code,
                                sourceMilestoneKey: milestoneId!,
                              }
                            }}
                            onDragEnd={() => {
                              // Clean up after drag ends (whether drop or cancel)
                              setTimeout(() => {
                                if (
                                  (window as unknown as Record<string, unknown>).__dragMI
                                ) {
                                  delete (window as unknown as Record<string, unknown>).__dragMI
                                }
                              }, 200)
                            }}
                            className={cn(
                              'group flex items-center gap-2 rounded-md px-2 py-1.5 hover:bg-bg-alt text-sm',
                              !isMITerminal
                                ? 'cursor-grab active:cursor-grabbing active:opacity-50'
                                : 'cursor-pointer',
                            )}
                            data-testid={`mi-drag-${mi.bizKey}`}
                          >
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center gap-1.5">
                                <span className="text-tertiary text-xs font-mono">
                                  {mi.code}
                                </span>
                                <span
                                  className={`truncate ${
                                    isCancelled
                                      ? 'text-tertiary'
                                      : 'text-secondary'
                                  }`}
                                >
                                  {mi.title}
                                </span>
                              </div>
                              <div className="flex items-center gap-2 mt-0.5">
                                <span className="text-xs text-tertiary">
                                  {mi.completion}%
                                </span>
                                <StatusBadge
                                  status={mi.itemStatus}
                                  className="text-[11px]"
                                />
                              </div>
                            </div>
                            {!isMITerminal && (
                              <button
                                onClick={(e) => {
                                  e.stopPropagation()
                                  unbindMutation.mutate(mi.bizKey)
                                }}
                                className="opacity-0 group-hover:opacity-100 text-tertiary hover:text-error transition-opacity shrink-0"
                                aria-label={`解绑事项 ${mi.code}`}
                              >
                                <X className="h-4 w-4" />
                              </button>
                            )}
                          </div>
                        )
                      })}
                    </div>
                  )}
                </div>
              )}

              {/* Danger zone: delete */}
              {canDelete && (
                <div className="pt-4 border-t border-border">
                  <span className="text-xs text-tertiary block mb-2">
                    危险操作
                  </span>
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => setDeleteConfirmOpen(true)}
                  >
                    <Trash2 className="h-4 w-4 mr-1" />
                    删除里程碑
                  </Button>
                </div>
              )}
            </>
          ) : null}
        </div>
      </div>

      {/* Delete confirmation dialog */}
      <ConfirmDialog
        open={deleteConfirmOpen}
        onOpenChange={setDeleteConfirmOpen}
        title={`确定删除里程碑 ${milestone?.milestoneName ?? ''}？`}
        description={`关联的 ${relatedMIs.length} 个事项将解除绑定，里程碑数据不可恢复。`}
        confirmLabel="确认删除"
        confirmVariant="danger"
        onConfirm={() => deleteMutation.mutate()}
      />
    </>
  )
}
