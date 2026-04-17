import client from './client'
import type { ReportPreviewResp } from '@/types'

export function getWeeklyReportPreviewApi(teamId: number, weekStart: string): Promise<ReportPreviewResp> {
  return client.get<never, ReportPreviewResp>(`/teams/${teamId}/reports/weekly/preview`, { params: { weekStart } })
}

export function exportWeeklyReportApi(teamId: number, weekStart: string): Promise<Blob> {
  return client.get<never, Blob>(`/teams/${teamId}/reports/weekly/export`, {
    params: { weekStart },
    responseType: 'blob',
  })
}
