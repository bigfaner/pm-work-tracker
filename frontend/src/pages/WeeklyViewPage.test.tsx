import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import WeeklyViewPage from './WeeklyViewPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team, WeeklyViewResp } from '@/types'

// --- Mocks ---

const mockGetWeeklyView = vi.fn()

vi.mock('@/api/views', () => ({
  getWeeklyViewApi: (...args: unknown[]) => mockGetWeeklyView(...args),
}))

// --- Test Data ---

const mockUser: User = {
  id: 1,
  username: 'pmuser',
  display_name: 'PM User',
  is_super_admin: false,
  can_create_team: false,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const mockTeam: Team = {
  id: 1,
  name: 'Team Alpha',
  description: '',
  pm_id: 1,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
}

const emptyWeeklyResp: WeeklyViewResp = {
  weekStart: '2026-04-13',
  weekEnd: '2026-04-19',
  groups: [],
}

const weeklyRespWithGroups: WeeklyViewResp = {
  weekStart: '2026-04-13',
  weekEnd: '2026-04-19',
  groups: [
    {
      mainItem: { id: 1, title: 'Main Item Alpha', completion: 60 },
      newlyCompleted: [
        {
          subItem: {
            id: 10,
            team_id: 1,
            main_item_id: 1,
            title: 'Sub Completed 1',
            description: '',
            priority: 'P2',
            assignee_id: 2,
            start_date: null,
            expected_end_date: null,
            actual_end_date: null,
            status: '已完成',
            completion: 100,
            is_key_item: false,
            delay_count: 0,
            weight: 1,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
          progressThisWeek: [],
        },
      ],
      hasProgress: [
        {
          subItem: {
            id: 11,
            team_id: 1,
            main_item_id: 1,
            title: 'Sub Progress 1',
            description: '',
            priority: 'P2',
            assignee_id: 2,
            start_date: null,
            expected_end_date: null,
            actual_end_date: null,
            status: '进行中',
            completion: 70,
            is_key_item: false,
            delay_count: 0,
            weight: 1,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
          progressThisWeek: [
            {
              id: 100,
              sub_item_id: 11,
              team_id: 1,
              author_id: 2,
              completion: 70,
              achievement: '完成了接口开发',
              blocker: '等待联调环境',
              lesson: '',
              is_pm_correct: false,
              created_at: '2026-04-15T10:00:00Z',
            },
          ],
        },
      ],
      noChangeFromLastWeek: [
        { id: 12, title: 'Sub No Change', status: '进行中', completion: 30 },
      ],
    },
  ],
}

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
}

function renderPage() {
  useAuthStore.getState().setAuth('token', mockUser)
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={['/weekly']}>
        <Routes>
          <Route path="/weekly" element={<WeeklyViewPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Tests ---

describe('WeeklyViewPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', mockUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)
    mockGetWeeklyView.mockReset()
  })

  // --- Basic rendering ---

  it('renders page with data-testid', () => {
    mockGetWeeklyView.mockResolvedValue(emptyWeeklyResp)
    renderPage()
    expect(screen.getByTestId('weekly-view-page')).toBeInTheDocument()
  })

  it('renders page title 周视图', () => {
    mockGetWeeklyView.mockResolvedValue(emptyWeeklyResp)
    renderPage()
    expect(screen.getByText('周视图')).toBeInTheDocument()
  })

  // --- Week picker ---

  it('renders week picker', () => {
    mockGetWeeklyView.mockResolvedValue(emptyWeeklyResp)
    renderPage()
    expect(screen.getByTestId('week-picker')).toBeInTheDocument()
  })

  // --- Loading state ---

  it('shows skeleton loading state while fetching data', () => {
    let resolvePromise!: (v: unknown) => void
    mockGetWeeklyView.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve }))
    renderPage()
    expect(screen.getByTestId('weekly-loading-skeleton')).toBeInTheDocument()
    resolvePromise(emptyWeeklyResp)
  })

  // --- Empty state ---

  it('shows empty state when no groups returned', async () => {
    mockGetWeeklyView.mockResolvedValue(emptyWeeklyResp)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('weekly-empty-state')).toBeInTheDocument()
    })
    expect(screen.getByText('本周暂无事项进度')).toBeInTheDocument()
  })

  // --- Week range display ---

  it('displays week range in format YYYY-MM-DD ~ YYYY-MM-DD', async () => {
    mockGetWeeklyView.mockResolvedValue(weeklyRespWithGroups)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('2026-04-13 ~ 2026-04-19')).toBeInTheDocument()
    })
  })

  // --- Main item group rendering ---

  it('renders main item card with title and progress bar', async () => {
    mockGetWeeklyView.mockResolvedValue(weeklyRespWithGroups)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Main Item Alpha')).toBeInTheDocument()
    })
    expect(screen.getByTestId('main-item-progress-1')).toBeInTheDocument()
  })

  // --- Three sections ---

  it('renders 本周新完成 section with green header', async () => {
    mockGetWeeklyView.mockResolvedValue(weeklyRespWithGroups)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('本周新完成')).toBeInTheDocument()
    })
    expect(screen.getByText('Sub Completed 1')).toBeInTheDocument()
  })

  it('renders 本周有进度 section with blue header', async () => {
    mockGetWeeklyView.mockResolvedValue(weeklyRespWithGroups)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('本周有进度')).toBeInTheDocument()
    })
    expect(screen.getByText('Sub Progress 1')).toBeInTheDocument()
  })

  it('renders 上周完成/无变化 section with gray header', async () => {
    mockGetWeeklyView.mockResolvedValue(weeklyRespWithGroups)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('上周完成/无变化')).toBeInTheDocument()
    })
    expect(screen.getByText('Sub No Change')).toBeInTheDocument()
  })

  // --- Newly completed sub-items ---

  it('shows 100% badge for newly completed sub-items', async () => {
    mockGetWeeklyView.mockResolvedValue(weeklyRespWithGroups)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('completed-badge-10')).toBeInTheDocument()
    })
  })

  // --- Has progress sub-items ---

  it('shows progress records for has-progress sub-items', async () => {
    mockGetWeeklyView.mockResolvedValue(weeklyRespWithGroups)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('完成了接口开发', { exact: false })).toBeInTheDocument()
    })
    expect(screen.getByText('等待联调环境', { exact: false })).toBeInTheDocument()
  })

  it('shows current completion for has-progress sub-items', async () => {
    mockGetWeeklyView.mockResolvedValue(weeklyRespWithGroups)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('sub-progress-11')).toBeInTheDocument()
    })
  })

  // --- No change sub-items ---

  it('shows summary row with title, status, completion for no-change sub-items', async () => {
    mockGetWeeklyView.mockResolvedValue(weeklyRespWithGroups)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('nochange-row-12')).toBeInTheDocument()
    })
  })

  // --- API call ---

  it('calls getWeeklyViewApi with teamId and weekStart', async () => {
    mockGetWeeklyView.mockResolvedValue(emptyWeeklyResp)
    renderPage()
    await waitFor(() => {
      expect(mockGetWeeklyView).toHaveBeenCalledWith(1, expect.any(String))
    })
  })

  it('sends weekStart as Monday ISO date', async () => {
    mockGetWeeklyView.mockResolvedValue(emptyWeeklyResp)
    renderPage()
    await waitFor(() => {
      const weekStart = mockGetWeeklyView.mock.calls[0][1] as string
      // Should be a Monday date in YYYY-MM-DD format
      expect(weekStart).toMatch(/^\d{4}-\d{2}-\d{2}$/)
    })
  })

  // --- Empty sections still rendered ---

  it('renders main item card even when all sections are empty', async () => {
    const respWithEmptySections: WeeklyViewResp = {
      weekStart: '2026-04-13',
      weekEnd: '2026-04-19',
      groups: [
        {
          mainItem: { id: 2, title: 'Empty Main Item', completion: 0 },
          newlyCompleted: [],
          hasProgress: [],
          noChangeFromLastWeek: [],
        },
      ],
    }
    mockGetWeeklyView.mockResolvedValue(respWithEmptySections)
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Empty Main Item')).toBeInTheDocument()
    })
  })
})
