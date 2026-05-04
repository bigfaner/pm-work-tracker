import client from "./client";
import type {
  Role,
  RoleDetail,
  CreateRoleReq,
  UpdateRoleReq,
  RoleListParams,
  PageResult,
  PermissionGroup,
} from "@/types";

export function listRolesApi(
  params?: RoleListParams,
): Promise<PageResult<Role>> {
  return client.get<never, PageResult<Role>>("/admin/roles", { params });
}

export function getRoleApi(bizKey: string): Promise<RoleDetail> {
  return client.get<never, RoleDetail>(`/admin/roles/${bizKey}`);
}

export function createRoleApi(req: CreateRoleReq): Promise<Role> {
  return client.post<never, Role>("/admin/roles", req);
}

export function updateRoleApi(
  bizKey: string,
  req: UpdateRoleReq,
): Promise<RoleDetail> {
  return client.put<never, RoleDetail>(`/admin/roles/${bizKey}`, req);
}

export function deleteRoleApi(bizKey: string): Promise<void> {
  return client.delete<never, void>(`/admin/roles/${bizKey}`);
}

export function listPermissionsApi(): Promise<PermissionGroup[]> {
  return client.get<never, PermissionGroup[]>("/admin/permissions");
}
