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
import { getStatusName, MAIN_TERMINAL_STATUSES, SUB_TERMINAL_STATUSES } from '@/lib/status'
import {
  getMainItemTransitionsApi,
  changeMainItemStatusApi,
} from '@/api/mainItems'
import {
  getSubItemTransitionsApi,
  changeSubItemStatusApi,
} from '@/api/subItems'

export interface StatusTransitionDropdownProps {
  currentStatus: string
  itemType: 'main' | 'sub'
  teamId: string
  itemId: string
  onStatusChanged: () => void
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
}: StatusTransitionDropdownProps) {
  const qc = useQueryClient()
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [pendingStatus, setPendingStatus] = useState<string | null>(null)
  const [open, setOpen] = useState(false)
  const [showTip, setShowTip] = useState(false)

  const terminalStatuses = itemType === 'main' ? MAIN_TERMINAL_STATUSES : SUB_TERMINAL_STATUSES

  const queryKey = itemType === 'main'
    ? ['mainItemTransitions', teamId, itemId]
    : ['subItemTransitions', teamId, itemId]

  const fetchTransitions = itemType === 'main'
    ? () => getMainItemTransitionsApi(teamId, itemId)
    : () => getSubItemTransitionsApi(teamId, itemId)

  const { data: transitions = [], isFetched, isFetching } = useQuery({
    queryKey,
    queryFn: fetchTransitions,
    enabled: !!teamId && open,
  })

  useEffect(() => {
    if (open && isFetched && !isFetching && transitions.length === 0) {
      setOpen(false)
      setShowTip(true)
      setTimeout(() => setShowTip(false), 2000)
    }
  }, [open, isFetched, isFetching, transitions.length])

  const changeStatus = async (status: string): Promise<void> => {
    if (itemType === 'main') {
      await changeMainItemStatusApi(teamId, itemId, { status })
    } else {
      await changeSubItemStatusApi(teamId, itemId, { status })
    }
  }

  const statusChangeMutation = useMutation({
    mutationFn: ({ newStatus }: { newStatus: string }) => changeStatus(newStatus),
    onSuccess: () => {
      if (itemType === 'main') {
        qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
        qc.invalidateQueries({ queryKey: ['mainItem', teamId, itemId] })
      } else {
        qc.invalidateQueries({ queryKey: ['subItem', teamId, itemId] })
      }
      qc.invalidateQueries({ queryKey })
      setOpen(false)
      setConfirmOpen(false)
      setPendingStatus(null)
      onStatusChanged()
    },
  })

  const handleSelect = useCallback(async (status: string) => {
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
  }, [statusChangeMutation, terminalStatuses, onBeforeTerminalStatus])

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
        {showTip && (
          <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1.5 whitespace-nowrap text-xs px-2 py-1 rounded-md bg-primary text-white shadow-md pointer-events-none z-50">
            暂无可用流转
          </div>
        )}
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

      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent size="sm">
          <DialogHeader>
            <DialogTitle>确认变更状态</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <p className="text-sm text-secondary">
              确认将状态变更为「{getStatusName(pendingStatus || '') || pendingStatus}」？此操作可能不可逆。
            </p>
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => { setConfirmOpen(false); setPendingStatus(null) }}>取消</Button>
            <Button onClick={handleConfirm} disabled={statusChangeMutation.isPending}>确认</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
