import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const priorityVariantMap: Record<string, string> = {
  P1: 'priority-high',
  P2: 'priority-medium',
  P3: 'priority-low',
}

interface PriorityBadgeProps {
  priority: string
  className?: string
}

export default function PriorityBadge({ priority, className }: PriorityBadgeProps) {
  const variant = priorityVariantMap[priority] || 'default'
  return (
    <Badge variant={variant as any} className={cn('text-[11px]', className)}>
      {priority}
    </Badge>
  )
}
