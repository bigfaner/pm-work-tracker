import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, within, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import AdminPage from './AdminPage'
import { useAuthStore } from '@/store/auth'
import type { User, AdminUser, AdminTeam, PageResult } from '@/types'

// --- Mocks ---

const mockListUsers = vi.fn()
const mockSetCanCreateTeam = vi.fn()
const mockListAdminTeams = vi.fn()

vi.mock('@/api/admin', () => ({
  listUsersApi: (...args: unknown[]) => mockListUsers(...args),
  setCanCreateTeamApi: (...args: unknown[]) => mockSetCanCreateTeam(...args),
  listAdminTeamsApi: (...args: unknown[]) => mockListAdminTeams(...args),
}))

// --- Test Data ---

const superAdminUser: User = {
  id: 1,
  username: 'admin',
  display_name: 'Admin',
  is_super_admin: true,
  can_create_team: true,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const normalUser: User = {
  id: 10,
  username: 'normal',
  display_name: 'Normal',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockAdminUsers: PageResult<AdminUser> = {
  items: [
    { id: 1, username: 'admin', displayName: 'Admin', canCreateTeam: true, isSuperAdmin: true },
    { id: 2, username: 'user1', displayName: 'User One', canCreateTeam: false, isSuperAdmin: false },
    { id: 3, username: 'user2', displayName: 'User Two', canCreateTeam: true, isSuperAdmin: false },
  ],
  total: 3,
  page: 1,
  pageSize: 50,
}

const mockAdminTeams: PageResult<AdminTeam> = {
  items: [
    { id: 1, name: 'Team A', pm: { displayName: 'PM A' }, memberCount: 5, mainItemCount: 10, createdAt: '2024-01-15T00:00:00Z' },
    { id: 2, name: 'Team B', pm: { displayName: 'PM B' }, memberCount: 3, mainItemCount: 7, createdAt: '2024-02-20T00:00:00Z' },
  ],
  total: 2,
  page: 1,
  pageSize: 50,
}

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderPage(user: User = superAdminUser) {
  useAuthStore.getState().setAuth('token', user)
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={['/admin']}>
        <Routes>
          <Route path="/admin" element={<AdminPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Tests ---

describe('AdminPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', superAdminUser)

    mockListUsers.mockResolvedValue(mockAdminUsers)
    mockListAdminTeams.mockResolvedValue(mockAdminTeams)
    mockSetCanCreateTeam.mockReset()
  })

  // --- Basic rendering ---

  it('renders page with data-testid', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('admin-page')).toBeInTheDocument()
    })
  })

  it('renders Tabs with 用户管理 and 团队列表 tabs', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('admin-page')).toBeInTheDocument()
    })
    expect(screen.getByText('用户管理')).toBeInTheDocument()
    expect(screen.getByText('团队列表')).toBeInTheDocument()
  })

  // --- Skeleton loading ---

  it('shows skeleton on initial load for user tab', () => {
    mockListUsers.mockReturnValue(new Promise(() => {}))
    renderPage()
    expect(screen.getByTestId('admin-users-skeleton')).toBeInTheDocument()
  })

  // --- User management tab ---

  it('renders user management table with correct column headers', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })
    expect(screen.getByText('账号')).toBeInTheDocument()
    expect(screen.getByText('姓名')).toBeInTheDocument()
    expect(screen.getByText('创建团队权限')).toBeInTheDocument()
    expect(screen.getAllByText('超级管理员').length).toBeGreaterThanOrEqual(1)
  })

  it('renders user rows with correct data', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })
    // Check usernames are displayed
    expect(screen.getByText('admin')).toBeInTheDocument()
    expect(screen.getByText('user1')).toBeInTheDocument()
    expect(screen.getByText('user2')).toBeInTheDocument()
  })

  it('renders 超级管理员 Tag for super admin users', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })
    // The super admin tag should appear
    const tags = screen.getAllByText('超级管理员')
    expect(tags.length).toBeGreaterThanOrEqual(1)
  })

  it('renders Switch for canCreateTeam column for non-superadmin rows', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })
    // Should have switches - one for user1 (off) and one for user2 (on)
    const switches = screen.getAllByRole('switch')
    expect(switches.length).toBeGreaterThanOrEqual(2)
  })

  it('self-row Switch is disabled with tooltip', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Admin')).toBeInTheDocument()
    })
    // The switch for the admin (self) should be disabled
    // Find the admin row
    const adminCell = screen.getAllByText('admin').find((el) => el.closest('td'))
    const adminRow = adminCell!.closest('tr')!
    const selfSwitch = within(adminRow as HTMLElement).getByRole('switch')
    expect(selfSwitch).toBeDisabled()
  })

  it('calls setCanCreateTeamApi on Switch toggle for non-self row', async () => {
    const user = userEvent.setup()
    mockSetCanCreateTeam.mockResolvedValue(undefined)

    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    // Find user1 row (canCreateTeam is false) and toggle switch
    const user1Cell = screen.getAllByText('user1').find((el) => el.closest('td'))
    const user1Row = user1Cell!.closest('tr')!
    const user1Switch = within(user1Row as HTMLElement).getByRole('switch')
    expect(user1Switch).not.toBeDisabled()

    await user.click(user1Switch)

    await waitFor(() => {
      expect(mockSetCanCreateTeam).toHaveBeenCalledWith(2, { canCreateTeam: true })
    })
  })

  it('shows loading state on Switch during mutation', async () => {
    let resolveMutation!: () => void
    mockSetCanCreateTeam.mockReturnValue(new Promise<void>((resolve) => { resolveMutation = resolve }))

    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    const user1Cell = screen.getAllByText('user1').find((el) => el.closest('td'))
    const user1Row = user1Cell!.closest('tr')!
    const user1Switch = within(user1Row as HTMLElement).getByRole('switch')

    const user = userEvent.setup()
    await user.click(user1Switch)

    // Switch should be loading (disabled while mutating)
    await waitFor(() => {
      expect(user1Switch).toBeDisabled()
    })

    resolveMutation()

    await waitFor(() => {
      expect(user1Switch).not.toBeDisabled()
    })
  })

  it('reverts Switch state on mutation error', async () => {
    const user = userEvent.setup()
    mockSetCanCreateTeam.mockRejectedValue(new Error('Network error'))

    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    // user1 has canCreateTeam=false, so switch should be off
    const user1Cell = screen.getAllByText('user1').find((el) => el.closest('td'))
    const user1Row = user1Cell!.closest('tr')!
    const user1Switch = within(user1Row as HTMLElement).getByRole('switch')

    expect(user1Switch).toHaveAttribute('aria-checked', 'false')

    await user.click(user1Switch)

    // After error, should revert to original state
    await waitFor(() => {
      expect(user1Switch).toHaveAttribute('aria-checked', 'false')
    })
  })

  it('invalidates users query after successful toggle', async () => {
    const evtUser = userEvent.setup()
    mockSetCanCreateTeam.mockResolvedValue(undefined)

    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    mockListUsers.mockClear()

    const user1Cell = screen.getAllByText('user1').find((el) => el.closest('td'))
    const user1Row = user1Cell!.closest('tr')!
    const user1Switch = within(user1Row as HTMLElement).getByRole('switch')
    await evtUser.click(user1Switch)

    await waitFor(() => {
      expect(mockListUsers).toHaveBeenCalled()
    })
  })

  // --- Team list tab ---

  it('renders team list table when 团队列表 tab is clicked', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    await user.click(screen.getByText('团队列表'))

    await waitFor(() => {
      expect(screen.getByText('Team A')).toBeInTheDocument()
    })
  })

  it('renders team list table with correct column headers', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    await user.click(screen.getByText('团队列表'))

    await waitFor(() => {
      expect(screen.getByText('Team A')).toBeInTheDocument()
    })
    expect(screen.getByText('团队名称')).toBeInTheDocument()
    expect(screen.getByText('PM姓名')).toBeInTheDocument()
    expect(screen.getByText('成员数')).toBeInTheDocument()
    expect(screen.getByText('主事项数')).toBeInTheDocument()
    expect(screen.getByText('创建时间')).toBeInTheDocument()
  })

  it('renders team rows with correct data', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    await user.click(screen.getByText('团队列表'))

    await waitFor(() => {
      expect(screen.getByText('Team A')).toBeInTheDocument()
    })
    expect(screen.getByText('PM A')).toBeInTheDocument()
    expect(screen.getByText('Team B')).toBeInTheDocument()
    expect(screen.getByText('PM B')).toBeInTheDocument()
  })

  it('renders member count and main item count', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    await user.click(screen.getByText('团队列表'))

    await waitFor(() => {
      expect(screen.getByText('Team A')).toBeInTheDocument()
    })
    expect(screen.getByText('5')).toBeInTheDocument()
    expect(screen.getByText('10')).toBeInTheDocument()
  })

  it('renders creation time in correct format', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    await user.click(screen.getByText('团队列表'))

    await waitFor(() => {
      expect(screen.getByText('Team A')).toBeInTheDocument()
    })
    expect(screen.getByText('2024-01-15')).toBeInTheDocument()
    expect(screen.getByText('2024-02-20')).toBeInTheDocument()
  })

  // --- Pagination ---

  it('renders user table with pagination', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })
    // Antd pagination has a .ant-pagination element
    const userTab = screen.getByTestId('admin-users-table')
    const pagination = within(userTab).getByRole('list')
    expect(pagination).toBeInTheDocument()
  })

  it('renders team table with pagination', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('User One')).toBeInTheDocument()
    })

    await user.click(screen.getByText('团队列表'))

    await waitFor(() => {
      expect(screen.getByText('Team A')).toBeInTheDocument()
    })
    const teamTab = screen.getByTestId('admin-teams-table')
    const pagination = within(teamTab).getByRole('list')
    expect(pagination).toBeInTheDocument()
  })
})
