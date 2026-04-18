import client from './client'
import type {
  AdminUser,
  SetCanCreateTeamReq,
  AdminTeam,
  PageResult,
  CreateUserReq,
  CreateUserResp,
  UpdateUserReq,
  UpdateUserResp,
  ToggleUserStatusReq,
  ToggleUserStatusResp,
  GetUserResp,
} from '@/types'

export function listUsersApi(params?: { page?: number; pageSize?: number; search?: string; canCreateTeam?: string }): Promise<PageResult<AdminUser>> {
  return client.get<never, PageResult<AdminUser>>('/admin/users', { params })
}

export function setCanCreateTeamApi(userId: number, req: SetCanCreateTeamReq): Promise<void> {
  return client.put<never, void>(`/admin/users/${userId}/can-create-team`, req)
}

export function listAdminTeamsApi(page?: number, pageSize?: number): Promise<PageResult<AdminTeam>> {
  return client.get<never, PageResult<AdminTeam>>('/admin/teams', { params: { page, pageSize } })
}

export function createUserApi(req: CreateUserReq): Promise<CreateUserResp> {
  return client.post<never, CreateUserResp>('/admin/users', req)
}

export function getUserApi(userId: number): Promise<GetUserResp> {
  return client.get<never, GetUserResp>(`/admin/users/${userId}`)
}

export function updateUserApi(userId: number, req: UpdateUserReq): Promise<UpdateUserResp> {
  return client.put<never, UpdateUserResp>(`/admin/users/${userId}`, req)
}

export function toggleUserStatusApi(userId: number, req: ToggleUserStatusReq): Promise<ToggleUserStatusResp> {
  return client.put<never, ToggleUserStatusResp>(`/admin/users/${userId}/status`, req)
}
