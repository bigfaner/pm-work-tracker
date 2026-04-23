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
import PaginationBar from '@/components/shared/PaginationBar'
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
  onAppendProgress: (subItemId: number, subItemTitle: string, subItemCompletion: number) => void
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
      <div data-testid="detail-table" className="overflow-x-auto">
        <Table className="min-w-[1100px]">
          <TableHeader>
            <TableRow>
              <TableHead className="whitespace-nowrap">编号</TableHead>
              <TableHead className="whitespace-nowrap">优先级</TableHead>
              <TableHead className="min-w-[180px]">标题</TableHead>
              <TableHead className="whitespace-nowrap">负责人</TableHead>
              <TableHead className="whitespace-nowrap">进度</TableHead>
              <TableHead className="whitespace-nowrap">状态</TableHead>
              <TableHead className="whitespace-nowrap">开始时间</TableHead>
              <TableHead className="whitespace-nowrap">预期完成时间</TableHead>
              <TableHead className="whitespace-nowrap">结束时间</TableHead>
              <TableHead className="whitespace-nowrap">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => {
              const subs = subItemsMap[item.id]
              return (
                <Fragment key={item.id}>
                  <TableRow className={subs?.length ? 'bg-blue-50/40' : ''}>
                    <TableCell className="whitespace-nowrap">
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
                    <TableCell className="whitespace-nowrap">{memberName(item.assigneeId)}</TableCell>
                    <TableCell className="whitespace-nowrap">
                      <span className="text-xs">{item.completion}%</span>
                    </TableCell>
                    <TableCell className="whitespace-nowrap">
                      <StatusTransitionDropdown currentStatus={item.status} itemType="main" teamId={teamId} itemId={item.id} onStatusChanged={onRefresh} />
                    </TableCell>
                    <TableCell className="text-xs whitespace-nowrap">{formatDate(item.startDate)}</TableCell>
                    <TableCell className="text-xs whitespace-nowrap">
                      <span>{formatDate(item.expectedEndDate)}</span>
                      {isOverdue(item.expectedEndDate ?? undefined, item.status, new Date()) && (
                        <Badge variant="error" className="ml-1">延期</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-xs whitespace-nowrap">{formatDate(item.actualEndDate)}</TableCell>
                    <TableCell>
                      <div className="flex gap-0.5 whitespace-nowrap">
                        <Link to={`/items/${item.id}`}><Button variant="ghost" size="sm" className="text-primary-600" disabled={!!MAIN_ITEM_STATUSES[item.status as keyof typeof MAIN_ITEM_STATUSES]?.terminal}><Pencil size={14} />编辑</Button></Link>
                        <Button variant="ghost" size="sm" className="text-primary-600" disabled={!!MAIN_ITEM_STATUSES[item.status as keyof typeof MAIN_ITEM_STATUSES]?.terminal} onClick={() => onAddSubItem(item.id, item.title)}><Plus size={14} />添加子事项</Button>
                      </div>
                    </TableCell>
                  </TableRow>
                  {subs?.map((sub) => (
                    <TableRow key={`sub-${sub.id}`} className="bg-bg-alt/60">
                      <TableCell className="whitespace-nowrap">
                        <span className="font-mono text-[11px] text-tertiary ml-4">{sub.code}</span>
                      </TableCell>
                      <TableCell>
                        <PriorityBadge priority={sub.priority} />
                      </TableCell>
                      <TableCell>
                        <Link to={`/items/${item.id}/sub/${sub.id}`} className="font-medium text-primary-600 hover:text-primary-700 hover:underline ml-4">
                          {sub.title}
                        </Link>
                      </TableCell>
                      <TableCell className="whitespace-nowrap">{memberName(sub.assigneeId)}</TableCell>
                      <TableCell className="whitespace-nowrap">
                        <span className="text-xs">{sub.completion}%</span>
                      </TableCell>
                      <TableCell className="whitespace-nowrap">
                        <StatusTransitionDropdown currentStatus={sub.status} itemType="sub" teamId={teamId} itemId={sub.id} onStatusChanged={onRefresh} />
                      </TableCell>
                      <TableCell className="text-xs whitespace-nowrap">{formatDate(sub.startDate)}</TableCell>
                      <TableCell className="text-xs whitespace-nowrap">{formatDate(sub.expectedEndDate)}</TableCell>
                      <TableCell className="text-xs whitespace-nowrap">{formatDate(sub.actualEndDate)}</TableCell>
                      <TableCell>
                        <div className="flex gap-0.5 whitespace-nowrap">
                          <PermissionGuard code="main_item:update">
                            <Button variant="ghost" size="sm" className="text-primary-600" disabled={!!SUB_ITEM_STATUSES[sub.status as keyof typeof SUB_ITEM_STATUSES]?.terminal} onClick={() => onEditSubItem(sub)}><Pencil size={14} />编辑</Button>
                          </PermissionGuard>
                          <PermissionGuard code="progress:update">
                            <Button variant="ghost" size="sm" className="text-primary-600" disabled={!!SUB_ITEM_STATUSES[sub.status as keyof typeof SUB_ITEM_STATUSES]?.terminal} onClick={() => onAppendProgress(sub.id, sub.title, sub.completion)}><Plus size={14} />追加进度</Button>
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
      <PaginationBar
        currentPage={currentPage}
        totalPages={totalPages}
        total={totalItems}
        onPageChange={onPageChange}
        pageSize={pageSize}
        onPageSizeChange={onPageSizeChange}
        pageSizeOptions={[5, 10, 20, 50]}
      />
    </div>
  )
}
