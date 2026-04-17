import client from './client'
import type {
  CreateSubItemReq,
  UpdateSubItemReq,
  ChangeStatusReq,
  AssignSubItemReq,
  SubItemFilter,
  SubItem,
  PageResult,
} from '@/types'

export function createSubItemApi(teamId: number, mainId: number, req: CreateSubItemReq): Promise<SubItem> {
  return client.post<never, SubItem>(`/teams/${teamId}/main-items/${mainId}/sub-items`, req)
}

export function listSubItemsApi(teamId: number, mainId: number, filter?: SubItemFilter): Promise<PageResult<SubItem>> {
  return client.get<never, PageResult<SubItem>>(`/teams/${teamId}/main-items/${mainId}/sub-items`, { params: filter })
}

export function getSubItemApi(teamId: number, itemId: number): Promise<SubItem> {
  return client.get<never, SubItem>(`/teams/${teamId}/sub-items/${itemId}`)
}

export function updateSubItemApi(teamId: number, itemId: number, req: UpdateSubItemReq): Promise<SubItem> {
  return client.put<never, SubItem>(`/teams/${teamId}/sub-items/${itemId}`, req)
}

export function changeSubItemStatusApi(teamId: number, itemId: number, req: ChangeStatusReq): Promise<void> {
  return client.put<never, void>(`/teams/${teamId}/sub-items/${itemId}/status`, req)
}

export function assignSubItemApi(teamId: number, itemId: number, req: AssignSubItemReq): Promise<void> {
  return client.put<never, void>(`/teams/${teamId}/sub-items/${itemId}/assignee`, req)
}
