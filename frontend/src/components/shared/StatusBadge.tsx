import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { getStatusVariant, getStatusName } from '@/lib/status'

interface StatusBadgeProps {
  status: string
  statusName?: string
  className?: string
}

export default function StatusBadge({ status, statusName, className }: StatusBadgeProps) {
  const variant = `status-${getStatusVariant(status)}` as 'status-planning' | 'status-in-progress' | 'status-completed' | 'status-on-hold' | 'status-cancelled' | 'status-overdue' | 'status-pending' | 'default'
  const displayText = statusName || getStatusName(status) || status
  return (
    <Badge variant={variant} className={cn('text-[11px]', className)}>
      {displayText}
    </Badge>
  )
}
