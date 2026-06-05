import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listMainItemsApi } from '@/api/mainItems'
import { moveSubItemApi } from '@/api/subItems'
import { MAIN_ITEM_STATUSES } from '@/lib/status'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'

interface MoveSubItemDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  subItemBizKey: string
  currentMainItemBizKey: string
  teamId: string
  onSuccess?: () => void
}

export default function MoveSubItemDialog({
  open,
  onOpenChange,
  subItemBizKey,
  currentMainItemBizKey,
  teamId,
  onSuccess,
}: MoveSubItemDialogProps) {
  const [target, setTarget] = useState('')
  const qc = useQueryClient()

  const { data: mainItemsPage } = useQuery({
    queryKey: ['mainItems', teamId, 'for-move'],
    queryFn: () => listMainItemsApi(teamId),
    enabled: open,
  })

  const moveMutation = useMutation({
    mutationFn: (targetMainItemBizKey: string) =>
      moveSubItemApi(teamId, subItemBizKey, targetMainItemBizKey),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      onOpenChange(false)
      setTarget('')
      onSuccess?.()
    },
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>移动到其他主事项</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <Select value={target} onValueChange={setTarget}>
            <SelectTrigger>
              <SelectValue placeholder="选择目标主事项" />
            </SelectTrigger>
            <SelectContent>
              {(mainItemsPage?.items || [])
                .filter((mi) => {
                  if (mi.bizKey === currentMainItemBizKey) return false
                  const statusDef = MAIN_ITEM_STATUSES[
                    mi.itemStatus as keyof typeof MAIN_ITEM_STATUSES
                  ]
                  return !statusDef?.terminal
                })
                .map((mi) => (
                  <SelectItem key={mi.bizKey} value={mi.bizKey}>
                    {mi.code} - {mi.title}
                  </SelectItem>
                ))}
            </SelectContent>
          </Select>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={() => target && moveMutation.mutate(target)}
            disabled={!target || moveMutation.isPending}
          >
            确认移动
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
