import { cn } from '@/lib/utils'
import ProgressBar from '@/components/shared/ProgressBar'
import type { Milestone } from '@/types'

interface MilestoneNodeProps {
  milestone: Milestone;
  selected?: boolean;
  onClick?: () => void;
  style?: React.CSSProperties;
}

/** Map milestone status to status dot color class */
function getStatusDotClass(status: string): string {
  switch (status) {
    case 'in_progress':
      return 'text-primary'
    case 'completed':
      return 'text-success'
    case 'not_started':
    case 'cancelled':
    default:
      return 'text-tertiary'
  }
}

export default function MilestoneNode({
  milestone,
  selected = false,
  onClick,
  style,
}: MilestoneNodeProps) {
  const isCancelled = milestone.milestoneStatus === 'cancelled'

  return (
    <div
      data-testid={`milestone-node-${milestone.bizKey}`}
      role="button"
      tabIndex={0}
      aria-label={milestone.milestoneName}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter') onClick?.()
      }}
      style={style}
      className={cn(
        'w-40 rounded-xl border border-border bg-white p-3.5 cursor-pointer transition-all duration-200',
        isCancelled && 'opacity-50',
        selected ? 'border-primary ring-2 ring-primary-100' : 'hover:bg-bg-alt',
      )}
    >
      {/* Row 1: Status dot + Name + Completion % */}
      <div className="flex items-center gap-1.5 mb-1.5">
        <span
          className={cn(
            'inline-block w-2 h-2 rounded-full bg-current shrink-0',
            getStatusDotClass(milestone.milestoneStatus),
          )}
        />
        <span className="text-sm font-semibold text-primary truncate flex-1">
          {milestone.milestoneName}
        </span>
        <span className="text-xs text-secondary font-medium shrink-0">
          {Math.round(milestone.completion)}%
        </span>
      </div>

      {/* Row 2: Date */}
      <div className="text-xs text-tertiary mb-1">
        {milestone.expectedEndDate ?? '未设置'}
      </div>

      {/* Row 3: MI count */}
      <div className="text-xs text-tertiary mb-2">
        {milestone.relatedMICount} 个事项
      </div>

      {/* Row 4: Progress bar */}
      <ProgressBar value={milestone.completion} size="sm" />
    </div>
  )
}
