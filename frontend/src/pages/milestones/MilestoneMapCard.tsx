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

      {/* Row 4: Milestone node summary */}
      <div className="flex flex-wrap gap-1.5">
        {map.milestoneSummary && map.milestoneSummary.length > 0
          ? map.milestoneSummary.map((ms) => (
              <span
                key={ms.bizKey}
                className="inline-flex items-center gap-1 rounded-md bg-bg-alt px-1.5 py-0.5 text-xs text-secondary"
              >
                <span
                  className={`w-1.5 h-1.5 rounded-full ${
                    ms.status === 'completed'
                      ? 'bg-success'
                      : ms.status === 'in_progress'
                        ? 'bg-warning'
                        : 'bg-gray-300'
                  }`}
                />
                <span className="truncate max-w-[80px]">{ms.name}</span>
                <span className="text-tertiary">
                  {Math.round(ms.progress)}%
                </span>
              </span>
            ))
          : map.milestoneCount === 0 && (
              <span className="text-xs text-tertiary">暂无里程碑</span>
            )}
      </div>
    </Link>
  )
}
