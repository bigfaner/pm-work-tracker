import client from './client'
import type { PermissionData } from '@/types'

export function getPermissionsApi(): Promise<PermissionData> {
  return client.get<never, PermissionData>('/me/permissions')
}
