import client from './client'
import type { TableFilter, WeeklyViewResp, WeeklyViewResponse, GanttViewResp, TableRow, PageResult } from '@/types'

export function getWeeklyViewApi(teamId: number, weekStart: string): Promise<WeeklyViewResponse> {
  return client.get<never, WeeklyViewResponse>(`/teams/${teamId}/views/weekly`, { params: { weekStart } })
}

export function getGanttViewApi(teamId: number, status?: string): Promise<GanttViewResp> {
  return client.get<never, GanttViewResp>(`/teams/${teamId}/views/gantt`, { params: { status } })
}

export function getTableViewApi(teamId: number, filter?: TableFilter): Promise<PageResult<TableRow>> {
  return client.get<never, PageResult<TableRow>>(`/teams/${teamId}/views/table`, { params: filter })
}

export function exportTableCsvApi(teamId: number, filter?: TableFilter): Promise<Blob> {
  return client.get<never, Blob>(`/teams/${teamId}/views/table/export`, {
    params: filter,
    responseType: 'blob',
  })
}
