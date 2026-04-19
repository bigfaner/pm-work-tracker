import client from './client'
import type {
  CreateMainItemReq,
  UpdateMainItemReq,
  MainItemFilter,
  MainItem,
  SubItem,
  PageResult,
} from '@/types'

export function createMainItemApi(teamId: number, req: CreateMainItemReq): Promise<MainItem> {
  const body: Record<string, unknown> = {
    title: req.title,
    priority: req.priority,
    assignee_id: req.assigneeId,
    start_date: req.startDate,
    expected_end_date: req.expectedEndDate,
  }
  if (req.description) body.description = req.description
  return client.post<never, MainItem>(`/teams/${teamId}/main-items`, body)
}

export function listMainItemsApi(teamId: number, filter?: MainItemFilter): Promise<PageResult<MainItem>> {
  return client.get<never, PageResult<MainItem>>(`/teams/${teamId}/main-items`, { params: filter })
}

export function getMainItemApi(teamId: number, itemId: number): Promise<MainItem & { subItems: SubItem[] }> {
  return client.get<never, MainItem & { subItems: SubItem[] }>(`/teams/${teamId}/main-items/${itemId}`)
}

export function updateMainItemApi(teamId: number, itemId: number, req: UpdateMainItemReq): Promise<MainItem> {
  const body: Record<string, unknown> = {}
  if (req.title !== undefined) body.title = req.title
  if (req.priority !== undefined) body.priority = req.priority
  if (req.assigneeId !== undefined) body.assignee_id = req.assigneeId
  if (req.startDate !== undefined) body.start_date = req.startDate
  if (req.expectedEndDate !== undefined) body.expected_end_date = req.expectedEndDate
  if (req.actualEndDate !== undefined) body.actual_end_date = req.actualEndDate
  if (req.status !== undefined) body.status = req.status
  return client.put<never, MainItem>(`/teams/${teamId}/main-items/${itemId}`, body)
}

export function archiveMainItemApi(teamId: number, itemId: number): Promise<void> {
  return client.post<never, void>(`/teams/${teamId}/main-items/${itemId}/archive`)
}
