import { Input } from '@/components/ui/input'
import { DateInput } from '@/components/ui/date-input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { PrioritySelectItems } from '@/components/shared/PrioritySelect'
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
import type { Milestone } from '@/types'

export interface EditMainItemFormState {
  title: string
  priority: string
  assigneeKey: string
  startDate: string
  expectedEndDate: string
  description: string
  milestoneKey: string
}

interface EditMainItemDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  form: EditMainItemFormState
  onFormChange: (
    updater: (prev: EditMainItemFormState) => EditMainItemFormState,
  ) => void
  members: { userKey: string, displayName: string }[]
  milestones: Milestone[]
  milestonesError: boolean
  onSubmit: () => void
  isPending: boolean
}

const UNASSIGNED = '__unassigned__'

export default function EditMainItemDialog({
  open,
  onOpenChange,
  form,
  onFormChange,
  members,
  milestones,
  milestonesError,
  onSubmit,
  isPending,
}: EditMainItemDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>编辑主事项</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              标题 <span className="text-error">*</span>
            </label>
            <Input
              maxLength={100}
              value={form.title}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, title: e.target.value }))
              }
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                优先级
              </label>
              <Select
                value={form.priority}
                onValueChange={(v) =>
                  onFormChange((f) => ({ ...f, priority: v }))
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <PrioritySelectItems />
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                负责人
              </label>
              <MemberSelect
                members={members}
                selectedId={form.assigneeKey}
                onSelect={(v) =>
                  onFormChange((f) => ({ ...f, assigneeKey: v }))
                }
              />
            </div>
          </div>
          <div className="mt-4">
            <label className="block text-sm font-medium text-primary mb-1">
              所属里程碑
            </label>
            <Select
              value={form.milestoneKey || UNASSIGNED}
              onValueChange={(v) =>
                onFormChange((f) => ({
                  ...f,
                  milestoneKey: v === UNASSIGNED ? '' : v,
                }))
              }
              disabled={milestonesError}
            >
              <SelectTrigger>
                {milestonesError ? (
                  <span className="text-tertiary">加载失败</span>
                ) : (
                  <SelectValue />
                )}
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={UNASSIGNED}>未分配</SelectItem>
                {milestones.map((ms) => (
                  <SelectItem key={ms.bizKey} value={ms.bizKey}>
                    {ms.milestoneName}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="grid grid-cols-2 gap-4 mt-4">
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                开始时间
              </label>
              <DateInput
                value={form.startDate}
                onChange={(e) =>
                  onFormChange((f) => ({ ...f, startDate: e.target.value }))
                }
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                预期完成时间
              </label>
              <DateInput
                value={form.expectedEndDate}
                onChange={(e) =>
                  onFormChange((f) => ({
                    ...f,
                    expectedEndDate: e.target.value,
                  }))
                }
              />
            </div>
          </div>
          <div className="mt-4">
            <label className="block text-sm font-medium text-primary mb-1">
              描述
            </label>
            <Textarea
              rows={3}
              value={form.description}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, description: e.target.value }))
              }
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button onClick={onSubmit} disabled={!form.title.trim() || isPending}>
            确认
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
