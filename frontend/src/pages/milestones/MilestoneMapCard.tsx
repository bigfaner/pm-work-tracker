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
        <span>{map.milestoneCount} 个里程碑</span>
        <span>{map.itemCount} 个事项</span>
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

      {/* Row 4: Milestone node thumbnail — status-colored dots connected by lines */}
      {map.milestoneSummary && map.milestoneSummary.length > 0 ? (
        <div className="flex items-center gap-1 mt-3 pt-3 border-t border-border">
          <span className="text-xs text-tertiary whitespace-nowrap shrink-0">
            {map.milestoneSummary[0].name}
          </span>
          {map.milestoneSummary.map((ms, i) => (
            <span key={ms.bizKey} className="contents">
              <span
                data-testid="milestone-dot"
                className={`w-2 h-2 rounded-full shrink-0 ${
                  ms.status === 'completed'
                    ? 'bg-success'
                    : ms.status === 'in_progress'
                      ? 'bg-info'
                      : ms.status === 'cancelled'
                        ? 'bg-gray-300 opacity-50'
                        : 'bg-gray-300'
                }`}
              />
              {i < map.milestoneSummary.length - 1 && (
                <span
                  data-testid="milestone-line"
                  className="flex-1 h-px bg-border min-w-2"
                />
              )}
            </span>
          ))}
          <span className="text-xs text-tertiary whitespace-nowrap shrink-0">
            {map.milestoneSummary[map.milestoneSummary.length - 1].name}
          </span>
        </div>
      ) : map.milestoneCount === 0 ? (
        <span className="text-xs text-tertiary">暂无里程碑</span>
      ) : null}
    </Link>
  )
}
