import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'

export default function AdminRoute() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const isSuperAdmin = useAuthStore((s) => s.isSuperAdmin)

  if (!isAuthenticated) {
    return <Navigate replace to="/login" />
  }

  if (!isSuperAdmin) {
    return <Navigate replace to="/" />
  }

  return <Outlet />
}
