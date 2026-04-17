import { Routes, Route, Navigate } from 'react-router-dom'
import ProtectedRoute from '@/components/ProtectedRoute'
import AdminRoute from '@/components/AdminRoute'
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

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<Navigate replace to="/items" />} />
      <Route element={<ProtectedRoute />}>
        <Route path="/items" element={<ItemViewPage />} />
        <Route path="/items/:mainItemId" element={<MainItemDetailPage />} />
        <Route path="/items/:mainItemId/sub/:subItemId" element={<SubItemDetailPage />} />
        <Route path="/weekly" element={<WeeklyViewPage />} />
        <Route path="/gantt" element={<GanttViewPage />} />
        <Route path="/table" element={<TableViewPage />} />
        <Route path="/item-pool" element={<ItemPoolPage />} />
        <Route path="/report" element={<ReportPage />} />
      </Route>
      <Route element={<AdminRoute />}>
        <Route path="/admin" element={<AdminPage />} />
      </Route>
    </Routes>
  )
}
