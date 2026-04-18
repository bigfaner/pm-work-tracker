import client from './client'
import type { AppendProgressReq, CorrectCompletionReq, ProgressRecord } from '@/types'

export function appendProgressApi(teamId: number, subItemId: number, req: AppendProgressReq): Promise<ProgressRecord> {
  return client.post<never, ProgressRecord>(`/teams/${teamId}/sub-items/${subItemId}/progress`, req)
}

export function listProgressApi(teamId: number, subItemId: number): Promise<ProgressRecord[]> {
  return client.get<never, ProgressRecord[]>(`/teams/${teamId}/sub-items/${subItemId}/progress`)
}

export function correctCompletionApi(teamId: number, recordId: number, req: CorrectCompletionReq): Promise<ProgressRecord> {
  return client.patch<never, ProgressRecord>(`/teams/${teamId}/progress/${recordId}/completion`, req)
}
