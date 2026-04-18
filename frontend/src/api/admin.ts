import client from './client'
import type { AdminUser, SetCanCreateTeamReq, AdminTeam, PageResult } from '@/types'

export function listUsersApi(page?: number, pageSize?: number): Promise<PageResult<AdminUser>> {
  return client.get<never, PageResult<AdminUser>>('/admin/users', { params: { page, pageSize } })
}

export function setCanCreateTeamApi(userId: number, req: SetCanCreateTeamReq): Promise<void> {
  return client.put<never, void>(`/admin/users/${userId}/can-create-team`, req)
}

export function listAdminTeamsApi(page?: number, pageSize?: number): Promise<PageResult<AdminTeam>> {
  return client.get<never, PageResult<AdminTeam>>('/admin/teams', { params: { page, pageSize } })
}
