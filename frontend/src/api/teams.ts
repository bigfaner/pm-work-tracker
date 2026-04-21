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

export function listTeamsApi(): Promise<Team[]> {
  return client.get<never, Team[]>('/teams')
}

export function getTeamApi(teamId: number): Promise<TeamDetailResp> {
  return client.get<never, TeamDetailResp>(`/teams/${teamId}`)
}

export function updateTeamApi(teamId: number, req: UpdateTeamReq): Promise<Team> {
  return client.put<never, Team>(`/teams/${teamId}`, req)
}

export function deleteTeamApi(teamId: number, req: DeleteTeamReq): Promise<void> {
  return client.delete<never, void>(`/teams/${teamId}`, { data: req })
}

export function listMembersApi(teamId: number): Promise<TeamMemberResp[]> {
  return client.get<never, TeamMemberResp[]>(`/teams/${teamId}/members`)
}

export function inviteMemberApi(teamId: number, req: InviteMemberReq): Promise<void> {
  return client.post<never, void>(`/teams/${teamId}/members`, req)
}

export function removeMemberApi(teamId: number, userId: number): Promise<void> {
  return client.delete<never, void>(`/teams/${teamId}/members/${userId}`)
}

export function transferPmApi(teamId: number, req: TransferPMReq): Promise<void> {
  return client.put<never, void>(`/teams/${teamId}/pm`, req)
}

export function changeMemberRoleApi(teamId: number, memberId: number, req: { roleId: number }): Promise<void> {
  return client.put<never, void>(`/teams/${teamId}/members/${memberId}/role`, req)
}

export interface UserSearchResult {
  id: number
  username: string
  displayName: string
}

export function searchAvailableUsersApi(teamId: number, search?: string): Promise<UserSearchResult[]> {
  return client.get<never, UserSearchResult[]>(`/teams/${teamId}/search-users`, { params: { search } })
}
