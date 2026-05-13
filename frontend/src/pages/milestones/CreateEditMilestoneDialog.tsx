import { useState, useEffect } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { DateInput } from '@/components/ui/date-input'
import type { Milestone } from '@/types'

interface CreateEditMilestoneDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  mode: 'create' | 'edit'
  milestone?: Milestone | null
  onSubmit: (data: { milestoneName: string, expectedEndDate: string }) => void
  isPending: boolean
}

export default function CreateEditMilestoneDialog({
  open,
  onOpenChange,
  mode,
  milestone,
  onSubmit,
  isPending,
}: CreateEditMilestoneDialogProps) {
  const [name, setName] = useState('')
  const [endDate, setEndDate] = useState('')
  const [nameError, setNameError] = useState('')
  const [dateError, setDateError] = useState('')

  useEffect(() => {
    if (open && mode === 'edit' && milestone) {
      setName(milestone.milestoneName)
      setEndDate(milestone.expectedEndDate)
    } else if (open && mode === 'create') {
      setName('')
      setEndDate('')
    }
    setNameError('')
    setDateError('')
  }, [open, mode, milestone])

  function handleConfirm() {
    let valid = true
    if (!name.trim()) {
      setNameError('请输入里程碑名称')
      valid = false
    }
    if (!endDate) {
      setDateError('请选择计划完成时间')
      valid = false
    }
    if (!valid) return

    onSubmit({ milestoneName: name.trim(), expectedEndDate: endDate })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>{mode === 'create' ? '创建里程碑' : '编辑里程碑'}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="mb-4">
            <label className="block text-[13px] font-medium text-secondary mb-1.5">
              名称 <span className="text-error">*</span>
            </label>
            <Input
              value={name}
              onChange={(e) => {
                setName(e.target.value)
                if (e.target.value.trim()) setNameError('')
              }}
              placeholder="请输入里程碑名称"
              maxLength={100}
            />
            {nameError && (
              <p className="text-[12px] text-error-text mt-1">{nameError}</p>
            )}
          </div>
          <div>
            <label className="block text-[13px] font-medium text-secondary mb-1.5">
              计划完成时间 <span className="text-error">*</span>
            </label>
            <DateInput
              value={endDate}
              onChange={(e) => {
                setEndDate(e.target.value)
                if (e.target.value) setDateError('')
              }}
            />
            {dateError && (
              <p className="text-[12px] text-error-text mt-1">{dateError}</p>
            )}
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button onClick={handleConfirm} disabled={isPending}>
            {isPending ? '保存中...' : '确认'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
