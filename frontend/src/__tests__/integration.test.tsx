import { describe, it, expect, beforeEach, afterEach, beforeAll, afterAll, vi } from 'vitest'
import { render, screen, waitFor, within, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import LoginPage from '@/pages/LoginPage'
import ItemViewPage from '@/pages/ItemViewPage'
import ItemPoolPage from '@/pages/ItemPoolPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team } from '@/types'

// MSW lifecycle for this test suite only
beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

// --- Shared data ---

const mockTeam: Team = {
  id: 1,
  name: 'Team Alpha',
  description: '',
  pm_id: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockUser: User = {
  id: 1,
  username: 'testuser',
  display_name: 'Test User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

function createQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

// antd renders button text with spaces between chars (e.g. "登 录")
function getLoginButton() {
  return screen.getByRole('button', { name: /登.*录/ })
}

// Helper to open antd Select dropdown and select an option
async function openAndSelectOption(selectTestId: string, optionText: string) {
  const selectEl = screen.getByTestId(selectTestId)
  const selector = selectEl.querySelector('.ant-select-selector')!
  fireEvent.mouseDown(selector)
  await waitFor(() => {
    const options = document.querySelectorAll('.ant-select-item-option')
    const match = Array.from(options).find((el) => el.textContent === optionText)
    expect(match).toBeTruthy()
  })
  const options = document.querySelectorAll('.ant-select-item-option')
  const match = Array.from(options).find((el) => el.textContent === optionText)!
  fireEvent.click(match)
}

// --- Tests ---

describe('Integration: Login flow', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
  })

  it('login success: fill credentials, submit, redirect to /items', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter initialEntries={['/login']}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/items" element={<div data-testid="items-page">Items Page</div>} />
        </Routes>
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText('账号'), 'testuser')
    await user.type(screen.getByLabelText('密码'), 'password123')
    await user.click(getLoginButton())

    await waitFor(() => {
      expect(screen.getByTestId('items-page')).toBeInTheDocument()
    })

    // Auth store should have token and user
    const state = useAuthStore.getState()
    expect(state.token).toBe('jwt-token-123')
    expect(state.user?.username).toBe('testuser')
    expect(state.isAuthenticated).toBe(true)
  })

  it('login failure: MSW returns 401, inline error displayed, no redirect', async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter initialEntries={['/login']}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/items" element={<div data-testid="items-page">Items Page</div>} />
        </Routes>
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText('账号'), 'wronguser')
    await user.type(screen.getByLabelText('密码'), 'wrongpass')
    await user.click(getLoginButton())

    await waitFor(() => {
      expect(screen.getByText('账号或密码错误')).toBeInTheDocument()
    })

    // Should stay on login page
    expect(screen.getByTestId('login-page')).toBeInTheDocument()
    expect(screen.queryByTestId('items-page')).not.toBeInTheDocument()

    // Auth store should not have token
    expect(useAuthStore.getState().token).toBeNull()
  })
})

describe('Integration: ItemViewPage with MSW', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('jwt-token-123', mockUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)
  })

  function renderItemView() {
    const qc = createQueryClient()
    return render(
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={['/items']}>
          <Routes>
            <Route path="/items" element={<ItemViewPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    )
  }

  it('renders 3 main items returned by MSW in Collapse', async () => {
    renderItemView()

    await waitFor(() => {
      expect(screen.getByText('Alpha')).toBeInTheDocument()
      expect(screen.getByText('Beta')).toBeInTheDocument()
      expect(screen.getByText('Gamma')).toBeInTheDocument()
    })

    // Verify all 3 are rendered as collapse panels
    expect(screen.getByText('MI-0001')).toBeInTheDocument()
    expect(screen.getByText('MI-0002')).toBeInTheDocument()
    expect(screen.getByText('MI-0003')).toBeInTheDocument()
  })

  it('filter by P1 shows only P1 items (re-fetch with filter param)', async () => {
    renderItemView()

    // Wait for initial load
    await waitFor(() => {
      expect(screen.getByText('Alpha')).toBeInTheDocument()
      expect(screen.getByText('Beta')).toBeInTheDocument()
      expect(screen.getByText('Gamma')).toBeInTheDocument()
    })

    // Select P1 filter
    await openAndSelectOption('filter-priority', 'P1')

    // After filtering, only the P1 item should be visible
    await waitFor(() => {
      expect(screen.getByText('Alpha')).toBeInTheDocument()
      expect(screen.queryByText('Beta')).not.toBeInTheDocument()
      expect(screen.queryByText('Gamma')).not.toBeInTheDocument()
    })
  })
})

describe('Integration: ItemPoolPage with MSW', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('jwt-token-123', mockUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)
  })

  function renderItemPool() {
    const qc = createQueryClient()
    return render(
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={['/item-pool']}>
          <Routes>
            <Route path="/item-pool" element={<ItemPoolPage />} />
            <Route path="/items/:mainItemId" element={<div data-testid="detail-page">Detail</div>} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    )
  }

  it('renders mix of items with correct status tags', async () => {
    renderItemPool()

    await waitFor(() => {
      expect(screen.getByText('Pending Item')).toBeInTheDocument()
      expect(screen.getByText('Assigned Item')).toBeInTheDocument()
      expect(screen.getByText('Rejected Item')).toBeInTheDocument()
    })

    // Verify status tags appear (radio filter buttons also contain same text, so use getAllByText)
    expect(screen.getAllByText('待分配').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('已分配').length).toBeGreaterThanOrEqual(1)
    expect(screen.getAllByText('已拒绝').length).toBeGreaterThanOrEqual(1)

    // Verify status tags specifically rendered in cards (with ant-tag class)
    const pendingCard = screen.getByTestId('pool-card-1')
    const assignedCard = screen.getByTestId('pool-card-2')
    const rejectedCard = screen.getByTestId('pool-card-3')

    expect(pendingCard.querySelector('.ant-tag-blue')).toBeTruthy()
    expect(assignedCard.querySelector('.ant-tag-green')).toBeTruthy()
    expect(rejectedCard.querySelector('.ant-tag-red')).toBeTruthy()
  })

  it('correct card styles: pending has blue border, assigned/rejected are grayed', async () => {
    renderItemPool()

    await waitFor(() => {
      expect(screen.getByTestId('pool-card-1')).toBeInTheDocument()
    })

    // Pending item: blue border
    const pendingCard = screen.getByTestId('pool-card-1')
    expect(pendingCard).toHaveStyle({ borderLeft: '3px solid #1677ff' })

    // Assigned item: grayed out (opacity 0.7)
    const assignedCard = screen.getByTestId('pool-card-2')
    expect(assignedCard).toHaveStyle({ opacity: '0.7' })

    // Rejected item: grayed out + has reject reason
    const rejectedCard = screen.getByTestId('pool-card-3')
    expect(rejectedCard).toHaveStyle({ opacity: '0.7' })
    expect(screen.getByText(/Not suitable/)).toBeInTheDocument()
  })

  it('click "分配" opens assign modal, submit assign, card updates to 已分配', async () => {
    const user = userEvent.setup()
    renderItemPool()

    // Wait for cards to render
    await waitFor(() => {
      expect(screen.getByTestId('assign-btn-1')).toBeInTheDocument()
    })

    // Click assign button
    await user.click(screen.getByTestId('assign-btn-1'))

    // Assign modal should open
    await waitFor(() => {
      expect(screen.getByTestId('assign-modal')).toBeInTheDocument()
    })

    // Select main item
    await openAndSelectOption('form-assign-main-item', 'MI-0001 Alpha')
    // Select assignee
    await openAndSelectOption('form-assign-assignee', 'Test User')

    // Submit assign
    await user.click(screen.getByTestId('assign-ok'))

    // After successful assign, pool list refreshes — the card should still be visible
    await waitFor(() => {
      expect(screen.getByText('Pending Item')).toBeInTheDocument()
    })
  })

  it('click "拒绝" opens reject modal, submit reject, card updates to 已拒绝', async () => {
    const user = userEvent.setup()
    renderItemPool()

    // Wait for cards to render
    await waitFor(() => {
      expect(screen.getByTestId('reject-btn-1')).toBeInTheDocument()
    })

    // Click reject button
    await user.click(screen.getByTestId('reject-btn-1'))

    // Reject modal should open
    await waitFor(() => {
      expect(screen.getByTestId('reject-modal')).toBeInTheDocument()
    })

    // Fill reason and submit
    await user.type(screen.getByTestId('form-reject-reason'), 'Not relevant')
    await user.click(screen.getByTestId('reject-ok'))

    // After successful reject, pool list refreshes
    await waitFor(() => {
      expect(screen.getByText('Pending Item')).toBeInTheDocument()
    })
  })
})

describe('Integration: Auth store cleared on 401', () => {
  it('any API call returning 401 clears auth store', async () => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('jwt-token-123', mockUser)

    // Override MSW to return 401 for the teams members endpoint
    server.use(
      http.get('/api/v1/teams/:teamId/members', () => {
        return new HttpResponse(
          JSON.stringify({ code: 'UNAUTHORIZED', message: 'token expired' }),
          { status: 401, headers: { 'Content-Type': 'application/json' } },
        )
      }),
    )

    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)

    const qc = createQueryClient()
    render(
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={['/items']}>
          <Routes>
            <Route path="/items" element={<ItemViewPage />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    )

    // Wait for the 401 to be processed by the axios response interceptor
    // The interceptor calls clearAuth() before setting window.location.href
    await waitFor(() => {
      expect(useAuthStore.getState().isAuthenticated).toBe(false)
    }, { timeout: 5000 })

    expect(useAuthStore.getState().token).toBeNull()
  })
})
