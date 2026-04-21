import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { getStatusVariant, getStatusName } from '@/lib/status'

interface StatusBadgeProps {
  status: string
  statusName?: string
  className?: string
}

export default function StatusBadge({ status, statusName, className }: StatusBadgeProps) {
  const variant = getStatusVariant(status)
  const displayText = statusName || getStatusName(status) || status
  return (
    <Badge variant={`status-${variant}` as any} className={cn('text-[11px]', className)}>
      {displayText}
    </Badge>
  )
}
