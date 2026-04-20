import { useState } from 'react'
import { getWeeklyReportPreviewApi, exportWeeklyReportApi } from '@/api/reports'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { PermissionGuard } from '@/components/PermissionGuard'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import type { ReportPreviewResp } from '@/types'

function formatWeekLabel(weekValue: string): string {
  if (!weekValue) return ''
  const [year, week] = weekValue.split('-W')
  const weekNum = parseInt(week, 10)
  // Calculate start date of the ISO week
  const jan4 = new Date(parseInt(year, 10), 0, 4)
  const dayOfWeek = jan4.getDay() || 7
  const startOfYear = new Date(parseInt(year, 10), 0, 1)
  const weekStart = new Date(startOfYear)
  weekStart.setDate(jan4.getDate() - dayOfWeek + 1 + (weekNum - 1) * 7)
  const weekEnd = new Date(weekStart)
  weekEnd.setDate(weekStart.getDate() + 6)
  const fmt = (d: Date) => `${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  return `${year}年第${weekNum}周 (${fmt(weekStart)} ~ ${fmt(weekEnd)})`
}

function renderMarkdown(preview: ReportPreviewResp): string {
  const [year, week] = preview.weekStart.split('-')
  const weekNum = (() => {
    const d = new Date(preview.weekStart)
    const start = new Date(d.getFullYear(), 0, 1)
    const diff = d.getTime() - start.getTime()
    return Math.ceil((diff / 86400000 + start.getDay() + 1) / 7)
  })()
  let md = `## ${year}年第${weekNum}周 工作周报\n\n`

  for (const section of preview.sections) {
    md += `### ${section.mainItem.title}\n`
    md += `完成度：${section.mainItem.completion}%\n\n`
    for (const sub of section.subItems) {
      const status = sub.completion === 100 ? '已完成' : `进行中 (${sub.completion}%)`
      md += `  - **${sub.title}** -- ${status}\n`
      for (const a of sub.achievements) {
        md += `    成果：${a}\n`
      }
      for (const b of sub.blockers) {
        md += `    卡点：${b}\n`
      }
      md += '\n'
    }
  }

  const user = useAuthStore.getState().user
  const now = new Date().toISOString().slice(0, 10)
  md += `---\n导出时间 ${now} by ${user?.displayName || ''}\n`
  return md
}

export default function ReportPage() {
  const currentTeamId = useTeamStore((s) => s.currentTeamId)
  const [weekValue, setWeekValue] = useState(() => {
    const now = new Date()
    const startOfYear = new Date(now.getFullYear(), 0, 1)
    const diff = now.getTime() - startOfYear.getTime()
    const weekNum = Math.ceil((diff / 86400000 + startOfYear.getDay() + 1) / 7)
    return `${now.getFullYear()}-W${String(weekNum).padStart(2, '0')}`
  })
  const [preview, setPreview] = useState<ReportPreviewResp | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handlePreview = async () => {
    if (!currentTeamId) {
      setError('请先选择团队')
      return
    }
    setError(null)
    setLoading(true)
    try {
      // Convert week input value to weekStart date (Monday of that week)
      const [year, week] = weekValue.split('-W')
      const jan4 = new Date(parseInt(year, 10), 0, 4)
      const dayOfWeek = jan4.getDay() || 7
      const startOfYear = new Date(parseInt(year, 10), 0, 1)
      const weekStart = new Date(startOfYear)
      weekStart.setDate(jan4.getDate() - dayOfWeek + 1 + (parseInt(week, 10) - 1) * 7)
      const weekStartStr = weekStart.toISOString().slice(0, 10)

      const resp = await getWeeklyReportPreviewApi(currentTeamId, weekStartStr)
      setPreview(resp)
    } catch (err: any) {
      setError(err?.response?.data?.message || '获取预览失败')
    } finally {
      setLoading(false)
    }
  }

  const handleExport = async () => {
    if (!currentTeamId || !preview) return
    try {
      const [year, week] = weekValue.split('-W')
      const jan4 = new Date(parseInt(year, 10), 0, 4)
      const dayOfWeek = jan4.getDay() || 7
      const startOfYear = new Date(parseInt(year, 10), 0, 1)
      const weekStart = new Date(startOfYear)
      weekStart.setDate(jan4.getDate() - dayOfWeek + 1 + (parseInt(week, 10) - 1) * 7)
      const weekStartStr = weekStart.toISOString().slice(0, 10)

      const blob = await exportWeeklyReportApi(currentTeamId, weekStartStr)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `weekly-report-${weekValue}.md`
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
    } catch {
      // Export failed silently - user can retry
    }
  }

  return (
    <div data-testid="report-page">
      <div className="page-header">
        <h1>周报导出</h1>
      </div>

      {/* Config Card */}
      <Card className="mb-5">
        <CardContent className="flex items-end gap-4">
          <div>
            <label htmlFor="week-input" className="block text-sm font-medium text-secondary mb-1.5">
              选择周次
            </label>
            <Input
              id="week-input"
              type="week"
              value={weekValue}
              onChange={(e) => setWeekValue(e.target.value)}
              className="w-[180px]"
            />
          </div>
          <div className="text-sm text-secondary pb-2">
            {formatWeekLabel(weekValue)}
          </div>
          <Button onClick={handlePreview} disabled={loading}>
            {loading ? '生成中...' : '生成预览'}
          </Button>
        </CardContent>
      </Card>

      {error && (
        <div className="bg-error-bg border border-error-text/20 rounded-lg px-4 py-2.5 text-sm text-error-text mb-4">
          {error}
        </div>
      )}

      {/* Preview Card */}
      {preview && (
        <Card>
          <CardHeader>
            <h3>预览</h3>
            <PermissionGuard code="report:export">
              <Button variant="primary" size="sm" onClick={handleExport}>
                导出 Markdown
              </Button>
            </PermissionGuard>
          </CardHeader>
          <CardContent>
            <pre className="font-mono text-[13px] leading-[1.8] text-secondary whitespace-pre-wrap bg-bg-alt p-5 rounded-lg border border-border">
              {renderMarkdown(preview)}
            </pre>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
