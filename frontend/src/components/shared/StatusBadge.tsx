import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const statusVariantMap: Record<string, string> = {
  '未开始': 'status-planning',
  '进行中': 'status-in-progress',
  '待评审': 'status-pending',
  '已完成': 'status-completed',
  '已关闭': 'status-cancelled',
  '阻塞中': 'status-overdue',
  '延期': 'status-on-hold',
}

interface StatusBadgeProps {
  status: string
  className?: string
}

export default function StatusBadge({ status, className }: StatusBadgeProps) {
  const variant = statusVariantMap[status] || 'default'
  return (
    <Badge variant={variant as any} className={cn('text-[11px]', className)}>
      {status}
    </Badge>
  )
}
