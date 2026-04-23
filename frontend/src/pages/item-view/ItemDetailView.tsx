import { Fragment } from 'react'
import { Link } from 'react-router-dom'
import { Pencil, Plus } from 'lucide-react'
import { MainItem, SubItem } from '@/types'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { Pagination, PaginationPageSize } from '@/components/ui/pagination'
import PriorityBadge from '@/components/shared/PriorityBadge'
import StatusTransitionDropdown from '@/components/shared/StatusTransitionDropdown'
import { Badge } from '@/components/ui/badge'
import { PermissionGuard } from '@/components/PermissionGuard'
import { MAIN_ITEM_STATUSES, SUB_ITEM_STATUSES } from '@/lib/status'
import { isOverdue } from '@/lib/status'

interface DetailViewProps {
  items: (MainItem & { subItems?: SubItem[] })[]
  subItemsMap: Record<number, SubItem[]>
  memberName: (id: number | null) => string
  formatDate: (date: string | null) => string
  teamId: number
  onRefresh: () => void
  currentPage: number
  totalPages: number
  pageSize: number
  onPageChange: (page: number) => void
  onPageSizeChange: (size: number) => void
  totalItems: number
  onAddSubItem: (mainItemId: number, mainItemTitle: string) => void
  onEditMainItem: (item: MainItem) => void
  onAppendProgress: (subItemId: number, subItemTitle: string) => void
  onEditSubItem: (sub: SubItem) => void
}

export default function ItemDetailView({
  items,
  subItemsMap,
  memberName,
  formatDate,
  teamId,
  onRefresh,
  currentPage,
  totalPages,
  pageSize,
  onPageChange,
  onPageSizeChange,
  totalItems,
  onAddSubItem,
  onEditMainItem,
  onAppendProgress,
  onEditSubItem,
}: DetailViewProps) {
  return (
    <div className="rounded-xl border border-border bg-white shadow-sm">
      <div data-testid="detail-table">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>编号</TableHead>
              <TableHead>优先级</TableHead>
              <TableHead>标题</TableHead>
              <TableHead>负责人</TableHead>
              <TableHead>进度</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>开始时间</TableHead>
              <TableHead>预期完成时间</TableHead>
              <TableHead>结束时间</TableHead>
              <TableHead>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => {
              const subs = subItemsMap[item.id]
              return (
                <Fragment key={item.id}>
                  <TableRow className={subs?.length ? 'bg-blue-50/40' : ''}>
                    <TableCell>
                      <span className="font-mono text-xs">{item.code}</span>
                    </TableCell>
                    <TableCell>
                      <PriorityBadge priority={item.priority} />
                    </TableCell>
                    <TableCell>
                      <Link to={`/items/${item.id}`} className="font-medium text-primary-600 hover:text-primary-700 hover:underline truncate block max-w-xs" title={item.title}>
                        {item.title}
                      </Link>
                    </TableCell>
                    <TableCell>{memberName(item.assigneeId)}</TableCell>
                    <TableCell>
                      <span className="text-xs">{item.completion}%</span>
                    </TableCell>
                    <TableCell>
                      <StatusTransitionDropdown currentStatus={item.status} itemType="main" teamId={teamId} itemId={item.id} onStatusChanged={onRefresh} />
                    </TableCell>
                    <TableCell className="text-xs">{formatDate(item.startDate)}</TableCell>
                    <TableCell className="text-xs">
                      <span>{formatDate(item.expectedEndDate)}</span>
                      {isOverdue(item.expectedEndDate ?? undefined, item.status) && (
                        <Badge variant="error" className="ml-1">延期</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-xs">{formatDate(item.actualEndDate)}</TableCell>
                    <TableCell>
                      <div className="flex gap-0.5 whitespace-nowrap">
                        <Link to={`/items/${item.id}`}><Button variant="ghost" size="sm" disabled={!!MAIN_ITEM_STATUSES[item.status as keyof typeof MAIN_ITEM_STATUSES]?.terminal}><Pencil size={14} />编辑</Button></Link>
                        <Button variant="ghost" size="sm" disabled={!!MAIN_ITEM_STATUSES[item.status as keyof typeof MAIN_ITEM_STATUSES]?.terminal} onClick={() => onAddSubItem(item.id, item.title)}><Plus size={14} />添加子事项</Button>
                      </div>
                    </TableCell>
                  </TableRow>
                  {subs?.map((sub) => (
                    <TableRow key={`sub-${sub.id}`} className="bg-bg-alt/60">
                      <TableCell>
                        <span className="font-mono text-[11px] text-tertiary ml-4">SI-{String(item.id).padStart(3, '0')}-{String(sub.id).slice(-2)}</span>
                      </TableCell>
                      <TableCell>
                        <PriorityBadge priority={sub.priority} />
                      </TableCell>
                      <TableCell>
                        <Link to={`/items/${item.id}/sub/${sub.id}`} className="font-medium text-primary-600 hover:text-primary-700 hover:underline ml-4">
                          {sub.title}
                        </Link>
                      </TableCell>
                      <TableCell>{memberName(sub.assigneeId)}</TableCell>
                      <TableCell>
                        <span className="text-xs">{sub.completion}%</span>
                      </TableCell>
                      <TableCell>
                        <StatusTransitionDropdown currentStatus={sub.status} itemType="sub" teamId={teamId} itemId={sub.id} onStatusChanged={onRefresh} />
                      </TableCell>
                      <TableCell className="text-xs">{formatDate(sub.startDate)}</TableCell>
                      <TableCell className="text-xs">{formatDate(sub.expectedEndDate)}</TableCell>
                      <TableCell className="text-xs">{formatDate(sub.actualEndDate)}</TableCell>
                      <TableCell>
                        <div className="flex gap-0.5 whitespace-nowrap">
                          <PermissionGuard code="main_item:update">
                            <Button variant="ghost" size="sm" disabled={!!SUB_ITEM_STATUSES[sub.status as keyof typeof SUB_ITEM_STATUSES]?.terminal} onClick={() => onEditSubItem(sub)}><Pencil size={14} />编辑</Button>
                          </PermissionGuard>
                          <PermissionGuard code="progress:update">
                            <Button variant="ghost" size="sm" disabled={!!SUB_ITEM_STATUSES[sub.status as keyof typeof SUB_ITEM_STATUSES]?.terminal} onClick={() => onAppendProgress(sub.id, sub.title)}><Plus size={14} />追加进度</Button>
                          </PermissionGuard>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </Fragment>
              )
            })}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-center gap-3 px-5 py-3 border-t border-border">
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={onPageChange}
        />
        <PaginationPageSize
          pageSize={pageSize}
          onPageSizeChange={onPageSizeChange}
          options={[5, 10, 20, 50]}
        />
        <span className="text-[13px] text-tertiary">共 {totalItems} 条</span>
      </div>
    </div>
  )
}
