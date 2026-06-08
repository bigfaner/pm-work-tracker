import { Link } from 'react-router-dom'
import StatusBadge from '@/components/shared/StatusBadge'
import ProgressBar from '@/components/shared/ProgressBar'
import type { MilestoneMap } from '@/types'

interface MilestoneMapCardProps {
  map: MilestoneMap
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return ''
  return dateStr.slice(0, 10)
}

export default function MilestoneMapCard({ map }: MilestoneMapCardProps) {
  const startLabel = formatDate(map.planStartDate)
  const endLabel = formatDate(map.expectedEndDate)
  const dateSpan =
    startLabel && endLabel
      ? `${startLabel} ~ ${endLabel}`
      : startLabel
        ? `${startLabel} ~`
        : endLabel
          ? `~ ${endLabel}`
          : ''

  return (
    <Link
      to={`/milestones/${map.bizKey}`}
      data-testid={`milestone-map-card-${map.bizKey}`}
      className="block rounded-xl border border-border bg-white p-4 shadow-sm transition-all hover:border-primary-300 hover:shadow-md"
    >
      {/* Row 1: Name + StatusBadge */}
      <div className="flex items-center justify-between gap-2 mb-2">
        <span className="text-sm font-medium text-primary truncate">
          {map.mapName}
        </span>
        <StatusBadge status={map.mapStatus} statusName={map.statusName} />
      </div>

      {/* Row 2: Milestone count + item count + assignee (space-between) */}
      <div className="flex items-center justify-between text-xs text-secondary mb-2">
        <span>
          {map.milestoneCount} 个里程碑 · {map.itemCount} 个事项
        </span>
        <span>{map.assigneeName}</span>
      </div>

      {/* Row 3: Date span + overall progress bar + % */}
      <div className="flex items-center justify-between gap-3 mb-2">
        <span className="text-xs text-tertiary shrink-0">{dateSpan}</span>
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <span className="text-xs text-tertiary shrink-0">整体进度</span>
          <ProgressBar
            value={map.overallProgress}
            size="sm"
            className="flex-1"
          />
          <span className="text-xs text-secondary shrink-0">
            {Math.round(map.overallProgress)}%
          </span>
        </div>
      </div>

      {/* Row 4: Milestone node thumbnail */}
      <div className="flex items-center gap-0.5">
        {Array.from({ length: map.milestoneCount }, (_, i) => (
          <span key={i} className="flex items-center">
            <span className="w-1.5 h-1.5 rounded-full bg-primary-300" />
            {i < map.milestoneCount - 1 && (
              <span className="w-3 h-px bg-border" />
            )}
          </span>
        ))}
        {map.milestoneCount === 0 && (
          <span className="text-xs text-tertiary">暂无里程碑</span>
        )}
      </div>
    </Link>
  )
}
