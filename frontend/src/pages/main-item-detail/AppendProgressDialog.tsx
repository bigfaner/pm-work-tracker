import { Input } from '@/components/ui/input'
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

export interface AppendProgressFormState {
  completion: string
  achievement: string
  blocker: string
}

interface AppendProgressDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  form: AppendProgressFormState
  onFormChange: (updater: (prev: AppendProgressFormState) => AppendProgressFormState) => void
  onSubmit: () => void
  isPending: boolean
}

export default function AppendProgressDialog({
  open,
  onOpenChange,
  form,
  onFormChange,
  onSubmit,
  isPending,
}: AppendProgressDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>追加进度</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              完成度 (%) <span className="text-error">*</span>
            </label>
            <Input
              type="number"
              min={0}
              max={100}
              value={form.completion}
              onChange={(e) => onFormChange((f) => ({ ...f, completion: e.target.value }))}
            />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">成果</label>
            <Textarea
              rows={2}
              placeholder="本次进展成果（可选）"
              value={form.achievement}
              onChange={(e) => onFormChange((f) => ({ ...f, achievement: e.target.value }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-primary mb-1">卡点</label>
            <Textarea
              rows={2}
              placeholder="遇到的阻碍（可选）"
              value={form.blocker}
              onChange={(e) => onFormChange((f) => ({ ...f, blocker: e.target.value }))}
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)}>取消</Button>
          <Button onClick={onSubmit} disabled={form.completion === '' || isPending}>确认</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
