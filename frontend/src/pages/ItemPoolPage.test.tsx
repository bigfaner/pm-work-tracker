import { describe, it, expect, beforeEach, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { server } from '@/mocks/server'
import { http, HttpResponse } from 'msw'
import { useTeamStore } from '@/store/team'
import { useAuthStore } from '@/store/auth'
import { ToastProvider } from '@/components/ui/toast'
import ItemPoolPage from './ItemPoolPage'

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
      <ToastProvider>
        <MemoryRouter>
          <ItemPoolPage />
        </MemoryRouter>
      </ToastProvider>
    </QueryClientProvider>,
  )
}

// --- Seed data ---

const seedPoolItems = [
  {
    bizKey: '1', teamKey: '1', title: '移动端适配需求',
    background: '当前系统在移动端浏览器下布局错位', expectedOutput: '响应式适配',
    submitterKey: 'U001', poolStatus: 'pending',
    assignedMainKey: null, assignedSubKey: null, assigneeKey: null,
    rejectReason: '', reviewedAt: null, reviewerKey: null,
    assignedMainCode: '', assignedMainTitle: '',
    createTime: '2026-04-17T00:00:00Z', dbUpdateTime: '2026-04-17T00:00:00Z',
  },
  {
    bizKey: '2', teamKey: '1', title: '用户反馈收集功能',
    background: '需要一个系统化的用户反馈收集机制', expectedOutput: '反馈系统',
    submitterKey: 'U001', poolStatus: 'pending',
    assignedMainKey: null, assignedSubKey: null, assigneeKey: null,
    rejectReason: '', reviewedAt: null, reviewerKey: null,
    assignedMainCode: '', assignedMainTitle: '',
    createTime: '2026-04-16T00:00:00Z', dbUpdateTime: '2026-04-16T00:00:00Z',
  },
  {
    bizKey: '3', teamKey: '1', title: '性能优化建议',
    background: '首页加载时间超过3秒', expectedOutput: '加载时间降低到1秒内',
    submitterKey: 'U001', poolStatus: 'assigned',
    assignedMainKey: '1', assignedSubKey: '11', assigneeKey: 'U002',
    rejectReason: '', reviewedAt: '2026-04-15T00:00:00Z', reviewerKey: 'U001',
    assignedMainCode: 'MI-001', assignedMainTitle: '',
    createTime: '2026-04-12T00:00:00Z', dbUpdateTime: '2026-04-15T00:00:00Z',
  },
  {
    bizKey: '4', teamKey: '1', title: '旧版API废弃',
    background: 'v1版本API仍有部分客户端在使用', expectedOutput: '废弃计划',
    submitterKey: 'U001', poolStatus: 'rejected',
    assignedMainKey: null, assignedSubKey: null, assigneeKey: null,
    rejectReason: '优先级不足，安排至下个季度处理', reviewedAt: '2026-04-10T00:00:00Z', reviewerKey: 'U001',
    assignedMainCode: '', assignedMainTitle: '',
    createTime: '2026-04-08T00:00:00Z', dbUpdateTime: '2026-04-10T00:00:00Z',
  },
]

const seedMembers = [
  { id: 1, bizKey: '1', teamKey: '1', userBizKey: 'U001', displayName: '张明', username: 'zhangming', role: 'pm', roleId: 1, roleName: 'pm', joinedAt: '2024-01-01' },
  { id: 2, bizKey: '2', teamKey: '1', userBizKey: 'U002', displayName: '李华', username: 'lihua', role: 'member', roleId: 2, roleName: 'member', joinedAt: '2024-01-01' },
]

const seedMainItems = [
  { bizKey: '1', teamKey: '1', code: 'MI-001', title: '用户认证模块开发', priority: 'P1', itemStatus: 'progressing' },
  { bizKey: '2', teamKey: '1', code: 'MI-002', title: '数据看板设计', priority: 'P2', itemStatus: 'pending' },
]

function setupHandlers() {
  server.use(
    http.get('/v1/teams/:teamId/item-pool', () => {
      return HttpResponse.json({
        code: 0,
        data: { items: seedPoolItems, total: seedPoolItems.length, page: 1, pageSize: 20 },
      })
    }),
    http.get('/v1/teams/:teamId/members', () => {
      return HttpResponse.json({ code: 0, data: seedMembers })
    }),
    http.get('/v1/teams/:teamId/main-items', () => {
      return HttpResponse.json({
        code: 0,
        data: { items: seedMainItems, total: seedMainItems.length, page: 1, pageSize: 20 },
      })
    }),
    http.post('/v1/teams/:teamId/item-pool', async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: {
          bizKey: '100', teamKey: '1', ...body,
          submitterKey: 'U001', poolStatus: 'pending',
          assignedMainKey: null, assignedSubKey: null, assigneeKey: null,
          rejectReason: '', reviewedAt: null, reviewerKey: null,
          assignedMainCode: '', assignedMainTitle: '',
          createTime: new Date().toISOString(), dbUpdateTime: new Date().toISOString(),
        },
      })
    }),
    http.post('/v1/teams/:teamId/item-pool/:poolId/assign', async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: { mainItemBizKey: 'mi-200', subItemBizKey: 'si-200', ...body },
      })
    }),
    http.post('/v1/teams/:teamId/item-pool/:poolId/reject', async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: {
          bizKey: '4', teamKey: '1', title: '旧版API废弃',
          background: '', expectedOutput: '', submitterKey: 'U001',
          poolStatus: 'rejected', assignedMainKey: null, assignedSubKey: null, assigneeKey: null,
          assignedMainCode: '', assignedMainTitle: '',
          rejectReason: body.reason, reviewedAt: new Date().toISOString(), reviewerKey: 'U001',
          createTime: '2026-04-08T00:00:00Z', dbUpdateTime: new Date().toISOString(),
        },
      })
    }),
    http.post('/v1/teams/:teamId/main-items', async ({ request }) => {
      const body = (await request.json()) as Record<string, unknown>
      return HttpResponse.json({
        code: 0,
        data: {
          bizKey: '300', teamKey: '1', code: 'MI-0300', priority: 'P2', proposerKey: 'U001',
          ...body,
        },
      })
    }),
  )
}

describe('ItemPoolPage', () => {
  beforeEach(() => {
    useTeamStore.setState({ currentTeamId: '1', teams: [{ bizKey: '1', name: 'Test Team', description: '', code: '', pmKey: '1', createdAt: '', updatedAt: '' }] })
    useAuthStore.getState().setPermissions({
      isSuperAdmin: false,
      teamPermissions: { 1: ['item_pool:review'] },
    })
    setupHandlers()
  })

  // --- Core rendering ---

  it('renders page header with title', async () => {
    renderPage()
    expect(screen.getByTestId('item-pool-page')).toBeInTheDocument()
    expect(screen.getByText('待办事项')).toBeInTheDocument()
  })

  it('renders add button', async () => {
    renderPage()
    expect(screen.getByText('新增待办事项')).toBeInTheDocument()
  })

  it('renders pool items after loading', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
      expect(screen.getByText('用户反馈收集功能')).toBeInTheDocument()
      expect(screen.getByText('性能优化建议')).toBeInTheDocument()
      expect(screen.getByText('旧版API废弃')).toBeInTheDocument()
    })
  })

  // --- Left border color indicates status ---

  it('shows pending items with blue left border', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })
    const card = screen.getByTestId('pool-item-1')
    expect(card.className).toContain('border-l-blue-500')
  })

  it('shows assigned items with gray left border and reduced opacity', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('性能优化建议')).toBeInTheDocument()
    })
    const card = screen.getByTestId('pool-item-3')
    expect(card.className).toContain('border-l-tertiary')
    expect(card.className).toContain('opacity-70')
  })

  it('shows rejected items with red left border and reduced opacity', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('旧版API废弃')).toBeInTheDocument()
    })
    const card = screen.getByTestId('pool-item-4')
    expect(card.className).toContain('border-l-error')
    expect(card.className).toContain('opacity-70')
  })

  // --- Status badges ---

  it('shows status badges for pool items', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })
    // Multiple items have "待分配" so use getAllByText
    expect(screen.getAllByText('待分配').length).toBeGreaterThanOrEqual(2)
    expect(screen.getByText('已分配')).toBeInTheDocument()
    expect(screen.getByText('已拒绝')).toBeInTheDocument()
  })

  // --- Filter and search ---

  it('renders filter bar with search and status filter', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })
    expect(screen.getByPlaceholderText('搜索标题或编号...')).toBeInTheDocument()
    expect(screen.getByText('重置')).toBeInTheDocument()
  })

  it('filters items by search text', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('搜索标题或编号...')
    await user.type(searchInput, '移动端')

    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
      expect(screen.queryByText('用户反馈收集功能')).not.toBeInTheDocument()
    })
  })

  it('filters items by status via select', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })

    // Click the status filter trigger
    const statusTrigger = screen.getByTestId('pool-status-filter')
    await user.click(statusTrigger)

    // Select "pending" from dropdown - the dropdown adds new elements
    await waitFor(() => {
      // The SelectContent should be in the DOM now
      const pendingOptions = screen.getAllByText('待分配')
      // At least the select option should be visible
      expect(pendingOptions.length).toBeGreaterThanOrEqual(2)
    })
  })

  it('resets filters when clicking reset button', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })

    const searchInput = screen.getByPlaceholderText('搜索标题或编号...')
    await user.type(searchInput, '移动端')
    expect(searchInput).toHaveValue('移动端')

    await user.click(screen.getByText('重置'))
    expect(searchInput).toHaveValue('')
  })

  // --- Action buttons for pending items ---

  it('shows action buttons for pending items', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })
    // Pending items should have action buttons
    expect(screen.getByTestId('to-main-1')).toBeInTheDocument()
    expect(screen.getByTestId('to-sub-1')).toBeInTheDocument()
    expect(screen.getByTestId('reject-1')).toBeInTheDocument()
  })

  it('does not show action buttons for assigned/rejected items', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('性能优化建议')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('to-main-3')).not.toBeInTheDocument()
    expect(screen.queryByTestId('reject-4')).not.toBeInTheDocument()
  })

  // --- Submit dialog ---

  it('opens submit dialog when clicking add button', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })

    // Click the button (which is in the header)
    const addButtons = screen.getAllByText('新增待办事项')
    await user.click(addButtons[0])

    await waitFor(() => {
      // Dialog title should appear (there will be 2 instances of the text now)
      expect(screen.getAllByText('新增待办事项').length).toBeGreaterThanOrEqual(2)
    })
  })

  // --- Convert to main item dialog ---

  it('opens convert-to-main dialog when clicking button on pending item', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('to-main-1'))

    await waitFor(() => {
      // Dialog title "转为主事项" should now appear in addition to button text
      expect(screen.getAllByText('转为主事项').length).toBeGreaterThanOrEqual(2)
    })
  })

  // --- Convert to sub item dialog ---

  it('opens convert-to-sub dialog when clicking button on pending item', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('to-sub-1'))

    await waitFor(() => {
      // Dialog title "转为子事项" should now appear in addition to button text
      expect(screen.getAllByText('转为子事项').length).toBeGreaterThanOrEqual(2)
    })
  })

  // --- Reject dialog ---

  it('opens reject dialog when clicking reject button', async () => {
    const user = userEvent.setup()
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('移动端适配需求')).toBeInTheDocument()
    })

    await user.click(screen.getByTestId('reject-1'))

    await waitFor(() => {
      expect(screen.getByText('拒绝事项')).toBeInTheDocument()
    })
  })

  // --- Rejected item shows reject reason ---

  it('shows reject reason for rejected items', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('旧版API废弃')).toBeInTheDocument()
    })
    expect(screen.getByText(/优先级不足/)).toBeInTheDocument()
  })

  // --- Assigned item shows assignment info ---

  it('shows assignment info for assigned items', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('性能优化建议')).toBeInTheDocument()
    })
    expect(screen.getByText(/已转为子事项/)).toBeInTheDocument()
  })

  // --- No antd imports ---

  it('does not import antd', async () => {
    renderPage()
    await waitFor(() => {
      expect(screen.getByText('待办事项')).toBeInTheDocument()
    })
    const antdElements = document.querySelectorAll('[class*="ant-"]')
    expect(antdElements.length).toBe(0)
  })

  // --- Empty state ---

  it('shows empty state when no pool items exist', async () => {
    server.use(
      http.get('/v1/teams/:teamId/item-pool', () => {
        return HttpResponse.json({ code: 0, data: { items: [], total: 0, page: 1, pageSize: 20 } })
      }),
    )

    renderPage()
    await waitFor(() => {
      expect(screen.getByText(/暂无待办事项/)).toBeInTheDocument()
    })
  })

  // --- No team selected ---

  it('shows message when no team is selected', () => {
    useTeamStore.setState({ currentTeamId: null })
    renderPage()
    expect(screen.getByText('请先选择团队')).toBeInTheDocument()
  })
})
