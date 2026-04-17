import { useState, useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { DatePicker, Card, Progress, Tag, Empty, Skeleton, Badge } from 'antd'
import { useTeamStore } from '@/store/team'
import { getWeeklyViewApi } from '@/api/views'
import type { WeeklyGroup, SubItemWithProgress } from '@/types'
import dayjs from 'dayjs'

const STATUS_TAG_COLOR_MAP: Record<string, string> = {
  '未开始': 'default',
  '进行中': 'processing',
  '待评审': 'warning',
  '已完成': 'success',
  '已关闭': 'default',
  '阻塞中': 'error',
  '延期': 'orange',
  '归档': 'default',
}

/** Get Monday of the given week (dayjs week starts Sunday by default) */
function getMonday(d: dayjs.Dayjs): dayjs.Dayjs {
  const day = d.day()
  // day() returns 0=Sun, 1=Mon, ..., 6=Sat
  // If Sunday (0), go back 6 days to previous Monday; otherwise subtract (day-1)
  return day === 0 ? d.subtract(6, 'day') : d.subtract(day - 1, 'day')
}

interface NoChangeRowProps {
  item: { id: number; title: string; status: string; completion: number }
}

function NoChangeRow({ item }: NoChangeRowProps) {
  return (
    <div
      data-testid={`nochange-row-${item.id}`}
      style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '4px 0' }}
    >
      <span style={{ flex: 1 }}>{item.title}</span>
      <Tag color={STATUS_TAG_COLOR_MAP[item.status] || 'default'}>{item.status}</Tag>
      <Progress type="line" percent={item.completion} size="small" style={{ width: 100 }} />
    </div>
  )
}

interface SubItemCardProps {
  data: SubItemWithProgress
}

function ProgressSubItemCard({ data }: SubItemCardProps) {
  const { subItem, progressThisWeek } = data
  return (
    <div style={{ padding: '8px 0' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
        <span style={{ fontWeight: 500 }}>{subItem.title}</span>
        <Progress
          data-testid={`sub-progress-${subItem.id}`}
          type="line"
          percent={subItem.completion}
          size="small"
          style={{ width: 100 }}
        />
      </div>
      {progressThisWeek.length > 0 && (
        <div style={{ paddingLeft: 12 }}>
          {progressThisWeek.map((record) => (
            <div key={record.id} style={{ fontSize: 13, color: '#595959', marginBottom: 4 }}>
              {record.achievement && <div>成果: {record.achievement}</div>}
              {record.blocker && <div>卡点: {record.blocker}</div>}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

interface GroupCardProps {
  group: WeeklyGroup
}

function GroupCard({ group }: GroupCardProps) {
  const { mainItem, newlyCompleted, hasProgress, noChangeFromLastWeek } = group

  return (
    <Card
      size="small"
      style={{ marginBottom: 16 }}
      title={
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ fontWeight: 600 }}>{mainItem.title}</span>
          <Progress
            data-testid={`main-item-progress-${mainItem.id}`}
            type="line"
            percent={mainItem.completion}
            size="small"
            style={{ width: 120 }}
          />
        </div>
      }
    >
      {/* Section: 本周新完成 (green) */}
      <div style={{ marginBottom: 12 }}>
        <div style={{ color: '#52c41a', fontWeight: 600, marginBottom: 4 }}>本周新完成</div>
        {newlyCompleted.length > 0 ? (
          newlyCompleted.map((item) => (
            <div
              key={item.subItem.id}
              data-testid={`completed-item-${item.subItem.id}`}
              style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '4px 0' }}
            >
              <span>{item.subItem.title}</span>
              <Badge
                data-testid={`completed-badge-${item.subItem.id}`}
                count="100%"
                style={{ backgroundColor: '#52c41a', fontSize: 12 }}
              />
            </div>
          ))
        ) : (
          <span style={{ color: '#bfbfbf', fontSize: 13 }}>暂无</span>
        )}
      </div>

      {/* Section: 本周有进度 (blue) */}
      <div style={{ marginBottom: 12 }}>
        <div style={{ color: '#1677ff', fontWeight: 600, marginBottom: 4 }}>本周有进度</div>
        {hasProgress.length > 0 ? (
          hasProgress.map((item) => (
            <ProgressSubItemCard key={item.subItem.id} data={item} />
          ))
        ) : (
          <span style={{ color: '#bfbfbf', fontSize: 13 }}>暂无</span>
        )}
      </div>

      {/* Section: 上周完成/无变化 (gray) */}
      <div>
        <div style={{ color: '#8c8c8c', fontWeight: 600, marginBottom: 4 }}>上周完成/无变化</div>
        {noChangeFromLastWeek.length > 0 ? (
          noChangeFromLastWeek.map((item) => (
            <NoChangeRow key={item.id} item={item} />
          ))
        ) : (
          <span style={{ color: '#bfbfbf', fontSize: 13 }}>暂无</span>
        )}
      </div>
    </Card>
  )
}

export default function WeeklyViewPage() {
  const { currentTeamId } = useTeamStore()

  const [selectedWeek, setSelectedWeek] = useState<dayjs.Dayjs>(() => dayjs())

  const weekStart = useMemo(() => getMonday(selectedWeek).startOf('day'), [selectedWeek])
  const weekStartStr = weekStart.format('YYYY-MM-DD')
  const weekEndStr = useMemo(() => weekStart.add(6, 'day').format('YYYY-MM-DD'), [weekStart])

  const { data, isLoading } = useQuery({
    queryKey: ['weeklyView', currentTeamId, weekStartStr],
    queryFn: () => getWeeklyViewApi(currentTeamId!, weekStartStr),
    enabled: !!currentTeamId,
  })

  const groups = data?.groups ?? []

  const handleWeekChange = useCallback((date: dayjs.Dayjs | null) => {
    if (date) {
      setSelectedWeek(date)
    }
  }, [])

  return (
    <div data-testid="weekly-view-page">
      {/* Page Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <h2 style={{ margin: 0 }}>周视图</h2>
          {data && (
            <span data-testid="week-range" style={{ color: '#8c8c8c', fontSize: 14 }}>
              {data.weekStart} ~ {data.weekEnd}
            </span>
          )}
        </div>
        <DatePicker
          data-testid="week-picker"
          picker="week"
          defaultValue={dayjs()}
          onChange={handleWeekChange}
        />
      </div>

      {/* Content */}
      {isLoading ? (
        <div data-testid="weekly-loading-skeleton">
          <Skeleton active paragraph={{ rows: 3 }} />
          <Skeleton active paragraph={{ rows: 3 }} style={{ marginTop: 16 }} />
          <Skeleton active paragraph={{ rows: 3 }} style={{ marginTop: 16 }} />
        </div>
      ) : groups.length === 0 ? (
        <div data-testid="weekly-empty-state">
          <Empty description="本周暂无事项进度" />
        </div>
      ) : (
        <div>
          {groups.map((group) => (
            <GroupCard key={group.mainItem.id} group={group} />
          ))}
        </div>
      )}
    </div>
  )
}
