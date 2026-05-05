import { useState, useCallback, useEffect, useRef } from "react";
import { listDecisionLogsApi } from "@/api/decisionLogs";
import type { DecisionLog } from "@/api/decisionLogs";
import { usePermission } from "@/hooks/usePermission";
import { useAuthStore } from "@/store/auth";
import {
  Card,
  CardHeader,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { MAIN_ITEM_STATUSES } from "@/lib/status";
import { ChevronDown, Pencil, Plus, RefreshCw } from "lucide-react";

// Category label map
const CATEGORY_LABELS: Record<string, string> = {
  technical: "技术",
  resource: "资源",
  requirement: "需求",
  schedule: "进度",
  risk: "风险",
  other: "其他",
};

// Category badge variant map
const CATEGORY_VARIANTS: Record<string, "default" | "warning" | "error"> = {
  technical: "default",
  resource: "default",
  requirement: "default",
  schedule: "warning",
  risk: "error",
  other: "default",
};

// Max tags to show before overflow
const MAX_VISIBLE_TAGS = 3;

function formatTime(iso: string): string {
  const d = new Date(iso);
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  const h = String(d.getHours()).padStart(2, "0");
  const min = String(d.getMinutes()).padStart(2, "0");
  return `${y}-${m}-${day} ${h}:${min}`;
}

interface DecisionTimelineProps {
  teamId: string;
  mainItemId: string;
  mainStatus: string;
  onAdd: () => void;
  onEdit: (bizKey: string) => void;
  collapsed?: boolean;
  onToggleCollapse?: () => void;
  refreshKey?: number;
}

export function DecisionTimeline({
  teamId,
  mainItemId,
  mainStatus,
  onAdd,
  onEdit,
  collapsed = false,
  onToggleCollapse,
  refreshKey,
}: DecisionTimelineProps) {
  const canUpdate = usePermission("main_item:update");
  const currentUserBizKey = useAuthStore((s) => s.user?.bizKey);
  const isTerminal =
    !!MAIN_ITEM_STATUSES[mainStatus as keyof typeof MAIN_ITEM_STATUSES]
      ?.terminal;

  const [logs, setLogs] = useState<DecisionLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState(false);
  const [paginationError, setPaginationError] = useState(false);
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const sentinelRef = useRef<HTMLDivElement>(null);
  const pageSize = 20;

  const fetchLogs = useCallback(
    async (pageNum: number) => {
      try {
        if (pageNum === 1) {
          setLoading(true);
          setError(false);
        } else {
          setLoadingMore(true);
          setPaginationError(false);
        }
        const result = await listDecisionLogsApi(
          teamId,
          mainItemId,
          pageNum,
          pageSize,
        );
        if (pageNum === 1) {
          setLogs(result.items);
        } else {
          setLogs((prev) => [...prev, ...result.items]);
        }
        setTotal(result.total);
        setPage(pageNum);
      } catch {
        if (pageNum === 1) {
          setError(true);
        } else {
          setPaginationError(true);
        }
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [teamId, mainItemId],
  );

  // Initial fetch + refresh on refreshKey change
  useEffect(() => {
    fetchLogs(1);
  }, [fetchLogs, refreshKey]);

  // Infinite scroll via IntersectionObserver
  useEffect(() => {
    if (!sentinelRef.current) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (
          entries[0].isIntersecting &&
          !loading &&
          !loadingMore &&
          logs.length < total
        ) {
          fetchLogs(page + 1);
        }
      },
      { rootMargin: "100px" },
    );
    observer.observe(sentinelRef.current);
    return () => observer.disconnect();
  }, [loading, loadingMore, logs.length, total, page, fetchLogs]);

  const toggleExpand = useCallback((bizKey: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(bizKey)) {
        next.delete(bizKey);
      } else {
        next.add(bizKey);
      }
      return next;
    });
  }, []);

  const showAddButton = canUpdate && !isTerminal;

  // --- Render ---

  const collapsibleHeader = (
    headerRight?: React.ReactNode,
  ) => (
    <CardHeader
      className={onToggleCollapse ? "cursor-pointer select-none" : ""}
      onClick={onToggleCollapse}
    >
      <h3 className="text-sm font-semibold text-primary m-0">决策记录</h3>
      <div className="flex items-center gap-2">
        {headerRight}
        {onToggleCollapse && (
          <ChevronDown
            className={`w-4 h-4 text-tertiary transition-transform ${collapsed ? "" : "rotate-180"}`}
          />
        )}
      </div>
    </CardHeader>
  );

  // Error state (first page)
  if (error) {
    return (
      <Card className="mb-5">
        {collapsibleHeader()}
        {!collapsed && (
          <CardContent>
            <div className="flex items-center justify-between rounded-lg bg-error-bg px-4 py-3 text-error-text text-[13px]">
              <span>加载失败</span>
              <Button variant="ghost" size="sm" onClick={() => fetchLogs(1)}>
                <RefreshCw size={12} />
                重试
              </Button>
            </div>
          </CardContent>
        )}
      </Card>
    );
  }

  // Loading state
  if (loading) {
    return (
      <Card className="mb-5">
        {collapsibleHeader()}
        {!collapsed && (
          <CardContent>
            <div className="space-y-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <div
                  key={i}
                  data-testid="timeline-skeleton"
                  className="h-16 rounded-lg bg-bg-alt animate-pulse"
                />
              ))}
            </div>
          </CardContent>
        )}
      </Card>
    );
  }

  // Empty state
  if (logs.length === 0) {
    return (
      <Card className="mb-5">
        {collapsibleHeader(
          showAddButton ? (
            <Button
              variant="primary"
              size="sm"
              onClick={(e) => { e.stopPropagation(); onAdd(); }}
            >
              <Plus size={14} />
              添加决策
            </Button>
          ) : undefined,
        )}
        {!collapsed && (
          <CardContent>
            <div className="py-6 text-center">
              <p className="text-tertiary text-[13px]">暂无决策记录</p>
              {showAddButton && (
                <Button
                  variant="primary"
                  size="sm"
                  className="mt-3"
                  onClick={onAdd}
                >
                  <Plus size={14} />
                  添加决策
                </Button>
              )}
            </div>
          </CardContent>
        )}
      </Card>
    );
  }

  // Populated state
  const allLoaded = logs.length >= total;

  return (
    <Card className="mb-5">
      {collapsibleHeader(
        showAddButton ? (
          <Button
            variant="primary"
            size="sm"
            onClick={(e) => { e.stopPropagation(); onAdd(); }}
          >
            添加决策
          </Button>
        ) : undefined,
      )}
      {!collapsed && (
        <>
          <CardContent>
            <div role="feed" aria-label="决策记录">
              {logs.map((log, index) => {
                const isDraft = log.logStatus === "draft";
                const canEdit = isDraft && log.createdBy === currentUserBizKey;
                const isExpanded = expandedIds.has(log.bizKey);
                const visibleTags = log.tags.slice(0, MAX_VISIBLE_TAGS);
                const overflowCount = log.tags.length - MAX_VISIBLE_TAGS;

                return (
                  <div
                    key={log.bizKey}
                    role="article"
                    aria-label={`${CATEGORY_LABELS[log.category] ?? log.category} — ${formatTime(log.createTime)}`}
                    tabIndex={0}
                    className={`flex gap-3 ${index > 0 ? "border-t border-border pt-3 mt-3" : ""}`}
                  >
                    {/* Left Rail */}
                    <div className="flex flex-col items-center w-6 shrink-0">
                      <div
                        className={`w-1.5 h-1.5 rounded-full mt-1.5 ${isDraft ? "bg-warning" : "bg-primary-500"}`}
                      />
                      {index < logs.length - 1 && (
                        <div className="w-px flex-1 bg-border mt-1" />
                      )}
                    </div>

                    {/* Right Content */}
                    <div className="flex-1 min-w-0">
                      {/* Row 1: Category + Draft badge + Time */}
                      <div className="flex items-center gap-2 mb-1">
                        <Badge
                          variant={CATEGORY_VARIANTS[log.category] ?? "default"}
                          aria-label={CATEGORY_LABELS[log.category] ?? log.category}
                        >
                          {CATEGORY_LABELS[log.category] ?? log.category}
                        </Badge>
                        {isDraft && (
                          <Badge
                            variant="warning"
                            className="bg-warning-bg text-warning-text"
                            aria-label="草稿"
                          >
                            草稿
                          </Badge>
                        )}
                        <span className="text-tertiary text-xs ml-auto">
                          {formatTime(log.createTime)}
                        </span>
                      </div>

                      {/* Row 2: Content */}
                      <div
                        role="button"
                        tabIndex={0}
                        aria-expanded={isExpanded}
                        aria-controls={`content-${log.bizKey}`}
                        onClick={() => toggleExpand(log.bizKey)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter" || e.key === " ") {
                            e.preventDefault();
                            toggleExpand(log.bizKey);
                          }
                        }}
                        className={`text-[13px] text-secondary cursor-pointer hover:text-primary transition-colors ${!isExpanded ? "line-clamp-2" : ""}`}
                      >
                        {log.content}
                      </div>
                      <div id={`content-${log.bizKey}`} className="sr-only" />

                      {/* Row 3: Tags + Creator + Edit */}
                      <div className="flex items-center gap-2 mt-2 flex-wrap">
                        {visibleTags.map((tag) => (
                          <Badge
                            key={tag}
                            className="max-w-[120px] overflow-hidden text-ellipsis whitespace-nowrap"
                          >
                            {tag}
                          </Badge>
                        ))}
                        {overflowCount > 0 && <Badge>+{overflowCount}</Badge>}
                        <span className="text-tertiary text-xs">
                          {log.creatorName}
                        </span>
                        {canEdit && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="ml-auto text-primary-600"
                            onClick={(e) => {
                              e.stopPropagation();
                              onEdit(log.bizKey);
                            }}
                          >
                            <Pencil size={12} />
                            编辑
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Infinite scroll sentinel */}
            {!allLoaded && <div ref={sentinelRef} className="h-1" />}

            {/* Loading more spinner */}
            {loadingMore && (
              <div className="flex justify-center py-3">
                <div className="h-6 w-6 border-2 border-t-transparent rounded-full animate-spin border-primary-500" />
              </div>
            )}

            {/* Pagination error */}
            {paginationError && (
              <div className="flex items-center justify-between rounded-lg bg-error-bg px-4 py-3 text-error-text text-[13px] mt-3">
                <span>加载更多失败</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => fetchLogs(page + 1)}
                >
                  <RefreshCw size={12} />
                  重试
                </Button>
              </div>
            )}
          </CardContent>
          {allLoaded && (
            <CardFooter>
              <span
                className="text-tertiary text-xs"
                aria-live="polite"
                aria-atomic="true"
              >
                已加载 {logs.length} 条
              </span>
            </CardFooter>
          )}
        </>
      )}
    </Card>
  );
}
