import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import AppLayout from './AppLayout'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team } from '@/types'

// Mock the teams API
vi.mock('@/api/teams', () => ({
  listTeamsApi: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, pageSize: 20 }),
}))

const mockUser: User = {
  bizKey: '1',
  username: 'testuser',
  displayName: 'Test User',
  isSuperAdmin: false,
  createTime: '',
}

const mockTeams: Team[] = [
  {
    bizKey: '1',
    name: 'Team Alpha',
    description: '',
    code: '',
    pmKey: '1',
  createdAt: '2024-01-01',
  updatedAt: '2024-01-01',
  },
]

function renderWithRouter(initialPath = '/items') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route element={<AppLayout />}>
          <Route path="/items" element={<div data-testid="page-items">Items</div>} />
          <Route path="/weekly" element={<div data-testid="page-weekly">Weekly</div>} />
          <Route path="/teams" element={<div data-testid="page-teams">Teams</div>} />
          <Route path="/users" element={<div data-testid="page-users">Users</div>} />
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

  it('renders sidebar and scrollable main content area', () => {
    renderWithRouter()
    expect(screen.getByTestId('sidebar')).toBeInTheDocument()
    expect(screen.getByTestId('content-area')).toBeInTheDocument()
  })

  it('renders child route content in main content area', () => {
    renderWithRouter('/items')
    expect(screen.getByTestId('page-items')).toBeInTheDocument()
  })

  it('fetches teams on mount', async () => {
    const { listTeamsApi } = await import('@/api/teams')
    const mocked = vi.mocked(listTeamsApi)
    mocked.mockResolvedValueOnce({ items: mockTeams, total: mockTeams.length, page: 1, pageSize: 20 })

    renderWithRouter()
    await screen.findByTestId('app-layout')
    expect(mocked).toHaveBeenCalledOnce()
  })
})
