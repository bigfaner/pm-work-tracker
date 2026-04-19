import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import Sidebar from './Sidebar'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team } from '@/types'

const mockUser: User = {
  id: 1,
  username: 'testuser',
  displayName: '张明',
  isSuperAdmin: false,
  canCreateTeam: false,
}

const superAdminUser: User = {
  ...mockUser,
  id: 2,
  username: 'admin',
  displayName: 'Admin',
  isSuperAdmin: true,
}

const mockTeams: Team[] = [
  {
    id: 1,
    name: '产品研发团队',
    description: '',
    pmId: 1,
  createdAt: '2024-01-01',
  updatedAt: '2024-01-01',
  },
  {
    id: 2,
    name: '设计团队',
    description: '',
    pmId: 2,
    createdAt: '2024-01-01',
    updatedAt: '2024-01-01',
  },
]

function renderWithRouter(initialPath = '/items') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route path="*" element={<Sidebar />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('Sidebar', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', mockUser)
    useTeamStore.getState().setTeams([])
    useTeamStore.getState().setCurrentTeam(null)
    vi.clearAllMocks()
  })

  it('renders PM Tracker brand', () => {
    renderWithRouter()
    expect(screen.getByText('PM Tracker')).toBeInTheDocument()
  })

  it('renders navigation items with correct labels', () => {
    renderWithRouter()
    expect(screen.getByText('事项清单')).toBeInTheDocument()
    expect(screen.getByText('每周进展')).toBeInTheDocument()
    expect(screen.getByText('整体进度')).toBeInTheDocument()
    expect(screen.getByText('待办事项')).toBeInTheDocument()
    expect(screen.getByText('周报导出')).toBeInTheDocument()
    expect(screen.getByText('团队管理')).toBeInTheDocument()
  })

  it('renders user info with first character of display name', () => {
    renderWithRouter()
    expect(screen.getByText('张')).toBeInTheDocument()
  })

  it('renders logout button', () => {
    renderWithRouter()
    // Logout is a button with an SVG icon (LogOut from lucide)
    const logoutBtn = screen.getByTestId('sidebar-logout')
    expect(logoutBtn).toBeInTheDocument()
  })

  it('highlights current active route', () => {
    renderWithRouter('/items')
    const itemsLink = screen.getByText('事项清单').closest('a')
    // Active links get bg-primary-50 and text-primary-700
    expect(itemsLink?.className).toContain('bg-primary-50')
    expect(itemsLink?.className).toContain('text-primary-700')
  })

  it('shows user management nav item only for SuperAdmin', () => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', superAdminUser)
    renderWithRouter()
    expect(screen.getByText('用户管理')).toBeInTheDocument()
  })

  it('hides user management nav item for regular users', () => {
    renderWithRouter()
    expect(screen.queryByText('用户管理')).not.toBeInTheDocument()
  })

  it('renders team selector with teams from store', () => {
    useTeamStore.getState().setTeams(mockTeams)
    useTeamStore.getState().setCurrentTeam(1)
    renderWithRouter()
    // The team selector should show the current team name
    expect(screen.getByText('产品研发团队')).toBeInTheDocument()
  })

  it('clicking logout clears auth', async () => {
    const user = userEvent.setup()
    renderWithRouter()
    await user.click(screen.getByTestId('sidebar-logout'))
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
  })

  it('renders 6 standard nav items for regular user', () => {
    renderWithRouter()
    const navLinks = screen.getAllByRole('link')
    // 6 standard: items, weekly, gantt, item-pool, report, teams
    expect(navLinks.length).toBe(6)
  })

  it('renders 7 nav items (including user mgmt) for SuperAdmin', () => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', superAdminUser)
    renderWithRouter()
    const navLinks = screen.getAllByRole('link')
    expect(navLinks.length).toBe(7)
  })
})
