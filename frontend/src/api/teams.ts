import client from './client'
import type {
  CreateTeamReq,
  UpdateTeamReq,
  DeleteTeamReq,
  InviteMemberReq,
  TransferPMReq,
  Team,
  TeamDetailResp,
  TeamMemberResp,
} from '@/types'

export function createTeamApi(req: CreateTeamReq): Promise<Team> {
  return client.post<never, Team>('/teams', req)
}

export interface TeamListParams {
  search?: string
  page?: number
  pageSize?: number
}

export interface TeamListResp {
  items: Team[]
  total: number
  page: number
  pageSize: number
}

export function listTeamsApi(params?: TeamListParams): Promise<TeamListResp> {
  return client.get<never, TeamListResp>('/teams', { params })
}

export function getTeamApi(bizKey: string): Promise<TeamDetailResp> {
  return client.get<never, TeamDetailResp>(`/teams/${bizKey}`)
}

export function updateTeamApi(bizKey: string, req: UpdateTeamReq): Promise<Team> {
  return client.put<never, Team>(`/teams/${bizKey}`, req)
}

export function deleteTeamApi(bizKey: string, req: DeleteTeamReq): Promise<void> {
  return client.delete<never, void>(`/teams/${bizKey}`, { data: req })
}

export function listMembersApi(bizKey: string): Promise<TeamMemberResp[]> {
  return client.get<never, TeamMemberResp[]>(`/teams/${bizKey}/members`)
}

export function inviteMemberApi(bizKey: string, req: InviteMemberReq): Promise<void> {
  return client.post<never, void>(`/teams/${bizKey}/members`, req)
}

export function removeMemberApi(bizKey: string, userKey: string): Promise<void> {
  return client.delete<never, void>(`/teams/${bizKey}/members/${userKey}`)
}

export function transferPmApi(bizKey: string, req: TransferPMReq): Promise<void> {
  return client.put<never, void>(`/teams/${bizKey}/pm`, req)
}

export function changeMemberRoleApi(bizKey: string, userKey: string, req: { roleKey: string }): Promise<void> {
  return client.put<never, void>(`/teams/${bizKey}/members/${userKey}/role`, req)
}

export interface UserSearchResult {
  bizKey: string
  username: string
  displayName: string
}

export function searchAvailableUsersApi(bizKey: string, search?: string): Promise<UserSearchResult[]> {
  return client.get<never, UserSearchResult[]>(`/teams/${bizKey}/search-users`, { params: { search } })
}
