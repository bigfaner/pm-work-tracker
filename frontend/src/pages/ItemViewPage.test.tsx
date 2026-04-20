import { describe, it, expect, vi, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
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
    id: 1, teamId: 1, code: 'MI-0001', title: 'Alpha Task', priority: 'P1',
    proposerId: 1, assigneeId: 1, startDate: '2026-04-01', expectedEndDate: '2026-04-15',
    actualEndDate: null, status: 'progressing', completion: 65, isKeyItem: false,
    delayCount: 0, archivedAt: null,
    createdAt: '2026-04-01T00:00:00Z', updatedAt: '2026-04-01T00:00:00Z',
    subItems: [
      {
        id: 11, teamId: 1, mainItemId: 1, title: 'Sub Alpha 1', description: '',
        priority: 'P1', assigneeId: 2, startDate: '2026-04-01', expectedEndDate: '2026-04-10',
        actualEndDate: '2026-04-09', status: 'completed', completion: 100, isKeyItem: false,
        delayCount: 0, weight: 1, createdAt: '2026-04-01T00:00:00Z', updatedAt: '2026-04-09T00:00:00Z',
      },
      {
        id: 12, teamId: 1, mainItemId: 1, title: 'Sub Alpha 2', description: '',
        priority: 'P2', assigneeId: 3, startDate: '2026-04-08', expectedEndDate: '2026-04-18',
        actualEndDate: null, status: 'progressing', completion: 80, isKeyItem: false,
        delayCount: 0, weight: 1, createdAt: '2026-04-01T00:00:00Z', updatedAt: '2026-04-08T00:00:00Z',
      },
    ],
  },
  {
    id: 2, teamId: 1, code: 'MI-0002', title: 'Beta Task', priority: 'P2',
    proposerId: 1, assigneeId: 2, startDate: '2026-04-15', expectedEndDate: '2026-04-25',
    actualEndDate: null, status: 'progressing', completion: 40, isKeyItem: false,
    delayCount: 0, archivedAt: null,
    createdAt: '2026-04-15T00:00:00Z', updatedAt: '2026-04-15T00:00:00Z',
    subItems: [],
  },
  {
    id: 3, teamId: 1, code: 'MI-0003', title: 'Gamma Task', priority: 'P3',
    proposerId: 1, assigneeId: 3, startDate: '2026-04-05', expectedEndDate: '2026-04-12',
    actualEndDate: '2026-04-12', status: 'completed', completion: 100, isKeyItem: false,
    delayCount: 0, archivedAt: null,
    createdAt: '2026-04-05T00:00:00Z', updatedAt: '2026-04-12T00:00:00Z',
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

    // List sub-items for a main item
    http.get('/api/v1/teams/:teamId/main-items/:mainId/sub-items', ({ params }) => {
      const item = seedMainItems.find(i => i.id === Number(params.mainId))
      const subs = item?.subItems || []
      return HttpResponse.json({ code: 0, data: { items: subs, total: subs.length, page: 1, pageSize: 20 } })
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
          id: 100, teamId: 1, code: 'MI-0100', priority: 'P2', proposerId: 1,
          assigneeId: null, startDate: null, expectedEndDate: null, actualEndDate: null,
          status: 'pending', completion: 0, isKeyItem: false, delayCount: 0, archivedAt: null,
          createdAt: new Date().toISOString(), updatedAt: new Date().toISOString(),
          ...body,
        },
      })
    }),

    // Create sub-item
    http.post('/api/v1/teams/:teamId/main-items/:mainId/sub-items', async ({ request }) => {
      const body = await request.json() as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: {
          id: 200, teamId: 1, mainItemId: Number(body.main_item_id), description: '',
          priority: 'P2', assigneeId: null, startDate: null, expectedEndDate: null, actualEndDate: null,
          status: 'pending', completion: 0, isKeyItem: false, delayCount: 0, weight: 1,
          createdAt: new Date().toISOString(), updatedAt: new Date().toISOString(),
          ...body,
        },
      })
    }),
  )
}

describe('ItemViewPage', () => {
  beforeEach(() => {
    useTeamStore.setState({ currentTeamId: 1, teams: [{ id: 1, name: 'Test Team', description: '', pmId: 1, createdAt: '', updatedAt: '' }] })
    useAuthStore.getState().setPermissions({
      isSuperAdmin: false,
      teamPermissions: { 1: ['main_item:create', 'progress:update'] },
    })
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
    expect(screen.getByText('新增主事项')).toBeInTheDocument()
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

    await user.click(screen.getByText('新增主事项'))

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

  // --- Create sub-item dialog (e2e) ---

  it('create sub-item dialog has required fields: priority, startDate, expectedEndDate', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    // Click "新增子事项" on the first card
    const addSubBtns = screen.getAllByRole('button', { name: /新增子事项/ })
    await user.click(addSubBtns[0])

    await waitFor(() => {
      const dialogTitles = screen.getAllByText(/新增子事项/).filter(el => el.closest('[role="dialog"]'))
      expect(dialogTitles.length).toBe(1)
    })

    // Priority should have required marker
    const priorityLabels = screen.getAllByText(/优先级/)
    const priorityWithRequired = priorityLabels.find(el => el.closest('label')?.innerHTML.includes('*'))
    expect(priorityWithRequired).toBeTruthy()

    // Start date field should exist with required marker
    const startDateLabels = screen.getAllByText(/开始时间/)
    const startDateWithRequired = startDateLabels.find(el => el.closest('label')?.innerHTML.includes('*'))
    expect(startDateWithRequired).toBeTruthy()

    // Expected end date should have required marker
    const endDateLabels = screen.getAllByText(/预期完成时间/)
    const endDateWithRequired = endDateLabels.find(el => el.closest('label')?.innerHTML.includes('*'))
    expect(endDateWithRequired).toBeTruthy()
  })

  it('create sub-item submit is disabled when required fields are empty', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    const addSubBtns = screen.getAllByRole('button', { name: /新增子事项/ })
    await user.click(addSubBtns[0])

    await waitFor(() => {
      const dialogTitles = screen.getAllByText(/新增子事项/).filter(el => el.closest('[role="dialog"]'))
      expect(dialogTitles.length).toBe(1)
    })

    // Fill only title, leave other required fields empty
    await user.type(screen.getByPlaceholderText('请输入子事项标题'), 'Test Sub')

    const submitBtn = screen.getByRole('button', { name: '确认' })
    expect(submitBtn).toBeDisabled()
  })

  it('create sub-item submit is enabled when all required fields are filled', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    const addSubBtns = screen.getAllByRole('button', { name: /新增子事项/ })
    await user.click(addSubBtns[0])

    await waitFor(() => {
      const dialogTitles = screen.getAllByText(/新增子事项/).filter(el => el.closest('[role="dialog"]'))
      expect(dialogTitles.length).toBe(1)
    })

    // Fill all required fields
    await user.type(screen.getByPlaceholderText('请输入子事项标题'), 'Test Sub')

    // Select priority
    const priorityLabels = screen.getAllByText(/优先级/)
    const priorityLabelInDialog = priorityLabels.find(el => el.closest('label')?.innerHTML.includes('*'))
    const priorityContainer = priorityLabelInDialog!.closest('div')?.parentElement
    const selectTrigger = priorityContainer?.querySelector('button')
    await user.click(selectTrigger!)
    await user.click(screen.getByRole('option', { name: 'P2' }))

    // Select assignee
    const allSelects = document.querySelectorAll('[role="dialog"] button[role="combobox"], [role="dialog"] button[data-state]')
    const assigneeBtn = Array.from(allSelects).find(btn => btn.textContent?.includes('不指定'))
    if (assigneeBtn) {
      await user.click(assigneeBtn)
      await waitFor(() => {
        expect(screen.getByRole('option', { name: 'Test User' })).toBeInTheDocument()
      })
      await user.click(screen.getByRole('option', { name: 'Test User' }))
    }

    // Fill date inputs
    const allDateInputs = document.querySelectorAll('input[type="date"]')
    const dialogDateInputs = Array.from(allDateInputs).filter(input => {
      const label = input.closest('div')?.querySelector('label')
      return label && (label.textContent?.includes('开始时间') || label.textContent?.includes('预期完成时间'))
    })
    expect(dialogDateInputs.length).toBe(2)
    fireEvent.change(dialogDateInputs[0], { target: { value: '2026-04-20' } })
    fireEvent.change(dialogDateInputs[1], { target: { value: '2026-04-30' } })

    await waitFor(() => {
      const submitBtn = screen.getByRole('button', { name: '确认' })
      expect(submitBtn).toBeEnabled()
    })
  })
})
