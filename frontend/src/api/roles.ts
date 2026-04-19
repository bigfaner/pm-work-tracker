import client from './client'
import type { Role, RoleDetail, CreateRoleReq, UpdateRoleReq, RoleListParams, PageResult, PermissionGroup } from '@/types'

export function listRolesApi(params?: RoleListParams): Promise<PageResult<Role>> {
  return client.get<never, PageResult<Role>>('/admin/roles', { params })
}

export function getRoleApi(id: number): Promise<RoleDetail> {
  return client.get<never, RoleDetail>(`/admin/roles/${id}`)
}

export function createRoleApi(req: CreateRoleReq): Promise<Role> {
  return client.post<never, Role>('/admin/roles', req)
}

export function updateRoleApi(id: number, req: UpdateRoleReq): Promise<RoleDetail> {
  return client.put<never, RoleDetail>(`/admin/roles/${id}`, req)
}

export function deleteRoleApi(id: number): Promise<void> {
  return client.delete<never, void>(`/admin/roles/${id}`)
}

export function listPermissionsApi(): Promise<PermissionGroup[]> {
  return client.get<never, PermissionGroup[]>('/admin/permissions')
}
