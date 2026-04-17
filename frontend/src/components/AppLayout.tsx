import { useState, useEffect } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Select, Avatar, Dropdown } from 'antd'
import {
  UnorderedListOutlined,
  CalendarOutlined,
  FundOutlined,
  TableOutlined,
  InboxOutlined,
  ExportOutlined,
  SettingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import type { MenuProps } from 'antd'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import { listTeamsApi } from '@/api/teams'

const { Sider } = Layout

const navItems = [
  { key: '/items', label: '事项视图', icon: <UnorderedListOutlined /> },
  { key: '/weekly', label: '周视图', icon: <CalendarOutlined /> },
  { key: '/gantt', label: '甘特图', icon: <FundOutlined /> },
  { key: '/table', label: '表格视图', icon: <TableOutlined /> },
  { key: '/item-pool', label: '事项池', icon: <InboxOutlined /> },
  { key: '/report', label: '周报导出', icon: <ExportOutlined /> },
]

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()

  const { user, clearAuth, isSuperAdmin } = useAuthStore()
  const { teams, currentTeamId, setCurrentTeam, setTeams } = useTeamStore()

  // Fetch teams on mount
  useEffect(() => {
    listTeamsApi()
      .then((data) => {
        setTeams(data)
        if (data.length > 0 && !currentTeamId) {
          setCurrentTeam(data[0].id)
        }
      })
      .catch(() => {
        // Silently handle - error already shown by API client interceptor
      })
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    navigate(key)
  }

  const handleTeamChange = (teamId: number) => {
    setCurrentTeam(teamId)
  }

  const handleLogout = () => {
    clearAuth()
    navigate('/login')
  }

  const dropdownItems: MenuProps['items'] = [
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  // Determine active menu key from current path
  const activeKey = '/' + location.pathname.split('/').filter(Boolean)[0]

  const allMenuItems = isSuperAdmin
    ? [
        ...navItems,
        { type: 'divider' as const },
        { key: '/admin', label: '管理后台', icon: <SettingOutlined /> },
      ]
    : navItems

  return (
    <Layout data-testid="app-layout" style={{ minHeight: '100vh' }}>
      <Sider
        data-testid="sidebar"
        width={220}
        collapsedWidth={64}
        collapsed={collapsed}
        trigger={null}
        style={{
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          display: 'flex',
          flexDirection: 'column',
          overflow: 'auto',
        }}
      >
        {/* Top: Logo + Team switcher */}
        <div style={{ padding: '16px 12px 8px' }}>
          <div
            style={{
              fontSize: collapsed ? 14 : 16,
              fontWeight: 700,
              color: '#fff',
              textAlign: 'center',
              marginBottom: 12,
              whiteSpace: 'nowrap',
              overflow: 'hidden',
            }}
          >
            {collapsed ? 'PM' : 'PM Tracker'}
          </div>
          {!collapsed && (
            <Select
              data-testid="team-switcher"
              value={currentTeamId ?? undefined}
              onChange={handleTeamChange}
              style={{ width: 180 }}
              placeholder="选择团队"
              options={teams.map((t) => ({ value: t.id, label: t.name }))}
            />
          )}
        </div>

        {/* Navigation */}
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[activeKey]}
          items={allMenuItems}
          onClick={handleMenuClick}
          style={{ flex: 1, borderRight: 0 }}
        />

        {/* Bottom: User section + Collapse toggle */}
        <div style={{ padding: '12px', borderTop: '1px solid rgba(255,255,255,0.1)' }}>
          <Dropdown menu={{ items: dropdownItems }} placement="topRight" trigger={['click']}>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                cursor: 'pointer',
                color: '#fff',
              }}
            >
              <Avatar size="small" style={{ backgroundColor: '#1677ff', flexShrink: 0 }}>
                {user?.display_name?.charAt(0) ?? '?'}
              </Avatar>
              {!collapsed && (
                <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {user?.display_name}
                </span>
              )}
            </div>
          </Dropdown>
          <div
            data-testid="sidebar-toggle"
            onClick={() => setCollapsed(!collapsed)}
            style={{
              color: '#fff',
              cursor: 'pointer',
              textAlign: 'center',
              marginTop: 12,
              fontSize: 16,
            }}
          >
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </div>
        </div>
      </Sider>

      {/* Content area */}
      <Layout style={{ marginLeft: collapsed ? 64 : 220, transition: 'margin-left 0.2s' }}>
        <div
          data-testid="content-area"
          style={{
            padding: 24,
            maxWidth: 1440,
            width: '100%',
            margin: '0 auto',
          }}
        >
          <Outlet />
        </div>
      </Layout>
    </Layout>
  )
}
