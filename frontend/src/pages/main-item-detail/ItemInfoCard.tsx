import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { isOverdue } from '@/lib/status'
import { formatDate } from '@/lib/format'

interface ItemInfoCardProps {
  assigneeName: string
  startDate: string | null
  expectedEndDate: string | null
  actualEndDate: string | null
  status: string
  description?: string
}

export default function ItemInfoCard({
  assigneeName,
  startDate,
  expectedEndDate,
  actualEndDate,
  status,
  description,
}: ItemInfoCardProps) {
  return (
    <Card className="mb-5">
      <CardContent>
        <div className="grid grid-cols-4 gap-4 mb-4">
          <div>
            <div className="text-xs text-tertiary mb-1">负责人</div>
            <span className="text-[13px] font-medium">{assigneeName}</span>
          </div>
          <div>
            <div className="text-xs text-tertiary mb-1">开始时间</div>
            <span className="text-[13px] font-medium">{formatDate(startDate)}</span>
          </div>
          <div>
            <div className="text-xs text-tertiary mb-1">预期完成时间</div>
            <div className="flex items-center gap-1.5">
              <span className="text-[13px] font-medium">{formatDate(expectedEndDate)}</span>
              {isOverdue(expectedEndDate ?? undefined, status) && (
                <Badge variant="error">延期</Badge>
              )}
            </div>
          </div>
          <div>
            <div className="text-xs text-tertiary mb-1">结束时间</div>
            <span className="text-[13px] font-medium">{formatDate(actualEndDate)}</span>
          </div>
        </div>
        <div>
          <div className="text-xs text-tertiary mb-1">描述</div>
          <span className="text-[13px] text-secondary leading-relaxed">{description || '暂无描述'}</span>
        </div>
      </CardContent>
    </Card>
  )
}
