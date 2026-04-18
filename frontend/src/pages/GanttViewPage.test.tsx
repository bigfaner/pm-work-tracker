import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import GanttViewPage from './GanttViewPage'
import { useAuthStore } from '@/store/auth'
import { useTeamStore } from '@/store/team'
import type { User, Team, GanttViewResp } from '@/types'

// --- Mocks ---

const mockGetGanttView = vi.fn()

vi.mock('@/api/views', () => ({
  getGanttViewApi: (...args: unknown[]) => mockGetGanttView(...args),
}))

// Mock frappe-gantt constructor calls to capture tasks and options
const mockGanttRefresh = vi.fn()
const mockGanttInstance = {
  change_view_mode: vi.fn(),
  refresh: mockGanttRefresh,
}

const capturedCalls: { tasks: unknown[]; options: unknown }[] = []

vi.mock('frappe-gantt', () => {
  return {
    default: class MockGantt {
      constructor(wrapper: unknown, tasks: unknown[], options?: unknown) {
        capturedCalls.push({ tasks, options: options || {} })
        Object.assign(this, mockGanttInstance)
      }
    },
  }
})

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

const ganttRespWithItems: GanttViewResp = {
  items: [
    {
      id: 1,
      title: 'Main Item A',
      priority: 'P1',
      startDate: '2026-04-01',
      expectedEndDate: '2026-04-30',
      completion: 60,
      status: '进行中',
      isOverdue: false,
      subItems: [
        {
          id: 10,
          title: 'Sub Item A1',
          startDate: '2026-04-05',
          expectedEndDate: '2026-04-20',
          completion: 80,
          status: '进行中',
        },
      ],
    },
    {
      id: 2,
      title: 'Main Item B',
      priority: 'P2',
      startDate: '2026-04-10',
      expectedEndDate: '2026-04-25',
      completion: 30,
      status: '待开始',
      isOverdue: false,
      subItems: [],
    },
  ],
}

const ganttRespOverdue: GanttViewResp = {
  items: [
    {
      id: 3,
      title: 'Overdue Item',
      priority: 'P3',
      startDate: '2026-03-01',
      expectedEndDate: '2026-03-31',
      completion: 40,
      status: '已延期',
      isOverdue: true,
      subItems: [],
    },
  ],
}

const ganttRespNoDates: GanttViewResp = {
  items: [
    {
      id: 4,
      title: 'No Dates Item',
      priority: 'P2',
      startDate: null,
      expectedEndDate: null,
      completion: 0,
      status: '待开始',
      isOverdue: false,
      subItems: [],
    },
  ],
}

const emptyGanttResp: GanttViewResp = {
  items: [],
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
      <MemoryRouter initialEntries={['/gantt']}>
        <Routes>
          <Route path="/gantt" element={<GanttViewPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Tests ---

describe('GanttViewPage', () => {
  beforeEach(() => {
    useAuthStore.getState().clearAuth()
    useAuthStore.getState().setAuth('token', mockUser)
    useTeamStore.getState().setTeams([mockTeam])
    useTeamStore.getState().setCurrentTeam(1)
    mockGetGanttView.mockReset()
    capturedCalls.length = 0
    mockGanttRefresh.mockReset()
  })

  // --- Basic rendering ---

  it('renders page with data-testid', () => {
    mockGetGanttView.mockResolvedValue(emptyGanttResp)
    renderPage()
    expect(screen.getByTestId('gantt-view-page')).toBeInTheDocument()
  })

  it('renders page title 甘特图', () => {
    mockGetGanttView.mockResolvedValue(emptyGanttResp)
    renderPage()
    expect(screen.getByText('甘特图')).toBeInTheDocument()
  })

  // --- Status filter ---

  it('renders status filter Select', () => {
    mockGetGanttView.mockResolvedValue(emptyGanttResp)
    renderPage()
    expect(screen.getByTestId('gantt-status-filter')).toBeInTheDocument()
  })

  it('calls API with status filter when changed', async () => {
    mockGetGanttView.mockResolvedValue(emptyGanttResp)
    renderPage()
    // Wait for initial load
    await waitFor(() => {
      expect(mockGetGanttView).toHaveBeenCalledWith(1, undefined)
    })
  })

  // --- Loading state ---

  it('shows Spin loading state while fetching data', () => {
    let resolvePromise!: (v: unknown) => void
    mockGetGanttView.mockReturnValue(new Promise((resolve) => { resolvePromise = resolve }))
    renderPage()
    expect(screen.getByTestId('gantt-loading')).toBeInTheDocument()
    resolvePromise(emptyGanttResp)
  })

  // --- Empty state ---

  it('shows empty state when no items returned', async () => {
    mockGetGanttView.mockResolvedValue(emptyGanttResp)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('gantt-empty-state')).toBeInTheDocument()
    })
    expect(screen.getByText('暂无事项数据')).toBeInTheDocument()
  })

  // --- Gantt rendering ---

  it('renders gantt chart container when data is loaded', async () => {
    mockGetGanttView.mockResolvedValue(ganttRespWithItems)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('gantt-chart-container')).toBeInTheDocument()
    })
  })

  it('passes tasks to frappe-gantt with correct format', async () => {
    mockGetGanttView.mockResolvedValue(ganttRespWithItems)
    renderPage()
    await waitFor(() => {
      expect(capturedCalls.length).toBeGreaterThan(0)
    })
    const firstCall = capturedCalls[0]
    const tasks = firstCall.tasks as Array<{ id: string; name: string; start: string; end: string; progress: number; custom_class?: string }>
    // Should have main items only (collapsed by default)
    expect(tasks.length).toBe(2)
    expect(tasks[0].id).toBe('1')
    expect(tasks[0].name).toBe('Main Item A')
    expect(tasks[0].start).toBe('2026-04-01')
    expect(tasks[0].end).toBe('2026-04-30')
    expect(tasks[0].progress).toBe(60)
  })

  // --- Priority coloring ---

  it('applies P1 (orange) custom class for P1 items', async () => {
    mockGetGanttView.mockResolvedValue(ganttRespWithItems)
    renderPage()
    await waitFor(() => {
      expect(capturedCalls.length).toBeGreaterThan(0)
    })
    const tasks = capturedCalls[0].tasks as Array<{ id: string; custom_class?: string }>
    const p1Task = tasks.find((t) => t.id === '1')
    expect(p1Task?.custom_class).toContain('bar-p1')
  })

  it('applies P2 (blue) custom class for P2 items', async () => {
    mockGetGanttView.mockResolvedValue(ganttRespWithItems)
    renderPage()
    await waitFor(() => {
      expect(capturedCalls.length).toBeGreaterThan(0)
    })
    const tasks = capturedCalls[0].tasks as Array<{ id: string; custom_class?: string }>
    const p2Task = tasks.find((t) => t.id === '2')
    expect(p2Task?.custom_class).toContain('bar-p2')
  })

  // --- Overdue coloring ---

  it('applies overdue (red) custom class for overdue items', async () => {
    mockGetGanttView.mockResolvedValue(ganttRespOverdue)
    renderPage()
    await waitFor(() => {
      expect(capturedCalls.length).toBeGreaterThan(0)
    })
    const tasks = capturedCalls[0].tasks as Array<{ custom_class?: string }>
    expect(tasks[0].custom_class).toContain('bar-overdue')
  })

  // --- Null dates handling ---

  it('uses today as fallback for null start/end dates', async () => {
    mockGetGanttView.mockResolvedValue(ganttRespNoDates)
    renderPage()
    await waitFor(() => {
      expect(capturedCalls.length).toBeGreaterThan(0)
    })
    const tasks = capturedCalls[0].tasks as Array<{ start: string; end: string }>
    // Should have valid date strings (not null)
    expect(tasks[0].start).toBeTruthy()
    expect(tasks[0].end).toBeTruthy()
  })

  // --- Expand/collapse ---

  it('clicking main item row expands sub-items', async () => {
    mockGetGanttView.mockResolvedValue(ganttRespWithItems)
    renderPage()
    await waitFor(() => {
      expect(capturedCalls.length).toBeGreaterThan(0)
    })
    // Initially collapsed: only main items
    const initialTasks = capturedCalls[0].tasks as unknown[]
    expect(initialTasks.length).toBe(2)

    // Click expand button for first main item
    const expandBtn = screen.getByTestId('gantt-expand-1')
    fireEvent.click(expandBtn)

    // After expand, gantt.refresh should be called with sub-items inserted
    await waitFor(() => {
      expect(mockGanttRefresh).toHaveBeenCalled()
    })
    expect(mockGanttRefresh.mock.calls.length).toBeGreaterThanOrEqual(1)
    // Find the call with 3 tasks
    const expandedCall = mockGanttRefresh.mock.calls.find((call: unknown[]) => (call[0] as unknown[]).length === 3)
    expect(expandedCall).toBeDefined()
  })

  it('clicking expanded main item collapses sub-items', async () => {
    mockGetGanttView.mockResolvedValue(ganttRespWithItems)
    renderPage()
    await waitFor(() => {
      expect(capturedCalls.length).toBeGreaterThan(0)
    })

    // Expand first
    const expandBtn = screen.getByTestId('gantt-expand-1')
    fireEvent.click(expandBtn)

    await waitFor(() => {
      expect(mockGanttRefresh).toHaveBeenCalled()
    })
    mockGanttRefresh.mockClear()

    // Click again to collapse
    fireEvent.click(expandBtn)
    await waitFor(() => {
      expect(mockGanttRefresh).toHaveBeenCalled()
    })
    const collapsedTasks = mockGanttRefresh.mock.calls[0][0] as unknown[]
    expect(collapsedTasks.length).toBe(2)
  })

  // --- Completion percentage ---

  it('renders completion percentage for each main item row', async () => {
    mockGetGanttView.mockResolvedValue(ganttRespWithItems)
    renderPage()
    await waitFor(() => {
      expect(screen.getByTestId('gantt-row-1')).toBeInTheDocument()
    })
    expect(screen.getByText('60%')).toBeInTheDocument()
    expect(screen.getByText('30%')).toBeInTheDocument()
  })

  // --- API call ---

  it('calls getGanttViewApi with teamId', async () => {
    mockGetGanttView.mockResolvedValue(emptyGanttResp)
    renderPage()
    await waitFor(() => {
      expect(mockGetGanttView).toHaveBeenCalledWith(1, undefined)
    })
  })
})
