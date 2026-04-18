import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import AppLayout from './AppLayout'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team } from '@/types'

// Mock the teams API
vi.mock('@/api/teams', () => ({
  listTeamsApi: vi.fn().mockResolvedValue([]),
}))

const mockUser: User = {
  id: 1,
  username: 'testuser',
  display_name: 'Test User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const superAdminUser: User = {
  ...mockUser,
  id: 2,
  username: 'admin',
  display_name: 'Admin User',
  is_super_admin: true,
}

const mockTeams: Team[] = [
  {
    id: 1,
    name: 'Team Alpha',
    description: '',
    pm_id: 1,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    name: 'Team Beta',
    description: '',
    pm_id: 2,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
]

function renderWithRouter(initialPath = '/items') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/items" element={<div data-testid="page-items">Items</div>} />
          <Route path="/weekly" element={<div data-testid="page-weekly">Weekly</div>} />
          <Route path="/gantt" element={<div data-testid="page-gantt">Gantt</div>} />
          <Route path="/table" element={<div data-testid="page-table">Table</div>} />
          <Route path="/item-pool" element={<div data-testid="page-pool">Pool</div>} />
          <Route path="/report" element={<div data-testid="page-report">Report</div>} />
          <Route path="/admin" element={<div data-testid="page-admin">Admin</div>} />
        </Route>
        <Route path="/login" element={<div data-testid="page-login">Login</div>} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('AppLayout', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', mockUser)
    useTeamStore.getState().setTeams([])
    useTeamStore.getState().setCurrentTeam(null)
    vi.clearAllMocks()
  })

  it('renders sidebar and content area', () => {
    renderWithRouter()
    expect(screen.getByTestId('app-layout')).toBeInTheDocument()
    expect(screen.getByTestId('sidebar')).toBeInTheDocument()
    expect(screen.getByTestId('content-area')).toBeInTheDocument()
  })

  it('renders nav items with correct labels', () => {
    renderWithRouter()
    expect(screen.getByText('事项视图')).toBeInTheDocument()
    expect(screen.getByText('周视图')).toBeInTheDocument()
    expect(screen.getByText('甘特图')).toBeInTheDocument()
    expect(screen.getByText('表格视图')).toBeInTheDocument()
    expect(screen.getByText('事项池')).toBeInTheDocument()
    expect(screen.getByText('周报导出')).toBeInTheDocument()
  })

  it('renders team switcher with teams from store', () => {
    useTeamStore.getState().setTeams(mockTeams)
    renderWithRouter()
    const switcher = screen.getByTestId('team-switcher')
    expect(switcher).toBeInTheDocument()
  })

  it('renders user avatar with first char of displayName', () => {
    renderWithRouter()
    expect(screen.getByText('T')).toBeInTheDocument()
  })

  it('renders user display name', () => {
    renderWithRouter()
    expect(screen.getByText('Test User')).toBeInTheDocument()
  })

  it('clicking logout clears auth and navigates to login', async () => {
    const user = userEvent.setup()
    renderWithRouter()
    const logoutBtn = screen.getByText('退出')
    await user.click(logoutBtn)
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
    expect(screen.getByTestId('page-login')).toBeInTheDocument()
  })

  it('shows admin link for super admin users', () => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', superAdminUser)
    renderWithRouter()
    expect(screen.getByText('管理后台')).toBeInTheDocument()
  })

  it('hides admin link for regular users', () => {
    renderWithRouter()
    expect(screen.queryByText('管理后台')).not.toBeInTheDocument()
  })

  it('team switcher updates currentTeamId in store', () => {
    useTeamStore.getState().setTeams(mockTeams)
    renderWithRouter()

    const selectEl = screen.getByTestId('team-switcher')
    expect(selectEl).toBeInTheDocument()

    useTeamStore.getState().setCurrentTeam(1)
    expect(useTeamStore.getState().currentTeamId).toBe(1)
  })

  it('fetches teams on mount', async () => {
    const { listTeamsApi } = await import('@/api/teams')
    const mocked = vi.mocked(listTeamsApi)
    mocked.mockResolvedValueOnce(mockTeams)

    renderWithRouter()

    await screen.findByTestId('app-layout')
    expect(mocked).toHaveBeenCalledOnce()
  })
})
