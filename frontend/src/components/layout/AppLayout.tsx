import { useEffect } from 'react'
import { Outlet } from 'react-router-dom'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
import { listTeamsApi } from '@/api/teams'
import Sidebar from './Sidebar'

export default function AppLayout() {
  const { teams, currentTeamId, setCurrentTeam, setTeams } = useTeamStore()
  const fetchPermissions = useAuthStore((s) => s.fetchPermissions)

  useEffect(() => {
    listTeamsApi()
      .then((data) => {
        setTeams(data)
        if (data.length > 0 && !currentTeamId) {
          setCurrentTeam(data[0].id)
        }
      })
      .catch(() => {})
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    fetchPermissions()
  }, [fetchPermissions])

  return (
    <div data-testid="app-layout" className="flex min-h-screen">
      <Sidebar />
      <main
        data-testid="content-area"
        className="ml-[240px] flex-1 p-6 bg-bg-alt min-h-screen overflow-auto"
      >
        <Outlet />
      </main>
    </div>
  )
}
