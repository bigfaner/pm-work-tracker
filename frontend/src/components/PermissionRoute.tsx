import { Navigate, Outlet } from "react-router-dom";
import { useAuthStore } from "@/store/auth";
import { usePermission } from "@/hooks/usePermission";

interface PermissionRouteProps {
  code: string;
}

export default function PermissionRoute({ code }: PermissionRouteProps) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const _hasHydrated = useAuthStore((s) => s._hasHydrated);
  const hasPermission = usePermission(code);

  if (!_hasHydrated) return null;

  if (!isAuthenticated) {
    return <Navigate replace to="/login" />;
  }

  if (!hasPermission) {
    return <Navigate replace to="/" />;
  }

  return <Outlet />;
}
