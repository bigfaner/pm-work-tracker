import { Input } from "@/components/ui/input";
import { DateInput } from "@/components/ui/date-input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PrioritySelectItems } from "@/components/shared/PrioritySelect";
import { MemberSelect } from "@/components/shared/MemberSelect";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

export interface CreateSubItemFormState {
  title: string;
  priority: string;
  assigneeKey: string;
  startDate: string;
  expectedEndDate: string;
  description: string;
}

interface CreateSubItemDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  targetName: string;
  form: CreateSubItemFormState;
  onFormChange: (
    updater: (prev: CreateSubItemFormState) => CreateSubItemFormState,
  ) => void;
  members: { userKey: string; displayName: string }[];
  onSubmit: () => void;
  isPending: boolean;
}

export default function CreateSubItemDialog({
  open,
  onOpenChange,
  targetName,
  form,
  onFormChange,
  members,
  onSubmit,
  isPending,
}: CreateSubItemDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>新增子事项 → {targetName}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              标题 <span className="text-error">*</span>
            </label>
            <Input
              placeholder="请输入子事项标题"
              value={form.title}
              onChange={(e) =>
                onFormChange((f) => ({ ...f, title: e.target.value }))
              }
            />
          </div>
          <div className="grid grid-cols-2 gap-3 mb-4">
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                优先级 <span className="text-error">*</span>
              </label>
              <Select
                value={form.priority}
                onValueChange={(v) =>
                  onFormChange((f) => ({ ...f, priority: v }))
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="请选择优先级" />
                </SelectTrigger>
                <SelectContent>
                  <PrioritySelectItems />
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                负责人 <span className="text-error">*</span>
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
          <div className="grid grid-cols-2 gap-3 mb-4">
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                开始时间 <span className="text-error">*</span>
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
                预期完成时间 <span className="text-error">*</span>
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
          <div>
            <label className="block text-sm font-medium text-primary mb-1">
              描述
            </label>
            <Textarea
              rows={3}
              placeholder="请输入子事项描述（可选）"
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
          <Button
            onClick={onSubmit}
            disabled={
              !form.title.trim() ||
              !form.priority ||
              !form.assigneeKey ||
              !form.startDate ||
              !form.expectedEndDate ||
              isPending
            }
          >
            确认
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
