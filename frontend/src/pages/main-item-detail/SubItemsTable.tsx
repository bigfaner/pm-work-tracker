import { Pencil, Plus } from 'lucide-react'
import type { SubItem } from '@/types'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Link } from 'react-router-dom'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { Card, CardHeader } from '@/components/ui/card'
import StatusTransitionDropdown from '@/components/shared/StatusTransitionDropdown'
import { MAIN_ITEM_STATUSES } from '@/lib/status'
import { formatDate } from '@/lib/format'

interface SubItemsTableProps {
  subItems: SubItem[]
  mainItemId: number
  mainStatus: string
  teamId: number
  memberName: (assigneeId: number | null) => string
  onStatusChanged: () => void
  onEditSub: (sub: SubItem) => void
  onAppendProgress: (sub: SubItem) => void
  onCreateSub: () => void
}

export default function SubItemsTable({
  subItems,
  mainItemId,
  mainStatus,
  teamId,
  memberName,
  onStatusChanged,
  onEditSub,
  onAppendProgress,
  onCreateSub,
}: SubItemsTableProps) {
  const mainTerminal = (MAIN_ITEM_STATUSES as Record<string, { terminal: boolean }>)[mainStatus]?.terminal ?? false

  return (
    <Card>
      <CardHeader>
        <h3 className="text-sm font-semibold text-primary m-0">子事项列表</h3>
        <Button variant="ghost" size="sm" className="text-primary-600" disabled={mainTerminal} onClick={onCreateSub}>+ 新增子事项</Button>
      </CardHeader>
      {subItems.length > 0 && (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>编号</TableHead>
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
            {subItems.map((sub) => (
              <TableRow key={sub.id}>
                <TableCell>
                  <Badge variant="default" className="font-mono text-[11px]">{sub.code}</Badge>
                </TableCell>
                <TableCell>
                  <Link
                    to={`/items/${mainItemId}/sub/${sub.id}`}
                    className="text-[13px] font-medium text-primary-600 hover:text-primary-700 hover:underline truncate block max-w-xs"
                    title={sub.title}
                  >
                    {sub.title}
                  </Link>
                </TableCell>
                <TableCell>{memberName(sub.assigneeId)}</TableCell>
                <TableCell>
                  <span>{sub.completion}%</span>
                </TableCell>
                <TableCell>
                  <StatusTransitionDropdown
                    currentStatus={sub.status}
                    itemType="sub"
                    teamId={teamId}
                    itemId={sub.id}
                    onStatusChanged={onStatusChanged}
                  />
                </TableCell>
                <TableCell className="text-xs">{formatDate(sub.startDate)}</TableCell>
                <TableCell className="text-xs">{formatDate(sub.expectedEndDate)}</TableCell>
                <TableCell className="text-xs">{formatDate(sub.actualEndDate)}</TableCell>
                <TableCell>
                  <div className="flex gap-0.5">
                    <Button variant="ghost" size="sm" className="text-primary-600" disabled={mainTerminal} onClick={() => onEditSub(sub)}><Pencil size={12} />编辑</Button>
                    <Button variant="ghost" size="sm" className="text-primary-600" disabled={mainTerminal} onClick={() => onAppendProgress(sub)}><Plus size={12} />追加进度</Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </Card>
  )
}
