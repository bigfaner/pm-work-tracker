import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { getWeeklyViewApi } from '@/api/views'
import { WeeklyViewResponse, WeeklyComparisonGroup, SubItemSnapshot } from '@/types'
import PriorityBadge from '@/components/shared/PriorityBadge'
import StatusBadge from '@/components/shared/StatusBadge'
import ProgressBar from '@/components/shared/ProgressBar'
import { WeekPicker } from '@/components/shared/WeekPicker'
import { RotateCcw } from 'lucide-react'
import { getCurrentWeekStart, getWeekNumber } from '@/utils/weekUtils'
import { isOverdue } from '@/lib/status'
import { formatDate } from '@/lib/format'
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from '@/components/ui/tooltip'

// --- Main Component ---

export default function WeeklyViewPage() {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const [weekStart, setWeekStart] = useState(getCurrentWeekStart)

  const { data, isLoading } = useQuery({
    queryKey: ['weeklyView', teamId, weekStart],
    queryFn: () => getWeeklyViewApi(teamId!, weekStart),
    enabled: !!teamId,
  })

  return (
    <div data-testid="weekly-view-page">
      {!teamId && <div className="p-6 text-tertiary">请先选择团队</div>}
      {teamId && (
        <>
          {/* Page Header */}
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-xl font-semibold text-primary">每周进展</h1>
            <div className="flex items-center gap-2">
              <span className="text-[13px] text-secondary">选择周次：</span>
              <WeekPicker
                weekStart={weekStart}
                onChange={setWeekStart}
                maxWeek={getCurrentWeekStart()}
              />
              <button
                onClick={() => setWeekStart(getCurrentWeekStart())}
                title="回到本周"
                className="h-8 w-8 inline-flex items-center justify-center rounded-md border border-border bg-white text-secondary hover:text-primary-500 hover:border-primary-500 transition-colors"
              >
                <RotateCcw className="h-4 w-4" />
              </button>
            </div>
          </div>

          {isLoading ? (
            <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
          ) : !data || !data.groups || data.groups.length === 0 ? (
            <div className="py-12 text-center">
              <p className="text-tertiary text-sm">暂无周数据</p>
            </div>
          ) : (
            <>
              {/* Stats Row */}
              <StatsBar stats={data.stats} />

              {/* Comparison Cards */}
              {data.groups.map((group) => (
                <ComparisonCard key={group.mainItem.id} group={group} weekStart={weekStart} />
              ))}

            </>
          )}
        </>
      )}
    </div>
  )
}

// --- Stats Bar ---

interface StatsBarProps {
  stats: WeeklyViewResponse['stats']
}

function StatsBar({ stats }: StatsBarProps) {
  return (
    <div className="grid grid-cols-4 gap-4 mb-6">
      <div className="rounded-xl border border-border bg-white p-4 text-center">
        <div className="text-2xl font-semibold" data-testid="stat-active">
          {stats.activeSubItems}
        </div>
        <div className="text-[13px] text-tertiary mt-0.5">本周活跃子事项</div>
        <div className="text-[11px] text-tertiary/70 mt-1">本周有进度记录或新完成</div>
      </div>
      <div className="rounded-xl border border-border bg-white p-4 text-center">
        <div className="text-2xl font-semibold text-success-text" data-testid="stat-newly-completed">
          {stats.newlyCompleted}
        </div>
        <div className="text-[13px] text-tertiary mt-0.5">本周新完成</div>
        <div className="text-[11px] text-tertiary/70 mt-1">完成时间落在本周内</div>
      </div>
      <div className="rounded-xl border border-border bg-white p-4 text-center">
        <div className="text-2xl font-semibold text-primary-600" data-testid="stat-in-progress">
          {stats.inProgress}
        </div>
        <div className="text-[13px] text-tertiary mt-0.5">进度推进中</div>
        <div className="text-[11px] text-tertiary/70 mt-1">活跃子事项中状态为进行中</div>
      </div>
      <div className="rounded-xl border border-border bg-white p-4 text-center">
        <div className="text-2xl font-semibold text-error" data-testid="stat-blocked">
          {stats.blocked}
        </div>
        <div className="text-[13px] text-tertiary mt-0.5">阻塞中</div>
        <div className="text-[11px] text-tertiary/70 mt-1">活跃子事项中状态为阻塞</div>
      </div>
    </div>
  )
}

// --- Comparison Card ---

interface ComparisonCardProps {
  group: WeeklyComparisonGroup
  weekStart: string
}

function ComparisonCard({ group, weekStart }: ComparisonCardProps) {
  const [expanded, setExpanded] = useState(false)
  const { mainItem } = group
  const lastWeek = group.lastWeek ?? []
  const thisWeek = group.thisWeek ?? []
  const completedNoChange = group.completedNoChange ?? []

  const weekNum = getWeekNumber(weekStart)
  const hasCompleted = completedNoChange.length > 0

  return (
    <div className="rounded-xl border border-border bg-white shadow-sm mb-5" data-testid={`group-card-${mainItem.id}`}>
      {/* Card Header */}
      <div className="flex items-center justify-between flex-wrap gap-2 px-5 py-3 bg-bg-alt rounded-t-xl border-b border-border">
        <div className="flex items-center gap-2 flex-wrap">
          <Link
            to={`/items/${mainItem.id}`}
            className="text-[15px] font-semibold text-primary-600 hover:text-primary-700 hover:underline"
          >
            <span className="text-tertiary font-normal mr-1">{mainItem.code}</span>{mainItem.title}
          </Link>
          <PriorityBadge priority={mainItem.priority} className="text-[11px]" />
          <span className="text-xs text-tertiary whitespace-nowrap">
            计划周期 {formatDate(mainItem.startDate)}~{formatDate(mainItem.expectedEndDate)}
          </span>
          {isOverdue(mainItem.expectedEndDate, mainItem.status) && (
            <span className="inline-flex items-center rounded-md bg-error-bg px-2 py-0.5 text-[11px] font-medium text-error-text">
              延期
            </span>
          )}
          {(mainItem.status === 'completed' || mainItem.status === 'closed') && mainItem.actualEndDate && (
            <span className="text-xs text-tertiary whitespace-nowrap">
              结束于 {formatDate(mainItem.actualEndDate)}
            </span>
          )}
          <span className="inline-flex items-center rounded-md bg-primary-50 px-2 py-0.5 text-[11px] font-medium text-primary-700">
            {mainItem.subItemCount} 个子事项
          </span>
          {hasCompleted && (
            <button
              data-testid={`expand-completed-${mainItem.id}`}
              className="p-1 rounded hover:bg-bg-alt"
              onClick={() => setExpanded((v) => !v)}
              title={expanded ? '折叠已完成、无变化的子事项' : '已完成、无变化的子事项已被折叠，点击展开'}
            >
              <svg
                className={`w-4 h-4 text-tertiary transition-transform ${expanded ? 'rotate-180' : ''}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 15l7-7 7 7" />
              </svg>
            </button>
          )}
        </div>
        <div className="flex items-center gap-3">
          <div className="w-30">
            <ProgressBar
              value={mainItem.completion}
              size="sm"
              showPercentage
            />
          </div>
        </div>
      </div>

      {/* Two-column layout */}
      <div className="grid grid-cols-[1fr_1px_1fr]">
        {/* Left: Last Week */}
        <div className="p-4 px-5 bg-[#fafbfc]">
          <div className="text-xs font-semibold text-tertiary uppercase tracking-wider mb-3">
            上周 (W{weekNum - 1})
          </div>
          <div className="flex flex-col gap-2.5">
            {lastWeek.map((item) => (
              <SubItemRow key={item.id} item={item} mainItemId={mainItem.id} />
            ))}
            {lastWeek.length === 0 && (
              <div className="text-xs text-tertiary">无活跃事项</div>
            )}
          </div>
          {/* Collapsed: completed-no-change */}
          {hasCompleted && expanded && (
            <div className="mt-2.5 pt-2.5 border-t border-dashed border-border">
              <div className="text-xs text-tertiary mb-1.5">已完成无变化</div>
              {completedNoChange.map((item) => (
                <div key={item.id} className="flex items-center gap-1.5 flex-wrap opacity-70">
                  <StatusBadge status={item.status} className="text-[11px]" />
                  <PriorityBadge priority={item.priority} className="text-[10px]" />
                  <TooltipProvider delayDuration={300}>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Link to={`/items/${mainItem.id}/sub/${item.id}`} className="text-[13px] text-primary-600 hover:text-primary-700 hover:underline truncate max-w-[160px]">{item.title}</Link>
                      </TooltipTrigger>
                      <TooltipContent>{item.title}</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Divider */}
        <div className="bg-border" />

        {/* Right: This Week */}
        <div className="p-4 px-5">
          <div className="text-xs font-semibold text-primary-700 uppercase tracking-wider mb-3">
            本周 (W{weekNum})
          </div>
          <div className="flex flex-col gap-2.5">
            {thisWeek.map((item) => (
              <SubItemRow key={item.id} item={item} mainItemId={mainItem.id} showDelta />
            ))}
            {thisWeek.length === 0 && (
              <div className="text-xs text-tertiary">无活跃事项</div>
            )}
          </div>
          {/* Collapsed: completed-no-change */}
          {hasCompleted && expanded && (
            <div className="mt-2.5 pt-2.5 border-t border-dashed border-border">
              <div className="text-xs text-tertiary mb-1.5">已完成无变化</div>
              {completedNoChange.map((item) => (
                <div key={item.id} className="flex items-center gap-1.5 flex-wrap opacity-70">
                  <StatusBadge status={item.status} className="text-[11px]" />
                  <PriorityBadge priority={item.priority} className="text-[10px]" />
                  <TooltipProvider delayDuration={300}>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Link to={`/items/${mainItem.id}/sub/${item.id}`} className="text-[13px] text-primary-600 hover:text-primary-700 hover:underline truncate max-w-[160px]">{item.title}</Link>
                      </TooltipTrigger>
                      <TooltipContent>{item.title}</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// --- Sub-Item Row ---

interface SubItemRowProps {
  item: SubItemSnapshot
  mainItemId: number
  showDelta?: boolean
}

function SubItemRow({ item, mainItemId, showDelta }: SubItemRowProps) {
  const overdue = isOverdue(item.expectedEndDate, item.status)
  const periodText = item.startDate && item.expectedEndDate
    ? `计划周期 ${formatDate(item.startDate)}~${formatDate(item.expectedEndDate)}`
    : item.expectedEndDate ? `计划 ${formatDate(item.expectedEndDate)}` : ''

  return (
    <div className="py-1">
      <div className="flex items-center gap-1.5 flex-wrap">
        <StatusBadge status={item.status} className="text-[11px]" />
        <PriorityBadge priority={item.priority} className="text-[10px]" />
        <TooltipProvider delayDuration={300}>
          <Tooltip>
            <TooltipTrigger asChild>
              <Link
                to={`/items/${mainItemId}/sub/${item.id}`}
                className="text-[13px] text-primary-600 hover:text-primary-700 hover:underline truncate max-w-[160px]"
              >
                {item.title}
              </Link>
            </TooltipTrigger>
            <TooltipContent className="flex items-center gap-1.5 whitespace-nowrap">
              <span className="text-white/70 font-mono">{item.code}</span>
              <span>{item.title}</span>
              {periodText && <span className="text-white/70">{periodText}</span>}
              {overdue && (
                <span className="inline-flex items-center rounded bg-error-bg px-1.5 py-px text-[11px] font-medium text-error-text">
                  延期
                </span>
              )}
              {(item.status === 'completed' || item.status === 'closed') && item.actualEndDate && (
                <span className="text-white/70">结束于 {formatDate(item.actualEndDate)}</span>
              )}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <span className={`text-[11px] font-semibold ${item.completion === 100 ? 'text-success-text' : 'text-secondary'}`}>
          {item.completion}%
        </span>
        <span className="text-[11px] text-tertiary whitespace-nowrap">
          {item.assigneeName}
        </span>
        {showDelta && item.delta != null && item.delta > 0 && !item.justCompleted && (
          <span className="text-[11px] font-semibold text-success-text bg-success-bg px-1.5 py-px rounded">
            +{item.delta}%
          </span>
        )}
        {showDelta && item.justCompleted && (
          <span className="text-[11px] font-semibold text-success-text bg-success-bg px-1.5 py-px rounded">
            完成 ✓
          </span>
        )}
        {showDelta && item.isNew && (
          <span className="text-[11px] font-semibold text-warning-text bg-warning-bg px-1.5 py-px rounded">
            NEW
          </span>
        )}
      </div>
      {item.progressRecords && item.progressRecords.length > 0 ? (
        <div className="pl-14 mt-1.5 flex flex-col gap-1">
          {item.progressRecords.map((record) => (
            <div key={record.id} className="text-xs text-tertiary">
              {record.achievement && (
                <span className="text-success-text">成果：{record.achievement}</span>
              )}
              {record.blocker && (
                <span className="ml-2 text-error">卡点：{record.blocker}</span>
              )}
            </div>
          ))}
        </div>
      ) : item.progressDescription ? (
        <div className="pl-14 text-xs text-tertiary mt-1 py-1">
          {item.completion}% · {item.progressDescription}
        </div>
      ) : null}
    </div>
  )
}
