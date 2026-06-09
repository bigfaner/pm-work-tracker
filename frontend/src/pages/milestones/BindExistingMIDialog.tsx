import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listMainItemsApi, updateMainItemApi } from '@/api/mainItems'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { CheckboxGroup } from '@/components/ui/checkbox-group'
import { useToast } from '@/components/ui/toast'

interface BindExistingMIDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  teamId: string
  milestoneBizKey: string
  milestoneName: string
}

export default function BindExistingMIDialog({
  open,
  onOpenChange,
  teamId,
  milestoneBizKey,
  milestoneName,
}: BindExistingMIDialogProps) {
  const [selected, setSelected] = useState<string[]>([])
  const qc = useQueryClient()
  const { addToast } = useToast()

  const { data: unboundResult, isLoading } = useQuery({
    queryKey: ['unassignedMIs', teamId],
    queryFn: () =>
      listMainItemsApi(teamId, { milestoneKey: 'unassigned', pageSize: 200 }),
    enabled: open && !!teamId,
  })

  const options = (unboundResult?.items ?? []).map((mi) => ({
    value: mi.bizKey,
    label: `${mi.code} - ${mi.title}`,
  }))

  const bindMutation = useMutation({
    mutationFn: (bizKeys: string[]) =>
      Promise.all(
        bizKeys.map((bizKey) =>
          updateMainItemApi(teamId, bizKey, { milestoneKey: milestoneBizKey }),
        ),
      ),
    onSuccess: (_data, bizKeys) => {
      qc.invalidateQueries({ queryKey: ['milestoneMIs', teamId, milestoneBizKey] })
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      qc.invalidateQueries({ queryKey: ['milestones', teamId] })
      qc.invalidateQueries({ queryKey: ['milestone', teamId, milestoneBizKey] })
      addToast(`已关联 ${bizKeys.length} 个事项到「${milestoneName}」`, 'success')
      setSelected([])
      onOpenChange(false)
    },
  })

  const handleSubmit = () => {
    if (selected.length > 0) {
      bindMutation.mutate(selected)
    }
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) setSelected([])
    onOpenChange(nextOpen)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>关联已有事项到「{milestoneName}」</DialogTitle>
        </DialogHeader>
        <DialogBody>
          {isLoading ? (
            <div className="space-y-2 animate-pulse">
              <div className="h-5 bg-bg-alt rounded w-3/4" />
              <div className="h-5 bg-bg-alt rounded w-1/2" />
              <div className="h-5 bg-bg-alt rounded w-2/3" />
            </div>
          ) : options.length === 0 ? (
            <p className="text-sm text-tertiary py-4 text-center">
              暂无可关联的事项
            </p>
          ) : (
            <CheckboxGroup
              title="选择事项"
              options={options}
              selected={selected}
              onChange={setSelected}
            />
          )}
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => handleOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={selected.length === 0 || bindMutation.isPending}
          >
            {bindMutation.isPending ? '关联中...' : `关联 (${selected.length})`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
