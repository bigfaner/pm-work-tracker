import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

interface ConfirmDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: string
  confirmLabel?: string
  confirmVariant?: 'primary' | 'danger'
  confirmPlaceholder?: string
  onConfirm: (inputValue?: string) => void
}

export default function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = '确认',
  confirmVariant = 'primary',
  confirmPlaceholder,
  onConfirm,
}: ConfirmDialogProps) {
  const [inputValue, setInputValue] = useState('')

  const handleConfirm = () => {
    onConfirm(confirmPlaceholder ? inputValue : undefined)
    setInputValue('')
  }

  const handleOpenChange = (value: boolean) => {
    if (!value) setInputValue('')
    onOpenChange(value)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent size="sm">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        <DialogBody>
          {confirmPlaceholder && (
            <Input
              placeholder={confirmPlaceholder}
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
            />
          )}
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => handleOpenChange(false)}>
            取消
          </Button>
          <Button
            variant={confirmVariant === 'danger' ? 'danger' : 'primary'}
            onClick={handleConfirm}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
