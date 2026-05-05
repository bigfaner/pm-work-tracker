import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

export interface AppendProgressFormState {
  completion: string;
  achievement: string;
  blocker: string;
}

interface AppendProgressDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  targetName: string;
  form: AppendProgressFormState;
  onFormChange: (
    updater: (prev: AppendProgressFormState) => AppendProgressFormState,
  ) => void;
  onSubmit: () => void;
  isPending: boolean;
}

export default function AppendProgressDialog({
  open,
  onOpenChange,
  targetName,
  form,
  onFormChange,
  onSubmit,
  isPending,
}: AppendProgressDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>追加进度 → {targetName}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              进度 (0-100) <span className="text-error">*</span>
            </label>
            <Input
              type="number"
              min={0}
              max={100}
              placeholder="请输入进度"
              value={form.completion}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, completion: e.target.value }))
              }
            />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              完成情况
            </label>
            <Textarea
              rows={3}
              placeholder="请输入完成情况（可选）"
              value={form.achievement}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, achievement: e.target.value }))
              }
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-primary mb-1">
              阻塞问题
            </label>
            <Textarea
              rows={3}
              placeholder="请输入阻塞问题（可选）"
              value={form.blocker}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, blocker: e.target.value }))
              }
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button onClick={onSubmit} disabled={!form.completion || isPending}>
            确认
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
