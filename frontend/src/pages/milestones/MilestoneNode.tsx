import { cn } from '@/lib/utils'
import StatusBadge from '@/components/shared/StatusBadge'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import type { Milestone } from '@/types'

interface MilestoneNodeProps {
  milestone: Milestone
  selected?: boolean
  onClick?: () => void
  style?: React.CSSProperties
  isDragOver?: boolean
  onDragOver?: (e: React.DragEvent) => void
  onDragLeave?: () => void
  onDrop?: (e: React.DragEvent) => void
}

export default function MilestoneNode({
  milestone,
  selected = false,
  onClick,
  style,
  isDragOver = false,
  onDragOver,
  onDragLeave,
  onDrop,
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
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
      style={style}
      className={cn(
        'w-40 rounded-xl border border-border bg-white p-3.5 cursor-pointer transition-all duration-200',
        isCancelled && 'opacity-50',
        isDragOver
          ? 'border-primary ring-2 ring-primary-200 bg-primary-50'
          : selected
            ? 'border-primary ring-2 ring-primary-100'
            : 'hover:bg-bg-alt',
      )}
    >
      {/* Row 1: Name + Status badge */}
      <div className="flex items-center gap-1.5 mb-1.5">
        <span className="text-sm font-semibold text-primary truncate flex-1">
          {milestone.milestoneName}
        </span>
        <StatusBadge
          status={milestone.milestoneStatus}
          statusName={milestone.statusName}
        />
      </div>

      {/* Row 2: Description (truncated, tooltip for full text) */}
      {milestone.milestoneDesc ? (
        <Tooltip>
          <TooltipTrigger asChild>
            <p
              className="text-xs text-secondary"
              style={{
                display: '-webkit-box',
                WebkitLineClamp: 2,
                WebkitBoxOrient: 'vertical',
                overflow: 'hidden',
              }}
            >
              {milestone.milestoneDesc}
            </p>
          </TooltipTrigger>
          <TooltipContent side="bottom" className="max-w-xs">
            {milestone.milestoneDesc}
          </TooltipContent>
        </Tooltip>
      ) : null}
    </div>
  )
}
