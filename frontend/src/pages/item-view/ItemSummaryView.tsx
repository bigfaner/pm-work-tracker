import { Link } from "react-router-dom";
import { Pencil, Plus } from "lucide-react";
import { MainItem, SubItem } from "@/types";
import { Button } from "@/components/ui/button";
import PriorityBadge from "@/components/shared/PriorityBadge";
import ProgressBar from "@/components/shared/ProgressBar";
import StatusTransitionDropdown from "@/components/shared/StatusTransitionDropdown";
import { Badge } from "@/components/ui/badge";
import { PermissionGuard } from "@/components/PermissionGuard";
import { MAIN_ITEM_STATUSES, SUB_ITEM_STATUSES } from "@/lib/status";
import { isOverdue } from "@/lib/status";
import { showToast } from "@/lib/toast";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

function copyLink(path: string, title: string) {
  navigator.clipboard.writeText(`${window.location.origin}${path} ${title}`);
  showToast("链接已复制", "success");
}

function CodeBadge({
  label,
  path,
  title,
  className,
}: {
  label: string;
  path: string;
  title: string;
  className?: string;
}) {
  return (
    <TooltipProvider delayDuration={300}>
      <Tooltip>
        <TooltipTrigger asChild>
          <span
            className={`font-mono cursor-pointer transition-colors ${className ?? ""}`}
            onClick={(e) => {
              e.stopPropagation();
              copyLink(path, title);
            }}
          >
            {label}
          </span>
        </TooltipTrigger>
        <TooltipContent>
          <p>点击复制链接</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

interface SummaryViewProps {
  items: (MainItem & { subItems?: SubItem[] })[];
  expandedCards: Set<string>;
  onToggleExpand: (id: string) => void;
  subItemsMap: Record<string, SubItem[]>;
  memberName: (id: string | null) => string;
  formatDate: (date: string | null) => string;
  hasMore: boolean;
  sentinelRef: React.RefObject<HTMLDivElement>;
  teamId: string;
  onRefresh: () => void;
  onAddSubItem: (mainItemId: string, mainItemTitle: string) => void;
  onEditMainItem: (item: MainItem) => void;
  onAppendProgress: (
    subItemId: string,
    subItemTitle: string,
    subItemCompletion: number,
  ) => void;
  onEditSubItem: (sub: SubItem, mainItemBizKey: string) => void;
}

export default function ItemSummaryView({
  items,
  expandedCards,
  onToggleExpand,
  subItemsMap,
  memberName,
  formatDate,
  hasMore,
  sentinelRef,
  teamId,
  onRefresh,
  onAddSubItem,
  onEditMainItem,
  onAppendProgress,
  onEditSubItem,
}: SummaryViewProps) {
  return (
    <div>
      {items.map((item) => (
        <div key={item.bizKey} className="mb-3">
          <div
            className="rounded-xl border border-border bg-white shadow-sm cursor-pointer"
            onClick={() => onToggleExpand(item.bizKey)}
          >
            <div className="flex items-center gap-2 px-4 py-3">
              {/* Expand chevron */}
              <svg
                className={`w-3.5 h-3.5 shrink-0 text-tertiary transition-transform ${
                  expandedCards.has(item.bizKey) ? "rotate-90" : ""
                }`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M9 5l7 7-7 7"
                />
              </svg>

              {/* Code */}
              <CodeBadge
                label={item.code}
                path={`/items/${item.bizKey}`}
                title={item.title}
                className="text-xs text-tertiary bg-bg-alt px-1.5 py-0.5 rounded hover:bg-primary-100 hover:text-primary-600"
              />

              {/* Priority */}
              <PriorityBadge priority={item.priority} />

              {/* Title + date range */}
              <div className="flex items-center gap-1.5 flex-1 min-w-0">
                <Link
                  to={`/items/${item.bizKey}`}
                  className="text-sm font-medium text-primary-600 hover:text-primary-700 hover:underline truncate"
                  title={item.title}
                  onClick={(e) => e.stopPropagation()}
                >
                  {item.title}
                </Link>
                {item.planStartDate && item.expectedEndDate && (
                  <span className="text-xs text-secondary whitespace-nowrap">
                    计划周期 {formatDate(item.planStartDate)} ~{" "}
                    {formatDate(item.expectedEndDate)}
                  </span>
                )}
                {isOverdue(
                  item.expectedEndDate ?? undefined,
                  item.itemStatus,
                  new Date(),
                ) && <Badge variant="error">延期</Badge>}
                {MAIN_ITEM_STATUSES[
                  item.itemStatus as keyof typeof MAIN_ITEM_STATUSES
                ]?.terminal &&
                  item.actualEndDate && (
                    <span className="text-xs text-tertiary whitespace-nowrap">
                      结束于 {formatDate(item.actualEndDate)}
                    </span>
                  )}
              </div>

              {/* Assignee */}
              <span className="text-[13px] text-secondary whitespace-nowrap">
                {memberName(item.assigneeKey)}
              </span>

              {/* Progress */}
              <div className="w-16 shrink-0">
                <ProgressBar value={item.completion} size="sm" showPercentage />
              </div>

              {/* Status */}
              <div onClick={(e) => e.stopPropagation()}>
                <StatusTransitionDropdown
                  currentStatus={item.itemStatus}
                  itemType="main"
                  teamId={teamId}
                  itemId={item.bizKey}
                  onStatusChanged={onRefresh}
                />
              </div>

              {/* Actions */}
              <div
                className="flex gap-0.5"
                onClick={(e) => e.stopPropagation()}
              >
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-primary-600"
                  disabled={
                    !!MAIN_ITEM_STATUSES[
                      item.itemStatus as keyof typeof MAIN_ITEM_STATUSES
                    ]?.terminal
                  }
                  onClick={() => onEditMainItem(item)}
                >
                  <Pencil size={14} />
                  编辑
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-primary-600"
                  disabled={
                    !!MAIN_ITEM_STATUSES[
                      item.itemStatus as keyof typeof MAIN_ITEM_STATUSES
                    ]?.terminal
                  }
                  onClick={() => onAddSubItem(item.bizKey, item.title)}
                >
                  <Plus size={14} />
                  新增子事项
                </Button>
              </div>
            </div>

            {/* Expanded sub-items */}
            {expandedCards.has(item.bizKey) && (
              <div className="border-t border-border px-5 py-3 pl-12">
                {!subItemsMap[item.bizKey] && (
                  <div className="text-xs text-tertiary py-2">加载中...</div>
                )}
                {subItemsMap[item.bizKey]?.length === 0 && (
                  <div className="text-xs text-tertiary py-2">暂无子事项</div>
                )}
                {subItemsMap[item.bizKey]?.map((sub) => (
                  <div
                    key={sub.bizKey}
                    className="flex items-center gap-2 py-2 border-b border-border/50 last:border-b-0"
                  >
                    <CodeBadge
                      label={sub.code.split("-").pop()!}
                      path={`/items/${item.bizKey}/sub/${sub.bizKey}`}
                      title={sub.title}
                      className="text-[11px] text-tertiary bg-bg-alt px-1.5 py-0.5 rounded hover:bg-primary-100 hover:text-primary-600"
                    />
                    <PriorityBadge
                      priority={sub.priority}
                      className="text-[10px]"
                    />
                    <Link
                      to={`/items/${item.bizKey}/sub/${sub.bizKey}`}
                      className="text-[13px] font-medium text-primary-600 hover:text-primary-700 hover:underline truncate"
                    >
                      {sub.title}
                    </Link>
                    <span className="text-[11px] text-tertiary whitespace-nowrap">
                      {sub.planStartDate && sub.expectedEndDate
                        ? `计划周期 ${formatDate(sub.planStartDate)} ~ ${formatDate(sub.expectedEndDate)}`
                        : "-"}
                    </span>
                    {isOverdue(
                      sub.expectedEndDate ?? undefined,
                      sub.itemStatus,
                      new Date(),
                    ) && <Badge variant="error">延期</Badge>}
                    {SUB_ITEM_STATUSES[
                      sub.itemStatus as keyof typeof SUB_ITEM_STATUSES
                    ]?.terminal &&
                      sub.actualEndDate && (
                        <span className="text-[11px] text-tertiary whitespace-nowrap">
                          结束于 {formatDate(sub.actualEndDate)}
                        </span>
                      )}
                    <span className="ml-auto text-[13px] text-secondary">
                      {memberName(sub.assigneeKey)}
                    </span>
                    <div className="w-16 shrink-0">
                      <ProgressBar
                        value={sub.completion}
                        size="sm"
                        showPercentage
                      />
                    </div>
                    <div onClick={(e) => e.stopPropagation()}>
                      <StatusTransitionDropdown
                        currentStatus={sub.itemStatus}
                        itemType="sub"
                        teamId={teamId}
                        itemId={sub.bizKey}
                        parentItemId={item.bizKey}
                        onStatusChanged={onRefresh}
                      />
                    </div>
                    <div
                      className="flex gap-0.5"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <PermissionGuard code="sub_item:update">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-[11px] h-6 px-1.5 text-primary-600"
                          data-testid={`edit-sub-${sub.bizKey}`}
                          disabled={
                            !!SUB_ITEM_STATUSES[
                              sub.itemStatus as keyof typeof SUB_ITEM_STATUSES
                            ]?.terminal
                          }
                          onClick={() => onEditSubItem(sub, item.bizKey)}
                        >
                          <Pencil size={12} />
                          编辑
                        </Button>
                      </PermissionGuard>
                      <PermissionGuard code="progress:update">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-[11px] h-6 px-1.5 text-primary-600"
                          disabled={
                            !!SUB_ITEM_STATUSES[
                              sub.itemStatus as keyof typeof SUB_ITEM_STATUSES
                            ]?.terminal
                          }
                          onClick={() =>
                            onAppendProgress(
                              sub.bizKey,
                              sub.title,
                              sub.completion,
                            )
                          }
                        >
                          <Plus size={12} />
                          追加进度
                        </Button>
                      </PermissionGuard>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
          {/* Expand button for test targeting */}
          <button
            data-testid={`expand-card-${item.bizKey}`}
            className="hidden"
            onClick={() => onToggleExpand(item.bizKey)}
          />
        </div>
      ))}

      {/* Infinite scroll sentinel */}
      <div ref={sentinelRef} className="py-4 text-center">
        {hasMore ? (
          <span className="text-xs text-tertiary">加载中...</span>
        ) : (
          items.length > 0 && (
            <span className="text-xs text-tertiary">-- 没有更多了 --</span>
          )
        )}
      </div>
    </div>
  );
}
