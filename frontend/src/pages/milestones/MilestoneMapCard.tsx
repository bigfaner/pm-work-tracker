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

function dotColor(status: string) {
  if (status === 'completed') return 'bg-success'
  if (status === 'in_progress') return 'bg-success'
  if (status === 'cancelled') return 'bg-gray-400 opacity-50'
  return 'bg-gray-400'
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

      {/* Row 4: Milestone node thumbnail — name ● ── ● ── ● name */}
      {map.milestoneSummary && map.milestoneSummary.length > 0 ? (
        <div className="flex items-center gap-1 mt-3 pt-3 border-t border-border overflow-hidden">
          <span className="text-xs text-tertiary truncate max-w-[60px] shrink-0">
            {map.milestoneSummary[0].name}
          </span>
          {buildThumbnailElements(map.milestoneSummary)}
        </div>
      ) : map.milestoneCount === 0 ? (
        <span className="text-xs text-tertiary">暂无里程碑</span>
      ) : null}
    </Link>
  )
}

function buildThumbnailElements(summary: MilestoneMap['milestoneSummary']) {
  const elements: React.ReactNode[] = []

  // First dot
  elements.push(
    <span
      key="dot-0"
      data-testid="milestone-dot"
      className={`w-2 h-2 rounded-full shrink-0 ${dotColor(summary[0].status)}`}
    />,
  )

  // Middle dots with connecting lines
  for (let i = 1; i < summary.length; i++) {
    elements.push(
      <span
        key={`line-${i}`}
        data-testid="milestone-line"
        className="flex-1 h-0.5 bg-border min-w-4"
      />,
    )
    elements.push(
      <span
        key={`dot-${i}`}
        data-testid="milestone-dot"
        className={`w-2 h-2 rounded-full shrink-0 ${dotColor(summary[i].status)}`}
      />,
    )
  }

  // Last milestone name (only when > 1 milestone)
  if (summary.length > 1) {
    elements.push(
      <span
        key="last-name"
        className="text-xs text-tertiary truncate max-w-[60px] shrink-0"
      >
        {summary[summary.length - 1].name}
      </span>,
    )
  }

  return elements
}
