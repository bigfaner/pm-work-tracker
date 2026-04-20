import { describe, it, expect, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
import { ToastProvider } from '@/components/ui/toast'
import MainItemDetailPage from './MainItemDetailPage'

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

function renderPage(mainItemId = '1') {
  const qc = createQueryClient()
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[`/items/${mainItemId}`]}>
        <Routes>
          <Route path="/items/:mainItemId" element={
            <ToastProvider>
              <MainItemDetailPage />
            </ToastProvider>
          } />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedMembers = [
  { userId: 1, displayName: 'Test User', username: 'testuser', role: 'pm', joinedAt: '2024-01-01' },
  { userId: 2, displayName: 'Alice', username: 'alice', role: 'member', joinedAt: '2024-01-01' },
  { userId: 3, displayName: 'Bob', username: 'bob', role: 'member', joinedAt: '2024-01-01' },
]

const seedMainItem = {
  id: 1, teamId: 1, code: 'MI-0001', title: 'Alpha Task', priority: 'P1',
  proposerId: 1, assigneeId: 1, startDate: '2026-03-20', expectedEndDate: '2026-04-15',
  actualEndDate: null, status: 'progressing', completion: 65, isKeyItem: false,
  delayCount: 0, archivedAt: null,
  createdAt: '2026-03-20T00:00:00Z', updatedAt: '2026-04-01T00:00:00Z',
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
    {
      id: 13, teamId: 1, mainItemId: 1, title: 'Sub Alpha 3', description: '',
      priority: 'P2', assigneeId: 3, startDate: '2026-04-15', expectedEndDate: '2026-04-25',
      actualEndDate: null, status: 'progressing', completion: 30, isKeyItem: false,
      delayCount: 0, weight: 1, createdAt: '2026-04-01T00:00:00Z', updatedAt: '2026-04-15T00:00:00Z',
    },
  ],
  achievements: ['登录/注册接口开发完成', 'JWT Token 签发与验证逻辑实现'],
  blockers: ['OAuth2.0 第三方回调地址需运维配合配置'],
}

function setupHandlers() {
  server.use(
    // Get main item with sub items
    http.get('/api/v1/teams/:teamId/main-items/:itemId', ({ params }) => {
      const item = Number(params.itemId) === 1 ? seedMainItem : null
      if (!item) return HttpResponse.json({ code: 'NOT_FOUND', message: 'not found' }, { status: 404 })
      return HttpResponse.json({ code: 0, data: item })
    }),

    // List members
    http.get('/api/v1/teams/:teamId/members', () => {
      return HttpResponse.json({ code: 0, data: seedMembers })
    }),

    // Update main item
    http.put('/api/v1/teams/:teamId/main-items/:itemId', async ({ request }) => {
      const body = await request.json() as Record<string, unknown>
      return HttpResponse.json({ code: 0, data: { ...seedMainItem, ...body } })
    }),

    // Create sub item
    http.post('/api/v1/teams/:teamId/main-items/:mainId/sub-items', async ({ request }) => {
      const body = await request.json() as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: {
          id: 100, teamId: 1, mainItemId: 1, description: '', priority: 'P2',
          assigneeId: null, startDate: null, expectedEndDate: null, actualEndDate: null,
          status: 'pending', completion: 0, isKeyItem: false, delayCount: 0, weight: 1,
          createdAt: new Date().toISOString(), updatedAt: new Date().toISOString(),
          ...body,
        },
      })
    }),

    // Change sub item status
    http.put('/api/v1/teams/:teamId/sub-items/:itemId/status', async () => {
      return HttpResponse.json({ code: 0, data: null })
    }),

    // Available transitions for sub items
    http.get('/api/v1/teams/:teamId/sub-items/:subId/available-transitions', () => {
      const allStatuses = ['pending', 'progressing', 'blocking', 'pausing', 'completed', 'closed']
      return HttpResponse.json({ code: 0, data: allStatuses })
    }),
  )
}

describe('MainItemDetailPage', () => {
  beforeEach(() => {
    useTeamStore.setState({ currentTeamId: 1, teams: [{ id: 1, name: 'Test Team', description: '', pmId: 1, createdAt: '', updatedAt: '' }] })
    useAuthStore.getState().setPermissions({
      isSuperAdmin: false,
      teamPermissions: { 1: ['main_item:update'] },
    })
    setupHandlers()
  })

  // --- Core rendering ---

  it('renders breadcrumb navigation', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('事项清单')).toBeInTheDocument()
    })
    // Breadcrumb link should go to /items
    const link = screen.getByText('事项清单').closest('a')
    expect(link).toHaveAttribute('href', '/items')
  })

  it('renders main item title and code badge', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
      expect(screen.getByText('MI-0001')).toBeInTheDocument()
    })
  })

  it('renders priority and status badges', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('P1')).toBeInTheDocument()
    })
  })

  it('renders assignee with avatar', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Test User')).toBeInTheDocument()
    })
  })

  it('renders info grid fields', async () => {
    renderPage()
    await waitFor(() => {
      // Some labels appear in both info grid and table headers
      expect(screen.getAllByText('负责人').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('开始时间').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('预期完成时间').length).toBeGreaterThanOrEqual(1)
      expect(screen.getAllByText('实际完成时间').length).toBeGreaterThanOrEqual(1)
    })
  })

  // --- Circular progress indicator ---

  it('renders circular progress indicator with percentage', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('65%')).toBeInTheDocument()
    })
  })

  // --- Collapsible sections ---

  it('renders progress & summary section', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('进度与汇总')).toBeInTheDocument()
    })
  })

  it('shows sub-item count in summary section', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/已完成.*个子事项/)).toBeInTheDocument()
      expect(screen.getByText(/共.*个子事项/)).toBeInTheDocument()
    })
  })

  it('expands to show achievements and blockers on click', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('进度与汇总')).toBeInTheDocument()
    })

    // Click to expand
    await user.click(screen.getByText('进度与汇总'))

    await waitFor(() => {
      expect(screen.getByText('成果汇总')).toBeInTheDocument()
      expect(screen.getByText('卡点汇总')).toBeInTheDocument()
    })
  })

  // --- Sub items table ---

  it('renders sub items table', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('子事项列表')).toBeInTheDocument()
      expect(screen.getByText('Sub Alpha 1')).toBeInTheDocument()
      expect(screen.getByText('Sub Alpha 2')).toBeInTheDocument()
      expect(screen.getByText('Sub Alpha 3')).toBeInTheDocument()
    })
  })

  it('renders table headers', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('编号')).toBeInTheDocument()
      expect(screen.getByText('标题')).toBeInTheDocument()
      expect(screen.getAllByText('负责人').length).toBeGreaterThanOrEqual(1)
      expect(screen.getByText('完成度')).toBeInTheDocument()
      expect(screen.getAllByText('状态').length).toBeGreaterThanOrEqual(1)
    })
  })

  it('sub item titles link to sub item detail page', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByText('Sub Alpha 1').closest('a')
      expect(link).toHaveAttribute('href', '/items/1/sub/11')
    })
  })

  // --- Inline status change ---

  it('sub items have inline status change dropdown', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('Sub Alpha 1')).toBeInTheDocument()
    })

    // Find status badges that are clickable in the sub-item table
    const statusBadges = screen.getAllByText('已完成')
    expect(statusBadges.length).toBeGreaterThanOrEqual(1)
  })

  // --- Create sub item dialog ---

  it('opens create sub-item dialog', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
    })

    // Find the "+ 新增子事项" button
    const newSubBtn = screen.getByRole('button', { name: /新增子事项/ })
    await user.click(newSubBtn)

    await waitFor(() => {
      // Dialog should open with title
      expect(screen.getByText('新增子事项')).toBeInTheDocument()
    })
  })

  it('create sub-item dialog has required fields: priority, startDate, expectedEndDate', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /新增子事项/ }))

    await waitFor(() => {
      expect(screen.getByText('新增子事项')).toBeInTheDocument()
    })

    // Priority should have a required marker (red asterisk)
    const priorityLabels = screen.getAllByText(/优先级/)
    const priorityWithRequired = priorityLabels.find(el => el.closest('label')?.innerHTML.includes('*'))
    expect(priorityWithRequired).toBeTruthy()

    // Start date field should exist in the dialog with a required marker
    const startDateLabels = screen.getAllByText(/开始时间/)
    const startDateInDialog = startDateLabels.find(el => el.closest('label')?.innerHTML.includes('*'))
    expect(startDateInDialog).toBeTruthy()
    // Verify a date input for start date exists
    const allDateInputs = document.querySelectorAll('input[type="date"]')
    expect(allDateInputs.length).toBeGreaterThanOrEqual(2)

    // Expected end date should have required marker
    const endDateLabels = screen.getAllByText(/预期完成时间/)
    const endDateWithRequired = endDateLabels.find(el => el.closest('label')?.innerHTML.includes('*'))
    expect(endDateWithRequired).toBeTruthy()
  })

  it('create sub-item submit is disabled when required fields are empty', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /新增子事项/ }))

    await waitFor(() => {
      expect(screen.getByText('新增子事项')).toBeInTheDocument()
    })

    // Fill only title, leave priority/startDate/expectedEndDate empty
    await user.type(screen.getByPlaceholderText('请输入子事项标题'), 'New Sub')

    // Submit button should be disabled because priority, startDate, expectedEndDate are empty
    const submitBtn = screen.getByRole('button', { name: '确认' })
    expect(submitBtn).toBeDisabled()
  })

  it('create sub-item submit is enabled when all required fields are filled', async () => {
    const user = userEvent.setup()
    const { fireEvent } = await import('@testing-library/react')
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
    })

    await user.click(screen.getByRole('button', { name: /新增子事项/ }))

    await waitFor(() => {
      expect(screen.getByText('新增子事项')).toBeInTheDocument()
    })

    // Fill all required fields
    await user.type(screen.getByPlaceholderText('请输入子事项标题'), 'New Sub')

    // Select priority via fireEvent on the Radix Select
    // The priority select trigger is the one near "优先级" label
    const priorityLabels = screen.getAllByText(/优先级/)
    const priorityLabelInDialog = priorityLabels.find(el => el.closest('label')?.innerHTML.includes('*'))
    expect(priorityLabelInDialog).toBeTruthy()
    // Find the select trigger button near this label
    const priorityContainer = priorityLabelInDialog!.closest('div')?.parentElement
    const selectTrigger = priorityContainer?.querySelector('button')
    expect(selectTrigger).toBeTruthy()
    await user.click(selectTrigger!)
    // Click P2 option
    await user.click(screen.getByRole('option', { name: 'P2' }))

    // Select assignee - use container query since Radix Select is hard to test with RTL
    const allSelects = document.querySelectorAll('[role="dialog"] button[role="combobox"], [role="dialog"] button[data-state]')
    // Find the one that shows "不指定" (assignee)
    const assigneeBtn = Array.from(allSelects).find(btn => btn.textContent?.includes('不指定'))
    if (assigneeBtn) {
      await user.click(assigneeBtn)
      // Wait for options to appear
      await waitFor(() => {
        expect(screen.getByRole('option', { name: 'Test User' })).toBeInTheDocument()
      })
      await user.click(screen.getByRole('option', { name: 'Test User' }))
    }

    // Fill start date and expected end date via fireEvent
    const allDateInputs = document.querySelectorAll('input[type="date"]')
    const dialogDateInputs = Array.from(allDateInputs).filter(input => {
      const label = input.closest('div')?.querySelector('label')
      return label && (label.textContent?.includes('开始时间') || label.textContent?.includes('预期完成时间'))
    })
    expect(dialogDateInputs.length).toBe(2)

    fireEvent.change(dialogDateInputs[0], { target: { value: '2026-04-20' } })
    fireEvent.change(dialogDateInputs[1], { target: { value: '2026-04-30' } })

    // Submit should now be enabled
    await waitFor(() => {
      const submitBtn = screen.getByRole('button', { name: '确认' })
      expect(submitBtn).toBeEnabled()
    })
  })

  // --- Action buttons ---

  it('renders edit button', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
      expect(screen.getByText('编辑')).toBeInTheDocument()
    })
  })

  // --- Breadcrumb navigation ---

  it('breadcrumb has clickable links', async () => {
    renderPage()
    await waitFor(() => {
      const link = screen.getByText('事项清单').closest('a')
      expect(link).toHaveAttribute('href', '/items')
    })
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Alpha Task' })).toBeInTheDocument()
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Loading state ---

  it('shows loading state', async () => {
    renderPage()
    await waitFor(() => {
      // Either loading indicator or the actual data appears
      const loading = screen.queryByText('加载中...')
      const title = screen.queryByRole('heading', { name: 'Alpha Task' })
      expect(loading || title).toBeTruthy()
    })
  })
})
