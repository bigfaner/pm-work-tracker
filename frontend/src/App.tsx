import { Routes, Route, Navigate } from 'react-router-dom'
import ProtectedRoute from '@/components/ProtectedRoute'
import AdminRoute from '@/components/AdminRoute'
import AppLayout from '@/components/layout/AppLayout'
import LoginPage from '@/pages/LoginPage'
import ItemViewPage from '@/pages/ItemViewPage'
import MainItemDetailPage from '@/pages/MainItemDetailPage'
import SubItemDetailPage from '@/pages/SubItemDetailPage'
import WeeklyViewPage from '@/pages/WeeklyViewPage'
import GanttViewPage from '@/pages/GanttViewPage'
import TableViewPage from '@/pages/TableViewPage'
import ItemPoolPage from '@/pages/ItemPoolPage'
import ReportPage from '@/pages/ReportPage'
import AdminPage from '@/pages/AdminPage'
import TeamManagementPage from '@/pages/TeamManagementPage'
import UserManagementPage from '@/pages/UserManagementPage'

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<Navigate replace to="/items" />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<AppLayout />}>
          <Route path="/items" element={<ItemViewPage />} />
          <Route path="/items/:mainItemId" element={<MainItemDetailPage />} />
          <Route path="/items/:mainItemId/sub/:subItemId" element={<SubItemDetailPage />} />
          <Route path="/weekly" element={<WeeklyViewPage />} />
          <Route path="/gantt" element={<GanttViewPage />} />
          <Route path="/table" element={<TableViewPage />} />
          <Route path="/item-pool" element={<ItemPoolPage />} />
          <Route path="/report" element={<ReportPage />} />
          <Route path="/teams/:teamId" element={<TeamManagementPage />} />
          <Route path="/teams" element={<AdminPage />} />
        </Route>
      </Route>
      <Route element={<AdminRoute />}>
        <Route element={<AppLayout />}>
          <Route path="/users" element={<UserManagementPage />} />
        </Route>
      </Route>
    </Routes>
  )
}
