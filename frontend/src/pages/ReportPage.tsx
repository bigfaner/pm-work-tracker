import { useState, useCallback, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Card,
  Button,
  DatePicker,
  Empty,
  Tag,
  Progress,
  Skeleton,
  message,
} from 'antd'
import { DownloadOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import isoWeek from 'dayjs/plugin/isoWeek'
import { useTeamStore } from '@/store/team'
import { getWeeklyReportPreviewApi, exportWeeklyReportApi } from '@/api/reports'
import type { ReportPreviewResp, ReportSection, ReportSubItem } from '@/types'

dayjs.extend(isoWeek)

function getMondayOfWeek(d: dayjs.Dayjs): dayjs.Dayjs {
  return d.isoWeekday(1)
}

function ReportSectionView({ section }: { section: ReportSection }) {
  return (
    <div style={{ marginBottom: 24 }}>
      <h3 style={{ marginBottom: 12 }}>
        {section.mainItem.isKeyItem && (
          <Tag color="orange" style={{ marginRight: 8 }}>重点</Tag>
        )}
        {section.mainItem.title}
        {' '}
        <span style={{ color: '#8c8c8c', fontWeight: 400, fontSize: 14 }}>
          (完成度 {Math.round(section.mainItem.completion)}%)
        </span>
      </h3>
      {section.subItems.map((sub) => (
        <SubItemView key={sub.id} sub={sub} />
      ))}
    </div>
  )
}

function SubItemView({ sub }: { sub: ReportSubItem }) {
  return (
    <div style={{ marginBottom: 12, paddingLeft: 16 }}>
      <div style={{ marginBottom: 4, fontWeight: 500 }}>{sub.title}</div>
      <Progress percent={Math.round(sub.completion)} size="small" style={{ width: 200, marginBottom: 4 }} />
      {sub.achievements.length > 0 && (
        <div style={{ marginBottom: 4 }}>
          <span style={{ fontWeight: 500 }}>成果：</span>
          <ul style={{ margin: '4px 0', paddingLeft: 20 }}>
            {sub.achievements.map((a, i) => (
              <li key={i}>{a}</li>
            ))}
          </ul>
        </div>
      )}
      {sub.blockers.length > 0 && (
        <div>
          <span style={{ fontWeight: 500 }}>卡点：</span>
          <ul style={{ margin: '4px 0', paddingLeft: 20 }}>
            {sub.blockers.map((b, i) => (
              <li key={i}>{b}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

function PreviewSkeleton() {
  return (
    <div data-testid="preview-loading">
      {[1, 2, 3].map((i) => (
        <div key={i} style={{ marginBottom: 24 }}>
          <Skeleton.Input active style={{ width: 300, height: 24, marginBottom: 12 }} />
          <div style={{ paddingLeft: 16 }}>
            <Skeleton active paragraph={{ rows: 3 }} />
          </div>
        </div>
      ))}
    </div>
  )
}

export default function ReportPage() {
  const { currentTeamId } = useTeamStore()
  const [selectedWeek, setSelectedWeek] = useState<dayjs.Dayjs>(getMondayOfWeek(dayjs()))
  const [previewTriggered, setPreviewTriggered] = useState(false)
  const [exporting, setExporting] = useState(false)
  const [previewData, setPreviewData] = useState<ReportPreviewResp | null>(null)

  const weekStartStr = useMemo(() => selectedWeek.format('YYYY-MM-DD'), [selectedWeek])

  const {
    refetch: fetchPreview,
    isFetching: previewLoading,
  } = useQuery({
    queryKey: ['reportPreview', currentTeamId, weekStartStr],
    queryFn: async () => {
      const data = await getWeeklyReportPreviewApi(currentTeamId!, weekStartStr)
      setPreviewData(data)
      return data
    },
    enabled: false,
  })

  const handlePreview = useCallback(async () => {
    setPreviewTriggered(true)
    setPreviewData(null)
    try {
      await fetchPreview()
    } catch (err: unknown) {
      const error = err as { response?: { data?: { code?: string } } }
      if (error?.response?.data?.code === 'NO_DATA') {
        message.warning('所选周暂无数据')
      }
    }
  }, [fetchPreview])

  const handleExport = useCallback(async () => {
    if (!currentTeamId || !previewData) return
    setExporting(true)
    try {
      const blob = await exportWeeklyReportApi(currentTeamId, weekStartStr)
      const url = URL.createObjectURL(new Blob([blob], { type: 'text/markdown' }))
      const a = document.createElement('a')
      a.href = url
      a.download = `weekly-report-${weekStartStr}.md`
      a.click()
      URL.revokeObjectURL(url)
      message.success('周报已导出')
    } catch (err: unknown) {
      const error = err as { response?: { data?: { code?: string } } }
      if (error?.response?.data?.code === 'NO_DATA') {
        message.warning('所选周暂无数据')
      } else {
        message.error('导出失败')
      }
    } finally {
      setExporting(false)
    }
  }, [currentTeamId, previewData, weekStartStr])

  const handleWeekChange = useCallback((date: dayjs.Dayjs | null) => {
    if (date) {
      setSelectedWeek(getMondayOfWeek(date))
      setPreviewTriggered(false)
      setPreviewData(null)
    }
  }, [])

  return (
    <div data-testid="report-page">
      {/* Page Header */}
      <h2 style={{ marginBottom: 16 }}>周报导出</h2>

      {/* Config Card */}
      <Card style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
          <DatePicker
            data-testid="week-picker"
            picker="week"
            defaultValue={selectedWeek}
            onChange={handleWeekChange}
            disabledDate={(d) => d && d.isAfter(dayjs(), 'day')}
          />
          <Button
            data-testid="preview-btn"
            onClick={handlePreview}
            loading={previewLoading}
          >
            生成预览
          </Button>
        </div>
      </Card>

      {/* Preview Card */}
      <Card
        title="预览"
        extra={
          <Button
            data-testid="export-btn"
            type="primary"
            icon={<DownloadOutlined />}
            onClick={handleExport}
            loading={exporting}
            disabled={exporting || !previewData}
          >
            导出 Markdown
          </Button>
        }
      >
        {previewLoading && <PreviewSkeleton />}

        {!previewLoading && !previewTriggered && (
          <div data-testid="preview-empty">
            <Empty description="请选择周次后点击预览" />
          </div>
        )}

        {!previewLoading && previewTriggered && !previewData && (
          <div data-testid="preview-empty">
            <Empty description="所选周暂无数据，无法导出" />
          </div>
        )}

        {!previewLoading && previewData && (
          <div data-testid="preview-content">
            <h2 style={{ marginBottom: 16 }}>
              {selectedWeek.format('YYYY')}年第{selectedWeek.isoWeek()}周 工作周报
            </h2>
            {previewData.sections.map((section) => (
              <ReportSectionView key={section.mainItem.id} section={section} />
            ))}
            <div style={{ color: '#8c8c8c', fontSize: 12, marginTop: 24 }}>
              导出时间：{dayjs().format('YYYY-MM-DD HH:mm')}
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}
