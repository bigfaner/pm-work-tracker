import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
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
import { Button } from '@/components/ui/button'
import StatusBadge from '@/components/shared/StatusBadge'
import {
  getStatusName,
  MAIN_TERMINAL_STATUSES,
  SUB_TERMINAL_STATUSES,
  MILESTONE_MAP_STATUSES,
  MILESTONE_STATUSES,
} from '@/lib/status'
import {
  getMainItemTransitionsApi,
  changeMainItemStatusApi,
} from '@/api/mainItems'
import {
  getSubItemTransitionsApi,
  changeSubItemStatusApi,
} from '@/api/subItems'
import {
  getMilestoneMapTransitionsApi,
  changeMilestoneMapStatusApi,
  getMilestoneTransitionsApi,
  changeMilestoneStatusApi,
} from '@/api/milestones'
import { X } from 'lucide-react'
import { isAxiosError } from 'axios'

export interface StatusTransitionDropdownProps {
  currentStatus: string
  itemType: 'main' | 'sub' | 'milestone-map' | 'milestone'
  teamId: string
  itemId: string
  onStatusChanged: () => void
  /** For sub-items: the parent main item's bizKey, needed to invalidate the correct query */
  parentItemId?: string
  /** Called before terminal status transition. Return true to proceed, false to cancel. */
  onBeforeTerminalStatus?: (status: string) => Promise<boolean>
  disabled?: boolean
}

export default function StatusTransitionDropdown({
  currentStatus,
  itemType,
  teamId,
  itemId,
  onStatusChanged,
  onBeforeTerminalStatus,
  disabled,
  parentItemId,
}: StatusTransitionDropdownProps) {
  const qc = useQueryClient()
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingStatus, setPendingStatus] = useState<string | null>(null)
  const [open, setOpen] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const terminalStatuses =
    itemType === 'main'
      ? MAIN_TERMINAL_STATUSES
      : itemType === 'sub'
        ? SUB_TERMINAL_STATUSES
        : itemType === 'milestone-map'
          ? Object.entries(MILESTONE_MAP_STATUSES).filter(([, v]) => v.terminal).map(([k]) => k)
          : Object.entries(MILESTONE_STATUSES).filter(([, v]) => v.terminal).map(([k]) => k)

  const queryKey = [`${itemType}Transitions`, teamId, itemId]

  const fetchTransitions =
    itemType === 'main'
      ? () => getMainItemTransitionsApi(teamId, itemId)
      : itemType === 'sub'
        ? () => getSubItemTransitionsApi(teamId, itemId)
        : itemType === 'milestone-map'
          ? () => getMilestoneMapTransitionsApi(teamId, itemId)
          : () => getMilestoneTransitionsApi(teamId, itemId)

  const {
    data: transitions = [],
    isFetched,
    isFetching,
  } = useQuery({
    queryKey,
    queryFn: fetchTransitions,
    enabled: !!teamId && open,
  })

  useEffect(() => {
    if (open && isFetched && !isFetching && transitions.length === 0) {
      setOpen(false)
    }
  }, [open, isFetched, isFetching, transitions.length])

  const changeStatus = async (status: string): Promise<void> => {
    if (itemType === 'main') {
      await changeMainItemStatusApi(teamId, itemId, { status })
    } else if (itemType === 'sub') {
      await changeSubItemStatusApi(teamId, itemId, { status })
    } else if (itemType === 'milestone-map') {
      await changeMilestoneMapStatusApi(teamId, itemId, { status })
    } else {
      await changeMilestoneStatusApi(teamId, itemId, { status })
    }
  }

  const statusChangeMutation = useMutation({
    mutationFn: ({ newStatus }: { newStatus: string }) =>
      changeStatus(newStatus),
    onSuccess: () => {
      if (itemType === 'main') {
        qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
        qc.invalidateQueries({ queryKey: ['mainItem', teamId, itemId] })
      } else if (itemType === 'sub') {
        qc.invalidateQueries({
          queryKey: ['subItems', teamId, parentItemId || itemId],
        })
      } else if (itemType === 'milestone-map') {
        qc.invalidateQueries({ queryKey: ['milestoneMaps', teamId] })
        qc.invalidateQueries({ queryKey: ['milestoneMap', teamId, itemId] })
      } else {
        qc.invalidateQueries({ queryKey: ['milestones', teamId] })
        qc.invalidateQueries({ queryKey: ['milestone', teamId, itemId] })
      }
      qc.invalidateQueries({ queryKey })
      setOpen(false)
      setConfirmOpen(false)
      setPendingStatus(null)
      setErrorMessage(null)
      onStatusChanged()
    },
    onError: (err) => {
      if (isAxiosError(err) && err.response?.data?.message) {
        setErrorMessage(err.response.data.message)
      } else {
        setErrorMessage('操作失败，请稍后重试')
      }
    },
  })

  const handleSelect = useCallback(
    async (status: string) => {
      if (!terminalStatuses.includes(status)) {
        statusChangeMutation.mutate({ newStatus: status })
        return
      }
      setPendingStatus(status)
      if (onBeforeTerminalStatus) {
        const proceed = await onBeforeTerminalStatus(status)
        if (!proceed) {
          setPendingStatus(null)
          return
        }
      }
      setConfirmOpen(true)
    },
    [statusChangeMutation, terminalStatuses, onBeforeTerminalStatus],
  )

  const handleConfirm = useCallback(() => {
    if (pendingStatus) {
      statusChangeMutation.mutate({ newStatus: pendingStatus })
    }
  }, [pendingStatus, statusChangeMutation])

  if (disabled) {
    return <StatusBadge status={currentStatus} />
  }

  return (
    <>
      <div className="relative inline-flex">
        <DropdownMenu open={open} onOpenChange={setOpen}>
          <DropdownMenuTrigger asChild>
            <button className="focus:outline-none">
              <StatusBadge status={currentStatus} className="cursor-pointer" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="min-w-0 w-auto">
            {transitions.map((status) => (
              <DropdownMenuItem
                key={status}
                className="text-[13px] justify-center"
                onSelect={(e) => {
                  e.preventDefault()
                  handleSelect(status)
                }}
              >
                {getStatusName(status) || status}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      {errorMessage && (
        <div
          role="alert"
          className="mt-2 flex items-start gap-2 rounded-md border border-[var(--color-error)] bg-[var(--color-error-bg)] px-3 py-2 text-sm text-[var(--color-error-text)]"
        >
          <span className="flex-1">{errorMessage}</span>
          <button
            type="button"
            onClick={() => setErrorMessage(null)}
            className="shrink-0 opacity-70 hover:opacity-100"
            aria-label="关闭错误提示"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>确认变更状态</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <p className="text-sm text-secondary">
              确认将状态变更为「
              {getStatusName(pendingStatus || '') || pendingStatus}
              」？此操作可能不可逆。
            </p>
          </DialogBody>
          <DialogFooter>
            <Button
              variant="secondary"
              onClick={() => {
                setConfirmOpen(false)
                setPendingStatus(null)
              }}
            >
              取消
            </Button>
            <Button
              onClick={handleConfirm}
              disabled={statusChangeMutation.isPending}
            >
              确认
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
