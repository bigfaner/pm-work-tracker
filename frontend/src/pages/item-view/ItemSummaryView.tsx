import { Link } from 'react-router-dom'
import { Pencil, Plus } from 'lucide-react'
import { MainItem, SubItem } from '@/types'
import { Button } from '@/components/ui/button'
import PriorityBadge from '@/components/shared/PriorityBadge'
import ProgressBar from '@/components/shared/ProgressBar'
import StatusTransitionDropdown from '@/components/shared/StatusTransitionDropdown'
import { Badge } from '@/components/ui/badge'
import { PermissionGuard } from '@/components/PermissionGuard'
import { MAIN_ITEM_STATUSES, SUB_ITEM_STATUSES } from '@/lib/status'
import { isOverdue } from '@/lib/status'

interface SummaryViewProps {
  items: (MainItem & { subItems?: SubItem[] })[]
  expandedCards: Set<number>
  onToggleExpand: (id: number) => void
  subItemsMap: Record<number, SubItem[]>
  memberName: (id: number | null) => string
  formatDate: (date: string | null) => string
  hasMore: boolean
  sentinelRef: React.RefObject<HTMLDivElement>
  teamId: number
  onRefresh: () => void
  onAddSubItem: (mainItemId: number, mainItemTitle: string) => void
  onEditMainItem: (item: MainItem) => void
  onAppendProgress: (subItemId: number, subItemTitle: string, subItemCompletion: number) => void
  onEditSubItem: (sub: SubItem) => void
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
        <div key={item.id} className="mb-3">
          <div
            className="rounded-xl border border-border bg-white shadow-sm cursor-pointer"
            onClick={() => onToggleExpand(item.id)}
          >
            <div className="flex items-center gap-2 px-4 py-3">
              {/* Expand chevron */}
              <svg
                className={`w-3.5 h-3.5 shrink-0 text-tertiary transition-transform ${
                  expandedCards.has(item.id) ? 'rotate-90' : ''
                }`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
              </svg>

              {/* Code */}
              <span className="font-mono text-xs text-tertiary bg-bg-alt px-1.5 py-0.5 rounded">
                {item.code}
              </span>

              {/* Priority */}
              <PriorityBadge priority={item.priority} />

              {/* Title + date range */}
              <div className="flex items-center gap-1.5 flex-1 min-w-0">
                <Link
                  to={`/items/${item.id}`}
                  className="text-sm font-medium text-primary-600 hover:text-primary-700 hover:underline truncate"
                  title={item.title}
                  onClick={(e) => e.stopPropagation()}
                >
                  {item.title}
                </Link>
                {item.startDate && item.expectedEndDate && (
                  <span className="text-xs text-secondary whitespace-nowrap">
                    计划周期 {formatDate(item.startDate)} ~ {formatDate(item.expectedEndDate)}
                  </span>
                )}
                {isOverdue(item.expectedEndDate ?? undefined, item.status) && (
                  <Badge variant="error">延期</Badge>
                )}
                {MAIN_ITEM_STATUSES[item.status as keyof typeof MAIN_ITEM_STATUSES]?.terminal && item.actualEndDate && (
                  <span className="text-xs text-tertiary whitespace-nowrap">
                    结束于 {formatDate(item.actualEndDate)}
                  </span>
                )}
              </div>

              {/* Assignee */}
              <span className="text-[13px] text-secondary whitespace-nowrap">
                {memberName(item.assigneeId)}
              </span>

              {/* Progress */}
              <div className="w-16 shrink-0">
                <ProgressBar value={item.completion} size="sm" showPercentage />
              </div>

              {/* Status */}
              <div onClick={(e) => e.stopPropagation()}>
                <StatusTransitionDropdown currentStatus={item.status} itemType="main" teamId={teamId} itemId={item.id} onStatusChanged={onRefresh} />
              </div>

              {/* Actions */}
              <div className="flex gap-0.5" onClick={(e) => e.stopPropagation()}>
                <Button variant="ghost" size="sm" className="text-primary-600" disabled={!!MAIN_ITEM_STATUSES[item.status as keyof typeof MAIN_ITEM_STATUSES]?.terminal} onClick={() => onEditMainItem(item)}><Pencil size={14} />编辑</Button>
                <Button variant="ghost" size="sm" className="text-primary-600" disabled={!!MAIN_ITEM_STATUSES[item.status as keyof typeof MAIN_ITEM_STATUSES]?.terminal} onClick={() => onAddSubItem(item.id, item.title)}><Plus size={14} />新增子事项</Button>
              </div>
            </div>

            {/* Expanded sub-items */}
            {expandedCards.has(item.id) && (
              <div className="border-t border-border px-5 py-3 pl-12">
                {!subItemsMap[item.id] && (
                  <div className="text-xs text-tertiary py-2">加载中...</div>
                )}
                {subItemsMap[item.id]?.length === 0 && (
                  <div className="text-xs text-tertiary py-2">暂无子事项</div>
                )}
                {subItemsMap[item.id]?.map((sub) => (
                  <div
                    key={sub.id}
                    className="flex items-center gap-2 py-2 border-b border-border/50 last:border-b-0"
                  >
                    <span className="font-mono text-[11px] text-tertiary bg-bg-alt px-1.5 py-0.5 rounded">
                      {sub.code}
                    </span>
                    <PriorityBadge priority={sub.priority} className="text-[10px]" />
                    <Link
                      to={`/items/${item.id}/sub/${sub.id}`}
                      className="text-[13px] font-medium text-primary-600 hover:text-primary-700 hover:underline truncate"
                    >
                      {sub.title}
                    </Link>
                    <span className="text-[11px] text-tertiary whitespace-nowrap">
                      {sub.startDate && sub.expectedEndDate
                        ? `计划周期 ${formatDate(sub.startDate)} ~ ${formatDate(sub.expectedEndDate)}`
                        : '-'}
                    </span>
                    {isOverdue(sub.expectedEndDate ?? undefined, sub.status) && (
                      <Badge variant="error">延期</Badge>
                    )}
                    {SUB_ITEM_STATUSES[sub.status as keyof typeof SUB_ITEM_STATUSES]?.terminal && sub.actualEndDate && (
                      <span className="text-[11px] text-tertiary whitespace-nowrap">
                        结束于 {formatDate(sub.actualEndDate)}
                      </span>
                    )}
                    <span className="ml-auto text-[13px] text-secondary">
                      {memberName(sub.assigneeId)}
                    </span>
                    <div className="w-16 shrink-0">
                      <ProgressBar value={sub.completion} size="sm" showPercentage />
                    </div>
                    <div onClick={(e) => e.stopPropagation()}>
                      <StatusTransitionDropdown currentStatus={sub.status} itemType="sub" teamId={teamId} itemId={sub.id} onStatusChanged={onRefresh} />
                    </div>
                    <PermissionGuard code="main_item:update">
                      <Button variant="ghost" size="sm" className="text-[11px] h-6 px-1.5 text-primary-600" disabled={!!SUB_ITEM_STATUSES[sub.status as keyof typeof SUB_ITEM_STATUSES]?.terminal} onClick={() => onEditSubItem(sub)}><Pencil size={12} />编辑</Button>
                    </PermissionGuard>
                    <PermissionGuard code="progress:update">
                      <Button variant="ghost" size="sm" className="text-[11px] h-6 px-1.5 text-primary-600" disabled={!!SUB_ITEM_STATUSES[sub.status as keyof typeof SUB_ITEM_STATUSES]?.terminal} onClick={() => onAppendProgress(sub.id, sub.title, sub.completion)}><Plus size={12} />追加进度</Button>
                    </PermissionGuard>
                  </div>
                ))}
              </div>
            )}
          </div>
          {/* Expand button for test targeting */}
          <button
            data-testid={`expand-card-${item.id}`}
            className="hidden"
            onClick={() => onToggleExpand(item.id)}
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
  )
}
