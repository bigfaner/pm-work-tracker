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

export function createMainItemApi(
  teamBizKey: string,
  req: CreateMainItemReq,
): Promise<MainItem> {
  return client.post<never, MainItem>(`/teams/${teamBizKey}/main-items`, req)
}

// serializeParams sends array values as repeated keys (?status=a&status=b)
// so Gin binds []string correctly instead of treating comma-joined as one value.
const serializeParams = (p: Record<string, unknown>) => {
  const parts: string[] = []
  for (const [key, value] of Object.entries(p)) {
    if (value == null) continue
    const values = Array.isArray(value) ? value : [value]
    for (const v of values) {
      parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(v))}`)
    }
  }
  return parts.join('&')
}

export function listMainItemsApi(
  teamBizKey: string,
  filter?: MainItemFilter,
): Promise<PageResult<MainItem>> {
  const params: Record<string, unknown> = { ...filter }
  if (Array.isArray(params.status) && params.status.length === 0) {
    delete params.status
  }
  return client.get<never, PageResult<MainItem>>(
    `/teams/${teamBizKey}/main-items`,
    { params, paramsSerializer: { serialize: serializeParams } },
  )
}

export type MainItemDetailResp = MainItem & {
  subItems: SubItem[]
  achievements?: string[]
  blockers?: string[]
};

export function getMainItemApi(
  teamBizKey: string,
  bizKey: string,
): Promise<MainItemDetailResp> {
  return client.get<never, MainItemDetailResp>(
    `/teams/${teamBizKey}/main-items/${bizKey}`,
  )
}

export function updateMainItemApi(
  teamBizKey: string,
  bizKey: string,
  req: UpdateMainItemReq,
): Promise<MainItem> {
  return client.put<never, MainItem>(
    `/teams/${teamBizKey}/main-items/${bizKey}`,
    req,
  )
}

export function changeMainItemStatusApi(
  teamBizKey: string,
  bizKey: string,
  req: ChangeStatusReq,
): Promise<MainItem> {
  return client.put<never, MainItem>(
    `/teams/${teamBizKey}/main-items/${bizKey}/status`,
    req,
  )
}

export function getMainItemTransitionsApi(
  teamBizKey: string,
  bizKey: string,
): Promise<string[]> {
  return client
    .get<
      never,
      { transitions: string[] }
    >(`/teams/${teamBizKey}/main-items/${bizKey}/available-transitions`)
    .then((res) => res.transitions ?? [])
}

export function deleteMainItemApi(
  teamBizKey: string,
  bizKey: string,
): Promise<{ message: string }> {
  return client.delete<never, { message: string }>(
    `/teams/${teamBizKey}/main-items/${bizKey}`,
  )
}
