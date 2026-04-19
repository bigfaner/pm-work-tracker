import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'

export default function AdminRoute() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const isSuperAdmin = useAuthStore((s) => s.isSuperAdmin)
  const _hasHydrated = useAuthStore((s) => s._hasHydrated)

  if (!_hasHydrated) return null

  if (!isAuthenticated) {
    return <Navigate replace to="/login" />
  }

  if (!isSuperAdmin) {
    return <Navigate replace to="/" />
  }

  return <Outlet />
}
