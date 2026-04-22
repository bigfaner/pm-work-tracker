import client from './client'
import type {
  CreateMainItemReq,
  UpdateMainItemReq,
  ChangeStatusReq,
  MainItemFilter,
  MainItem,
  SubItem,
  PageResult,
} from '@/types'

export function createMainItemApi(teamId: number, req: CreateMainItemReq): Promise<MainItem> {
  return client.post<never, MainItem>(`/teams/${teamId}/main-items`, req)
}

export function listMainItemsApi(teamId: number, filter?: MainItemFilter): Promise<PageResult<MainItem>> {
  return client.get<never, PageResult<MainItem>>(`/teams/${teamId}/main-items`, { params: filter })
}

export type MainItemDetailResp = MainItem & {
  subItems: SubItem[]
  achievements?: string[]
  blockers?: string[]
}

export function getMainItemApi(teamId: number, itemId: number): Promise<MainItemDetailResp> {
  return client.get<never, MainItemDetailResp>(`/teams/${teamId}/main-items/${itemId}`)
}

export function updateMainItemApi(teamId: number, itemId: number, req: UpdateMainItemReq): Promise<MainItem> {
  return client.put<never, MainItem>(`/teams/${teamId}/main-items/${itemId}`, req)
}

export function changeMainItemStatusApi(teamId: number, itemId: number, req: ChangeStatusReq): Promise<MainItem> {
  return client.put<never, MainItem>(`/teams/${teamId}/main-items/${itemId}/status`, req)
}

export function getMainItemTransitionsApi(teamId: number, itemId: number): Promise<string[]> {
  return client.get<never, string[]>(`/teams/${teamId}/main-items/${itemId}/available-transitions`)
}
