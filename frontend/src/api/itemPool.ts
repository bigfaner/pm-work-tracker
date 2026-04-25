import client from './client'
import type {
  SubmitItemPoolReq,
  AssignItemPoolReq,
  ConvertToMainItemReq,
  RejectItemPoolReq,
  ItemPoolFilter,
  ItemPool,
  AssignItemPoolResp,
  PageResult,
} from '@/types'

export function submitItemPoolApi(teamBizKey: string, req: SubmitItemPoolReq): Promise<ItemPool> {
  return client.post<never, ItemPool>(`/teams/${teamBizKey}/item-pool`, req)
}

export function listItemPoolApi(teamBizKey: string, filter?: ItemPoolFilter): Promise<PageResult<ItemPool>> {
  return client.get<never, PageResult<ItemPool>>(`/teams/${teamBizKey}/item-pool`, { params: filter })
}

export function assignItemPoolApi(teamBizKey: string, poolBizKey: string, req: AssignItemPoolReq): Promise<AssignItemPoolResp> {
  return client.post<never, AssignItemPoolResp>(`/teams/${teamBizKey}/item-pool/${poolBizKey}/assign`, req)
}

export function convertToMainApi(teamBizKey: string, poolBizKey: string, req: ConvertToMainItemReq): Promise<{ mainItemId: number }> {
  return client.post<never, { mainItemId: number }>(`/teams/${teamBizKey}/item-pool/${poolBizKey}/convert-to-main`, req)
}

export function rejectItemPoolApi(teamBizKey: string, poolBizKey: string, req: RejectItemPoolReq): Promise<ItemPool> {
  return client.post<never, ItemPool>(`/teams/${teamBizKey}/item-pool/${poolBizKey}/reject`, req)
}
