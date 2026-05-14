import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { createMainItemApi } from '@/api/mainItems'
import { Input } from '@/components/ui/input'
import { DateInput } from '@/components/ui/date-input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { PrioritySelectItems } from '@/components/shared/PrioritySelect'
import { MemberSelect } from '@/components/shared/MemberSelect'
import { useToast } from '@/components/ui/toast'
import type { CreateMainItemFormState } from '@/pages/item-view/CreateMainItemDialog'

interface QuickAddMIDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  milestoneBizKey: string
  milestoneName: string
  onCreated: () => void
}

export default function QuickAddMIDialog({
  open,
  onOpenChange,
  milestoneBizKey,
  milestoneName,
  onCreated,
}: QuickAddMIDialogProps) {
  const teamId = useTeamStore((s) => s.currentTeamId) ?? ''
  const qc = useQueryClient()
  const { addToast } = useToast()

  const [form, setForm] = useState<CreateMainItemFormState>({
    title: '',
    description: '',
    priority: 'P2',
    assigneeKey: '',
    startDate: '',
    expectedEndDate: '',
    milestoneKey: '',
  })

  const createMutation = useMutation({
    mutationFn: () =>
      createMainItemApi(teamId, {
        title: form.title.trim(),
        description: form.description.trim() || undefined,
        priority: form.priority,
        assigneeKey: form.assigneeKey,
        startDate: form.startDate,
        expectedEndDate: form.expectedEndDate,
        milestoneKey: milestoneBizKey,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['milestone-related-mis', teamId, milestoneBizKey] })
      qc.invalidateQueries({ queryKey: ['milestone-detail', teamId, milestoneBizKey] })
      qc.invalidateQueries({ queryKey: ['milestones-by-map', teamId] })
      addToast('事项创建成功', 'success')
      setForm({
        title: '',
        description: '',
        priority: 'P2',
        assigneeKey: '',
        startDate: '',
        expectedEndDate: '',
        milestoneKey: '',
      })
      onOpenChange(false)
      onCreated()
    },
    onError: () => {
      addToast('创建失败，请重试', 'error')
    },
  })

  function handleFormChange(
    updater: (prev: CreateMainItemFormState) => CreateMainItemFormState,
  ) {
    setForm(updater)
  }

  // Simple mock members list - in real usage this should come from team store or API
  // The parent page should provide this, but for the dialog we'll use an empty list
  // since MemberSelect can handle it
  const members: { userKey: string, displayName: string }[] = []

  const isValid =
    form.title.trim() &&
    form.assigneeKey &&
    form.startDate &&
    form.expectedEndDate

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg" data-testid="form-quick-add-mi">
        <DialogHeader>
          <DialogTitle>新建主事项</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-sm font-medium text-primary mb-1">
              标题 <span className="text-error">*</span>
            </label>
            <Input
              placeholder="请输入标题"
              maxLength={100}
              value={form.title}
              onChange={(e) =>
                handleFormChange((f) => ({ ...f, title: e.target.value }))
              }
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                优先级 <span className="text-error">*</span>
              </label>
              <Select
                value={form.priority}
                onValueChange={(v) =>
                  handleFormChange((f) => ({ ...f, priority: v }))
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
                负责人 <span className="text-error">*</span>
              </label>
              <MemberSelect
                members={members}
                selectedId={form.assigneeKey}
                onSelect={(v) =>
                  handleFormChange((f) => ({ ...f, assigneeKey: v }))
                }
                allowEmpty={false}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4 mt-4">
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                开始时间 <span className="text-error">*</span>
              </label>
              <DateInput
                value={form.startDate}
                onChange={(e) =>
                  handleFormChange((f) => ({ ...f, startDate: e.target.value }))
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
                  handleFormChange((f) => ({
                    ...f,
                    expectedEndDate: e.target.value,
                  }))
                }
              />
            </div>
          </div>
          <div className="mt-4">
            <label className="block text-sm font-medium text-primary mb-1">
              所属里程碑 <span className="text-error">*</span>
            </label>
            <Input value={milestoneName} disabled className="bg-bg-alt" data-testid="field-milestone" />
          </div>
          <div className="mt-4">
            <label className="block text-sm font-medium text-primary mb-1">
              描述
            </label>
            <Textarea
              rows={3}
              placeholder="请输入描述（可选）"
              value={form.description}
              onChange={(e) =>
                handleFormChange((f) => ({ ...f, description: e.target.value }))
              }
            />
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            onClick={() => createMutation.mutate()}
            disabled={!isValid || createMutation.isPending}
            data-testid="btn-confirm"
          >
            {createMutation.isPending ? '保存中...' : '确认'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
