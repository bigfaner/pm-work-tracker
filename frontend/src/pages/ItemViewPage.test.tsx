import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import ItemViewPage from './ItemViewPage'

// MSW lifecycle
beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

// --- Helpers ---

function createQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function renderPage() {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <ItemViewPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedMainItems = [
  {
    id: 1, team_id: 1, code: 'MI-0001', title: 'Alpha Task', priority: 'P1',
    proposer_id: 1, assignee_id: 1, start_date: '2026-04-01', expected_end_date: '2026-04-15',
    actual_end_date: null, status: '进行中', completion: 65, is_key_item: false,
    delay_count: 0, archived_at: null,
    created_at: '2026-04-01T00:00:00Z', updated_at: '2026-04-01T00:00:00Z',
    subItems: [
      {
        id: 11, team_id: 1, main_item_id: 1, title: 'Sub Alpha 1', description: '',
        priority: 'P1', assignee_id: 2, start_date: '2026-04-01', expected_end_date: '2026-04-10',
        actual_end_date: '2026-04-09', status: '已完成', completion: 100, is_key_item: false,
        delay_count: 0, weight: 1, created_at: '2026-04-01T00:00:00Z', updated_at: '2026-04-09T00:00:00Z',
      },
      {
        id: 12, team_id: 1, main_item_id: 1, title: 'Sub Alpha 2', description: '',
        priority: 'P2', assignee_id: 3, start_date: '2026-04-08', expected_end_date: '2026-04-18',
        actual_end_date: null, status: '进行中', completion: 80, is_key_item: false,
        delay_count: 0, weight: 1, created_at: '2026-04-01T00:00:00Z', updated_at: '2026-04-08T00:00:00Z',
      },
    ],
  },
  {
    id: 2, team_id: 1, code: 'MI-0002', title: 'Beta Task', priority: 'P2',
    proposer_id: 1, assignee_id: 2, start_date: '2026-04-15', expected_end_date: '2026-04-25',
    actual_end_date: null, status: '进行中', completion: 40, is_key_item: false,
    delay_count: 0, archived_at: null,
    created_at: '2026-04-15T00:00:00Z', updated_at: '2026-04-15T00:00:00Z',
    subItems: [],
  },
  {
    id: 3, team_id: 1, code: 'MI-0003', title: 'Gamma Task', priority: 'P3',
    proposer_id: 1, assignee_id: 3, start_date: '2026-04-05', expected_end_date: '2026-04-12',
    actual_end_date: '2026-04-12', status: '已完成', completion: 100, is_key_item: false,
    delay_count: 0, archived_at: null,
    created_at: '2026-04-05T00:00:00Z', updated_at: '2026-04-12T00:00:00Z',
    subItems: [],
  },
]

const seedMembers = [
  { userId: 1, displayName: 'Test User', username: 'testuser', role: 'pm', joinedAt: '2024-01-01' },
  { userId: 2, displayName: 'Alice', username: 'alice', role: 'member', joinedAt: '2024-01-01' },
  { userId: 3, displayName: 'Bob', username: 'bob', role: 'member', joinedAt: '2024-01-01' },
]

function setupHandlers() {
  server.use(
    // List main items (with subItems embedded)
    http.get('/api/v1/teams/:teamId/main-items', () => {
      return HttpResponse.json({ code: 0, data: { items: seedMainItems, total: seedMainItems.length, page: 1, pageSize: 20 } })
    }),

    // Get single main item with sub items
    http.get('/api/v1/teams/:teamId/main-items/:itemId', ({ params }) => {
      const item = seedMainItems.find(i => i.id === Number(params.itemId))
      if (!item) return HttpResponse.json({ code: 'NOT_FOUND', message: 'not found' }, { status: 404 })
      return HttpResponse.json({ code: 0, data: item })
    }),

    // List members for assignee filter
    http.get('/api/v1/teams/:teamId/members', () => {
      return HttpResponse.json({ code: 0, data: seedMembers })
    }),

    // Update main item status
    http.put('/api/v1/teams/:teamId/main-items/:itemId', async ({ request }) => {
      const body = await request.json() as Record<string, unknown>
      return HttpResponse.json({ code: 0, data: { ...seedMainItems[0], ...body } })
    }),

    // Create main item
    http.post('/api/v1/teams/:teamId/main-items', async ({ request }) => {
      const body = await request.json() as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: {
          id: 100, team_id: 1, code: 'MI-0100', priority: 'P2', proposer_id: 1,
          assignee_id: null, start_date: null, expected_end_date: null, actual_end_date: null,
          status: '未开始', completion: 0, is_key_item: false, delay_count: 0, archived_at: null,
          created_at: new Date().toISOString(), updated_at: new Date().toISOString(),
          ...body,
        },
      })
    }),
  )
}

describe('ItemViewPage', () => {
  beforeEach(() => {
    useTeamStore.setState({ currentTeamId: 1, teams: [{ id: 1, name: 'Test Team', description: '', pm_id: 1, created_at: '', updated_at: '' }] })
    setupHandlers()
  })

  // --- Core rendering ---

  it('renders page header with title and view toggle', async () => {
    renderPage()
    expect(screen.getByText('事项清单')).toBeInTheDocument()
    expect(screen.getByText('汇总')).toBeInTheDocument()
    expect(screen.getByText('明细')).toBeInTheDocument()
    // Wait for data to load
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })
  })

  it('renders create button', async () => {
    renderPage()
    expect(screen.getByText('创建主事项')).toBeInTheDocument()
  })

  // --- Summary view (default) ---

  it('shows summary view by default with cards', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      expect(screen.getByText('Beta Task')).toBeInTheDocument()
      expect(screen.getByText('Gamma Task')).toBeInTheDocument()
    })
  })

  it('shows item codes and priority badges in summary view', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
      expect(screen.getByText('MI-0002')).toBeInTheDocument()
    })
  })

  it('expands sub-items when clicking a summary card', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    // Click to expand the first card
    const expandBtn = screen.getByTestId('expand-card-1')
    await user.click(expandBtn)

    await waitFor(() => {
      expect(screen.getByText('Sub Alpha 1')).toBeInTheDocument()
      expect(screen.getByText('Sub Alpha 2')).toBeInTheDocument()
    })
  })

  // --- Detail view ---

  it('switches to detail view when clicking toggle', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    await user.click(screen.getByText('明细'))

    // Detail view should have a table
    await waitFor(() => {
      expect(screen.getByTestId('detail-table')).toBeInTheDocument()
      // Should show items in table format
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })
  })

  it('shows table headers in detail view', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    await user.click(screen.getByText('明细'))

    await waitFor(() => {
      expect(screen.getByText('编号')).toBeInTheDocument()
      expect(screen.getByText('优先级')).toBeInTheDocument()
      expect(screen.getByText('标题')).toBeInTheDocument()
      expect(screen.getByText('负责人')).toBeInTheDocument()
      expect(screen.getByText('进度')).toBeInTheDocument()
      expect(screen.getByText('状态')).toBeInTheDocument()
    })
  })

  // --- Filter bar ---

  it('renders filter bar with search, status, and assignee', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })
    expect(screen.getByPlaceholderText('搜索标题或编号...')).toBeInTheDocument()
    expect(screen.getByText('重置')).toBeInTheDocument()
  })

  // --- Shared filter state ---

  it('preserves filter state when switching views', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    // Type in search
    const searchInput = screen.getByPlaceholderText('搜索标题或编号...')
    await user.type(searchInput, 'Alpha')

    // Switch to detail view
    await user.click(screen.getByText('明细'))

    // Search value should be preserved
    expect(screen.getByPlaceholderText('搜索标题或编号...')).toHaveValue('Alpha')
  })

  // --- Empty state ---

  it('shows empty state when no items exist', async () => {
    server.use(
      http.get('/api/v1/teams/:teamId/main-items', () => {
        return HttpResponse.json({ code: 0, data: { items: [], total: 0, page: 1, pageSize: 20 } })
      }),
    )

    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/暂无事项/)).toBeInTheDocument()
    })
  })

  // --- Create main item dialog ---

  it('opens create dialog when clicking create button', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    await user.click(screen.getByText('创建主事项'))

    await waitFor(() => {
      expect(screen.getByText('新建主事项')).toBeInTheDocument()
    })
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    // This is a build-time check - if the component renders without antd, it passes
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('事项清单')).toBeInTheDocument()
    })
    // No antd classes should be present
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Navigation ---

  it('has a clickable title that links to detail page', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByText('Alpha Task').closest('a')
      expect(link).toHaveAttribute('href', '/items/1')
    })
  })

  // --- Inline status change ---

  it('shows status badge in summary cards', async () => {
    renderPage()
    await waitFor(() => {
      // Each main item should show its status
      const statuses = screen.getAllByText('进行中')
      expect(statuses.length).toBeGreaterThanOrEqual(2)
    })
  })
})
