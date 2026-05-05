import { Routes, Route, Navigate } from "react-router-dom";
import ProtectedRoute from "@/components/ProtectedRoute";
import PermissionRoute from "@/components/PermissionRoute";
import AppLayout from "@/components/layout/AppLayout";
import LoginPage from "@/pages/LoginPage";
import ItemViewPage from "@/pages/ItemViewPage";
import MainItemDetailPage from "@/pages/MainItemDetailPage";
import SubItemDetailPage from "@/pages/SubItemDetailPage";
import WeeklyViewPage from "@/pages/WeeklyViewPage";
import GanttViewPage from "@/pages/GanttViewPage";
import TableViewPage from "@/pages/TableViewPage";
import ItemPoolPage from "@/pages/ItemPoolPage";
import ReportPage from "@/pages/ReportPage";
import TeamManagementPage from "@/pages/TeamManagementPage";
import TeamDetailPage from "@/pages/TeamDetailPage";
import UserManagementPage from "@/pages/UserManagementPage";
import RoleManagementPage from "@/pages/RoleManagementPage";

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<Navigate replace to="/items" />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route path="/items" element={<ItemViewPage />} />
          <Route path="/items/:mainItemId" element={<MainItemDetailPage />} />
          <Route
            path="/items/:mainItemId/sub/:subItemId"
            element={<SubItemDetailPage />}
          />
          <Route path="/weekly" element={<WeeklyViewPage />} />
          <Route path="/gantt" element={<GanttViewPage />} />
          <Route path="/table" element={<TableViewPage />} />
          <Route path="/item-pool" element={<ItemPoolPage />} />
          <Route path="/report" element={<ReportPage />} />
          <Route path="/teams/:teamId" element={<TeamDetailPage />} />
          <Route path="/teams" element={<TeamManagementPage />} />
        </Route>
      </Route>
      <Route element={<PermissionRoute code="user:read" />}>
        <Route element={<AppLayout />}>
          <Route path="/users" element={<UserManagementPage />} />
        </Route>
      </Route>
      <Route element={<PermissionRoute code="user:manage_role" />}>
        <Route element={<AppLayout />}>
          <Route path="/roles" element={<RoleManagementPage />} />
        </Route>
      </Route>
    </Routes>
  );
}
