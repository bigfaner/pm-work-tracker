import { useAuthStore } from '@/store/auth'

export function usePermission(code: string, teamId?: number): boolean {
  return useAuthStore((s) => s.hasPermission(code, teamId))
}
