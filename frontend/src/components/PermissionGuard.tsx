import type { ReactNode } from 'react'
import { usePermission } from '@/hooks/usePermission'

interface PermissionGuardProps {
  code: string
  teamId?: number
  children: ReactNode
}

export function PermissionGuard({ code, teamId, children }: PermissionGuardProps) {
  const allowed = usePermission(code, teamId)
  if (!allowed) return null
  return <>{children}</>
}
