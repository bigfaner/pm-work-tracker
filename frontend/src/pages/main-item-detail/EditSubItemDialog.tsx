import { Input } from '@/components/ui/input'
import { DateInput } from '@/components/ui/date-input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { PrioritySelectItems } from '@/components/shared/PrioritySelect'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

export interface EditSubItemFormState {
  title: string
  priority: string
  expectedEndDate: string
  description: string
}

interface EditSubItemDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  form: EditSubItemFormState
  onFormChange: (updater: (prev: EditSubItemFormState) => EditSubItemFormState) => void
  onSubmit: () => void
  isPending: boolean
}

export default function EditSubItemDialog({
  open,
  onOpenChange,
  form,
  onFormChange,
  onSubmit,
  isPending,
}: EditSubItemDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>编辑子事项</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              标题 <span className="text-error">*</span>
            </label>
            <Input
              value={form.title}
              onChange={(e) => onFormChange((f) => ({ ...f, title: e.target.value }))}
            />
          </div>
          <div className="grid grid-cols-2 gap-3 mb-4">
            <div>
              <label className="block text-sm font-medium text-primary mb-1">优先级</label>
              <Select value={form.priority} onValueChange={(v) => onFormChange((f) => ({ ...f, priority: v }))}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <PrioritySelectItems />
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="block text-sm font-medium text-primary mb-1">预期完成时间</label>
              <DateInput
                value={form.expectedEndDate}
                onChange={(e) => onFormChange((f) => ({ ...f, expectedEndDate: e.target.value }))}
              />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-primary mb-1">描述</label>
            <Textarea
              rows={3}
              value={form.description}
              onChange={(e) => onFormChange((f) => ({ ...f, description: e.target.value }))}
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)}>取消</Button>
          <Button onClick={onSubmit} disabled={!form.title.trim() || isPending}>保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
