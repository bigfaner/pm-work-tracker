import { Input } from '@/components/ui/input'
import { DateInput } from '@/components/ui/date-input'
import { Textarea } from '@/components/ui/textarea'
import { MemberSelect } from '@/components/shared/MemberSelect'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import type { MilestoneMap } from '@/types'

export interface MilestoneMapFormState {
  mapName: string
  assigneeKey: string
  planStartDate: string
  expectedEndDate: string
  mapDesc: string
}

interface CreateMilestoneMapDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  form: MilestoneMapFormState
  onFormChange: (
    updater: (prev: MilestoneMapFormState) => MilestoneMapFormState,
  ) => void
  members: { userKey: string; displayName: string }[]
  onSubmit: () => void
  isPending: boolean
  /** Provide to switch to edit mode with pre-filled title */
  milestoneMap?: MilestoneMap
}

export default function CreateMilestoneMapDialog({
  open,
  onOpenChange,
  form,
  onFormChange,
  members,
  onSubmit,
  isPending,
  milestoneMap,
}: CreateMilestoneMapDialogProps) {
  const isEdit = !!milestoneMap
  const hasDateError =
    !!form.planStartDate &&
    !!form.expectedEndDate &&
    form.expectedEndDate < form.planStartDate

  const canSubmit =
    form.mapName.trim() !== '' &&
    form.assigneeKey !== '' &&
    !hasDateError &&
    !isPending

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? '编辑里程碑图' : '创建里程碑图'}
          </DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              名称 <span className="text-error">*</span>
            </label>
            <Input
              placeholder="请输入里程碑图名称"
              maxLength={100}
              value={form.mapName}
              disabled={isPending}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, mapName: e.target.value }))
              }
            />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              负责人 <span className="text-error">*</span>
            </label>
            <MemberSelect
              members={members}
              selectedId={form.assigneeKey}
              onSelect={(v) =>
                onFormChange((f) => ({ ...f, assigneeKey: v }))
              }
              placeholder="选择负责人"
              allowEmpty={false}
            />
          </div>
          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                计划开始时间
              </label>
              <DateInput
                value={form.planStartDate}
                disabled={isPending}
                onChange={(e) =>
                  onFormChange((f) => ({ ...f, planStartDate: e.target.value }))
                }
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                计划完成时间
              </label>
              <DateInput
                value={form.expectedEndDate}
                disabled={isPending}
                onChange={(e) =>
                  onFormChange((f) => ({
                    ...f,
                    expectedEndDate: e.target.value,
                  }))
                }
              />
            </div>
          </div>
          {hasDateError && (
            <p className="text-xs text-error mb-2">
              计划完成时间不得早于计划开始时间
            </p>
          )}
          <div>
            <label className="block text-sm font-medium text-primary mb-1">
              描述
            </label>
            <Textarea
              placeholder="请输入描述（可选）"
              value={form.mapDesc}
              disabled={isPending}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, mapDesc: e.target.value }))
              }
              className="h-[160px] resize-y"
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)} disabled={isPending}>
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
