import client from './client'
import type {
  AdminUser,
  AdminTeam,
  PageResult,
  CreateUserReq,
  CreateUserResp,
  UpdateUserReq,
  UpdateUserResp,
  ToggleUserStatusReq,
  ToggleUserStatusResp,
  GetUserResp,
  ResetPasswordReq,
  ResetPasswordResp,
} from '@/types'

export function listUsersApi(params?: { page?: number; pageSize?: number; search?: string }): Promise<PageResult<AdminUser>> {
  return client.get<never, PageResult<AdminUser>>('/admin/users', { params })
}

export function listAdminTeamsApi(page?: number, pageSize?: number): Promise<PageResult<AdminTeam>> {
  return client.get<never, PageResult<AdminTeam>>('/admin/teams', { params: { page, pageSize } })
}

export function createUserApi(req: CreateUserReq): Promise<CreateUserResp> {
  return client.post<never, CreateUserResp>('/admin/users', req)
}

export function getUserApi(userId: string): Promise<GetUserResp> {
  return client.get<never, GetUserResp>(`/admin/users/${userId}`)
}

export function updateUserApi(userId: string, req: UpdateUserReq): Promise<UpdateUserResp> {
  return client.put<never, UpdateUserResp>(`/admin/users/${userId}`, req)
}

export function toggleUserStatusApi(userId: string, req: ToggleUserStatusReq): Promise<ToggleUserStatusResp> {
  return client.put<never, ToggleUserStatusResp>(`/admin/users/${userId}/status`, req)
}

export function resetPasswordApi(userId: string, req: ResetPasswordReq): Promise<ResetPasswordResp> {
  return client.put<never, ResetPasswordResp>(`/admin/users/${userId}/password`, req)
}

export function deleteUserApi(userId: string): Promise<void> {
  return client.delete<never, void>(`/admin/users/${userId}`)
}
