import client from './client'
import type {
  TableFilter,
  WeeklyViewResponse,
  GanttViewResp,
  TableRow,
  PageResult,
} from '@/types'

export function getWeeklyViewApi(
  teamBizKey: string,
  weekStart: string,
): Promise<WeeklyViewResponse> {
  return client.get<never, WeeklyViewResponse>(
    `/teams/${teamBizKey}/views/weekly`,
    { params: { weekStart } },
  )
}

export function getGanttViewApi(
  teamBizKey: string,
  statuses?: string[],
): Promise<GanttViewResp> {
  const params: Record<string, string | string[] | undefined> = {}
  if (statuses && statuses.length > 0) {
    params.status = statuses
  }
  return client.get<never, GanttViewResp>(`/teams/${teamBizKey}/views/gantt`, {
    params,
    paramsSerializer: {
      serialize: (p) => {
        const parts: string[] = []
        for (const [key, value] of Object.entries(p)) {
          if (value == null) continue
          const values = Array.isArray(value) ? value : [value]
          for (const v of values) {
            parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(v)}`)
          }
        }
        return parts.join('&')
      },
    },
  })
}

export function getTableViewApi(
  teamBizKey: string,
  filter?: TableFilter,
): Promise<PageResult<TableRow>> {
  return client.get<never, PageResult<TableRow>>(
    `/teams/${teamBizKey}/views/table`,
    { params: filter },
  )
}

export function exportTableCsvApi(
  teamBizKey: string,
  filter?: TableFilter,
): Promise<Blob> {
  return client.get<never, Blob>(`/teams/${teamBizKey}/views/table/export`, {
    params: filter,
    responseType: 'blob',
  })
}
