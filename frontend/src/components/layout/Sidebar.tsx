import { useNavigate, useLocation } from 'react-router-dom'
import {
  LayoutGrid,
  Calendar,
  AlignLeft,
  Inbox,
  FileDown,
  UserCog,
  Users,
  LogOut,
  Shield,
} from 'lucide-react'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import { PermissionGuard } from '@/components/PermissionGuard'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const businessItems = [
  { key: '/items', label: '事项清单', icon: LayoutGrid },
  { key: '/item-pool', label: '待办事项', icon: Inbox },
  { key: '/weekly', label: '每周进展', icon: Calendar },
  { key: '/gantt', label: '整体进度', icon: AlignLeft, permission: 'view:gantt' },
  { key: '/report', label: '周报导出', icon: FileDown },
]

const adminItems = [
  { key: '/teams', label: '团队管理', icon: Users },
  { key: '/users', label: '用户管理', icon: UserCog, permission: 'user:list' },
  { key: '/roles', label: '角色管理', icon: Shield, permission: 'role:read' },
]

export default function Sidebar() {
  const navigate = useNavigate()
  const location = useLocation()

  const { user, clearAuth } = useAuthStore()
  const { teams, currentTeamId, setCurrentTeam } = useTeamStore()

  const activeKey = '/' + location.pathname.split('/').filter(Boolean)[0]

  const handleLogout = () => {
    clearAuth()
    navigate('/login')
  }

  return (
    <nav
      data-testid="sidebar"
      className="w-[240px] bg-white border-r border-border flex flex-col fixed top-0 left-0 bottom-0 z-40"
    >
      <div className="px-4 py-5 border-b border-border text-center">
        <h2 className="text-lg font-semibold text-primary">PM Tracker</h2>
        <p className="text-xs text-tertiary mt-0.5">工作事项追踪</p>
      </div>

      {teams.length > 0 && (
        <div className="px-3 pt-3">
          <Select
            value={currentTeamId ?? undefined}
            onValueChange={(val) => setCurrentTeam(val)}
          >
            <SelectTrigger data-testid="team-switcher" className="w-full h-8 text-[13px] [&>span:first-child]:flex-1 [&>span:first-child]:text-center">
              <SelectValue placeholder="选择团队" />
            </SelectTrigger>
            <SelectContent>
              {teams.map((t) => (
                <SelectItem key={t.bizKey} value={String(t.bizKey)} className="justify-center">
                  {t.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      <div className="flex-1 px-2 py-3 overflow-y-auto">
        {businessItems.map((item) => {
          const Icon = item.icon
          const isActive = activeKey === item.key
          const navLink = (
            <a
              key={item.key}
              href={item.key}
              onClick={(e) => { e.preventDefault(); navigate(item.key) }}
              className={`flex items-center justify-center gap-2.5 px-3 py-2.5 rounded-lg text-[13px] font-medium transition-all ${
                isActive
                  ? 'bg-primary-50 text-primary-700'
                  : 'text-secondary hover:bg-bg-alt hover:text-primary'
              }`}
            >
              <Icon className="w-[18px] h-[18px] shrink-0" />
              {item.label}
            </a>
          )
          if (item.permission) {
            return (
              <PermissionGuard key={item.key} code={item.permission}>
                {navLink}
              </PermissionGuard>
            )
          }
          return navLink
        })}

        {adminItems.map((item, idx) => {
          const Icon = item.icon
          const isActive = activeKey === item.key
          const navLink = (
            <a
              key={item.key}
              href={item.key}
              onClick={(e) => { e.preventDefault(); navigate(item.key) }}
              className={`flex items-center justify-center gap-2.5 px-3 py-2.5 rounded-lg text-[13px] font-medium transition-all ${
                idx === 0 ? 'mt-2 pt-2 border-t border-border' : ''
              } ${
                isActive
                  ? 'bg-primary-50 text-primary-700'
                  : 'text-secondary hover:bg-bg-alt hover:text-primary'
              }`}
            >
              <Icon className="w-[18px] h-[18px] shrink-0" />
              {item.label}
            </a>
          )
          if (item.permission) {
            return (
              <PermissionGuard key={item.key} code={item.permission}>
                {navLink}
              </PermissionGuard>
            )
          }
          return navLink
        })}
      </div>

      <div className="px-4 py-3 border-t border-border flex items-center gap-2.5">
        <div className="w-8 h-8 rounded-full bg-primary-100 text-primary-700 flex items-center justify-center text-[13px] font-semibold">
          {user?.displayName?.charAt(0) ?? '?'}
        </div>
        <div className="flex-1 min-w-0">
          <div className="text-[13px] font-medium text-primary truncate">{user?.displayName}</div>
        </div>
        <button
          data-testid="sidebar-logout"
          onClick={handleLogout}
          className="w-7 h-7 flex items-center justify-center text-secondary hover:text-primary hover:bg-bg-alt rounded-md transition-colors"
        >
          <LogOut className="w-4 h-4" />
        </button>
      </div>
    </nav>
  )
}
