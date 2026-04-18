import { useState, useMemo, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useTeamStore } from '@/store/team'
import { getWeeklyViewApi } from '@/api/views'
import type { WeeklyViewResponse, WeeklyComparisonGroup, SubItemSnapshot } from '@/types'
import PriorityBadge from '@/components/shared/PriorityBadge'
import StatusBadge from '@/components/shared/StatusBadge'
import ProgressBar from '@/components/shared/ProgressBar'

// --- Helpers ---

function formatDate(date: string) {
  return date.replace(/-/g, '/')
}

function getWeekNumber(dateStr: string): number {
  const d = new Date(dateStr)
  const startOfYear = new Date(d.getFullYear(), 0, 1)
  const diff = d.getTime() - startOfYear.getTime()
  const oneWeek = 604800000
  return Math.ceil((diff / oneWeek) + ((startOfYear.getDay() + 1) / 7))
}

function getCurrentWeekStart(): string {
  const now = new Date()
  const day = now.getDay()
  const diff = now.getDate() - day + (day === 0 ? -6 : 1)
  const monday = new Date(now)
  monday.setDate(diff)
  return monday.toISOString().slice(0, 10)
}

function toWeekInputValue(weekStart: string): string {
  const d = new Date(weekStart)
  const year = d.getFullYear()
  const jan1 = new Date(year, 0, 1)
  const days = Math.floor((d.getTime() - jan1.getTime()) / 86400000)
  const weekNum = Math.ceil((days + jan1.getDay() + 1) / 7)
  return `${year}-W${String(weekNum).padStart(2, '0')}`
}

// --- Main Component ---

export default function WeeklyViewPage() {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const [weekStart, setWeekStart] = useState(getCurrentWeekStart)

  const maxWeek = useMemo(() => toWeekInputValue(getCurrentWeekStart()), [])

  const { data, isLoading } = useQuery({
    queryKey: ['weeklyView', teamId, weekStart],
    queryFn: () => getWeeklyViewApi(teamId!, weekStart),
    enabled: !!teamId,
  })

  const handleWeekChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value
    if (!val) return
    // Convert YYYY-Www to YYYY-MM-DD (Monday of that week)
    const [yearStr, weekStr] = val.split('-W')
    const year = parseInt(yearStr)
    const week = parseInt(weekStr)
    const jan1 = new Date(year, 0, 1)
    const dayOfWeek = jan1.getDay()
    const offset = dayOfWeek <= 4 ? 1 - dayOfWeek : 8 - dayOfWeek
    const monday = new Date(year, 0, offset + (week - 1) * 7)
    setWeekStart(monday.toISOString().slice(0, 10))
  }, [])

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
              <input
                type="week"
                data-testid="week-selector"
                className="h-8 rounded-md border border-border bg-white px-2 text-sm"
                value={toWeekInputValue(weekStart)}
                onChange={handleWeekChange}
                max={maxWeek}
                style={{ width: 180 }}
              />
              {data && (
                <span className="text-[13px] text-secondary">
                  {formatDate(data.weekStart)} ~ {formatDate(data.weekEnd)}
                </span>
              )}
            </div>
          </div>

          {isLoading ? (
            <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
          ) : !data || data.groups.length === 0 ? (
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

              {/* Legend */}
              <Legend />
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
        <div className="text-[13px] text-tertiary">本周活跃子事项</div>
      </div>
      <div className="rounded-xl border border-border bg-white p-4 text-center">
        <div className="text-2xl font-semibold text-emerald-600" data-testid="stat-newly-completed">
          {stats.newlyCompleted}
        </div>
        <div className="text-[13px] text-tertiary">本周新完成</div>
      </div>
      <div className="rounded-xl border border-border bg-white p-4 text-center">
        <div className="text-2xl font-semibold text-primary-600" data-testid="stat-in-progress">
          {stats.inProgress}
        </div>
        <div className="text-[13px] text-tertiary">进度推进中</div>
      </div>
      <div className="rounded-xl border border-border bg-white p-4 text-center">
        <div className="text-2xl font-semibold text-red-600" data-testid="stat-blocked">
          {stats.blocked}
        </div>
        <div className="text-[13px] text-tertiary">阻塞中</div>
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
  const { mainItem, lastWeek, thisWeek, completedNoChange } = group

  const weekNum = getWeekNumber(weekStart)
  const hasCompleted = completedNoChange.length > 0

  return (
    <div className="rounded-xl border border-border bg-white shadow-sm mb-5" data-testid={`group-card-${mainItem.id}`}>
      {/* Card Header */}
      <div className="flex items-center justify-between flex-wrap gap-2 px-5 py-3 bg-bg-alt rounded-t-xl border-b border-border">
        <div className="flex items-center gap-2 flex-wrap">
          <Link
            to={`/items/${mainItem.id}`}
            className="text-[15px] font-semibold text-primary hover:text-primary-600"
          >
            {mainItem.title}
          </Link>
          <PriorityBadge priority={mainItem.priority} className="text-[11px]" />
          <span className="text-xs text-tertiary whitespace-nowrap">
            计划 {formatDate(mainItem.startDate)}~{formatDate(mainItem.expectedEndDate)}
          </span>
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
          <div className="w-[120px]">
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
              <SubItemRow key={item.id} item={item} />
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
                  <span className="text-[13px]">{item.title}</span>
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
              <SubItemRow key={item.id} item={item} showDelta />
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
                  <span className="text-[13px]">{item.title}</span>
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
  showDelta?: boolean
}

function SubItemRow({ item, showDelta }: SubItemRowProps) {
  return (
    <div>
      <div className="flex items-center gap-1.5 flex-wrap">
        <StatusBadge status={item.status} className="text-[11px]" />
        <PriorityBadge priority={item.priority} className="text-[10px]" />
        <span className="text-[13px]">{item.title}</span>
        <span className="text-[11px] text-tertiary whitespace-nowrap">
          {item.assigneeName}
        </span>
        <span className="text-[11px] text-tertiary whitespace-nowrap">
          计划 {formatDate(item.expectedEndDate)}
        </span>
        {showDelta && item.delta != null && item.delta > 0 && !item.justCompleted && (
          <span className="text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-1.5 py-px rounded">
            +{item.delta}%
          </span>
        )}
        {showDelta && item.justCompleted && (
          <span className="text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-1.5 py-px rounded">
            完成 ✓
          </span>
        )}
        {showDelta && item.isNew && (
          <span className="text-[11px] font-semibold text-amber-700 bg-amber-50 px-1.5 py-px rounded">
            NEW
          </span>
        )}
      </div>
      {item.progressDescription && (
        <div className="pl-14 text-xs text-tertiary -mt-1">
          {item.completion}% · {item.progressDescription}
        </div>
      )}
    </div>
  )
}

// --- Legend ---

function Legend() {
  return (
    <div className="flex items-center gap-4 py-3 text-xs text-tertiary flex-wrap">
      <span>图例：</span>
      <span className="flex items-center gap-1">
        <PriorityBadge priority="P1" className="text-[10px]" />
        <PriorityBadge priority="P2" className="text-[10px]" />
        <PriorityBadge priority="P3" className="text-[10px]" />
        优先级
      </span>
      <span className="flex items-center gap-1">
        <span className="text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-1.5 py-px rounded">+30%</span>
        本周进度增量
      </span>
      <span className="flex items-center gap-1">
        <span className="text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-1.5 py-px rounded">完成 ✓</span>
        本周新完成
      </span>
      <span className="flex items-center gap-1">
        <span className="text-[11px] font-semibold text-amber-700 bg-amber-50 px-1.5 py-px rounded">NEW</span>
        本周新增事项
      </span>
    </div>
  )
}
