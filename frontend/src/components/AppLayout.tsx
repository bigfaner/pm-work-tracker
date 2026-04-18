import { useState, useEffect } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import { listTeamsApi } from '@/api/teams'

const navItems = [
  { key: '/items', label: '事项视图' },
  { key: '/weekly', label: '周视图' },
  { key: '/gantt', label: '甘特图' },
  { key: '/table', label: '表格视图' },
  { key: '/item-pool', label: '事项池' },
  { key: '/report', label: '周报导出' },
]

export default function AppLayout() {
  const navigate = useNavigate()
  const location = useLocation()

  const { user, clearAuth, isSuperAdmin } = useAuthStore()
  const { teams, currentTeamId, setCurrentTeam, setTeams } = useTeamStore()

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

  const activeKey = '/' + location.pathname.split('/').filter(Boolean)[0]

  const handleLogout = () => {
    clearAuth()
    navigate('/login')
  }

  const allNavItems = isSuperAdmin
    ? [...navItems, { key: '/admin', label: '管理后台' }]
    : navItems

  return (
    <div data-testid="app-layout" style={{ display: 'flex', minHeight: '100vh' }}>
      <div data-testid="sidebar" style={{ width: 220, position: 'fixed', left: 0, top: 0, bottom: 0, background: '#001529', color: '#fff', display: 'flex', flexDirection: 'column', overflow: 'auto' }}>
        <div style={{ padding: '16px 12px 8px', textAlign: 'center', fontWeight: 700, fontSize: 16 }}>
          PM Tracker
        </div>
        <div style={{ padding: '0 12px 12px' }}>
          <select
            data-testid="team-switcher"
            value={currentTeamId ?? ''}
            onChange={(e) => setCurrentTeam(Number(e.target.value))}
            style={{ width: '100%', padding: '4px 8px', borderRadius: 4, border: 'none' }}
          >
            {teams.map((t) => (
              <option key={t.id} value={t.id}>{t.name}</option>
            ))}
          </select>
        </div>
        <nav style={{ flex: 1, padding: '0 8px' }}>
          {allNavItems.map((item) => (
            <div
              key={item.key}
              data-testid={`nav-${item.key.slice(1)}`}
              onClick={() => navigate(item.key)}
              style={{
                padding: '10px 16px',
                borderRadius: 6,
                cursor: 'pointer',
                color: activeKey === item.key ? '#fff' : 'rgba(255,255,255,0.65)',
                background: activeKey === item.key ? '#1677ff' : 'transparent',
                marginBottom: 4,
              }}
            >
              {item.label}
            </div>
          ))}
        </nav>
        <div style={{ padding: 12, borderTop: '1px solid rgba(255,255,255,0.1)' }}>
          <div data-testid="sidebar-user" style={{ display: 'flex', alignItems: 'center', gap: 8, color: '#fff' }}>
            <span style={{ width: 28, height: 28, borderRadius: '50%', background: '#1677ff', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 13 }}>
              {user?.display_name?.charAt(0) ?? '?'}
            </span>
            <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {user?.display_name}
            </span>
            <button onClick={handleLogout} style={{ background: 'none', border: 'none', color: '#fff', cursor: 'pointer' }}>
              退出
            </button>
          </div>
        </div>
      </div>
      <div data-testid="content-area" style={{ marginLeft: 220, flex: 1, padding: 24 }}>
        <Outlet />
      </div>
    </div>
  )
}
