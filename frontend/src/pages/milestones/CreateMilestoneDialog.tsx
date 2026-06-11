import { Input } from '@/components/ui/input'
import { DateInput } from '@/components/ui/date-input'
import { Textarea } from '@/components/ui/textarea'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import type { Milestone } from '@/types'

export interface MilestoneFormState {
  milestoneName: string
  expectedEndDate: string
  milestoneDesc: string
}

interface CreateMilestoneDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  form: MilestoneFormState
  onFormChange: (
    updater: (prev: MilestoneFormState) => MilestoneFormState,
  ) => void
  onSubmit: () => void
  isPending: boolean
  /** Provide to switch to edit mode with pre-filled title */
  milestone?: Milestone
}

export default function CreateMilestoneDialog({
  open,
  onOpenChange,
  form,
  onFormChange,
  onSubmit,
  isPending,
  milestone,
}: CreateMilestoneDialogProps) {
  const isEdit = !!milestone

  const canSubmit =
    form.milestoneName.trim() !== '' &&
    form.expectedEndDate !== '' &&
    !isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? '编辑里程碑' : '创建里程碑'}
          </DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              名称 <span className="text-error">*</span>
            </label>
            <Input
              placeholder="请输入里程碑名称"
              maxLength={100}
              value={form.milestoneName}
              disabled={isPending}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, milestoneName: e.target.value }))
              }
            />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              计划完成时间 <span className="text-error">*</span>
            </label>
            <DateInput
              value={form.expectedEndDate}
              disabled={isPending}
              aria-label="计划完成时间"
              onChange={(e) =>
                onFormChange((f) => ({
                  ...f,
                  expectedEndDate: e.target.value,
                }))
              }
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-primary mb-1">
              描述
            </label>
            <Textarea
              placeholder="请输入描述（可选）"
              value={form.milestoneDesc}
              disabled={isPending}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, milestoneDesc: e.target.value }))
              }
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button
            variant="secondary"
            onClick={() => onOpenChange(false)}
            disabled={isPending}
          >
            取消
          </Button>
          <Button onClick={onSubmit} disabled={!canSubmit}>
            确认
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
