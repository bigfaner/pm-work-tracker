import { useState } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTeamStore } from "@/store/team";
import { getWeeklyViewApi } from "@/api/views";
import {
  WeeklyViewResponse,
  WeeklyComparisonGroup,
  SubItemSnapshot,
} from "@/types";
import PriorityBadge from "@/components/shared/PriorityBadge";
import StatusBadge from "@/components/shared/StatusBadge";
import ProgressBar from "@/components/shared/ProgressBar";
import { WeekPicker } from "@/components/shared/WeekPicker";
import { RotateCcw } from "lucide-react";
import { getCurrentWeekStart, getWeekNumber } from "@/utils/weekUtils";
import { isOverdue } from "@/lib/status";
import { formatDate } from "@/lib/format";
import {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
  TooltipProvider,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

// --- Main Component ---

export default function WeeklyViewPage() {
  const teamId = useTeamStore((s) => s.currentTeamId);
  const [weekStart, setWeekStart] = useState(getCurrentWeekStart);

  const { data, isLoading, isError } = useQuery({
    queryKey: ["weeklyView", teamId, weekStart],
    queryFn: () => getWeeklyViewApi(teamId!, weekStart),
    enabled: !!teamId,
  });

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
            <div className="py-8 text-center text-tertiary text-sm">
              加载中...
            </div>
          ) : (
            <>
              {/* Stats Row — always visible (shows "-" on error) */}
              <StatsBar stats={isError ? undefined : data?.stats} />

              {/* Comparison Cards */}
              {!data || !data.groups || data.groups.length === 0 ? (
                <div className="py-12 text-center">
                  <p className="text-tertiary text-sm">暂无周数据</p>
                </div>
              ) : (
                <>
                  {(() => {
                    if (!data.weekEnd) throw new Error("weekEnd is required");
                    const referenceDate = new Date(data.weekEnd);
                    return data.groups.map((group) => (
                      <ComparisonCard
                        key={group.mainItem.bizKey}
                        group={group}
                        weekStart={weekStart}
                        referenceDate={referenceDate}
                      />
                    ));
                  })()}
                </>
              )}
            </>
          )}
        </>
      )}
    </div>
  );
}

// --- Stats Bar ---

interface StatsBarProps {
  stats: WeeklyViewResponse["stats"] | undefined;
}

function StatsBar({ stats }: StatsBarProps) {
  const v = (n: number | undefined) => (stats === undefined ? "-" : (n ?? 0));
  return (
    <TooltipProvider delayDuration={300}>
      <div className="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-7 gap-3 mb-6">
        <StatCard
          value={v(stats?.activeSubItems)}
          label="本周活跃"
          tooltip="本周有进展记录，或计划周期与本周重叠的子事项总数（含本周新完成）"
          testId="stat-active"
        />
        <StatCard
          value={v(stats?.newlyCompleted)}
          label="本周新完成"
          tooltip="本周内实际完成（actualEndDate 落在本周）的子事项数"
          valueClassName="text-success-text"
          testId="stat-newly-completed"
        />
        <StatCard
          value={v(stats?.inProgress)}
          label="进行中"
          tooltip={'状态为"进行中"且本周活跃的子事项数'}
          valueClassName="text-primary-600"
          testId="stat-in-progress"
        />
        <StatCard
          value={v(stats?.blocked)}
          label="阻塞中"
          tooltip={'状态为"阻塞中"且本周活跃的子事项数'}
          valueClassName="text-error"
          testId="stat-blocked"
        />
        <StatCard
          value={v(stats?.pending)}
          label="未开始"
          tooltip="已创建但尚未启动（状态为 pending）且本周活跃的子事项数"
          valueClassName="text-secondary"
          testId="stat-pending"
        />
        <StatCard
          value={v(stats?.pausing)}
          label="暂停中"
          tooltip={'状态为"暂停中"且本周活跃的子事项数'}
          valueClassName="text-warning"
          testId="stat-pausing"
        />
        <StatCard
          value={v(stats?.overdue)}
          label="逾期中"
          tooltip="计划截止日在本周结束前已过、尚未完成/关闭且本周活跃的子事项数"
          valueClassName="text-error"
          testId="stat-overdue"
        />
      </div>
    </TooltipProvider>
  );
}

// --- Stat Card ---

interface StatCardProps {
  value: number | string;
  label: string;
  tooltip: string;
  valueClassName?: string;
  testId?: string;
}

function StatCard({
  value,
  label,
  tooltip,
  valueClassName,
  testId,
}: StatCardProps) {
  const [open, setOpen] = useState(false);
  return (
    <Tooltip open={open} onOpenChange={setOpen}>
      <TooltipTrigger asChild>
        <button
          className="rounded-xl border border-border bg-white p-4 text-center w-full"
          onClick={() => setOpen((v) => !v)}
        >
          <div
            className={cn("text-2xl font-semibold", valueClassName)}
            data-testid={testId}
          >
            {value}
          </div>
          <div className="text-[13px] text-tertiary mt-0.5">{label}</div>
        </button>
      </TooltipTrigger>
      <TooltipContent side="top">
        <p className="max-w-[200px] text-center">{tooltip}</p>
      </TooltipContent>
    </Tooltip>
  );
}

// --- Comparison Card ---

interface ComparisonCardProps {
  group: WeeklyComparisonGroup;
  weekStart: string;
  referenceDate: Date;
}

function ComparisonCard({
  group,
  weekStart,
  referenceDate,
}: ComparisonCardProps) {
  const [expanded, setExpanded] = useState(false);
  const { mainItem } = group;
  const lastWeek = group.lastWeek ?? [];
  const thisWeek = group.thisWeek ?? [];
  const completedNoChange = group.completedNoChange ?? [];

  const weekNum = getWeekNumber(weekStart);
  const hasCompleted = completedNoChange.length > 0;

  return (
    <div
      className="rounded-xl border border-border bg-white shadow-sm mb-5"
      data-testid={`group-card-${mainItem.bizKey}`}
    >
      {/* Card Header */}
      <div className="flex items-center justify-between flex-wrap gap-2 px-5 py-3 bg-bg-alt rounded-t-xl border-b border-border">
        <div className="flex items-center gap-2 flex-wrap">
          <Link
            to={`/items/${mainItem.bizKey}`}
            className="text-[15px] font-semibold text-primary-600 hover:text-primary-700 hover:underline"
          >
            <span className="text-tertiary font-normal mr-1">
              {mainItem.code}
            </span>
            {mainItem.title}
          </Link>
          <PriorityBadge priority={mainItem.priority} className="text-[11px]" />
          <span className="text-xs text-tertiary whitespace-nowrap">
            计划周期 {formatDate(mainItem.startDate)}~
            {formatDate(mainItem.expectedEndDate)}
          </span>
          {isOverdue(
            mainItem.expectedEndDate,
            mainItem.itemStatus,
            referenceDate,
          ) && (
            <span className="inline-flex items-center rounded-md bg-error-bg px-2 py-0.5 text-[11px] font-medium text-error-text">
              延期
            </span>
          )}
          {(mainItem.itemStatus === "completed" ||
            mainItem.itemStatus === "closed") &&
            mainItem.actualEndDate && (
              <span className="text-xs text-tertiary whitespace-nowrap">
                结束于 {formatDate(mainItem.actualEndDate)}
              </span>
            )}
          <span className="inline-flex items-center rounded-md bg-primary-50 px-2 py-0.5 text-[11px] font-medium text-primary-700">
            {mainItem.subItemCount} 个子事项
          </span>
          {hasCompleted && (
            <button
              data-testid={`expand-completed-${mainItem.bizKey}`}
              className="p-1 rounded hover:bg-bg-alt"
              onClick={() => setExpanded((v) => !v)}
              title={
                expanded
                  ? "折叠已完成、无变化的子事项"
                  : "已完成、无变化的子事项已被折叠，点击展开"
              }
            >
              <svg
                className={`w-4 h-4 text-tertiary transition-transform ${expanded ? "rotate-180" : ""}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M5 15l7-7 7 7"
                />
              </svg>
            </button>
          )}
        </div>
        <div className="flex items-center gap-3">
          <div className="w-30">
            <ProgressBar value={mainItem.completion} size="sm" showPercentage />
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
              <SubItemRow
                key={item.bizKey}
                item={item}
                mainItemId={mainItem.bizKey}
                referenceDate={referenceDate}
              />
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
                <div
                  key={item.bizKey}
                  className="flex items-center gap-1.5 flex-wrap opacity-70"
                >
                  <StatusBadge
                    status={item.itemStatus}
                    className="text-[11px]"
                  />
                  <PriorityBadge
                    priority={item.priority}
                    className="text-[10px]"
                  />
                  <TooltipProvider delayDuration={300}>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Link
                          to={`/items/${mainItem.bizKey}/sub/${item.bizKey}`}
                          className="text-[13px] text-primary-600 hover:text-primary-700 hover:underline truncate max-w-[160px]"
                        >
                          {item.title}
                        </Link>
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
              <SubItemRow
                key={item.bizKey}
                item={item}
                mainItemId={mainItem.bizKey}
                showDelta
                referenceDate={referenceDate}
              />
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
                <div
                  key={item.bizKey}
                  className="flex items-center gap-1.5 flex-wrap opacity-70"
                >
                  <StatusBadge
                    status={item.itemStatus}
                    className="text-[11px]"
                  />
                  <PriorityBadge
                    priority={item.priority}
                    className="text-[10px]"
                  />
                  <TooltipProvider delayDuration={300}>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Link
                          to={`/items/${mainItem.bizKey}/sub/${item.bizKey}`}
                          className="text-[13px] text-primary-600 hover:text-primary-700 hover:underline truncate max-w-[160px]"
                        >
                          {item.title}
                        </Link>
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
  );
}

// --- Sub-Item Row ---

interface SubItemRowProps {
  item: SubItemSnapshot;
  mainItemId: string;
  showDelta?: boolean;
  referenceDate: Date;
}

function SubItemRow({
  item,
  mainItemId,
  showDelta,
  referenceDate,
}: SubItemRowProps) {
  const overdue = isOverdue(
    item.expectedEndDate,
    item.itemStatus,
    referenceDate,
  );
  const periodText =
    item.startDate && item.expectedEndDate
      ? `计划周期 ${formatDate(item.startDate)}~${formatDate(item.expectedEndDate)}`
      : item.expectedEndDate
        ? `计划 ${formatDate(item.expectedEndDate)}`
        : "";

  return (
    <div className="py-1">
      <div className="flex items-center gap-1.5 flex-wrap">
        <StatusBadge status={item.itemStatus} className="text-[11px]" />
        <PriorityBadge priority={item.priority} className="text-[10px]" />
        <TooltipProvider delayDuration={300}>
          <Tooltip>
            <TooltipTrigger asChild>
              <Link
                to={`/items/${mainItemId}/sub/${item.bizKey}`}
                className="text-[13px] text-primary-600 hover:text-primary-700 hover:underline truncate max-w-[160px]"
              >
                {item.title}
              </Link>
            </TooltipTrigger>
            <TooltipContent className="flex items-center gap-1.5 whitespace-nowrap">
              <span className="text-white/70 font-mono">{item.code}</span>
              <span>{item.title}</span>
              {periodText && (
                <span className="text-white/70">{periodText}</span>
              )}
              {overdue && (
                <span className="inline-flex items-center rounded bg-error-bg px-1.5 py-px text-[11px] font-medium text-error-text">
                  延期
                </span>
              )}
              {(item.itemStatus === "completed" ||
                item.itemStatus === "closed") &&
                item.actualEndDate && (
                  <span className="text-white/70">
                    结束于 {formatDate(item.actualEndDate)}
                  </span>
                )}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <span
          className={`text-[11px] font-semibold ${item.completion === 100 ? "text-success-text" : "text-secondary"}`}
        >
          {Math.round(item.completion)}%
        </span>
        <span className="text-[11px] text-tertiary whitespace-nowrap">
          {item.assigneeName}
        </span>
        {showDelta &&
          item.delta != null &&
          item.delta > 0 &&
          !item.justCompleted && (
            <span className="text-[11px] font-semibold text-success-text bg-success-bg px-1.5 py-px rounded">
              +{Math.round(item.delta)}%
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
          {item.progressRecords.map((record, idx) => (
            <div
              key={`${record.createdAt}-${idx}`}
              className="text-xs text-tertiary"
            >
              {record.achievement && (
                <span className="text-success-text">
                  成果：{record.achievement}
                </span>
              )}
              {record.blocker && (
                <span className="ml-2 text-error">卡点：{record.blocker}</span>
              )}
            </div>
          ))}
        </div>
      ) : item.progressDescription ? (
        <div className="pl-14 text-xs text-tertiary mt-1 py-1">
          {Math.round(item.completion)}% · {item.progressDescription}
        </div>
      ) : null}
    </div>
  );
}
