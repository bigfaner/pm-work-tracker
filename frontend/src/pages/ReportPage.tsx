import { useState } from 'react'
import { getWeeklyReportPreviewApi } from '@/api/reports'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { PermissionGuard } from '@/components/PermissionGuard'
import { Button } from '@/components/ui/button'
import { WeekPicker } from '@/components/shared/WeekPicker'
import { getCurrentWeekStart, getWeekNumber, getISOWeekYear } from '@/utils/weekUtils'
import type { ReportPreviewResp } from '@/types'

function renderMarkdown(preview: ReportPreviewResp, filterUserId?: number): string {
  const isoYear = getISOWeekYear(preview.weekStart)
  const weekNum = getWeekNumber(preview.weekStart)
  const user = useAuthStore.getState().user
  const isPersonal = filterUserId != null

  let md = `## ${isoYear}年第${weekNum}周 工作周报`
  if (isPersonal && user?.displayName) {
    md += ` — ${user.displayName}`
  }
  md += '\n\n'

  for (const section of preview.sections) {
    const subs = isPersonal
      ? section.subItems.filter((s) => s.assigneeId === filterUserId)
      : section.subItems
    if (subs.length === 0) continue

    md += `### ${section.mainItem.title}\n`
    md += `进度：${section.mainItem.completion}%\n\n`
    for (const sub of subs) {
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

  const now = new Date().toISOString().slice(0, 10)
  md += `---\n导出时间 ${now} by ${user?.displayName || ''}\n`
  return md
}

function downloadMarkdown(content: string, filename: string) {
  const blob = new Blob([content], { type: 'text/markdown' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

export default function ReportPage() {
  const currentTeamId = useTeamStore((s) => s.currentTeamId)
  const currentUser = useAuthStore((s) => s.user)
  const [weekValue, setWeekValue] = useState(getCurrentWeekStart)
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
      const resp = await getWeeklyReportPreviewApi(currentTeamId, weekValue)
      setPreview(resp)
    } catch (err: any) {
      setError(err?.response?.data?.message || '获取预览失败')
    } finally {
      setLoading(false)
    }
  }

  const handleExportFull = () => {
    if (!preview) return
    downloadMarkdown(renderMarkdown(preview), `weekly-report-${weekValue}.md`)
  }

  const handleExportPersonal = () => {
    if (!preview || !currentUser) return
    downloadMarkdown(
      renderMarkdown(preview, currentUser.id),
      `weekly-report-${weekValue}-${currentUser.username}.md`,
    )
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
            <label className="block text-sm font-medium text-secondary mb-1.5">
              选择周次
            </label>
            <WeekPicker weekStart={weekValue} onChange={setWeekValue} />
          </div>
          <Button size="sm" onClick={handlePreview} disabled={loading}>
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
              <div className="flex gap-2">
                <Button variant="secondary" size="sm" onClick={handleExportPersonal} data-testid="export-personal-btn">
                  导出个人周报
                </Button>
                <Button variant="primary" size="sm" onClick={handleExportFull} data-testid="export-full-btn">
                  导出完整周报
                </Button>
              </div>
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
