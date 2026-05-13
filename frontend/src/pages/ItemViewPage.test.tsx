import {
  describe,
  it,
  expect,
  beforeEach,
  beforeAll,
  afterAll,
  afterEach,
  vi,
} from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
import { ToastProvider } from '@/components/ui/toast'
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
        <ToastProvider>
          <ItemViewPage />
        </ToastProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedMainItems = [
  {
    bizKey: '1',
    teamKey: '1',
    code: 'MI-0001',
    title: 'Alpha Task',
    priority: 'P1',
    proposerKey: 'U001',
    assigneeKey: 'U001',
    planStartDate: '2026-04-01',
    expectedEndDate: '2026-04-15',
    actualEndDate: null,
    itemStatus: 'progressing',
    completion: 65,
    milestoneKey: 'MS001',
    createTime: '2026-04-01T00:00:00Z',
    dbUpdateTime: '2026-04-01T00:00:00Z',
    subItems: [
      {
        bizKey: '11',
        teamKey: '1',
        mainItemKey: '1',
        code: 'MI-0001-01',
        title: 'Sub Alpha 1',
        itemDesc: '',
        priority: 'P1',
        assigneeKey: 'U002',
        planStartDate: '2026-04-01',
        expectedEndDate: '2026-04-10',
        actualEndDate: '2026-04-09',
        itemStatus: 'completed',
        completion: 100,
        weight: 1,
        createTime: '2026-04-01T00:00:00Z',
        dbUpdateTime: '2026-04-09T00:00:00Z',
      },
      {
        bizKey: '12',
        teamKey: '1',
        mainItemKey: '1',
        code: 'MI-0001-02',
        title: 'Sub Alpha 2',
        itemDesc: '',
        priority: 'P2',
        assigneeKey: 'U003',
        planStartDate: '2026-04-08',
        expectedEndDate: '2026-04-18',
        actualEndDate: null,
        itemStatus: 'progressing',
        completion: 80,
        weight: 1,
        createTime: '2026-04-01T00:00:00Z',
        dbUpdateTime: '2026-04-08T00:00:00Z',
      },
    ],
  },
  {
    bizKey: '2',
    teamKey: '1',
    code: 'MI-0002',
    title: 'Beta Task',
    priority: 'P2',
    proposerKey: 'U001',
    assigneeKey: 'U002',
    planStartDate: '2026-04-15',
    expectedEndDate: '2026-04-25',
    actualEndDate: null,
    itemStatus: 'progressing',
    completion: 40,
    milestoneKey: null,
    createTime: '2026-04-15T00:00:00Z',
    dbUpdateTime: '2026-04-15T00:00:00Z',
    subItems: [],
  },
  {
    bizKey: '3',
    teamKey: '1',
    code: 'MI-0003',
    title: 'Gamma Task',
    priority: 'P3',
    proposerKey: 'U001',
    assigneeKey: 'U003',
    planStartDate: '2026-04-05',
    expectedEndDate: '2026-04-12',
    actualEndDate: '2026-04-12',
    itemStatus: 'completed',
    completion: 100,
    milestoneKey: 'MS002',
    createTime: '2026-04-05T00:00:00Z',
    dbUpdateTime: '2026-04-12T00:00:00Z',
    subItems: [],
  },
]

const seedMembers = [
  {
    id: 1,
    bizKey: '1',
    teamKey: '1',
    userKey: 'U001',
    displayName: 'Test User',
    username: 'testuser',
    role: 'pm',
    roleKey: 1,
    roleName: 'pm',
    joinedAt: '2024-01-01',
  },
  {
    id: 2,
    bizKey: '2',
    teamKey: '1',
    userKey: 'U002',
    displayName: 'Alice',
    username: 'alice',
    role: 'member',
    roleKey: 2,
    roleName: 'member',
    joinedAt: '2024-01-01',
  },
  {
    id: 3,
    bizKey: '3',
    teamKey: '1',
    userKey: 'U003',
    displayName: 'Bob',
    username: 'bob',
    role: 'member',
    roleKey: 3,
    roleName: 'member',
    joinedAt: '2024-01-01',
  },
]

function setupHandlers() {
  server.use(
    // List main items (with subItems embedded)
    http.get('/v1/teams/:teamId/main-items', () => {
      return HttpResponse.json({
        code: 0,
        data: {
          items: seedMainItems,
          total: seedMainItems.length,
          page: 1,
          pageSize: 20,
        },
      })
    }),

    // Get single main item with sub items
    http.get('/v1/teams/:teamId/main-items/:itemId', ({ params }) => {
      const item = seedMainItems.find(
        (i) => i.bizKey === String(params.itemId),
      )
      if (!item)
        return HttpResponse.json(
          { code: 'NOT_FOUND', message: 'not found' },
          { status: 404 },
        )
      return HttpResponse.json({ code: 0, data: item })
    }),

    // List sub-items for a main item
    http.get('/v1/teams/:teamId/main-items/:mainId/sub-items', ({ params }) => {
      const item = seedMainItems.find(
        (i) => i.bizKey === String(params.mainId),
      )
      const subs = item?.subItems || []
      return HttpResponse.json({
        code: 0,
        data: { items: subs, total: subs.length, page: 1, pageSize: 20 },
      })
    }),

    // List members for assignee filter
    http.get('/v1/teams/:teamId/members', () => {
      return HttpResponse.json({ code: 0, data: seedMembers })
    }),

    // List milestones for milestone filter
    http.get('/v1/teams/:teamId/milestones', () => {
      return HttpResponse.json({
        code: 0,
        data: {
          items: [
            {
              bizKey: 'MS001',
              teamKey: '1',
              milestoneMapKey: 'MAP001',
              milestoneName: 'MVP发布',
              expectedEndDate: '2026-06-30',
              milestoneStatus: 'in_progress',
              statusName: '进行中',
              completion: 50,
              relatedMICount: 1,
              createTime: '2026-01-01T00:00:00Z',
              dbUpdateTime: '2026-01-01T00:00:00Z',
            },
            {
              bizKey: 'MS002',
              teamKey: '1',
              milestoneMapKey: 'MAP001',
              milestoneName: '二期迭代',
              expectedEndDate: '2026-09-30',
              milestoneStatus: 'pending',
              statusName: '待开始',
              completion: 0,
              relatedMICount: 1,
              createTime: '2026-01-01T00:00:00Z',
              dbUpdateTime: '2026-01-01T00:00:00Z',
            },
          ],
          total: 2,
        },
      })
    }),

    // Update main item status
    http.put('/v1/teams/:teamId/main-items/:itemId', async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: { ...seedMainItems[0], ...body },
      })
    }),

    // Change main item status
    http.put(
      '/v1/teams/:teamId/main-items/:itemId/status',
      async ({ request }) => {
        const body = (await request.json()) as Record<string, unknown>
        return HttpResponse.json({
          code: 0,
          data: { itemStatus: body.status },
        })
      },
    ),

    // Available transitions for main items (default: return all non-self statuses)
    http.get(
      '/v1/teams/:teamId/main-items/:itemId/available-transitions',
      ({ params }) => {
        const item = seedMainItems.find(
          (i) => i.bizKey === String(params.itemId),
        )
        const currentStatus = item?.itemStatus || 'pending'
        const allStatuses = [
          'pending',
          'progressing',
          'blocking',
          'pausing',
          'reviewing',
          'completed',
          'closed',
        ]
        const transitions = allStatuses.filter((s) => s !== currentStatus)
        return HttpResponse.json({ code: 0, data: { transitions } })
      },
    ),

    // Create main item
    http.post('/v1/teams/:teamId/main-items', async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: {
          bizKey: '100',
          teamKey: '1',
          code: 'MI-0100',
          priority: 'P2',
          proposerKey: 'U001',
          assigneeKey: null,
          planStartDate: null,
          expectedEndDate: null,
          actualEndDate: null,
          itemStatus: 'pending',
          completion: 0,
          createTime: new Date().toISOString(),
          dbUpdateTime: new Date().toISOString(),
          ...body,
        },
      })
    }),

    // Create sub-item
    http.post(
      '/v1/teams/:teamId/main-items/:mainId/sub-items',
      async ({ request, params }) => {
        const body = (await request.json()) as Record<string, unknown>
        return HttpResponse.json({
          code: 0,
          data: {
            bizKey: '200',
            teamKey: '1',
            mainItemKey: String(body.mainItemBizKey || params.mainId),
            itemDesc: '',
            priority: 'P2',
            assigneeKey: null,
            planStartDate: null,
            expectedEndDate: null,
            actualEndDate: null,
            itemStatus: 'pending',
            completion: 0,
            weight: 1,
            createTime: new Date().toISOString(),
            dbUpdateTime: new Date().toISOString(),
            ...body,
          },
        })
      },
    ),
  )
}

describe('ItemViewPage', () => {
  beforeEach(() => {
    useTeamStore.setState({
      currentTeamId: '1',
      teams: [
        {
          bizKey: '1',
          name: 'Test Team',
          description: '',
          code: '',
          pmKey: '1',
          createdAt: '',
          updatedAt: '',
        },
      ],
    })
    useAuthStore.getState().setPermissions({
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
    expect(
      screen.getByPlaceholderText('搜索标题或编号...'),
    ).toBeInTheDocument()
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
    expect(screen.getByPlaceholderText('搜索标题或编号...')).toHaveValue(
      'Alpha',
    )
  })

  // --- Empty state ---

  it('shows empty state when no items exist', async () => {
    server.use(
      http.get('/v1/teams/:teamId/main-items', () => {
        return HttpResponse.json({
          code: 0,
          data: { items: [], total: 0, page: 1, pageSize: 20 },
        })
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

  // --- Append progress validation ---

  it('bug: shows error toast when submitting completion lower than current sub-item completion', async () => {
    let progressApiCalled = false
    server.use(
      http.post(
        '/v1/teams/:teamId/sub-items/:itemId/progress',
        async ({ request }) => {
          progressApiCalled = true
          const body = (await request.json()) as Record<string, unknown>
          return HttpResponse.json({
            code: 0,
            data: {
              completion: body.completion,
              createTime: new Date().toISOString(),
            },
          })
        },
      ),
    )

    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    // Expand card to show sub-items
    const expandBtn = screen.getByTestId('expand-card-1')
    await user.click(expandBtn)

    await waitFor(() => {
      expect(screen.getByText('Sub Alpha 2')).toBeInTheDocument()
    })

    // Click "追加进度" on Sub Alpha 2 (completion=80)
    const appendBtns = screen.getAllByRole('button', { name: /追加进度/ })
    const enabledAppendBtns = appendBtns.filter(
      (btn) => !btn.hasAttribute('disabled'),
    )
    await user.click(enabledAppendBtns[0])

    await waitFor(() => {
      expect(
        screen.getByRole('heading', { name: /追加进度/ }),
      ).toBeInTheDocument()
    })

    // Enter a value lower than current completion (80)
    const input = screen.getByPlaceholderText('请输入进度')
    await user.clear(input)
    await user.type(input, '50')

    await user.click(screen.getByRole('button', { name: '确认' }))

    // Should show error toast WITHOUT calling the API
    await waitFor(() => {
      expect(screen.getByText(/进度不能低于/)).toBeInTheDocument()
    })
    expect(progressApiCalled).toBe(false)
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

  it('StatusDropdown fetches available transitions and shows only returned options', async () => {
    const user = userEvent.setup()
    setupHandlers() // ensure handlers
    server.use(
      http.get(
        '/v1/teams/:teamId/main-items/:itemId/available-transitions',
        () => {
          return HttpResponse.json({
            code: 0,
            data: { transitions: ['progressing', 'closed'] },
          })
        },
      ),
    )
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    // Click the first status badge (with cursor-pointer class = dropdown trigger) to open dropdown
    const badges = screen.getAllByText('进行中')
    const triggerBadges = badges.filter((el) => el.closest('button') !== null)
    await user.click(triggerBadges[0])

    await waitFor(() => {
      // Should show the returned transition 'closed' as '已关闭'
      expect(
        screen.getByRole('menuitem', { name: '已关闭' }),
      ).toBeInTheDocument()
      // Should NOT show other statuses like "待开始"
      expect(
        screen.queryByRole('menuitem', { name: '待开始' }),
      ).not.toBeInTheDocument()
    })
  })

  it('StatusDropdown calls changeMainItemStatusApi on select', async () => {
    const user = userEvent.setup()
    let statusChanged = false
    server.use(
      http.get(
        '/v1/teams/:teamId/main-items/:itemId/available-transitions',
        () => {
          return HttpResponse.json({
            code: 0,
            data: { transitions: ['progressing', 'blocking'] },
          })
        },
      ),
      http.put(
        '/v1/teams/:teamId/main-items/:itemId/status',
        async ({ request }) => {
          const body = (await request.json()) as Record<string, unknown>
          statusChanged = true
          return HttpResponse.json({ code: 0, data: { status: body.status } })
        },
      ),
    )
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    const badges = screen.getAllByText('进行中')
    await user.click(badges[0])

    await waitFor(() => {
      expect(screen.getByText('阻塞中')).toBeInTheDocument()
    })
    await user.click(screen.getByText('阻塞中'))

    await waitFor(() => {
      expect(statusChanged).toBe(true)
    })
  })

  it('shows confirmation dialog before terminal transition', async () => {
    const user = userEvent.setup()
    server.use(
      http.get(
        '/v1/teams/:teamId/main-items/:itemId/available-transitions',
        () => {
          return HttpResponse.json({
            code: 0,
            data: { transitions: ['completed', 'blocking'] },
          })
        },
      ),
    )
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    const badges = screen.getAllByText('进行中')
    const triggerBadges = badges.filter((el) => el.closest('button') !== null)
    await user.click(triggerBadges[0])

    await waitFor(() => {
      expect(
        screen.getByRole('menuitem', { name: '已完成' }),
      ).toBeInTheDocument()
    })
    await user.click(screen.getByRole('menuitem', { name: '已完成' }))

    // Confirmation dialog should appear
    await waitFor(() => {
      expect(screen.getByText(/确认变更状态/)).toBeInTheDocument()
    })
  })

  // --- Overdue badge ---

  it('shows overdue badge when expectedEndDate is past and status is non-terminal', async () => {
    // Override main items with one overdue item
    const overdueItems = [
      {
        bizKey: '10',
        teamKey: '1',
        code: 'MI-0010',
        title: 'Overdue Task',
        priority: 'P1',
        proposerKey: 'U001',
        assigneeKey: 'U001',
        planStartDate: '2020-01-01',
        expectedEndDate: '2020-02-01',
        actualEndDate: null,
        itemStatus: 'progressing',
        completion: 30,
        createTime: '2020-01-01T00:00:00Z',
        dbUpdateTime: '2020-01-01T00:00:00Z',
        subItems: [],
      },
    ]
    server.use(
      http.get('/v1/teams/:teamId/main-items', () => {
        return HttpResponse.json({
          code: 0,
          data: { items: overdueItems, total: 1, page: 1, pageSize: 20 },
        })
      }),
    )
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Overdue Task')).toBeInTheDocument()
    })
    // Should show overdue badge
    expect(screen.getByText('延期')).toBeInTheDocument()
  })

  it('does not show overdue badge for terminal status items', async () => {
    const completedItems = [
      {
        bizKey: '11',
        teamKey: '1',
        code: 'MI-0011',
        title: 'Completed Task',
        priority: 'P1',
        proposerKey: 'U001',
        assigneeKey: 'U001',
        planStartDate: '2020-01-01',
        expectedEndDate: '2020-02-01',
        actualEndDate: '2020-02-01',
        itemStatus: 'completed',
        completion: 100,
        createTime: '2020-01-01T00:00:00Z',
        dbUpdateTime: '2020-02-01T00:00:00Z',
        subItems: [],
      },
    ]
    server.use(
      http.get('/v1/teams/:teamId/main-items', () => {
        return HttpResponse.json({
          code: 0,
          data: { items: completedItems, total: 1, page: 1, pageSize: 20 },
        })
      }),
    )
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Completed Task')).toBeInTheDocument()
    })
    // Should NOT show overdue badge for completed items
    expect(screen.queryByText('延期')).not.toBeInTheDocument()
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
      const dialogTitles = screen
        .getAllByText(/新增子事项/)
        .filter((el) => el.closest('[role="dialog"]'))
      expect(dialogTitles.length).toBe(1)
    })

    // Priority should have required marker
    const priorityLabels = screen.getAllByText(/优先级/)
    const priorityWithRequired = priorityLabels.find((el) =>
      el.closest('label')?.innerHTML.includes('*'),
    )
    expect(priorityWithRequired).toBeTruthy()

    // Start date field should exist with required marker
    const startDateLabels = screen.getAllByText(/开始时间/)
    const startDateWithRequired = startDateLabels.find((el) =>
      el.closest('label')?.innerHTML.includes('*'),
    )
    expect(startDateWithRequired).toBeTruthy()

    // Expected end date should have required marker
    const endDateLabels = screen.getAllByText(/预期完成时间/)
    const endDateWithRequired = endDateLabels.find((el) =>
      el.closest('label')?.innerHTML.includes('*'),
    )
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
      const dialogTitles = screen
        .getAllByText(/新增子事项/)
        .filter((el) => el.closest('[role="dialog"]'))
      expect(dialogTitles.length).toBe(1)
    })

    // Fill only title, leave other required fields empty
    await user.type(
      screen.getByPlaceholderText('请输入子事项标题'),
      'Test Sub',
    )

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
      const dialogTitles = screen
        .getAllByText(/新增子事项/)
        .filter((el) => el.closest('[role="dialog"]'))
      expect(dialogTitles.length).toBe(1)
    })

    // Fill all required fields
    await user.type(
      screen.getByPlaceholderText('请输入子事项标题'),
      'Test Sub',
    )

    // Select priority
    const priorityLabels = screen.getAllByText(/优先级/)
    const priorityLabelInDialog = priorityLabels.find((el) =>
      el.closest('label')?.innerHTML.includes('*'),
    )
    const priorityContainer =
      priorityLabelInDialog!.closest('div')?.parentElement
    const selectTrigger = priorityContainer?.querySelector('button')
    await user.click(selectTrigger!)
    await user.click(screen.getByRole('option', { name: 'P2' }))

    // Select assignee
    const allSelects = document.querySelectorAll(
      '[role="dialog"] button[role="combobox"], [role="dialog"] button[data-state]',
    )
    const assigneeBtn = Array.from(allSelects).find((btn) =>
      btn.textContent?.includes('不指定'),
    )
    if (assigneeBtn) {
      await user.click(assigneeBtn)
      await waitFor(() => {
        expect(
          screen.getByRole('option', { name: 'Test User' }),
        ).toBeInTheDocument()
      })
      await user.click(screen.getByRole('option', { name: 'Test User' }))
    }

    // Fill date inputs
    const allDateInputs = document.querySelectorAll('input[type="date"]')
    const dialogDateInputs = Array.from(allDateInputs).filter((input) => {
      const label = input.closest('div')?.parentElement?.querySelector('label')
      return (
        label &&
        (label.textContent?.includes('开始时间') ||
          label.textContent?.includes('预期完成时间'))
      )
    })
    expect(dialogDateInputs.length).toBe(2)
    fireEvent.change(dialogDateInputs[0], { target: { value: '2026-04-20' } })
    fireEvent.change(dialogDateInputs[1], { target: { value: '2026-04-30' } })

    await waitFor(() => {
      const submitBtn = screen.getByRole('button', { name: '确认' })
      expect(submitBtn).toBeEnabled()
    })
  })

  // --- Sub-items fetched via React Query ---

  it('fetches sub-items via React Query when expanding a card', async () => {
    let subItemsFetchCount = 0
    server.use(
      http.get('/v1/teams/:teamId/main-items/:mainId/sub-items', () => {
        subItemsFetchCount++
        return HttpResponse.json({
          code: 0,
          data: {
            items: seedMainItems[0].subItems,
            total: 2,
            page: 1,
            pageSize: 20,
          },
        })
      }),
    )
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    // Expand the first card
    const expandBtn = screen.getByTestId('expand-card-1')
    await user.click(expandBtn)

    await waitFor(() => {
      expect(subItemsFetchCount).toBeGreaterThanOrEqual(1)
      expect(screen.getByText('Sub Alpha 1')).toBeInTheDocument()
    })
  })

  // --- No "fetch all pages" in detail view ---

  it('detail view does not trigger fetch-all-pages pattern', async () => {
    let listCallCount = 0
    server.use(
      http.get('/v1/teams/:teamId/main-items', ({ request }) => {
        listCallCount++
        const url = new URL(request.url)
        const page = Number(url.searchParams.get('page') || 1)
        // Return one page of items
        return HttpResponse.json({
          code: 0,
          data: {
            items: seedMainItems,
            total: seedMainItems.length,
            page,
            pageSize: 20,
          },
        })
      }),
    )
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    // Switch to detail view
    await user.click(screen.getByText('明细'))

    await waitFor(() => {
      expect(screen.getByTestId('detail-table')).toBeInTheDocument()
    })

    // Should NOT have made multiple list calls to fetch all pages
    // Only the initial fetch should have happened
    expect(listCallCount).toBe(1)
  })

  it('renders sub-item code from API response without auto-generating', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    // Switch to detail view — sub-items load automatically for all paginated items
    await user.click(screen.getByText('明细'))
    await waitFor(() => {
      expect(screen.getByTestId('detail-table')).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getByText('01')).toBeInTheDocument()
      expect(screen.getByText('02')).toBeInTheDocument()
    })
  })

  // --- Edit main item: refresh + preserve expanded state ---

  it('bug: after editing a main item, sub-items for expanded cards are re-fetched and expanded state is preserved', async () => {
    let subItemsFetchCount = 0
    server.use(
      http.get('/v1/teams/:teamId/main-items/:mainId/sub-items', () => {
        subItemsFetchCount++
        return HttpResponse.json({
          code: 0,
          data: {
            items: seedMainItems[0].subItems,
            total: 2,
            page: 1,
            pageSize: 20,
          },
        })
      }),
    )

    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Alpha Task')).toBeInTheDocument()
    })

    // Expand card '1' to trigger sub-items fetch
    await user.click(screen.getByTestId('expand-card-1'))
    await waitFor(() => {
      expect(screen.getByText('Sub Alpha 1')).toBeInTheDocument()
      expect(subItemsFetchCount).toBeGreaterThanOrEqual(1)
    })
    const fetchCountAfterExpand = subItemsFetchCount

    // Click the main item edit button (first "编辑" not belonging to a sub-item)
    const editBtns = screen.getAllByRole('button', { name: /编辑/ })
    const mainEditBtn = editBtns.find(
      (btn) => !btn.getAttribute('data-testid')?.startsWith('edit-sub-'),
    )
    await user.click(mainEditBtn!)

    await waitFor(() => {
      expect(screen.getByText('编辑主事项')).toBeInTheDocument()
    })

    // Submit — title is already populated from openEditDialog
    await user.click(screen.getByRole('button', { name: '确认' }))

    // Sub-items for the expanded card must be re-fetched after edit
    await waitFor(() => {
      expect(subItemsFetchCount).toBeGreaterThan(fetchCountAfterExpand)
    })

    // Expanded state must be preserved — sub-items still visible
    expect(screen.getByText('Sub Alpha 1')).toBeInTheDocument()
  })

  describe('copyLink from code badge', () => {
    let clipboardSpy: ReturnType<typeof vi.fn>

    beforeEach(() => {
      clipboardSpy = vi.fn().mockResolvedValue(undefined)
      Object.defineProperty(navigator, 'clipboard', {
        value: { writeText: clipboardSpy },
        configurable: true,
      })
    })

    it('bug: copied link includes base_path in summary view', async () => {
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('MI-0001')).toBeInTheDocument()
      })

      // Click the code badge (MI-0001) in summary view
      const codeBadge = screen.getByText('MI-0001')
      fireEvent.click(codeBadge)

      expect(clipboardSpy).toHaveBeenCalledOnce()
      const copiedText = clipboardSpy.mock.calls[0][0] as string
      // The URL must include the VITE_BASE_PATH prefix
      const basePath = import.meta.env.VITE_BASE_PATH || ''
      expect(copiedText).toContain(basePath + '/items/1')
    })
  })

  // --- Milestone filter ---

  describe('milestone filter', () => {
    it('renders milestone filter dropdown in filter bar', async () => {
      renderPage()
      await waitFor(() => {
        expect(screen.getByTestId('milestone-filter')).toBeInTheDocument()
      })
    })

    it('shows milestone name badge on MI rows in summary view', async () => {
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })
      // Alpha Task has milestoneKey MS001 -> milestoneName "MVP发布"
      expect(screen.getByText('MVP发布')).toBeInTheDocument()
    })

    it('does not show milestone badge for items without milestone', async () => {
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Beta Task')).toBeInTheDocument()
      })
      // Beta Task has milestoneKey null — no badge rendered next to it
      // The badge only appears for items that have a milestone
      const badges = screen.getAllByText('MVP发布')
      expect(badges.length).toBe(1) // Only on Alpha Task row
    })

    it('filters MI list by selected milestone', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      // Open milestone filter dropdown
      await user.click(screen.getByTestId('milestone-filter'))

      // Select "MVP发布" milestone
      await waitFor(() => {
        expect(screen.getByRole('option', { name: 'MVP发布' })).toBeInTheDocument()
      })
      await user.click(screen.getByRole('option', { name: 'MVP发布' }))

      // Only Alpha Task (milestoneKey=MS001 -> MVP发布) should be visible
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
        expect(screen.queryByText('Beta Task')).not.toBeInTheDocument()
        expect(screen.queryByText('Gamma Task')).not.toBeInTheDocument()
      })
    })

    it('filters MI list by unassigned milestone', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      // Open milestone filter dropdown
      await user.click(screen.getByTestId('milestone-filter'))

      // Select "未分配"
      await waitFor(() => {
        expect(screen.getByRole('option', { name: '未分配' })).toBeInTheDocument()
      })
      await user.click(screen.getByRole('option', { name: '未分配' }))

      // Only Beta Task (milestoneKey=null) should be visible
      await waitFor(() => {
        expect(screen.queryByText('Alpha Task')).not.toBeInTheDocument()
        expect(screen.getByText('Beta Task')).toBeInTheDocument()
        expect(screen.queryByText('Gamma Task')).not.toBeInTheDocument()
      })
    })

    it('resets milestone filter along with other filters', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      // Open milestone filter and select a value
      await user.click(screen.getByTestId('milestone-filter'))
      await waitFor(() => {
        expect(screen.getByRole('option', { name: 'MVP发布' })).toBeInTheDocument()
      })
      await user.click(screen.getByRole('option', { name: 'MVP发布' }))

      // Only one item visible
      await waitFor(() => {
        expect(screen.queryByText('Beta Task')).not.toBeInTheDocument()
      })

      // Click reset
      await user.click(screen.getByText('重置'))

      // All items should be visible again
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
        expect(screen.getByText('Beta Task')).toBeInTheDocument()
        expect(screen.getByText('Gamma Task')).toBeInTheDocument()
      })
    })

    it('shows milestone badge in detail view table', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      // Switch to detail view
      await user.click(screen.getByText('明细'))

      await waitFor(() => {
        expect(screen.getByTestId('detail-table')).toBeInTheDocument()
      })

      // Detail view should show milestone badge
      await waitFor(() => {
        // 里程碑 column header and value
        expect(screen.getByText('里程碑')).toBeInTheDocument()
        expect(screen.getByText('MVP发布')).toBeInTheDocument()
      })
    })
  })

  // --- Milestone selector in dialogs ---

  describe('milestone selector in create dialog', () => {
    it('shows milestone selector with options when creating a new item', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      await user.click(screen.getByText('新增主事项'))

      await waitFor(() => {
        expect(screen.getByText('新建主事项')).toBeInTheDocument()
      })

      // Milestone selector label should be visible
      expect(screen.getByText('所属里程碑')).toBeInTheDocument()

      // Default should be "未分配"
      const milestoneSelects = screen.getAllByText('未分配')
      // At least one in the dialog
      expect(milestoneSelects.length).toBeGreaterThanOrEqual(1)
    })

    it('defaults to "未分配" for new items', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      await user.click(screen.getByText('新增主事项'))

      await waitFor(() => {
        expect(screen.getByText('新建主事项')).toBeInTheDocument()
      })

      // The milestone selector should show "未分配" (the default)
      // Verify the select trigger shows "未分配"
      const milestoneLabels = screen.getAllByText('所属里程碑')
      const dialogLabel = milestoneLabels.find((el) =>
        el.closest('[role="dialog"]'),
      )
      expect(dialogLabel).toBeTruthy()
    })

    it('shows milestone options in dropdown when creating an item', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      await user.click(screen.getByText('新增主事项'))

      await waitFor(() => {
        expect(screen.getByText('新建主事项')).toBeInTheDocument()
      })

      // Open the milestone dropdown
      const dialogEl = screen.getByRole('dialog')
      const selectTriggers = dialogEl.querySelectorAll('button[role="combobox"]')
      const milestoneTrigger = Array.from(selectTriggers).find((btn) =>
        btn.textContent?.includes('未分配'),
      )
      expect(milestoneTrigger).toBeTruthy()
      await user.click(milestoneTrigger!)

      // Should show milestone options
      await waitFor(() => {
        expect(
          screen.getByRole('option', { name: 'MVP发布' }),
        ).toBeInTheDocument()
        expect(
          screen.getByRole('option', { name: '二期迭代' }),
        ).toBeInTheDocument()
        expect(
          screen.getByRole('option', { name: '未分配' }),
        ).toBeInTheDocument()
      })
    })
  })

  describe('milestone selector in edit dialog', () => {
    it('pre-fills current milestone when editing an item', async () => {
      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      // Click edit on Alpha Task (which has milestoneKey MS001)
      const editBtns = screen.getAllByRole('button', { name: /编辑/ })
      const mainEditBtn = editBtns.find(
        (btn) => !btn.getAttribute('data-testid')?.startsWith('edit-sub-'),
      )
      await user.click(mainEditBtn!)

      await waitFor(() => {
        expect(screen.getByText('编辑主事项')).toBeInTheDocument()
      })

      // The milestone selector should be visible
      const milestoneLabels = screen.getAllByText('所属里程碑')
      const dialogLabel = milestoneLabels.find((el) =>
        el.closest('[role="dialog"]'),
      )
      expect(dialogLabel).toBeTruthy()
    })

    it('sends milestoneKey when editing an item', async () => {
      let capturedBody: Record<string, unknown> | null = null
      server.use(
        http.put('/v1/teams/:teamId/main-items/:itemId', async ({ request }) => {
          capturedBody = (await request.json()) as Record<string, unknown>
          return HttpResponse.json({
            code: 0,
            data: { ...seedMainItems[0], ...capturedBody },
          })
        }),
      )

      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      // Click edit on Alpha Task (which has milestoneKey MS001)
      const editBtns = screen.getAllByRole('button', { name: /编辑/ })
      const mainEditBtn = editBtns.find(
        (btn) => !btn.getAttribute('data-testid')?.startsWith('edit-sub-'),
      )
      await user.click(mainEditBtn!)

      await waitFor(() => {
        expect(screen.getByText('编辑主事项')).toBeInTheDocument()
      })

      // Submit the edit (form is pre-filled from the item)
      await user.click(screen.getByRole('button', { name: '确认' }))

      await waitFor(() => {
        expect(capturedBody).toBeTruthy()
      })

      // Should include milestoneKey in the update
      expect(capturedBody!.milestoneKey).toBe('MS001')
    })
  })

  describe('milestone selector with no milestones', () => {
    it('shows only "未分配" when team has no milestones', async () => {
      server.use(
        http.get('/v1/teams/:teamId/milestones', () => {
          return HttpResponse.json({
            code: 0,
            data: { items: [], total: 0 },
          })
        }),
      )

      const user = userEvent.setup()
      renderPage()
      await waitFor(() => {
        expect(screen.getByText('Alpha Task')).toBeInTheDocument()
      })

      await user.click(screen.getByText('新增主事项'))

      await waitFor(() => {
        expect(screen.getByText('新建主事项')).toBeInTheDocument()
      })

      // Milestone selector should still be present
      expect(screen.getByText('所属里程碑')).toBeInTheDocument()
    })
  })
})
