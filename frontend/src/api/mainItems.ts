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
  return client.post<never, MainItem>(`/teams/${teamId}/main-items`, req)
}

export function listMainItemsApi(teamId: number, filter?: MainItemFilter): Promise<PageResult<MainItem>> {
  return client.get<never, PageResult<MainItem>>(`/teams/${teamId}/main-items`, { params: filter })
}

export function getMainItemApi(teamId: number, itemId: number): Promise<MainItem & { subItems: SubItem[] }> {
  return client.get<never, MainItem & { subItems: SubItem[] }>(`/teams/${teamId}/main-items/${itemId}`)
}

export function updateMainItemApi(teamId: number, itemId: number, req: UpdateMainItemReq): Promise<MainItem> {
  return client.put<never, MainItem>(`/teams/${teamId}/main-items/${itemId}`, req)
}

export function archiveMainItemApi(teamId: number, itemId: number): Promise<void> {
  return client.post<never, void>(`/teams/${teamId}/main-items/${itemId}/archive`)
}
