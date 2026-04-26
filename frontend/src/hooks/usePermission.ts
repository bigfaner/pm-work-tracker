import { useAuthStore } from '@/store/auth'

export function usePermission(code: string, teamId?: string): boolean {
  return useAuthStore((s) => s.hasPermission(code, teamId))
}
