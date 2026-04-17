import client from './client'
import type {
  SubmitItemPoolReq,
  AssignItemPoolReq,
  RejectItemPoolReq,
  ItemPoolFilter,
  ItemPool,
  AssignItemPoolResp,
  PageResult,
} from '@/types'

export function submitItemPoolApi(teamId: number, req: SubmitItemPoolReq): Promise<ItemPool> {
  return client.post<never, ItemPool>(`/teams/${teamId}/item-pool`, req)
}

export function listItemPoolApi(teamId: number, filter?: ItemPoolFilter): Promise<PageResult<ItemPool>> {
  return client.get<never, PageResult<ItemPool>>(`/teams/${teamId}/item-pool`, { params: filter })
}

export function getItemPoolApi(teamId: number, poolId: number): Promise<ItemPool> {
  return client.get<never, ItemPool>(`/teams/${teamId}/item-pool/${poolId}`)
}

export function assignItemPoolApi(teamId: number, poolId: number, req: AssignItemPoolReq): Promise<AssignItemPoolResp> {
  return client.post<never, AssignItemPoolResp>(`/teams/${teamId}/item-pool/${poolId}/assign`, req)
}

export function rejectItemPoolApi(teamId: number, poolId: number, req: RejectItemPoolReq): Promise<ItemPool> {
  return client.post<never, ItemPool>(`/teams/${teamId}/item-pool/${poolId}/reject`, req)
}
