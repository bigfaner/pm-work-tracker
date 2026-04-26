import client from './client'
import type {
  CreateSubItemReq,
  UpdateSubItemReq,
  ChangeStatusReq,
  SubItemFilter,
  SubItem,
  PageResult,
} from '@/types'

export function createSubItemApi(teamBizKey: string, mainBizKey: string, req: CreateSubItemReq): Promise<SubItem> {
  return client.post<never, SubItem>(`/teams/${teamBizKey}/main-items/${mainBizKey}/sub-items`, {
    ...req,
    mainItemBizKey: mainBizKey,
  })
}

export function listSubItemsApi(teamBizKey: string, mainBizKey: string, filter?: SubItemFilter): Promise<PageResult<SubItem>> {
  return client.get<never, PageResult<SubItem>>(`/teams/${teamBizKey}/main-items/${mainBizKey}/sub-items`, { params: filter })
}

export function getSubItemApi(teamBizKey: string, bizKey: string): Promise<SubItem> {
  return client.get<never, SubItem>(`/teams/${teamBizKey}/sub-items/${bizKey}`)
}

export function updateSubItemApi(teamBizKey: string, bizKey: string, req: UpdateSubItemReq): Promise<SubItem> {
  return client.put<never, SubItem>(`/teams/${teamBizKey}/sub-items/${bizKey}`, req)
}

export function changeSubItemStatusApi(teamBizKey: string, bizKey: string, req: ChangeStatusReq): Promise<void> {
  return client.put<never, void>(`/teams/${teamBizKey}/sub-items/${bizKey}/status`, req)
}

export function getSubItemTransitionsApi(teamBizKey: string, bizKey: string): Promise<string[]> {
  return client.get<never, { transitions: string[] }>(`/teams/${teamBizKey}/sub-items/${bizKey}/available-transitions`)
    .then((res) => res.transitions ?? [])
}
